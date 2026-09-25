import { get, post } from './http'
import type { ContestItem } from './types'

export const getContestList = (params: { platform?: string; page?: number; count?: number }) =>
  get<{ contests: ContestItem[]; total: number; page_total: number }>('/api/contest/list', params)

export const getContestDetail = (id: string) =>
  get<{ contest: ContestItem }>('/api/contest/detail', { id })

export const toggleBooking = (contestId: string) =>
  post<{ booked: boolean }>('/api/contest/booking', { contest_id: contestId })

export const getBookingMap = (ids: string[]) =>
  get<{ booked: Record<string, boolean> }>('/api/contest/booking', { ids: ids.join(',') })

// ---- 管理端 ----
export const adminCreateContest = (data: {
  title: string
  start_time: number
  end_time: number
  url: string
}) => post<{ id: string }>('/api/admin/contest/create', data)

export const adminUpdateContest = (data: {
  id: string
  title: string
  start_time: number
  end_time: number
  url: string
}) => post('/api/admin/contest/update', data)

export const adminDeleteContest = (contestId: string) =>
  post('/api/admin/contest/delete', { contest_id: contestId })

export const adminSetRecommend = (contestId: string, value: boolean) =>
  post('/api/admin/contest/recommend', { contest_id: contestId, value })
