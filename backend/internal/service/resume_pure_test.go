package service

import (
	"testing"

	"acking/internal/model"
)

func TestResumeTransitionAllowed(t *testing.T) {
	tests := []struct {
		name string
		from int
		to   int
		want bool
	}{
		{"待审核 -> 已通过", model.ResumePending, model.ResumeApproved, true},
		{"待审核 -> 未通过", model.ResumePending, model.ResumeRejected, true},
		{"待审核 -> 待审核", model.ResumePending, model.ResumePending, false},
		{"已通过 -> 未通过（终态）", model.ResumeApproved, model.ResumeRejected, false},
		{"未通过 -> 已通过（终态，重新投递需本人修改）", model.ResumeRejected, model.ResumeApproved, false},
		{"未通过 -> 未通过", model.ResumeRejected, model.ResumeRejected, false},
		{"非法目标状态", model.ResumePending, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resumeTransitionAllowed(tt.from, tt.to); got != tt.want {
				t.Errorf("resumeTransitionAllowed(%d, %d) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestResumeStatusEditable(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{"待审核可修改", model.ResumePending, true},
		{"未通过可修改后重新投递", model.ResumeRejected, true},
		{"已通过不可修改", model.ResumeApproved, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resumeStatusEditable(tt.status); got != tt.want {
				t.Errorf("resumeStatusEditable(%d) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestValidateResumeUsername(t *testing.T) {
	base := func(username string) ResumeSubmitReq {
		return ResumeSubmitReq{
			RealName: "张三", Grade: 25, StudentNo: "20250101",
			Email: "test@example.com", Username: username,
		}
	}
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"两字用户名", "小明", false},
		{"30 字用户名", "一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十", false},
		{"单字用户名", "明", true},
		{"超长用户名", "一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一", true},
		{"空用户名", "", true},
		{"首尾空格会被裁剪", "  小明  ", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := base(tt.username)
			err := validateResume(&req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateResume(username=%q) error = %v, wantErr %v", tt.username, err, tt.wantErr)
			}
			if !tt.wantErr && req.Username != "小明" && tt.username == "  小明  " {
				t.Errorf("用户名应裁剪首尾空格, got %q", req.Username)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"6 位密码", "123456", false},
		{"30 位密码", "abcdefghijklmnopqrstuvwxyz1234", false},
		{"空密码（投递时不合格）", "", true},
		{"5 位密码", "12345", true},
		{"31 位密码", "abcdefghijklmnopqrstuvwxyz12345", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePassword(len=%d) error = %v, wantErr %v", len(tt.password), err, tt.wantErr)
			}
		})
	}
}
