package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"server/app/rpc/program/internal/svc"
	"server/app/rpc/program/programpb/programpb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const eventDetailCacheTTL = 10 * time.Minute

type GetEventDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetEventDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEventDetailLogic {
	return &GetEventDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetEventDetailLogic) GetEventDetail(in *programpb.GetEventDetailReq) (*programpb.GetEventDetailResp, error) {
	if in.EventId <= 0 {
		return nil, fmt.Errorf("event_id is required")
	}

	cacheKey := svc.EventDetailKey(in.EventId)
	var cached programpb.GetEventDetailResp
	if readCache(l.ctx, l.svcCtx.Redis, cacheKey, &cached) {
		l.Infof("GetEventDetail cache hit: event_id=%d", in.EventId)
		return &cached, nil
	}

	q := l.svcCtx.Query

	event, err := q.Event.WithContext(l.ctx).Where(q.Event.ID.Eq(in.EventId)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("query event failed: %w", err)
	}

	category, err := q.Category.WithContext(l.ctx).Where(q.Category.ID.Eq(event.CategoryID)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("query category failed: %w", err)
	}

	city, err := q.City.WithContext(l.ctx).Where(q.City.ID.Eq(event.CityID)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("query city failed: %w", err)
	}

	venue, err := q.Venue.WithContext(l.ctx).Where(q.Venue.ID.Eq(event.VenueID)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("query venue failed: %w", err)
	}

	var venueCityName string
	if venue != nil {
		venueCity, venueCityErr := q.City.WithContext(l.ctx).Where(q.City.ID.Eq(venue.CityID)).First()
		if venueCityErr != nil && !errors.Is(venueCityErr, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("query venue city failed: %w", venueCityErr)
		}
		if venueCity != nil {
			venueCityName = venueCity.Name
		}
	}

	tiers, err := q.TicketTier.WithContext(l.ctx).
		Where(q.TicketTier.EventID.Eq(event.ID)).
		Order(q.TicketTier.SortOrder).
		Order(q.TicketTier.ID).
		Find()
	if err != nil {
		return nil, fmt.Errorf("query ticket tiers failed: %w", err)
	}

	tierInfos := make([]*programpb.TicketTierInfo, 0, len(tiers))
	for _, tier := range tiers {
		tierInfos = append(tierInfos, modelTicketTierToInfo(tier))
	}

	cityName := ""
	if city != nil {
		cityName = city.Name
	}

	resp := &programpb.GetEventDetailResp{
		Event: &programpb.EventDetail{
			Id:             event.ID,
			Title:          event.Title,
			Description:    event.Description,
			PosterUrl:      event.PosterURL,
			Category:       modelCategoryToInfo(category),
			Venue:          modelVenueToInfo(venue, venueCityName),
			City:           cityName,
			Artist:         event.Artist,
			EventStartTime: formatTime(event.EventStartTime),
			EventEndTime:   formatTime(event.EventEndTime),
			SaleStartTime:  formatTime(event.SaleStartTime),
			SaleEndTime:    formatTime(event.SaleEndTime),
			Status:         int32(event.Status),
			PurchaseLimit:  event.PurchaseLimit,
			NeedRealName:   int32(event.NeedRealName),
			TicketType:     int32(event.TicketType),
			TicketTiers:    tierInfos,
		},
	}

	writeCache(l.ctx, l.svcCtx.Redis, cacheKey, resp, eventDetailCacheTTL)
	return resp, nil
}
