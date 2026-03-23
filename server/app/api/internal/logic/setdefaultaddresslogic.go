package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetDefaultAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetDefaultAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultAddressLogic {
	return &SetDefaultAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetDefaultAddressLogic) SetDefaultAddress(req *types.SetDefaultAddressReq) (resp *types.OperationResp, err error) {
	addressID, err := parseID(req.AddressId, "addressId")
	if err != nil {
		return nil, err
	}

	setResp, err := l.svcCtx.UserRpc.SetDefaultAddress(l.ctx, &userservice.SetDefaultAddressReq{
		AddressId: addressID,
		UserId:    authctx.UserID(l.ctx),
	})
	if err != nil {
		return nil, err
	}

	return &types.OperationResp{
		Success: setResp.Success,
	}, nil
}
