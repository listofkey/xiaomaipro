package logic

import (
	"server/app/api/internal/types"
	"server/app/rpc/order/orderservice"
	"server/app/rpc/payment/paymentservice"
)

func mapOrderBuyer(info *orderservice.OrderBuyerInfo) types.OrderBuyerInfo {
	if info == nil {
		return types.OrderBuyerInfo{}
	}
	return types.OrderBuyerInfo{
		Id:     formatID(info.Id),
		Name:   info.Name,
		IdCard: info.IdCard,
		Phone:  info.Phone,
	}
}

func mapOrderTicket(info *orderservice.OrderTicketInfo) types.OrderTicketInfo {
	if info == nil {
		return types.OrderTicketInfo{}
	}
	return types.OrderTicketInfo{
		Id:            formatID(info.Id),
		TicketBuyerId: formatID(info.TicketBuyerId),
		TicketCode:    info.TicketCode,
		QrCodeUrl:     info.QrCodeUrl,
		Status:        info.Status,
		SeatInfo:      info.SeatInfo,
		VerifiedAt:    info.VerifiedAt,
		Buyer:         mapOrderBuyer(info.Buyer),
	}
}

func mapOrderTickets(items []*orderservice.OrderTicketInfo) []types.OrderTicketInfo {
	result := make([]types.OrderTicketInfo, 0, len(items))
	for _, item := range items {
		result = append(result, mapOrderTicket(item))
	}
	return result
}

func mapOrderPayment(info *orderservice.OrderPaymentInfo) types.OrderPaymentInfo {
	if info == nil {
		return types.OrderPaymentInfo{}
	}
	return types.OrderPaymentInfo{
		Id:           formatID(info.Id),
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

func mapOrderRefund(info *orderservice.OrderRefundInfo) types.OrderRefundInfo {
	if info == nil {
		return types.OrderRefundInfo{}
	}
	return types.OrderRefundInfo{
		Id:           formatID(info.Id),
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

func mapPaymentInfo(info *paymentservice.PaymentInfo) types.OrderPaymentInfo {
	if info == nil {
		return types.OrderPaymentInfo{}
	}
	return types.OrderPaymentInfo{
		Id:           formatID(info.Id),
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

func mapPaymentRefund(info *paymentservice.RefundInfo) types.OrderRefundInfo {
	if info == nil {
		return types.OrderRefundInfo{}
	}
	return types.OrderRefundInfo{
		Id:           formatID(info.Id),
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

func mapOrderDelivery(info *orderservice.OrderDeliveryInfo) types.OrderDeliveryInfo {
	if info == nil {
		return types.OrderDeliveryInfo{}
	}
	return types.OrderDeliveryInfo{
		TicketType:     info.TicketType,
		DeliveryMethod: info.DeliveryMethod,
		AddressId:      formatID(info.AddressId),
		ReceiverName:   info.ReceiverName,
		ReceiverPhone:  info.ReceiverPhone,
		Province:       info.Province,
		City:           info.City,
		District:       info.District,
		Detail:         info.Detail,
		DeliveryStatus: info.DeliveryStatus,
	}
}

func mapOrderSummary(info *orderservice.OrderSummary) types.OrderSummary {
	if info == nil {
		return types.OrderSummary{}
	}
	return types.OrderSummary{
		Id:             formatID(info.Id),
		OrderNo:        info.OrderNo,
		EventId:        formatID(info.EventId),
		TicketTierId:   formatID(info.TicketTierId),
		EventTitle:     info.EventTitle,
		PosterUrl:      info.PosterUrl,
		VenueName:      info.VenueName,
		City:           info.City,
		EventStartTime: info.EventStartTime,
		EventEndTime:   info.EventEndTime,
		TicketTierName: info.TicketTierName,
		Quantity:       info.Quantity,
		UnitPrice:      info.UnitPrice,
		TotalAmount:    info.TotalAmount,
		Status:         info.Status,
		StatusText:     info.StatusText,
		PayDeadline:    info.PayDeadline,
		PaidAt:         info.PaidAt,
		CancelledAt:    info.CancelledAt,
		CreatedAt:      info.CreatedAt,
		TicketType:     info.TicketType,
	}
}

func mapOrderSummaryList(items []*orderservice.OrderSummary) []types.OrderSummary {
	result := make([]types.OrderSummary, 0, len(items))
	for _, item := range items {
		result = append(result, mapOrderSummary(item))
	}
	return result
}

func mapOrderDetail(info *orderservice.OrderDetail) types.OrderDetail {
	if info == nil {
		return types.OrderDetail{}
	}
	return types.OrderDetail{
		Id:             formatID(info.Id),
		OrderNo:        info.OrderNo,
		UserId:         formatID(info.UserId),
		EventId:        formatID(info.EventId),
		TicketTierId:   formatID(info.TicketTierId),
		EventTitle:     info.EventTitle,
		Description:    info.Description,
		PosterUrl:      info.PosterUrl,
		VenueName:      info.VenueName,
		VenueAddress:   info.VenueAddress,
		City:           info.City,
		EventStartTime: info.EventStartTime,
		EventEndTime:   info.EventEndTime,
		SaleStartTime:  info.SaleStartTime,
		SaleEndTime:    info.SaleEndTime,
		TicketTierName: info.TicketTierName,
		Quantity:       info.Quantity,
		UnitPrice:      info.UnitPrice,
		TotalAmount:    info.TotalAmount,
		Status:         info.Status,
		StatusText:     info.StatusText,
		PayDeadline:    info.PayDeadline,
		PaidAt:         info.PaidAt,
		CancelledAt:    info.CancelledAt,
		CreatedAt:      info.CreatedAt,
		PurchaseLimit:  info.PurchaseLimit,
		NeedRealName:   info.NeedRealName,
		TicketType:     info.TicketType,
		Delivery:       mapOrderDelivery(info.Delivery),
		Tickets:        mapOrderTickets(info.Tickets),
		Payment:        mapOrderPayment(info.Payment),
		Refund:         mapOrderRefund(info.Refund),
	}
}
