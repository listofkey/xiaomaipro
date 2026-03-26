package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"server/app/rpc/model"
	"server/app/rpc/order/internal/svc"
	"server/app/rpc/order/orderpb"
	"server/app/rpc/payment/paymentservice"
	"server/common"
	"server/pkg/monitoring"

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

	cancelReasonNone    int16 = 0
	cancelReasonManual  int16 = 1
	cancelReasonTimeout int16 = 2
	cancelReasonRefund  int16 = 3

	paymentStatusPending int16 = 0
	paymentStatusSuccess int16 = 1
	paymentStatusFailed  int16 = 2

	refundStatusPending int16 = 1
	refundStatusSuccess int16 = 3

	orderTicketStatusUnused   int16 = 0
	orderTicketStatusVerified int16 = 1
	orderTicketStatusVoided   int16 = 2

	eventStatusOnSale    int16 = 1
	ticketTierStatusLive int16 = 1
	ticketTypeEticket    int16 = 1
	ticketTypePaper      int16 = 2

	queueStatusQueued     int32 = 1
	queueStatusProcessing int32 = 2
	queueStatusSuccess    int32 = 3
	queueStatusFailed     int32 = 4

	defaultListPage     int32 = 1
	defaultListPageSize int32 = 20
	maxListPageSize     int32 = 100

	timeLayout = "2006-01-02 15:04:05"
)

var reserveInventoryScript = redis.NewScript(`
local stock = tonumber(redis.call('GET', KEYS[1]))
if not stock then
  return {-3, 0, 0}
end

local purchased = tonumber(redis.call('GET', KEYS[2]) or '0')
local qty = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])

if purchased + qty > limit then
  return {-2, stock, purchased}
end

if stock < qty then
  return {-1, stock, purchased}
end

local remain = redis.call('DECRBY', KEYS[1], qty)
local count = redis.call('INCRBY', KEYS[2], qty)
if ttl > 0 then
  redis.call('EXPIRE', KEYS[1], ttl)
  redis.call('EXPIRE', KEYS[2], ttl)
end

return {1, remain, count}
`)

var releaseInventoryScript = redis.NewScript(`
local qty = tonumber(ARGV[1])
local ttl = tonumber(ARGV[2])

if redis.call('EXISTS', KEYS[1]) == 1 then
  redis.call('INCRBY', KEYS[1], qty)
  if ttl > 0 then
    redis.call('EXPIRE', KEYS[1], ttl)
  end
end

local current = tonumber(redis.call('GET', KEYS[2]) or '0')
if current <= qty then
  redis.call('SET', KEYS[2], 0)
else
  redis.call('DECRBY', KEYS[2], qty)
end

if ttl > 0 then
  redis.call('EXPIRE', KEYS[2], ttl)
end

return 1
`)

type OrderCore struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

type queuedOrderMessage struct {
	OrderNo        string  `json:"orderNo"`
	UserID         int64   `json:"userId"`
	EventID        int64   `json:"eventId"`
	TicketTierID   int64   `json:"ticketTierId"`
	Quantity       int32   `json:"quantity"`
	TicketBuyerIDs []int64 `json:"ticketBuyerIds"`
	AddressID      int64   `json:"addressId"`
	PayMethod      int32   `json:"payMethod"`
	RequestID      string  `json:"requestId"`
}

type queueState struct {
	QueueToken  string `json:"queueToken"`
	OrderNo     string `json:"orderNo"`
	UserID      int64  `json:"userId"`
	Status      int32  `json:"status"`
	Message     string `json:"message"`
	OrderID     int64  `json:"orderId"`
	OrderStatus int32  `json:"orderStatus"`
	UpdatedAt   string `json:"updatedAt"`
}

type orderInfoRecord struct {
	ID           int64      `gorm:"column:id;primaryKey"`
	OrderNo      string     `gorm:"column:order_no"`
	UserID       int64      `gorm:"column:user_id"`
	EventID      int64      `gorm:"column:event_id"`
	TicketTierID int64      `gorm:"column:ticket_tier_id"`
	Quantity     int32      `gorm:"column:quantity"`
	UnitPrice    float64    `gorm:"column:unit_price"`
	TotalAmount  float64    `gorm:"column:total_amount"`
	Status       int16      `gorm:"column:status"`
	CancelReason int16      `gorm:"column:cancel_reason"`
	PayDeadline  *time.Time `gorm:"column:pay_deadline"`
	PaidAt       *time.Time `gorm:"column:paid_at"`
	CancelledAt  *time.Time `gorm:"column:cancelled_at"`
	AddressID    *int64     `gorm:"column:address_id"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (orderInfoRecord) TableName() string {
	return "order_info"
}

type orderTicketRecord struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	OrderID       int64      `gorm:"column:order_id"`
	TicketBuyerID int64      `gorm:"column:ticket_buyer_id"`
	TicketCode    string     `gorm:"column:ticket_code"`
	QrCodeURL     string     `gorm:"column:qr_code_url"`
	Status        int16      `gorm:"column:status"`
	SeatInfo      string     `gorm:"column:seat_info"`
	VerifiedAt    *time.Time `gorm:"column:verified_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (orderTicketRecord) TableName() string {
	return "order_ticket"
}

type paymentRecord struct {
	ID           int64      `gorm:"column:id;primaryKey"`
	PaymentNo    string     `gorm:"column:payment_no"`
	OrderID      int64      `gorm:"column:order_id"`
	UserID       int64      `gorm:"column:user_id"`
	PayMethod    int16      `gorm:"column:pay_method"`
	Amount       float64    `gorm:"column:amount"`
	Status       int16      `gorm:"column:status"`
	TradeNo      *string    `gorm:"column:trade_no"`
	PaidAt       *time.Time `gorm:"column:paid_at"`
	CallbackData *string    `gorm:"column:callback_data"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (paymentRecord) TableName() string {
	return "payment"
}

type refundRecord struct {
	ID           int64      `gorm:"column:id;primaryKey"`
	RefundNo     string     `gorm:"column:refund_no"`
	OrderID      int64      `gorm:"column:order_id"`
	PaymentID    int64      `gorm:"column:payment_id"`
	UserID       int64      `gorm:"column:user_id"`
	RefundAmount float64    `gorm:"column:refund_amount"`
	Status       int16      `gorm:"column:status"`
	Reason       string     `gorm:"column:reason"`
	RejectReason *string    `gorm:"column:reject_reason"`
	TradeNo      *string    `gorm:"column:trade_no"`
	AuditedBy    *int64     `gorm:"column:audited_by"`
	AuditedAt    *time.Time `gorm:"column:audited_at"`
	RefundedAt   *time.Time `gorm:"column:refunded_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (refundRecord) TableName() string {
	return "refund"
}

