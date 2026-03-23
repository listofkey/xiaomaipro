package logic

import (
	"context"
	"errors"
	"time"

	"server/app/rpc/model"
	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"
	"server/common"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAddressLogic {
	return &CreateAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateAddressLogic) CreateAddress(in *userpb.CreateAddressReq) (*userpb.CreateAddressResp, error) {
	if in.UserId <= 0 {
		return nil, errors.New("用户ID无效")
	}
	if in.ReceiverName == "" || in.ReceiverPhone == "" {
		return nil, errors.New("收件人姓名和电话不能为空")
	}
	if in.Province == "" || in.City == "" || in.District == "" || in.Detail == "" {
		return nil, errors.New("地址信息不完整")
	}

	q := l.svcCtx.Query

	// 若设为默认，先取消旧的默认
	if in.IsDefault == 1 {
		q.Address.WithContext(l.ctx).
			Where(q.Address.UserID.Eq(in.UserId), q.Address.IsDefault.Eq(1)).
			Update(q.Address.IsDefault, 0)
	}

	addr := &model.Address{
		ID:            common.GenerateId(),
		UserID:        in.UserId,
		ReceiverName:  in.ReceiverName,
		ReceiverPhone: in.ReceiverPhone,
		Province:      in.Province,
		City:          in.City,
		District:      in.District,
		Detail:        in.Detail,
		IsDefault:     int16(in.IsDefault),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := q.Address.WithContext(l.ctx).Create(addr); err != nil {
		return nil, errors.New("新增地址失败: " + err.Error())
	}

	return &userpb.CreateAddressResp{
		Address: modelAddressToInfo(addr),
	}, nil
}
