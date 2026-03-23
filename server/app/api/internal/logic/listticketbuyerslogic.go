package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTicketBuyersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTicketBuyersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTicketBuyersLogic {
	return &ListTicketBuyersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTicketBuyersLogic) ListTicketBuyers() (resp *types.TicketBuyerListResp, err error) {
	listResp, err := l.svcCtx.UserRpc.ListTicketBuyer(l.ctx, &userservice.ListTicketBuyerReq{
		UserId: authctx.UserID(l.ctx),
	})
	if err != nil {
		return nil, err
	}

	return &types.TicketBuyerListResp{
		TicketBuyers: mapTicketBuyerList(listResp.TicketBuyers),
	}, nil
}
