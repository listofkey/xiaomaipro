package logic

import (
	"server/app/rpc/model"
	"server/app/rpc/user/internal/pkg/encrypt"
	"server/app/rpc/user/userpb/userpb"
)

// modelUserToInfo 将 model.User 转换为 proto UserInfo，做必要的脱敏处理
func modelUserToInfo(u *model.User, aesKey string) *userpb.UserInfo {
	realName := ""
	if u.IsVerified == 1 && u.RealName != "" {
		decrypted, err := encrypt.AESDecrypt(u.RealName, aesKey)
		if err == nil {
			realName = encrypt.MaskRealName(decrypted)
		}
	}

	return &userpb.UserInfo{
		Id:         u.ID,
		Phone:      encrypt.MaskPhone(u.Phone),
		Email:      u.Email,
		Nickname:   u.Nickname,
		Avatar:     u.Avatar,
		Status:     int32(u.Status),
		IsVerified: int32(u.IsVerified),
		RealName:   realName,
		CreatedAt:  u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// modelAddressToInfo 将 model.Address 转换为 proto AddressInfo
func modelAddressToInfo(a *model.Address) *userpb.AddressInfo {
	return &userpb.AddressInfo{
		Id:            a.ID,
		UserId:        a.UserID,
		ReceiverName:  a.ReceiverName,
		ReceiverPhone: a.ReceiverPhone,
		Province:      a.Province,
		City:          a.City,
		District:      a.District,
		Detail:        a.Detail,
		IsDefault:     int32(a.IsDefault),
		CreatedAt:     a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// modelTicketBuyerToInfo 将 model.TicketBuyer 转换为 proto TicketBuyerInfo（IDCard 脱敏）
func modelTicketBuyerToInfo(tb *model.TicketBuyer, aesKey string) *userpb.TicketBuyerInfo {
	idCard := ""
	if tb.IDCard != "" {
		decrypted, err := encrypt.AESDecrypt(tb.IDCard, aesKey)
		if err == nil {
			idCard = encrypt.MaskIDCard(decrypted)
		}
	}
	return &userpb.TicketBuyerInfo{
		Id:        tb.ID,
		UserId:    tb.UserID,
		Name:      tb.Name,
		IdCard:    idCard,
		Phone:     tb.Phone,
		IsDefault: int32(tb.IsDefault),
		CreatedAt: tb.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
