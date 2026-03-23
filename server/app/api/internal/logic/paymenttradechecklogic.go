package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/payment/paymentservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type PaymentTradeCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPaymentTradeCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PaymentTradeCheckLogic {
	return &PaymentTradeCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PaymentTradeCheckLogic) PaymentTradeCheck(req *types.PaymentTradeCheckReq) (*types.PaymentTradeCheckResp, error) {
	resp, err := l.svcCtx.PaymentRpc.TradeCheck(l.ctx, &paymentservice.TradeCheckReq{
		UserId:    authctx.UserID(l.ctx),
		OrderNo:   req.OrderNo,
		PaymentNo: req.PaymentNo,
		Channel:   req.Channel,
	})
	if err != nil {
		return nil, err
	}

	return &types.PaymentTradeCheckResp{
		Success:           resp.Success,
		Paid:              resp.Paid,
		Payment:           mapPaymentInfo(resp.Payment),
		OrderStatus:       resp.OrderStatus,
		PaidAt:            resp.PaidAt,
		CheckoutSessionId: resp.CheckoutSessionId,
	}, nil
}
