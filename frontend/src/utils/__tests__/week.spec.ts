// 周记周期算法测试（与后端 pkg/weekcode 保持同一套规则）
import { describe, expect, it } from 'vitest'
import { getWeekCode, getValidSubmissionTime, getStudyTimeString } from '../week'

// 带时区偏移的时间字符串 -> 毫秒时间戳
const t = (s: string) => new Date(s).getTime()

describe('getWeekCode', () => {
  it('周一上午归属上一周期且在窗口内', () => {
    const w = getWeekCode(new Date(t('2026-09-21T10:00:00+08:00'))) // 北京 10:00
    expect(w.code).toBe('2026-9-2')
    expect(w.name).toBe('2026年9月 第2周')
    expect(w.valid).toBe(true)
  })

  it('周一晚上仍在窗口内', () => {
    expect(getWeekCode(new Date(t('2026-09-21T22:00:00+08:00'))).valid).toBe(true)
  })

  it('周二上午仍在窗口内', () => {
    expect(getWeekCode(new Date(t('2026-09-22T11:59:00+08:00'))).valid).toBe(true)
  })

  it('周二中午窗口关闭', () => {
    const w = getWeekCode(new Date(t('2026-09-22T12:01:00+08:00'))) // 北京 12:01
    expect(w.code).toBe('2026-9-2')
    expect(w.valid).toBe(false)
  })

  it('周三翻转到本周', () => {
    const w = getWeekCode(new Date(t('2026-09-23T15:00:00+08:00')))
    expect(w.code).toBe('2026-9-3')
    expect(w.valid).toBe(false)
  })

  it('周日中午窗口开启', () => {
    expect(getWeekCode(new Date(t('2026-09-20T12:01:00+08:00'))).valid).toBe(true)
  })

  it('周日上午窗口未开', () => {
    expect(getWeekCode(new Date(t('2026-09-20T11:59:00+08:00'))).valid).toBe(false)
  })

  it('跨月归属', () => {
    expect(getWeekCode(new Date(t('2026-10-01T10:00:00+08:00'))).code).toBe('2026-9-4')
    expect(getWeekCode(new Date(t('2026-10-07T10:00:00+08:00'))).code).toBe('2026-10-1')
  })
})

describe('打卡/学习窗口', () => {
  it('打卡窗口为周日 12:00 ~ 周二 12:00（北京时间）', () => {
    const { from, to } = getValidSubmissionTime(new Date(t('2026-09-21T10:00:00+08:00')))
    expect(new Date(from).toISOString()).toBe('2026-09-20T04:00:00.000Z')
    expect(new Date(to).toISOString()).toBe('2026-09-22T04:00:00.000Z')
  })

  it('学习窗口为周一 0:00 ~ 下周一 0:00（北京时间）', () => {
    const { from, to } = getStudyTimeString(new Date(t('2026-09-21T10:00:00+08:00')))
    expect(new Date(from).toISOString()).toBe('2026-09-13T16:00:00.000Z') // 09-14 00:00 +08
    expect(new Date(to).toISOString()).toBe('2026-09-20T16:00:00.000Z') // 09-21 00:00 +08
  })
})
