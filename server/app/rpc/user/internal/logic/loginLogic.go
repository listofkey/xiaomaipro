package logic

import (
	"context"
	"errors"

	"server/app/rpc/user/internal/pkg/jwt"
	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *userpb.LoginReq) (*userpb.LoginResp, error) {
	if in.Account == "" {
		return nil, errors.New("账号不能为空")
	}
	if in.Password == "" {
		return nil, errors.New("密码不能为空")
	}

	q := l.svcCtx.Query
	cfg := l.svcCtx.Config

	// 支持手机号或邮箱登录
	u, err := q.User.WithContext(l.ctx).
		Where(q.User.Phone.Eq(in.Account)).
		Or(q.User.Email.Eq(in.Account)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("账号不存在")
		}
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)); err != nil {
		return nil, errors.New("密码错误")
	}

	// 检查账号状态
	if u.Status != 1 {
		return nil, errors.New("账号已被禁用，请联系客服")
	}

	// 生成 JWT
	accessToken, err := jwt.GenerateAccessToken(u.ID, cfg.JWT.AccessSecret, cfg.JWT.AccessExpire)
	if err != nil {
		return nil, errors.New("Token 生成失败")
	}
	refreshToken, err := jwt.GenerateRefreshToken(u.ID, cfg.JWT.AccessSecret, cfg.JWT.RefreshExpire)
	if err != nil {
		return nil, errors.New("Refresh Token 生成失败")
	}

	return &userpb.LoginResp{
		UserId:       u.ID,
		UserInfo:     modelUserToInfo(u, cfg.AES.Key),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    cfg.JWT.AccessExpire,
	}, nil
}
