package orderservice

import (
	"context"

	"server/app/rpc/order/orderpb"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	ApplyRefundReq         = orderpb.ApplyRefundReq
	ApplyRefundResp        = orderpb.ApplyRefundResp
	CancelOrderReq         = orderpb.CancelOrderReq
	CancelOrderResp        = orderpb.CancelOrderResp
	CreateOrderReq         = orderpb.CreateOrderReq
	CreateOrderResp        = orderpb.CreateOrderResp
	GetOrderDetailReq      = orderpb.GetOrderDetailReq
	GetOrderDetailResp     = orderpb.GetOrderDetailResp
	GetQueueStatusReq      = orderpb.GetQueueStatusReq
	GetQueueStatusResp     = orderpb.GetQueueStatusResp
	InventoryPreheatResult = orderpb.InventoryPreheatResult
	ListOrderReq           = orderpb.ListOrderReq
	ListOrderResp          = orderpb.ListOrderResp
	OrderBuyerInfo         = orderpb.OrderBuyerInfo
	OrderDeliveryInfo      = orderpb.OrderDeliveryInfo
	OrderDetail            = orderpb.OrderDetail
	OrderPaymentInfo       = orderpb.OrderPaymentInfo
	OrderRefundInfo        = orderpb.OrderRefundInfo
	OrderSummary           = orderpb.OrderSummary
	OrderTicketInfo        = orderpb.OrderTicketInfo
	PayOrderReq            = orderpb.PayOrderReq
	PayOrderResp           = orderpb.PayOrderResp
	PreheatInventoryReq    = orderpb.PreheatInventoryReq
	PreheatInventoryResp   = orderpb.PreheatInventoryResp

	OrderService interface {
		PreheatInventory(ctx context.Context, in *PreheatInventoryReq, opts ...grpc.CallOption) (*PreheatInventoryResp, error)
		CreateOrder(ctx context.Context, in *CreateOrderReq, opts ...grpc.CallOption) (*CreateOrderResp, error)
		GetQueueStatus(ctx context.Context, in *GetQueueStatusReq, opts ...grpc.CallOption) (*GetQueueStatusResp, error)
		PayOrder(ctx context.Context, in *PayOrderReq, opts ...grpc.CallOption) (*PayOrderResp, error)
		CancelOrder(ctx context.Context, in *CancelOrderReq, opts ...grpc.CallOption) (*CancelOrderResp, error)
		ApplyRefund(ctx context.Context, in *ApplyRefundReq, opts ...grpc.CallOption) (*ApplyRefundResp, error)
		ListOrder(ctx context.Context, in *ListOrderReq, opts ...grpc.CallOption) (*ListOrderResp, error)
		GetOrderDetail(ctx context.Context, in *GetOrderDetailReq, opts ...grpc.CallOption) (*GetOrderDetailResp, error)
	}

	defaultOrderService struct {
		cli zrpc.Client
	}
)

func NewOrderService(cli zrpc.Client) OrderService {
	return &defaultOrderService{cli: cli}
}

func (m *defaultOrderService) PreheatInventory(ctx context.Context, in *PreheatInventoryReq, opts ...grpc.CallOption) (*PreheatInventoryResp, error) {
	return orderpb.NewOrderServiceClient(m.cli.Conn()).PreheatInventory(ctx, in, opts...)
}

func (m *defaultOrderService) CreateOrder(ctx context.Context, in *CreateOrderReq, opts ...grpc.CallOption) (*CreateOrderResp, error) {
	return orderpb.NewOrderServiceClient(m.cli.Conn()).CreateOrder(ctx, in, opts...)
}

func (m *defaultOrderService) GetQueueStatus(ctx context.Context, in *GetQueueStatusReq, opts ...grpc.CallOption) (*GetQueueStatusResp, error) {
	return orderpb.NewOrderServiceClient(m.cli.Conn()).GetQueueStatus(ctx, in, opts...)
}

func (m *defaultOrderService) PayOrder(ctx context.Context, in *PayOrderReq, opts ...grpc.CallOption) (*PayOrderResp, error) {
	return orderpb.NewOrderServiceClient(m.cli.Conn()).PayOrder(ctx, in, opts...)
}

func (m *defaultOrderService) CancelOrder(ctx context.Context, in *CancelOrderReq, opts ...grpc.CallOption) (*CancelOrderResp, error) {
	return orderpb.NewOrderServiceClient(m.cli.Conn()).CancelOrder(ctx, in, opts...)
}

func (m *defaultOrderService) ApplyRefund(ctx context.Context, in *ApplyRefundReq, opts ...grpc.CallOption) (*ApplyRefundResp, error) {
	return orderpb.NewOrderServiceClient(m.cli.Conn()).ApplyRefund(ctx, in, opts...)
}

func (m *defaultOrderService) ListOrder(ctx context.Context, in *ListOrderReq, opts ...grpc.CallOption) (*ListOrderResp, error) {
	return orderpb.NewOrderServiceClient(m.cli.Conn()).ListOrder(ctx, in, opts...)
}

func (m *defaultOrderService) GetOrderDetail(ctx context.Context, in *GetOrderDetailReq, opts ...grpc.CallOption) (*GetOrderDetailResp, error) {
	return orderpb.NewOrderServiceClient(m.cli.Conn()).GetOrderDetail(ctx, in, opts...)
}
