package logic

import (
	"context"

	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/program/programservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHotRecommendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetHotRecommendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHotRecommendLogic {
	return &GetHotRecommendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHotRecommendLogic) GetHotRecommend(req *types.GetHotRecommendReq) (*types.ProgramHotRecommendResp, error) {
	resp, err := l.svcCtx.ProgramRpc.GetHotRecommend(l.ctx, &programservice.GetHotRecommendReq{
		City:  req.City,
		Limit: req.Limit,
	})
	if err != nil {
		return nil, err
	}

	return &types.ProgramHotRecommendResp{
		Events: mapProgramEventBriefList(resp.Events),
	}, nil
}
