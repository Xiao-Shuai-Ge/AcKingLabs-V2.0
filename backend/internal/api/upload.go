package api

import (
	"github.com/gin-gonic/gin"

	"acking/internal/service"
	"acking/pkg/response"
)

// UploadImage POST /api/file/upload（multipart 字段名 file；需登录）
func UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	url, err := service.UploadImage(file)
	response.Auto(c, gin.H{"url": url}, err)
}
