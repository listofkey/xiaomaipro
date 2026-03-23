package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAddressLogic {
	return &UpdateAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAddressLogic) UpdateAddress(req *types.UpdateAddressReq) (resp *types.AddressResp, err error) {
	addressID, err := parseID(req.AddressId, "addressId")
	if err != nil {
		return nil, err
	}

	updateResp, err := l.svcCtx.UserRpc.UpdateAddress(l.ctx, &userservice.UpdateAddressReq{
		AddressId:     addressID,
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
		Address: mapAddressInfo(updateResp.Address),
	}, nil
}
