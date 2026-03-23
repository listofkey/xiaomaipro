import { defineStore } from 'pinia'
import { ref } from 'vue'

const TOKEN_KEY = 'token'
const REFRESH_TOKEN_KEY = 'refreshToken'
const USER_INFO_KEY = 'userInfo'

export interface UserProfile {
  id: string
  nickname: string
  avatar: string
  phone: string
  email: string
  status: number
  isVerified: boolean
  isRealName: boolean
  realName: string
  createdAt: string
}

function defaultUserProfile(): UserProfile {
  return {
    id: '',
    nickname: '',
    avatar: '',
    phone: '',
    email: '',
    status: 0,
    isVerified: false,
    isRealName: false,
    realName: '',
    createdAt: '',
  }
}

function loadUserProfile(): UserProfile {
  const raw = localStorage.getItem(USER_INFO_KEY)
  if (!raw) {
    return defaultUserProfile()
  }

  try {
    const parsed = JSON.parse(raw) as Partial<UserProfile>
    return {
      ...defaultUserProfile(),
      ...parsed,
    }
  } catch {
    return defaultUserProfile()
  }
}

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem(TOKEN_KEY) || '')
  const refreshToken = ref(localStorage.getItem(REFRESH_TOKEN_KEY) || '')
  const userInfo = ref<UserProfile>(loadUserProfile())

  function setToken(newToken: string) {
    token.value = newToken
    localStorage.setItem(TOKEN_KEY, newToken)
  }

  function setRefreshToken(newRefreshToken: string) {
    refreshToken.value = newRefreshToken
    localStorage.setItem(REFRESH_TOKEN_KEY, newRefreshToken)
  }

  function setTokens(newToken: string, newRefreshToken?: string) {
    setToken(newToken)
    if (newRefreshToken) {
      setRefreshToken(newRefreshToken)
    }
  }

  function setUserInfo(profile: Partial<UserProfile>) {
    userInfo.value = {
      ...defaultUserProfile(),
      ...userInfo.value,
      ...profile,
      isRealName: profile.isRealName ?? profile.isVerified ?? userInfo.value.isRealName,
      isVerified: profile.isVerified ?? profile.isRealName ?? userInfo.value.isVerified,
    }
    localStorage.setItem(USER_INFO_KEY, JSON.stringify(userInfo.value))
  }

  function setAuth(payload: {
    accessToken: string
    refreshToken: string
    userInfo?: Partial<UserProfile>
  }) {
    setToken(payload.accessToken)
    setRefreshToken(payload.refreshToken)
    if (payload.userInfo) {
      setUserInfo(payload.userInfo)
    }
  }

  function clearAuth() {
    token.value = ''
    refreshToken.value = ''
    userInfo.value = defaultUserProfile()
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    localStorage.removeItem(USER_INFO_KEY)
  }

  function clearToken() {
    clearAuth()
  }

  return {
    token,
    refreshToken,
    userInfo,
    setToken,
    setRefreshToken,
    setTokens,
    setUserInfo,
    setAuth,
    clearAuth,
    clearToken,
  }
})
