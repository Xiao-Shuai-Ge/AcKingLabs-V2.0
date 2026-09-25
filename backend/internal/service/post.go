package service

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"acking/internal/app"
	"acking/internal/model"
	"acking/internal/repo"
	"acking/pkg/response"
	"acking/pkg/weekcode"
)

// 经验值规则
const (
	XpDiaryPublic   = 8  // 公开周记
	XpDiaryPrivate  = 4  // 私密周记
	XpAdminLike     = 5  // 帖子首次被管理员点赞
	XpQualityAnswer = 4  // 求助帖顶层评论首次被管理员点赞（优质解答）
	XpFeatured      = 20 // 帖子首次精选
)

// 内容长度上限（按角色，与前端 utils/contentLimit.ts 保持一致）
func contentLimits(role int) (postLen, commentLen int) {
	switch {
	case role >= model.RoleAdmin:
		return 50000, 5000
	case role == model.RoleMember:
		return 30000, 2000
	default:
		return 15000, 1000
	}
}

// isDuplicateEntry 判断是否唯一索引冲突（不依赖具体驱动类型，按 MySQL 1062 判定）
func isDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}
	return strings.Contains(err.Error(), "Error 1062")
}

// ---- DTO ----

// PostItem 帖子列表项（列表接口内嵌作者与点赞态，消灭前端 N+1）
type PostItem struct {
	ID           string       `json:"id"`
	UserID       string       `json:"user_id"`
	Title        string       `json:"title"`
	ContentShort string       `json:"content_short"`
	Type         string       `json:"type"`
	Source       string       `json:"source"`
	LikeCount    int          `json:"like_count"`
	CommentCount int          `json:"comment_count"`
	ViewCount    int          `json:"view_count"`
	IsAdminLike  bool         `json:"is_admin_like"`
	IsFeatured   bool         `json:"is_featured"`
	IsPrivate    bool         `json:"is_private"`
	IsHidden     bool         `json:"is_hidden,omitempty"`
	WeekCode     string       `json:"week_code,omitempty"`
	HotScore     int64        `json:"hot_score"`
	CreatedAt    int64        `json:"created_at"`
	UpdatedAt    int64        `json:"updated_at"`
	Author       *AuthorBrief `json:"author"`
	Liked        bool         `json:"liked"`
}

func buildPostItems(posts []model.Post, viewerID int64, withHidden bool) []PostItem {
	ids := make([]int64, 0, len(posts))
	userIDs := make([]int64, 0, len(posts))
	for _, p := range posts {
		ids = append(ids, p.ID)
		userIDs = append(userIDs, p.UserID)
	}
	authors := authorMap(userIDs)
	liked, _ := repo.PostLikedMap(app.DB, viewerID, ids)

	items := make([]PostItem, 0, len(posts))
	for _, p := range posts {
		short := ""
		if p.IsPrivate {
			short = "帖子为私密状态，仅作者和管理员可见"
		} else {
			short = contentShort(p.Content, 300)
		}
		item := PostItem{
			ID:           formatID(p.ID),
			UserID:       formatID(p.UserID),
			Title:        p.Title,
			ContentShort: short,
			Type:         p.Type,
			Source:       p.Source,
			LikeCount:    p.LikeCount,
			CommentCount: p.CommentCount,
			ViewCount:    p.ViewCount,
			IsAdminLike:  p.IsAdminLike,
			IsFeatured:   p.IsFeatured,
			IsPrivate:    p.IsPrivate,
			WeekCode:     deref(p.WeekCode),
			HotScore:     p.HotScore,
			CreatedAt:    p.CreatedAt.UnixMilli(),
			UpdatedAt:    p.UpdatedAt.UnixMilli(),
			Author:       toBrief(authors[p.UserID]),
			Liked:        liked[p.ID],
		}
		if withHidden {
			item.IsHidden = p.IsHidden
		}
		items = append(items, item)
	}
	return items
}

// contentShort 截取前 max 个字符（按字符不按字节，避免截断 UTF-8）
func contentShort(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "..."
}

func formatID(id int64) string {
	return strconv.FormatInt(id, 10)
}

// ---- 发布 / 编辑 / 删除 ----

type PostWriteReq struct {
	OperatorID   int64
	OperatorRole int
	Title        string
	Content      string
	Type         string
	Source       string
	IsPrivate    bool
}

