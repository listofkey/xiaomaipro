package pay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	stripe "github.com/stripe/stripe-go/v84"
	stripecheckout "github.com/stripe/stripe-go/v84/checkout/session"
	striperefund "github.com/stripe/stripe-go/v84/refund"
	stripewebhook "github.com/stripe/stripe-go/v84/webhook"
)

const stripeTimeLayout = "2006-01-02 15:04:05"

type StripeStrategy struct {
	config StripeConfig
}

type stripeEventEnvelope struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object stripeSessionSnapshot `json:"object"`
	} `json:"data"`
}

type stripeSessionSnapshot struct {
	ID                string            `json:"id"`
	ClientReferenceID string            `json:"client_reference_id"`
	Metadata          map[string]string `json:"metadata"`
	PaymentIntent     json.RawMessage   `json:"payment_intent"`
	URL               string            `json:"url"`
	Status            string            `json:"status"`
	PaymentStatus     string            `json:"payment_status"`
	AmountTotal       int64             `json:"amount_total"`
	ExpiresAt         int64             `json:"expires_at"`
	Created           int64             `json:"created"`
}

func NewStripeStrategy(cfg StripeConfig) (*StripeStrategy, error) {
	cfg = normalizeStripeConfig(cfg)
	if strings.TrimSpace(cfg.SecretKey) == "" {
		return nil, errors.New("stripe secret key is empty")
	}
	return &StripeStrategy{config: cfg}, nil
}

func (s *StripeStrategy) Pay(req *PayRequest) (*PayResult, error) {
	if req == nil {
		return nil, errors.New("pay request is empty")
	}

	amountMinor, err := decimalToMinor(req.Amount)
	if err != nil {
		return nil, err
	}
	if amountMinor <= 0 {
		return nil, errors.New("pay amount must be greater than 0")
	}

	s.setAPIKey()
	reqCtx, cancel := s.requestContext(req.Context)
	defer cancel()

	metadata := map[string]string{
		"payment_no": strings.TrimSpace(req.PaymentNo),
		"order_no":   strings.TrimSpace(req.OrderNo),
	}
	for key, value := range req.Metadata {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		metadata[key] = strings.TrimSpace(value)
	}

	params := &stripe.CheckoutSessionParams{
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		PaymentMethodTypes: stripe.StringSlice(s.config.PaymentMethodTypes),
		ClientReferenceID:  stripe.String(strings.TrimSpace(req.PaymentNo)),
		SuccessURL: stripe.String(buildCheckoutReturnURL(req.SuccessURL, map[string]string{
			"orderNo":       strings.TrimSpace(req.OrderNo),
			"paymentNo":     strings.TrimSpace(req.PaymentNo),
			"paymentResult": "success",
			"session_id":    "{CHECKOUT_SESSION_ID}",
		})),
		CancelURL: stripe.String(buildCheckoutReturnURL(req.CancelURL, map[string]string{
			"orderNo":       strings.TrimSpace(req.OrderNo),
			"paymentNo":     strings.TrimSpace(req.PaymentNo),
			"paymentResult": "cancel",
		})),
		Metadata: metadata,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(strings.ToLower(strings.TrimSpace(req.Currency))),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(firstNonEmpty(req.Subject, req.OrderNo, req.PaymentNo)),
					},
					UnitAmount: stripe.Int64(amountMinor),
				},
			},
		},
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: metadata,
		},
	}
	params.Context = reqCtx
	if email := strings.TrimSpace(req.CustomerEmail); email != "" {
		params.CustomerEmail = stripe.String(email)
	}
	if expiresAt := s.checkoutExpiresAt(); expiresAt != nil {
		params.ExpiresAt = stripe.Int64(expiresAt.Unix())
	}

	result, err := stripecheckout.New(params)
	if err != nil {
		return nil, wrapStripeRequestError("create checkout session", err)
	}

	rawData, _ := json.Marshal(result)
	return &PayResult{
		Success:     true,
		Body:        result.URL,
		CheckoutURL: result.URL,
		SessionID:   result.ID,
		TradeNo:     result.ID,
		ExpiresAt:   formatUnix(result.ExpiresAt),
		RawData:     string(rawData),
	}, nil
}

