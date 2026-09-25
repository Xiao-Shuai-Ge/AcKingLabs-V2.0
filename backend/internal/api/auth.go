// Package api HTTP 处理层：只做参数绑定与响应，业务在 service。
package api

import (
	"github.com/gin-gonic/gin"

	"acking/internal/service"
	"acking/pkg/response"
)

type sendCodeReq struct {
	Email string `json:"email" binding:"required"`
}

func SendCode(c *gin.Context) {
	var req sendCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	response.Auto(c, nil, service.SendCode(req.Email))
}

type registerReq struct {
	Email    string `json:"email" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required"`
	Username string `json:"username" binding:"required"`
	Invite   string `json:"invitation_code"`
}

func Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	access, refresh, err := service.Register(service.RegisterReq{
		Email: req.Email, Code: req.Code, Password: req.Password,
		Username: req.Username, Invite: req.Invite,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"access_token": access, "refresh_token": refresh})
}

type loginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Remember bool   `json:"is_remember"`
}

func Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	access, refresh, err := service.Login(req.Email, req.Password, req.Remember)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"access_token": access, "refresh_token": refresh})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func RefreshToken(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	access, err := service.Refresh(req.RefreshToken)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"access_token": access})
}

type resetPasswordReq struct {
	Email    string `json:"email" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func ResetPassword(c *gin.Context) {
	var req resetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	response.Auto(c, nil, service.ResetPassword(req.Email, req.Code, req.Password))
}
