package logic

import (
	"context"

	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTicketBuyerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTicketBuyerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTicketBuyerLogic {
	return &ListTicketBuyerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListTicketBuyerLogic) ListTicketBuyer(in *userpb.ListTicketBuyerReq) (*userpb.ListTicketBuyerResp, error) {
	q := l.svcCtx.Query
	cfg := l.svcCtx.Config

	buyers, err := q.TicketBuyer.WithContext(l.ctx).
		Where(q.TicketBuyer.UserID.Eq(in.UserId)).
		Order(q.TicketBuyer.IsDefault.Desc(), q.TicketBuyer.CreatedAt.Desc()).
		Find()
	if err != nil {
		return nil, err
	}

	result := make([]*userpb.TicketBuyerInfo, 0, len(buyers))
	for _, tb := range buyers {
		result = append(result, modelTicketBuyerToInfo(tb, cfg.AES.Key))
	}

	return &userpb.ListTicketBuyerResp{TicketBuyers: result}, nil
}
