package logic

import (
	"fmt"
	"strconv"
	"strings"

	"server/app/api/internal/types"
	"server/app/rpc/user/userservice"
)

func mapUserInfo(info *userservice.UserInfo) types.UserInfo {
	if info == nil {
		return types.UserInfo{}
	}

	return types.UserInfo{
		Id:         formatID(info.Id),
		Phone:      info.Phone,
		Email:      info.Email,
		Nickname:   info.Nickname,
		Avatar:     info.Avatar,
		Status:     info.Status,
		IsVerified: info.IsVerified == 1,
		RealName:   info.RealName,
		CreatedAt:  info.CreatedAt,
	}
}

func mapAddressInfo(info *userservice.AddressInfo) types.AddressInfo {
	if info == nil {
		return types.AddressInfo{}
	}

	return types.AddressInfo{
		Id:            formatID(info.Id),
		ReceiverName:  info.ReceiverName,
		ReceiverPhone: info.ReceiverPhone,
		Province:      info.Province,
		City:          info.City,
		District:      info.District,
		Detail:        info.Detail,
		IsDefault:     info.IsDefault == 1,
		CreatedAt:     info.CreatedAt,
	}
}

func mapAddressList(items []*userservice.AddressInfo) []types.AddressInfo {
	result := make([]types.AddressInfo, 0, len(items))
	for _, item := range items {
		result = append(result, mapAddressInfo(item))
	}
	return result
}

func mapTicketBuyerInfo(info *userservice.TicketBuyerInfo) types.TicketBuyerInfo {
	if info == nil {
		return types.TicketBuyerInfo{}
	}

	return types.TicketBuyerInfo{
		Id:        formatID(info.Id),
		Name:      info.Name,
		IdCard:    info.IdCard,
		Phone:     info.Phone,
		IsDefault: info.IsDefault == 1,
		CreatedAt: info.CreatedAt,
	}
}

func mapTicketBuyerList(items []*userservice.TicketBuyerInfo) []types.TicketBuyerInfo {
	result := make([]types.TicketBuyerInfo, 0, len(items))
	for _, item := range items {
		result = append(result, mapTicketBuyerInfo(item))
	}
	return result
}

func boolToInt32(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

func formatID(id int64) string {
	return strconv.FormatInt(id, 10)
}

func parseID(raw string, field string) (int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, fmt.Errorf("%s is required", field)
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid int64 string", field)
	}

	return id, nil
}

func parseOptionalID(raw string, field string) (int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, nil
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid int64 string", field)
	}
	return id, nil
}

func parseIDs(values []string, field string) ([]int64, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("%s is required", field)
	}

	result := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, raw := range values {
		id, err := parseID(raw, field)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result, nil
}
