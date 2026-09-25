package repo

import (
	"gorm.io/gorm"

	"acking/internal/model"
)

func CreatePost(db *gorm.DB, p *model.Post) error {
	return db.Create(p).Error
}

func GetPostByID(db *gorm.DB, id int64) (*model.Post, error) {
	var p model.Post
	if err := db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func SavePost(db *gorm.DB, p *model.Post) error {
	return db.Save(p).Error
}

func UpdatePostColumns(db *gorm.DB, id int64, cols map[string]interface{}) error {
	return db.Model(&model.Post{}).Where("id = ?", id).Updates(cols).Error
}

// PostListFilter 帖子列表筛选
type PostListFilter struct {
	Type string // diary / all(非周记) / 具体类型
	// Sort: hot(hot_score DESC) / new(id DESC) / featured(仅精选, id DESC)
	Sort     string
	Source   string
	WeekCode string // 周记周期筛选（diary 专用）
	UserID   int64
	// IncludePrivate 包含该用户的私密帖（仅当查询自己的帖子或管理员时由业务层置真）
	IncludePrivate bool
	// IncludeHidden 包含被管理员隐藏的帖（仅管理员）
	IncludeHidden bool
	// OnlyHidden / OnlyFeatured 管理后台状态筛选
	OnlyHidden   bool
	OnlyFeatured bool
	Keyword      string // 标题/内容搜索
	Page, Count  int
}

func buildPostQuery(db *gorm.DB, f PostListFilter) *gorm.DB {
	q := db.Model(&model.Post{})
	switch f.Type {
	case "", "all": // 学习页语义：全部 = 非周记
		q = q.Where("type != ?", model.PostTypeDiary)
	case "*": // 管理后台语义：真正的全部类型
	default:
		q = q.Where("type = ?", f.Type)
	}
	if !f.IncludePrivate {
		q = q.Where("is_private = 0")
	}
	if !f.IncludeHidden {
		q = q.Where("is_hidden = 0")
	}
	if f.OnlyHidden {
		q = q.Where("is_hidden = 1")
	}
	if f.OnlyFeatured {
		q = q.Where("is_featured = 1")
	}
	if f.Source != "" {
		q = q.Where("source = ?", f.Source)
	}
	if f.WeekCode != "" {
		q = q.Where("week_code = ?", f.WeekCode)
	}
	if f.UserID > 0 {
		q = q.Where("user_id = ?", f.UserID)
	}
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		q = q.Where("title LIKE ? OR content LIKE ?", kw, kw)
	}
	return q
}

// PostListPage 分页帖子列表（返回总数供计算 page_total）
func PostListPage(db *gorm.DB, f PostListFilter) ([]model.Post, int64, error) {
	q := buildPostQuery(db, f)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	switch f.Sort {
	case "new":
		q = q.Order("id DESC")
	case "featured":
		q = q.Where("is_featured = 1").Order("id DESC")
	default: // hot
		q = q.Order("hot_score DESC, id DESC")
	}
	var posts []model.Post
	err := q.Offset((f.Page - 1) * f.Count).Limit(f.Count).Find(&posts).Error
	return posts, total, err
}

// PostMoreCursor 游标分页（打卡页"加载更多"）
// sort=hot 时 cursor 是 hot_score，new 时 cursor 是 id；返回按排序取的下一页
func PostMoreCursor(db *gorm.DB, f PostListFilter, cursor int64) ([]model.Post, error) {
	q := buildPostQuery(db, f)
	if f.Sort == "hot" {
		q = q.Where("hot_score < ?", cursor).Order("hot_score DESC, id DESC")
	} else {
		q = q.Where("id < ?", cursor).Order("id DESC")
	}
	var posts []model.Post
	return posts, q.Limit(f.Count).Find(&posts).Error
}

// UserDiaryCodes 某用户全部周记的 (id, week_code)
func UserDiaryCodes(db *gorm.DB, userID int64) ([]model.Post, error) {
	var posts []model.Post
	err := db.Select("id", "week_code", "is_private").
		Where("type = ? AND user_id = ?", model.PostTypeDiary, userID).
		Order("id DESC").Find(&posts).Error
	return posts, err
}

// PostLikesExist 单个点赞关系
func PostLikesExist(db *gorm.DB, postID, userID int64) (bool, error) {
	var n int64
	err := db.Model(&model.PostLike{}).Where("post_id = ? AND user_id = ?", postID, userID).Count(&n).Error
	return n > 0, err
}