// CreatePost 发布帖子；周记的 week_code 与打卡窗口由服务端决定
func CreatePost(req PostWriteReq) (int64, error) {
	postLen, _ := contentLimits(req.OperatorRole)
	if len([]rune(req.Content)) > postLen {
		return 0, response.NewErr(response.CodeContentTooLong)
	}

	post := model.Post{
		UserID:    req.OperatorID,
		Title:     strings.TrimSpace(req.Title),
		Content:   req.Content,
		Type:      req.Type,
		Source:    "",
		IsPrivate: false,
	}

	switch {
	case req.Type == model.PostTypeDiary:
		info := weekcode.NowInfo()
		if !info.Valid {
			return 0, response.NewErr(response.CodeNotDiaryTime)
		}
		code := info.Code
		post.WeekCode = &code
		post.IsPrivate = req.IsPrivate
		if post.Title == "" {
			post.Title = info.Name + " 学习周记"
		}
	case model.IsValidPostType(req.Type):
		if req.Type == model.PostTypeOfficial && req.OperatorRole < model.RoleAdmin {
			return 0, response.NewErr(response.CodeForbidden)
		}
		post.Source = truncateRunes(strings.TrimSpace(req.Source), 255)
		if post.Title == "" {
			return 0, response.NewErrMsg(response.CodeBadRequest, "标题不能为空")
		}
	default:
		return 0, response.NewErrMsg(response.CodeBadRequest, "帖子类型不合法")
	}
	if len([]rune(post.Title)) > 50 {
		return 0, response.NewErrMsg(response.CodeBadRequest, "标题最多 50 字")
	}

	var xpDelta int
	err := app.DB.Transaction(func(tx *gorm.DB) error {
		if err := repo.CreatePost(tx, &post); err != nil {
			if isDuplicateEntry(err) && post.WeekCode != nil {
				return response.NewErr(response.CodeDiaryExists)
			}
			return err
		}
		if post.Type == model.PostTypeDiary {
			xpDelta = XpDiaryPublic
			if post.IsPrivate {
				xpDelta = XpDiaryPrivate
			}
			return repo.AddUserXp(tx, post.UserID, xpDelta)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	// @提及通知 + 求助帖广播（失败不影响发帖）
	notifyMentions(app.DB, req.OperatorID, post.Content, "/"+post.Type+"/"+formatID(post.ID))
	if post.Type == model.PostTypeHelp {
		go BroadcastHelpPost(req.OperatorID, post.Title, "/help/"+formatID(post.ID))
	}
	return post.ID, nil
}

// EditPost 编辑帖子；权限与字段可变性规则：
//   - 作者或管理员可编辑
//   - 周记的 type/week_code 锁定；作者可切换私密
//   - 非管理员不能变更非周记帖子的类型
func EditPost(req PostWriteReq, postID int64) error {
	postLen, _ := contentLimits(req.OperatorRole)
	if len([]rune(req.Content)) > postLen {
		return response.NewErr(response.CodeContentTooLong)
	}

	post, err := repo.GetPostByID(app.DB, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewErr(response.CodeNotFound)
		}
		return err
	}
	if post.UserID != req.OperatorID && req.OperatorRole < model.RoleAdmin {
		return response.NewErr(response.CodeNoPermission)
	}

	title := strings.TrimSpace(req.Title)
	if len([]rune(title)) > 50 {
		return response.NewErrMsg(response.CodeBadRequest, "标题最多 50 字")
	}

	cols := map[string]interface{}{
		"title":   title,
		"content": req.Content,
	}

	if post.Type == model.PostTypeDiary {
		if title == "" {
			title = post.Title
			cols["title"] = title
		}
		if req.OperatorID == post.UserID || req.OperatorRole >= model.RoleAdmin {
			cols["is_private"] = req.IsPrivate
		}
	} else {
		if title == "" {
			return response.NewErrMsg(response.CodeBadRequest, "标题不能为空")
		}
		newType := req.Type
		if newType == "" {
			newType = post.Type
		}
		if !model.IsValidPostType(newType) {
			return response.NewErrMsg(response.CodeBadRequest, "帖子类型不合法")
		}
		if newType != post.Type {
			if req.OperatorRole < model.RoleAdmin {
				return response.NewErrMsg(response.CodeForbidden, "只有管理员可以变更帖子类型")
			}
			if newType == model.PostTypeDiary || post.Type == model.PostTypeDiary {
				return response.NewErrMsg(response.CodeBadRequest, "周记与其他类型不能互转")
			}
			cols["type"] = newType
		}
		cols["source"] = truncateRunes(strings.TrimSpace(req.Source), 255)
	}

	if err := repo.UpdatePostColumns(app.DB, post.ID, cols); err != nil {
		return err
	}
	notifyMentions(app.DB, req.OperatorID, req.Content, "/"+post.Type+"/"+formatID(post.ID))
	return nil
}

// DeletePost 删除帖子（周记为打卡记录，非管理员不可删）；事务级联清理点赞/评论
func DeletePost(operatorID int64, operatorRole int, postID int64) error {
	post, err := repo.GetPostByID(app.DB, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewErr(response.CodeNotFound)
		}
		return err
	}
	if post.UserID != operatorID && operatorRole < model.RoleAdmin {
		return response.NewErr(response.CodeNoPermission)
	}
	if post.Type == model.PostTypeDiary && operatorRole < model.RoleAdmin {
		return response.NewErrMsg(response.CodeForbidden, "周记打卡记录不能删除")
	}
	return app.DB.Transaction(func(tx *gorm.DB) error {
		return repo.DeletePostCascadeData(tx, post.ID)
	})
}

// ---- 详情与列表 ----

// PostDetailDTO 详情
type PostDetailDTO struct {
	PostItem
	Content   string `json:"content"`
	CanEdit   bool   `json:"can_edit"`
	CanDelete bool   `json:"can_delete"`
}

// GetPostDetail 帖子详情（顺带浏览量计数）
func GetPostDetail(viewerID int64, viewerRole int, postID int64, viewerKey string) (*PostDetailDTO, error) {
	post, err := repo.GetPostByID(app.DB, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewErr(response.CodeNotFound)
		}
		return nil, err
	}
	isOwner := viewerID == post.UserID
	isAdmin := viewerRole >= model.RoleAdmin
	if post.IsPrivate && !isOwner && !isAdmin {
		return nil, response.NewErrMsg(response.CodeNotFound, "帖子为私密状态")
	}
	if post.IsHidden && !isAdmin {
		return nil, response.NewErr(response.CodeNotFound)
	}

	// 浏览量：作者自己不计，同一访客 30 分钟内只计一次（登录按用户、游客按 IP）
	if !isOwner {
		if recordView(postID, viewerKey) {
			post.ViewCount++ // 真正计数时才即时回显（Redis 路径实际落库由定时任务完成）
		}
	}

	items := buildPostItems([]model.Post{*post}, viewerID, isAdmin)
	dto := &PostDetailDTO{
		PostItem:  items[0],
		Content:   post.Content,
		CanEdit:   isOwner || isAdmin,
		CanDelete: (post.Type != model.PostTypeDiary && isOwner) || isAdmin,
	}
	return dto, nil
}

