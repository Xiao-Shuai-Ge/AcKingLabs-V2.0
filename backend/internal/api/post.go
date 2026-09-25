package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"acking/internal/middleware"
	"acking/internal/service"
	"acking/pkg/response"
)

type postWriteReq struct {
	Title     string `json:"title"`
	Content   string `json:"content" binding:"required"`
	Type      string `json:"type" binding:"required"`
	Source    string `json:"source"`
	IsPrivate bool   `json:"is_private"`
}

func postWriteServiceReq(c *gin.Context, req postWriteReq) service.PostWriteReq {
	uid, role := middleware.CurrentUser(c)
	return service.PostWriteReq{
		OperatorID: uid, OperatorRole: role,
		Title: req.Title, Content: req.Content, Type: req.Type,
		Source: req.Source, IsPrivate: req.IsPrivate,
	}
}

// CreatePost POST /api/post/create
func CreatePost(c *gin.Context) {
	var req postWriteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	id, err := service.CreatePost(postWriteServiceReq(c, req))
	response.Auto(c, gin.H{"id": strconv.FormatInt(id, 10)}, err)
}

type postEditReq struct {
	PostID    int64  `json:"post_id,string" binding:"required"`
	Title     string `json:"title"`
	Content   string `json:"content" binding:"required"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	IsPrivate bool   `json:"is_private"`
}

// EditPost POST /api/post/edit
func EditPost(c *gin.Context) {
	var req postEditReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	svcReq := postWriteServiceReq(c, postWriteReq{
		Title: req.Title, Content: req.Content, Type: req.Type,
		Source: req.Source, IsPrivate: req.IsPrivate,
	})
	response.Auto(c, nil, service.EditPost(svcReq, req.PostID))
}

type postIdReq struct {
	PostID int64 `json:"post_id,string" binding:"required"`
}

// DeletePost POST /api/post/delete
func DeletePost(c *gin.Context) {
	var req postIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	uid, role := middleware.CurrentUser(c)
	response.Auto(c, nil, service.DeletePost(uid, role, req.PostID))
}

// GetPostDetail GET /api/post/detail?id=（游客可看公开帖；浏览量在此计数）
func GetPostDetail(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil || postID <= 0 {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	uid, role := middleware.CurrentUser(c)
	viewerKey := "ip:" + c.ClientIP()
	if uid > 0 {
		viewerKey = "u:" + strconv.FormatInt(uid, 10)
	}
	dto, err := service.GetPostDetail(uid, role, postID, viewerKey)
	response.Auto(c, dto, err)
}

func parseListQuery(c *gin.Context) service.ListQuery {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "5"))
	var userID int64
	if v := c.Query("user_id"); v != "" {
		userID, _ = strconv.ParseInt(v, 10, 64)
	}
	return service.ListQuery{
		Type: c.Query("type"), Sort: c.DefaultQuery("sort", c.DefaultQuery("by", "new")),
		Source: c.Query("source"), UserID: userID, Keyword: c.Query("keyword"),
		Page: page, Count: count,
	}
}

// GetPostList GET /api/post/list?type&sort=hot|new|featured&source&user_id&keyword&page&count
func GetPostList(c *gin.Context) {
	uid, role := middleware.CurrentUser(c)
	q := parseListQuery(c)
	items, total, pageTotal, err := service.GetPostList(uid, role, q)
	response.Auto(c, gin.H{"posts": items, "length": total, "total": total, "page_total": pageTotal}, err)
}

// SearchPosts GET /api/post/search?keyword&page&count（搜索即关键词列表，MySQL 实现）
func SearchPosts(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.OK(c, gin.H{"posts": []interface{}{}, "length": 0, "page_total": 0})
		return
	}
	uid, role := middleware.CurrentUser(c)
	q := parseListQuery(c)
	q.Keyword = keyword
	items, total, pageTotal, err := service.GetPostList(uid, role, q)
	response.Auto(c, gin.H{"posts": items, "length": total, "total": total, "page_total": pageTotal}, err)
}

// GetPostMore GET /api/post/more?type&sort&source&user_id&cursor&count（游标分页）
func GetPostMore(c *gin.Context) {
	uid, role := middleware.CurrentUser(c)
	q := parseListQuery(c)
	if q.Count <= 0 || q.Count > 50 {
		q.Count = 20
	}
	cursor, _ := strconv.ParseInt(c.DefaultQuery("cursor", c.DefaultQuery("before_id", "0")), 10, 64)
	items, err := service.GetPostMoreCursor(uid, role, q, cursor)
	response.Auto(c, gin.H{"posts": items, "length": len(items)}, err)
}

// GetWeekStatus GET /api/post/week-status（当前打卡周期）
func GetWeekStatus(c *gin.Context) {
	response.OK(c, service.GetWeekStatus())
}

// GetUserDiaryWeeks GET /api/post/diary-weeks?user_id=（登录后可查，打卡页状态判断）
func GetUserDiaryWeeks(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	weeks, err := service.GetUserDiaryWeeks(userID)
	response.Auto(c, gin.H{"weeks": weeks, "length": len(weeks)}, err)
}

// TogglePostLike POST /api/post/like
func TogglePostLike(c *gin.Context) {
	var req postIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	uid, role := middleware.CurrentUser(c)
	liked, count, err := service.TogglePostLike(uid, role, req.PostID)
	response.Auto(c, gin.H{"liked": liked, "like_count": count}, err)
}

// GetComments GET /api/post/comments?post_id&father_id&before&after&count
func GetComments(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Query("post_id"), 10, 64)
	if err != nil || postID <= 0 {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	fatherID, _ := strconv.ParseInt(c.Query("father_id"), 10, 64)
	before, _ := strconv.ParseInt(c.Query("before"), 10, 64)
	after, _ := strconv.ParseInt(c.Query("after"), 10, 64)
	count, _ := strconv.ParseInt(c.DefaultQuery("count", "10"), 10, 64)
	uid, role := middleware.CurrentUser(c)
	list, err := service.GetComments(uid, role, postID, fatherID, before, after, count)
	response.Auto(c, gin.H{"comments": list, "length": len(list)}, err)
}

type commentCreateReq struct {
	PostID    int64  `json:"post_id,string" binding:"required"`
	FatherID  int64  `json:"father_id,string"`
	ReplyToID int64  `json:"reply_to_id,string"`
	Content   string `json:"content" binding:"required"`
}

// CreateComment POST /api/post/comment
func CreateComment(c *gin.Context) {
	var req commentCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	uid, role := middleware.CurrentUser(c)
	dto, err := service.CreateComment(uid, role, req.PostID, req.FatherID, req.ReplyToID, req.Content)
	response.Auto(c, dto, err)
}

type commentIdReq struct {
	CommentID int64 `json:"comment_id,string" binding:"required"`
}

// DeleteComment POST /api/post/comment/delete
func DeleteComment(c *gin.Context) {
	var req commentIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	uid, role := middleware.CurrentUser(c)
	response.Auto(c, nil, service.DeleteComment(uid, role, req.CommentID))
}

// ToggleCommentLike POST /api/post/comment/like
func ToggleCommentLike(c *gin.Context) {
	var req commentIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	uid, role := middleware.CurrentUser(c)
	liked, count, err := service.ToggleCommentLike(uid, role, req.CommentID)
	response.Auto(c, gin.H{"liked": liked, "like_count": count}, err)
}

// ---- 管理后台 ----

// AdminPostList GET /api/admin/post/list?keyword&type&status&page&count
func AdminPostList(c *gin.Context) {
	_, role := middleware.CurrentUser(c)
	q := parseListQuery(c)
	if q.Count <= 0 || q.Count > 100 {
		q.Count = 20
	}
	status := c.Query("status")
	items, total, pageTotal, err := service.AdminPostList(role, q, status == "hidden", status == "featured")
	response.Auto(c, gin.H{"posts": items, "length": total, "total": total, "page_total": pageTotal}, err)
}

type setFlagReq struct {
	PostID int64 `json:"post_id,string" binding:"required"`
	Value  bool  `json:"value"`
}

// AdminSetPostHidden POST /api/admin/post/hide
func AdminSetPostHidden(c *gin.Context) {
	var req setFlagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	uid, _ := middleware.CurrentUser(c)
	response.Auto(c, nil, service.SetPostHidden(uid, req.PostID, req.Value))
}

// AdminSetPostFeatured POST /api/admin/post/feature
func AdminSetPostFeatured(c *gin.Context) {
	var req setFlagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	uid, _ := middleware.CurrentUser(c)
	response.Auto(c, nil, service.SetPostFeatured(uid, req.PostID, req.Value))
}

// AdminDeletePost POST /api/admin/post/delete
func AdminDeletePost(c *gin.Context) {
	var req postIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeBadRequest)
		return
	}
	response.Auto(c, nil, service.AdminDeletePost(req.PostID))
}
