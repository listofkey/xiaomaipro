package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/order/orderservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelOrderLogic) CancelOrder(req *types.CancelOrderReq) (*types.OperationResp, error) {
	resp, err := l.svcCtx.OrderRpc.CancelOrder(l.ctx, &orderservice.CancelOrderReq{
		UserId:  authctx.UserID(l.ctx),
		OrderNo: req.OrderNo,
	})
	if err != nil {
		return nil, err
	}

	return &types.OperationResp{Success: resp.Success}, nil
}
