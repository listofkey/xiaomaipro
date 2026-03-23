package logic

import (
	"context"
	"errors"

	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteTicketBuyerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTicketBuyerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTicketBuyerLogic {
	return &DeleteTicketBuyerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteTicketBuyerLogic) DeleteTicketBuyer(in *userpb.DeleteTicketBuyerReq) (*userpb.DeleteTicketBuyerResp, error) {
	if in.BuyerId <= 0 || in.UserId <= 0 {
		return nil, errors.New("参数无效")
	}

	q := l.svcCtx.Query

	tb, err := q.TicketBuyer.WithContext(l.ctx).
		Where(q.TicketBuyer.ID.Eq(in.BuyerId), q.TicketBuyer.UserID.Eq(in.UserId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("购票人不存在")
		}
		return nil, err
	}

	if _, err := q.TicketBuyer.WithContext(l.ctx).Delete(tb); err != nil {
		return nil, errors.New("删除购票人失败")
	}

	return &userpb.DeleteTicketBuyerResp{Success: true}, nil
}
