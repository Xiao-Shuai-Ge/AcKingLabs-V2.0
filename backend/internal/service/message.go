package service

import (
	"log/slog"
	"regexp"
	"strconv"

	"gorm.io/gorm"

	"acking/internal/app"
	"acking/internal/model"
	"acking/internal/repo"
	"acking/pkg/emailkit"
)

var mentionRe = regexp.MustCompile(`\[@[^\]]+\]\(/profile/(\d+)\)`)

// notify 写一条站内消息。
// kind 决定受哪个通知偏好开关控制（并映射到数据库消息类型）：
//
//	like -> 点赞通知；comment -> 评论/回复通知；mention -> 提及通知；
//	system -> 系统消息（不受开关限制，可选同步邮件）；broadcast -> 求助帖广播
func notify(tx *gorm.DB, recipient, sender int64, kind, content, url string) {
	if recipient == sender || recipient <= 0 {
		return
	}
	u, err := repo.GetUserByID(tx, recipient)
	if err != nil {
		return
	}
	s := u.NotifyOf()

	dbType := ""
	switch kind {
	case model.MsgTypeLike:
		if !s.Like {
			return
		}
		dbType = model.MsgTypeLike
	case "comment":
		if !s.Comment {
			return
		}
		dbType = model.MsgTypeComment
	case "mention":
		if !s.Mention {
			return
		}
		dbType = model.MsgTypeComment
	case "system":
		dbType = model.MsgTypeSystem
	case "broadcast":
		if !s.HelpPost {
			return
		}
		dbType = model.MsgTypeSystem
	default:
		return
	}

	msg := &model.Message{UserID: recipient, SenderID: sender, Type: dbType, Content: content, Url: url}
	if err := repo.CreateMessage(tx, msg); err != nil {
		// 消息失败不影响主流程
		_ = repo.CreateMessage(app.DB, msg)
	}

	// 系统消息且用户开启邮件同步：异步发送，不阻塞请求
	if kind == "system" && s.SystemEmail && u.Email != "" && app.Cfg.EmailConfigured() {
		to, content, url := u.Email, content, url
		go func() {
			if err := emailkit.SendSystemNotice(to, content, url); err != nil {
				slog.Error("系统消息邮件发送失败", "to", to, "err", err)
			}
		}()
	}
}

// notifyMentions 解析 @提及（[@名字](/profile/id)）并通知被提及用户
func notifyMentions(tx *gorm.DB, sender int64, content, url string) {
	seen := map[int64]bool{}
	for _, m := range mentionRe.FindAllStringSubmatch(content, 20) {
		uid, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil || uid <= 0 || seen[uid] {
			continue
		}
		seen[uid] = true
		notify(tx, uid, sender, "mention", "在帖子/评论中提到了你", url)
	}
}

// MessageCounts 未读数
type MessageCounts struct {
	Count   int64 `json:"count"`
	Like    int64 `json:"like_count"`
	Comment int64 `json:"comment_count"`
	System  int64 `json:"system_count"`
}

func GetMessageCounts(userID int64) (*MessageCounts, error) {
	total, like, comment, system, err := repo.UnreadCounts(app.DB, userID)
	if err != nil {
		return nil, err
	}
	return &MessageCounts{Count: total, Like: like, Comment: comment, System: system}, nil
}

// MessageItem 消息行
type MessageItem struct {
	ID        int64        `json:"id,string"`
	Type      string       `json:"type"`
	Content   string       `json:"content"`
	Url       string       `json:"url"`
	IsRead    bool         `json:"is_read"`
	CreatedAt int64        `json:"created_at"`
	Sender    *AuthorBrief `json:"sender"`
}

func GetMessageList(userID int64, typ string, before, count int64) ([]*MessageItem, error) {
	if before <= 0 {
		before = 1 << 62
	}
	if count <= 0 || count > 50 {
		count = 20
	}
	list, err := repo.MessageList(app.DB, userID, typ, before, count)
	if err != nil {
		return nil, err
	}
	senderIDs := make([]int64, 0, len(list))
	for _, m := range list {
		if m.SenderID > 0 {
			senderIDs = append(senderIDs, m.SenderID)
		}
	}
	senders := authorMap(senderIDs)
	out := make([]*MessageItem, 0, len(list))
	for _, m := range list {
		item := &MessageItem{
			ID: m.ID, Type: m.Type, Content: m.Content, Url: m.Url,
			IsRead: m.IsRead, CreatedAt: m.CreatedAt.UnixMilli(),
		}
		if m.SenderID > 0 {
			item.Sender = toBrief(senders[m.SenderID])
		} else {
			item.Sender = &AuthorBrief{Username: "系统", Avatar: "/assets/AcKing.png"}
		}
		out = append(out, item)
	}
	return out, nil
}

// MarkRead 单条已读；id=0 且 all=true 时全部已读
func MarkRead(userID int64, id int64, all bool, typ string) error {
	if all {
		return repo.MarkAllMessagesRead(app.DB, userID, typ)
	}
	if id <= 0 {
		return repo.MarkAllMessagesRead(app.DB, userID, "all")
	}
	return repo.MarkMessageRead(app.DB, userID, id)
}

// BroadcastHelpPost 新求助帖广播：通知所有开启"求助帖提醒"的用户（站内消息）
func BroadcastHelpPost(sender int64, title, url string) {
	var ids []int64
	err := app.DB.Raw(`
SELECT id FROM users
WHERE deleted_at IS NULL AND id != ?
  AND (settings IS NULL
       OR JSON_UNQUOTE(JSON_EXTRACT(settings, '$.notify.help_post')) = 'true')`, sender).Scan(&ids).Error
	if err != nil {
		slog.Error("查询求助帖广播目标失败", "err", err)
		return
	}
	msgs := make([]model.Message, 0, len(ids))
	for _, uid := range ids {
		msgs = append(msgs, model.Message{
			UserID: uid, SenderID: sender, Type: model.MsgTypeSystem,
			Content: "新求助帖：《" + truncateRunes(title, 20) + "》", Url: url,
		})
	}
	if len(msgs) == 0 {
		return
	}
	if err := app.DB.CreateInBatches(&msgs, 500).Error; err != nil {
		slog.Error("求助帖广播写入失败", "err", err)
	} else {
		slog.Info("求助帖广播完成", "recipients", len(msgs))
	}
}
