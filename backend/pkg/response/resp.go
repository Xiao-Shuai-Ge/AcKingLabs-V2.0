// Package response 统一响应封装：永远 HTTP 200 + {"code", "message", "data"}。
// code=0 成功；非 0 见 codes.go。错误信息不得包含 SQL / 内部细节。
package response

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// AppErr 业务错误，service 层返回它，api 层统一翻译成响应
type AppErr struct {
	Code int
	Msg  string
}

func (e *AppErr) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return Message(e.Code)
}

// NewErr 构造业务错误（使用错误码默认文案）
func NewErr(code int) error {
	return &AppErr{Code: code}
}

// NewErrMsg 构造带自定义文案的业务错误
func NewErrMsg(code int, msg string) error {
	return &AppErr{Code: code, Msg: msg}
}

// AsAppErr 取出业务错误；非业务错误返回 nil
func AsAppErr(err error) *AppErr {
	var ae *AppErr
	if errors.As(err, &ae) {
		return ae
	}
	return nil
}

func OK(c *gin.Context, data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	c.JSON(200, Body{Code: CodeOK, Message: Message(CodeOK), Data: data})
}

// Auto 服务调用快捷封装：err 非 nil 时返回错误，否则返回 data
func Auto(c *gin.Context, data interface{}, err error) {
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, data)
}

// Fail 返回业务错误；err 不是 AppErr 时按内部错误处理（不透出细节）
func Fail(c *gin.Context, err error) {
	if ae := AsAppErr(err); ae != nil {
		c.JSON(200, Body{Code: ae.Code, Message: ae.Error(), Data: gin.H{}})
		return
	}
	c.JSON(200, Body{Code: CodeInternal, Message: Message(CodeInternal), Data: gin.H{}})
}

// FailCode 直接按错误码返回
func FailCode(c *gin.Context, code int) {
	c.JSON(200, Body{Code: code, Message: Message(code), Data: gin.H{}})
}

// FailMsg 按错误码 + 自定义文案返回
func FailMsg(c *gin.Context, code int, msg string) {
	c.JSON(200, Body{Code: code, Message: msg, Data: gin.H{}})
}

// Errf 快捷构造
func Errf(code int, format string, args ...interface{}) error {
	return &AppErr{Code: code, Msg: fmt.Sprintf(format, args...)}
}
