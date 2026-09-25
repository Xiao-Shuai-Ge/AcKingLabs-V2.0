// Package repo 数据访问层：只写 GORM 查询，不做业务判断。
// 所有函数第一个参数传 *gorm.DB，事务里传 tx 即可参与事务。
package repo

import (
	"errors"

	"gorm.io/gorm"

	"acking/internal/model"
)

func CreateUser(db *gorm.DB, u *model.User) error {
	return db.Create(u).Error
}

func GetUserByID(db *gorm.DB, id int64) (*model.User, error) {
	var u model.User
	if err := db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByEmail(db *gorm.DB, email string) (*model.User, error) {
	var u model.User
	err := db.Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUsersByIDs 批量取用户（列表页消灭 N+1 的基础）
func GetUsersByIDs(db *gorm.DB, ids []int64) (map[int64]*model.User, error) {
	result := make(map[int64]*model.User, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var users []model.User
	if err := db.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	for i := range users {
		result[users[i].ID] = &users[i]
	}
	return result, nil
}

func SaveUser(db *gorm.DB, u *model.User) error {
	return db.Save(u).Error
}

func UpdateUserColumns(db *gorm.DB, id int64, cols map[string]interface{}) error {
	return db.Model(&model.User{}).Where("id = ?", id).Updates(cols).Error
}

// UserListFilter 管理后台用户列表筛选
type UserListFilter struct {
	Keyword string
	Role    *int
	Page    int
	Count   int
}

func UserListPage(db *gorm.DB, f UserListFilter) ([]model.User, int64, error) {
	q := db.Model(&model.User{})
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		q = q.Where("username LIKE ? OR email LIKE ? OR real_name LIKE ? OR student_no LIKE ?", kw, kw, kw, kw)
	}
	if f.Role != nil {
		q = q.Where("role = ?", *f.Role)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	err := q.Order("id ASC").Offset((f.Page - 1) * f.Count).Limit(f.Count).Find(&users).Error
	return users, total, err
}

func CountUsers(db *gorm.DB) (int64, error) {
	var n int64
	return n, db.Model(&model.User{}).Count(&n).Error
}

// UsernameExists 用户名是否被其他用户占用
func UsernameExists(db *gorm.DB, username string, excludeID int64) (bool, error) {
	var n int64
	err := db.Model(&model.User{}).Where("username = ? AND id != ?", username, excludeID).Count(&n).Error
	return n > 0, err
}

// UserRankingsPage 经验值排行榜（xp 降序，同分按 id 升序）
func UserRankingsPage(db *gorm.DB, page, count int) ([]model.User, int64, error) {
	q := db.Model(&model.User{}).Where("xp > 0")
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	err := q.Order("xp DESC, id ASC").Offset((page - 1) * count).Limit(count).Find(&users).Error
	return users, total, err
}

// AddUserXp 经验值增减（无负数下限保护，业务层保证合法）
func AddUserXp(db *gorm.DB, userID int64, delta int) error {
	return db.Model(&model.User{}).Where("id = ?", userID).
		Update("xp", gorm.Expr("xp + ?", delta)).Error
}
