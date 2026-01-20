package onlinestore

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/8liang/kit"
	"github.com/redis/go-redis/v9"
)

// New 创建一个新的在线状态存储实例
func New(client *redis.Client, opts ...Option) (*Store, error) {
	if client == nil {
		return nil, kit.ErrRedisOptIsRequired
	}

	cfg := defaultConfig(client, opts...)

	if cfg.onlineTTLSeconds <= 0 {
		return nil, fmt.Errorf("%w: OnlineTTLSeconds must be positive", kit.ErrParameterInvalid)
	}
	if cfg.maxPageSize <= 0 {
		return nil, fmt.Errorf("%w: MaxPageSize must be positive", kit.ErrParameterInvalid)
	}
	if cfg.key == "" {
		return nil, fmt.Errorf("%w: Key cannot be empty", kit.ErrParameterInvalid)
	}
	if cfg.ctx == nil {
		cfg.ctx = context.Background()
	}

	// 测试Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := cfg.client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis failed: %w", err)
	}

	return &Store{config: cfg}, nil
}

// Store 在线状态存储
type Store struct {
	config *Config
}

// Heartbeat 记录用户心跳，标记用户在线
func (s *Store) Heartbeat(userId string) error {
	return s.config.client.ZAdd(s.config.ctx, s.config.key, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: userId,
	}).Err()
}

// SetOffline 设置用户离线
func (s *Store) SetOffline(userId string) error {
	return s.config.client.ZRem(s.config.ctx, s.config.key, userId).Err()
}

// OnlineCount 获取当前在线用户数量
func (s *Store) OnlineCount() (int64, error) {
	return s.config.client.ZCount(
		s.config.ctx,
		s.config.key,
		strconv.FormatInt(s.staleTime(), 10),
		"+inf",
	).Result()
}

// OnlineUsers 获取在线用户列表
func (s *Store) OnlineUsers(opts ...OnlineOption) ([]string, error) {
	zRangeBy := &redis.ZRangeBy{
		Min: strconv.FormatInt(s.staleTime(), 10),
		Max: "+inf",
	}
	for _, opt := range opts {
		opt(zRangeBy)
	}
	return s.config.client.ZRangeByScore(s.config.ctx, s.config.key, zRangeBy).Result()
}

// EachOnlineUser 遍历所有在线用户
func (s *Store) EachOnlineUser(fn func(userId string) error) error {
	total, err := s.OnlineCount()
	if err != nil {
		return err
	}

	pages := total / s.config.maxPageSize
	if total%s.config.maxPageSize != 0 {
		pages++
	}

	for i := int64(1); i <= pages; i++ {
		users, err := s.OnlineUsers(WithPage(i, s.config.maxPageSize))
		if err != nil {
			return err
		}
		for _, userId := range users {
			if err := fn(userId); err != nil {
				return err
			}
		}
	}
	return nil
}

// FilterOnlineUsers 从给定的用户ID列表中筛选出在线的用户
// 返回 map[userId]lastActiveTimestamp
func (s *Store) FilterOnlineUsers(userIDs []string) (map[string]float64, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	pipe := s.config.client.Pipeline()
	cmds := make([]*redis.FloatCmd, 0, len(userIDs))
	for _, uid := range userIDs {
		cmds = append(cmds, pipe.ZScore(s.config.ctx, s.config.key, uid))
	}

	_, err := pipe.Exec(s.config.ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	result := make(map[string]float64)
	staleTime := float64(s.staleTime())
	for i, cmd := range cmds {
		score, _ := cmd.Result()
		if score >= staleTime {
			result[userIDs[i]] = score
		}
	}
	return result, nil
}

// CleanupOffline 清理离线用户数据
// 返回清理的用户数量
func (s *Store) CleanupOffline() (int64, error) {
	return s.config.client.ZRemRangeByScore(
		s.config.ctx,
		s.config.key,
		"0",
		strconv.FormatInt(s.staleTime(), 10),
	).Result()
}

// OnlineOption 在线用户查询选项
type OnlineOption func(*redis.ZRangeBy)

// WithPage 设置分页参数
func WithPage(page, pageSize int64) OnlineOption {
	return func(zr *redis.ZRangeBy) {
		if page <= 0 || pageSize <= 0 {
			return
		}
		zr.Offset = pageSize * (page - 1)
		zr.Count = pageSize
	}
}

// staleTime 计算过期时间戳
func (s *Store) staleTime() int64 {
	return time.Now().Unix() - s.config.onlineTTLSeconds
}
