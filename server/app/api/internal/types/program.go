package types

type ProgramCategoryInfo struct {
	Id        int32  `json:"id"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	SortOrder int32  `json:"sortOrder"`
	Status    int32  `json:"status"`
}

type ProgramCityInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ProgramVenueInfo struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	City        string `json:"city"`
	Address     string `json:"address"`
	Capacity    int32  `json:"capacity"`
	SeatMapUrl  string `json:"seatMapUrl"`
	Description string `json:"description"`
}

type ProgramTicketTierInfo struct {
	Id          string  `json:"id"`
	EventId     string  `json:"eventId"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	TotalStock  int32   `json:"totalStock"`
	SoldCount   int32   `json:"soldCount"`
	LockedCount int32   `json:"lockedCount"`
	Status      int32   `json:"status"`
	SortOrder   int32   `json:"sortOrder"`
	RemainStock int32   `json:"remainStock"`
}

type ProgramEventBrief struct {
	Id             string              `json:"id"`
	Title          string              `json:"title"`
	PosterUrl      string              `json:"posterUrl"`
	Category       ProgramCategoryInfo `json:"category"`
	VenueName      string              `json:"venueName"`
	City           string              `json:"city"`
	Artist         string              `json:"artist"`
	EventStartTime string              `json:"eventStartTime"`
	EventEndTime   string              `json:"eventEndTime"`
	Status         int32               `json:"status"`
	MinPrice       float64             `json:"minPrice"`
	TicketType     int32               `json:"ticketType"`
	IsHot          bool                `json:"isHot"`
}

type ProgramEventDetail struct {
	Id             string                  `json:"id"`
	Title          string                  `json:"title"`
	Description    string                  `json:"description"`
	PosterUrl      string                  `json:"posterUrl"`
	Category       ProgramCategoryInfo     `json:"category"`
	Venue          ProgramVenueInfo        `json:"venue"`
	City           string                  `json:"city"`
	Artist         string                  `json:"artist"`
	EventStartTime string                  `json:"eventStartTime"`
	EventEndTime   string                  `json:"eventEndTime"`
	SaleStartTime  string                  `json:"saleStartTime"`
	SaleEndTime    string                  `json:"saleEndTime"`
	Status         int32                   `json:"status"`
	PurchaseLimit  int32                   `json:"purchaseLimit"`
	NeedRealName   int32                   `json:"needRealName"`
	TicketType     int32                   `json:"ticketType"`
	TicketTiers    []ProgramTicketTierInfo `json:"ticketTiers"`
}

type ListEventsReq struct {
	Page       int32  `form:"page,optional" json:"page,optional"`
	PageSize   int32  `form:"pageSize,optional" json:"pageSize,optional"`
	CategoryId int32  `form:"categoryId,optional" json:"categoryId,optional"`
	City       string `form:"city,optional" json:"city,optional"`
	StartDate  string `form:"startDate,optional" json:"startDate,optional"`
	EndDate    string `form:"endDate,optional" json:"endDate,optional"`
	SortBy     string `form:"sortBy,optional" json:"sortBy,optional"`
}

type SearchEventsReq struct {
	Keyword    string `form:"keyword,optional" json:"keyword,optional"`
	CategoryId int32  `form:"categoryId,optional" json:"categoryId,optional"`
	City       string `form:"city,optional" json:"city,optional"`
	StartDate  string `form:"startDate,optional" json:"startDate,optional"`
	EndDate    string `form:"endDate,optional" json:"endDate,optional"`
	Page       int32  `form:"page,optional" json:"page,optional"`
	PageSize   int32  `form:"pageSize,optional" json:"pageSize,optional"`
}

type GetEventDetailReq struct {
	EventId string `form:"eventId" json:"eventId,optional"`
}

type ListCategoriesReq struct {
	Status int32 `form:"status,optional" json:"status,optional"`
}

type GetHotRecommendReq struct {
	City  string `form:"city,optional" json:"city,optional"`
	Limit int32  `form:"limit,optional" json:"limit,optional"`
}

type ProgramEventListResp struct {
	Events   []ProgramEventBrief `json:"events"`
	Total    int64               `json:"total"`
	Page     int32               `json:"page"`
	PageSize int32               `json:"pageSize"`
}

type ProgramEventDetailResp struct {
	Event ProgramEventDetail `json:"event"`
}

type ProgramCategoryListResp struct {
	Categories []ProgramCategoryInfo `json:"categories"`
	Cities     []ProgramCityInfo     `json:"cities"`
}

type ProgramHotRecommendResp struct {
	Events []ProgramEventBrief `json:"events"`
}
