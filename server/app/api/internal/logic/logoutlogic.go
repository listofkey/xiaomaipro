package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout(req *types.LogoutReq) (*types.OperationResp, error) {
	_ = req

	_, err := l.svcCtx.UserRpc.Logout(l.ctx, &userservice.LogoutReq{
		UserId:      authctx.UserID(l.ctx),
		AccessToken: authctx.AccessToken(l.ctx),
	})
	if err != nil {
		return nil, err
	}

	return &types.OperationResp{Success: true}, nil
}
