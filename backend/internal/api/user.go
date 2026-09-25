package api

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"acking/internal/middleware"
	"acking/internal/model"
	"acking/internal/service"
	"acking/pkg/response"
)

func Me(c *gin.Context) {
	uid, _ := middleware.CurrentUser(c)
	dto, err := service.Me(uid)
	response.Auto(c, dto, err)
}

// BatchInfo 批量用户基础信息 GET /api/user/batch?ids=1,2,3
func BatchInfo(c *gin.Context) {
	var ids []int64
	for _, p := range strings.Split(c.Query("ids"), ",") {
		if id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		response.OK(c, gin.H{"users": []interface{}{}})
		return
	}
	if len(ids) > 100 {
		ids = ids[:100]
	}
	users, err := service.BatchInfo(ids)
	response.Auto(c, gin.H{"users": users}, err)
}

// GetProfile GET /api/user/profile?id=
func GetProfile(c *gin.Context) {
	viewerID, viewerRole := middleware.CurrentUser(c)
	targetID, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil || targetID <= 0 {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	dto, err := service.GetProfile(viewerID, viewerRole, targetID)
	response.Auto(c, dto, err)
}

type updateProfileReq struct {
	ID           int64          `json:"id,string"`
	Username     *string        `json:"username"`
	Avatar       *string        `json:"avatar"`
	Signature    *string        `json:"signature"`
	CodeforcesID *string        `json:"codeforces_id"`
	Awards       *[]model.Award `json:"awards"`
	Grade        *int           `json:"grade"`
	StudentNo    *string        `json:"student_no"`
	RealName     *string        `json:"real_name"`
	Xp           *int           `json:"xp"`
}

// UpdateProfile POST /api/user/profile（本人或管理员）
func UpdateProfile(c *gin.Context) {
	var req updateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil || req.ID <= 0 {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	uid, role := middleware.CurrentUser(c)
	err := service.UpdateProfile(service.UpdateProfileReq{
		OperatorID: uid, OperatorRole: role, TargetID: req.ID,
		Username: req.Username, Avatar: req.Avatar, Signature: req.Signature,
		CodeforcesID: req.CodeforcesID, Awards: req.Awards,
		Grade: req.Grade, StudentNo: req.StudentNo, RealName: req.RealName,
	})
	response.Auto(c, nil, err)
}

// GetUserSetting GET /api/user/setting（返回分组设置，当前含 notify）
func GetUserSetting(c *gin.Context) {
	uid, _ := middleware.CurrentUser(c)
	s, err := service.GetUserSettings(uid)
	response.Auto(c, s, err)
}

type updateUserSettingReq struct {
	Settings model.UserSettings `json:"settings" binding:"required"`
}

// UpdateUserSetting POST /api/user/setting
func UpdateUserSetting(c *gin.Context) {
	var req updateUserSettingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	uid, _ := middleware.CurrentUser(c)
	response.Auto(c, nil, service.UpdateUserSettings(uid, req.Settings))
}

// GetRankings GET /api/user/rankings?page&count
func GetRankings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "10"))
	items, total, pageTotal, err := service.GetRankings(page, count)
	response.Auto(c, gin.H{
		"rankings": items, "length": total, "total": total, "page_total": pageTotal,
	}, err)
}

type setRoleReq struct {
	ID   int64 `json:"id,string" binding:"required"`
	Role *int  `json:"role" binding:"required"`
}

// SetRole POST /api/admin/user/role（仅超管）
func SetRole(c *gin.Context) {
	var req setRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	_, role := middleware.CurrentUser(c)
	response.Auto(c, nil, service.SetRole(role, req.ID, *req.Role))
}

// AdminUserList GET /api/admin/user/list?keyword&role&page&count
func AdminUserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "20"))
	var rolePtr *int
	if v := c.Query("role"); v != "" {
		if r, err := strconv.Atoi(v); err == nil {
			rolePtr = &r
		}
	}
	items, total, pageTotal, err := service.AdminUserList(c.Query("keyword"), rolePtr, page, count)
	response.Auto(c, gin.H{
		"users": items, "length": total, "total": total, "page_total": pageTotal,
	}, err)
}

// AdminUpdateUser POST /api/admin/user/update（管理员编辑用户资料，可改经验值）
func AdminUpdateUser(c *gin.Context) {
	var req updateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil || req.ID <= 0 {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	uid, role := middleware.CurrentUser(c)
	err := service.UpdateProfile(service.UpdateProfileReq{
		OperatorID: uid, OperatorRole: role, TargetID: req.ID,
		Username: req.Username, Avatar: req.Avatar, Signature: req.Signature,
		CodeforcesID: req.CodeforcesID, Awards: req.Awards,
		Grade: req.Grade, StudentNo: req.StudentNo, RealName: req.RealName,
		Xp: req.Xp,
	})
	response.Auto(c, nil, err)
}

type idReq struct {
	ID int64 `json:"id,string" binding:"required"`
}

// AdminDeleteUser POST /api/admin/user/delete
func AdminDeleteUser(c *gin.Context) {
	var req idReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	uid, _ := middleware.CurrentUser(c)
	response.Auto(c, nil, service.DeleteUserAdmin(uid, req.ID))
}
