package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App struct {
		Host           string   `yaml:"host"`
		Port           int      `yaml:"port"`
		BaseURL        string   `yaml:"base_url"`
		UploadDir      string   `yaml:"upload_dir"`
		StaticDir      string   `yaml:"static_dir"`
		TrustedProxies []string `yaml:"trusted_proxies"`
	} `yaml:"app"`

	Log struct {
		Level string `yaml:"level"`
		Dir   string `yaml:"dir"`
	} `yaml:"log"`

	Database struct {
		DSN     string `yaml:"dsn"`
		Migrate bool   `yaml:"migrate"`
	} `yaml:"database"`

	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`

	JWT struct {
		Secret string `yaml:"secret"`
	} `yaml:"jwt"`

	Email struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"email"`

	Admin struct {
		Email    string `yaml:"email"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"admin"`

	OSS struct {
		Endpoint        string `yaml:"endpoint"`
		AccessKeyID     string `yaml:"access_key_id"`
		AccessKeySecret string `yaml:"access_key_secret"`
		Bucket          string `yaml:"bucket"`
	} `yaml:"oss"`

	Invitation struct {
		Code string `yaml:"code"`
	} `yaml:"invitation"`
}

// Load 从给定路径读取配置；path 为空时依次尝试 ./config.yaml、config.example.yaml
func Load(path string) (*Config, error) {
	candidates := []string{path, "./config.yaml"}
	for _, p := range candidates {
		if p == "" {
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		cfg := &Config{}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("解析配置文件 %s 失败: %w", p, err)
		}
		fillDefaults(cfg)
		return cfg, nil
	}
	return nil, fmt.Errorf("找不到配置文件（尝试过: %v）", candidates)
}

func fillDefaults(c *Config) {
	if c.App.Port == 0 {
		c.App.Port = 8080
	}
	if c.App.UploadDir == "" {
		c.App.UploadDir = "./data/uploads"
	}
	if c.App.StaticDir == "" {
		c.App.StaticDir = "./static"
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Redis.Addr == "" {
		c.Redis.Addr = "127.0.0.1:6379"
	}
}

// OSSConfigured OSS 是否配置完整（配置后上传走 OSS，否则本地磁盘）
func (c *Config) OSSConfigured() bool {
	return c.OSS.Endpoint != "" && c.OSS.AccessKeyID != "" && c.OSS.AccessKeySecret != "" && c.OSS.Bucket != ""
}

// EmailConfigured 邮箱是否可用（投递简历通知、预约提醒等都依赖）
func (c *Config) EmailConfigured() bool {
	return c.Email.Host != "" && c.Email.Host != "smtp.example.com" && c.Email.Username != ""
}

// AccessTTL / RefreshTTL 令牌有效期
const (
	AccessTTL  = 2 * time.Hour
	RefreshTTL = 14 * 24 * time.Hour
	// RefreshTTLShort 未勾选"记住我"时的刷新令牌有效期
	RefreshTTLShort = 24 * time.Hour
)
