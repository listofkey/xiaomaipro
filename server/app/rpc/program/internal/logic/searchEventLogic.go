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

const searchEventCacheTTL = 1 * time.Minute

func searchEventCacheKey(in *programpb.SearchEventReq) string {
	return fmt.Sprintf(
		"%skw%s_cat%d_city%s_sd%s_ed%s_p%d_ps%d",
		svc.PrefixEventSearch,
		strings.TrimSpace(in.Keyword),
		in.CategoryId,
		strings.TrimSpace(in.City),
		in.StartDate,
		in.EndDate,
		in.Page,
		in.PageSize,
	)
}

type SearchEventLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchEventLogic {
	return &SearchEventLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SearchEventLogic) SearchEvent(in *programpb.SearchEventReq) (*programpb.SearchEventResp, error) {
	page := normalizePagination(in.Page, in.PageSize)
	cacheKey := searchEventCacheKey(&programpb.SearchEventReq{
		Keyword:    strings.TrimSpace(in.Keyword),
		CategoryId: in.CategoryId,
		City:       strings.TrimSpace(in.City),
		StartDate:  strings.TrimSpace(in.StartDate),
		EndDate:    strings.TrimSpace(in.EndDate),
		Page:       page.page,
		PageSize:   page.pageSize,
	})

	var cached programpb.SearchEventResp
	if readCache(l.ctx, l.svcCtx.Redis, cacheKey, &cached) {
		l.Infof("SearchEvent cache hit: %s", cacheKey)
		return &cached, nil
	}

	db, hasResult, err := applyEventFilters(
		l.ctx,
		l.svcCtx,
		l.svcCtx.DB,
		in.CategoryId,
		in.City,
		in.Keyword,
		in.StartDate,
		in.EndDate,
	)
	if err != nil {
		return nil, err
	}
	if !hasResult {
		return &programpb.SearchEventResp{
			Events:   []*programpb.EventBrief{},
			Total:    0,
			Page:     page.page,
			PageSize: page.pageSize,
		}, nil
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count search events failed: %w", err)
	}

	events := make([]*model.Event, 0, page.pageSize)
	if err := db.Order("event_start_time ASC").Order("id DESC").
		Offset(page.offset).
		Limit(int(page.pageSize)).
		Find(&events).Error; err != nil {
		return nil, fmt.Errorf("search events failed: %w", err)
	}

	briefs, err := buildEventBriefs(l.ctx, l.svcCtx, events, false)
	if err != nil {
		return nil, fmt.Errorf("assemble search events failed: %w", err)
	}

	resp := &programpb.SearchEventResp{
		Events:   briefs,
		Total:    total,
		Page:     page.page,
		PageSize: page.pageSize,
	}
	writeCache(l.ctx, l.svcCtx.Redis, cacheKey, resp, searchEventCacheTTL)

	return resp, nil
}
