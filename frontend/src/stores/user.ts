// 当前登录用户状态
import { defineStore } from 'pinia'
import { me } from '@/api/user'
import { clearTokens, getAccessToken, getRefreshToken } from '@/api/http'

export interface CurrentUser {
  id: string
  username: string
  avatar: string
  xp: number
  role: number
}

export const useUserStore = defineStore('user', {
  // 仅当本地还有令牌时才从缓存恢复用户，避免"幽灵登录态"（令牌已被清除、
  // 缓存用户却让界面显示已登录）。
  state: () => ({
    currentUser: getAccessToken()
      ? loadCachedUser()
      : (localStorage.removeItem('cached_user'), null),
  }),
  getters: {
    // 登录态只看 currentUser：它仅在拿到有效令牌并成功拉取用户信息后被赋值，
    // 令牌失效时 refreshUser/logout 会将其清空，单独即可代表登录态。
    // 注意：不要在 getter 里混合 localStorage 等非响应式条件——若 && 首个条件
    // 为假，后续响应式依赖不会被读取/追踪，getter 将缓存过期值不再更新。
    isLogin: (state) => !!state.currentUser,
    isAdmin: (state) => (state.currentUser?.role ?? -1) >= 3,
  },
  actions: {
    async refreshUser(): Promise<boolean> {
      if (!getAccessToken() || !getRefreshToken()) {
        this.logout()
        return false
      }
      try {
        this.currentUser = await me()
        cacheUser(this.currentUser)
        return true
      } catch {
        // 静默刷新失败时由 http 层统一处理登出
        return false
      }
    },
    setUser(user: CurrentUser | null) {
      this.currentUser = user
      cacheUser(user)
    },
    logout() {
      clearTokens()
      this.currentUser = null
      localStorage.removeItem('cached_user')
    },
  },
})

function loadCachedUser(): CurrentUser | null {
  try {
    const raw = localStorage.getItem('cached_user')
    return raw ? (JSON.parse(raw) as CurrentUser) : null
  } catch {
    return null
  }
}

function cacheUser(user: CurrentUser | null) {
  if (user) localStorage.setItem('cached_user', JSON.stringify(user))
  else localStorage.removeItem('cached_user')
}
