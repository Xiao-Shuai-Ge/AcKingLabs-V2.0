package model

import (
	"time"

	"gorm.io/gorm"
)

// 角色等级
const (
	RoleVisitor = 0 // 未实名游客（仅历史数据会出现）
	RoleUser    = 1 // 普通用户（注册即得）
	RoleMember  = 2 // 正式成员
	RoleAdmin   = 3 // 管理员
	RoleSuper   = 4 // 超级管理员
)

// Award 获奖经历（users.awards JSON 数组元素）
type Award struct {
	Name  string `json:"name"`
	Level int    `json:"level"` // 1金 2银 3铜
}

type User struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Username string `gorm:"size:32;not null;default:''"`
	Password string `gorm:"size:255;not null;default:''"` // bcrypt
	Email    string `gorm:"size:255;not null;uniqueIndex"`
	Avatar   string `gorm:"size:512;not null;default:''"`

	Xp        int    `gorm:"not null;default:0"`
	Grade     int    `gorm:"not null;default:0"` // 入学年级，如 23
	StudentNo string `gorm:"size:32;not null;default:''"`
	RealName  string `gorm:"size:32;not null;default:''"`

	CodeforcesID     string `gorm:"size:64;not null;default:''"`
	CodeforcesRating int    `gorm:"not null;default:0"`

	Signature string  `gorm:"size:255;not null;default:''"`
	Awards    []Award `gorm:"type:json;serializer:json"`
	Role      int     `gorm:"not null;default:1"`

	// 用户设置（JSON 分组结构：notify = 通知偏好；后续如隐私/展示等
	// 设置加平行分组即可）。NULL = 未自定义，代码层按默认值处理。
	Settings *UserSettings `gorm:"type:json;serializer:json"`
}

// UserSettings 用户设置（按功能分组）
type UserSettings struct {
	Notify NotifySettings `json:"notify"`
}

// NotifySettings 通知偏好
type NotifySettings struct {
	Like          bool `json:"like"`            // 点赞通知（站内）
	Comment       bool `json:"comment"`         // 评论/回复通知（站内）
	Mention       bool `json:"mention"`         // @提及通知（站内）
	HelpPost      bool `json:"help_post"`       // 新求助帖提醒（站内）
	SystemEmail   bool `json:"system_email"`    // 系统消息同步发送邮件
	NewResumeEmail bool `json:"new_resume_email"` // 新简历邮件提醒（仅管理员生效，非管理员保存时强制 false）
}

// DefaultNotifySettings 默认偏好：站内通知全开，邮件默认关（避免骚扰）
func DefaultNotifySettings() NotifySettings {
	return NotifySettings{
		Like:           true,
		Comment:        true,
		Mention:        true,
		HelpPost:       true,
		SystemEmail:    false,
		NewResumeEmail: false,
	}
}

// NotifyOf 取用户生效的通知偏好（未设置视为默认）
func (u *User) NotifyOf() NotifySettings {
	if u.Settings == nil {
		return DefaultNotifySettings()
	}
	return u.Settings.Notify
}

func (User) TableName() string { return "users" }
