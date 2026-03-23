package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"server/app/rpc/dao"
	"server/app/rpc/model"
	"server/app/rpc/program/internal/svc"
	"server/app/rpc/program/programpb/programpb"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	defaultPage       int32 = 1
	defaultPageSize   int32 = 20
	maxPageSize       int32 = 100
	defaultHotLimit   int32 = 8
	maxHotLimit       int32 = 50
	activeEventStatus int16 = 1
)

type pagination struct {
	page     int32
	pageSize int32
	offset   int
}

type minPriceRow struct {
	EventID  int64   `gorm:"column:event_id"`
	MinPrice float64 `gorm:"column:min_price"`
}

func normalizePagination(page, pageSize int32) pagination {
	if page <= 0 {
		page = defaultPage
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	return pagination{
		page:     page,
		pageSize: pageSize,
		offset:   int((page - 1) * pageSize),
	}
}

func normalizeHotLimit(limit int32) int32 {
	if limit <= 0 {
		return defaultHotLimit
	}
	if limit > maxHotLimit {
		return maxHotLimit
	}
	return limit
}

func normalizeSortBy(sortBy string) string {
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "time", "price", "hot":
		return strings.ToLower(strings.TrimSpace(sortBy))
	default:
		return "hot"
	}
}

func normalizeCategoryStatus(status int32) int32 {
	switch status {
	case -1, 0, 1:
		return status
	default:
		return 1
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func parseDateRange(startDate, endDate string) (*time.Time, *time.Time, error) {
	var (
		start        *time.Time
		endExclusive *time.Time
		err          error
	)

	if strings.TrimSpace(startDate) != "" {
		start, err = parseDate(startDate, false)
		if err != nil {
			return nil, nil, err
		}
	}
	if strings.TrimSpace(endDate) != "" {
		endExclusive, err = parseDate(endDate, true)
		if err != nil {
			return nil, nil, err
		}
	}

	if start != nil && endExclusive != nil && !start.Before(*endExclusive) {
		return nil, nil, fmt.Errorf("start_date must be earlier than or equal to end_date")
	}

	return start, endExclusive, nil
}

func parseDate(raw string, endExclusive bool) (*time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(raw), time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q, expected YYYY-MM-DD", raw)
	}
	if endExclusive {
		parsed = parsed.AddDate(0, 0, 1)
	}
	return &parsed, nil
}

func readCache(ctx context.Context, redisClient *redis.Client, key string, dest any) bool {
	if redisClient == nil {
		return false
	}

	data, err := redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return false
	}

	return json.Unmarshal(data, dest) == nil
}

func writeCache(ctx context.Context, redisClient *redis.Client, key string, value any, ttl time.Duration) {
	if redisClient == nil {
		return
	}

	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	_ = redisClient.Set(ctx, key, data, ttl).Err()
}

func applyEventFilters(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	db *gorm.DB,
	categoryID int32,
	cityName string,
	keyword string,
	startDate string,
	endDate string,
) (*gorm.DB, bool, error) {
	db = db.WithContext(ctx).
		Model(&model.Event{}).
		Where("status = ?", activeEventStatus).
		Where("event_end_time >= ?", time.Now())

	if categoryID > 0 {
		db = db.Where("category_id = ?", categoryID)
	}

	if strings.TrimSpace(keyword) != "" {
		like := "%" + strings.TrimSpace(keyword) + "%"
		db = db.Where("(title ILIKE ? OR artist ILIKE ?)", like, like)
	}

	if strings.TrimSpace(cityName) != "" {
		cityID, found, err := lookupCityID(ctx, svcCtx.Query, cityName)
		if err != nil {
			return nil, false, err
		}
		if !found {
			return db, false, nil
		}
		db = db.Where("city_id = ?", cityID)
	}

	start, endExclusive, err := parseDateRange(startDate, endDate)
	if err != nil {
		return nil, false, err
	}
	if start != nil {
		db = db.Where("event_end_time >= ?", *start)
	}
	if endExclusive != nil {
		db = db.Where("event_start_time < ?", *endExclusive)
	}

	return db, true, nil
}

func applyListSort(db *gorm.DB, svcCtx *svc.ServiceContext, sortBy string) *gorm.DB {
	switch normalizeSortBy(sortBy) {
	case "time":
		return db.Order("event_start_time ASC").Order("id DESC")
	case "price":
		priceSubQuery := svcCtx.DB.Table("ticket_tier").
			Select("event_id, COALESCE(MIN(CASE WHEN status = 1 THEN price END), MIN(price)) AS min_price").
			Group("event_id")

		return db.Joins("LEFT JOIN (?) AS ticket_price ON ticket_price.event_id = event.id", priceSubQuery).
			Order("ticket_price.min_price ASC NULLS LAST").
			Order("event_start_time ASC").
			Order("id DESC")
	default:
		return db.Order("created_at DESC").Order("id DESC")
	}
}