func NewOrderCore(ctx context.Context, svcCtx *svc.ServiceContext) *OrderCore {
	return &OrderCore{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OrderCore) PreheatInventory(in *orderpb.PreheatInventoryReq) (*orderpb.PreheatInventoryResp, error) {
	now := time.Now()

	eventQuery := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.Event{}).
		Where("status = ?", eventStatusOnSale).
		Where("sale_end_time >= ?", now)

	eventIDs := uniqueInt64(in.GetEventIds())
	if len(eventIDs) > 0 {
		eventQuery = eventQuery.Where("id IN ?", eventIDs)
	}

	var events []model.Event
	if err := eventQuery.Find(&events).Error; err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return &orderpb.PreheatInventoryResp{Results: []*orderpb.InventoryPreheatResult{}}, nil
	}

	validEventIDs := make([]int64, 0, len(events))
	for _, item := range events {
		validEventIDs = append(validEventIDs, item.ID)
	}

	var tiers []model.TicketTier
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("event_id IN ?", validEventIDs).
		Find(&tiers).Error; err != nil {
		return nil, err
	}

	results := make([]*orderpb.InventoryPreheatResult, 0, len(tiers))
	for _, tier := range tiers {
		stock := tier.TotalStock - tier.SoldCount - tier.LockedCount
		if stock < 0 {
			stock = 0
		}

		key := l.tierStockKey(tier.ID)
		ttl := l.inventoryTTL()
		var err error
		if in.GetForce() {
			err = l.svcCtx.Redis.Set(l.ctx, key, stock, ttl).Err()
		} else {
			err = l.svcCtx.Redis.SetNX(l.ctx, key, stock, ttl).Err()
		}

		result := &orderpb.InventoryPreheatResult{
			EventId:      tier.EventID,
			TicketTierId: tier.ID,
			Stock:        stock,
			Success:      err == nil,
		}
		if err != nil {
			result.Message = err.Error()
		} else {
			result.Message = "ok"
		}
		results = append(results, result)
	}

	return &orderpb.PreheatInventoryResp{Results: results}, nil
}

func (l *OrderCore) CreateOrder(in *orderpb.CreateOrderReq) (*orderpb.CreateOrderResp, error) {
	start := time.Now()
	result := "error"
	defer func() {
		monitoring.RecordOperation("order", "create_order", result, time.Since(start))
	}()

	if in == nil {
		return nil, errors.New("request is empty")
	}

	msg, event, err := l.prepareQueuedOrder(in)
	if err != nil {
		return nil, err
	}

	if err := l.ensureTierInventoryCached(msg.TicketTierID, event.PurchaseLimit); err != nil {
		return nil, err
	}
	if err := l.ensurePurchaseCounterCached(msg.UserID, msg.EventID); err != nil {
		return nil, err
	}

	orderNo, created, err := l.acquireRequestSlot(msg.UserID, msg.EventID, msg.RequestID)
	if err != nil {
		return nil, err
	}
	msg.OrderNo = orderNo

	if !created {
		state, _ := l.readQueueState(orderNo)
		result = "success"
		return &orderpb.CreateOrderResp{
			OrderNo:     orderNo,
			QueueToken:  orderNo,
			QueueStatus: fallbackQueueStatus(state),
			Message:     fallbackQueueMessage(state),
		}, nil
	}

	if err := l.reserveInventoryAndQuota(msg.EventID, msg.TicketTierID, msg.UserID, msg.Quantity, event.PurchaseLimit); err != nil {
		_ = l.deleteRequestSlot(msg.UserID, msg.EventID, msg.RequestID)
		return nil, err
	}

	state := &queueState{
		QueueToken: orderNo,
		OrderNo:    orderNo,
		UserID:     msg.UserID,
		Status:     queueStatusQueued,
		Message:    "queued",
		UpdatedAt:  time.Now().Format(timeLayout),
	}
	if err := l.writeQueueState(state); err != nil {
		_ = l.releaseInventoryAndQuota(msg.EventID, msg.TicketTierID, msg.UserID, msg.Quantity)
		_ = l.deleteRequestSlot(msg.UserID, msg.EventID, msg.RequestID)
		return nil, err
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		_ = l.releaseInventoryAndQuota(msg.EventID, msg.TicketTierID, msg.UserID, msg.Quantity)
		_ = l.deleteRequestSlot(msg.UserID, msg.EventID, msg.RequestID)
		return nil, err
	}

	if err := l.svcCtx.PublishOrderMessage(l.ctx, payload); err != nil {
		_ = l.releaseInventoryAndQuota(msg.EventID, msg.TicketTierID, msg.UserID, msg.Quantity)
		_ = l.deleteRequestSlot(msg.UserID, msg.EventID, msg.RequestID)
		_ = l.writeQueueState(&queueState{
			QueueToken: orderNo,
			OrderNo:    orderNo,
			UserID:     msg.UserID,
			Status:     queueStatusFailed,
			Message:    "publish failed",
			UpdatedAt:  time.Now().Format(timeLayout),
		})
		return nil, err
	}

	result = "success"
	return &orderpb.CreateOrderResp{
		OrderNo:     orderNo,
		QueueToken:  orderNo,
		QueueStatus: queueStatusQueued,
		Message:     "queued",
	}, nil
}

func (l *OrderCore) GetQueueStatus(in *orderpb.GetQueueStatusReq) (*orderpb.GetQueueStatusResp, error) {
	if in == nil {
		return nil, errors.New("request is empty")
	}
	if in.UserId <= 0 {
		return nil, errors.New("user_id is required")
	}
	queueToken := strings.TrimSpace(in.QueueToken)
	if queueToken == "" {
		return nil, errors.New("queue_token is required")
	}

	state, err := l.readQueueState(queueToken)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	if state != nil {
		if state.UserID != 0 && state.UserID != in.UserId {
			return nil, errors.New("queue not found")
		}

		resp := &orderpb.GetQueueStatusResp{
			QueueToken:  queueToken,
			OrderNo:     state.OrderNo,
			QueueStatus: state.Status,
			Message:     state.Message,
		}
		if state.Status == queueStatusSuccess || state.OrderStatus > 0 {
			summary, err := l.findOrderSummary(in.UserId, state.OrderNo)
			if err == nil && summary != nil {
				resp.Order = summary
			}
		}
		return resp, nil
	}

	summary, err := l.findOrderSummary(in.UserId, queueToken)
	if err == nil && summary != nil {
		return &orderpb.GetQueueStatusResp{
			QueueToken:  queueToken,
			OrderNo:     queueToken,
			QueueStatus: queueStatusSuccess,
			Message:     "success",
			Order:       summary,
		}, nil
	}

	return &orderpb.GetQueueStatusResp{
		QueueToken:  queueToken,
		OrderNo:     queueToken,
		QueueStatus: queueStatusFailed,
		Message:     "queue state expired",
	}, nil
}

