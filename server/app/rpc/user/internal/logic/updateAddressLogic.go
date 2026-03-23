package logic

import (
	"context"
	"errors"
	"time"

	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAddressLogic {
	return &UpdateAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateAddressLogic) UpdateAddress(in *userpb.UpdateAddressReq) (*userpb.UpdateAddressResp, error) {
	if in.AddressId <= 0 || in.UserId <= 0 {
		return nil, errors.New("参数无效")
	}

	q := l.svcCtx.Query

	// 确认地址归属
	addr, err := q.Address.WithContext(l.ctx).
		Where(q.Address.ID.Eq(in.AddressId), q.Address.UserID.Eq(in.UserId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("地址不存在")
		}
		return nil, err
	}

	// 若设为默认，先取消旧的默认
	if in.IsDefault == 1 {
		q.Address.WithContext(l.ctx).
			Where(q.Address.UserID.Eq(in.UserId), q.Address.IsDefault.Eq(1)).
			Update(q.Address.IsDefault, 0)
	}

	updates := map[string]interface{}{"updated_at": time.Now()}
	if in.ReceiverName != "" {
		updates["receiver_name"] = in.ReceiverName
	}
	if in.ReceiverPhone != "" {
		updates["receiver_phone"] = in.ReceiverPhone
	}
	if in.Province != "" {
		updates["province"] = in.Province
	}
	if in.City != "" {
		updates["city"] = in.City
	}
	if in.District != "" {
		updates["district"] = in.District
	}
	if in.Detail != "" {
		updates["detail"] = in.Detail
	}
	updates["is_default"] = in.IsDefault

	if _, err := q.Address.WithContext(l.ctx).
		Where(q.Address.ID.Eq(in.AddressId)).
		Updates(updates); err != nil {
		return nil, errors.New("更新地址失败")
	}

	// 重新查询
	addr, _ = q.Address.WithContext(l.ctx).Where(q.Address.ID.Eq(in.AddressId)).First()

	return &userpb.UpdateAddressResp{
		Address: modelAddressToInfo(addr),
	}, nil
}
