package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/payment/paymentservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type PaymentDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPaymentDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PaymentDetailLogic {
	return &PaymentDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PaymentDetailLogic) PaymentDetail(req *types.PaymentDetailReq) (*types.PaymentDetailResp, error) {
	resp, err := l.svcCtx.PaymentRpc.Detail(l.ctx, &paymentservice.DetailReq{
		UserId:  authctx.UserID(l.ctx),
		OrderNo: req.OrderNo,
	})
	if err != nil {
		return nil, err
	}

	return &types.PaymentDetailResp{
		Payment:     mapPaymentInfo(resp.Payment),
		Refund:      mapPaymentRefund(resp.Refund),
		OrderStatus: resp.OrderStatus,
		PaidAt:      resp.PaidAt,
	}, nil
}
