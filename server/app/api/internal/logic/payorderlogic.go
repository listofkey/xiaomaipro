package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/order/orderservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PayOrderLogic {
	return &PayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayOrderLogic) PayOrder(req *types.PayOrderReq) (*types.PayOrderResp, error) {
	resp, err := l.svcCtx.OrderRpc.PayOrder(l.ctx, &orderservice.PayOrderReq{
		UserId:    authctx.UserID(l.ctx),
		OrderNo:   req.OrderNo,
		PayMethod: req.PayMethod,
		Channel:   req.Channel,
	})
	if err != nil {
		return nil, err
	}

	return &types.PayOrderResp{
		Success:           resp.Success,
		PayForm:           resp.PayForm,
		Payment:           mapOrderPayment(resp.Payment),
		OrderStatus:       resp.OrderStatus,
		PaidAt:            resp.PaidAt,
		CheckoutUrl:       resp.CheckoutUrl,
		CheckoutSessionId: resp.CheckoutSessionId,
		SessionExpiresAt:  resp.SessionExpiresAt,
	}, nil
}
