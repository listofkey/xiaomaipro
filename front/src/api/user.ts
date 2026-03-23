import request from '@/utils/request'

export interface UserProfilePayload {
  id: string
  phone: string
  email: string
  nickname: string
  avatar: string
  status: number
  isVerified: boolean
  realName?: string
  createdAt?: string
}

export interface UpdateProfilePayload {
  nickname?: string
  avatar?: string
  email?: string
}

export interface ChangePasswordPayload {
  oldPassword: string
  newPassword: string
}

export interface TicketBuyerPayload {
  id: string
  name: string
  idCard: string
  phone: string
  isDefault: boolean
  createdAt: string
}

export interface TicketBuyerMutationPayload {
  name: string
  idCard?: string
  phone?: string
  isDefault: boolean
}

export interface UpdateTicketBuyerPayload extends TicketBuyerMutationPayload {
  buyerId: string
}

export interface AddressPayload {
  id: string
  receiverName: string
  receiverPhone: string
  province: string
  city: string
  district: string
  detail: string
  isDefault: boolean
  createdAt: string
}

export interface AddressMutationPayload {
  receiverName: string
  receiverPhone: string
  province: string
  city: string
  district: string
  detail: string
  isDefault: boolean
}

export interface UpdateAddressPayload extends AddressMutationPayload {
  addressId: string
}

export interface OperationResp {
  success: boolean
}

export interface TicketBuyerResp {
  ticketBuyer: TicketBuyerPayload
}

export interface TicketBuyerListResp {
  ticketBuyers: TicketBuyerPayload[]
}

export interface AddressResp {
  address: AddressPayload
}

export interface AddressListResp {
  addresses: AddressPayload[]
}

export function getProfile() {
  return request.get<UserProfilePayload, UserProfilePayload>('/user/profile')
}

export function updateProfile(data: UpdateProfilePayload) {
  return request.put<UserProfilePayload, UserProfilePayload>('/user/profile', data)
}

export function changePassword(data: ChangePasswordPayload) {
  return request.put<OperationResp, OperationResp>('/user/password', data)
}

export function listTicketBuyers() {
  return request.get<TicketBuyerListResp, TicketBuyerListResp>('/user/ticket-buyers')
}

export function createTicketBuyer(data: TicketBuyerMutationPayload) {
  return request.post<TicketBuyerResp, TicketBuyerResp>('/user/ticket-buyers', data)
}

export function updateTicketBuyer(data: UpdateTicketBuyerPayload) {
  return request.put<TicketBuyerResp, TicketBuyerResp>('/user/ticket-buyers', data)
}

export function deleteTicketBuyer(buyerId: string) {
  return request.delete<OperationResp, OperationResp>('/user/ticket-buyers', {
    data: { buyerId },
  })
}

export function setDefaultTicketBuyer(buyerId: string) {
  return request.put<OperationResp, OperationResp>('/user/ticket-buyers/default', {
    buyerId,
  })
}

export function listAddresses() {
  return request.get<AddressListResp, AddressListResp>('/user/addresses')
}

export function createAddress(data: AddressMutationPayload) {
  return request.post<AddressResp, AddressResp>('/user/addresses', data)
}

export function updateAddress(data: UpdateAddressPayload) {
  return request.put<AddressResp, AddressResp>('/user/addresses', data)
}

export function deleteAddress(addressId: string) {
  return request.delete<OperationResp, OperationResp>('/user/addresses', {
    data: { addressId },
  })
}

export function setDefaultAddress(addressId: string) {
  return request.put<OperationResp, OperationResp>('/user/addresses/default', {
    addressId,
  })
}
