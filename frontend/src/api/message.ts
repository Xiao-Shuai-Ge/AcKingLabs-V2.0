import { get, post } from './http'
import type { MessageItem } from './types'

export const getMessageCount = () =>
  get<{ count: number; like_count: number; comment_count: number; system_count: number }>(
    '/api/message/count',
  )

export const getMessageList = (params: { type?: string; before?: string; count?: number }) =>
  get<{ messages: MessageItem[]; length: number }>('/api/message/list', params)

export const markMessageRead = (data: { id?: string; all?: boolean; type?: string }) =>
  post('/api/message/read', data)
