// AcKingLabs V2.0 后端入口。
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"acking/internal/app"
	"acking/internal/config"
	"acking/internal/database"
	"acking/internal/router"
)

func main() {
	// 配置：默认 ./config.yaml，可用 -c 指定
	path := "./config.yaml"
	if len(os.Args) >= 3 && os.Args[1] == "-c" {
		path = os.Args[2]
	}
	cfg, err := config.Load(path)
	if err != nil {
		slog.Error("加载配置失败", "err", err)
		os.Exit(1)
	}
	app.Cfg = cfg

	// 日志
	level := slog.LevelInfo
	switch cfg.Log.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
	slog.Info("AcKingLabs V2.0 后端启动中...")

	// MySQL
	if err := database.Init(); err != nil {
		slog.Error("数据库初始化失败", "err", err)
		os.Exit(1)
	}

	// Redis（失败不阻断启动，相关功能自动降级）
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		slog.Warn("Redis 不可用，浏览量/验证码将降级为本地实现", "err", err)
	} else {
		app.RDB = rdb
		slog.Info("Redis 已连接", "addr", cfg.Redis.Addr)
	}

	// HTTP + 定时任务
	go router.Run()

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("收到退出信号，服务停止")
	if app.RDB != nil {
		_ = app.RDB.Close()
	}
}
