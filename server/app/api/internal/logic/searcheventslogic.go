package logic

import (
	"context"

	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/program/programservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchEventsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchEventsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchEventsLogic {
	return &SearchEventsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchEventsLogic) SearchEvents(req *types.SearchEventsReq) (*types.ProgramEventListResp, error) {
	resp, err := l.svcCtx.ProgramRpc.SearchEvent(l.ctx, &programservice.SearchEventReq{
		Keyword:    req.Keyword,
		CategoryId: req.CategoryId,
		City:       req.City,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Page:       req.Page,
		PageSize:   req.PageSize,
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
