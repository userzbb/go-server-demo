// Package repository 提供数据访问层接口与 PostgreSQL 实现
package repository

import (
	"context"
	"errors"

	"omega-server/internal/model"
)

// 数据访问层通用错误
var (
	// ErrNotFound 记录不存在
	ErrNotFound = errors.New("记录不存在")
	// ErrUsernameExists 用户名已被占用
	ErrUsernameExists = errors.New("用户名已存在")
)

// PlayerRepository 玩家数据访问接口，便于单元测试使用 Mock 实现
type PlayerRepository interface {
	// Create 创建玩家并填充生成的 ID 与时间戳
	Create(ctx context.Context, p *model.Player) (*model.Player, error)
	// GetByUsername 按用户名查询玩家，不存在返回 ErrNotFound
	GetByUsername(ctx context.Context, username string) (*model.Player, error)
	// UpdateLastLogin 更新玩家最后登录时间
	UpdateLastLogin(ctx context.Context, playerID string) error
}
