package model

import (
	"time"

	"gorm.io/gorm"
)

// 帖子类型
const (
	PostTypeDiary    = "diary"    // 周记打卡
	PostTypeTutorial = "tutorial" // 教程
	PostTypeSolution = "solution" // 题解
	PostTypeContest  = "contest"  // 比赛复盘
	PostTypeFun      = "fun"      // 闲聊
	PostTypeHelp     = "help"     // 求助
	PostTypeOfficial = "official" // 官方
)

var PostTypes = []string{
	PostTypeDiary, PostTypeTutorial, PostTypeSolution,
	PostTypeContest, PostTypeFun, PostTypeHelp, PostTypeOfficial,
}

// IsValidPostType 类型是否合法
func IsValidPostType(t string) bool {
	for _, v := range PostTypes {
		if v == t {
			return true
		}
	}
	return false
}

type Post struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	UserID  int64  `gorm:"not null;index:idx_user_type,priority:1;uniqueIndex:uk_user_week,priority:1"`
	Title   string `gorm:"size:100;not null"`
	Content string `gorm:"type:mediumtext;not null"`
	Type    string `gorm:"size:20;not null;index:idx_user_type,priority:2;index:idx_type_hot,priority:1"`
	// diary: 空；solution: 比赛链接；其他: 自定义来源/网址
	Source string `gorm:"size:255;not null;default:'';index"`
	// 周记周期编码（仅 diary 有值，其余 NULL——唯一索引忽略 NULL，避免非周记帖互相冲突）
	// 与 user_id 组成唯一索引 => 数据库层防重复打卡
	WeekCode *string `gorm:"size:20;uniqueIndex:uk_user_week,priority:2"`

	LikeCount    int `gorm:"not null;default:0"`
	CommentCount int `gorm:"not null;default:0"`
	ViewCount    int `gorm:"not null;default:0"`

	IsAdminLike bool `gorm:"not null;default:false"` // 管理员点赞过（前端"管理推荐"角标）
	IsFeatured  bool `gorm:"not null;default:false"` // 精选
	IsPrivate   bool `gorm:"not null;default:false"` // 私密（仅本人与管理员可见）
	IsHidden    bool `gorm:"not null;default:false"` // 管理员隐藏（原 AI 审核下架的替代）

	// 热度分 = 基准时间 + 点赞*1h + 评论*1h + 管理员点赞*3d + 精选*14d（官方帖基准用 updated_at）
	// 仅由定时任务重算，写路径不维护，避免点赞/评论时的写放大。
	HotScore int64 `gorm:"not null;default:0;index:idx_type_hot,priority:2"`
}

func (Post) TableName() string { return "posts" }

type PostLike struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time

	PostID int64 `gorm:"not null;uniqueIndex:uk_post_user,priority:1"`
	UserID int64 `gorm:"not null;uniqueIndex:uk_post_user,priority:2;index"`
}

func (PostLike) TableName() string { return "post_likes" }

// Comment 两级评论：father_id=0 为顶层评论；子评论的 father_id 指向其顶层评论，
// reply_to_id 记录"回复了谁"（仅子评论之间回复时有值）。
type Comment struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	PostID    int64  `gorm:"not null;index:idx_post_id,priority:1"`
	FatherID  int64  `gorm:"not null;default:0;index:idx_post_id,priority:2"`
	ReplyToID int64  `gorm:"not null;default:0"`
	UserID    int64  `gorm:"not null;index"`
	Content   string `gorm:"type:text;not null"`

	LikeCount   int  `gorm:"not null;default:0"`
	IsAdminLike bool `gorm:"not null;default:false"`
}

func (Comment) TableName() string { return "comments" }

type CommentLike struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time

	CommentID int64 `gorm:"not null;uniqueIndex:uk_comment_user,priority:1"`
	UserID    int64 `gorm:"not null;uniqueIndex:uk_comment_user,priority:2;index"`
}

func (CommentLike) TableName() string { return "comment_likes" }
