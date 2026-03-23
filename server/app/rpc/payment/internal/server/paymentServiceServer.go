package server

import (
	"context"

	"server/app/rpc/payment/internal/logic"
	"server/app/rpc/payment/internal/svc"
	"server/app/rpc/payment/paymentpb"
)

type PaymentServiceServer struct {
	svcCtx *svc.ServiceContext
	paymentpb.UnimplementedPaymentServiceServer
}

func NewPaymentServiceServer(svcCtx *svc.ServiceContext) *PaymentServiceServer {
	return &PaymentServiceServer{svcCtx: svcCtx}
}

func (s *PaymentServiceServer) PayOrder(ctx context.Context, in *paymentpb.PayOrderReq) (*paymentpb.PayOrderResp, error) {
	l := logic.NewPaymentCore(ctx, s.svcCtx)
	return l.PayOrder(in)
}

func (s *PaymentServiceServer) Notify(ctx context.Context, in *paymentpb.NotifyReq) (*paymentpb.NotifyResp, error) {
	l := logic.NewPaymentCore(ctx, s.svcCtx)
	return l.Notify(in)
}

func (s *PaymentServiceServer) TradeCheck(ctx context.Context, in *paymentpb.TradeCheckReq) (*paymentpb.TradeCheckResp, error) {
	l := logic.NewPaymentCore(ctx, s.svcCtx)
	return l.TradeCheck(in)
}

func (s *PaymentServiceServer) Refund(ctx context.Context, in *paymentpb.RefundReq) (*paymentpb.RefundResp, error) {
	l := logic.NewPaymentCore(ctx, s.svcCtx)
	return l.Refund(in)
}

func (s *PaymentServiceServer) Detail(ctx context.Context, in *paymentpb.DetailReq) (*paymentpb.DetailResp, error) {
	l := logic.NewPaymentCore(ctx, s.svcCtx)
	return l.Detail(in)
}
