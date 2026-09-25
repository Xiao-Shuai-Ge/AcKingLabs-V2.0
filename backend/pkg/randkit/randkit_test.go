package randkit

import (
	"strings"
	"testing"
)

func TestDigitCode(t *testing.T) {
	for _, n := range []int{1, 6, 8} {
		code := DigitCode(n)
		if len(code) != n {
			t.Errorf("长度 = %d, want %d", len(code), n)
		}
		if strings.Trim(code, "0123456789") != "" {
			t.Errorf("出现非数字字符: %q", code)
		}
	}
}

func TestLetterCode(t *testing.T) {
	const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	for _, n := range []int{1, 6, 16} {
		code := LetterCode(n)
		if len(code) != n {
			t.Errorf("长度 = %d, want %d", len(code), n)
		}
		if strings.Trim(code, letters) != "" {
			t.Errorf("出现字符集外字符: %q", code)
		}
	}
	// 6 位码不应出现易混淆的 I / O
	if strings.ContainsAny(LetterCode(64), "IO") {
		t.Error("邀请码出现了易混淆字符 I/O")
	}
}

func TestCodesNotConstant(t *testing.T) {
	// 连续生成的码不应全部相同（概率意义上）
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		seen[DigitCode(6)] = true
	}
	if len(seen) < 10 {
		t.Errorf("随机性可疑，50 次只出现 %d 个不同结果", len(seen))
	}
}
