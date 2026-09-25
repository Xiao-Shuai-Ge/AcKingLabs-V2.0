// Package task 定时任务：热度重算、浏览量落库、比赛抓取、预约提醒。
// 所有任务都带 Recover，单个任务 panic 不拖垮进程。
package task

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"acking/internal/app"
	"acking/internal/service"
)

func safe(name string, fn func()) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("定时任务 panic", "task", name, "panic", r)
			}
		}()
		start := time.Now()
		fn()
		slog.Debug("定时任务完成", "task", name, "cost", time.Since(start).String())
	}
}

// addFunc 注册定时任务，注册失败直接 panic（cron 表达式是编译期常量，错了应当在启动时就暴露）
func addFunc(c *cron.Cron, spec, name string, fn func()) {
	if _, err := c.AddFunc(spec, safe(name, fn)); err != nil {
		panic(fmt.Sprintf("注册定时任务 %s 失败: %v", name, err))
	}
}

// StartCron 启动全部定时任务
func StartCron() {
	c := cron.New(cron.WithSeconds())

	// 浏览量落库：30 秒一次
	addFunc(c, "*/30 * * * * *", "flush-views", service.FlushViews)

	// 热度重算：5 分钟一次
	addFunc(c, "0 */5 * * * *", "hot-score", recomputeHotScore)

	// 比赛抓取：30 分钟一次 + 启动时先抓一次
	addFunc(c, "0 */30 * * * *", "scrape-contests", service.ScrapeContests)
	go func() {
		time.Sleep(5 * time.Second) // 等数据库/网络就绪
		safe("scrape-contests-boot", service.ScrapeContests)()
	}()

	// 预约开赛提醒：1 分钟一次
	addFunc(c, "0 * * * * *", "booking-notify", service.NotifyBookings)

	c.Start()
	slog.Info("定时任务已启动")
}

// recomputeHotScore 批量重算热度分：
//
//	hot_score = 基准时间 + 点赞*1h + 评论*1h + 管理员点赞*3d + 精选*14d
//	官方帖基准用 updated_at，其余用 created_at
func recomputeHotScore() {
	err := app.DB.Exec(`
UPDATE posts SET hot_score =
  (CASE WHEN type = 'official' THEN UNIX_TIMESTAMP(updated_at) ELSE UNIX_TIMESTAMP(created_at) END) * 1000
  + like_count * 3600000
  + comment_count * 3600000
  + is_admin_like * 259200000
  + is_featured * 1209600000
WHERE deleted_at IS NULL`).Error
	if err != nil {
		slog.Error("热度重算失败", "err", err)
	}
}
