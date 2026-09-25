// Package database MySQL 初始化、建表、超管种子。
package database

import (
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"acking/internal/app"
	"acking/internal/model"
)

// Init 连接 MySQL；migrate 开启时执行 AutoMigrate，users 为空时播种超管
func Init() error {
	db, err := gorm.Open(mysql.Open(app.Cfg.Database.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // SQL 不落日志，避免噪音与信息泄露
	})
	if err != nil {
		return fmt.Errorf("连接 MySQL 失败: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)
	app.DB = db

	if app.Cfg.Database.Migrate {
		if err := db.AutoMigrate(
			&model.User{}, &model.Post{}, &model.PostLike{},
			&model.Comment{}, &model.CommentLike{},
			&model.Contest{}, &model.Booking{},
			&model.Resume{}, &model.Message{},
		); err != nil {
			return fmt.Errorf("自动建表失败: %w", err)
		}
		slog.Info("数据库表结构已就绪")
	}

	return seedAdmin(db)
}

// seedAdmin users 表为空时创建超级管理员
func seedAdmin(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	cfg := app.Cfg.Admin
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	admin := model.User{
		Username: cfg.Username,
		Password: string(hash),
		Email:    cfg.Email,
		Role:     model.RoleSuper,
	}
	if err := db.Create(&admin).Error; err != nil {
		return err
	}
	slog.Info("已创建超级管理员账号", "email", cfg.Email)
	return nil
}
