// 帖子类型元信息与格式化工具测试
import { describe, expect, it } from 'vitest'
import { postTypeName, postTypeClass } from '../postMeta'
import { formatDuration, contestStatus } from '../format'

describe('postMeta', () => {
  it('类型名', () => {
    expect(postTypeName('diary')).toBe('周记')
    expect(postTypeName('solution')).toBe('题解')
    expect(postTypeName('help')).toBe('求助')
    expect(postTypeName('unknown-x')).toBe('unknown-x')
  })
  it('角标配色包含文字色与边框色', () => {
    expect(postTypeClass('official')).toContain('border-red-500')
    expect(postTypeClass('nope')).toContain('border-gray-400')
  })
})

describe('formatDuration', () => {
  it('秒/分/小时/天', () => {
    expect(formatDuration(30)).toBe('30 秒')
    expect(formatDuration(90)).toBe('1 分 30 秒')
    expect(formatDuration(3600)).toBe('1 小时')
    expect(formatDuration(90000)).toContain('1 天')
  })
  it('零', () => {
    expect(formatDuration(0)).toBe('0 秒')
  })
})

describe('contestStatus', () => {
  const now = Date.now()
  it('按当前时间判定状态', () => {
    expect(contestStatus(now + 1000, now + 2000)).toBe('upcoming')
    expect(contestStatus(now - 1000, now + 2000)).toBe('ongoing')
    expect(contestStatus(now - 2000, now - 1000)).toBe('ended')
  })
})
