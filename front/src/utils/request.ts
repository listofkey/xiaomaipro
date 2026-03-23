import axios from 'axios'
import type { AxiosError, AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'

import { useUserStore } from '@/store/user'

type ApiEnvelope<T> = {
  code: number
  msg?: string
  data: T
}

type RetryableRequestConfig = InternalAxiosRequestConfig & {
  _retry?: boolean
  silentError?: boolean
}

const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'

const request: AxiosInstance = axios.create({
  baseURL,
  timeout: 0
})

const refreshClient = axios.create({
  baseURL,
  timeout: 10000,
})

let refreshPromise: Promise<string | null> | null = null

function unwrapResponseData<T>(payload: T | ApiEnvelope<T>): T {
  if (
    payload &&
    typeof payload === 'object' &&
    'code' in payload &&
    typeof (payload as ApiEnvelope<T>).code === 'number'
  ) {
    const envelope = payload as ApiEnvelope<T>
    if (envelope.code === 200) {
      return envelope.data
    }

    throw new Error(envelope.msg || 'Request Error')
  }

  return payload as T
}

async function refreshAccessToken(): Promise<string | null> {
  const userStore = useUserStore()
  if (!userStore.refreshToken) {
    return null
  }

  if (!refreshPromise) {
    refreshPromise = refreshClient
      .post('/auth/token/refresh', {
        refreshToken: userStore.refreshToken,
      })
      .then((response: AxiosResponse) => {
        const data = unwrapResponseData<{
          accessToken: string
          refreshToken: string
          expiresIn: number
        }>(response.data)

        userStore.setTokens(data.accessToken, data.refreshToken)
        return data.accessToken
      })
      .catch(() => {
        userStore.clearAuth()
        return null
      })
      .finally(() => {
        refreshPromise = null
      })
  }

  return refreshPromise
}

function redirectToLogin() {
  if (window.location.pathname !== '/login') {
    window.location.href = '/login'
  }
}

request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    return config
  },
  (error: AxiosError) => Promise.reject(error),
)

request.interceptors.response.use(
  (response: AxiosResponse) => unwrapResponseData(response.data),
  async (error: AxiosError) => {
    const response = error.response
    const config = error.config as RetryableRequestConfig | undefined
    const requestURL = String(config?.url || '')


    // todo 后续加入token刷新机制
    
    // if (
    //   response?.status === 401 &&
    //   config &&
    //   !config._retry &&
    //   !requestURL.includes('/auth/login') &&
    //   !requestURL.includes('/auth/register') &&
    //   !requestURL.includes('/auth/token/refresh')
    // ) {
    //   config._retry = true

    //   const newToken = await refreshAccessToken()
    //   if (newToken) {
    //     config.headers.Authorization = `Bearer ${newToken}`
    //     return request(config)
    //   }
    // }

    // if (response?.status === 401) {
    //   ElMessage.error((response.data as { msg?: string } | undefined)?.msg || '登录状态已失效，请重新登录')
    //   const userStore = useUserStore()
    //   userStore.clearAuth()
    //   redirectToLogin()
    //   return Promise.reject(error)
    // }

    if (!config?.silentError) {
      if (response) {
        ElMessage.error((response.data as { msg?: string } | undefined)?.msg || `HTTP Error ${response.status}`)
      } else {
        ElMessage.error('Network Error')
      }
    }

    return Promise.reject(error)
  },
)

export default request
