package onlinestore

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Config 在线状态存储配置
type Config struct {
	// client Redis客户端（必需）
	client *redis.Client

	// onlineTTLSeconds 在线状态过期时间（秒），超过此时间未心跳的用户视为离线
	// 默认: 300秒（5分钟）
	onlineTTLSeconds int64

	// maxPageSize 分页查询时的最大页面大小
	// 默认: 1000
	maxPageSize int64

	// key Redis中存储在线状态的键名
	// 默认: "kit:online-store"
	key string

	// ctx 上下文，用于所有Redis操作
	// 默认: context.Background()
	ctx context.Context
}

// Option 配置选项函数
type Option func(*Config)

// WithOnlineTTLSeconds 设置在线状态过期时间（秒）
func WithOnlineTTLSeconds(seconds int64) Option {
	return func(c *Config) {
		c.onlineTTLSeconds = seconds
	}
}

// WithMaxPageSize 设置分页查询的最大页面大小
func WithMaxPageSize(size int64) Option {
	return func(c *Config) {
		c.maxPageSize = size
	}
}

// WithKey 设置Redis存储键名
func WithKey(key string) Option {
	return func(c *Config) {
		c.key = key
	}
}

// WithContext 设置上下文
func WithContext(ctx context.Context) Option {
	return func(c *Config) {
		c.ctx = ctx
	}
}

// defaultConfig 创建默认配置
func defaultConfig(client *redis.Client, opts ...Option) *Config {
	cfg := &Config{
		client:           client,
		onlineTTLSeconds: 300,
		maxPageSize:      1000,
		key:              "kit:online-store",
		ctx:              context.Background(),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}
