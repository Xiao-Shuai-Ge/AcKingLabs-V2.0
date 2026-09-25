// 内容字数限制测试（与后端 contentLimits 保持一致）
import { describe, expect, it } from 'vitest'
import { getContentLimit } from '../contentLimit'

describe('getContentLimit', () => {
  it('普通用户', () => {
    const l = getContentLimit(1)
    expect(l.maxPostLength).toBe(15000)
    expect(l.maxCommentLength).toBe(1000)
  })
  it('正式成员', () => {
    const l = getContentLimit(2)
    expect(l.maxPostLength).toBe(30000)
    expect(l.maxCommentLength).toBe(2000)
  })
  it('管理员及以上', () => {
    const l = getContentLimit(3)
    expect(l.maxPostLength).toBe(50000)
    expect(l.maxCommentLength).toBe(5000)
    expect(getContentLimit(4).maxPostLength).toBe(50000)
  })
})