func (s *StripeStrategy) ParseNotify(req *NotifyRequest) (*NotifyResult, error) {
	if req == nil {
		return nil, errors.New("notify request is empty")
	}
	if len(req.RawBody) == 0 {
		return nil, errors.New("stripe webhook body is empty")
	}
	if strings.TrimSpace(s.config.WebhookSecret) == "" {
		return nil, errors.New("stripe webhook secret is empty")
	}

	signature := firstNonEmpty(req.Signature, req.Headers["Stripe-Signature"], req.Headers["stripe-signature"])
	event, err := stripewebhook.ConstructEventWithOptions(
		req.RawBody,
		signature,
		s.config.WebhookSecret,
		stripewebhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook signature: %w", err)
	}

	resp := &NotifyResult{
		Success:   true,
		Handled:   false,
		EventID:   event.ID,
		EventType: string(event.Type),
		RawData:   string(req.RawBody),
	}

	switch event.Type {
	case "checkout.session.completed", "checkout.session.async_payment_succeeded", "checkout.session.expired", "checkout.session.async_payment_failed":
	default:
		return resp, nil
	}

	var snapshot stripeSessionSnapshot
	if err := json.Unmarshal(event.Data.Raw, &snapshot); err != nil {
		return nil, fmt.Errorf("unmarshal checkout session event failed: %w", err)
	}

	resp.Handled = true
	resp.PaymentNo = extractPaymentNo(snapshot)
	resp.TradeNo = strings.TrimSpace(snapshot.ID)
	if snapshot.AmountTotal > 0 {
		resp.TotalAmount = minorToDecimal(snapshot.AmountTotal)
	}

	switch event.Type {
	case "checkout.session.expired", "checkout.session.async_payment_failed":
		resp.PayBillStatus = PayBillStatusCancel
		return resp, nil
	}

	if strings.EqualFold(snapshot.PaymentStatus, string(stripe.CheckoutSessionPaymentStatusPaid)) {
		resp.PayBillStatus = PayBillStatusPay
		resp.PaidAt = formatUnix(firstNonZero(snapshot.Created, event.Created))
		return resp, nil
	}

	if strings.EqualFold(snapshot.Status, string(stripe.CheckoutSessionStatusExpired)) {
		resp.PayBillStatus = PayBillStatusCancel
		return resp, nil
	}

	resp.PayBillStatus = PayBillStatusNoPay
	return resp, nil
}

func (s *StripeStrategy) QueryTrade(req *TradeQueryRequest) (*TradeResult, error) {
	if req == nil {
		return nil, errors.New("trade query request is empty")
	}

	sessionID := strings.TrimSpace(req.TradeNo)
	if sessionID == "" {
		snapshot, ok := decodeStripeSessionSnapshot(req.CallbackData)
		if ok {
			sessionID = strings.TrimSpace(snapshot.ID)
		}
	}
	if sessionID == "" {
		return &TradeResult{Success: false}, nil
	}

	reqCtx, cancel := s.requestContext(req.Context)
	defer cancel()

	sessionObj, err := s.getCheckoutSession(reqCtx, sessionID)
	if err != nil {
		return nil, wrapStripeRequestError("query checkout session", err)
	}

	rawData, _ := json.Marshal(sessionObj)
	resp := &TradeResult{
		Success:     true,
		OutTradeNo:  firstNonEmpty(sessionObj.ClientReferenceID, req.PaymentNo, sessionObj.Metadata["payment_no"]),
		TradeNo:     strings.TrimSpace(sessionObj.ID),
		TotalAmount: minorToDecimal(sessionObj.AmountTotal),
		RawData:     string(rawData),
	}

	switch {
	case strings.EqualFold(string(sessionObj.PaymentStatus), string(stripe.CheckoutSessionPaymentStatusPaid)):
		resp.PayBillStatus = PayBillStatusPay
		resp.PaidAt = formatUnix(firstNonZero(extractPaymentIntentCreated(sessionObj.PaymentIntent), sessionObj.Created))
	case strings.EqualFold(string(sessionObj.Status), string(stripe.CheckoutSessionStatusExpired)):
		resp.PayBillStatus = PayBillStatusCancel
	default:
		resp.PayBillStatus = PayBillStatusNoPay
	}

	return resp, nil
}

