import { get, post } from './http'
import type { ResumeItem, ResumeExtra } from './types'
import { sendCode } from './auth'

export { sendCode }

export interface ResumeSubmitData {
  avatar: string
  real_name: string
  grade: number
  student_no: string
  email: string
  username: string // 审核通过自动开通账号时使用的用户名
  code: string
  extra: ResumeExtra
}

export const submitResume = (data: ResumeSubmitData) =>
  post<{ id: string }>('/api/resume/submit', data)

export const updateResume = (data: ResumeSubmitData & { id: string }) =>
  post('/api/resume/update', data)

export const getResumeBySelf = (email: string, code: string) =>
  get<ResumeItem>('/api/resume/detail', { email, code })

// ---- 管理端 ----
export const adminResumeList = (params: {
  keyword?: string
  status?: number
  page?: number
  count?: number
}) =>
  get<{ resumes: ResumeItem[]; total: number; page_total: number }>(
    '/api/admin/resume/list',
    params,
  )

export const adminResumeDetail = (id: string) => get<ResumeItem>('/api/admin/resume/detail', { id })

export const adminResumeStatus = (id: string, status: number) =>
  post<ResumeItem>('/api/admin/resume/status', { id, status })

export const adminDeleteResume = (id: string) => post('/api/admin/resume/delete', { id })
