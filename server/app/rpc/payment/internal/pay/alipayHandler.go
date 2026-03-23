package pay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"math/big"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"server/common"

	"github.com/smartwalle/alipay/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

type AlipayStrategy struct {
	client *alipay.Client
	config AlipayConfig
}

func NewAlipayStrategy(cfg AlipayConfig) (*AlipayStrategy, error) {
	if strings.TrimSpace(cfg.AppId) == "" {
		return nil, errors.New("alipay app id is empty")
	}
	if strings.TrimSpace(cfg.MerchantPrivateKey) == "" {
		return nil, errors.New("alipay merchant private key is empty")
	}
	if strings.TrimSpace(cfg.AlipayPublicKey) == "" {
		return nil, errors.New("alipay public key is empty")
	}

	opts := make([]alipay.OptionFunc, 0, 1)
	gatewayURL := strings.TrimSpace(cfg.GatewayUrl)
	if gatewayURL != "" {
		if cfg.Production {
			opts = append(opts, alipay.WithProductionGateway(gatewayURL))
		} else {
			opts = append(opts, alipay.WithSandboxGateway(gatewayURL))
		}
	}

	client, err := alipay.New(cfg.AppId, cfg.MerchantPrivateKey, cfg.Production, opts...)
	if err != nil {
		return nil, err
	}
	if err := client.LoadAliPayPublicKey(cfg.AlipayPublicKey); err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.ContentKey) != "" {
		if err := client.SetEncryptKey(cfg.ContentKey); err != nil {
			logx.Errorf("set alipay content key failed: %v", err)
		}
	}

	return &AlipayStrategy{
		client: client,
		config: cfg,
	}, nil
}

func (a *AlipayStrategy) Pay(outTradeNo, price, subject, notifyURL, returnURL string) (*PayResult, error) {
	param := alipay.TradePagePay{
		Trade: alipay.Trade{
			NotifyURL:   firstNonEmpty(notifyURL, a.config.NotifyUrl),
			ReturnURL:   firstNonEmpty(returnURL, a.config.ReturnUrl),
			Subject:     strings.TrimSpace(subject),
			OutTradeNo:  strings.TrimSpace(outTradeNo),
			TotalAmount: strings.TrimSpace(price),
			ProductCode: "FAST_INSTANT_TRADE_PAY",
		},
	}

	payURL, err := a.client.TradePagePay(param)
	if err != nil {
		return nil, err
	}

	return &PayResult{
		Success: true,
		Body:    buildAutoSubmitForm(payURL),
	}, nil
}

func (a *AlipayStrategy) SignVerify(params map[string]string) (bool, error) {
	values := mapToValues(params)
	if err := a.client.VerifySign(context.Background(), values); err != nil {
		return false, err
	}
	return true, nil
}

func (a *AlipayStrategy) DataVerify(params map[string]string, bill PayBillSnapshot) (bool, error) {
	notifyAmount, ok := normalizeAmount(params["total_amount"])
	if !ok {
		return false, nil
	}
	payAmount, ok := normalizeAmount(bill.PayAmount)
	if !ok {
		return false, nil
	}
	if notifyAmount != payAmount {
		return false, nil
	}

	notifySellerId := strings.TrimSpace(params["seller_id"])
	if notifySellerId == "" || notifySellerId != strings.TrimSpace(a.config.SellerId) {
		return false, nil
	}

	notifyAppId := strings.TrimSpace(params["app_id"])
	if notifyAppId == "" || notifyAppId != strings.TrimSpace(a.config.AppId) {
		return false, nil
	}

	tradeStatus := strings.ToUpper(strings.TrimSpace(params["trade_status"]))
	if tradeStatus != string(alipay.TradeStatusSuccess) {
		return false, nil
	}
	return true, nil
}

