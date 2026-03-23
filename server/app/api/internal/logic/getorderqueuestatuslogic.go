package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/order/orderservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderQueueStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderQueueStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderQueueStatusLogic {
	return &GetOrderQueueStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderQueueStatusLogic) GetOrderQueueStatus(req *types.GetOrderQueueStatusReq) (*types.OrderQueueStatusResp, error) {
	resp, err := l.svcCtx.OrderRpc.GetQueueStatus(l.ctx, &orderservice.GetQueueStatusReq{
		UserId:     authctx.UserID(l.ctx),
		QueueToken: req.QueueToken,
	})
	if err != nil {
		return nil, err
	}

	return &types.OrderQueueStatusResp{
		QueueToken:  resp.QueueToken,
		OrderNo:     resp.OrderNo,
		QueueStatus: resp.QueueStatus,
		Message:     resp.Message,
		Order:       mapOrderSummary(resp.Order),
	}, nil
}
