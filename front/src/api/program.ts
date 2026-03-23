import request from '@/utils/request'

export interface ProgramCategoryInfo {
  id: number
  name: string
  icon: string
  sortOrder: number
  status: number
}

export interface ProgramCityInfo {
  id: string
  name: string
}

export interface ProgramVenueInfo {
  id: string
  name: string
  city: string
  address: string
  capacity: number
  seatMapUrl: string
  description: string
}

export interface ProgramTicketTierInfo {
  id: string
  eventId: string
  name: string
  price: number
  totalStock: number
  soldCount: number
  lockedCount: number
  status: number
  sortOrder: number
  remainStock: number
}

export interface ProgramEventBrief {
  id: string
  title: string
  posterUrl: string
  category: ProgramCategoryInfo
  venueName: string
  city: string
  artist: string
  eventStartTime: string
  eventEndTime: string
  status: number
  minPrice: number
  ticketType: number
  isHot: boolean
}

export interface ProgramEventDetail {
  id: string
  title: string
  description: string
  posterUrl: string
  category: ProgramCategoryInfo
  venue: ProgramVenueInfo
  city: string
  artist: string
  eventStartTime: string
  eventEndTime: string
  saleStartTime: string
  saleEndTime: string
  status: number
  purchaseLimit: number
  needRealName: number
  ticketType: number
  ticketTiers: ProgramTicketTierInfo[]
}

export interface ListEventsParams {
  page?: number
  pageSize?: number
  categoryId?: number
  city?: string
  startDate?: string
  endDate?: string
  sortBy?: 'hot' | 'time' | 'price'
}

export interface SearchEventsParams {
  keyword?: string
  categoryId?: number
  city?: string
  startDate?: string
  endDate?: string
  page?: number
  pageSize?: number
}

export interface ListCategoriesParams {
  status?: number
}

export interface GetHotRecommendParams {
  city?: string
  limit?: number
}

export interface ProgramEventListResp {
  events: ProgramEventBrief[]
  total: number
  page: number
  pageSize: number
}

export interface ProgramEventDetailResp {
  event: ProgramEventDetail
}

export interface ProgramCategoryListResp {
  categories: ProgramCategoryInfo[]
  cities: ProgramCityInfo[]
}

export interface ProgramHotRecommendResp {
  events: ProgramEventBrief[]
}

export function listEvents(params: ListEventsParams = {}) {
  return request.get<ProgramEventListResp, ProgramEventListResp>('/program/events', {
    params,
  })
}

export function searchEvents(params: SearchEventsParams = {}) {
  return request.get<ProgramEventListResp, ProgramEventListResp>('/program/events/search', {
    params,
  })
}

export function getEventDetail(eventId: string) {
  return request.get<ProgramEventDetailResp, ProgramEventDetailResp>('/program/events/detail', {
    params: { eventId },
  })
}

export function listProgramCategories(params: ListCategoriesParams = {}) {
  return request.get<ProgramCategoryListResp, ProgramCategoryListResp>('/program/categories', {
    params,
  })
}

export function getHotRecommend(params: GetHotRecommendParams = {}) {
  return request.get<ProgramHotRecommendResp, ProgramHotRecommendResp>('/program/hot-recommend', {
    params,
  })
}
