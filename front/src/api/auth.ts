import request from '@/utils/request'

export interface AuthUserInfo {
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

export interface AuthPayload {
  userId: string
  accessToken: string
  refreshToken: string
  expiresIn: number
  userInfo: AuthUserInfo
}

export interface RefreshTokenPayload {
  accessToken: string
  refreshToken: string
  expiresIn: number
}

export interface RegisterPayload {
  phone?: string
  email?: string
  password: string
  nickname?: string
  code?: string
}

export interface LoginPayload {
  account: string
  password: string
}

export function register(data: RegisterPayload) {
  return request.post<AuthPayload, AuthPayload>('/auth/register', data)
}

export function login(data: LoginPayload) {
  return request.post<AuthPayload, AuthPayload>('/auth/login', data)
}

export function refreshToken(data: { refreshToken: string }) {
  return request.post<RefreshTokenPayload, RefreshTokenPayload>('/auth/token/refresh', data)
}

export function validateToken() {
  return request.get<
    {
      valid: boolean
      userId: string
      status: number
      userInfo: AuthUserInfo
    },
    {
      valid: boolean
      userId: string
      status: number
      userInfo: AuthUserInfo
    }
  >('/auth/token/validate')
}

export function logout() {
  return request.post<{ success: boolean }, { success: boolean }>('/auth/logout', {})
}
