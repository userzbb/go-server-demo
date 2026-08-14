package model

import "time"

// 房间状态常量，与 docs/SCHEMA.md 第 4.1 节保持一致
const (
	// RoomStatusWaiting 等待中
	RoomStatusWaiting = "waiting"
	// RoomStatusPlaying 游戏中
	RoomStatusPlaying = "playing"
	// RoomStatusClosed 已关闭
	RoomStatusClosed = "closed"
)

// Room 房间实体，对应 rooms 表；Alpha 阶段采用内存管理
type Room struct {
	ID         string
	Name       string
	OwnerID    string
	MaxPlayers int
	Players    []string // 成员列表，按加入顺序
	Status     string
	CreatedAt  time.Time
}
