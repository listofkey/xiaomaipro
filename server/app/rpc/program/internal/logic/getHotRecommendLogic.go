package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"server/app/rpc/model"
	"server/app/rpc/program/internal/svc"
	"server/app/rpc/program/programpb/programpb"

	"github.com/zeromicro/go-zero/core/logx"
)

const hotRecommendCacheTTL = 5 * time.Minute

type GetHotRecommendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHotRecommendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHotRecommendLogic {
	return &GetHotRecommendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetHotRecommendLogic) GetHotRecommend(in *programpb.GetHotRecommendReq) (*programpb.GetHotRecommendResp, error) {
	limit := normalizeHotLimit(in.Limit)
	city := strings.TrimSpace(in.City)
	cacheKey := fmt.Sprintf("%s:lim%d", svc.HotRecommendKey(city), limit)

	var cached programpb.GetHotRecommendResp
	if readCache(l.ctx, l.svcCtx.Redis, cacheKey, &cached) {
		l.Infof("GetHotRecommend cache hit: city=%s", city)
		return &cached, nil
	}

	baseQuery := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.Event{}).
		Where("status = ?", activeEventStatus).
		Where("event_end_time >= ?", time.Now()).
		Order("created_at DESC").
		Order("id DESC")

	events := make([]*model.Event, 0, limit)
	excludeIDs := make([]int64, 0, limit)

	if city != "" {
		if cityID, found, err := lookupCityID(l.ctx, l.svcCtx.Query, city); err != nil {
			return nil, err
		} else if found {
			cityEvents := make([]*model.Event, 0, limit)
			if err := baseQuery.Where("city_id = ?", cityID).Limit(int(limit)).Find(&cityEvents).Error; err != nil {
				return nil, fmt.Errorf("query city hot events failed: %w", err)
			}
			events = append(events, cityEvents...)
			for _, event := range cityEvents {
				excludeIDs = append(excludeIDs, event.ID)
			}
		}
	}

	if int32(len(events)) < limit {
		need := int(limit) - len(events)
		globalQuery := baseQuery
		if len(excludeIDs) > 0 {
			globalQuery = globalQuery.Not("id IN ?", excludeIDs)
		}

		globalEvents := make([]*model.Event, 0, need)
		if err := globalQuery.Limit(need).Find(&globalEvents).Error; err != nil {
			return nil, fmt.Errorf("query global hot events failed: %w", err)
		}
		events = append(events, globalEvents...)
	}

	briefs, err := buildEventBriefs(l.ctx, l.svcCtx, events, true)
	if err != nil {
		return nil, fmt.Errorf("assemble hot events failed: %w", err)
	}

	resp := &programpb.GetHotRecommendResp{Events: briefs}
	writeCache(l.ctx, l.svcCtx.Redis, cacheKey, resp, hotRecommendCacheTTL)

	return resp, nil
}
