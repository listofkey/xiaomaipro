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

const listEventCacheTTL = 2 * time.Minute

func listEventCacheKey(in *programpb.ListEventReq) string {
	return fmt.Sprintf(
		"%scat%d_city%s_sd%s_ed%s_sort%s_p%d_ps%d",
		svc.PrefixEventList,
		in.CategoryId,
		strings.TrimSpace(in.City),
		in.StartDate,
		in.EndDate,
		normalizeSortBy(in.SortBy),
		in.Page,
		in.PageSize,
	)
}

type ListEventLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListEventLogic {
	return &ListEventLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListEventLogic) ListEvent(in *programpb.ListEventReq) (*programpb.ListEventResp, error) {
	page := normalizePagination(in.Page, in.PageSize)
	cacheKey := listEventCacheKey(&programpb.ListEventReq{
		Page:       page.page,
		PageSize:   page.pageSize,
		CategoryId: in.CategoryId,
		City:       strings.TrimSpace(in.City),
		StartDate:  strings.TrimSpace(in.StartDate),
		EndDate:    strings.TrimSpace(in.EndDate),
		SortBy:     normalizeSortBy(in.SortBy),
	})

	var cached programpb.ListEventResp
	if readCache(l.ctx, l.svcCtx.Redis, cacheKey, &cached) {
		l.Infof("ListEvent cache hit: %s", cacheKey)
		return &cached, nil
	}

	db, hasResult, err := applyEventFilters(
		l.ctx,
		l.svcCtx,
		l.svcCtx.DB,
		in.CategoryId,
		in.City,
		"",
		in.StartDate,
		in.EndDate,
	)
	if err != nil {
		return nil, err
	}
	if !hasResult {
		return &programpb.ListEventResp{
			Events:   []*programpb.EventBrief{},
			Total:    0,
			Page:     page.page,
			PageSize: page.pageSize,
		}, nil
	}

	db = applyListSort(db, l.svcCtx, in.SortBy)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count events failed: %w", err)
	}

	events := make([]*model.Event, 0, page.pageSize)
	if err := db.Offset(page.offset).Limit(int(page.pageSize)).Find(&events).Error; err != nil {
		return nil, fmt.Errorf("list events failed: %w", err)
	}

	briefs, err := buildEventBriefs(l.ctx, l.svcCtx, events, false)
	if err != nil {
		return nil, fmt.Errorf("assemble events failed: %w", err)
	}

	resp := &programpb.ListEventResp{
		Events:   briefs,
		Total:    total,
		Page:     page.page,
		PageSize: page.pageSize,
	}
	writeCache(l.ctx, l.svcCtx.Redis, cacheKey, resp, listEventCacheTTL)

	return resp, nil
}