// ListQuery 列表查询参数
type ListQuery struct {
	Type        string
	Sort        string // hot / new / featured
	Source      string
	UserID      int64
	Keyword     string
	Page, Count int
}

// GetPostList 分页列表（学习页/个人主页/管理后台共用）
func GetPostList(viewerID int64, viewerRole int, q ListQuery) (items []PostItem, total, pageTotal int64, err error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Count <= 0 || q.Count > 50 {
		q.Count = 5
	}
	f := repo.PostListFilter{
		Type: q.Type, Sort: q.Sort, Source: diarySource(q), WeekCode: diaryWeek(q), UserID: q.UserID,
		Keyword: q.Keyword, Page: q.Page, Count: q.Count,
		IncludeHidden:  viewerRole >= model.RoleAdmin,
		IncludePrivate: q.UserID > 0 && (q.UserID == viewerID || viewerRole >= model.RoleAdmin),
	}
	posts, total, err := repo.PostListPage(app.DB, f)
	if err != nil {
		return
	}
	items = buildPostItems(posts, viewerID, viewerRole >= model.RoleAdmin)
	pageTotal = (total + int64(q.Count) - 1) / int64(q.Count)
	return
}

// GetPostMoreCursor 游标列表（打卡页加载更多）
func GetPostMoreCursor(viewerID int64, viewerRole int, q ListQuery, cursor int64) ([]PostItem, error) {
	if q.Count <= 0 || q.Count > 50 {
		q.Count = 20
	}
	if cursor <= 0 {
		cursor = 1 << 62
	}
	f := repo.PostListFilter{
		Type: q.Type, Sort: q.Sort, Source: diarySource(q), WeekCode: diaryWeek(q), UserID: q.UserID, Count: q.Count,
		IncludeHidden:  viewerRole >= model.RoleAdmin,
		IncludePrivate: q.UserID > 0 && (q.UserID == viewerID || viewerRole >= model.RoleAdmin),
	}
	posts, err := repo.PostMoreCursor(app.DB, f, cursor)
	if err != nil {
		return nil, err
	}
	return buildPostItems(posts, viewerID, viewerRole >= model.RoleAdmin), nil
}

