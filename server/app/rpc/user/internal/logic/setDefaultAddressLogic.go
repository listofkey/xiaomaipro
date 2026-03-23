package logic

import (
	"context"
	"errors"

	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type SetDefaultAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetDefaultAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultAddressLogic {
	return &SetDefaultAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetDefaultAddressLogic) SetDefaultAddress(in *userpb.SetDefaultAddressReq) (*userpb.SetDefaultAddressResp, error) {
	if in.AddressId <= 0 || in.UserId <= 0 {
		return nil, errors.New("参数无效")
	}

	q := l.svcCtx.Query

	// 确认地址归属
	_, err := q.Address.WithContext(l.ctx).
		Where(q.Address.ID.Eq(in.AddressId), q.Address.UserID.Eq(in.UserId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("地址不存在")
		}
		return nil, err
	}

	// 事务：先全部取消默认，再设置新默认
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&struct{ IsDefault int32 }{}).
			Table("address").
			Where("user_id = ?", in.UserId).
			Update("is_default", 0).Error; err != nil {
			return err
		}
		return tx.Table("address").
			Where("id = ?", in.AddressId).
			Update("is_default", 1).Error
	})
	if err != nil {
		return nil, errors.New("设置默认地址失败")
	}

	return &userpb.SetDefaultAddressResp{Success: true}, nil
}
