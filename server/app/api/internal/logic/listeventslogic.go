package logic

import (
	"context"

	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/program/programservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListEventsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListEventsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListEventsLogic {
	return &ListEventsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListEventsLogic) ListEvents(req *types.ListEventsReq) (*types.ProgramEventListResp, error) {
	resp, err := l.svcCtx.ProgramRpc.ListEvent(l.ctx, &programservice.ListEventReq{
		Page:       req.Page,
		PageSize:   req.PageSize,
		CategoryId: req.CategoryId,
		City:       req.City,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		SortBy:     req.SortBy,
	})
	if err != nil {
		return nil, err
	}

	return &types.ProgramEventListResp{
		Events:   mapProgramEventBriefList(resp.Events),
		Total:    resp.Total,
		Page:     resp.Page,
		PageSize: resp.PageSize,
	}, nil
}
