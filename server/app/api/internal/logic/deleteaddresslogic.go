package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAddressLogic {
	return &DeleteAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAddressLogic) DeleteAddress(req *types.DeleteAddressReq) (resp *types.OperationResp, err error) {
	addressID, err := parseID(req.AddressId, "addressId")
	if err != nil {
		return nil, err
	}

	deleteResp, err := l.svcCtx.UserRpc.DeleteAddress(l.ctx, &userservice.DeleteAddressReq{
		AddressId: addressID,
		UserId:    authctx.UserID(l.ctx),
	})
	if err != nil {
		return nil, err
	}

	return &types.OperationResp{
		Success: deleteResp.Success,
	}, nil
}
