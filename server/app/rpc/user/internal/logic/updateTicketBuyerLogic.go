package logic

import (
	"context"
	"errors"
	"time"

	"server/app/rpc/user/internal/pkg/encrypt"
	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateTicketBuyerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTicketBuyerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTicketBuyerLogic {
	return &UpdateTicketBuyerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateTicketBuyerLogic) UpdateTicketBuyer(in *userpb.UpdateTicketBuyerReq) (*userpb.UpdateTicketBuyerResp, error) {
	if in.BuyerId <= 0 || in.UserId <= 0 {
		return nil, errors.New("参数无效")
	}

	q := l.svcCtx.Query
	cfg := l.svcCtx.Config

	// 确认归属
	tb, err := q.TicketBuyer.WithContext(l.ctx).
		Where(q.TicketBuyer.ID.Eq(in.BuyerId), q.TicketBuyer.UserID.Eq(in.UserId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("购票人不存在")
		}
		return nil, err
	}

	// 若设为默认，先取消旧的默认
	if in.IsDefault == 1 {
		q.TicketBuyer.WithContext(l.ctx).
			Where(q.TicketBuyer.UserID.Eq(in.UserId), q.TicketBuyer.IsDefault.Eq(1)).
			Update(q.TicketBuyer.IsDefault, 0)
	}

	updates := map[string]interface{}{
		"is_default": in.IsDefault,
		"updated_at": time.Now(),
	}
	if in.Name != "" {
		updates["name"] = in.Name
	}
	if in.IdCard != "" {
		if !idCardRegex.MatchString(in.IdCard) {
			return nil, errors.New("身份证号格式不正确")
		}
		encIDCard, err := encrypt.AESEncrypt(in.IdCard, cfg.AES.Key)
		if err != nil {
			return nil, errors.New("数据加密失败")
		}
		updates["id_card"] = encIDCard
	}
	if in.Phone != "" {
		updates["phone"] = in.Phone
	}

	if _, err := q.TicketBuyer.WithContext(l.ctx).
		Where(q.TicketBuyer.ID.Eq(in.BuyerId)).
		Updates(updates); err != nil {
		return nil, errors.New("更新购票人失败")
	}

	tb, _ = q.TicketBuyer.WithContext(l.ctx).Where(q.TicketBuyer.ID.Eq(in.BuyerId)).First()

	return &userpb.UpdateTicketBuyerResp{
		TicketBuyer: modelTicketBuyerToInfo(tb, cfg.AES.Key),
	}, nil
}
