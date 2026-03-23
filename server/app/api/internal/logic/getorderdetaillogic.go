package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/order/orderservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderDetailLogic {
	return &GetOrderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderDetailLogic) GetOrderDetail(req *types.GetOrderDetailReq) (*types.OrderDetailResp, error) {
	resp, err := l.svcCtx.OrderRpc.GetOrderDetail(l.ctx, &orderservice.GetOrderDetailReq{
		UserId:  authctx.UserID(l.ctx),
		OrderNo: req.OrderNo,
	})
	if err != nil {
		return nil, err
	}

	return &types.OrderDetailResp{
		Order: mapOrderDetail(resp.Order),
	}, nil
}
