package logic

import (
	"context"

	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/program/programservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEventDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetEventDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEventDetailLogic {
	return &GetEventDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetEventDetailLogic) GetEventDetail(req *types.GetEventDetailReq) (*types.ProgramEventDetailResp, error) {
	eventID, err := parseID(req.EventId, "eventId")
	if err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.ProgramRpc.GetEventDetail(l.ctx, &programservice.GetEventDetailReq{
		EventId: eventID,
	})
	if err != nil {
		return nil, err
	}

	return &types.ProgramEventDetailResp{
		Event: mapProgramEventDetail(resp.Event),
	}, nil
}
