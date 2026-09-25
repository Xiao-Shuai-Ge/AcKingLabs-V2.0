// Package app 存放进程级全局对象（数据库/Redis/配置）。
package app

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"acking/internal/config"
)

var (
	Cfg *config.Config
	DB  *gorm.DB
	RDB *redis.Client // 可能为 nil（Redis 未启用时功能自动降级）
)

// RedisOK Redis 是否可用
func RedisOK() bool {
	return RDB != nil
}
