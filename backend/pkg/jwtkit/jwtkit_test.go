package jwtkit

import (
	"testing"
	"time"

	"acking/internal/app"
	"acking/internal/config"
)

func setup(t *testing.T) {
	t.Helper()
	app.Cfg = &config.Config{}
	app.Cfg.JWT.Secret = "unit-test-secret"
}

func TestGenerateAndParse(t *testing.T) {
	setup(t)

	access, err := Generate(42, 3, TypeAccess, time.Minute)
	if err != nil {
		t.Fatalf("签发失败: %v", err)
	}
	claims, err := Parse(access, TypeAccess)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if claims.UserID != 42 || claims.Role != 3 {
		t.Errorf("claims = uid %d role %d, want 42 / 3", claims.UserID, claims.Role)
	}
}

func TestParseTypeMismatch(t *testing.T) {
	setup(t)

	refresh, _ := Generate(1, 1, TypeRefresh, time.Minute)
	if _, err := Parse(refresh, TypeAccess); err == nil {
		t.Error("refresh 令牌不应通过 access 类型校验")
	}
}

func TestParseWrongSecret(t *testing.T) {
	setup(t)
	token, _ := Generate(1, 1, TypeAccess, time.Minute)

	app.Cfg.JWT.Secret = "another-secret"
	if _, err := Parse(token, TypeAccess); err == nil {
		t.Error("换密钥后旧令牌不应通过校验")
	}
}

func TestParseGarbage(t *testing.T) {
	setup(t)
	for _, bad := range []string{"", "not-a-jwt", "a.b.c"} {
		if _, err := Parse(bad, TypeAccess); err == nil {
			t.Errorf("垃圾输入 %q 不应解析成功", bad)
		}
	}
}

func TestGeneratePair(t *testing.T) {
	setup(t)
	access, refresh, err := GeneratePair(7, 1, true)
	if err != nil || access == "" || refresh == "" {
		t.Fatalf("GeneratePair 失败: %v", err)
	}
	if access == refresh {
		t.Error("两类令牌不应相同")
	}
}
