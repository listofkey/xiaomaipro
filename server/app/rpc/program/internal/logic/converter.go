package logic

import (
	"server/app/rpc/model"
	"server/app/rpc/program/programpb/programpb"
)

func modelCategoryToInfo(category *model.Category) *programpb.CategoryInfo {
	if category == nil {
		return nil
	}

	return &programpb.CategoryInfo{
		Id:        category.ID,
		Name:      category.Name,
		Icon:      category.Icon,
		SortOrder: category.SortOrder,
		Status:    int32(category.Status),
	}
}

func modelCityToInfo(city *model.City) *programpb.CityInfo {
	if city == nil {
		return nil
	}

	return &programpb.CityInfo{
		Id:   city.ID,
		Name: city.Name,
	}
}

func modelVenueToInfo(venue *model.Venue, cityName string) *programpb.VenueInfo {
	if venue == nil {
		return nil
	}

	return &programpb.VenueInfo{
		Id:          venue.ID,
		Name:        venue.Name,
		City:        cityName,
		Address:     venue.Address,
		Capacity:    venue.Capacity,
		SeatMapUrl:  venue.SeatMapURL,
		Description: venue.Description,
	}
}

func modelTicketTierToInfo(tier *model.TicketTier) *programpb.TicketTierInfo {
	if tier == nil {
		return nil
	}

	remainStock := tier.TotalStock - tier.SoldCount - tier.LockedCount
	if remainStock < 0 {
		remainStock = 0
	}

	return &programpb.TicketTierInfo{
		Id:          tier.ID,
		EventId:     tier.EventID,
		Name:        tier.Name,
		Price:       tier.Price,
		TotalStock:  tier.TotalStock,
		SoldCount:   tier.SoldCount,
		LockedCount: tier.LockedCount,
		Status:      int32(tier.Status),
		SortOrder:   tier.SortOrder,
		RemainStock: remainStock,
	}
}

func modelEventToBrief(
	event *model.Event,
	category *model.Category,
	cityName string,
	venueName string,
	minPrice float64,
	isHot bool,
) *programpb.EventBrief {
	if event == nil {
		return nil
	}

	return &programpb.EventBrief{
		Id:             event.ID,
		Title:          event.Title,
		PosterUrl:      event.PosterURL,
		Category:       modelCategoryToInfo(category),
		VenueName:      venueName,
		City:           cityName,
		Artist:         event.Artist,
		EventStartTime: formatTime(event.EventStartTime),
		EventEndTime:   formatTime(event.EventEndTime),
		Status:         int32(event.Status),
		MinPrice:       minPrice,
		TicketType:     int32(event.TicketType),
		IsHot:          isHot,
	}
}
