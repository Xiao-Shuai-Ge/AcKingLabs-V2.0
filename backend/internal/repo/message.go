package repo

import (
	"gorm.io/gorm"

	"acking/internal/model"
)

func CreateMessage(db *gorm.DB, m *model.Message) error {
	return db.Create(m).Error
}

// UnreadCounts 各类型未读数
func UnreadCounts(db *gorm.DB, userID int64) (total, like, comment, system int64, err error) {
	var rows []struct {
		Type string
		N    int64
	}
	err = db.Model(&model.Message{}).
		Select("type, COUNT(*) AS n").
		Where("user_id = ? AND is_read = 0", userID).
		Group("type").Scan(&rows).Error
	if err != nil {
		return
	}
	for _, r := range rows {
		total += r.N
		switch r.Type {
		case model.MsgTypeLike:
			like = r.N
		case model.MsgTypeComment:
			comment = r.N
		case model.MsgTypeSystem:
			system = r.N
		}
	}
	return
}

// MessageList 消息列表（游标 id < before）
func MessageList(db *gorm.DB, userID int64, typ string, before, count int64) ([]model.Message, error) {
	q := db.Model(&model.Message{}).Where("user_id = ? AND id < ?", userID, before)
	if typ != "" && typ != "all" {
		q = q.Where("type = ?", typ)
	}
	var list []model.Message
	err := q.Order("id DESC").Limit(int(count)).Find(&list).Error
	return list, err
}

func MarkMessageRead(db *gorm.DB, userID, messageID int64) error {
	return db.Model(&model.Message{}).
		Where("id = ? AND user_id = ?", messageID, userID).
		Update("is_read", true).Error
}

func MarkAllMessagesRead(db *gorm.DB, userID int64, typ string) error {
	q := db.Model(&model.Message{}).Where("user_id = ? AND is_read = 0", userID)
	if typ != "" && typ != "all" {
		q = q.Where("type = ?", typ)
	}
	return q.Update("is_read", true).Error
}
