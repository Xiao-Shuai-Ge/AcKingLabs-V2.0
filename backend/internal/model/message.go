package model

import (
	"time"
)

// 站内消息类型
const (
	MsgTypeLike    = "like"
	MsgTypeComment = "comment"
	MsgTypeSystem  = "system"
)

type Message struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time

	UserID   int64  `gorm:"not null;index:idx_user_read,priority:1"` // 收件人
	SenderID int64  `gorm:"not null;default:0"`                      // 发送者（系统消息为 0）
	Type     string `gorm:"size:16;not null"`
	Content  string `gorm:"size:255;not null;default:''"`
	Url      string `gorm:"size:255;not null;default:''"`
	IsRead   bool   `gorm:"not null;default:false;index:idx_user_read,priority:2"`
}

func (Message) TableName() string { return "messages" }