func (l *OrderCore) PayOrder(in *orderpb.PayOrderReq) (*orderpb.PayOrderResp, error) {
	start := time.Now()
	result := "error"
	defer func() {
		monitoring.RecordOperation("order", "pay_order", result, time.Since(start))
	}()

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

	var (
		order model.OrderInfo
	)

	releaseAfterCommit := false
	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_no = ? AND user_id = ?", orderNo, in.UserId).
			First(&order).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("order not found")
			}
			return err
		}

		switch order.Status {
		case orderStatusPaid, orderStatusCompleted:
			return nil
		case orderStatusCancelled:
			return errors.New("order already cancelled")
		case orderStatusRefunded:
			return errors.New("order already refunded")
		case orderStatusPendingPay:
			now := time.Now()
			if !order.PayDeadline.IsZero() && now.After(order.PayDeadline) {
				if err := l.cancelPendingOrderTx(tx, &order, cancelReasonTimeout, &now); err != nil {
					return err
				}
				releaseAfterCommit = true
				return errors.New("order payment timeout")
			}
			return nil
		default:
			return errors.New("order status is invalid")
		}
	})
	if err != nil {
		if releaseAfterCommit {
			_ = l.releaseInventoryAndQuota(order.EventID, order.TicketTierID, order.UserID, order.Quantity)
		}
		return nil, err
	}

	payResp, err := l.svcCtx.PaymentRpc.PayOrder(l.ctx, &paymentservice.PayOrderReq{
		UserId:    in.UserId,
		OrderNo:   orderNo,
		PayMethod: in.PayMethod,
		Channel:   in.Channel,
	})
	if err != nil {
		return nil, err
	}

	resp := &orderpb.PayOrderResp{
		Success:           payResp.Success,
		PayForm:           payResp.PayForm,
		Payment:           mapPaymentServiceInfo(payResp.Payment),
		OrderStatus:       payResp.OrderStatus,
		PaidAt:            payResp.PaidAt,
		CheckoutUrl:       payResp.CheckoutUrl,
		CheckoutSessionId: payResp.CheckoutSessionId,
		SessionExpiresAt:  payResp.SessionExpiresAt,
	}
	if resp.Success {
		result = "success"
	} else {
		result = "business_failed"
	}

	return resp, nil
}

func (l *OrderCore) CancelOrder(in *orderpb.CancelOrderReq) (*orderpb.CancelOrderResp, error) {
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

	order, release, err := l.cancelOrderByNo(in.UserId, orderNo, cancelReasonManual)
	if err != nil {
		return nil, err
	}
	if release {
		_ = l.releaseInventoryAndQuota(order.EventID, order.TicketTierID, order.UserID, order.Quantity)
	}

	return &orderpb.CancelOrderResp{
		Success: true,
		OrderNo: orderNo,
	}, nil
}

func (l *OrderCore) ApplyRefund(in *orderpb.ApplyRefundReq) (*orderpb.ApplyRefundResp, error) {
	start := time.Now()
	result := "error"
	defer func() {
		monitoring.RecordOperation("order", "apply_refund", result, time.Since(start))
	}()

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

	var order model.OrderInfo
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("order_no = ? AND user_id = ?", orderNo, in.UserId).
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}

	refundResp, err := l.svcCtx.PaymentRpc.Refund(l.ctx, &paymentservice.RefundReq{
		UserId:  in.UserId,
		OrderNo: orderNo,
		Reason:  in.Reason,
	})
	if err != nil {
		return nil, err
	}

	if refundResp.ShouldRelease {
		_ = l.releaseInventoryAndQuota(order.EventID, order.TicketTierID, order.UserID, order.Quantity)
	}

	resp := &orderpb.ApplyRefundResp{
		Success: refundResp.Success,
		OrderNo: orderNo,
		Refund:  mapPaymentServiceRefund(refundResp.Refund),
	}
	if resp.Success {
		result = "success"
	} else {
		result = "business_failed"
	}

	return resp, nil
}

