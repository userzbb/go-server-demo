// Package model 定义核心数据模型，字段与 docs/SCHEMA.md 对应
package model

import "time"

// Player 玩家实体，对应 players 表
type Player struct {
	ID           string
	Username     string
	PasswordHash string
	Email        string
	Nickname     string
	Level        int
	Exp          int64
	Gold         int64
	Diamond      int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
}
