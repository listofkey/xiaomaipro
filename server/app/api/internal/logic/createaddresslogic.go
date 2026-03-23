package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAddressLogic {
	return &CreateAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateAddressLogic) CreateAddress(req *types.CreateAddressReq) (resp *types.AddressResp, err error) {
	createResp, err := l.svcCtx.UserRpc.CreateAddress(l.ctx, &userservice.CreateAddressReq{
		UserId:        authctx.UserID(l.ctx),
		ReceiverName:  req.ReceiverName,
		ReceiverPhone: req.ReceiverPhone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		Detail:        req.Detail,
		IsDefault:     boolToInt32(req.IsDefault),
	})
	if err != nil {
		return nil, err
	}

	return &types.AddressResp{
		Address: mapAddressInfo(createResp.Address),
	}, nil
}