func (l *OrderCore) ListOrder(in *orderpb.ListOrderReq) (*orderpb.ListOrderResp, error) {
	if in == nil {
		return nil, errors.New("request is empty")
	}
	if in.UserId <= 0 {
		return nil, errors.New("user_id is required")
	}

	page, pageSize, offset := normalizeListPagination(in.Page, in.PageSize)

	query := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.OrderInfo{}).
		Where("user_id = ?", in.UserId)
	if in.Status > 0 {
		query = query.Where("status = ?", in.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var orders []model.OrderInfo
	if err := query.Order("created_at desc").
		Offset(offset).
		Limit(int(pageSize)).
		Find(&orders).Error; err != nil {
		return nil, err
	}

	summaries, err := l.buildOrderSummaries(orders)
	if err != nil {
		return nil, err
	}

	return &orderpb.ListOrderResp{
		Orders:   summaries,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (l *OrderCore) GetOrderDetail(in *orderpb.GetOrderDetailReq) (*orderpb.GetOrderDetailResp, error) {
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

	var order model.OrderInfo
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("order_no = ? AND user_id = ?", orderNo, in.UserId).
		First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}

	event, tier, venue, city, err := l.loadOrderBaseData(order)
	if err != nil {
		return nil, err
	}

	var address *model.Address
	if order.AddressID > 0 {
		var addressModel model.Address
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", order.AddressID).First(&addressModel).Error; err == nil {
			address = &addressModel
		}
	}

	var tickets []model.OrderTicket
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("order_id = ?", order.ID).
		Order("created_at asc").
		Find(&tickets).Error; err != nil {
		return nil, err
	}

	buyerIDs := make([]int64, 0, len(tickets))
	for _, item := range tickets {
		buyerIDs = append(buyerIDs, item.TicketBuyerID)
	}

	buyerMap, err := l.loadBuyerMap(buyerIDs)
	if err != nil {
		return nil, err
	}

	var payment model.Payment
	_ = l.svcCtx.DB.WithContext(l.ctx).
		Where("order_id = ?", order.ID).
		Order("created_at desc").
		First(&payment).Error

	var refund model.Refund
	refundErr := l.svcCtx.DB.WithContext(l.ctx).
		Where("order_id = ?", order.ID).
		Order("created_at desc").
		First(&refund).Error

	resp := &orderpb.OrderDetail{
		Id:             order.ID,
		OrderNo:        order.OrderNo,
		UserId:         order.UserID,
		EventId:        order.EventID,
		TicketTierId:   order.TicketTierID,
		EventTitle:     event.Title,
		Description:    event.Description,
		PosterUrl:      event.PosterURL,
		VenueName:      venue.Name,
		VenueAddress:   venue.Address,
		City:           city.Name,
		EventStartTime: formatTimeValue(event.EventStartTime),
		EventEndTime:   formatTimeValue(event.EventEndTime),
		SaleStartTime:  formatTimeValue(event.SaleStartTime),
		SaleEndTime:    formatTimeValue(event.SaleEndTime),
		TicketTierName: tier.Name,
		Quantity:       order.Quantity,
		UnitPrice:      order.UnitPrice,
		TotalAmount:    order.TotalAmount,
		Status:         int32(order.Status),
		StatusText:     orderStatusText(order.Status),
		PayDeadline:    formatTimeValue(order.PayDeadline),
		PaidAt:         formatTimeValue(order.PaidAt),
		CancelledAt:    formatTimeValue(order.CancelledAt),
		CreatedAt:      formatTimeValue(order.CreatedAt),
		PurchaseLimit:  event.PurchaseLimit,
		NeedRealName:   int32(event.NeedRealName),
		TicketType:     int32(event.TicketType),
		Delivery:       buildDeliveryInfo(order, event, address),
		Tickets:        make([]*orderpb.OrderTicketInfo, 0, len(tickets)),
	}

	for _, item := range tickets {
		resp.Tickets = append(resp.Tickets, &orderpb.OrderTicketInfo{
			Id:            item.ID,
			TicketBuyerId: item.TicketBuyerID,
			TicketCode:    item.TicketCode,
			QrCodeUrl:     item.QrCodeURL,
			Status:        int32(item.Status),
			SeatInfo:      item.SeatInfo,
			VerifiedAt:    formatTimeValue(item.VerifiedAt),
			Buyer:         mapBuyerInfo(buyerMap[item.TicketBuyerID]),
		})
	}

	if payment.ID > 0 {
		resp.Payment = mapPaymentInfo(paymentRecord{
			ID:           payment.ID,
			PaymentNo:    payment.PaymentNo,
			OrderID:      payment.OrderID,
			UserID:       payment.UserID,
			PayMethod:    payment.PayMethod,
			Amount:       payment.Amount,
			Status:       payment.Status,
			TradeNo:      stringPtrIfNotEmpty(payment.TradeNo),
			PaidAt:       timePtrIfNotZero(payment.PaidAt),
			CallbackData: stringPtrIfNotEmpty(payment.CallbackData),
			CreatedAt:    payment.CreatedAt,
			UpdatedAt:    payment.UpdatedAt,
		})
	}
	if refundErr == nil && refund.ID > 0 {
		resp.Refund = mapRefundInfo(refundRecord{
			ID:           refund.ID,
			RefundNo:     refund.RefundNo,
			OrderID:      refund.OrderID,
			PaymentID:    refund.PaymentID,
			UserID:       refund.UserID,
			RefundAmount: refund.RefundAmount,
			Status:       refund.Status,
			Reason:       refund.Reason,
			RejectReason: stringPtrIfNotEmpty(refund.RejectReason),
			TradeNo:      stringPtrIfNotEmpty(refund.TradeNo),
			RefundedAt:   timePtrIfNotZero(refund.RefundedAt),
			CreatedAt:    refund.CreatedAt,
			UpdatedAt:    refund.UpdatedAt,
		})
	}

	return &orderpb.GetOrderDetailResp{Order: resp}, nil
}

func (l *OrderCore) processQueuedOrderMessage(raw []byte) error {
	var msg queuedOrderMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return err
	}

	_ = l.writeQueueState(&queueState{
		QueueToken: msg.OrderNo,
		OrderNo:    msg.OrderNo,
		UserID:     msg.UserID,
		Status:     queueStatusProcessing,
		Message:    "processing",
		UpdatedAt:  time.Now().Format(timeLayout),
	})

	state, err := l.createQueuedOrderInDB(&msg)
	if err != nil {
		_ = l.releaseInventoryAndQuota(msg.EventID, msg.TicketTierID, msg.UserID, msg.Quantity)
		_ = l.writeQueueState(&queueState{
			QueueToken: msg.OrderNo,
			OrderNo:    msg.OrderNo,
			UserID:     msg.UserID,
			Status:     queueStatusFailed,
			Message:    err.Error(),
			UpdatedAt:  time.Now().Format(timeLayout),
		})
		return err
	}

	return l.writeQueueState(state)
}

func (l *OrderCore) expirePendingOrders(limit int) (int, error) {
	if limit <= 0 {
		limit = 100
	}

	type expiredOrder struct {
		ID           int64  `gorm:"column:id"`
		OrderNo      string `gorm:"column:order_no"`
		UserID       int64  `gorm:"column:user_id"`
		EventID      int64  `gorm:"column:event_id"`
		TicketTierID int64  `gorm:"column:ticket_tier_id"`
		Quantity     int32  `gorm:"column:quantity"`
	}

	expired := make([]expiredOrder, 0, limit)
	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		query := `
SELECT id, order_no, user_id, event_id, ticket_tier_id, quantity
FROM order_info
WHERE status = ? AND pay_deadline IS NOT NULL AND pay_deadline <= ?
ORDER BY pay_deadline ASC
LIMIT ?
FOR UPDATE SKIP LOCKED
`
		if err := tx.Raw(query, orderStatusPendingPay, time.Now(), limit).Scan(&expired).Error; err != nil {
			return err
		}

		for _, item := range expired {
			var order model.OrderInfo
			if err := tx.Where("id = ?", item.ID).First(&order).Error; err != nil {
				return err
			}
			now := time.Now()
			if err := l.cancelPendingOrderTx(tx, &order, cancelReasonTimeout, &now); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	for _, item := range expired {
		_ = l.releaseInventoryAndQuota(item.EventID, item.TicketTierID, item.UserID, item.Quantity)
	}
	return len(expired), nil
}

func (l *OrderCore) prepareQueuedOrder(in *orderpb.CreateOrderReq) (*queuedOrderMessage, *model.Event, error) {
	if in.UserId <= 0 {
		return nil, nil, errors.New("user_id is required")
	}
	if in.EventId <= 0 {
		return nil, nil, errors.New("event_id is required")
	}
	if in.TicketTierId <= 0 {
		return nil, nil, errors.New("ticket_tier_id is required")
	}
	if in.Quantity <= 0 {
		return nil, nil, errors.New("quantity is required")
	}

	buyerIDs := uniqueInt64(in.TicketBuyerIds)
	if len(buyerIDs) == 0 {
		return nil, nil, errors.New("ticket_buyer_ids is required")
	}
	if int32(len(buyerIDs)) != in.Quantity {
		return nil, nil, errors.New("quantity must equal ticket_buyer_ids length")
	}

	var event model.Event
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("id = ?", in.EventId).
		First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("event not found")
		}
		return nil, nil, err
	}
	if err := checkEventOnSale(&event); err != nil {
		return nil, nil, err
	}
	if in.Quantity > event.PurchaseLimit {
		return nil, nil, fmt.Errorf("quantity exceeds purchase limit %d", event.PurchaseLimit)
	}

	var tier model.TicketTier
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("id = ? AND event_id = ?", in.TicketTierId, in.EventId).
		First(&tier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("ticket tier not found")
		}
		return nil, nil, err
	}
	if tier.Status != ticketTierStatusLive {
		return nil, nil, errors.New("ticket tier is not on sale")
	}

	var buyers []model.TicketBuyer
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("user_id = ? AND id IN ?", in.UserId, buyerIDs).
		Find(&buyers).Error; err != nil {
		return nil, nil, err
	}
	if len(buyers) != len(buyerIDs) {
		return nil, nil, errors.New("ticket buyer is invalid")
	}
	if event.NeedRealName == 1 {
		for _, buyer := range buyers {
			if strings.TrimSpace(buyer.IDCard) == "" {
				return nil, nil, errors.New("real-name ticket buyer is required")
			}
		}
	}

	if event.TicketType == ticketTypePaper {
		if in.AddressId <= 0 {
			return nil, nil, errors.New("address_id is required for paper tickets")
		}
		var address model.Address
		if err := l.svcCtx.DB.WithContext(l.ctx).
			Where("id = ? AND user_id = ?", in.AddressId, in.UserId).
			First(&address).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, errors.New("address not found")
			}
			return nil, nil, err
		}
	}

	msg := &queuedOrderMessage{
		UserID:         in.UserId,
		EventID:        in.EventId,
		TicketTierID:   in.TicketTierId,
		Quantity:       in.Quantity,
		TicketBuyerIDs: buyerIDs,
		AddressID:      in.AddressId,
		PayMethod:      normalizePayMethod(in.PayMethod),
		RequestID:      strings.TrimSpace(in.RequestId),
	}
	if msg.RequestID == "" {
		msg.RequestID = strconv.FormatInt(common.GenerateId(), 10)
	}

	return msg, &event, nil
}

