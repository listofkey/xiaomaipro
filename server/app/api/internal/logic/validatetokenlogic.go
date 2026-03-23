package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewValidateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateTokenLogic {
	return &ValidateTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ValidateTokenLogic) ValidateToken() (*types.ValidateTokenResp, error) {
	userID := authctx.UserID(l.ctx)
	status := authctx.UserStatus(l.ctx)

	getUserInfoResp, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &userservice.GetUserInfoReq{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}

	return &types.ValidateTokenResp{
		Valid:    true,
		UserId:   formatID(userID),
		Status:   status,
		UserInfo: mapUserInfo(getUserInfoResp.UserInfo),
	}, nil
}
