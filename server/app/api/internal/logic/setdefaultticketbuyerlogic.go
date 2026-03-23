package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetDefaultTicketBuyerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetDefaultTicketBuyerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultTicketBuyerLogic {
	return &SetDefaultTicketBuyerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetDefaultTicketBuyerLogic) SetDefaultTicketBuyer(req *types.SetDefaultTicketBuyerReq) (resp *types.OperationResp, err error) {
	buyerID, err := parseID(req.BuyerId, "buyerId")
	if err != nil {
		return nil, err
	}

	setResp, err := l.svcCtx.UserRpc.SetDefaultTicketBuyer(l.ctx, &userservice.SetDefaultTicketBuyerReq{
		BuyerId: buyerID,
		UserId:  authctx.UserID(l.ctx),
	})
	if err != nil {
		return nil, err
	}

	return &types.OperationResp{
		Success: setResp.Success,
	}, nil
}
