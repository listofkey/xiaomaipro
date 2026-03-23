package logic

import (
	"context"

	"server/app/api/internal/svc"
	"server/app/rpc/payment/paymentservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type AlipayNotifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAlipayNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AlipayNotifyLogic {
	return &AlipayNotifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AlipayNotifyLogic) AlipayNotify(params map[string]string) (string, error) {
	resp, err := l.svcCtx.PaymentRpc.Notify(l.ctx, &paymentservice.NotifyReq{
		Channel: "alipay",
		Params:  params,
	})
	if err != nil {
		return "", err
	}
	return resp.AckText, nil
}