func (l *OrderCore) acquireRequestSlot(userID, eventID int64, requestID string) (string, bool, error) {
	key := l.requestKey(userID, eventID, requestID)

	existing, err := l.svcCtx.Redis.Get(l.ctx, key).Result()
	if err == nil && existing != "" {
		return existing, false, nil
	}
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", false, err
	}

	orderNo := strconv.FormatInt(common.GenerateId(), 10)
	ok, err := l.svcCtx.Redis.SetNX(l.ctx, key, orderNo, l.queueStateTTL()).Result()
	if err != nil {
		return "", false, err
	}
	if !ok {
		existing, getErr := l.svcCtx.Redis.Get(l.ctx, key).Result()
		if getErr != nil {
			return "", false, getErr
		}
		return existing, false, nil
	}

	return orderNo, true, nil
}

func (l *OrderCore) deleteRequestSlot(userID, eventID int64, requestID string) error {
	return l.svcCtx.Redis.Del(l.ctx, l.requestKey(userID, eventID, requestID)).Err()
}

func (l *OrderCore) ensureTierInventoryCached(ticketTierID int64, _ int32) error {
	key := l.tierStockKey(ticketTierID)
	exists, err := l.svcCtx.Redis.Exists(l.ctx, key).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}

	var tier model.TicketTier
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", ticketTierID).First(&tier).Error; err != nil {
		return err
	}
	remain := tier.TotalStock - tier.SoldCount - tier.LockedCount
	if remain < 0 {
		remain = 0
	}

	return l.svcCtx.Redis.SetNX(l.ctx, key, remain, l.inventoryTTL()).Err()
}

func (l *OrderCore) ensurePurchaseCounterCached(userID, eventID int64) error {
	key := l.purchaseCounterKey(eventID, userID)
	exists, err := l.svcCtx.Redis.Exists(l.ctx, key).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}

	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.OrderInfo{}).
		Where("user_id = ? AND event_id = ? AND status IN ?", userID, eventID, []int16{
			orderStatusPendingPay,
			orderStatusPaid,
			orderStatusCompleted,
		}).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&count).Error; err != nil {
		return err
	}

	return l.svcCtx.Redis.SetNX(l.ctx, key, count, l.purchaseCounterTTL()).Err()
}

func (l *OrderCore) reserveInventoryAndQuota(eventID, ticketTierID, userID int64, quantity int32, purchaseLimit int32) error {
	result, err := reserveInventoryScript.Run(
		l.ctx,
		l.svcCtx.Redis,
		[]string{l.tierStockKey(ticketTierID), l.purchaseCounterKey(eventID, userID)},
		quantity,
		purchaseLimit,
		int(l.inventoryTTL().Seconds()),
	).Result()
	if err != nil {
		return err
	}

	values, ok := result.([]any)
	if !ok || len(values) == 0 {
		return errors.New("reserve inventory failed")
	}

	code := toInt64(values[0])
	switch code {
	case 1:
		return nil
	case -1:
		return errors.New("inventory is insufficient")
	case -2:
		return errors.New("purchase limit exceeded")
	case -3:
		return errors.New("inventory is not preheated")
	default:
		return errors.New("reserve inventory failed")
	}
}

func (l *OrderCore) releaseInventoryAndQuota(eventID, ticketTierID, userID int64, quantity int32) error {
	_, err := releaseInventoryScript.Run(
		l.ctx,
		l.svcCtx.Redis,
		[]string{l.tierStockKey(ticketTierID), l.purchaseCounterKey(eventID, userID)},
		quantity,
		int(l.inventoryTTL().Seconds()),
	).Result()
	return err
}

func (l *OrderCore) writeQueueState(state *queueState) error {
	payload, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return l.svcCtx.Redis.Set(l.ctx, l.queueStateKey(state.QueueToken), payload, l.queueStateTTL()).Err()
}

