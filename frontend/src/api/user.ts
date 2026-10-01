import { get, post } from './http'
import type { Award } from './types'

export interface UserBrief {
  id: string
  username: string
  avatar: string
  xp: number
  role: number
}

export interface UserProfile extends UserBrief {
  grade: number
  student_no?: string
  real_name?: string
  codeforces_id: string
  codeforces_rating: number
  signature: string
  awards: Award[]
}

export const me = () => get<UserBrief>('/api/user/me')

export const batchUsers = (ids: string[]) =>
  get<{ users: UserBrief[] }>('/api/user/batch', { ids: ids.join(',') })

export const getProfile = (id: string) => get<UserProfile>('/api/user/profile', { id })

export const setProfile = (
  data: Partial<{
    id: string
    username: string
    avatar: string
    signature: string
    codeforces_id: string
    awards: Award[]
  }>,
) => post('/api/user/profile', data)

// ---- 管理端 ----
export interface AdminUser extends UserProfile {
  email: string
  student_no: string
  real_name: string
  created_at: number
}

export const adminUserList = (params: {
  keyword?: string
  role?: number
  page?: number
  count?: number
}) => get<{ users: AdminUser[]; total: number; page_total: number }>('/api/admin/user/list', params)

export const adminUpdateUser = (data: Record<string, any>) => post('/api/admin/user/update', data)

export const adminSetRole = (data: { id: string; role: number }) =>
  post('/api/admin/user/role', data)

export const adminDeleteUser = (data: { id: string }) => post('/api/admin/user/delete', data)

// ---- 用户设置（分组结构：notify 通知偏好；后续可扩展其他分组） ----
export interface NotifySettings {
  like: boolean
  comment: boolean
  mention: boolean
  help_post: boolean
  system_email: boolean
  new_resume_email: boolean // 新简历邮件提醒（仅管理员生效）
}

export interface UserSettings {
  notify: NotifySettings
}

export const getUserSetting = () => get<UserSettings>('/api/user/setting')

export const updateUserSetting = (settings: UserSettings) => post('/api/user/setting', { settings })

// 修改密码（登录态，验证原密码）
export const changePassword = (data: { old_password: string; new_password: string }) =>
  post('/api/user/password', data)

// ---- 排行榜 ----
export interface RankingItem {
  rank: number
  id: string
  username: string
  avatar: string
  xp: number
  role: number
}

export const getRankings = (params: { page?: number; count?: number }) =>
  get<{ rankings: RankingItem[]; total: number; page_total: number }>('/api/user/rankings', params)
