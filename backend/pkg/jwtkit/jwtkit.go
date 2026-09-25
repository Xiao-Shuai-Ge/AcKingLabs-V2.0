// Package jwtkit JWT 双令牌：access（2h，接口鉴权用）/ refresh（14d，换新 access）。
package jwtkit

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"acking/internal/app"
	"acking/internal/config"
)

const (
	TypeAccess  = "access"
	TypeRefresh = "refresh"
)

type Claims struct {
	UserID int64  `json:"uid"`
	Role   int    `json:"role"`
	Typ    string `json:"typ"`
	jwt.RegisteredClaims
}

// Generate 签发令牌
func Generate(userID int64, role int, typ string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		Typ:    typ,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(app.Cfg.JWT.Secret))
}

// GeneratePair 签发 access + refresh
func GeneratePair(userID int64, role int, remember bool) (access, refresh string, err error) {
	refreshTTL := config.RefreshTTL
	if !remember {
		refreshTTL = config.RefreshTTLShort
	}
	access, err = Generate(userID, role, TypeAccess, config.AccessTTL)
	if err != nil {
		return
	}
	refresh, err = Generate(userID, role, TypeRefresh, refreshTTL)
	return
}

// Parse 解析并校验令牌；typ 传空串表示不限制类型
func Parse(tokenStr, typ string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("签名算法不合法")
		}
		return []byte(app.Cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("令牌无效")
	}
	if typ != "" && claims.Typ != typ {
		return nil, errors.New("令牌类型不匹配")
	}
	return claims, nil
}
