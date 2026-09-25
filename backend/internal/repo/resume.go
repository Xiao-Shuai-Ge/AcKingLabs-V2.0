package repo

import (
	"gorm.io/gorm"

	"acking/internal/model"
)

func CreateResume(db *gorm.DB, r *model.Resume) error {
	return db.Create(r).Error
}

func SaveResume(db *gorm.DB, r *model.Resume) error {
	return db.Save(r).Error
}

func GetResumeByID(db *gorm.DB, id int64) (*model.Resume, error) {
	var r model.Resume
	if err := db.First(&r, id).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func GetResumeByEmail(db *gorm.DB, email string) (*model.Resume, error) {
	var r model.Resume
	err := db.Where("email = ?", email).First(&r).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// ResumeListPage 管理后台列表：keyword 匹配姓名/学号/邮箱，status 为 -100 时表示全部
func ResumeListPage(db *gorm.DB, keyword string, status, page, count int) ([]model.Resume, int64, error) {
	q := db.Model(&model.Resume{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("real_name LIKE ? OR student_no LIKE ? OR email LIKE ?", kw, kw, kw)
	}
	if status != -100 {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Resume
	err := q.Order("id DESC").Offset((page - 1) * fCount(count)).Limit(fCount(count)).Find(&list).Error
	return list, total, err
}

// ConsumeResumeInviteCode 消耗邀请码（一次性）：仅当未被消耗时置空并返回是否成功
func ConsumeResumeInviteCode(db *gorm.DB, id int64) error {
	return db.Model(&model.Resume{}).Where("id = ?", id).
		Update("invite_code", "").Error
}
