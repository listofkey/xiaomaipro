package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/order/orderservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApplyRefundLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApplyRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyRefundLogic {
	return &ApplyRefundLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApplyRefundLogic) ApplyRefund(req *types.ApplyRefundReq) (*types.ApplyRefundResp, error) {
	resp, err := l.svcCtx.OrderRpc.ApplyRefund(l.ctx, &orderservice.ApplyRefundReq{
		UserId:  authctx.UserID(l.ctx),
		OrderNo: req.OrderNo,
		Reason:  req.Reason,
	})
	if err != nil {
		return nil, err
	}

	return &types.ApplyRefundResp{
		Success: resp.Success,
		OrderNo: resp.OrderNo,
		Refund:  mapOrderRefund(resp.Refund),
	}, nil
}