func (s *StripeStrategy) Refund(req *RefundRequest) (*RefundResult, error) {
	if req == nil {
		return nil, errors.New("refund request is empty")
	}

	reqCtx, cancel := s.requestContext(req.Context)
	defer cancel()

	paymentIntentID, err := s.resolvePaymentIntentID(reqCtx, req.TradeNo, req.CallbackData)
	if err != nil {
		return nil, wrapStripeRequestError("resolve refund payment intent", err)
	}
	if paymentIntentID == "" {
		return nil, errors.New("stripe payment intent not found")
	}

	amountMinor, err := decimalToMinor(req.Amount)
	if err != nil {
		return nil, err
	}
	if amountMinor <= 0 {
		return nil, errors.New("refund amount must be greater than 0")
	}

	s.setAPIKey()

	reason := string(stripe.RefundReasonRequestedByCustomer)
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
		Amount:        stripe.Int64(amountMinor),
		Reason:        &reason,
		Metadata: map[string]string{
			"payment_no": strings.TrimSpace(req.PaymentNo),
			"reason":     strings.TrimSpace(req.Reason),
		},
	}
	params.Context = reqCtx
	params.SetIdempotencyKey(buildRefundIdempotencyKey(req))
	for key, value := range req.Metadata {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		params.Metadata[key] = strings.TrimSpace(value)
	}

	result, err := striperefund.New(params)
	if err != nil {
		return nil, wrapStripeRequestError("create refund", err)
	}

	rawData, _ := json.Marshal(result)
	refundStatus := string(result.Status)
	refundResult := &RefundResult{
		Success: result != nil && (result.Status == stripe.RefundStatusSucceeded || result.Status == stripe.RefundStatusPending),
		Pending: result != nil && result.Status == stripe.RefundStatusPending,
		Body:    string(rawData),
		TradeNo: result.ID,
	}
	if refundResult.Success {
		return refundResult, nil
	}

	refundResult.Message = firstNonEmpty(
		strings.TrimSpace(string(result.FailureReason)),
		strings.TrimSpace(refundStatus),
		"stripe refund failed",
	)
	return refundResult, nil
}

func (s *StripeStrategy) Channel() string {
	return ChannelStripe
}

func (s *StripeStrategy) setAPIKey() {
	stripe.Key = s.config.SecretKey
}

func (s *StripeStrategy) checkoutExpiresAt() *time.Time {
	if s.config.CheckoutExpireMinutes <= 0 {
		return nil
	}
	value := time.Now().Add(time.Duration(s.config.CheckoutExpireMinutes) * time.Minute)
	return &value
}

func (s *StripeStrategy) getCheckoutSession(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error) {
	s.setAPIKey()

	params := &stripe.CheckoutSessionParams{}
	params.AddExpand("payment_intent")
	params.Context = ctx
	return stripecheckout.Get(strings.TrimSpace(sessionID), params)
}

func (s *StripeStrategy) resolvePaymentIntentID(ctx context.Context, tradeNo, callbackData string) (string, error) {
	tradeNo = strings.TrimSpace(tradeNo)
	if strings.HasPrefix(tradeNo, "pi_") {
		return tradeNo, nil
	}

	if snapshot, ok := decodeStripeSessionSnapshot(callbackData); ok {
		if paymentIntentID := extractIDFromRawMessage(snapshot.PaymentIntent); paymentIntentID != "" {
			return paymentIntentID, nil
		}
		if strings.HasPrefix(strings.TrimSpace(snapshot.ID), "cs_") {
			tradeNo = snapshot.ID
		}
	}

	if strings.HasPrefix(tradeNo, "cs_") {
		sessionObj, err := s.getCheckoutSession(ctx, tradeNo)
		if err != nil {
			return "", err
		}
		if sessionObj.PaymentIntent != nil {
			return strings.TrimSpace(sessionObj.PaymentIntent.ID), nil
		}
	}

	return "", nil
}

func (s *StripeStrategy) requestContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if s.config.RequestTimeoutSeconds <= 0 {
		return parent, func() {}
	}
	return context.WithTimeout(parent, time.Duration(s.config.RequestTimeoutSeconds)*time.Second)
}

func normalizeStripeConfig(cfg StripeConfig) StripeConfig {
	cfg.SecretKey = strings.TrimSpace(cfg.SecretKey)
	cfg.WebhookSecret = strings.TrimSpace(cfg.WebhookSecret)
	cfg.SuccessURL = strings.TrimSpace(cfg.SuccessURL)
	cfg.CancelURL = strings.TrimSpace(cfg.CancelURL)
	cfg.Currency = strings.ToLower(strings.TrimSpace(cfg.Currency))
	if cfg.Currency == "" {
		cfg.Currency = "cny"
	}

	methods := make([]string, 0, len(cfg.PaymentMethodTypes))
	for _, method := range cfg.PaymentMethodTypes {
		method = strings.ToLower(strings.TrimSpace(method))
		if method == "" {
			continue
		}
		methods = append(methods, method)
	}
	if len(methods) == 0 {
		methods = []string{"card"}
	}
	cfg.PaymentMethodTypes = methods

	if cfg.CheckoutExpireMinutes <= 0 {
		cfg.CheckoutExpireMinutes = 30
	}
	if cfg.RequestTimeoutSeconds <= 0 {
		cfg.RequestTimeoutSeconds = 45
	}
	return cfg
}

