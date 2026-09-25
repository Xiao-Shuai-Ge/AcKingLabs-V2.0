// 等级体系测试
import { describe, expect, it } from 'vitest'
import {
  CheckLevel,
  GetRoleName,
  GetRoleLabel,
  GetTextColor,
  GetBgColor,
  NextLevelLimit,
} from '../level'

describe('CheckLevel', () => {
  it('未实名 -> 0；未登录(传 -1)落在兜底档且颜色走默认灰', () => {
    expect(CheckLevel(0, 0)).toBe(0)
    expect(CheckLevel(0, -1)).toBe(-1)
    expect(GetTextColor(CheckLevel(0, -1))).toBe('text-gray-500')
    expect(GetRoleName(CheckLevel(0, -1))).toBe('未实名游客')
  })
  it('零经验注册用户 -> 1', () => {
    expect(CheckLevel(0, 1)).toBe(1)
  })
  it('经验决定普通等级', () => {
    expect(CheckLevel(40, 1)).toBe(2)
    expect(CheckLevel(100, 1)).toBe(3)
    expect(CheckLevel(200, 1)).toBe(4)
  })
  it('黄名需要正式成员角色', () => {
    expect(CheckLevel(200, 1)).toBe(4) // 紫名
    expect(CheckLevel(200, 2)).toBe(5) // 黄名
  })
  it('红名对应管理员', () => {
    expect(CheckLevel(400, 3)).toBe(7)
    expect(CheckLevel(400, 2)).toBe(6) // 橙名
  })
})

describe('称号与角色', () => {
  it('等级称号', () => {
    expect(GetRoleName(0)).toBe('未实名游客')
    expect(GetRoleName(1)).toBe('新人')
    expect(GetRoleName(5)).toBe('主力成员')
    expect(GetRoleName(6)).toBe('传奇成员')
    expect(GetRoleName(7)).toBe('管理员')
  })
  it('角色身份名（与管理后台一致）', () => {
    expect(GetRoleLabel(0)).toBe('游客')
    expect(GetRoleLabel(1)).toBe('普通用户')
    expect(GetRoleLabel(2)).toBe('正式成员')
    expect(GetRoleLabel(3)).toBe('管理员')
    expect(GetRoleLabel(4)).toBe('超级管理员')
  })
})

describe('颜色', () => {
  it('等级文字色', () => {
    expect(GetTextColor(0)).toBe('text-gray-500')
    expect(GetTextColor(7)).toBe('text-red-500')
  })
  it('等级背景色', () => {
    expect(GetBgColor(7)).toBe('bg-red-500')
  })
})

describe('NextLevelLimit', () => {
  it('下一等级门槛', () => {
    expect(NextLevelLimit(0, 1)).toBe(40) // Lv1 的下一档是 40xp
    expect(NextLevelLimit(40, 1)).toBe(100)
  })
})
