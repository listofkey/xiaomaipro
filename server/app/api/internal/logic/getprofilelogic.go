package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProfileLogic {
	return &GetProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProfileLogic) GetProfile() (resp *types.UserInfo, err error) {
	getUserInfoResp, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &userservice.GetUserInfoReq{
		UserId: authctx.UserID(l.ctx),
	})
	if err != nil {
		return nil, err
	}

	userInfo := mapUserInfo(getUserInfoResp.UserInfo)
	return &userInfo, nil
}
