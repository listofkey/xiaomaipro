package paymentservice

import (
	"context"
	"server/app/rpc/payment/paymentpb"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	DetailReq      = paymentpb.DetailReq
	DetailResp     = paymentpb.DetailResp
	NotifyReq      = paymentpb.NotifyReq
	NotifyResp     = paymentpb.NotifyResp
	PayOrderReq    = paymentpb.PayOrderReq
	PayOrderResp   = paymentpb.PayOrderResp
	PaymentInfo    = paymentpb.PaymentInfo
	RefundInfo     = paymentpb.RefundInfo
	RefundReq      = paymentpb.RefundReq
	RefundResp     = paymentpb.RefundResp
	TradeCheckReq  = paymentpb.TradeCheckReq
	TradeCheckResp = paymentpb.TradeCheckResp
	PaymentService interface {
		PayOrder(ctx context.Context, in *PayOrderReq, opts ...grpc.CallOption) (*PayOrderResp, error)
		Notify(ctx context.Context, in *NotifyReq, opts ...grpc.CallOption) (*NotifyResp, error)
		TradeCheck(ctx context.Context, in *TradeCheckReq, opts ...grpc.CallOption) (*TradeCheckResp, error)
		Refund(ctx context.Context, in *RefundReq, opts ...grpc.CallOption) (*RefundResp, error)
		Detail(ctx context.Context, in *DetailReq, opts ...grpc.CallOption) (*DetailResp, error)
	}

	defaultPaymentService struct {
		cli zrpc.Client
	}
)

func NewPaymentService(cli zrpc.Client) PaymentService {
	return &defaultPaymentService{cli: cli}
}

func (m *defaultPaymentService) PayOrder(ctx context.Context, in *PayOrderReq, opts ...grpc.CallOption) (*PayOrderResp, error) {
	return paymentpb.NewPaymentServiceClient(m.cli.Conn()).PayOrder(ctx, in, opts...)
}

func (m *defaultPaymentService) Notify(ctx context.Context, in *NotifyReq, opts ...grpc.CallOption) (*NotifyResp, error) {
	return paymentpb.NewPaymentServiceClient(m.cli.Conn()).Notify(ctx, in, opts...)
}

func (m *defaultPaymentService) TradeCheck(ctx context.Context, in *TradeCheckReq, opts ...grpc.CallOption) (*TradeCheckResp, error) {
	return paymentpb.NewPaymentServiceClient(m.cli.Conn()).TradeCheck(ctx, in, opts...)
}

func (m *defaultPaymentService) Refund(ctx context.Context, in *RefundReq, opts ...grpc.CallOption) (*RefundResp, error) {
	return paymentpb.NewPaymentServiceClient(m.cli.Conn()).Refund(ctx, in, opts...)
}

func (m *defaultPaymentService) Detail(ctx context.Context, in *DetailReq, opts ...grpc.CallOption) (*DetailResp, error) {
	return paymentpb.NewPaymentServiceClient(m.cli.Conn()).Detail(ctx, in, opts...)
}