// DiaryWeekItem 打卡周记录
type DiaryWeekItem struct {
	PostID    string `json:"post_id"`
	WeekCode  string `json:"week_code"`
	IsPrivate bool   `json:"is_private"`
}

// GetUserDiaryWeeks 用户的周记周期列表（打卡页状态判断）
func GetUserDiaryWeeks(userID int64) ([]DiaryWeekItem, error) {
	posts, err := repo.UserDiaryCodes(app.DB, userID)
	if err != nil {
		return nil, err
	}
	out := make([]DiaryWeekItem, 0, len(posts))
	for _, p := range posts {
		out = append(out, DiaryWeekItem{
			PostID:    formatID(p.ID),
			WeekCode:  deref(p.WeekCode),
			IsPrivate: p.IsPrivate,
		})
	}
	return out, nil
}

// GetWeekStatus 当前打卡周期信息（给前端展示按钮状态）
func GetWeekStatus() map[string]interface{} {
	info := weekcode.NowInfo()
	from, to := weekcode.ValidWindow(time.Now())
	studyFrom, studyTo := weekcode.StudyWindow(time.Now())
	return map[string]interface{}{
		"code":       info.Code,
		"name":       info.Name,
		"valid":      info.Valid,
		"from":       from.UnixMilli(),
		"to":         to.UnixMilli(),
		"study_from": studyFrom.UnixMilli(),
		"study_to":   studyTo.UnixMilli(),
	}
}

// ---- 点赞 ----

// TogglePostLike 帖子点赞切换；返回点赞后的状态
func TogglePostLike(operatorID int64, operatorRole int, postID int64) (liked bool, likeCount int, err error) {
	post, err := repo.GetPostByID(app.DB, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, 0, response.NewErr(response.CodeNotFound)
		}
		return false, 0, err
	}
	if post.IsPrivate || post.IsHidden {
		if !(post.UserID == operatorID || operatorRole >= model.RoleAdmin) {
			return false, 0, response.NewErr(response.CodeNotFound)
		}
	}

	isAdmin := operatorRole >= model.RoleAdmin
	err = app.DB.Transaction(func(tx *gorm.DB) error {
		exists, err := repo.PostLikesExist(tx, postID, operatorID)
		if err != nil {
			return err
		}
		if exists {
			if err := repo.DeletePostLike(tx, postID, operatorID); err != nil {
				return err
			}
			if err := tx.Model(&model.Post{}).Where("id = ?", postID).
				Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error; err != nil {
				return err
			}
			liked = false
			return nil
		}
		if err := repo.CreatePostLike(tx, &model.PostLike{PostID: postID, UserID: operatorID}); err != nil {
			if isDuplicateEntry(err) {
				liked = true
				return nil // 并发点击，视为已点赞
			}
			return err
		}
		if err := tx.Model(&model.Post{}).Where("id = ?", postID).
			Update("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
			return err
		}
		liked = true
		// 管理员首次点赞：打标 + 作者加经验 + 通知
		if isAdmin && !post.IsAdminLike {
			if err := tx.Model(&model.Post{}).Where("id = ?", postID).
				Update("is_admin_like", true).Error; err != nil {
				return err
			}
			if err := repo.AddUserXp(tx, post.UserID, XpAdminLike); err != nil {
				return err
			}
		}
		notify(tx, post.UserID, operatorID, "like", "赞了你的帖子《"+truncateRunes(post.Title, 20)+"》",
			"/"+post.Type+"/"+formatID(post.ID))
		return nil
	})
	if err != nil {
		return false, 0, err
	}
	// 重新读计数（事务内 GREATEST 表达式无法直接回读）
	fresh, err := repo.GetPostByID(app.DB, postID)
	if err == nil {
		likeCount = fresh.LikeCount
	}
	return liked, likeCount, nil
}

// ---- 评论 ----

// CommentItem 评论 DTO
type CommentItem struct {
	ID          string       `json:"id"`
	PostID      string       `json:"post_id"`
	FatherID    string       `json:"father_id"`
	ReplyToID   string       `json:"reply_to_id"`
	UserID      string       `json:"user_id"`
	Content     string       `json:"content"`
	LikeCount   int          `json:"like_count"`
	IsAdminLike bool         `json:"is_admin_like"`
	CreatedAt   int64        `json:"created_at"`
	Author      *AuthorBrief `json:"author"`
	ReplyTo     string       `json:"reply_to,omitempty"` // 被回复人昵称
	Liked       bool         `json:"liked"`
	CanDelete   bool         `json:"can_delete"`
}

