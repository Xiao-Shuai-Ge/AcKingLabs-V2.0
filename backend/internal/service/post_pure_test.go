package service

import (
	"errors"
	"testing"
)

func TestContentShort(t *testing.T) {
	tests := []struct {
		name  string
		input string
		max   int
		want  string
	}{
		{"短文本原样", "hello", 10, "hello"},
		{"英文截断加省略号", "hello world", 5, "hello..."},
		{"中文按字符不按字节", "一二三四五六七八九十", 4, "一二三四..."},
		{"emoji 不会被截坏", "😀😀😀😀", 2, "😀😀..."},
		{"刚好等于上限", "abcde", 5, "abcde"},
		{"空文本", "", 3, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contentShort(tt.input, tt.max); got != tt.want {
				t.Errorf("contentShort(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
			}
		})
	}
}

func TestTruncateRunes(t *testing.T) {
	if got := truncateRunes("一二三", 5); got != "一二三" {
		t.Errorf("短文本应原样: %q", got)
	}
	if got := truncateRunes("一二三四五", 2); got != "一二" {
		t.Errorf("截断结果: %q", got)
	}
}

func TestContentLimits(t *testing.T) {
	tests := []struct {
		role             int
		wantPost         int
		wantComment      int
	}{
		{0, 15000, 1000},
		{1, 15000, 1000},
		{2, 30000, 2000},
		{3, 50000, 5000},
		{4, 50000, 5000},
	}
	for _, tt := range tests {
		p, c := contentLimits(tt.role)
		if p != tt.wantPost || c != tt.wantComment {
			t.Errorf("role %d: got (%d,%d), want (%d,%d)", tt.role, p, c, tt.wantPost, tt.wantComment)
		}
	}
}

func TestDeref(t *testing.T) {
	if got := deref(nil); got != "" {
		t.Errorf("nil 应返回空串: %q", got)
	}
	s := "2026-9-3"
	if got := deref(&s); got != s {
		t.Errorf("应返回指针指向的值: %q", got)
	}
}

func TestIsDuplicateEntry(t *testing.T) {
	if isDuplicateEntry(nil) {
		t.Error("nil 不是重复键错误")
	}
	if isDuplicateEntry(errors.New("some error")) {
		t.Error("普通错误不是重复键错误")
	}
	if !isDuplicateEntry(errors.New("Error 1062 (23000): Duplicate entry '1' for key 'PRIMARY'")) {
		t.Error("1062 错误应被识别为重复键")
	}
}
