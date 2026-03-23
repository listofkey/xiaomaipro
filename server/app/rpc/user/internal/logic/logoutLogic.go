package logic

import (
	"context"
	"errors"

	"server/app/rpc/user/internal/pkg/jwt"
	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"
	"server/common/auth"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LogoutLogic) Logout(in *userpb.LogoutReq) (*userpb.LogoutResp, error) {
	if in.AccessToken == "" {
		return &userpb.LogoutResp{Success: true}, nil
	}

	claims, err := jwt.ParseTokenAllowExpired(in.AccessToken, l.svcCtx.Config.JWT.AccessSecret)
	if err != nil {
		return nil, errors.New("token 无效")
	}
	if claims.Type != "access" {
		return nil, errors.New("token 类型错误")
	}
	if in.UserId != 0 && in.UserId != claims.UserID.Int64() {
		return nil, errors.New("用户信息不匹配")
	}

	if err := auth.BlacklistToken(l.ctx, l.svcCtx.Redis, in.AccessToken, jwt.RemainingTTL(claims)); err != nil {
		return nil, err
	}

	l.Infof("user %d logout", claims.UserID.Int64())

	return &userpb.LogoutResp{Success: true}, nil
}
