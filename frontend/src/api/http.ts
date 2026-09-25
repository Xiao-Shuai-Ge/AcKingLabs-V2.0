// 统一 HTTP 客户端：自动携带 token；40100 时静默刷新令牌并重放请求；业务码 toast。
import axios from 'axios'
import type { AxiosRequestConfig, InternalAxiosRequestConfig, AxiosResponse } from 'axios'

export interface ApiBody<T = any> {
  code: number
  message: string
  data: T
}

// 后端地址：开发环境走 vite 代理（同源），生产由 nginx 反代 /api
export const BACKEND_URL = ''

const instance = axios.create({
  baseURL: BACKEND_URL,
  timeout: 30000,
})

// ---- 令牌存取 ----
export const getAccessToken = () => localStorage.getItem('access_token') || ''
export const getRefreshToken = () => localStorage.getItem('refresh_token') || ''
export const setTokens = (access: string, refresh?: string) => {
  localStorage.setItem('access_token', access)
  if (refresh !== undefined) localStorage.setItem('refresh_token', refresh)
}
export const clearTokens = () => {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
}

instance.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// ---- 静默刷新（单飞：并发的 401 只触发一次刷新，其余排队等待重放） ----
let refreshing: Promise<string> | null = null

async function refreshAccessToken(): Promise<string> {
  const rt = getRefreshToken()
  if (!rt) throw new Error('no refresh token')
  // 必须带超时：刷新请求挂起会让所有排队重放的请求一起卡死
  const resp = await axios.post<ApiBody<{ access_token: string }>>(
    BACKEND_URL + '/api/auth/refresh-token',
    { refresh_token: rt },
    { timeout: 10000 },
  )
  if (resp.data.code !== 0) throw new Error(resp.data.message)
  setTokens(resp.data.data.access_token)
  return resp.data.data.access_token
}

instance.interceptors.response.use(async (response: AxiosResponse<ApiBody>) => {
  const body = response.data
  if (body.code === 40100) {
    const cfg = response.config as InternalAxiosRequestConfig & { _retried?: boolean }
    if (cfg._retried || !getRefreshToken()) {
      forceLogout()
      return response
    }
    cfg._retried = true
    try {
      refreshing = refreshing || refreshAccessToken()
      const token = await refreshing
      cfg.headers.Authorization = `Bearer ${token}`
      return instance.request(cfg)
    } catch {
      forceLogout()
      return response
    } finally {
      refreshing = null
    }
  }
  return response
})

function forceLogout() {
  clearTokens()
  if (!location.pathname.startsWith('/login')) {
    location.href = '/login'
  }
}

// ---- 业务请求封装 ----

export async function request<T = any>(config: AxiosRequestConfig): Promise<T> {
  const resp = await instance.request<ApiBody<T>>(config)
  const body = resp.data
  if (body.code !== 0) {
    throw new ApiError(body.code, body.message)
  }
  return body.data
}

export class ApiError extends Error {
  code: number
  constructor(code: number, message: string) {
    super(message)
    this.code = code
  }
}

export const get = <T = any>(url: string, params?: Record<string, any>) =>
  request<T>({ method: 'GET', url, params })

export const post = <T = any>(url: string, data?: Record<string, any>) =>
  request<T>({ method: 'POST', url, data })
