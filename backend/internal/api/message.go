package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"acking/internal/middleware"
	"acking/internal/service"
	"acking/pkg/response"
)

// GetMessageCounts GET /api/message/count
func GetMessageCounts(c *gin.Context) {
	uid, _ := middleware.CurrentUser(c)
	dto, err := service.GetMessageCounts(uid)
	response.Auto(c, dto, err)
}

// GetMessageList GET /api/message/list?type=like|comment|system|all&before&count
func GetMessageList(c *gin.Context) {
	uid, _ := middleware.CurrentUser(c)
	before, _ := strconv.ParseInt(c.Query("before"), 10, 64)
	count, _ := strconv.ParseInt(c.DefaultQuery("count", "20"), 10, 64)
	list, err := service.GetMessageList(uid, c.Query("type"), before, count)
	response.Auto(c, gin.H{"messages": list, "length": len(list)}, err)
}

type messageReadReq struct {
	ID   int64  `json:"id,string"`
	All  bool   `json:"all"`
	Type string `json:"type"`
}

// MarkMessageRead POST /api/message/read（单条已读 / 全部已读）
func MarkMessageRead(c *gin.Context) {
	var req messageReadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	uid, _ := middleware.CurrentUser(c)
	response.Auto(c, nil, service.MarkRead(uid, req.ID, req.All, req.Type))
}
