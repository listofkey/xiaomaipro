package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/order/orderservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderReq) (*types.CreateOrderResp, error) {
	eventID, err := parseID(req.EventId, "eventId")
	if err != nil {
		return nil, err
	}
	ticketTierID, err := parseID(req.TicketTierId, "ticketTierId")
	if err != nil {
		return nil, err
	}
	addressID, err := parseOptionalID(req.AddressId, "addressId")
	if err != nil {
		return nil, err
	}
	buyerIDs, err := parseIDs(req.TicketBuyerIds, "ticketBuyerIds")
	if err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.OrderRpc.CreateOrder(l.ctx, &orderservice.CreateOrderReq{
		UserId:         authctx.UserID(l.ctx),
		EventId:        eventID,
		TicketTierId:   ticketTierID,
		Quantity:       req.Quantity,
		TicketBuyerIds: buyerIDs,
		AddressId:      addressID,
		PayMethod:      req.PayMethod,
		RequestId:      req.RequestId,
	})
	if err != nil {
		return nil, err
	}

	return &types.CreateOrderResp{
		OrderNo:     resp.OrderNo,
		QueueToken:  resp.QueueToken,
		QueueStatus: resp.QueueStatus,
		Message:     resp.Message,
	}, nil
}