func (l *OrderCore) readQueueState(queueToken string) (*queueState, error) {
	raw, err := l.svcCtx.Redis.Get(l.ctx, l.queueStateKey(queueToken)).Bytes()
	if err != nil {
		return nil, err
	}
	var state queueState
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (l *OrderCore) createQueuedOrderInDB(msg *queuedOrderMessage) (*queueState, error) {
	var state *queueState

	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var existed model.OrderInfo
		err := tx.Where("order_no = ?", msg.OrderNo).First(&existed).Error
		if err == nil {
			state = &queueState{
				QueueToken:  msg.OrderNo,
				OrderNo:     msg.OrderNo,
				UserID:      msg.UserID,
				Status:      queueStatusSuccess,
				Message:     "success",
				OrderID:     existed.ID,
				OrderStatus: int32(existed.Status),
				UpdatedAt:   time.Now().Format(timeLayout),
			}
			return nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var event model.Event
		if err := tx.Where("id = ?", msg.EventID).First(&event).Error; err != nil {
			return err
		}
		if err := checkEventOnSale(&event); err != nil {
			return err
		}

		var tier model.TicketTier
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND event_id = ?", msg.TicketTierID, msg.EventID).
			First(&tier).Error; err != nil {
			return err
		}
		if tier.Status != ticketTierStatusLive {
			return errors.New("ticket tier is not on sale")
		}
		if tier.TotalStock-tier.SoldCount-tier.LockedCount < msg.Quantity {
			return errors.New("database inventory is insufficient")
		}

		var buyers []model.TicketBuyer
		if err := tx.Where("user_id = ? AND id IN ?", msg.UserID, msg.TicketBuyerIDs).Find(&buyers).Error; err != nil {
			return err
		}
		if len(buyers) != len(msg.TicketBuyerIDs) {
			return errors.New("ticket buyer is invalid")
		}

		var addressID *int64
		if event.TicketType == ticketTypePaper {
			if msg.AddressID <= 0 {
				return errors.New("address is required")
			}
			var address model.Address
			if err := tx.Where("id = ? AND user_id = ?", msg.AddressID, msg.UserID).First(&address).Error; err != nil {
				return err
			}
			addressID = &address.ID
		}

		now := time.Now()
		payDeadline := now.Add(time.Duration(l.svcCtx.Config.Order.PayTimeoutMinutes) * time.Minute)
		orderID := common.GenerateId()
		order := orderInfoRecord{
			ID:           orderID,
			OrderNo:      msg.OrderNo,
			UserID:       msg.UserID,
			EventID:      msg.EventID,
			TicketTierID: msg.TicketTierID,
			Quantity:     msg.Quantity,
			UnitPrice:    tier.Price,
			TotalAmount:  tier.Price * float64(msg.Quantity),
			Status:       orderStatusPendingPay,
			CancelReason: cancelReasonNone,
			PayDeadline:  &payDeadline,
			AddressID:    addressID,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		tickets := make([]orderTicketRecord, 0, len(msg.TicketBuyerIDs))
		for idx, buyerID := range msg.TicketBuyerIDs {
			code := fmt.Sprintf("TK%s%02d", msg.OrderNo, idx+1)
			qrURL := fmt.Sprintf("mock://ticket/%s", code)
			tickets = append(tickets, orderTicketRecord{
				ID:            common.GenerateId(),
				OrderID:       orderID,
				TicketBuyerID: buyerID,
				TicketCode:    code,
				QrCodeURL:     qrURL,
				Status:        orderTicketStatusUnused,
				CreatedAt:     now,
				UpdatedAt:     now,
			})
		}
		if len(tickets) > 0 {
			if err := tx.Create(&tickets).Error; err != nil {
				return err
			}
		}

		payment := paymentRecord{
			ID:        common.GenerateId(),
			PaymentNo: fmt.Sprintf("P%s", strconv.FormatInt(common.GenerateId(), 10)),
			OrderID:   orderID,
			UserID:    msg.UserID,
			PayMethod: int16(msg.PayMethod),
			Amount:    order.TotalAmount,
			Status:    paymentStatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := tx.Create(&payment).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.TicketTier{}).
			Where("id = ? AND total_stock - sold_count - locked_count >= ?", tier.ID, msg.Quantity).
			Update("locked_count", gorm.Expr("locked_count + ?", msg.Quantity)).Error; err != nil {
			return err
		}

		state = &queueState{
			QueueToken:  msg.OrderNo,
			OrderNo:     msg.OrderNo,
			UserID:      msg.UserID,
			Status:      queueStatusSuccess,
			Message:     "success",
			OrderID:     orderID,
			OrderStatus: int32(orderStatusPendingPay),
			UpdatedAt:   now.Format(timeLayout),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (l *OrderCore) cancelOrderByNo(userID int64, orderNo string, reason int16) (*model.OrderInfo, bool, error) {
	var (
		order   model.OrderInfo
		release bool
	)

	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_no = ? AND user_id = ?", orderNo, userID).
			First(&order).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("order not found")
			}
			return err
		}

		switch order.Status {
		case orderStatusCancelled:
			return nil
		case orderStatusPendingPay:
			now := time.Now()
			if err := l.cancelPendingOrderTx(tx, &order, reason, &now); err != nil {
				return err
			}
			release = true
			return nil
		case orderStatusPaid, orderStatusCompleted:
			return errors.New("paid order can not be cancelled")
		case orderStatusRefunded:
			return errors.New("refunded order can not be cancelled")
		default:
			return errors.New("order status is invalid")
		}
	})
	if err != nil {
		return nil, false, err
	}

	return &order, release, nil
}

func (l *OrderCore) cancelPendingOrderTx(tx *gorm.DB, order *model.OrderInfo, reason int16, now *time.Time) error {
	if order == nil {
		return errors.New("order not found")
	}
	if order.Status != orderStatusPendingPay {
		return errors.New("only pending-pay order can be cancelled")
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

	if err := tx.Model(&orderInfoRecord{}).
		Where("id = ?", order.ID).
		Updates(map[string]any{
			"status":        orderStatusCancelled,
			"cancel_reason": reason,
			"cancelled_at":  now,
		}).Error; err != nil {
		return err
	}

	if err := tx.Model(&paymentRecord{}).
		Where("order_id = ? AND status = ?", order.ID, paymentStatusPending).
		Updates(map[string]any{
			"status":     paymentStatusFailed,
			"updated_at": now,
		}).Error; err != nil {
		return err
	}

	if err := tx.Model(&model.OrderTicket{}).
		Where("order_id = ?", order.ID).
		Update("status", orderTicketStatusVoided).Error; err != nil {
		return err
	}

	if err := tx.Model(&model.TicketTier{}).
		Where("id = ? AND locked_count >= ?", tier.ID, order.Quantity).
		Update("locked_count", gorm.Expr("locked_count - ?", order.Quantity)).Error; err != nil {
		return err
	}

	order.Status = orderStatusCancelled
	order.CancelReason = reason
	if now != nil {
		order.CancelledAt = *now
	}
	return nil
}

func (l *OrderCore) findOrderSummary(userID int64, orderNo string) (*orderpb.OrderSummary, error) {
	var order model.OrderInfo
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("order_no = ? AND user_id = ?", orderNo, userID).
		First(&order).Error; err != nil {
		return nil, err
	}

	list, err := l.buildOrderSummaries([]model.OrderInfo{order})
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[0], nil
}

func (l *OrderCore) buildOrderSummaries(orders []model.OrderInfo) ([]*orderpb.OrderSummary, error) {
	if len(orders) == 0 {
		return []*orderpb.OrderSummary{}, nil
	}

	eventIDs := make([]int64, 0, len(orders))
	tierIDs := make([]int64, 0, len(orders))
	for _, order := range orders {
		eventIDs = append(eventIDs, order.EventID)
		tierIDs = append(tierIDs, order.TicketTierID)
	}

	eventMap, err := l.loadEventMap(eventIDs)
	if err != nil {
		return nil, err
	}
	tierMap, err := l.loadTierMap(tierIDs)
	if err != nil {
		return nil, err
	}

	venueIDs := make([]int64, 0, len(eventMap))
	cityIDs := make([]int64, 0, len(eventMap))
	for _, event := range eventMap {
		venueIDs = append(venueIDs, event.VenueID)
		cityIDs = append(cityIDs, event.CityID)
	}

	venueMap, err := l.loadVenueMap(venueIDs)
	if err != nil {
		return nil, err
	}
	cityMap, err := l.loadCityMap(cityIDs)
	if err != nil {
		return nil, err
	}

	result := make([]*orderpb.OrderSummary, 0, len(orders))
	for _, order := range orders {
		event := eventMap[order.EventID]
		tier := tierMap[order.TicketTierID]

		var venueName, cityName, posterURL, eventTitle string
		var eventStart, eventEnd string
		var ticketType int32
		if event != nil {
			eventTitle = event.Title
			posterURL = event.PosterURL
			eventStart = formatTimeValue(event.EventStartTime)
			eventEnd = formatTimeValue(event.EventEndTime)
			ticketType = int32(event.TicketType)
			if venue := venueMap[event.VenueID]; venue != nil {
				venueName = venue.Name
			}
			if city := cityMap[event.CityID]; city != nil {
				cityName = city.Name
			}
		}

		tierName := ""
		if tier != nil {
			tierName = tier.Name
		}

		result = append(result, &orderpb.OrderSummary{
			Id:             order.ID,
			OrderNo:        order.OrderNo,
			EventId:        order.EventID,
			TicketTierId:   order.TicketTierID,
			EventTitle:     eventTitle,
			PosterUrl:      posterURL,
			VenueName:      venueName,
			City:           cityName,
			EventStartTime: eventStart,
			EventEndTime:   eventEnd,
			TicketTierName: tierName,
			Quantity:       order.Quantity,
			UnitPrice:      order.UnitPrice,
			TotalAmount:    order.TotalAmount,
			Status:         int32(order.Status),
			StatusText:     orderStatusText(order.Status),
			PayDeadline:    formatTimeValue(order.PayDeadline),
			PaidAt:         formatTimeValue(order.PaidAt),
			CancelledAt:    formatTimeValue(order.CancelledAt),
			CreatedAt:      formatTimeValue(order.CreatedAt),
			TicketType:     ticketType,
		})
	}
	return result, nil
}

func (l *OrderCore) loadOrderBaseData(order model.OrderInfo) (*model.Event, *model.TicketTier, *model.Venue, *model.City, error) {
	eventMap, err := l.loadEventMap([]int64{order.EventID})
	if err != nil {
		return nil, nil, nil, nil, err
	}
	tierMap, err := l.loadTierMap([]int64{order.TicketTierID})
	if err != nil {
		return nil, nil, nil, nil, err
	}

	event := eventMap[order.EventID]
	tier := tierMap[order.TicketTierID]
	if event == nil || tier == nil {
		return nil, nil, nil, nil, errors.New("order related event data is missing")
	}

	venueMap, err := l.loadVenueMap([]int64{event.VenueID})
	if err != nil {
		return nil, nil, nil, nil, err
	}
	cityMap, err := l.loadCityMap([]int64{event.CityID})
	if err != nil {
		return nil, nil, nil, nil, err
	}

	venue := venueMap[event.VenueID]
	city := cityMap[event.CityID]
	if venue == nil || city == nil {
		return nil, nil, nil, nil, errors.New("order related venue data is missing")
	}

	return event, tier, venue, city, nil
}

func (l *OrderCore) loadEventMap(ids []int64) (map[int64]*model.Event, error) {
	result := make(map[int64]*model.Event)
	uniqueIDs := uniqueInt64(ids)
	if len(uniqueIDs) == 0 {
		return result, nil
	}

	var list []model.Event
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id IN ?", uniqueIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	for idx := range list {
		item := list[idx]
		result[item.ID] = &item
	}
	return result, nil
}

func (l *OrderCore) loadTierMap(ids []int64) (map[int64]*model.TicketTier, error) {
	result := make(map[int64]*model.TicketTier)
	uniqueIDs := uniqueInt64(ids)
	if len(uniqueIDs) == 0 {
		return result, nil
	}

	var list []model.TicketTier
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id IN ?", uniqueIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	for idx := range list {
		item := list[idx]
		result[item.ID] = &item
	}
	return result, nil
}

