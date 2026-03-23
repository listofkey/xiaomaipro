package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTicketBuyerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTicketBuyerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTicketBuyerLogic {
	return &CreateTicketBuyerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTicketBuyerLogic) CreateTicketBuyer(req *types.CreateTicketBuyerReq) (resp *types.TicketBuyerResp, err error) {
	createResp, err := l.svcCtx.UserRpc.CreateTicketBuyer(l.ctx, &userservice.CreateTicketBuyerReq{
		UserId:    authctx.UserID(l.ctx),
		Name:      req.Name,
		IdCard:    req.IdCard,
		Phone:     req.Phone,
		IsDefault: boolToInt32(req.IsDefault),
	})
	if err != nil {
		return nil, err
	}

	return &types.TicketBuyerResp{
		TicketBuyer: mapTicketBuyerInfo(createResp.TicketBuyer),
	}, nil
}
