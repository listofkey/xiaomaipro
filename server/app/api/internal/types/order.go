package types

type CreateOrderReq struct {
	EventId        string   `json:"eventId"`
	TicketTierId   string   `json:"ticketTierId"`
	Quantity       int32    `json:"quantity"`
	TicketBuyerIds []string `json:"ticketBuyerIds"`
	AddressId      string   `json:"addressId,optional"`
	PayMethod      int32    `json:"payMethod,optional"`
	RequestId      string   `json:"requestId,optional"`
}

type GetOrderQueueStatusReq struct {
	QueueToken string `form:"queueToken" json:"queueToken,optional"`
}

type PayOrderReq struct {
	OrderNo   string `json:"orderNo"`
	PayMethod int32  `json:"payMethod,optional"`
	Channel   string `json:"channel,optional"`
}

type CancelOrderReq struct {
	OrderNo string `json:"orderNo"`
}

type ApplyRefundReq struct {
	OrderNo string `json:"orderNo"`
	Reason  string `json:"reason,optional"`
}

type ListOrderReq struct {
	Status   int32 `form:"status,optional" json:"status,optional"`
	Page     int32 `form:"page,optional" json:"page,optional"`
	PageSize int32 `form:"pageSize,optional" json:"pageSize,optional"`
}

type GetOrderDetailReq struct {
	OrderNo string `form:"orderNo" json:"orderNo,optional"`
}

type OrderBuyerInfo struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	IdCard string `json:"idCard"`
	Phone  string `json:"phone"`
}

type OrderTicketInfo struct {
	Id            string         `json:"id"`
	TicketBuyerId string         `json:"ticketBuyerId"`
	TicketCode    string         `json:"ticketCode"`
	QrCodeUrl     string         `json:"qrCodeUrl"`
	Status        int32          `json:"status"`
	SeatInfo      string         `json:"seatInfo"`
	VerifiedAt    string         `json:"verifiedAt"`
	Buyer         OrderBuyerInfo `json:"buyer"`
}

type OrderPaymentInfo struct {
	Id           string  `json:"id"`
	PaymentNo    string  `json:"paymentNo"`
	PayMethod    int32   `json:"payMethod"`
	Amount       float64 `json:"amount"`
	Status       int32   `json:"status"`
	TradeNo      string  `json:"tradeNo"`
	PaidAt       string  `json:"paidAt"`
	CreatedAt    string  `json:"createdAt"`
	CallbackData string  `json:"callbackData"`
}

type OrderRefundInfo struct {
	Id           string  `json:"id"`
	RefundNo     string  `json:"refundNo"`
	RefundAmount float64 `json:"refundAmount"`
	Status       int32   `json:"status"`
	Reason       string  `json:"reason"`
	RejectReason string  `json:"rejectReason"`
	TradeNo      string  `json:"tradeNo"`
	CreatedAt    string  `json:"createdAt"`
	RefundedAt   string  `json:"refundedAt"`
}

type OrderDeliveryInfo struct {
	TicketType     int32  `json:"ticketType"`
	DeliveryMethod string `json:"deliveryMethod"`
	AddressId      string `json:"addressId"`
	ReceiverName   string `json:"receiverName"`
	ReceiverPhone  string `json:"receiverPhone"`
	Province       string `json:"province"`
	City           string `json:"city"`
	District       string `json:"district"`
	Detail         string `json:"detail"`
	DeliveryStatus string `json:"deliveryStatus"`
}

type OrderSummary struct {
	Id             string  `json:"id"`
	OrderNo        string  `json:"orderNo"`
	EventId        string  `json:"eventId"`
	TicketTierId   string  `json:"ticketTierId"`
	EventTitle     string  `json:"eventTitle"`
	PosterUrl      string  `json:"posterUrl"`
	VenueName      string  `json:"venueName"`
	City           string  `json:"city"`
	EventStartTime string  `json:"eventStartTime"`
	EventEndTime   string  `json:"eventEndTime"`
	TicketTierName string  `json:"ticketTierName"`
	Quantity       int32   `json:"quantity"`
	UnitPrice      float64 `json:"unitPrice"`
	TotalAmount    float64 `json:"totalAmount"`
	Status         int32   `json:"status"`
	StatusText     string  `json:"statusText"`
	PayDeadline    string  `json:"payDeadline"`
	PaidAt         string  `json:"paidAt"`
	CancelledAt    string  `json:"cancelledAt"`
	CreatedAt      string  `json:"createdAt"`
	TicketType     int32   `json:"ticketType"`
}

type OrderDetail struct {
	Id             string            `json:"id"`
	OrderNo        string            `json:"orderNo"`
	UserId         string            `json:"userId"`
	EventId        string            `json:"eventId"`
	TicketTierId   string            `json:"ticketTierId"`
	EventTitle     string            `json:"eventTitle"`
	Description    string            `json:"description"`
	PosterUrl      string            `json:"posterUrl"`
	VenueName      string            `json:"venueName"`
	VenueAddress   string            `json:"venueAddress"`
	City           string            `json:"city"`
	EventStartTime string            `json:"eventStartTime"`
	EventEndTime   string            `json:"eventEndTime"`
	SaleStartTime  string            `json:"saleStartTime"`
	SaleEndTime    string            `json:"saleEndTime"`
	TicketTierName string            `json:"ticketTierName"`
	Quantity       int32             `json:"quantity"`
	UnitPrice      float64           `json:"unitPrice"`
	TotalAmount    float64           `json:"totalAmount"`
	Status         int32             `json:"status"`
	StatusText     string            `json:"statusText"`
	PayDeadline    string            `json:"payDeadline"`
	PaidAt         string            `json:"paidAt"`
	CancelledAt    string            `json:"cancelledAt"`
	CreatedAt      string            `json:"createdAt"`
	PurchaseLimit  int32             `json:"purchaseLimit"`
	NeedRealName   int32             `json:"needRealName"`
	TicketType     int32             `json:"ticketType"`
	Delivery       OrderDeliveryInfo `json:"delivery"`
	Tickets        []OrderTicketInfo `json:"tickets"`
	Payment        OrderPaymentInfo  `json:"payment"`
	Refund         OrderRefundInfo   `json:"refund"`
}

type CreateOrderResp struct {
	OrderNo     string `json:"orderNo"`
	QueueToken  string `json:"queueToken"`
	QueueStatus int32  `json:"queueStatus"`
	Message     string `json:"message"`
}

type OrderQueueStatusResp struct {
	QueueToken  string       `json:"queueToken"`
	OrderNo     string       `json:"orderNo"`
	QueueStatus int32        `json:"queueStatus"`
	Message     string       `json:"message"`
	Order       OrderSummary `json:"order,optional"`
}

type PayOrderResp struct {
	Success           bool             `json:"success"`
	PayForm           string           `json:"payForm"`
	Payment           OrderPaymentInfo `json:"payment"`
	OrderStatus       int32            `json:"orderStatus"`
	PaidAt            string           `json:"paidAt"`
	CheckoutUrl       string           `json:"checkoutUrl"`
	CheckoutSessionId string           `json:"checkoutSessionId"`
	SessionExpiresAt  string           `json:"sessionExpiresAt"`
}

type ApplyRefundResp struct {
	Success bool            `json:"success"`
	OrderNo string          `json:"orderNo"`
	Refund  OrderRefundInfo `json:"refund"`
}

type OrderListResp struct {
	Orders   []OrderSummary `json:"orders"`
	Total    int64          `json:"total"`
	Page     int32          `json:"page"`
	PageSize int32          `json:"pageSize"`
}

type OrderDetailResp struct {
	Order OrderDetail `json:"order"`
}