func (l *OrderCore) loadVenueMap(ids []int64) (map[int64]*model.Venue, error) {
	result := make(map[int64]*model.Venue)
	uniqueIDs := uniqueInt64(ids)
	if len(uniqueIDs) == 0 {
		return result, nil
	}

	var list []model.Venue
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id IN ?", uniqueIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	for idx := range list {
		item := list[idx]
		result[item.ID] = &item
	}
	return result, nil
}

func (l *OrderCore) loadCityMap(ids []int64) (map[int64]*model.City, error) {
	result := make(map[int64]*model.City)
	uniqueIDs := uniqueInt64(ids)
	if len(uniqueIDs) == 0 {
		return result, nil
	}

	var list []model.City
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id IN ?", uniqueIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	for idx := range list {
		item := list[idx]
		result[item.ID] = &item
	}
	return result, nil
}

func (l *OrderCore) loadBuyerMap(ids []int64) (map[int64]*model.TicketBuyer, error) {
	result := make(map[int64]*model.TicketBuyer)
	uniqueIDs := uniqueInt64(ids)
	if len(uniqueIDs) == 0 {
		return result, nil
	}

	var list []model.TicketBuyer
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id IN ?", uniqueIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	for idx := range list {
		item := list[idx]
		result[item.ID] = &item
	}
	return result, nil
}

func (l *OrderCore) checkRefundWindow(event *model.Event) error {
	if event == nil {
		return errors.New("event not found")
	}

	now := time.Now()
	if !now.Before(event.EventStartTime) {
		return errors.New("event already started")
	}

	deadlineHours := l.svcCtx.Config.Order.RefundDeadlineHours
	if deadlineHours <= 0 {
		return nil
	}
	if event.EventStartTime.Sub(now) < time.Duration(deadlineHours)*time.Hour {
		return fmt.Errorf("refund is only allowed %d hours before the event", deadlineHours)
	}
	return nil
}

func (l *OrderCore) buildMockPayForm(orderNo string) string {
	template := l.svcCtx.Config.Pay.MockPayFormTemplate
	if strings.Contains(template, "{order_no}") {
		return strings.ReplaceAll(template, "{order_no}", orderNo)
	}
	if strings.Contains(template, "{order_number}") {
		return strings.ReplaceAll(template, "{order_number}", orderNo)
	}
	if strings.Contains(template, "%s") {
		return fmt.Sprintf(template, orderNo)
	}
	return strings.TrimRight(template, "/") + "/" + orderNo
}

func (l *OrderCore) queueStateKey(queueToken string) string {
	return fmt.Sprintf("%s:queue:%s", l.svcCtx.Config.KeyPrefix, queueToken)
}

func (l *OrderCore) requestKey(userID, eventID int64, requestID string) string {
	return fmt.Sprintf("%s:request:%d:%d:%s", l.svcCtx.Config.KeyPrefix, userID, eventID, requestID)
}

func (l *OrderCore) tierStockKey(ticketTierID int64) string {
	return fmt.Sprintf("%s:stock:%d", l.svcCtx.Config.KeyPrefix, ticketTierID)
}