func decodeStripeSessionSnapshot(raw string) (*stripeSessionSnapshot, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false
	}

	var envelope stripeEventEnvelope
	if err := json.Unmarshal([]byte(raw), &envelope); err == nil && envelope.Data.Object.ID != "" {
		return &envelope.Data.Object, true
	}

	var snapshot stripeSessionSnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err == nil && isStripeSessionSnapshot(snapshot) {
		return &snapshot, true
	}

	return nil, false
}

func isStripeSessionSnapshot(snapshot stripeSessionSnapshot) bool {
	if strings.HasPrefix(strings.TrimSpace(snapshot.ID), "cs_") {
		return true
	}
	if strings.TrimSpace(snapshot.ClientReferenceID) != "" {
		return true
	}
	if len(snapshot.PaymentIntent) > 0 {
		return true
	}
	if snapshot.AmountTotal > 0 {
		return true
	}
	if strings.TrimSpace(snapshot.URL) != "" {
		return true
	}
	if strings.TrimSpace(snapshot.Status) != "" || strings.TrimSpace(snapshot.PaymentStatus) != "" {
		return true
	}
	return false
}

func extractPaymentNo(snapshot stripeSessionSnapshot) string {
	return firstNonEmpty(snapshot.ClientReferenceID, snapshot.Metadata["payment_no"])
}

func extractPaymentIntentCreated(intent *stripe.PaymentIntent) int64 {
	if intent == nil {
		return 0
	}
	return intent.Created
}

func extractIDFromRawMessage(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var id string
	if err := json.Unmarshal(raw, &id); err == nil {
		return strings.TrimSpace(id)
	}

	var obj struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil {
		return strings.TrimSpace(obj.ID)
	}

	return ""
}

func buildRefundIdempotencyKey(req *RefundRequest) string {
	if req == nil {
		return "refund"
	}

	parts := []string{
		"refund",
		strings.TrimSpace(req.PaymentNo),
		strings.TrimSpace(req.Amount),
	}
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		result = append(result, part)
	}
	if len(result) == 0 {
		return "refund"
	}
	return strings.Join(result, ":")
}

func wrapStripeRequestError(operation string, err error) error {
	if err == nil {
		return nil
	}

	lowered := strings.ToLower(err.Error())
	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(lowered, "context deadline exceeded") {
		return fmt.Errorf("stripe %s timeout: %w", operation, err)
	}
	return fmt.Errorf("stripe %s failed: %w", operation, err)
}

func buildCheckoutReturnURL(base string, params map[string]string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		return ""
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return base
	}

	query := parsed.Query()
	for key, value := range params {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		query.Set(key, value)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func decimalToMinor(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, errors.New("amount is required")
	}

	value, ok := new(big.Rat).SetString(raw)
	if !ok {
		return 0, errors.New("invalid amount")
	}
	if value.Sign() <= 0 {
		return 0, errors.New("amount must be greater than 0")
	}

	value.Mul(value, big.NewRat(100, 1))
	if value.Denom().Cmp(big.NewInt(1)) != 0 {
		return 0, errors.New("amount must keep 2 decimal places")
	}
	if !value.Num().IsInt64() {
		return 0, errors.New("amount is too large")
	}
	return value.Num().Int64(), nil
}

func minorToDecimal(amount int64) string {
	value := new(big.Rat).SetFrac64(amount, 100)
	return value.FloatString(2)
}

func formatUnix(unixValue int64) string {
	if unixValue <= 0 {
		return ""
	}
	return time.Unix(unixValue, 0).Format(stripeTimeLayout)
}

func firstNonZero(values ...int64) int64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func (s *StripeStrategy) String() string {
	return fmt.Sprintf("stripe(currency=%s, methods=%s)", s.config.Currency, strings.Join(s.config.PaymentMethodTypes, ","))
}
