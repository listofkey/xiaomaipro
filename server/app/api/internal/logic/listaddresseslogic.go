package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAddressesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAddressesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAddressesLogic {
	return &ListAddressesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAddressesLogic) ListAddresses() (resp *types.AddressListResp, err error) {
	listResp, err := l.svcCtx.UserRpc.ListAddress(l.ctx, &userservice.ListAddressReq{
		UserId: authctx.UserID(l.ctx),
	})
	if err != nil {
		return nil, err
	}

	return &types.AddressListResp{
		Addresses: mapAddressList(listResp.Addresses),
	}, nil
}