func buildCommentItems(list []model.Comment, viewerID int64, viewerRole int) []CommentItem {
	ids := make([]int64, 0, len(list))
	userIDs := make([]int64, 0, len(list))
	replyIDs := make([]int64, 0)
	for _, cm := range list {
		ids = append(ids, cm.ID)
		userIDs = append(userIDs, cm.UserID)
		if cm.ReplyToID > 0 {
			replyIDs = append(replyIDs, cm.ReplyToID)
		}
	}
	authors := authorMap(userIDs)
	liked, _ := repo.CommentLikedMap(app.DB, viewerID, ids)

	// 被回复评论 -> 被回复人昵称（可能不在本页，单独批量查）
	replyComments, _ := repo.CommentsByIDs(app.DB, replyIDs)
	replyUserIDs := make([]int64, 0, len(replyComments))
	for _, rc := range replyComments {
		replyUserIDs = append(replyUserIDs, rc.UserID)
	}
	replyAuthors := authorMap(replyUserIDs)
	replyNames := make(map[int64]string, len(replyComments))
	for cid, rc := range replyComments {
		if u := replyAuthors[rc.UserID]; u != nil {
			replyNames[cid] = u.Username
		}
	}

	items := make([]CommentItem, 0, len(list))
	for _, cm := range list {
		items = append(items, CommentItem{
			ID:          formatID(cm.ID),
			PostID:      formatID(cm.PostID),
			FatherID:    formatID(cm.FatherID),
			ReplyToID:   formatID(cm.ReplyToID),
			UserID:      formatID(cm.UserID),
			Content:     cm.Content,
			LikeCount:   cm.LikeCount,
			IsAdminLike: cm.IsAdminLike,
			CreatedAt:   cm.CreatedAt.UnixMilli(),
			Author:      toBrief(authors[cm.UserID]),
			ReplyTo:     replyNames[cm.ReplyToID],
			Liked:       liked[cm.ID],
			CanDelete:   cm.UserID == viewerID || viewerRole >= model.RoleAdmin,
		})
	}
	return items
}

// CommentNodeItem 顶层评论 + 内嵌子评论（列表页默认展开，免逐条请求）
type CommentNodeItem struct {
	CommentItem
	Children     []CommentItem `json:"children"`
	ChildrenMore bool          `json:"children_more"`
}

// CreateComment 发评论；两级结构：father_id 一定是顶层评论 id
func CreateComment(operatorID int64, operatorRole int, postID, fatherID, replyToID int64, content string) (*CommentItem, error) {
	_, commentLen := contentLimits(operatorRole)
	if n := len([]rune(content)); n == 0 || n > commentLen {
		return nil, response.NewErr(response.CodeContentTooLong)
	}
	post, err := repo.GetPostByID(app.DB, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewErr(response.CodeNotFound)
		}
		return nil, err
	}
	if post.IsHidden {
		return nil, response.NewErrMsg(response.CodeBadRequest, "帖子已被隐藏，无法评论")
	}
	if post.IsPrivate && post.UserID != operatorID && operatorRole < model.RoleAdmin {
		return nil, response.NewErr(response.CodeNoPermission)
	}

	cm := model.Comment{PostID: postID, FatherID: 0, ReplyToID: 0, UserID: operatorID, Content: content}

	// 归一化两级结构：回复子评论时，father 提升为其顶层评论，reply_to 指向被回复评论
	if fatherID > 0 {
		father, err := repo.GetCommentByID(app.DB, fatherID)
		if err != nil || father.PostID != postID || father.DeletedAt.Valid {
			return nil, response.NewErrMsg(response.CodeBadRequest, "回复的评论不存在")
		}
		cm.FatherID = fatherID
		if father.FatherID > 0 {
			cm.FatherID = father.FatherID
			cm.ReplyToID = father.ID
		}
	}
	if replyToID > 0 {
		rt, err := repo.GetCommentByID(app.DB, replyToID)
		if err != nil || rt.PostID != postID || rt.DeletedAt.Valid {
			return nil, response.NewErrMsg(response.CodeBadRequest, "回复的评论不存在")
		}
		if cm.FatherID == 0 {
			cm.FatherID = rt.ID
		}
		cm.ReplyToID = rt.ID
		if rt.FatherID > 0 {
			cm.FatherID = rt.FatherID
		}
	}

	err = app.DB.Transaction(func(tx *gorm.DB) error {
		if err := repo.CreateComment(tx, &cm); err != nil {
			return err
		}
		return tx.Model(&model.Post{}).Where("id = ?", postID).
			Update("comment_count", gorm.Expr("comment_count + 1")).Error
	})
	if err != nil {
		return nil, err
	}

	postURL := "/" + post.Type + "/" + formatID(post.ID)
	// 通知：顶层评论 -> 帖子作者；子评论 -> 被回复人 + 帖子作者
	if cm.FatherID == 0 {
		notify(app.DB, post.UserID, operatorID, "comment", "评论了你的帖子《"+truncateRunes(post.Title, 20)+"》", postURL)
	} else {
		if cm.ReplyToID > 0 {
			if rt, err := repo.GetCommentByID(app.DB, cm.ReplyToID); err == nil {
				notify(app.DB, rt.UserID, operatorID, "comment", "回复了你的评论", postURL)
			}
		}
		if father, err := repo.GetCommentByID(app.DB, cm.FatherID); err == nil && father.UserID != post.UserID {
			notify(app.DB, father.UserID, operatorID, "comment", "回复了你的评论", postURL)
		}
	}
	notifyMentions(app.DB, operatorID, content, postURL)

	items := buildCommentItems([]model.Comment{cm}, operatorID, operatorRole)
	return &items[0], nil
}

