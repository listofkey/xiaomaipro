package logic

import (
	"context"
	"errors"

	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *userpb.GetUserInfoReq) (*userpb.GetUserInfoResp, error) {
	q := l.svcCtx.Query
	cfg := l.svcCtx.Config

	u, err := q.User.WithContext(l.ctx).Where(q.User.ID.Eq(in.UserId)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	return &userpb.GetUserInfoResp{
		UserInfo: modelUserToInfo(u, cfg.AES.Key),
	}, nil
}
