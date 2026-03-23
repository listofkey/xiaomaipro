package logic

import (
	"context"
	"errors"
	"time"

	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserInfoLogic {
	return &UpdateUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserInfoLogic) UpdateUserInfo(in *userpb.UpdateUserInfoReq) (*userpb.UpdateUserInfoResp, error) {
	if in.UserId <= 0 {
		return nil, errors.New("用户ID无效")
	}

	q := l.svcCtx.Query
	cfg := l.svcCtx.Config

	u, err := q.User.WithContext(l.ctx).Where(q.User.ID.Eq(in.UserId)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 更新非空字段
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if in.Nickname != "" {
		updates["nickname"] = in.Nickname
	}
	if in.Avatar != "" {
		updates["avatar"] = in.Avatar
	}
	if in.Email != "" {
		if !isValidEmail(in.Email) {
			return nil, errors.New("邮箱格式不正确")
		}
		// 检查邮箱是否已被占用
		existing, err := q.User.WithContext(l.ctx).
			Where(q.User.Email.Eq(in.Email)).
			Where(q.User.ID.Neq(in.UserId)).
			First()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil {
			return nil, errors.New("该邮箱已被其他账号使用")
		}
		updates["email"] = in.Email
	}

	if _, err := q.User.WithContext(l.ctx).
		Where(q.User.ID.Eq(in.UserId)).
		Updates(updates); err != nil {
		return nil, errors.New("更新失败: " + err.Error())
	}

	// 重新查询最新用户信息
	u, _ = q.User.WithContext(l.ctx).Where(q.User.ID.Eq(in.UserId)).First()

	return &userpb.UpdateUserInfoResp{
		UserInfo: modelUserToInfo(u, cfg.AES.Key),
	}, nil
}
