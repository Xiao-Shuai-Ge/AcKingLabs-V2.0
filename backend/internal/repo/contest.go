package repo

import (
	"gorm.io/gorm"

	"acking/internal/model"
)

func UpsertContest(db *gorm.DB, c *model.Contest) error {
	// 按 url 幂等：存在则更新比赛信息，不存在则插入
	var exist model.Contest
	err := db.Where("url = ?", c.Url).First(&exist).Error
	if err == gorm.ErrRecordNotFound {
		return db.Create(c).Error
	}
	if err != nil {
		return err
	}
	return db.Model(&exist).Updates(map[string]interface{}{
		"title":      c.Title,
		"start_time": c.StartTime,
		"end_time":   c.EndTime,
		"duration":   c.Duration,
		"platform":   c.Platform,
	}).Error
}

func GetContestByID(db *gorm.DB, id int64) (*model.Contest, error) {
	var c model.Contest
	if err := db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func GetContestURLExists(db *gorm.DB, url string, excludeID int64) (bool, error) {
	var n int64
	err := db.Model(&model.Contest{}).Where("url = ? AND id != ?", url, excludeID).Count(&n).Error
	return n > 0, err
}

// ContestListPage platform: all / recommend / 具体平台名；按开始时间倒序
func ContestListPage(db *gorm.DB, platform string, page, count int) ([]model.Contest, int64, error) {
	q := db.Model(&model.Contest{})
	switch platform {
	case "", "all":
	case "recommend":
		q = q.Where("is_recommend = 1")
	default:
		q = q.Where("platform = ?", platform)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Contest
	err := q.Order("start_time DESC").Offset((page - 1) * fCount(count)).Limit(fCount(count)).Find(&list).Error
	return list, total, err
}

func fCount(c int) int {
	if c <= 0 || c > 100 {
		return 10
	}
	return c
}

func SaveContest(db *gorm.DB, c *model.Contest) error {
	return db.Save(c).Error
}

func GetContestURL(db *gorm.DB, url string) (*model.Contest, error) {
	var c model.Contest
	if err := db.Where("url = ?", url).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func UpdateContestRecommend(db *gorm.DB, id int64, recommend bool) error {
	return db.Model(&model.Contest{}).Where("id = ?", id).
		Update("is_recommend", recommend).Error
}

func DeleteContestCascade(db *gorm.DB, id int64) error {
	if err := db.Where("contest_id = ?", id).Delete(&model.Booking{}).Error; err != nil {
		return err
	}
	return db.Delete(&model.Contest{}, id).Error
}

func IsBooked(db *gorm.DB, contestID, userID int64) (bool, error) {
	var n int64
	err := db.Model(&model.Booking{}).Where("contest_id = ? AND user_id = ?", contestID, userID).Count(&n).Error
	return n > 0, err
}

// BookedMap 批量查询预约状态
func BookedMap(db *gorm.DB, userID int64, contestIDs []int64) (map[int64]bool, error) {
	result := make(map[int64]bool, len(contestIDs))
	if len(contestIDs) == 0 || userID <= 0 {
		return result, nil
	}
	var list []model.Booking
	if err := db.Where("user_id = ? AND contest_id IN ?", userID, contestIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, b := range list {
		result[b.ContestID] = true
	}
	return result, nil
}

func CreateBooking(db *gorm.DB, b *model.Booking) error {
	return db.Create(b).Error
}

func DeleteBooking(db *gorm.DB, contestID, userID int64) error {
	return db.Where("contest_id = ? AND user_id = ?", contestID, userID).Delete(&model.Booking{}).Error
}

// PendingNotifyBookings 未通知且比赛即将开始（20 分钟内）或刚开始的预约
func PendingNotifyBookings(db *gorm.DB, nowMs int64) ([]model.Booking, error) {
	var list []model.Booking
	err := db.
		Where("notified = 0 AND contest_id IN (?)",
			db.Model(&model.Contest{}).Select("id").
				Where("start_time > ? AND start_time <= ?", nowMs-2*3600*1000, nowMs+20*60*1000)).
		Find(&list).Error
	return list, err
}

func MarkBookingsNotified(db *gorm.DB, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return db.Model(&model.Booking{}).Where("id IN ?", ids).Update("notified", true).Error
}
