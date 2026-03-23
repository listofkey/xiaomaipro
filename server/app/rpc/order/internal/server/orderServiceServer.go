package server

import (
	"context"

	"server/app/rpc/order/internal/logic"
	"server/app/rpc/order/internal/svc"
	"server/app/rpc/order/orderpb"
)

type OrderServiceServer struct {
	svcCtx *svc.ServiceContext
	orderpb.UnimplementedOrderServiceServer
}

func NewOrderServiceServer(svcCtx *svc.ServiceContext) *OrderServiceServer {
	return &OrderServiceServer{svcCtx: svcCtx}
}

func (s *OrderServiceServer) PreheatInventory(ctx context.Context, in *orderpb.PreheatInventoryReq) (*orderpb.PreheatInventoryResp, error) {
	return logic.NewOrderCore(ctx, s.svcCtx).PreheatInventory(in)
}

func (s *OrderServiceServer) CreateOrder(ctx context.Context, in *orderpb.CreateOrderReq) (*orderpb.CreateOrderResp, error) {
	return logic.NewOrderCore(ctx, s.svcCtx).CreateOrder(in)
}

func (s *OrderServiceServer) GetQueueStatus(ctx context.Context, in *orderpb.GetQueueStatusReq) (*orderpb.GetQueueStatusResp, error) {
	return logic.NewOrderCore(ctx, s.svcCtx).GetQueueStatus(in)
}

func (s *OrderServiceServer) PayOrder(ctx context.Context, in *orderpb.PayOrderReq) (*orderpb.PayOrderResp, error) {
	return logic.NewOrderCore(ctx, s.svcCtx).PayOrder(in)
}

func (s *OrderServiceServer) CancelOrder(ctx context.Context, in *orderpb.CancelOrderReq) (*orderpb.CancelOrderResp, error) {
	return logic.NewOrderCore(ctx, s.svcCtx).CancelOrder(in)
}

func (s *OrderServiceServer) ApplyRefund(ctx context.Context, in *orderpb.ApplyRefundReq) (*orderpb.ApplyRefundResp, error) {
	return logic.NewOrderCore(ctx, s.svcCtx).ApplyRefund(in)
}

func (s *OrderServiceServer) ListOrder(ctx context.Context, in *orderpb.ListOrderReq) (*orderpb.ListOrderResp, error) {
	return logic.NewOrderCore(ctx, s.svcCtx).ListOrder(in)
}

func (s *OrderServiceServer) GetOrderDetail(ctx context.Context, in *orderpb.GetOrderDetailReq) (*orderpb.GetOrderDetailResp, error) {
	return logic.NewOrderCore(ctx, s.svcCtx).GetOrderDetail(in)
}