// GetComments 评论列表：father_id=0 顶层（id < before 倒序，内嵌子评论默认展开）；
// 否则子评论（id > after 正序）。
func GetComments(viewerID int64, viewerRole int, postID, fatherID, before, after, count int64) ([]CommentNodeItem, error) {
	post, err := repo.GetPostByID(app.DB, postID)
	if err != nil {
		return nil, response.NewErr(response.CodeNotFound)
	}
	if post.IsPrivate && post.UserID != viewerID && viewerRole < model.RoleAdmin {
		return nil, response.NewErr(response.CodeNoPermission)
	}
	if post.IsHidden && viewerRole < model.RoleAdmin {
		return nil, response.NewErr(response.CodeNotFound)
	}
	if count <= 0 || count > 50 {
		count = 10
	}
	var list []model.Comment
	if fatherID <= 0 {
		if before <= 0 {
			before = 1 << 62
		}
		list, err = repo.TopComments(app.DB, postID, before, count)
		if err != nil {
			return nil, err
		}
		// 一并取回这些顶层评论的子评论（单条 IN 查询，消灭前端逐楼请求）
		ids := make([]int64, 0, len(list))
		for _, c := range list {
			ids = append(ids, c.ID)
		}
		var childModels []model.Comment
		if len(ids) > 0 {
			if err := app.DB.Where("father_id IN ?", ids).Order("id ASC").Find(&childModels).Error; err != nil {
				return nil, err
			}
		}
		childMap := make(map[int64][]model.Comment, len(list))
		for _, c := range childModels {
			childMap[c.FatherID] = append(childMap[c.FatherID], c)
		}
		out := make([]CommentNodeItem, 0, len(list))
		for _, top := range list {
			children := childMap[top.ID]
			more := false
			if len(children) > 50 { // 超出部分走子评论游标接口
				children = children[:50]
				more = true
			}
			items := buildCommentItems([]model.Comment{top}, viewerID, viewerRole)
			childItems := buildCommentItems(children, viewerID, viewerRole)
			if childItems == nil {
				childItems = []CommentItem{}
			}
			out = append(out, CommentNodeItem{CommentItem: items[0], Children: childItems, ChildrenMore: more})
		}
		return out, nil
	}
	list, err = repo.ChildComments(app.DB, fatherID, after, count)
	if err != nil {
		return nil, err
	}
	items := buildCommentItems(list, viewerID, viewerRole)
	out := make([]CommentNodeItem, 0, len(items))
	for _, it := range items {
		out = append(out, CommentNodeItem{CommentItem: it, Children: []CommentItem{}})
	}
	return out, nil
}

// DeleteComment 删除评论（作者或管理员；事务级联子评论与点赞）
func DeleteComment(operatorID int64, operatorRole int, commentID int64) error {
	cm, err := repo.GetCommentByID(app.DB, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewErr(response.CodeNotFound)
		}
		return err
	}
	if cm.UserID != operatorID && operatorRole < model.RoleAdmin {
		return response.NewErr(response.CodeNoPermission)
	}
	return app.DB.Transaction(func(tx *gorm.DB) error {
		return repo.DeleteCommentCascadeData(tx, cm)
	})
}

