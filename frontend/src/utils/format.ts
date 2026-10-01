// 展示格式化工具
import { cstParts } from './week'

export function formatDateTime(ms: number): string {
  const p = cstParts(ms)
  const d = new Date(ms + 8 * 3600 * 1000)
  const hh = d.getUTCHours().toString().padStart(2, '0')
  const mi = d.getUTCMinutes().toString().padStart(2, '0')
  return `${p.y}-${String(p.m).padStart(2, '0')}-${String(p.d).padStart(2, '0')} ${hh}:${mi}`
}

export function formatDate(ms: number): string {
  const p = cstParts(ms)
  return `${p.y}-${String(p.m).padStart(2, '0')}-${String(p.d).padStart(2, '0')}`
}

// 相对时间：几分钟前/几小时前/几天前，超过 7 天显示日期
export function timeAgo(ms: number): string {
  const diff = Date.now() - ms
  if (diff < 60 * 1000) return '刚刚'
  if (diff < 3600 * 1000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 24 * 3600 * 1000) return `${Math.floor(diff / 3600000)} 小时前`
  if (diff < 7 * 24 * 3600 * 1000) return `${Math.floor(diff / 86400000)} 天前`
  return formatDate(ms)
}

export function formatDuration(seconds: number): string {
  const d = Math.floor(seconds / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  const parts: string[] = []
  if (d) parts.push(`${d} 天`)
  if (h) parts.push(`${h} 小时`)
  if (m) parts.push(`${m} 分`)
  if (s && !d) parts.push(`${s} 秒`)
  return parts.join(' ') || '0 秒'
}

// 比赛状态
export type ContestStatus = 'upcoming' | 'ongoing' | 'ended'

export function contestStatus(start: number, end: number): ContestStatus {
  const now = Date.now()
  if (now < start) return 'upcoming'
  if (now > end) return 'ended'
  return 'ongoing'
}

export function avatarUrl(avatar: string): string {
  if (!avatar) return '/assets/default_avatar.png'
  return avatar
}