// PostLikedMap 批量查询点赞关系
func PostLikedMap(db *gorm.DB, userID int64, postIDs []int64) (map[int64]bool, error) {
	result := make(map[int64]bool, len(postIDs))
	if len(postIDs) == 0 || userID <= 0 {
		return result, nil
	}
	var likes []model.PostLike
	if err := db.Where("user_id = ? AND post_id IN ?", userID, postIDs).Find(&likes).Error; err != nil {
		return nil, err
	}
	for _, l := range likes {
		result[l.PostID] = true
	}
	return result, nil
}

func CreatePostLike(db *gorm.DB, l *model.PostLike) error {
	return db.Create(l).Error
}

func DeletePostLike(db *gorm.DB, postID, userID int64) error {
	return db.Where("post_id = ? AND user_id = ?", postID, userID).Delete(&model.PostLike{}).Error
}

func CreateComment(db *gorm.DB, cm *model.Comment) error {
	return db.Create(cm).Error
}

func GetCommentByID(db *gorm.DB, id int64) (*model.Comment, error) {
	var cm model.Comment
	if err := db.First(&cm, id).Error; err != nil {
		return nil, err
	}
	return &cm, nil
}

// TopComments 顶层评论（游标 id < before，新的在前）
func TopComments(db *gorm.DB, postID, before, count int64) ([]model.Comment, error) {
	var list []model.Comment
	err := db.Where("post_id = ? AND father_id = 0 AND id < ?", postID, before).
		Order("id DESC").Limit(int(count)).Find(&list).Error
	return list, err
}

// ChildComments 子评论（游标 id > after，从早到晚）
func ChildComments(db *gorm.DB, fatherID, after, count int64) ([]model.Comment, error) {
	var list []model.Comment
	err := db.Where("father_id = ? AND id > ?", fatherID, after).
		Order("id ASC").Limit(int(count)).Find(&list).Error
	return list, err
}

// CommentsByIDs 批量取评论（回复目标展示用）
func CommentsByIDs(db *gorm.DB, ids []int64) (map[int64]*model.Comment, error) {
	result := make(map[int64]*model.Comment, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var list []model.Comment
	if err := db.Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	for i := range list {
		result[list[i].ID] = &list[i]
	}
	return result, nil
}

func CommentLikedMap(db *gorm.DB, userID int64, commentIDs []int64) (map[int64]bool, error) {
	result := make(map[int64]bool, len(commentIDs))
	if len(commentIDs) == 0 || userID <= 0 {
		return result, nil
	}
	var likes []model.CommentLike
	if err := db.Where("user_id = ? AND comment_id IN ?", userID, commentIDs).Find(&likes).Error; err != nil {
		return nil, err
	}
	for _, l := range likes {
		result[l.CommentID] = true
	}
	return result, nil
}

func CreateCommentLike(db *gorm.DB, l *model.CommentLike) error {
	return db.Create(l).Error
}

func DeleteCommentLike(db *gorm.DB, commentID, userID int64) error {
	return db.Where("comment_id = ? AND user_id = ?", commentID, userID).Delete(&model.CommentLike{}).Error
}

// DeletePostCascadeData 删除帖子关联数据（软删帖与评论，硬删点赞），需在事务中调用
func DeletePostCascadeData(db *gorm.DB, postID int64) error {
	if err := db.Delete(&model.Post{}, postID).Error; err != nil {
		return err
	}
	if err := db.Where("post_id = ?", postID).Delete(&model.PostLike{}).Error; err != nil {
		return err
	}
	if err := db.Where("post_id = ?", postID).Delete(&model.CommentLike{}).Error; err != nil {
		return err
	}
	return db.Where("post_id = ?", postID).Delete(&model.Comment{}).Error
}

// DeleteCommentCascadeData 删除评论及其子评论与点赞（在事务中调用）
func DeleteCommentCascadeData(db *gorm.DB, comment *model.Comment) error {
	// 收集要删的评论：自己 + 所有子孙（两级结构下即全部子评论）
	ids := []int64{comment.ID}
	var children []model.Comment
	if err := db.Where("father_id = ?", comment.ID).Find(&children).Error; err != nil {
		return err
	}
	for _, ch := range children {
		ids = append(ids, ch.ID)
	}
	if err := db.Where("comment_id IN ?", ids).Delete(&model.CommentLike{}).Error; err != nil {
		return err
	}
	if err := db.Where("id IN ?", ids).Delete(&model.Comment{}).Error; err != nil {
		return err
	}
	removed := int64(len(ids))
	return db.Model(&model.Post{}).Where("id = ?", comment.PostID).
		Update("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", removed)).Error
}
