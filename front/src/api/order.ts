import request from '@/utils/request'

export interface CreateOrderPayload {
  eventId: string
  ticketTierId: string
  quantity: number
  ticketBuyerIds: string[]
  addressId?: string
  payMethod?: number
  requestId?: string
}

export interface GetOrderQueueStatusParams {
  queueToken: string
}

export interface PayOrderPayload {
  orderNo: string
  payMethod?: number
  channel?: string
}

export interface CancelOrderPayload {
  orderNo: string
}

export interface ApplyRefundPayload {
  orderNo: string
  reason?: string
}

export interface ListOrderParams {
  status?: number
  page?: number
  pageSize?: number
}

export interface GetOrderDetailParams {
  orderNo: string
}

export interface OrderBuyerPayload {
  id: string
  name: string
  idCard: string
  phone: string
}

export interface OrderTicketPayload {
  id: string
  ticketBuyerId: string
  ticketCode: string
  qrCodeUrl: string
  status: number
  seatInfo: string
  verifiedAt: string
  buyer: OrderBuyerPayload
}

export interface OrderPaymentPayload {
  id: string
  paymentNo: string
  payMethod: number
  amount: number
  status: number
  tradeNo: string
  paidAt: string
  createdAt: string
  callbackData: string
}

export interface OrderRefundPayload {
  id: string
  refundNo: string
  refundAmount: number
  status: number
  reason: string
  rejectReason: string
  tradeNo: string
  createdAt: string
  refundedAt: string
}

export interface OrderDeliveryPayload {
  ticketType: number
  deliveryMethod: string
  addressId: string
  receiverName: string
  receiverPhone: string
  province: string
  city: string
  district: string
  detail: string
  deliveryStatus: string
}

export interface OrderSummaryPayload {
  id: string
  orderNo: string
  eventId: string
  ticketTierId: string
  eventTitle: string
  posterUrl: string
  venueName: string
  city: string
  eventStartTime: string
  eventEndTime: string
  ticketTierName: string
  quantity: number
  unitPrice: number
  totalAmount: number
  status: number
  statusText: string
  payDeadline: string
  paidAt: string
  cancelledAt: string
  createdAt: string
  ticketType: number
}

export interface OrderDetailPayload {
  id: string
  orderNo: string
  userId: string
  eventId: string
  ticketTierId: string
  eventTitle: string
  description: string
  posterUrl: string
  venueName: string
  venueAddress: string
  city: string
  eventStartTime: string
  eventEndTime: string
  saleStartTime: string
  saleEndTime: string
  ticketTierName: string
  quantity: number
  unitPrice: number
  totalAmount: number
  status: number
  statusText: string
  payDeadline: string
  paidAt: string
  cancelledAt: string
  createdAt: string
  purchaseLimit: number
  needRealName: number
  ticketType: number
  delivery: OrderDeliveryPayload
  tickets: OrderTicketPayload[]
  payment: OrderPaymentPayload
  refund: OrderRefundPayload
}

export interface CreateOrderResp {
  orderNo: string
  queueToken: string
  queueStatus: number
  message: string
}

export interface OrderQueueStatusResp {
  queueToken: string
  orderNo: string
  queueStatus: number
  message: string
  order?: OrderSummaryPayload
}

export interface PayOrderResp {
  success: boolean
  payForm: string
  payment: OrderPaymentPayload
  orderStatus: number
  paidAt: string
  checkoutUrl: string
  checkoutSessionId: string
  sessionExpiresAt: string
}

export interface ApplyRefundResp {
  success: boolean
  orderNo: string
  refund: OrderRefundPayload
}

export interface OrderListResp {
  orders: OrderSummaryPayload[]
  total: number
  page: number
  pageSize: number
}

export interface OrderDetailResp {
  order: OrderDetailPayload
}

export function createOrder(data: CreateOrderPayload) {
  return request.post<CreateOrderResp, CreateOrderResp>('/order/create', data)
}

export function getOrderQueueStatus(params: GetOrderQueueStatusParams) {
  return request.get<OrderQueueStatusResp, OrderQueueStatusResp>('/order/queue-status', {
    params,
  })
}

export function payOrder(data: PayOrderPayload) {
  return request.post<PayOrderResp, PayOrderResp>('/order/pay', data)
}

export function cancelOrder(data: CancelOrderPayload) {
  return request.post<{ success: boolean }, { success: boolean }>('/order/cancel', data)
}

export function applyRefund(data: ApplyRefundPayload) {
  return request.post<ApplyRefundResp, ApplyRefundResp>('/order/refund', data)
}

export function listOrders(params: ListOrderParams = {}) {
  return request.get<OrderListResp, OrderListResp>('/order/list', {
    params,
  })
}

export function getOrderDetail(params: GetOrderDetailParams) {
  return request.get<OrderDetailResp, OrderDetailResp>('/order/detail', {
    params,
  })
}