// ToggleCommentLike 评论点赞切换
func ToggleCommentLike(operatorID int64, operatorRole int, commentID int64) (liked bool, likeCount int, err error) {
	cm, err := repo.GetCommentByID(app.DB, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, 0, response.NewErr(response.CodeNotFound)
		}
		return false, 0, err
	}
	post, err := repo.GetPostByID(app.DB, cm.PostID)
	if err != nil {
		return false, 0, response.NewErr(response.CodeNotFound)
	}
	postURL := "/" + post.Type + "/" + formatID(post.ID)
	isAdmin := operatorRole >= model.RoleAdmin
	err = app.DB.Transaction(func(tx *gorm.DB) error {
		var n int64
		if err := tx.Model(&model.CommentLike{}).Where("comment_id = ? AND user_id = ?", commentID, operatorID).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			if err := repo.DeleteCommentLike(tx, commentID, operatorID); err != nil {
				return err
			}
			if err := tx.Model(&model.Comment{}).Where("id = ?", commentID).
				Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error; err != nil {
				return err
			}
			liked = false
			return nil
		}
		if err := repo.CreateCommentLike(tx, &model.CommentLike{CommentID: commentID, UserID: operatorID}); err != nil {
			if isDuplicateEntry(err) {
				liked = true
				return nil
			}
			return err
		}
		if err := tx.Model(&model.Comment{}).Where("id = ?", commentID).
			Update("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
			return err
		}
		liked = true
		// 求助帖顶层评论首次被管理员点赞 = 优质解答：打标 + 评论者加经验 + 通知
		if isAdmin && !cm.IsAdminLike {
			if err := tx.Model(&model.Comment{}).Where("id = ?", commentID).
				Update("is_admin_like", true).Error; err != nil {
				return err
			}
			if post.Type == model.PostTypeHelp && cm.FatherID == 0 {
				if err := repo.AddUserXp(tx, cm.UserID, XpQualityAnswer); err != nil {
					return err
				}
			}
			notify(tx, cm.UserID, operatorID, "like", "管理员推荐了你的评论", postURL)
		} else {
			notify(tx, cm.UserID, operatorID, "like", "赞了你的评论", postURL)
		}
		return nil
	})
	if err != nil {
		return false, 0, err
	}
	if fresh, err := repo.GetCommentByID(app.DB, commentID); err == nil {
		likeCount = fresh.LikeCount
	}
	return liked, likeCount, nil
}

// ---- 管理后台 ----