func (a *AlipayStrategy) QueryTrade(outTradeNo string) (*TradeResult, error) {
	result := &TradeResult{Success: false}
	rsp, err := a.client.TradeQuery(context.Background(), alipay.TradeQuery{
		OutTradeNo: strings.TrimSpace(outTradeNo),
	})
	if err != nil {
		logx.Errorf("alipay query trade failed, out_trade_no=%s err=%v", outTradeNo, err)
		return result, nil
	}
	if rsp == nil || !rsp.IsSuccess() {
		if rsp != nil {
			logx.Errorf("alipay query trade response failed, out_trade_no=%s code=%s msg=%s sub_code=%s sub_msg=%s",
				outTradeNo, rsp.Code, rsp.Msg, rsp.SubCode, rsp.SubMsg)
		}
		return result, nil
	}

	status, ok := convertPayBillStatus(rsp.TradeStatus)
	if !ok {
		logx.Errorf("alipay trade status not supported, out_trade_no=%s trade_status=%s", outTradeNo, rsp.TradeStatus)
		return result, nil
	}

	result.Success = true
	result.OutTradeNo = strings.TrimSpace(rsp.OutTradeNo)
	result.TradeNo = strings.TrimSpace(rsp.TradeNo)
	result.TotalAmount = strings.TrimSpace(rsp.TotalAmount)
	result.PayBillStatus = status
	result.PaidAt = strings.TrimSpace(rsp.SendPayDate)
	return result, nil
}

func (a *AlipayStrategy) Refund(outTradeNo, amount, reason string) (*RefundResult, error) {
	rsp, err := a.client.TradeRefund(context.Background(), alipay.TradeRefund{
		OutTradeNo:   strings.TrimSpace(outTradeNo),
		RefundAmount: strings.TrimSpace(amount),
		RefundReason: strings.TrimSpace(reason),
		OutRequestNo: strconv.FormatInt(common.GenerateId(), 10),
	})
	if err != nil {
		return nil, err
	}
	if rsp == nil {
		return &RefundResult{
			Success: false,
			Message: "alipay refund response is empty",
		}, nil
	}

	bodyBytes, _ := json.Marshal(rsp)
	message := strings.TrimSpace(rsp.SubMsg)
	if message == "" {
		message = strings.TrimSpace(rsp.Msg)
	}
	if message == "" && !rsp.IsSuccess() {
		message = "refund failed"
	}

	return &RefundResult{
		Success: rsp.IsSuccess(),
		Body:    string(bodyBytes),
		Message: message,
		TradeNo: strings.TrimSpace(rsp.TradeNo),
	}, nil
}

func (a *AlipayStrategy) Channel() string {
	return ChannelAlipay
}

func convertPayBillStatus(status alipay.TradeStatus) (int32, bool) {
	switch status {
	case alipay.TradeStatusWaitBuyerPay:
		return PayBillStatusNoPay, true
	case alipay.TradeStatusClosed:
		return PayBillStatusCancel, true
	case alipay.TradeStatusSuccess, alipay.TradeStatusFinished:
		return PayBillStatusPay, true
	default:
		return 0, false
	}
}

func mapToValues(params map[string]string) url.Values {
	values := make(url.Values, len(params))
	for key, value := range params {
		values.Set(key, value)
	}
	return values
}

func buildAutoSubmitForm(payURL *url.URL) string {
	if payURL == nil {
		return ""
	}

	actionURL := (&url.URL{
		Scheme: payURL.Scheme,
		Host:   payURL.Host,
		Path:   payURL.Path,
	}).String()
	values := payURL.Query()
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	builder.WriteString(`<form id="alipaysubmit" name="alipaysubmit" action="`)
	builder.WriteString(html.EscapeString(actionURL))
	builder.WriteString(`" method="POST">`)
	for _, key := range keys {
		for _, value := range values[key] {
			builder.WriteString(`<input type="hidden" name="`)
			builder.WriteString(html.EscapeString(key))
			builder.WriteString(`" value="`)
			builder.WriteString(html.EscapeString(value))
			builder.WriteString(`"/>`)
		}
	}
	builder.WriteString(`<input type="submit" value="ok" style="display:none;"/></form>`)
	builder.WriteString(`<script>document.forms['alipaysubmit'].submit();</script>`)
	return builder.String()
}

func normalizeAmount(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	value, ok := new(big.Rat).SetString(raw)
	if !ok {
		return "", false
	}
	if value.Sign() < 0 {
		return "", false
	}
	return value.FloatString(2), true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func (a *AlipayStrategy) String() string {
	return fmt.Sprintf("alipay(appId=%s)", a.config.AppId)
}
