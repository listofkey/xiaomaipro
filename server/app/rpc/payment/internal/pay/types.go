package pay

import "context"

const (
	ChannelStripe = "stripe"
	ChannelAlipay = "alipay"
)

const (
	PayBillStatusNoPay  int32 = 1
	PayBillStatusCancel int32 = 2
	PayBillStatusPay    int32 = 3
	PayBillStatusRefund int32 = 4
)

type PayBillSnapshot struct {
	PayAmount string
}

type PayRequest struct {
	Context       context.Context
	PaymentNo     string
	OrderNo       string
	Amount        string
	Subject       string
	SuccessURL    string
	CancelURL     string
	Currency      string
	CustomerEmail string
	Metadata      map[string]string
}

type NotifyRequest struct {
	RawBody   []byte
	Headers   map[string]string
	Signature string
}

type TradeQueryRequest struct {
	Context       context.Context
	PaymentNo    string
	TradeNo      string
	CallbackData string
}

type RefundRequest struct {
	Context       context.Context
	PaymentNo    string
	TradeNo      string
	CallbackData string
	Amount       string
	Reason       string
	Metadata     map[string]string
}

type PayResult struct {
	Success     bool
	Body        string
	CheckoutURL string
	SessionID   string
	TradeNo     string
	ExpiresAt   string
	RawData     string
}

type TradeResult struct {
	Success       bool
	PayBillStatus int32
	OutTradeNo    string
	TradeNo       string
	TotalAmount   string
	PaidAt        string
	RawData       string
}

type RefundResult struct {
	Success bool
	Pending bool
	Body    string
	Message string
	TradeNo string
}

type NotifyResult struct {
	Success       bool
	Handled       bool
	EventID       string
	EventType     string
	PaymentNo     string
	TradeNo       string
	TotalAmount   string
	PaidAt        string
	PayBillStatus int32
	RawData       string
}

type Strategy interface {
	Pay(req *PayRequest) (*PayResult, error)
	ParseNotify(req *NotifyRequest) (*NotifyResult, error)
	QueryTrade(req *TradeQueryRequest) (*TradeResult, error)
	Refund(req *RefundRequest) (*RefundResult, error)
	Channel() string
}

type StripeConfig struct {
	SecretKey             string
	WebhookSecret         string
	SuccessURL            string
	CancelURL             string
	Currency              string
	PaymentMethodTypes    []string
	CheckoutExpireMinutes int
	RequestTimeoutSeconds int
}

type AlipayConfig struct {
	AppId              string
	SellerId           string
	GatewayUrl         string
	MerchantPrivateKey string
	AlipayPublicKey    string
	ContentKey         string
	ReturnUrl          string
	NotifyUrl          string
	Production         bool
}
