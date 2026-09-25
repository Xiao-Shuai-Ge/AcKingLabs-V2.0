package model

import (
	"time"
)

// Contest 比赛起止时间用毫秒时间戳存储（抓取源直接是时间戳，前端展示也统一毫秒）
type Contest struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Url         string `gorm:"size:512;not null;uniqueIndex"`
	Platform    string `gorm:"size:32;not null;default:'';index"`
	Title       string `gorm:"size:255;not null"`
	StartTime   int64  `gorm:"not null"`
	EndTime     int64  `gorm:"not null"`
	Duration    int64  `gorm:"not null"` // 秒
	IsRecommend bool   `gorm:"not null;default:false"`
	// scrape: 定时抓取；manual: 管理员手动创建（平台为 AcKing）
	Source string `gorm:"size:16;not null;default:scrape"`
}

func (Contest) TableName() string { return "contests" }

// Booking 比赛预约；notified 标记是否已发过开赛提醒（预约状态本身保留）
type Booking struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time

	ContestID int64 `gorm:"not null;uniqueIndex:uk_contest_user,priority:1"`
	UserID    int64 `gorm:"not null;uniqueIndex:uk_contest_user,priority:2"`
	Notified  bool  `gorm:"not null;default:false"`
}

func (Booking) TableName() string { return "bookings" }
