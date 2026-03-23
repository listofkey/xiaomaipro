import request from '@/utils/request'

import type { OrderPaymentPayload, OrderRefundPayload } from './order'

export interface PaymentDetailParams {
  orderNo: string
}

export interface PaymentStatusPayload {
  success: boolean
  paid: boolean
  payment: OrderPaymentPayload
  orderStatus: number
  paidAt: string
  checkoutSessionId: string
}

export interface PaymentStatusParams {
  orderNo?: string
  paymentNo?: string
  channel?: string
}

export interface PaymentRequestOptions {
  silentError?: boolean
}

export interface PaymentDetailPayload {
  payment: OrderPaymentPayload
  refund: OrderRefundPayload
  orderStatus: number
  paidAt: string
}

export function getPaymentStatus(data: PaymentStatusParams, options: PaymentRequestOptions = {}) {
  return request.post<PaymentStatusPayload, PaymentStatusPayload>(
    '/payment/status',
    data,
    {
      silentError: options.silentError,
    } as any,
  )
}

export function getPaymentDetail(params: PaymentDetailParams) {
  return request.get<PaymentDetailPayload, PaymentDetailPayload>('/payment/detail', {
    params,
  })
}
