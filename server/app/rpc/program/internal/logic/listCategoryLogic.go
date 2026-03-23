package logic

import (
	"context"
	"fmt"
	"time"

	"server/app/rpc/model"
	"server/app/rpc/program/internal/svc"
	"server/app/rpc/program/programpb/programpb"

	"github.com/zeromicro/go-zero/core/logx"
)

const categoryCacheTTL = 30 * time.Minute

type ListCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCategoryLogic {
	return &ListCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListCategoryLogic) ListCategory(in *programpb.ListCategoryReq) (*programpb.ListCategoryResp, error) {
	status := normalizeCategoryStatus(in.Status)
	cacheKey := fmt.Sprintf("%s:status%d", svc.PrefixCategoryList, status)

	var cached programpb.ListCategoryResp
	if readCache(l.ctx, l.svcCtx.Redis, cacheKey, &cached) {
		l.Infof("ListCategory cache hit: status=%d", status)
		return &cached, nil
	}

	q := l.svcCtx.Query
	categoryQuery := q.Category.WithContext(l.ctx).Order(q.Category.SortOrder).Order(q.Category.ID)
	if status >= 0 {
		categoryQuery = categoryQuery.Where(q.Category.Status.Eq(int16(status)))
	}

	categories, err := categoryQuery.Find()
	if err != nil {
		return nil, fmt.Errorf("query categories failed: %w", err)
	}

	categoryInfos := make([]*programpb.CategoryInfo, 0, len(categories))
	for _, category := range categories {
		categoryInfos = append(categoryInfos, modelCategoryToInfo(category))
	}

	var cityIDs []int64
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.Event{}).
		Where("status = ?", activeEventStatus).
		Where("event_end_time >= ?", time.Now()).
		Distinct("city_id").
		Pluck("city_id", &cityIDs).Error; err != nil {
		return nil, fmt.Errorf("query event cities failed: %w", err)
	}

	cityInfos := make([]*programpb.CityInfo, 0, len(cityIDs))
	if len(cityIDs) > 0 {
		cities, err := q.City.WithContext(l.ctx).
			Where(q.City.ID.In(uniqueInt64(cityIDs)...)).
			Order(q.City.Name).
			Find()
		if err != nil {
			return nil, fmt.Errorf("query cities failed: %w", err)
		}
		for _, city := range cities {
			cityInfos = append(cityInfos, modelCityToInfo(city))
		}
	}

	resp := &programpb.ListCategoryResp{
		Categories: categoryInfos,
		Cities:     cityInfos,
	}
	writeCache(l.ctx, l.svcCtx.Redis, cacheKey, resp, categoryCacheTTL)

	return resp, nil
}
