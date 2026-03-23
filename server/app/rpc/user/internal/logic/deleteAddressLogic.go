package logic

import (
	"context"
	"errors"

	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAddressLogic {
	return &DeleteAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteAddressLogic) DeleteAddress(in *userpb.DeleteAddressReq) (*userpb.DeleteAddressResp, error) {
	if in.AddressId <= 0 || in.UserId <= 0 {
		return nil, errors.New("参数无效")
	}

	q := l.svcCtx.Query

	// 确认归属后删除
	addr, err := q.Address.WithContext(l.ctx).
		Where(q.Address.ID.Eq(in.AddressId), q.Address.UserID.Eq(in.UserId)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("地址不存在")
		}
		return nil, err
	}

	if _, err := q.Address.WithContext(l.ctx).Delete(addr); err != nil {
		return nil, errors.New("删除地址失败")
	}

	return &userpb.DeleteAddressResp{Success: true}, nil
}
