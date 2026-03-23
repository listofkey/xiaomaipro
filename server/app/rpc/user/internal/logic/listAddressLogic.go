package logic

import (
	"context"

	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAddressLogic {
	return &ListAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListAddressLogic) ListAddress(in *userpb.ListAddressReq) (*userpb.ListAddressResp, error) {
	q := l.svcCtx.Query

	addrs, err := q.Address.WithContext(l.ctx).
		Where(q.Address.UserID.Eq(in.UserId)).
		Order(q.Address.IsDefault.Desc(), q.Address.CreatedAt.Desc()).
		Find()
	if err != nil {
		return nil, err
	}
	// TODO：日志

	result := make([]*userpb.AddressInfo, 0, len(addrs))
	for _, a := range addrs {
		result = append(result, modelAddressToInfo(a))
	}

	return &userpb.ListAddressResp{Addresses: result}, nil
}
