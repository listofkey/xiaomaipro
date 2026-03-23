package logic

import (
	"context"

	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (*types.AuthPayload, error) {
	registerResp, err := l.svcCtx.UserRpc.Register(l.ctx, &userservice.RegisterReq{
		Phone:    req.Phone,
		Email:    req.Email,
		Password: req.Password,
		Nickname: req.Nickname,
		Code:     req.Code,
	})
	if err != nil {
		return nil, err
	}

	userInfo := types.UserInfo{
		Id:       formatID(registerResp.UserId),
		Phone:    req.Phone,
		Email:    req.Email,
		Nickname: req.Nickname,
	}

	getUserInfoResp, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &userservice.GetUserInfoReq{
		UserId: registerResp.UserId,
	})
	if err == nil {
		userInfo = mapUserInfo(getUserInfoResp.UserInfo)
	}

	return &types.AuthPayload{
		UserId:       formatID(registerResp.UserId),
		AccessToken:  registerResp.AccessToken,
		RefreshToken: registerResp.RefreshToken,
		ExpiresIn:    registerResp.ExpiresIn,
		UserInfo:     userInfo,
	}, nil
}
