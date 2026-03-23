package logic

import (
	"context"
	"errors"
	"time"

	"server/app/rpc/model"
	"server/app/rpc/user/internal/pkg/encrypt"
	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"
	"server/common"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTicketBuyerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTicketBuyerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTicketBuyerLogic {
	return &CreateTicketBuyerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateTicketBuyerLogic) CreateTicketBuyer(in *userpb.CreateTicketBuyerReq) (*userpb.CreateTicketBuyerResp, error) {
	if in.UserId <= 0 {
		return nil, errors.New("用户ID无效")
	}
	if in.Name == "" {
		return nil, errors.New("购票人姓名不能为空")
	}
	if in.IdCard != "" && !idCardRegex.MatchString(in.IdCard) {
		return nil, errors.New("身份证号格式不正确")
	}

	q := l.svcCtx.Query
	cfg := l.svcCtx.Config

	// 若设为默认，先取消旧的默认
	if in.IsDefault == 1 {
		q.TicketBuyer.WithContext(l.ctx).
			Where(q.TicketBuyer.UserID.Eq(in.UserId), q.TicketBuyer.IsDefault.Eq(1)).
			Update(q.TicketBuyer.IsDefault, 0)
	}

	// 加密身份证号
	encIDCard := ""
	if in.IdCard != "" {
		var err error
		encIDCard, err = encrypt.AESEncrypt(in.IdCard, cfg.AES.Key)
		if err != nil {
			return nil, errors.New("数据加密失败")
		}
	}

	tb := &model.TicketBuyer{
		ID:        common.GenerateId(),
		UserID:    in.UserId,
		Name:      in.Name,
		IDCard:    encIDCard,
		Phone:     in.Phone,
		IsDefault: int16(in.IsDefault),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := q.TicketBuyer.WithContext(l.ctx).Create(tb); err != nil {
		return nil, errors.New("新增购票人失败: " + err.Error())
	}

	return &userpb.CreateTicketBuyerResp{
		TicketBuyer: modelTicketBuyerToInfo(tb, cfg.AES.Key),
	}, nil
}