func lookupCityID(ctx context.Context, query *dao.Query, cityName string) (int64, bool, error) {
	city, err := query.City.WithContext(ctx).Where(query.City.Name.Eq(strings.TrimSpace(cityName))).First()
	if err == nil {
		return city.ID, true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, false, nil
	}
	return 0, false, err
}

func buildEventBriefs(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	events []*model.Event,
	isHot bool,
) ([]*programpb.EventBrief, error) {
	if len(events) == 0 {
		return []*programpb.EventBrief{}, nil
	}

	categoryIDs := make([]int32, 0, len(events))
	cityIDs := make([]int64, 0, len(events))
	venueIDs := make([]int64, 0, len(events))
	eventIDs := make([]int64, 0, len(events))
	for _, event := range events {
		categoryIDs = append(categoryIDs, event.CategoryID)
		cityIDs = append(cityIDs, event.CityID)
		venueIDs = append(venueIDs, event.VenueID)
		eventIDs = append(eventIDs, event.ID)
	}

	categoryMap, err := loadCategoryMap(ctx, svcCtx.Query, categoryIDs)
	if err != nil {
		return nil, err
	}
	cityMap, err := loadCityMap(ctx, svcCtx.Query, cityIDs)
	if err != nil {
		return nil, err
	}
	venueMap, err := loadVenueMap(ctx, svcCtx.Query, venueIDs)
	if err != nil {
		return nil, err
	}
	minPriceMap, err := loadMinPriceMap(ctx, svcCtx.DB, eventIDs)
	if err != nil {
		return nil, err
	}

	briefs := make([]*programpb.EventBrief, 0, len(events))
	for _, event := range events {
		cityName := ""
		if city := cityMap[event.CityID]; city != nil {
			cityName = city.Name
		}

		venueName := ""
		if venue := venueMap[event.VenueID]; venue != nil {
			venueName = venue.Name
		}

		briefs = append(briefs, modelEventToBrief(
			event,
			categoryMap[event.CategoryID],
			cityName,
			venueName,
			minPriceMap[event.ID],
			isHot,
		))
	}

	return briefs, nil
}

func loadCategoryMap(ctx context.Context, query *dao.Query, ids []int32) (map[int32]*model.Category, error) {
	result := make(map[int32]*model.Category)
	uniqueIDs := uniqueInt32(ids)
	if len(uniqueIDs) == 0 {
		return result, nil
	}

	items, err := query.Category.WithContext(ctx).Where(query.Category.ID.In(uniqueIDs...)).Find()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		result[item.ID] = item
	}
	return result, nil
}

func loadCityMap(ctx context.Context, query *dao.Query, ids []int64) (map[int64]*model.City, error) {
	result := make(map[int64]*model.City)
	uniqueIDs := uniqueInt64(ids)
	if len(uniqueIDs) == 0 {
		return result, nil
	}

	items, err := query.City.WithContext(ctx).Where(query.City.ID.In(uniqueIDs...)).Find()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		result[item.ID] = item
	}
	return result, nil
}

func loadVenueMap(ctx context.Context, query *dao.Query, ids []int64) (map[int64]*model.Venue, error) {
	result := make(map[int64]*model.Venue)
	uniqueIDs := uniqueInt64(ids)
	if len(uniqueIDs) == 0 {
		return result, nil
	}

	items, err := query.Venue.WithContext(ctx).Where(query.Venue.ID.In(uniqueIDs...)).Find()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		result[item.ID] = item
	}
	return result, nil
}

func loadMinPriceMap(ctx context.Context, db *gorm.DB, eventIDs []int64) (map[int64]float64, error) {
	result := make(map[int64]float64)
	uniqueIDs := uniqueInt64(eventIDs)
	if len(uniqueIDs) == 0 {
		return result, nil
	}

	rows := make([]minPriceRow, 0, len(uniqueIDs))
	if err := db.WithContext(ctx).
		Table("ticket_tier").
		Select("event_id, COALESCE(MIN(CASE WHEN status = 1 THEN price END), MIN(price)) AS min_price").
		Where("event_id IN ?", uniqueIDs).
		Group("event_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.EventID] = row.MinPrice
	}
	return result, nil
}

func uniqueInt32(values []int32) []int32 {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[int32]struct{}, len(values))
	result := make([]int32, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func uniqueInt64(values []int64) []int64 {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
