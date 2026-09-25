package model

import (
	"time"

	"gorm.io/gorm"
)

// 简历状态（全站统一语义）
const (
	ResumePending   = 0  // 待处理
	ResumeAssessing = 1  // 待考核
	ResumeAccepted  = 2  // 已通过（已发邀请码）
	ResumeRejected  = -1 // 未通过
)

// ResumeExtra 简历补充信息（resumes.extra JSON）
type ResumeExtra struct {
	Information   string `json:"information"`   // 个人介绍
	Skills        string `json:"skills"`        // 专业能力
	Reason        string `json:"reason"`        // 为什么要加入实验室
	Understanding string `json:"understanding"` // 对竞赛的理解
	FuturePlan    string `json:"future_plan"`   // 未来计划
}

type Resume struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Avatar     string      `gorm:"size:512;not null;default:''"`
	RealName   string      `gorm:"size:32;not null;default:''"`
	Grade      int         `gorm:"not null;default:0"`
	StudentNo  string      `gorm:"size:32;not null;default:''"`
	Email      string      `gorm:"size:255;not null;uniqueIndex"`
	Extra      ResumeExtra `gorm:"type:json;serializer:json"`
	InviteCode string      `gorm:"size:32;not null;default:''"` // 通过后生成，注册消耗后清空（一次性）
	Status     int         `gorm:"not null;default:0;index"`
}

func (Resume) TableName() string { return "resumes" }
