package logic

import (
	"context"
	"errors"

	"server/app/rpc/user/internal/pkg/jwt"
	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"
	"server/common/auth"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type RefreshTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefreshTokenLogic) RefreshToken(in *userpb.RefreshTokenReq) (*userpb.RefreshTokenResp, error) {
	if in.RefreshToken == "" {
		return nil, errors.New("refresh_token 不能为空")
	}

	blacklisted, err := auth.IsTokenBlacklisted(l.ctx, l.svcCtx.Redis, in.RefreshToken)
	if err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, errors.New("refresh_token 已失效")
	}

	cfg := l.svcCtx.Config
	q := l.svcCtx.Query

	claims, err := jwt.ParseToken(in.RefreshToken, cfg.JWT.AccessSecret)
	if err != nil {
		return nil, errors.New("refresh_token 无效或已过期")
	}
	if claims.Type != "refresh" {
		return nil, errors.New("token 类型错误")
	}

	user, err := q.User.WithContext(l.ctx).Where(q.User.ID.Eq(claims.UserID.Int64())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	newAccessToken, err := jwt.GenerateAccessToken(claims.UserID.Int64(), cfg.JWT.AccessSecret, cfg.JWT.AccessExpire)
	if err != nil {
		return nil, errors.New("token 生成失败")
	}
	newRefreshToken, err := jwt.GenerateRefreshToken(claims.UserID.Int64(), cfg.JWT.AccessSecret, cfg.JWT.RefreshExpire)
	if err != nil {
		return nil, errors.New("refresh token 生成失败")
	}

	if err := auth.BlacklistToken(l.ctx, l.svcCtx.Redis, in.RefreshToken, jwt.RemainingTTL(claims)); err != nil {
		return nil, err
	}

	return &userpb.RefreshTokenResp{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    cfg.JWT.AccessExpire,
	}, nil
}
