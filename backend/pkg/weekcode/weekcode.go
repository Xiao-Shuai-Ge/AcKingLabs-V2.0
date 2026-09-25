// Package weekcode 周记打卡的周期计算。
//
// 规则（全站固定按北京时间 UTC+8 计算，与部署机器时区无关）：
//   - 一"周"以 周一 12:00 为锚点；周一/周二算作"上一周"（-2 天回退）。
//   - 有效打卡窗口：目标周结束（下一个周一 12:00）前后各 24 小时，
//     即 周日 12:00 ~ 周二 12:00。
//   - 周记周期编码（唯一键）：`年-月-第几周`，"第几周" = 该月中有多少个周一
//     （从锚点往回数到月份变化）。
package weekcode

import (
	"fmt"
	"time"
)

var cst = time.FixedZone("UTC+8", 8*3600)

// WeekInfo 某时刻对应的周记周期
type WeekInfo struct {
	Code   string    // "2026-3-5"，数据库唯一键使用
	Name   string    // "2026年3月 第5周"，展示用
	Anchor time.Time // 本周周一 12:00（北京时间）
	Valid  bool      // now 是否处于可打卡窗口
}

// NowInfo 当前时间对应的周期
func NowInfo() WeekInfo {
	return Info(time.Now())
}

// Info 计算任意时刻对应的周期
func Info(now time.Time) WeekInfo {
	now = now.In(cst)
	ts := now.Add(-48 * time.Hour) // 周一/周二归属上一周

	// 回退到本周周一
	wd := int(ts.Weekday())
	if wd == 0 {
		wd = 7
	}
	ts = ts.Add(-time.Duration(wd-1) * 24 * time.Hour)

	// 对齐到北京时间当天 0 点，再加 12 小时 => 周一 12:00 锚点
	y, m, d := ts.Date()
	anchor := time.Date(y, m, d, 12, 0, 0, 0, cst)

	// 有效窗口：锚点 + 7 天（下个周一 12:00）± 24 小时
	valid := absDuration(now.Sub(anchor.Add(7*24*time.Hour))) <= 24*time.Hour

	// "第几周"：从锚点往回数，同一月份里有多少个周一
	month, year := anchor.Month(), anchor.Year()
	n := 0
	for t := anchor; t.Month() == month; t = t.Add(-7 * 24 * time.Hour) {
		n++
	}

	return WeekInfo{
		Code:   fmt.Sprintf("%d-%d-%d", year, int(month), n),
		Name:   fmt.Sprintf("%d年%d月 第%d周", year, int(month), n),
		Anchor: anchor,
		Valid:  valid,
	}
}

// ValidWindow 返回 now 所处周期的打卡窗口（周日 12:00 ~ 周二 12:00）
func ValidWindow(now time.Time) (from, to time.Time) {
	info := Info(now)
	end := info.Anchor.Add(7 * 24 * time.Hour)
	return end.Add(-24 * time.Hour), end.Add(24 * time.Hour)
}

// StudyWindow 学习时间区间：周一 0:00 ~ 下周一 0:00
func StudyWindow(now time.Time) (from, to time.Time) {
	info := Info(now)
	return info.Anchor.Add(-12 * time.Hour), info.Anchor.Add(7*24*time.Hour - 12*time.Hour)
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
