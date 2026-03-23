package types

type PaymentDetailReq struct {
	OrderNo string `form:"orderNo" json:"orderNo"`
}

type PaymentTradeCheckReq struct {
	OrderNo   string `json:"orderNo,optional"`
	PaymentNo string `json:"paymentNo,optional"`
	Channel   string `json:"channel,optional"`
}

type PaymentDetailResp struct {
	Payment     OrderPaymentInfo `json:"payment"`
	Refund      OrderRefundInfo  `json:"refund"`
	OrderStatus int32            `json:"orderStatus"`
	PaidAt      string           `json:"paidAt"`
}

type PaymentTradeCheckResp struct {
	Success           bool             `json:"success"`
	Paid              bool             `json:"paid"`
	Payment           OrderPaymentInfo `json:"payment"`
	OrderStatus       int32            `json:"orderStatus"`
	PaidAt            string           `json:"paidAt"`
	CheckoutSessionId string           `json:"checkoutSessionId"`
}
