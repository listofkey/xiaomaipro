package logic

import (
	"context"
	"errors"

	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type SetDefaultTicketBuyerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetDefaultTicketBuyerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultTicketBuyerLogic {
	return &SetDefaultTicketBuyerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetDefaultTicketBuyerLogic) SetDefaultTicketBuyer(in *userpb.SetDefaultTicketBuyerReq) (*userpb.SetDefaultTicketBuyerResp, error) {
	if in.BuyerId <= 0 || in.UserId <= 0 {
		return nil, errors.New("参数无效")
	}

	q := l.svcCtx.Query

	// 确认归属
	_, err := q.TicketBuyer.WithContext(l.ctx).
		Where(q.TicketBuyer.ID.Eq(in.BuyerId), q.TicketBuyer.UserID.Eq(in.UserId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("购票人不存在")
		}
		return nil, err
	}

	// 事务：先全部取消默认，再设置新默认
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("ticket_buyer").
			Where("user_id = ?", in.UserId).
			Update("is_default", 0).Error; err != nil {
			return err
		}
		return tx.Table("ticket_buyer").
			Where("id = ?", in.BuyerId).
			Update("is_default", 1).Error
	})
	if err != nil {
		return nil, errors.New("设置默认购票人失败")
	}

	return &userpb.SetDefaultTicketBuyerResp{Success: true}, nil
}
