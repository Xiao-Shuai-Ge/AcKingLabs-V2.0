package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"acking/internal/model"
	"acking/internal/service"
	"acking/pkg/response"
)

type resumeSubmitReq struct {
	Avatar    string            `json:"avatar"`
	RealName  string            `json:"real_name" binding:"required"`
	Grade     int               `json:"grade"`
	StudentNo string            `json:"student_no" binding:"required"`
	Email     string            `json:"email" binding:"required"`
	Code      string            `json:"code" binding:"required"`
	Extra     model.ResumeExtra `json:"extra"`
}

// SubmitResume POST /api/resume/submit（无需登录，凭邮箱验证码）
func SubmitResume(c *gin.Context) {
	var req resumeSubmitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	id, err := service.SubmitResume(service.ResumeSubmitReq{
		Avatar: req.Avatar, RealName: req.RealName, Grade: req.Grade,
		StudentNo: req.StudentNo, Email: req.Email, Extra: req.Extra,
	}, req.Code)
	response.Auto(c, gin.H{"id": strconv.FormatInt(id, 10)}, err)
}

type resumeUpdateReq struct {
	resumeSubmitReq
	ID int64 `json:"id,string" binding:"required"`
}

// UpdateResume POST /api/resume/update
func UpdateResume(c *gin.Context) {
	var req resumeUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	response.Auto(c, nil, service.UpdateResume(service.ResumeSubmitReq{
		Avatar: req.Avatar, RealName: req.RealName, Grade: req.Grade,
		StudentNo: req.StudentNo, Email: req.Email, Extra: req.Extra,
	}, req.Code, req.ID))
}

// GetResumeBySelf GET /api/resume/detail?email&code
func GetResumeBySelf(c *gin.Context) {
	dto, err := service.GetResumeBySelf(c.Query("email"), c.Query("code"))
	response.Auto(c, dto, err)
}

// ---- 管理端 ----

// AdminResumeList GET /api/admin/resume/list?keyword&status&page&count
func AdminResumeList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "20"))
	status := -100
	if v := c.Query("status"); v != "" {
		if s, err := strconv.Atoi(v); err == nil {
			status = s
		}
	}
	items, total, pageTotal, err := service.AdminResumeList(c.Query("keyword"), status, page, count)
	response.Auto(c, gin.H{"resumes": items, "length": total, "total": total, "page_total": pageTotal}, err)
}

// AdminGetResume GET /api/admin/resume/detail?id=
func AdminGetResume(c *gin.Context) {
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	dto, err := service.AdminGetResume(id)
	response.Auto(c, dto, err)
}

// AdminDeleteResume POST /api/admin/resume/delete
func AdminDeleteResume(c *gin.Context) {
	var req idReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	response.Auto(c, nil, service.AdminDeleteResume(req.ID))
}

type resumeStatusReq struct {
	ID     int64 `json:"id,string" binding:"required"`
	Status int   `json:"status" binding:"required"`
}

// AdminSetResumeStatus POST /api/admin/resume/status {id, status: 1待考核 2通过 -1拒绝}
func AdminSetResumeStatus(c *gin.Context) {
	var req resumeStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	dto, err := service.AdminSetResumeStatus(req.ID, req.Status)
	response.Auto(c, dto, err)
}