func (l *OrderCore) purchaseCounterKey(eventID, userID int64) string {
	return fmt.Sprintf("%s:purchase:%d:%d", l.svcCtx.Config.KeyPrefix, eventID, userID)
}

func (l *OrderCore) queueStateTTL() time.Duration {
	return time.Duration(l.svcCtx.Config.Order.QueueStatusTTLMinutes) * time.Minute
}

func (l *OrderCore) inventoryTTL() time.Duration {
	return time.Duration(l.svcCtx.Config.Order.InventoryTTLHours) * time.Hour
}

func (l *OrderCore) purchaseCounterTTL() time.Duration {
	return time.Duration(l.svcCtx.Config.Order.PurchaseCounterTTLHours) * time.Hour
}

func normalizeListPagination(page, pageSize int32) (int32, int32, int) {
	if page <= 0 {
		page = defaultListPage
	}
	if pageSize <= 0 {
		pageSize = defaultListPageSize
	}
	if pageSize > maxListPageSize {
		pageSize = maxListPageSize
	}
	offset := int((page - 1) * pageSize)
	return page, pageSize, offset
}

func normalizePayMethod(payMethod int32) int32 {
	if payMethod == 2 {
		return 2
	}
	return 1
}

func checkEventOnSale(event *model.Event) error {
	if event == nil {
		return errors.New("event not found")
	}
	now := time.Now()
	if event.Status != eventStatusOnSale {
		return errors.New("event is not on sale")
	}
	if now.Before(event.SaleStartTime) {
		return errors.New("sale has not started")
	}
	if now.After(event.SaleEndTime) {
		return errors.New("sale has ended")
	}
	if !now.Before(event.EventEndTime) {
		return errors.New("event has ended")
	}
	return nil
}

func orderStatusText(status int16) string {
	switch status {
	case orderStatusPendingPay:
		return "pending_pay"
	case orderStatusCancelled:
		return "cancelled"
	case orderStatusPaid:
		return "paid"
	case orderStatusRefunded:
		return "refunded"
	case orderStatusCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

func buildDeliveryInfo(order model.OrderInfo, event *model.Event, address *model.Address) *orderpb.OrderDeliveryInfo {
	if event == nil {
		return &orderpb.OrderDeliveryInfo{}
	}

	info := &orderpb.OrderDeliveryInfo{
		TicketType: int32(event.TicketType),
	}

	if event.TicketType == ticketTypePaper {
		info.DeliveryMethod = "paper_delivery"
		info.DeliveryStatus = paperDeliveryStatus(order.Status)
		if address != nil {
			info.AddressId = address.ID
			info.ReceiverName = address.ReceiverName
			info.ReceiverPhone = address.ReceiverPhone
			info.Province = address.Province
			info.City = address.City
			info.District = address.District
			info.Detail = address.Detail
		}
		return info
	}

	info.DeliveryMethod = "eticket"
	info.DeliveryStatus = eticketDeliveryStatus(order.Status)
	return info
}

func eticketDeliveryStatus(status int16) string {
	switch status {
	case orderStatusPendingPay:
		return "waiting_payment"
	case orderStatusPaid, orderStatusCompleted:
		return "ticket_generated"
	case orderStatusCancelled, orderStatusRefunded:
		return "invalid"
	default:
		return "unknown"
	}
}

func paperDeliveryStatus(status int16) string {
	switch status {
	case orderStatusPendingPay:
		return "waiting_payment"
	case orderStatusPaid, orderStatusCompleted:
		return "waiting_shipment"
	case orderStatusCancelled, orderStatusRefunded:
		return "no_shipment"
	default:
		return "unknown"
	}
}

func mapBuyerInfo(buyer *model.TicketBuyer) *orderpb.OrderBuyerInfo {
	if buyer == nil {
		return &orderpb.OrderBuyerInfo{}
	}
	return &orderpb.OrderBuyerInfo{
		Id:     buyer.ID,
		Name:   buyer.Name,
		IdCard: buyer.IDCard,
		Phone:  buyer.Phone,
	}
}

func mapPaymentInfo(payment paymentRecord) *orderpb.OrderPaymentInfo {
	return &orderpb.OrderPaymentInfo{
		Id:           payment.ID,
		PaymentNo:    payment.PaymentNo,
		PayMethod:    int32(payment.PayMethod),
		Amount:       payment.Amount,
		Status:       int32(payment.Status),
		TradeNo:      derefString(payment.TradeNo),
		PaidAt:       formatTimePtr(payment.PaidAt),
		CreatedAt:    formatTimeValue(payment.CreatedAt),
		CallbackData: derefString(payment.CallbackData),
	}
}

func mapRefundInfo(refund refundRecord) *orderpb.OrderRefundInfo {
	return &orderpb.OrderRefundInfo{
		Id:           refund.ID,
		RefundNo:     refund.RefundNo,
		RefundAmount: refund.RefundAmount,
		Status:       int32(refund.Status),
		Reason:       refund.Reason,
		RejectReason: derefString(refund.RejectReason),
		TradeNo:      derefString(refund.TradeNo),
		CreatedAt:    formatTimeValue(refund.CreatedAt),
		RefundedAt:   formatTimePtr(refund.RefundedAt),
	}
}

func mapPaymentServiceInfo(info *paymentservice.PaymentInfo) *orderpb.OrderPaymentInfo {
	if info == nil {
		return nil
	}
	return &orderpb.OrderPaymentInfo{
		Id:           info.Id,
		PaymentNo:    info.PaymentNo,
		PayMethod:    info.PayMethod,
		Amount:       info.Amount,
		Status:       info.Status,
		TradeNo:      info.TradeNo,
		PaidAt:       info.PaidAt,
		CreatedAt:    info.CreatedAt,
		CallbackData: info.CallbackData,
	}
}

func mapPaymentServiceRefund(info *paymentservice.RefundInfo) *orderpb.OrderRefundInfo {
	if info == nil {
		return nil
	}
	return &orderpb.OrderRefundInfo{
		Id:           info.Id,
		RefundNo:     info.RefundNo,
		RefundAmount: info.RefundAmount,
		Status:       info.Status,
		Reason:       info.Reason,
		RejectReason: info.RejectReason,
		TradeNo:      info.TradeNo,
		CreatedAt:    info.CreatedAt,
		RefundedAt:   info.RefundedAt,
	}
}

func uniqueInt64(values []int64) []int64 {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func toInt64(value any) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int32:
		return int64(v)
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		i, _ := strconv.ParseInt(v, 10, 64)
		return i
	default:
		return 0
	}
}

func fallbackQueueStatus(state *queueState) int32 {
	if state == nil || state.Status == 0 {
		return queueStatusQueued
	}
	return state.Status
}

func fallbackQueueMessage(state *queueState) string {
	if state == nil || state.Message == "" {
		return "queued"
	}
	return state.Message
}

func formatTimeValue(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(timeLayout)
}

func formatTimePtr(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format(timeLayout)
}

func timePtrIfNotZero(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

func stringPtrIfNotEmpty(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
