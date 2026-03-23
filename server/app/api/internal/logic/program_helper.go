package logic

import (
	"server/app/api/internal/types"
	"server/app/rpc/program/programservice"
)

func mapProgramCategoryInfo(info *programservice.CategoryInfo) types.ProgramCategoryInfo {
	if info == nil {
		return types.ProgramCategoryInfo{}
	}

	return types.ProgramCategoryInfo{
		Id:        info.Id,
		Name:      info.Name,
		Icon:      info.Icon,
		SortOrder: info.SortOrder,
		Status:    info.Status,
	}
}

func mapProgramCityInfo(info *programservice.CityInfo) types.ProgramCityInfo {
	if info == nil {
		return types.ProgramCityInfo{}
	}

	return types.ProgramCityInfo{
		Id:   formatID(info.Id),
		Name: info.Name,
	}
}

func mapProgramVenueInfo(info *programservice.VenueInfo) types.ProgramVenueInfo {
	if info == nil {
		return types.ProgramVenueInfo{}
	}

	return types.ProgramVenueInfo{
		Id:          formatID(info.Id),
		Name:        info.Name,
		City:        info.City,
		Address:     info.Address,
		Capacity:    info.Capacity,
		SeatMapUrl:  info.SeatMapUrl,
		Description: info.Description,
	}
}

func mapProgramTicketTierInfo(info *programservice.TicketTierInfo) types.ProgramTicketTierInfo {
	if info == nil {
		return types.ProgramTicketTierInfo{}
	}

	return types.ProgramTicketTierInfo{
		Id:          formatID(info.Id),
		EventId:     formatID(info.EventId),
		Name:        info.Name,
		Price:       info.Price,
		TotalStock:  info.TotalStock,
		SoldCount:   info.SoldCount,
		LockedCount: info.LockedCount,
		Status:      info.Status,
		SortOrder:   info.SortOrder,
		RemainStock: info.RemainStock,
	}
}

func mapProgramEventBrief(info *programservice.EventBrief) types.ProgramEventBrief {
	if info == nil {
		return types.ProgramEventBrief{}
	}

	return types.ProgramEventBrief{
		Id:             formatID(info.Id),
		Title:          info.Title,
		PosterUrl:      info.PosterUrl,
		Category:       mapProgramCategoryInfo(info.Category),
		VenueName:      info.VenueName,
		City:           info.City,
		Artist:         info.Artist,
		EventStartTime: info.EventStartTime,
		EventEndTime:   info.EventEndTime,
		Status:         info.Status,
		MinPrice:       info.MinPrice,
		TicketType:     info.TicketType,
		IsHot:          info.IsHot,
	}
}

func mapProgramEventBriefList(items []*programservice.EventBrief) []types.ProgramEventBrief {
	result := make([]types.ProgramEventBrief, 0, len(items))
	for _, item := range items {
		result = append(result, mapProgramEventBrief(item))
	}
	return result
}

func mapProgramTicketTierList(items []*programservice.TicketTierInfo) []types.ProgramTicketTierInfo {
	result := make([]types.ProgramTicketTierInfo, 0, len(items))
	for _, item := range items {
		result = append(result, mapProgramTicketTierInfo(item))
	}
	return result
}

func mapProgramEventDetail(info *programservice.EventDetail) types.ProgramEventDetail {
	if info == nil {
		return types.ProgramEventDetail{}
	}

	return types.ProgramEventDetail{
		Id:             formatID(info.Id),
		Title:          info.Title,
		Description:    info.Description,
		PosterUrl:      info.PosterUrl,
		Category:       mapProgramCategoryInfo(info.Category),
		Venue:          mapProgramVenueInfo(info.Venue),
		City:           info.City,
		Artist:         info.Artist,
		EventStartTime: info.EventStartTime,
		EventEndTime:   info.EventEndTime,
		SaleStartTime:  info.SaleStartTime,
		SaleEndTime:    info.SaleEndTime,
		Status:         info.Status,
		PurchaseLimit:  info.PurchaseLimit,
		NeedRealName:   info.NeedRealName,
		TicketType:     info.TicketType,
		TicketTiers:    mapProgramTicketTierList(info.TicketTiers),
	}
}

func mapProgramCategoryList(items []*programservice.CategoryInfo) []types.ProgramCategoryInfo {
	result := make([]types.ProgramCategoryInfo, 0, len(items))
	for _, item := range items {
		result = append(result, mapProgramCategoryInfo(item))
	}
	return result
}

func mapProgramCityList(items []*programservice.CityInfo) []types.ProgramCityInfo {
	result := make([]types.ProgramCityInfo, 0, len(items))
	for _, item := range items {
		result = append(result, mapProgramCityInfo(item))
	}
	return result
}
