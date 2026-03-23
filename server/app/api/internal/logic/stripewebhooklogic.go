package logic

import (
	"context"

	"server/app/api/internal/svc"
	"server/app/rpc/payment/paymentservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type StripeWebhookLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStripeWebhookLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StripeWebhookLogic {
	return &StripeWebhookLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StripeWebhookLogic) StripeWebhook(rawBody []byte, signature string, headers map[string]string) (string, error) {
	resp, err := l.svcCtx.PaymentRpc.Notify(l.ctx, &paymentservice.NotifyReq{
		Channel:   "stripe",
		RawBody:   rawBody,
		Headers:   headers,
		Signature: signature,
	})
	if err != nil {
		return "", err
	}
	return resp.AckText, nil
}
