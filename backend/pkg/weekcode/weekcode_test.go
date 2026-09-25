package weekcode

import (
	"testing"
	"time"
)

// 北京时间构造：YYYY-MM-DD HH:MM
func cstTime(t *testing.T, s string) time.Time {
	t.Helper()
	loc := time.FixedZone("CST", 8*3600)
	tm, err := time.ParseInLocation("2006-01-02 15:04", s, loc)
	if err != nil {
		t.Fatalf("解析时间 %q 失败: %v", s, err)
	}
	return tm
}

func TestWeekInfo(t *testing.T) {
	tests := []struct {
		name     string
		at       string // 北京时间
		code     string
		valid    bool
		weekName string
	}{
		// 周一中午前后：12:00 前属于上一周，之后仍属于上一周（到周二 12:00 截止）
		{"周一上午属于上一周期", "2026-09-21 10:00", "2026-9-2", true, "2026年9月 第2周"},
		{"周一晚上仍在窗口内", "2026-09-21 22:00", "2026-9-2", true, "2026年9月 第2周"},
		{"周二上午仍在窗口内", "2026-09-22 11:59", "2026-9-2", true, "2026年9月 第2周"},
		{"周二中午窗口关闭（周期尚未翻转）", "2026-09-22 12:01", "2026-9-2", false, "2026年9月 第2周"},
		{"周日上午窗口未开", "2026-09-20 11:59", "2026-9-2", false, "2026年9月 第2周"},
		{"周日中午窗口开启", "2026-09-20 12:01", "2026-9-2", true, "2026年9月 第2周"},
		{"周三不在窗口", "2026-09-23 15:00", "2026-9-3", false, "2026年9月 第3周"},
		{"月末周二上午仍在窗口", "2026-09-29 10:00", "2026-9-3", true, "2026年9月 第3周"},
		{"月末周二中午窗口关闭", "2026-09-29 12:01", "2026-9-3", false, "2026年9月 第3周"},
		// 跨月：10 月第 1 周的周一是 10-05，10-01（周四）归 9 月第 4 周
		{"跨月归属", "2026-10-01 10:00", "2026-9-4", false, "2026年9月 第4周"},
		{"跨月周二上午仍属上月周期", "2026-10-06 10:00", "2026-9-4", true, "2026年9月 第4周"},
		{"跨月周三翻转到下月周期", "2026-10-07 10:00", "2026-10-1", false, "2026年10月 第1周"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Info(cstTime(t, tt.at))
			if got.Code != tt.code {
				t.Errorf("code = %q, want %q", got.Code, tt.code)
			}
			if got.Name != tt.weekName {
				t.Errorf("name = %q, want %q", got.Name, tt.weekName)
			}
			if got.Valid != tt.valid {
				t.Errorf("valid = %v, want %v", got.Valid, tt.valid)
			}
		})
	}
}

func TestValidWindow(t *testing.T) {
	from, to := ValidWindow(cstTime(t, "2026-09-21 10:00"))
	// 目标周结束于周一(09-21)12:00，窗口为前后各 24h
	wantFrom := cstTime(t, "2026-09-20 12:00")
	wantTo := cstTime(t, "2026-09-22 12:00")
	if !from.Equal(wantFrom) || !to.Equal(wantTo) {
		t.Errorf("window = [%v, %v], want [%v, %v]", from, to, wantFrom, wantTo)
	}
}

func TestStudyWindow(t *testing.T) {
	from, to := StudyWindow(cstTime(t, "2026-09-21 10:00"))
	// 学习区间：周一 0:00 ~ 下周一 0:00
	if from.Format("01-02 15:04") != "09-14 00:00" || to.Format("01-02 15:04") != "09-21 00:00" {
		t.Errorf("study window = [%v, %v]", from, to)
	}
}

func TestInfoDeterministic(t *testing.T) {
	// 同一时刻多次计算结果一致
	a := Info(cstTime(t, "2026-09-21 10:00"))
	for i := 0; i < 10; i++ {
		b := Info(cstTime(t, "2026-09-21 10:00"))
		if a != b {
			t.Fatalf("结果不稳定: %+v vs %+v", a, b)
		}
	}
}
