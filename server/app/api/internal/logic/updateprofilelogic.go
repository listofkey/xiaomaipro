package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProfileLogic) UpdateProfile(req *types.UpdateProfileReq) (resp *types.UserInfo, err error) {
	updateResp, err := l.svcCtx.UserRpc.UpdateUserInfo(l.ctx, &userservice.UpdateUserInfoReq{
		UserId:   authctx.UserID(l.ctx),
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Email:    req.Email,
	})
	if err != nil {
		return nil, err
	}

	userInfo := mapUserInfo(updateResp.UserInfo)
	return &userInfo, nil
}
