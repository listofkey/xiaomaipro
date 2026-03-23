package logic

import (
	"context"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTicketBuyerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteTicketBuyerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTicketBuyerLogic {
	return &DeleteTicketBuyerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTicketBuyerLogic) DeleteTicketBuyer(req *types.DeleteTicketBuyerReq) (resp *types.OperationResp, err error) {
	buyerID, err := parseID(req.BuyerId, "buyerId")
	if err != nil {
		return nil, err
	}

	deleteResp, err := l.svcCtx.UserRpc.DeleteTicketBuyer(l.ctx, &userservice.DeleteTicketBuyerReq{
		BuyerId: buyerID,
		UserId:  authctx.UserID(l.ctx),
	})
	if err != nil {
		return nil, err
	}

	return &types.OperationResp{
		Success: deleteResp.Success,
	}, nil
}
