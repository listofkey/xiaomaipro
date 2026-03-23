package logic

import (
	"context"
	"errors"
	"regexp"
	"time"

	"server/app/rpc/user/internal/pkg/encrypt"
	"server/app/rpc/user/internal/svc"
	"server/app/rpc/user/userpb/userpb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type VerifyRealNameLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyRealNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyRealNameLogic {
	return &VerifyRealNameLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// idCardRegex 简单校验18位身份证格式
var idCardRegex = regexp.MustCompile(`^\d{17}[\dXx]$`)

func (l *VerifyRealNameLogic) VerifyRealName(in *userpb.VerifyRealNameReq) (*userpb.VerifyRealNameResp, error) {
	if in.UserId <= 0 {
		return nil, errors.New("用户ID无效")
	}
	if in.RealName == "" {
		return nil, errors.New("真实姓名不能为空")
	}
	if !idCardRegex.MatchString(in.IdCard) {
		return nil, errors.New("身份证号格式不正确")
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

	if u.IsVerified == 1 {
		return &userpb.VerifyRealNameResp{
			Success:    false,
			Message:    "已完成实名认证，无需重复提交",
			IsVerified: 1,
		}, nil
	}

	// 加密存储敏感信息（AES-256-GCM）
	encRealName, err := encrypt.AESEncrypt(in.RealName, cfg.AES.Key)
	if err != nil {
		return nil, errors.New("数据加密失败")
	}
	encIDCard, err := encrypt.AESEncrypt(in.IdCard, cfg.AES.Key)
	if err != nil {
		return nil, errors.New("数据加密失败")
	}

	// TODO: 接入第三方实名认证 API（公安部接口 / 银行四要素验证等）
	// 当前实现：格式校验通过即认为认证成功
	verifyPassed := true

	if !verifyPassed {
		return &userpb.VerifyRealNameResp{
			Success:    false,
			Message:    "实名认证失败，姓名与身份证不匹配",
			IsVerified: 0,
		}, nil
	}

	// 更新用户认证状态
	_, err = q.User.WithContext(l.ctx).
		Where(q.User.ID.Eq(in.UserId)).
		Updates(map[string]interface{}{
			"real_name":   encRealName,
			"id_card":     encIDCard,
			"is_verified": 1,
			"updated_at":  time.Now(),
		})
	if err != nil {
		return nil, errors.New("认证信息保存失败")
	}

	return &userpb.VerifyRealNameResp{
		Success:    true,
		Message:    "实名认证成功",
		IsVerified: 1,
	}, nil
}
