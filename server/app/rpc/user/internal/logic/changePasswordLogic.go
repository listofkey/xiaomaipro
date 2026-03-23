package logic

import (
	"context"
	"errors"

	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ChangePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChangePasswordLogic) ChangePassword(in *userpb.ChangePasswordReq) (*userpb.ChangePasswordResp, error) {
	if in.UserId <= 0 {
		return nil, errors.New("用户ID无效")
	}
	if in.OldPassword == "" || in.NewPassword == "" {
		return nil, errors.New("密码不能为空")
	}
	if len(in.NewPassword) < 6 {
		return nil, errors.New("新密码长度不能少于6位")
	}

	q := l.svcCtx.Query

	u, err := q.User.WithContext(l.ctx).Where(q.User.ID.Eq(in.UserId)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.OldPassword)); err != nil {
		return nil, errors.New("原密码错误")
	}

	// 生成新密码 Hash
	newHash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	if _, err := q.User.WithContext(l.ctx).
		Where(q.User.ID.Eq(in.UserId)).
		Update(q.User.PasswordHash, string(newHash)); err != nil {
		return nil, errors.New("密码修改失败")
	}

	return &userpb.ChangePasswordResp{Success: true}, nil
}