// AdminPostList 管理后台帖子列表（关键词/类型/状态筛选）
func AdminPostList(viewerRole int, q ListQuery, onlyHidden, onlyFeatured bool) (items []PostItem, total, pageTotal int64, err error) {
	if viewerRole < model.RoleAdmin {
		return nil, 0, 0, response.NewErr(response.CodeForbidden)
	}
	if q.Type == "" || q.Type == "all" {
		q.Type = "*" // 管理后台的"全部"包含周记
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Count <= 0 || q.Count > 100 {
		q.Count = 20
	}
	f := repo.PostListFilter{
		Type: q.Type, Sort: "new", Source: q.Source, UserID: q.UserID, Keyword: q.Keyword,
		Page: q.Page, Count: q.Count,
		IncludeHidden: true, IncludePrivate: true,
		OnlyHidden: onlyHidden, OnlyFeatured: onlyFeatured,
	}
	posts, total, err := repo.PostListPage(app.DB, f)
	if err != nil {
		return
	}
	items = buildPostItems(posts, 0, true)
	pageTotal = (total + int64(q.Count) - 1) / int64(q.Count)
	return
}

// SetPostHidden 隐藏/恢复帖子
func SetPostHidden(operatorID int64, postID int64, hidden bool) error {
	post, err := repo.GetPostByID(app.DB, postID)
	if err != nil {
		return response.NewErr(response.CodeNotFound)
	}
	if post.IsHidden == hidden {
		return nil
	}
	if err := repo.UpdatePostColumns(app.DB, postID, map[string]interface{}{"is_hidden": hidden}); err != nil {
		return err
	}
	url := "/" + post.Type + "/" + formatID(postID)
	if hidden {
		notify(app.DB, post.UserID, operatorID, "system", "你的帖子《"+truncateRunes(post.Title, 20)+"》已被管理员隐藏", url)
	} else {
		notify(app.DB, post.UserID, operatorID, "system", "你的帖子《"+truncateRunes(post.Title, 20)+"》已恢复展示", url)
	}
	return nil
}

// SetPostFeatured 精选/取消精选（首次精选加经验并通知）
func SetPostFeatured(operatorID int64, postID int64, featured bool) error {
	post, err := repo.GetPostByID(app.DB, postID)
	if err != nil {
		return response.NewErr(response.CodeNotFound)
	}
	if post.IsFeatured == featured {
		return nil
	}
	err = app.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Post{}).Where("id = ?", postID).
			Update("is_featured", featured).Error; err != nil {
			return err
		}
		if featured && !post.IsFeatured {
			if err := repo.AddUserXp(tx, post.UserID, XpFeatured); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if featured {
		notify(app.DB, post.UserID, operatorID, "system",
			"你的帖子《"+truncateRunes(post.Title, 20)+"》被设为精选", "/"+post.Type+"/"+formatID(postID))
	}
	return nil
}

// AdminDeletePost 管理员删帖
func AdminDeletePost(postID int64) error {
	if _, err := repo.GetPostByID(app.DB, postID); err != nil {
		return response.NewErr(response.CodeNotFound)
	}
	return app.DB.Transaction(func(tx *gorm.DB) error {
		return repo.DeletePostCascadeData(tx, postID)
	})
}

// ---- 浏览量 ----

// viewDedupWindow 同一访客对同一帖子的去重窗口
const viewDedupWindow = 30 * time.Minute

// recordView 浏览计数，返回是否真正计入：
//   - Redis 可用：SETNX 去重（30 分钟）+ Hash 累积增量，由定时任务批量落库；
//   - Redis 不可用：进程内 TTL 表去重（30 分钟）+ 直接落库（单机部署同样准确，
//     修复降级路径"每次点击都 +1"的问题）。
func recordView(postID int64, viewerKey string) bool {
	if app.RedisOK() {
		ctx := context.Background()
		dedup := "view:" + formatID(postID) + ":" + viewerKey
		ok, err := app.RDB.SetNX(ctx, dedup, 1, viewDedupWindow).Result()
		if err == nil {
			if !ok {
				return false // 窗口内已计过
			}
			app.RDB.HIncrBy(ctx, "views:delta", formatID(postID), 1)
			return true
		}
		// Redis 中途故障则退回进程内方案，不让这次浏览丢失
	}
	if memViewSet(postID, viewerKey) {
		if err := app.DB.Model(&model.Post{}).Where("id = ?", postID).
			Update("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
			slog.Error("浏览量落库失败", "post", postID, "err", err)
			return false
		}
		return true
	}
	return false
}

// ---- 进程内浏览去重表（Redis 不可用时的降级实现） ----

var (
	viewMu   sync.Mutex
	viewSeen = map[string]time.Time{}
)

// memViewSet 未在窗口内出现过则登记并返回 true
func memViewSet(postID int64, viewerKey string) bool {
	key := formatID(postID) + ":" + viewerKey
	viewMu.Lock()
	defer viewMu.Unlock()
	if t, ok := viewSeen[key]; ok && time.Since(t) < viewDedupWindow {
		return false
	}
	viewSeen[key] = time.Now()
	return true
}

func init() {
	// 定期清理过期去重记录，避免无限增长
	go func() {
		for range time.Tick(10 * time.Minute) {
			viewMu.Lock()
			for k, t := range viewSeen {
				if time.Since(t) >= viewDedupWindow {
					delete(viewSeen, k)
				}
			}
			viewMu.Unlock()
		}
	}()
}

// FlushViews 把 Redis 中的浏览增量原子地搬回数据库（Lua 保证不丢不重）
func FlushViews() {
	if !app.RedisOK() {
		return
	}
	script := `
local d = redis.call('HGETALL', KEYS[1])
if next(d) then redis.call('DEL', KEYS[1]) end
return d
`
	res, err := app.RDB.Eval(context.Background(), script, []string{"views:delta"}).Result()
	if err != nil {
		slog.Error("读取浏览增量失败", "err", err)
		return
	}
	fields, ok := res.([]interface{})
	if !ok || len(fields) == 0 {
		return
	}
	for i := 0; i+1 < len(fields); i += 2 {
		idStr, _ := fields[i].(string)
		deltaStr, _ := fields[i+1].(string)
		id, err1 := strconv.ParseInt(idStr, 10, 64)
		delta, err2 := strconv.ParseInt(deltaStr, 10, 64)
		if err1 != nil || err2 != nil || delta <= 0 {
			continue
		}
		if err := app.DB.Model(&model.Post{}).Where("id = ?", id).
			Update("view_count", gorm.Expr("view_count + ?", delta)).Error; err != nil {
			slog.Error("浏览量落库失败", "post", id, "err", err)
		}
	}
}

// diaryWeek: 查询周记时，source 参数的语义是周期编码（week_code 列）
func diaryWeek(q ListQuery) string {
	if q.Type == model.PostTypeDiary && q.Source != "" {
		return q.Source
	}
	return ""
}

// diarySource: 周记的 source 列恒为空
func diarySource(q ListQuery) string {
	if q.Type == model.PostTypeDiary {
		return ""
	}
	return q.Source
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}
