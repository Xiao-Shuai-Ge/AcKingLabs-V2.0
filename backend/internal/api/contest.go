package api

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"acking/internal/middleware"
	"acking/internal/service"
	"acking/pkg/response"
)

// GetContestList GET /api/contest/list?platform=all|recommend|Codeforces|...&page&count
func GetContestList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "5"))
	items, total, pageTotal, err := service.GetContestList(c.DefaultQuery("platform", c.DefaultQuery("type", "all")), page, count)
	response.Auto(c, gin.H{"contests": items, "length": total, "total": total, "page_total": pageTotal}, err)
}

// GetContestDetail GET /api/contest/detail?id=
func GetContestDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	dto, err := service.GetContestDetail(id)
	response.Auto(c, gin.H{"contest": dto}, err)
}

type bookingReq struct {
	ContestID int64 `json:"contest_id,string" binding:"required"`
}

// ToggleBooking POST /api/contest/booking
func ToggleBooking(c *gin.Context) {
	var req bookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	uid, _ := middleware.CurrentUser(c)
	booked, err := service.ToggleBooking(uid, req.ContestID)
	response.Auto(c, gin.H{"booked": booked}, err)
}

// GetBookingMap GET /api/contest/booking?ids=1,2,3（批量，消灭 N+1）
func GetBookingMap(c *gin.Context) {
	uid, _ := middleware.CurrentUser(c)
	var ids []int64
	for _, p := range strings.Split(c.Query("ids"), ",") {
		if id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	if len(ids) > 100 {
		ids = ids[:100]
	}
	m, err := service.GetBookedMap(uid, ids)
	response.Auto(c, gin.H{"booked": m}, err)
}

// ---- 管理端 ----

type contestWriteReq struct {
	ID        int64  `json:"id,string"`
	Title     string `json:"title" binding:"required"`
	StartTime int64  `json:"start_time" binding:"required"`
	EndTime   int64  `json:"end_time" binding:"required"`
	Url       string `json:"url" binding:"required"`
}

// AdminCreateContest POST /api/admin/contest/create
func AdminCreateContest(c *gin.Context) {
	var req contestWriteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	id, err := service.AdminCreateContest(service.ContestWriteReq{
		Title: req.Title, StartTime: req.StartTime, EndTime: req.EndTime, Url: req.Url,
	})
	response.Auto(c, gin.H{"id": strconv.FormatInt(id, 10)}, err)
}

// AdminUpdateContest POST /api/admin/contest/update
func AdminUpdateContest(c *gin.Context) {
	var req contestWriteReq
	if err := c.ShouldBindJSON(&req); err != nil || req.ID <= 0 {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	response.Auto(c, nil, service.AdminUpdateContest(req.ID, service.ContestWriteReq{
		Title: req.Title, StartTime: req.StartTime, EndTime: req.EndTime, Url: req.Url,
	}))
}

type contestIdReq struct {
	ContestID int64 `json:"contest_id,string" binding:"required"`
}

// AdminDeleteContest POST /api/admin/contest/delete
func AdminDeleteContest(c *gin.Context) {
	var req contestIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	response.Auto(c, nil, service.AdminDeleteContest(req.ContestID))
}

type contestRecommendReq struct {
	ContestID int64 `json:"contest_id,string" binding:"required"`
	Value     bool  `json:"value"`
}

// AdminSetRecommend POST /api/admin/contest/recommend
func AdminSetRecommend(c *gin.Context) {
	var req contestRecommendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	response.Auto(c, nil, service.AdminSetRecommend(req.ContestID, req.Value))
}
