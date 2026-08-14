package service

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"omega-server/internal/model"
)

// 房间业务错误，handler 据此映射协议错误码
var (
	// ErrRoomNotFound 房间不存在
	ErrRoomNotFound = errors.New("房间不存在")
	// ErrRoomFull 房间已满
	ErrRoomFull = errors.New("房间已满")
	// ErrAlreadyInRoom 玩家已在房间中
	ErrAlreadyInRoom = errors.New("玩家已在房间中")
	// ErrNotInRoom 玩家不在该房间中
	ErrNotInRoom = errors.New("玩家不在该房间中")
)

// RoomManager 管理房间生命周期（内存实现），线程安全
type RoomManager struct {
	mu     sync.RWMutex
	rooms  map[string]*model.Room
	nextID int
}

// NewRoomManager 创建房间管理器
func NewRoomManager() *RoomManager {
	return &RoomManager{rooms: make(map[string]*model.Room)}
}

// CreateRoom 创建房间，房主自动成为第一个成员
func (m *RoomManager) CreateRoom(ownerID, name string, maxPlayers int) (*model.Room, error) {
	if ownerID == "" || name == "" {
		return nil, ErrBadRequest
	}
	if maxPlayers < 2 || maxPlayers > 100 {
		return nil, ErrBadRequest
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	room := &model.Room{
		ID:         fmt.Sprintf("room_%03d", m.nextID),
		Name:       name,
		OwnerID:    ownerID,
		MaxPlayers: maxPlayers,
		Players:    []string{ownerID},
		Status:     model.RoomStatusWaiting,
		CreatedAt:  time.Now(),
	}
	m.rooms[room.ID] = room
	return copyRoom(room), nil
}

// JoinRoom 玩家加入房间
func (m *RoomManager) JoinRoom(roomID, playerID string) (*model.Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return nil, ErrRoomNotFound
	}
	for _, p := range room.Players {
		if p == playerID {
			return nil, ErrAlreadyInRoom
		}
	}
	if len(room.Players) >= room.MaxPlayers {
		return nil, ErrRoomFull
	}
	room.Players = append(room.Players, playerID)
	return copyRoom(room), nil
}

// LeaveRoom 玩家离开房间；房主离开时转移房主给剩余第一个成员，空房间被删除
func (m *RoomManager) LeaveRoom(roomID, playerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return ErrRoomNotFound
	}

	idx := -1
	for i, p := range room.Players {
		if p == playerID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return ErrNotInRoom
	}

	room.Players = append(room.Players[:idx], room.Players[idx+1:]...)
	if len(room.Players) == 0 {
		delete(m.rooms, roomID)
		return nil
	}
	if room.OwnerID == playerID {
		room.OwnerID = room.Players[0]
	}
	return nil
}

// GetRoom 按 ID 查询房间副本
func (m *RoomManager) GetRoom(roomID string) (*model.Room, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return nil, ErrRoomNotFound
	}
	return copyRoom(room), nil
}

// FindRoomByPlayer 查找玩家所在的房间，不在任何房间返回 ErrNotInRoom
func (m *RoomManager) FindRoomByPlayer(playerID string) (*model.Room, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, room := range m.rooms {
		for _, p := range room.Players {
			if p == playerID {
				return copyRoom(room), nil
			}
		}
	}
	return nil, ErrNotInRoom
}

// Count 返回当前房间总数
func (m *RoomManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms)
}

// copyRoom 返回房间的浅拷贝，Players 切片独立复制
func copyRoom(r *model.Room) *model.Room {
	cp := *r
	cp.Players = append([]string(nil), r.Players...)
	return &cp
}
