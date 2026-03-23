package logic

import (
	"context"
	"errors"
	"regexp"
	"time"

	"server/app/rpc/model"
	"server/app/rpc/user/internal/pkg/jwt"
	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"
	"server/common"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *userpb.RegisterReq) (*userpb.RegisterResp, error) {
	// 参数校验
	if in.Phone == "" && in.Email == "" {
		return nil, errors.New("手机号或邮箱不能为空")
	}
	if in.Password == "" {
		return nil, errors.New("密码不能为空")
	}

	q := l.svcCtx.Query
	cfg := l.svcCtx.Config

	// 检查手机号/邮箱是否已注册
	if in.Phone != "" {
		if !isValidPhone(in.Phone) {
			return nil, errors.New("手机号格式不正确")
		}
		existing, err := q.User.WithContext(l.ctx).Where(q.User.Phone.Eq(in.Phone)).First()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil {
			return nil, errors.New("该手机号已被注册")
		}
	}
	if in.Email != "" {
		if !isValidEmail(in.Email) {
			return nil, errors.New("邮箱格式不正确")
		}
		existing, err := q.User.WithContext(l.ctx).Where(q.User.Email.Eq(in.Email)).First()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil {
			return nil, errors.New("该邮箱已被注册")
		}
	}

	// 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	// 设置默认昵称
	nickname := in.Nickname
	if nickname == "" {
		if in.Phone != "" {
			nickname = maskPhoneForNickname(in.Phone)
		} else {
			nickname = emailPrefix(in.Email)
		}
	}

	// 创建用户
	user := &model.User{
		ID:           common.GenerateId(),
		Phone:        in.Phone,
		Email:        in.Email,
		PasswordHash: string(hash),
		Nickname:     nickname,
		Status:       1,
		IsVerified:   0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := q.User.WithContext(l.ctx).Create(user); err != nil {
		return nil, errors.New("注册失败: " + err.Error())
	}

	// 生成 JWT
	accessToken, err := jwt.GenerateAccessToken(user.ID, cfg.JWT.AccessSecret, cfg.JWT.AccessExpire)
	if err != nil {
		return nil, errors.New("Token 生成失败")
	}
	refreshToken, err := jwt.GenerateRefreshToken(user.ID, cfg.JWT.AccessSecret, cfg.JWT.RefreshExpire)
	if err != nil {
		return nil, errors.New("Refresh Token 生成失败")
	}

	return &userpb.RegisterResp{
		UserId:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    cfg.JWT.AccessExpire,
	}, nil
}

// ---------- 辅助函数 ----------

var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isValidPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func maskPhoneForNickname(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

func emailPrefix(email string) string {
	for i, c := range email {
		if c == '@' {
			return email[:i]
		}
	}
	return email
}
