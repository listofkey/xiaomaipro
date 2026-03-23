package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTicketBuyerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTicketBuyerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTicketBuyerLogic {
	return &UpdateTicketBuyerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTicketBuyerLogic) UpdateTicketBuyer(req *types.UpdateTicketBuyerReq) (resp *types.TicketBuyerResp, err error) {
	buyerID, err := parseID(req.BuyerId, "buyerId")
	if err != nil {
		return nil, err
	}

	updateResp, err := l.svcCtx.UserRpc.UpdateTicketBuyer(l.ctx, &userservice.UpdateTicketBuyerReq{
		BuyerId:   buyerID,
		UserId:    authctx.UserID(l.ctx),
		Name:      req.Name,
		IdCard:    req.IdCard,
		Phone:     req.Phone,
		IsDefault: boolToInt32(req.IsDefault),
	})
	if err != nil {
		return nil, err
	}

	return &types.TicketBuyerResp{
		TicketBuyer: mapTicketBuyerInfo(updateResp.TicketBuyer),
	}, nil
}
