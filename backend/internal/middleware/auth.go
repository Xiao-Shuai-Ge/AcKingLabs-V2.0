// Package middleware Gin 中间件：鉴权、限流、跨域。
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"acking/internal/model"
	"acking/pkg/jwtkit"
	"acking/pkg/response"
)

const (
	CtxUserID = "user_id"
	CtxRole   = "role"
)

// BearerToken 从请求头取令牌
func BearerToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	parts := strings.SplitN(h, " ", 3)
	if len(parts) == 2 && parts[0] == "Bearer" {
		return parts[1]
	}
	return ""
}

// Auth 强制登录且角色达标（minRole 见 model.Role* 常量）
func Auth(minRole int) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := BearerToken(c)
		if token == "" {
			response.FailCode(c, response.CodeUnauthorized)
			c.Abort()
			return
		}
		claims, err := jwtkit.Parse(token, jwtkit.TypeAccess)
		if err != nil {
			response.FailCode(c, response.CodeUnauthorized)
			c.Abort()
			return
		}
		if claims.Role < minRole {
			response.FailCode(c, response.CodeForbidden)
			c.Abort()
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

// OptionalAuth 有令牌就解析（游客 user_id=0 / role=-1）
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(CtxUserID, int64(0))
		c.Set(CtxRole, -1)
		token := BearerToken(c)
		if token == "" {
			c.Next()
			return
		}
		if claims, err := jwtkit.Parse(token, jwtkit.TypeAccess); err == nil {
			c.Set(CtxUserID, claims.UserID)
			c.Set(CtxRole, claims.Role)
		}
		c.Next()
	}
}

// CurrentUser 取当前用户（需经过 Auth/OptionalAuth）
func CurrentUser(c *gin.Context) (userID int64, role int) {
	uid, _ := c.Get(CtxUserID)
	rolev, _ := c.Get(CtxRole)
	id, _ := uid.(int64)
	r, _ := rolev.(int)
	return id, r
}

// IsAdmin 是否管理员
func IsAdmin(role int) bool { return role >= model.RoleAdmin }
