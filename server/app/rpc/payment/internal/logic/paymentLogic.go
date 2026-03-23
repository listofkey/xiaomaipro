package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"server/app/rpc/model"
	"server/app/rpc/payment/internal/pay"
	"server/app/rpc/payment/internal/svc"
	"server/common"
	"server/app/rpc/payment/paymentpb"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	orderStatusPendingPay int16 = 1
	orderStatusCancelled  int16 = 2
	orderStatusPaid       int16 = 3
	orderStatusRefunded   int16 = 4
	orderStatusCompleted  int16 = 5

	cancelReasonRefund int16 = 3

	paymentStatusPending int16 = 0
	paymentStatusSuccess int16 = 1
	paymentStatusFailed  int16 = 2

	refundStatusPending    int16 = 1
	refundStatusProcessing int16 = 2
	refundStatusSuccess    int16 = 3
	refundStatusFailed     int16 = 4

	orderTicketStatusVoided int16 = 2

	lockNamePayOrder   = "pay_order"
	lockNameNotify     = "notify"
	lockNameTradeCheck = "trade_check"
	lockNameRefund     = "refund"

	timeLayout = "2006-01-02 15:04:05"
	refundOperationTimeout = 2 * time.Minute
)

var releaseLockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
end
return 0
`)

type PaymentCore struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPaymentCore(ctx context.Context, svcCtx *svc.ServiceContext) *PaymentCore {
	return &PaymentCore{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PaymentCore) PayOrder(in *paymentpb.PayOrderReq) (*paymentpb.PayOrderResp, error) {
	if in == nil {
		return nil, errors.New("request is empty")
	}
	if in.UserId <= 0 {
		return nil, errors.New("user_id is required")
	}

	orderNo := strings.TrimSpace(in.OrderNo)
	if orderNo == "" {
		return nil, errors.New("order_no is required")
	}

	channel := l.resolveChannel(in.Channel)
	strategy, err := l.svcCtx.PayStrategies.Get(channel)
	if err != nil {
		return nil, err
	}

	resp := &paymentpb.PayOrderResp{}
	err = l.withLock(lockNamePayOrder, orderNo, func() error {
		order, payment, err := l.preparePayOrder(in)
		if err != nil {
			return err
		}

		resp.OrderStatus = int32(order.Status)
		resp.PaidAt = formatTimeValue(order.PaidAt)
		resp.Payment = paymentToProto(payment)

		if order.Status == orderStatusPaid || order.Status == orderStatusCompleted || payment.Status == paymentStatusSuccess {
			resp.Success = true
			return nil
		}

		if checkoutURL, sessionID, expiresAt, ok := extractCheckoutSessionInfo(payment); ok {
			resp.Success = true
			resp.PayForm = checkoutURL
			resp.CheckoutUrl = checkoutURL
			resp.CheckoutSessionId = sessionID
			resp.SessionExpiresAt = expiresAt
			return nil
		}

		subject := strings.TrimSpace(in.Subject)
		if subject == "" {
			subject = fmt.Sprintf("xiaomaipro order %s", orderNo)
		}

		result, err := strategy.Pay(&pay.PayRequest{
			PaymentNo:     payment.PaymentNo,
			OrderNo:       order.OrderNo,
			Amount:        formatAmount(payment.Amount),
			Subject:       subject,
			SuccessURL:    l.resolveStripeSuccessURL(in.ReturnUrl),
			CancelURL:     l.resolveStripeCancelURL(in.NotifyUrl),
			Currency:      l.svcCtx.Config.Stripe.Currency,
			CustomerEmail: l.findUserEmail(order.UserID),
			Metadata: map[string]string{
				"user_id": strconv.FormatInt(order.UserID, 10),
			},
		})
		if err != nil {
			return err
		}
		if err := l.updatePendingPaymentSession(payment, result); err != nil {
			return err
		}

		resp.Success = result != nil && result.Success
		if result != nil {
			resp.PayForm = result.Body
			resp.CheckoutUrl = result.CheckoutURL
			resp.CheckoutSessionId = result.SessionID
			resp.SessionExpiresAt = result.ExpiresAt
			resp.Payment = paymentToProto(payment)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (l *PaymentCore) Notify(in *paymentpb.NotifyReq) (*paymentpb.NotifyResp, error) {
	resp := &paymentpb.NotifyResp{
		Success: false,
		AckText: l.svcCtx.Config.Pay.NotifyFailureResult,
	}
	if in == nil {
		return resp, nil
	}

	channel := l.resolveChannel(in.Channel)
	strategy, err := l.svcCtx.PayStrategies.Get(channel)
	if err != nil {
		return nil, err
	}

	notifyResult, err := strategy.ParseNotify(&pay.NotifyRequest{
		RawBody:   in.RawBody,
		Headers:   in.Headers,
		Signature: in.Signature,
	})
	if err != nil {
		return nil, err
	}
	if notifyResult == nil {
		return resp, nil
	}
	if !notifyResult.Handled {
		resp.Success = true
		resp.AckText = l.svcCtx.Config.Pay.NotifySuccessResult
		return resp, nil
	}

	paymentNo := strings.TrimSpace(notifyResult.PaymentNo)
	if paymentNo == "" {
		resp.Success = true
		resp.AckText = l.svcCtx.Config.Pay.NotifySuccessResult
		return resp, nil
	}

	tradeNo := strings.TrimSpace(notifyResult.TradeNo)
	paidAt := parseThirdPartyTime(notifyResult.PaidAt)
	callbackData := firstNonEmpty(notifyResult.RawData, mustJSON(notifyResult))

	err = l.withLock(lockNameNotify, paymentNo, func() error {
		var order *model.OrderInfo
		var payment *model.Payment

		err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
			var err error
			payment, err = l.loadPaymentByNoTx(tx, paymentNo, true)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			if err != nil {
				return err
			}

			order, err = l.loadOrderByIDTx(tx, payment.OrderID, true)
			if err != nil {
				return err
			}

			if notifyResult.TotalAmount != "" {
				equal, err := amountEqual(formatAmount(payment.Amount), notifyResult.TotalAmount)
				if err != nil {
					l.Errorf("notify amount parse failed, payment_no=%s err=%v", paymentNo, err)
					return nil
				}
				if !equal {
					l.Errorf("notify amount mismatch, payment_no=%s notify_amount=%s db_amount=%s",
						paymentNo, notifyResult.TotalAmount, formatAmount(payment.Amount))
					return nil
				}
			}

			if payment.Status == paymentStatusSuccess {
				return nil
			}

			switch notifyResult.PayBillStatus {
			case pay.PayBillStatusPay:
				if order.Status == orderStatusCancelled {
					refundResult, err := strategy.Refund(&pay.RefundRequest{
						PaymentNo:    payment.PaymentNo,
						TradeNo:      firstNonEmpty(tradeNo, payment.TradeNo),
						CallbackData: callbackData,
						Amount:       formatAmount(payment.Amount),
						Reason:       "late payment auto refund",
						Metadata: map[string]string{
							"order_no": order.OrderNo,
						},
					})
					if err != nil {
						return err
					}
					if refundResult == nil || !refundResult.Success {
						if refundResult != nil && strings.TrimSpace(refundResult.Message) != "" {
							return errors.New(refundResult.Message)
						}
						return errors.New("late payment refund failed")
					}
					return l.applyLatePaymentRefundTx(
						tx,
						payment,
						order,
						firstNonEmpty(tradeNo, payment.TradeNo),
						paidAt,
						callbackData,
						"late payment auto refund",
						refundResult.TradeNo,
						refundResult.Pending,
					)
				}

				if order.Status == orderStatusRefunded {
					if err := l.updatePaymentSuccessOnlyTx(tx, payment, firstNonEmpty(tradeNo, payment.TradeNo), paidAt, callbackData); err != nil {
						return err
					}
					payment.Status = paymentStatusSuccess
					payment.TradeNo = firstNonEmpty(tradeNo, payment.TradeNo)
					payment.PaidAt = derefTime(paidAt)
					payment.CallbackData = callbackData
					return nil
				}

				return l.applyPaymentSuccessTx(tx, payment, order, firstNonEmpty(tradeNo, payment.TradeNo), paidAt, callbackData)
			case pay.PayBillStatusCancel:
				if payment.Status == paymentStatusPending {
					now := time.Now()
					if err := tx.Model(&model.Payment{}).
						Where("id = ?", payment.ID).
						Updates(map[string]any{
							"status":        paymentStatusFailed,
							"trade_no":      firstNonEmpty(tradeNo, payment.TradeNo),
							"callback_data": callbackData,
							"updated_at":    now,
						}).Error; err != nil {
						return err
					}
					payment.Status = paymentStatusFailed
					payment.TradeNo = firstNonEmpty(tradeNo, payment.TradeNo)
					payment.CallbackData = callbackData
					payment.UpdatedAt = now
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		if payment == nil || order == nil {
			resp.Success = true
			resp.AckText = l.svcCtx.Config.Pay.NotifySuccessResult
			return nil
		}

		resp.Success = true
		resp.AckText = l.svcCtx.Config.Pay.NotifySuccessResult
		resp.OrderNo = order.OrderNo
		resp.OrderStatus = int32(order.Status)
		resp.PaidAt = formatTimeValue(order.PaidAt)
		resp.Payment = paymentToProto(payment)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (l *PaymentCore) TradeCheck(in *paymentpb.TradeCheckReq) (*paymentpb.TradeCheckResp, error) {
	if in == nil {
		return nil, errors.New("request is empty")
	}
	if in.UserId <= 0 {
		return nil, errors.New("user_id is required")
	}

	orderNo := strings.TrimSpace(in.OrderNo)
	paymentNo := strings.TrimSpace(in.PaymentNo)
	if orderNo == "" && paymentNo == "" {
		return nil, errors.New("order_no or payment_no is required")
	}

	channel := l.resolveChannel(in.Channel)
	strategy, err := l.svcCtx.PayStrategies.Get(channel)
	if err != nil {
		return nil, err
	}

	resp := &paymentpb.TradeCheckResp{}
	lockKey := firstNonEmpty(paymentNo, orderNo)
	err = l.withLock(lockNameTradeCheck, lockKey, func() error {
		order, payment, err := l.findOrderAndPaymentForCheck(in)
		if err != nil {
			return err
		}
		if payment == nil {
			return nil
		}

		result, err := strategy.QueryTrade(&pay.TradeQueryRequest{
			PaymentNo:    payment.PaymentNo,
			TradeNo:      payment.TradeNo,
			CallbackData: payment.CallbackData,
		})
		if err != nil {
			return err
		}
		if result == nil || !result.Success {
			resp.Success = false
			resp.Paid = payment.Status == paymentStatusSuccess
			resp.Payment = paymentToProto(payment)
			resp.CheckoutSessionId = payment.TradeNo
			if order != nil {
				resp.OrderStatus = int32(order.Status)
				resp.PaidAt = formatTimeValue(order.PaidAt)
			}
			return nil
		}

		if equal, err := amountEqual(formatAmount(payment.Amount), result.TotalAmount); err != nil || !equal {
			if err != nil {
				l.Errorf("trade check amount parse failed, payment_no=%s err=%v", payment.PaymentNo, err)
			} else {
				l.Errorf("trade check amount mismatch, payment_no=%s query_amount=%s db_amount=%s",
					payment.PaymentNo, result.TotalAmount, formatAmount(payment.Amount))
			}
			resp.Success = false
			resp.Paid = payment.Status == paymentStatusSuccess
			resp.Payment = paymentToProto(payment)
			if order != nil {
				resp.OrderStatus = int32(order.Status)
				resp.PaidAt = formatTimeValue(order.PaidAt)
			}
			return nil
		}

		paidAt := parseThirdPartyTime(result.PaidAt)
		callbackData := firstNonEmpty(result.RawData, mustJSON(result))

		err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
			order, payment, err = l.reloadOrderAndPaymentByOrderNoTx(tx, in.UserId, orderNo, payment.PaymentNo)
			if err != nil {
				return err
			}
			if payment == nil || order == nil {
				return nil
			}

			switch result.PayBillStatus {
			case pay.PayBillStatusPay:
				if payment.Status == paymentStatusSuccess {
					return nil
				}
				if order.Status == orderStatusCancelled {
					refundResult, err := strategy.Refund(&pay.RefundRequest{
						PaymentNo:    payment.PaymentNo,
						TradeNo:      firstNonEmpty(result.TradeNo, payment.TradeNo),
						CallbackData: callbackData,
						Amount:       formatAmount(payment.Amount),
						Reason:       "trade check auto refund",
						Metadata: map[string]string{
							"order_no": order.OrderNo,
						},
					})
					if err != nil {
						return err
					}
					if refundResult == nil || !refundResult.Success {
						if refundResult != nil && strings.TrimSpace(refundResult.Message) != "" {
							return errors.New(refundResult.Message)
						}
						return errors.New("trade check auto refund failed")
					}
					return l.applyLatePaymentRefundTx(
						tx,
						payment,
						order,
						firstNonEmpty(result.TradeNo, payment.TradeNo),
						paidAt,
						callbackData,
						"trade check auto refund",
						refundResult.TradeNo,
						refundResult.Pending,
					)
				}
				if order.Status == orderStatusRefunded {
					if err := l.updatePaymentSuccessOnlyTx(tx, payment, firstNonEmpty(result.TradeNo, payment.TradeNo), paidAt, callbackData); err != nil {
						return err
					}
					payment.Status = paymentStatusSuccess
					payment.TradeNo = firstNonEmpty(result.TradeNo, payment.TradeNo)
					payment.PaidAt = derefTime(paidAt)
					payment.CallbackData = callbackData
					return nil
				}
				return l.applyPaymentSuccessTx(tx, payment, order, firstNonEmpty(result.TradeNo, payment.TradeNo), paidAt, callbackData)
			case pay.PayBillStatusCancel:
				if payment.Status == paymentStatusPending {
					now := time.Now()
					if err := tx.Model(&model.Payment{}).
						Where("id = ?", payment.ID).
						Updates(map[string]any{
							"status":        paymentStatusFailed,
							"trade_no":      firstNonEmpty(result.TradeNo, payment.TradeNo),
							"callback_data": callbackData,
							"updated_at":    now,
						}).Error; err != nil {
						return err
					}
					payment.Status = paymentStatusFailed
					payment.TradeNo = firstNonEmpty(result.TradeNo, payment.TradeNo)
					payment.CallbackData = callbackData
					payment.UpdatedAt = now
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		resp.Success = true
		resp.Paid = payment != nil && payment.Status == paymentStatusSuccess
		resp.Payment = paymentToProto(payment)
		if payment != nil {
			resp.CheckoutSessionId = firstNonEmpty(result.TradeNo, payment.TradeNo)
		} else {
			resp.CheckoutSessionId = result.TradeNo
		}
		if order != nil {
			resp.OrderStatus = int32(order.Status)
			resp.PaidAt = formatTimeValue(order.PaidAt)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (l *PaymentCore) Refund(in *paymentpb.RefundReq) (*paymentpb.RefundResp, error) {
	if in == nil {
		return nil, errors.New("request is empty")
	}
	if in.UserId <= 0 {
		return nil, errors.New("user_id is required")
	}

	orderNo := strings.TrimSpace(in.OrderNo)
	if orderNo == "" {
		return nil, errors.New("order_no is required")
	}

	channel := l.resolveChannel(in.Channel)
	strategy, err := l.svcCtx.PayStrategies.Get(channel)
	if err != nil {
		return nil, err
	}

	resp := &paymentpb.RefundResp{
		Success: false,
		OrderNo: orderNo,
	}
	opCtx, cancel := l.refundOperationContext()
	defer cancel()

	core := l.withContext(opCtx)
	err = core.withLock(lockNameRefund, orderNo, func() error {
		order, payment, existedRefund, refundAmount, reason, err := core.prepareRefund(in)
		if err != nil {
			return err
		}
		if existedRefund != nil {
			resp.Success = true
			resp.OrderStatus = int32(orderStatusRefunded)
			resp.Refund = refundToProto(existedRefund)
			resp.ShouldRelease = false
			return nil
		}

		refundResult, err := strategy.Refund(&pay.RefundRequest{
			Context:      opCtx,
			PaymentNo:    payment.PaymentNo,
			TradeNo:      payment.TradeNo,
			CallbackData: payment.CallbackData,
			Amount:       refundAmount,
			Reason:       reason,
			Metadata: map[string]string{
				"order_no": order.OrderNo,
			},
		})
		if err != nil {
			return err
		}
		if refundResult == nil || !refundResult.Success {
			if refundResult != nil && strings.TrimSpace(refundResult.Message) != "" {
				return errors.New(refundResult.Message)
			}
			return errors.New("refund failed")
		}

		var createdRefund *model.Refund
		err = core.svcCtx.DB.WithContext(core.ctx).Transaction(func(tx *gorm.DB) error {
			order, payment, createdRefund, err = core.applyRefundTx(tx, in.UserId, orderNo, refundAmount, reason, refundResult)
			return err
		})
		if err != nil {
			return err
		}

		resp.Success = true
		resp.OrderStatus = int32(order.Status)
		resp.Refund = refundToProto(createdRefund)
		resp.ShouldRelease = true
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (l *PaymentCore) Detail(in *paymentpb.DetailReq) (*paymentpb.DetailResp, error) {
	if in == nil {
		return nil, errors.New("request is empty")
	}
	if in.UserId <= 0 {
		return nil, errors.New("user_id is required")
	}

	orderNo := strings.TrimSpace(in.OrderNo)
	if orderNo == "" {
		return nil, errors.New("order_no is required")
	}

	order, err := l.findOrderByNo(in.UserId, orderNo)
	if err != nil {
		return nil, err
	}

	payment, err := l.findLatestPayment(order.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	refund, err := l.findLatestRefund(order.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &paymentpb.DetailResp{
		Payment:     paymentToProto(payment),
		Refund:      refundToProto(refund),
		OrderStatus: int32(order.Status),
		PaidAt:      formatTimeValue(order.PaidAt),
	}, nil
}

func (l *PaymentCore) preparePayOrder(in *paymentpb.PayOrderReq) (*model.OrderInfo, *model.Payment, error) {
	var (
		order   *model.OrderInfo
		payment *model.Payment
	)

	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		order, err = l.loadOrderByNoTx(tx, in.UserId, strings.TrimSpace(in.OrderNo), true)
		if err != nil {
			return err
		}

		switch order.Status {
		case orderStatusPaid, orderStatusCompleted:
			payment, err = l.loadLatestPaymentTx(tx, order.ID, true)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		case orderStatusCancelled:
			return errors.New("order already cancelled")
		case orderStatusRefunded:
			return errors.New("order already refunded")
		case orderStatusPendingPay:
		default:
			return errors.New("order status is invalid")
		}

		now := time.Now()
		if !order.PayDeadline.IsZero() && now.After(order.PayDeadline) {
			return errors.New("order payment timeout")
		}

		payment, err = l.loadLatestPaymentTx(tx, order.ID, true)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			payment, err = l.createPendingPaymentTx(tx, order, normalizePayMethod(in.PayMethod))
			return err
		}
		if err != nil {
			return err
		}

		switch payment.Status {
		case paymentStatusSuccess:
			return nil
		case paymentStatusFailed:
			payment, err = l.createPendingPaymentTx(tx, order, normalizePayMethod(in.PayMethod))
			return err
		case paymentStatusPending:
			method := int16(normalizePayMethod(in.PayMethod))
			if method == 0 {
				method = payment.PayMethod
			}
			if method == 0 {
				method = 1
			}
			if payment.PayMethod != method {
				if err := tx.Model(&model.Payment{}).
					Where("id = ?", payment.ID).
					Updates(map[string]any{
						"pay_method": method,
						"updated_at": now,
					}).Error; err != nil {
					return err
				}
				payment.PayMethod = method
				payment.UpdatedAt = now
			}
			return nil
		default:
			return errors.New("payment status is invalid")
		}
	})
	if err != nil {
		return nil, nil, err
	}

	if payment == nil {
		payment = &model.Payment{}
	}
	return order, payment, nil
}

func (l *PaymentCore) prepareRefund(in *paymentpb.RefundReq) (*model.OrderInfo, *model.Payment, *model.Refund, string, string, error) {
	var (
		order         *model.OrderInfo
		payment       *model.Payment
		existedRefund *model.Refund
		refundAmount  string
		reason        string
	)

	reason = strings.TrimSpace(in.Reason)
	if reason == "" {
		reason = "user apply"
	}

	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		order, err = l.loadOrderByNoTx(tx, in.UserId, strings.TrimSpace(in.OrderNo), true)
		if err != nil {
			return err
		}

		if order.Status == orderStatusRefunded {
			existedRefund, err = l.loadLatestRefundTx(tx, order.ID, true)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("order already refunded")
			}
			return err
		}

		if order.Status != orderStatusPaid && order.Status != orderStatusCompleted {
			return errors.New("only paid orders can be refunded")
		}

		var event model.Event
		if err := tx.Where("id = ?", order.EventID).First(&event).Error; err != nil {
			return err
		}
		if err := l.checkRefundWindow(&event); err != nil {
			return err
		}

		payment, err = l.loadLatestPaymentTx(tx, order.ID, true)
		if err != nil {
			return err
		}
		if payment.Status != paymentStatusSuccess {
			return errors.New("payment is not successful")
		}

		existedRefund, err = l.loadLatestRefundTx(tx, order.ID, true)
		if err == nil && existedRefund != nil && isTerminalRefund(existedRefund.Status) {
			return nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var tier model.TicketTier
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", order.TicketTierID).
			First(&tier).Error; err != nil {
			return err
		}
		if tier.SoldCount < order.Quantity {
			return errors.New("sold inventory is insufficient")
		}

		amount := in.RefundAmount
		if amount <= 0 {
			amount = order.TotalAmount
		}
		refundAmount = formatAmount(amount)

		greater, err := amountGreater(refundAmount, formatAmount(payment.Amount))
		if err != nil {
			return err
		}
		if greater {
			return errors.New("refund amount greater than pay amount")
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, "", "", err
	}

	return order, payment, existedRefund, refundAmount, reason, nil
}

func (l *PaymentCore) applyRefundTx(tx *gorm.DB, userID int64, orderNo, refundAmount, reason string, refundResult *pay.RefundResult) (*model.OrderInfo, *model.Payment, *model.Refund, error) {
	order, err := l.loadOrderByNoTx(tx, userID, orderNo, true)
	if err != nil {
		return nil, nil, nil, err
	}

	if order.Status == orderStatusRefunded {
		refund, err := l.loadLatestRefundTx(tx, order.ID, true)
		if err != nil {
			return order, nil, nil, err
		}
		payment, err := l.loadLatestPaymentTx(tx, order.ID, true)
		if err != nil {
			return order, nil, nil, err
		}
		return order, payment, refund, nil
	}

	if order.Status != orderStatusPaid && order.Status != orderStatusCompleted {
		return nil, nil, nil, errors.New("only paid orders can be refunded")
	}

	payment, err := l.loadLatestPaymentTx(tx, order.ID, true)
	if err != nil {
		return nil, nil, nil, err
	}
	if payment.Status != paymentStatusSuccess {
		return nil, nil, nil, errors.New("payment is not successful")
	}

	var tier model.TicketTier
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", order.TicketTierID).
		First(&tier).Error; err != nil {
		return nil, nil, nil, err
	}
	if tier.SoldCount < order.Quantity {
		return nil, nil, nil, errors.New("sold inventory is insufficient")
	}

	now := time.Now()
	refundAmountValue, _ := strconv.ParseFloat(refundAmount, 64)
	refund := &model.Refund{
		ID:           common.GenerateId(),
		RefundNo:     fmt.Sprintf("R%d", common.GenerateId()),
		OrderID:      order.ID,
		PaymentID:    payment.ID,
		UserID:       order.UserID,
		RefundAmount: refundAmountValue,
		Status:       resolveRefundStatus(refundResult),
		Reason:       reason,
		TradeNo:      firstNonEmpty(strings.TrimSpace(refundResult.TradeNo), payment.TradeNo),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if refund.Status == refundStatusSuccess {
		refund.RefundedAt = now
	}

	if err := tx.Create(refund).Error; err != nil {
		return nil, nil, nil, err
	}

	if err := tx.Model(&model.OrderInfo{}).
		Where("id = ?", order.ID).
		Updates(map[string]any{
			"status":        orderStatusRefunded,
			"cancel_reason": cancelReasonRefund,
			"cancelled_at":  now,
			"updated_at":    now,
		}).Error; err != nil {
		return nil, nil, nil, err
	}

	if err := tx.Model(&model.OrderTicket{}).
		Where("order_id = ?", order.ID).
		Update("status", orderTicketStatusVoided).Error; err != nil {
		return nil, nil, nil, err
	}

	if err := tx.Model(&model.TicketTier{}).
		Where("id = ? AND sold_count >= ?", tier.ID, order.Quantity).
		Update("sold_count", gorm.Expr("sold_count - ?", order.Quantity)).Error; err != nil {
		return nil, nil, nil, err
	}

	order.Status = orderStatusRefunded
	order.CancelReason = cancelReasonRefund
	order.CancelledAt = now
	order.UpdatedAt = now

	return order, payment, refund, nil
}

func (l *PaymentCore) applyPaymentSuccessTx(tx *gorm.DB, payment *model.Payment, order *model.OrderInfo, tradeNo string, paidAt *time.Time, callbackData string) error {
	if payment == nil || order == nil {
		return errors.New("payment or order is nil")
	}

	if order.Status == orderStatusPaid || order.Status == orderStatusCompleted {
		return l.updatePaymentSuccessOnlyTx(tx, payment, tradeNo, paidAt, callbackData)
	}
	if order.Status != orderStatusPendingPay {
		return fmt.Errorf("order status %d does not support payment success", order.Status)
	}

	var tier model.TicketTier
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", order.TicketTierID).
		First(&tier).Error; err != nil {
		return err
	}
	if tier.LockedCount < order.Quantity {
		return errors.New("locked inventory is insufficient")
	}

	now := time.Now()
	updatePaidAt := chooseTime(paidAt, &now)
	if err := tx.Model(&model.OrderInfo{}).
		Where("id = ?", order.ID).
		Updates(map[string]any{
			"status":     orderStatusPaid,
			"paid_at":    updatePaidAt,
			"updated_at": now,
		}).Error; err != nil {
		return err
	}

	if err := tx.Model(&model.Payment{}).
		Where("id = ?", payment.ID).
		Updates(map[string]any{
			"status":        paymentStatusSuccess,
			"trade_no":      tradeNo,
			"paid_at":       updatePaidAt,
			"callback_data": callbackData,
			"updated_at":    now,
		}).Error; err != nil {
		return err
	}

	if err := tx.Model(&model.TicketTier{}).
		Where("id = ? AND locked_count >= ?", tier.ID, order.Quantity).
		Updates(map[string]any{
			"locked_count": gorm.Expr("locked_count - ?", order.Quantity),
			"sold_count":   gorm.Expr("sold_count + ?", order.Quantity),
		}).Error; err != nil {
		return err
	}

	order.Status = orderStatusPaid
	order.PaidAt = derefTime(updatePaidAt)
	order.UpdatedAt = now
	payment.Status = paymentStatusSuccess
	payment.TradeNo = tradeNo
	payment.PaidAt = derefTime(updatePaidAt)
	payment.CallbackData = callbackData
	payment.UpdatedAt = now
	return nil
}

func (l *PaymentCore) updatePaymentSuccessOnlyTx(tx *gorm.DB, payment *model.Payment, tradeNo string, paidAt *time.Time, callbackData string) error {
	if payment == nil {
		return errors.New("payment is nil")
	}

	now := time.Now()
	updatePaidAt := chooseTime(paidAt, timePtrIfNotZero(payment.PaidAt))
	if err := tx.Model(&model.Payment{}).
		Where("id = ?", payment.ID).
		Updates(map[string]any{
			"status":        paymentStatusSuccess,
			"trade_no":      firstNonEmpty(tradeNo, payment.TradeNo),
			"paid_at":       updatePaidAt,
			"callback_data": callbackData,
			"updated_at":    now,
		}).Error; err != nil {
		return err
	}

	payment.Status = paymentStatusSuccess
	payment.TradeNo = firstNonEmpty(tradeNo, payment.TradeNo)
	payment.PaidAt = derefTime(updatePaidAt)
	payment.CallbackData = callbackData
	payment.UpdatedAt = now
	return nil
}

func (l *PaymentCore) applyLatePaymentRefundTx(tx *gorm.DB, payment *model.Payment, order *model.OrderInfo, tradeNo string, paidAt *time.Time, callbackData, reason, refundTradeNo string, refundPending bool) error {
	if err := l.updatePaymentSuccessOnlyTx(tx, payment, tradeNo, paidAt, callbackData); err != nil {
		return err
	}

	refund, err := l.loadLatestRefundTx(tx, order.ID, true)
	if err == nil && refund != nil && isTerminalRefund(refund.Status) {
		order.Status = orderStatusRefunded
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	now := time.Now()
	refundRecord := &model.Refund{
		ID:           common.GenerateId(),
		RefundNo:     fmt.Sprintf("R%d", common.GenerateId()),
		OrderID:      order.ID,
		PaymentID:    payment.ID,
		UserID:       order.UserID,
		RefundAmount: payment.Amount,
		Status:       refundStatusFromPending(refundPending),
		Reason:       reason,
		TradeNo:      firstNonEmpty(strings.TrimSpace(refundTradeNo), payment.TradeNo, tradeNo),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if refundRecord.Status == refundStatusSuccess {
		refundRecord.RefundedAt = now
	}
	if err := tx.Create(refundRecord).Error; err != nil {
		return err
	}

	if err := tx.Model(&model.OrderInfo{}).
		Where("id = ?", order.ID).
		Updates(map[string]any{
			"status":        orderStatusRefunded,
			"cancel_reason": cancelReasonRefund,
			"updated_at":    now,
		}).Error; err != nil {
		return err
	}

	order.Status = orderStatusRefunded
	order.CancelReason = cancelReasonRefund
	order.UpdatedAt = now
	return nil
}

func (l *PaymentCore) findOrderAndPaymentForCheck(in *paymentpb.TradeCheckReq) (*model.OrderInfo, *model.Payment, error) {
	if strings.TrimSpace(in.OrderNo) != "" {
		order, err := l.findOrderByNo(in.UserId, strings.TrimSpace(in.OrderNo))
		if err != nil {
			return nil, nil, err
		}

		payment, err := l.findLatestPayment(order.ID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return order, nil, nil
		}
		if err != nil {
			return nil, nil, err
		}
		return order, payment, nil
	}

	var payment model.Payment
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("payment_no = ?", strings.TrimSpace(in.PaymentNo)).
		Order("created_at desc").
		First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	order, err := l.findOrderByID(payment.OrderID)
	if err != nil {
		return nil, nil, err
	}
	if order.UserID != in.UserId {
		return nil, nil, errors.New("order not found")
	}
	return order, &payment, nil
}

func (l *PaymentCore) reloadOrderAndPaymentByOrderNoTx(tx *gorm.DB, userID int64, orderNo, paymentNo string) (*model.OrderInfo, *model.Payment, error) {
	var (
		order   *model.OrderInfo
		payment *model.Payment
		err     error
	)

	if strings.TrimSpace(orderNo) != "" {
		order, err = l.loadOrderByNoTx(tx, userID, orderNo, true)
		if err != nil {
			return nil, nil, err
		}
	} else {
		payment, err = l.loadPaymentByNoTx(tx, paymentNo, true)
		if err != nil {
			return nil, nil, err
		}
		order, err = l.loadOrderByIDTx(tx, payment.OrderID, true)
		if err != nil {
			return nil, nil, err
		}
		if order.UserID != userID {
			return nil, nil, errors.New("order not found")
		}
		return order, payment, nil
	}

	if strings.TrimSpace(paymentNo) != "" {
		payment, err = l.loadPaymentByNoTx(tx, paymentNo, true)
		if err != nil {
			return nil, nil, err
		}
	} else {
		payment, err = l.loadLatestPaymentTx(tx, order.ID, true)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return order, nil, nil
		}
		if err != nil {
			return nil, nil, err
		}
	}

	return order, payment, nil
}

func (l *PaymentCore) findOrderByNo(userID int64, orderNo string) (*model.OrderInfo, error) {
	return l.loadOrderByNoTx(l.svcCtx.DB.WithContext(l.ctx), userID, orderNo, false)
}

func (l *PaymentCore) findOrderByID(orderID int64) (*model.OrderInfo, error) {
	return l.loadOrderByIDTx(l.svcCtx.DB.WithContext(l.ctx), orderID, false)
}

func (l *PaymentCore) findLatestPayment(orderID int64) (*model.Payment, error) {
	return l.loadLatestPaymentTx(l.svcCtx.DB.WithContext(l.ctx), orderID, false)
}

func (l *PaymentCore) findLatestRefund(orderID int64) (*model.Refund, error) {
	return l.loadLatestRefundTx(l.svcCtx.DB.WithContext(l.ctx), orderID, false)
}

func (l *PaymentCore) loadOrderByNoTx(tx *gorm.DB, userID int64, orderNo string, forUpdate bool) (*model.OrderInfo, error) {
	var order model.OrderInfo
	query := tx
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return &order, nil
}

func (l *PaymentCore) loadOrderByIDTx(tx *gorm.DB, orderID int64, forUpdate bool) (*model.OrderInfo, error) {
	var order model.OrderInfo
	query := tx
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Where("id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (l *PaymentCore) loadLatestPaymentTx(tx *gorm.DB, orderID int64, forUpdate bool) (*model.Payment, error) {
	var payment model.Payment
	query := tx
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Where("order_id = ?", orderID).Order("created_at desc").First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (l *PaymentCore) loadPaymentByNoTx(tx *gorm.DB, paymentNo string, forUpdate bool) (*model.Payment, error) {
	var payment model.Payment
	query := tx
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Where("payment_no = ?", paymentNo).Order("created_at desc").First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (l *PaymentCore) loadLatestRefundTx(tx *gorm.DB, orderID int64, forUpdate bool) (*model.Refund, error) {
	var refund model.Refund
	query := tx
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Where("order_id = ?", orderID).Order("created_at desc").First(&refund).Error; err != nil {
		return nil, err
	}
	return &refund, nil
}

func (l *PaymentCore) createPendingPaymentTx(tx *gorm.DB, order *model.OrderInfo, payMethod int32) (*model.Payment, error) {
	now := time.Now()
	payment := &model.Payment{
		ID:        common.GenerateId(),
		PaymentNo: fmt.Sprintf("P%d", common.GenerateId()),
		OrderID:   order.ID,
		UserID:    order.UserID,
		PayMethod: int16(normalizePayMethod(payMethod)),
		Amount:    order.TotalAmount,
		Status:    paymentStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if payment.PayMethod == 0 {
		payment.PayMethod = 1
	}
	if err := tx.Create(payment).Error; err != nil {
		return nil, err
	}
	return payment, nil
}

func (l *PaymentCore) updatePendingPaymentSession(payment *model.Payment, result *pay.PayResult) error {
	if payment == nil || payment.ID == 0 || result == nil {
		return nil
	}

	now := time.Now()
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.Payment{}).
		Where("id = ? AND status = ?", payment.ID, paymentStatusPending).
		Updates(map[string]any{
			"trade_no":      firstNonEmpty(result.TradeNo, payment.TradeNo),
			"callback_data": firstNonEmpty(result.RawData, payment.CallbackData),
			"updated_at":    now,
		}).Error; err != nil {
		return err
	}

	payment.TradeNo = firstNonEmpty(result.TradeNo, payment.TradeNo)
	payment.CallbackData = firstNonEmpty(result.RawData, payment.CallbackData)
	payment.UpdatedAt = now
	return nil
}

func (l *PaymentCore) findUserEmail(userID int64) string {
	if userID <= 0 {
		return ""
	}

	var user model.User
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Select("email").
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		return ""
	}
	return strings.TrimSpace(user.Email)
}

func (l *PaymentCore) resolveStripeSuccessURL(override string) string {
	return firstNonEmpty(override, l.svcCtx.Config.Stripe.SuccessURL)
}

func (l *PaymentCore) resolveStripeCancelURL(override string) string {
	return firstNonEmpty(override, l.svcCtx.Config.Stripe.CancelURL)
}

func (l *PaymentCore) withLock(name, businessKey string, fn func() error) error {
	if l.svcCtx.Redis == nil {
		return fn()
	}

	lockKey := fmt.Sprintf("%s:lock:%s:%s", l.svcCtx.Config.KeyPrefix, name, businessKey)
	lockValue := strconv.FormatInt(common.GenerateId(), 10)
	lockTTL := time.Duration(l.svcCtx.Config.Lock.TTLSeconds) * time.Second
	retryTimes := l.svcCtx.Config.Lock.RetryTimes
	retryInterval := time.Duration(l.svcCtx.Config.Lock.RetryIntervalMillis) * time.Millisecond

	for i := 0; i < retryTimes; i++ {
		ok, err := l.svcCtx.Redis.SetNX(l.ctx, lockKey, lockValue, lockTTL).Result()
		if err != nil {
			return err
		}
		if ok {
			defer func() {
				if _, err := releaseLockScript.Run(l.ctx, l.svcCtx.Redis, []string{lockKey}, lockValue).Result(); err != nil {
					l.Errorf("release lock failed, key=%s err=%v", lockKey, err)
				}
			}()
			return fn()
		}
		time.Sleep(retryInterval)
	}

	return errors.New("service is processing")
}

func (l *PaymentCore) withContext(ctx context.Context) *PaymentCore {
	return &PaymentCore{
		ctx:    ctx,
		svcCtx: l.svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PaymentCore) refundOperationContext() (context.Context, context.CancelFunc) {
	base := context.Background()
	if l.ctx != nil {
		base = context.WithoutCancel(l.ctx)
	}
	return context.WithTimeout(base, refundOperationTimeout)
}

func (l *PaymentCore) resolveChannel(channel string) string {
	channel = strings.ToLower(strings.TrimSpace(channel))
	switch channel {
	case "", "web", "h5", "pc", "app":
		return l.svcCtx.Config.Pay.DefaultChannel
	default:
		return channel
	}
}

func (l *PaymentCore) checkRefundWindow(event *model.Event) error {
	if event == nil {
		return errors.New("event not found")
	}

	now := time.Now()
	if !now.Before(event.EventStartTime) {
		return errors.New("event already started")
	}

	deadlineHours := l.svcCtx.Config.Business.RefundDeadlineHours
	if deadlineHours <= 0 {
		return nil
	}
	if event.EventStartTime.Sub(now) < time.Duration(deadlineHours)*time.Hour {
		return fmt.Errorf("refund is only allowed %d hours before the event", deadlineHours)
	}
	return nil
}

func paymentToProto(payment *model.Payment) *paymentpb.PaymentInfo {
	if payment == nil || payment.ID == 0 {
		return nil
	}

	return &paymentpb.PaymentInfo{
		Id:           payment.ID,
		PaymentNo:    payment.PaymentNo,
		PayMethod:    int32(payment.PayMethod),
		Amount:       payment.Amount,
		Status:       int32(payment.Status),
		TradeNo:      payment.TradeNo,
		PaidAt:       formatTimeValue(payment.PaidAt),
		CreatedAt:    formatTimeValue(payment.CreatedAt),
		UpdatedAt:    formatTimeValue(payment.UpdatedAt),
		CallbackData: payment.CallbackData,
	}
}

func refundToProto(refund *model.Refund) *paymentpb.RefundInfo {
	if refund == nil || refund.ID == 0 {
		return nil
	}

	return &paymentpb.RefundInfo{
		Id:           refund.ID,
		RefundNo:     refund.RefundNo,
		RefundAmount: refund.RefundAmount,
		Status:       int32(refund.Status),
		Reason:       refund.Reason,
		RejectReason: refund.RejectReason,
		TradeNo:      refund.TradeNo,
		CreatedAt:    formatTimeValue(refund.CreatedAt),
		UpdatedAt:    formatTimeValue(refund.UpdatedAt),
		RefundedAt:   formatTimeValue(refund.RefundedAt),
	}
}

func parseThirdPartyTime(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	value, err := time.ParseInLocation(timeLayout, raw, time.Local)
	if err != nil {
		return nil
	}
	return &value
}

func extractCheckoutSessionInfo(payment *model.Payment) (string, string, string, bool) {
	if payment == nil || strings.TrimSpace(payment.CallbackData) == "" {
		return "", "", "", false
	}

	var payload struct {
		ID        string `json:"id"`
		URL       string `json:"url"`
		Status    string `json:"status"`
		ExpiresAt int64  `json:"expires_at"`
	}
	if err := json.Unmarshal([]byte(payment.CallbackData), &payload); err != nil {
		return "", "", "", false
	}
	if strings.TrimSpace(payload.ID) == "" || strings.TrimSpace(payload.URL) == "" {
		return "", "", "", false
	}
	if strings.EqualFold(strings.TrimSpace(payload.Status), "expired") {
		return "", "", "", false
	}

	var expiresAt string
	if payload.ExpiresAt > 0 {
		expires := time.Unix(payload.ExpiresAt, 0)
		if time.Now().After(expires) {
			return "", "", "", false
		}
		expiresAt = expires.Format(timeLayout)
	}

	return payload.URL, payload.ID, expiresAt, true
}

func resolveRefundStatus(result *pay.RefundResult) int16 {
	if result != nil && result.Pending {
		return refundStatusProcessing
	}
	return refundStatusSuccess
}

func refundStatusFromPending(pending bool) int16 {
	if pending {
		return refundStatusProcessing
	}
	return refundStatusSuccess
}

func isTerminalRefund(status int16) bool {
	return status == refundStatusProcessing || status == refundStatusSuccess
}

func normalizePayMethod(payMethod int32) int32 {
	if payMethod == 2 {
		return 2
	}
	return 1
}

func formatAmount(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func formatTimeValue(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(timeLayout)
}

func amountEqual(a, b string) (bool, error) {
	aa, err := parseAmount(a)
	if err != nil {
		return false, err
	}
	bb, err := parseAmount(b)
	if err != nil {
		return false, err
	}
	return aa.Cmp(bb) == 0, nil
}

func amountGreater(a, b string) (bool, error) {
	aa, err := parseAmount(a)
	if err != nil {
		return false, err
	}
	bb, err := parseAmount(b)
	if err != nil {
		return false, err
	}
	return aa.Cmp(bb) > 0, nil
}

func parseAmount(raw string) (*big.Rat, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("amount is required")
	}
	value, ok := new(big.Rat).SetString(raw)
	if !ok {
		return nil, errors.New("invalid amount")
	}
	return value, nil
}

func mustJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

func chooseTime(primary, fallback *time.Time) *time.Time {
	if primary != nil && !primary.IsZero() {
		return primary
	}
	if fallback != nil && !fallback.IsZero() {
		return fallback
	}
	return nil
}

func timePtrIfNotZero(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

func derefTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
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
