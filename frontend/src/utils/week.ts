// 周记周期算法 —— 与后端 pkg/weekcode 严格一致，固定按北京时间(UTC+8)计算，
// 与浏览器本地时区无关，跨时区设备结果一致。
// 规则：一周以周一 12:00 为锚点；周一/周二算上一周；有效打卡窗口为周日 12:00 ~ 周二 12:00。

const CST = 8 * 3600 * 1000
const DAY = 24 * 3600 * 1000

export interface CstParts {
  y: number
  m: number // 1-12
  d: number
  wd: number // ISO 星期：周一=1 ... 周日=7
}

export function cstParts(ms: number): CstParts {
  const d = new Date(ms + CST)
  const wd = d.getUTCDay() === 0 ? 7 : d.getUTCDay()
  return { y: d.getUTCFullYear(), m: d.getUTCMonth() + 1, d: d.getUTCDate(), wd }
}

// 北京时间某天的 0 点（毫秒）
function cstMidnight(y: number, m: number, d: number): number {
  return Date.UTC(y, m - 1, d) - CST
}

export interface WeekInfo {
  code: string // "2026-3-5"
  name: string // "2026年3月 第5周"
  anchor: number // 周一 12:00（北京）毫秒时间戳
  valid: boolean // 是否处于打卡窗口
}

export function getWeekCode(date: Date): WeekInfo {
  let ts = date.getTime() - 2 * DAY // 周一/周二归属上一周
  let p = cstParts(ts)
  ts -= (p.wd - 1) * DAY // 回退到本周周一
  p = cstParts(ts)
  const anchor = cstMidnight(p.y, p.m, p.d) + 12 * 3600 * 1000 // 周一 12:00

  const valid = Math.abs(date.getTime() - (anchor + 7 * DAY)) <= DAY

  // 第几周：从锚点往回数，该月中有多少个周一
  const { y, m } = cstParts(anchor)
  let n = 0
  let t = anchor
  while (cstParts(t).m === m) {
    n++
    t -= 7 * DAY
  }
  return {
    code: `${y}-${m}-${n}`,
    name: `${y}年${m}月 第${n}周`,
    anchor,
    valid,
  }
}

// 打卡窗口（周日 12:00 ~ 周二 12:00）
export function getValidSubmissionTime(date: Date) {
  const info = getWeekCode(date)
  const end = info.anchor + 7 * DAY
  return { from: end - DAY, to: end + DAY }
}

// 学习时间区间（周一 0:00 ~ 下周一 0:00）
export function getStudyTimeString(date: Date) {
  const info = getWeekCode(date)
  return { from: info.anchor - 12 * 3600 * 1000, to: info.anchor + 7 * DAY - 12 * 3600 * 1000 }
}

export function timestampFormat(ms: number): string {
  const p = cstParts(ms)
  const d = new Date(ms + CST)
  const hh = d.getUTCHours().toString().padStart(2, '0')
  const mm = d.getUTCMinutes().toString().padStart(2, '0')
  return `${p.y}-${p.m}-${p.d} ${hh}:${mm}`
}
