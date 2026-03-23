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

type ValidateTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateTokenLogic {
	return &ValidateTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ValidateTokenLogic) ValidateToken(in *userpb.ValidateTokenReq) (*userpb.ValidateTokenResp, error) {
	if in.AccessToken == "" {
		return &userpb.ValidateTokenResp{Valid: false}, nil
	}

	blacklisted, err := auth.IsTokenBlacklisted(l.ctx, l.svcCtx.Redis, in.AccessToken)
	if err != nil {
		return nil, err
	}
	if blacklisted {
		return &userpb.ValidateTokenResp{Valid: false}, nil
	}

	cfg := l.svcCtx.Config
	q := l.svcCtx.Query

	claims, err := jwt.ParseToken(in.AccessToken, cfg.JWT.AccessSecret)
	if err != nil {
		return &userpb.ValidateTokenResp{Valid: false}, nil
	}
	if claims.Type != "access" {
		return &userpb.ValidateTokenResp{Valid: false}, nil
	}

	user, err := q.User.WithContext(l.ctx).Where(q.User.ID.Eq(claims.UserID.Int64())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &userpb.ValidateTokenResp{Valid: false}, nil
		}
		return nil, err
	}

	return &userpb.ValidateTokenResp{
		Valid:  user.Status == 1,
		UserId: user.ID,
		Status: int32(user.Status),
	}, nil
}
