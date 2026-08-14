package service

import (
	"errors"
	"testing"

	"omega-server/internal/model"
)

func TestCreateRoom(t *testing.T) {
	m := NewRoomManager()

	tests := []struct {
		name       string
		ownerID    string
		roomName   string
		maxPlayers int
		wantErr    error
	}{
		{name: "正常创建", ownerID: "p1", roomName: "我的房间", maxPlayers: 10},
		{name: "空房主", ownerID: "", roomName: "房间", maxPlayers: 10, wantErr: ErrBadRequest},
		{name: "空房名", ownerID: "p1", roomName: "", maxPlayers: 10, wantErr: ErrBadRequest},
		{name: "人数过少", ownerID: "p1", roomName: "房间", maxPlayers: 1, wantErr: ErrBadRequest},
		{name: "人数过多", ownerID: "p1", roomName: "房间", maxPlayers: 101, wantErr: ErrBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room, err := m.CreateRoom(tt.ownerID, tt.roomName, tt.maxPlayers)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("期望错误 %v, 实际 %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("CreateRoom 返回错误: %v", err)
			}
			if room.OwnerID != tt.ownerID {
				t.Errorf("房主不匹配: %q", room.OwnerID)
			}
			if len(room.Players) != 1 || room.Players[0] != tt.ownerID {
				t.Errorf("房主应自动成为第一个成员: %v", room.Players)
			}
			if room.Status != model.RoomStatusWaiting {
				t.Errorf("新房间状态应为 waiting: %q", room.Status)
			}
		})
	}
}

func TestJoinRoom(t *testing.T) {
	m := NewRoomManager()
	room, err := m.CreateRoom("p1", "房间", 2)
	if err != nil {
		t.Fatalf("准备数据失败: %v", err)
	}

	tests := []struct {
		name     string
		roomID   string
		playerID string
		wantErr  error
	}{
		{name: "正常加入", roomID: room.ID, playerID: "p2"},
		{name: "重复加入", roomID: room.ID, playerID: "p2", wantErr: ErrAlreadyInRoom},
		{name: "房间不存在", roomID: "room_999", playerID: "p3", wantErr: ErrRoomNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			joined, err := m.JoinRoom(tt.roomID, tt.playerID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("期望错误 %v, 实际 %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("JoinRoom 返回错误: %v", err)
			}
			if len(joined.Players) != 2 {
				t.Errorf("成员数应为 2: %v", joined.Players)
			}
		})
	}

	t.Run("房间已满", func(t *testing.T) {
		if _, err := m.JoinRoom(room.ID, "p3"); !errors.Is(err, ErrRoomFull) {
			t.Fatalf("期望 ErrRoomFull, 实际 %v", err)
		}
	})
}

func TestLeaveRoom(t *testing.T) {
	t.Run("房主离开转移房主", func(t *testing.T) {
		m := NewRoomManager()
		room, _ := m.CreateRoom("p1", "房间", 4)
		_, _ = m.JoinRoom(room.ID, "p2")

		if err := m.LeaveRoom(room.ID, "p1"); err != nil {
			t.Fatalf("LeaveRoom 返回错误: %v", err)
		}
		got, err := m.GetRoom(room.ID)
		if err != nil {
			t.Fatalf("GetRoom 返回错误: %v", err)
		}
		if got.OwnerID != "p2" {
			t.Errorf("房主应转移给 p2, 实际 %q", got.OwnerID)
		}
		if len(got.Players) != 1 {
			t.Errorf("成员数应为 1: %v", got.Players)
		}
	})

	t.Run("最后一人离开删除房间", func(t *testing.T) {
		m := NewRoomManager()
		room, _ := m.CreateRoom("p1", "房间", 4)

		if err := m.LeaveRoom(room.ID, "p1"); err != nil {
			t.Fatalf("LeaveRoom 返回错误: %v", err)
		}
		if _, err := m.GetRoom(room.ID); !errors.Is(err, ErrRoomNotFound) {
			t.Fatalf("房间应已被删除, 实际 %v", err)
		}
		if m.Count() != 0 {
			t.Errorf("房间总数应为 0, 实际 %d", m.Count())
		}
	})

	t.Run("不在房间", func(t *testing.T) {
		m := NewRoomManager()
		room, _ := m.CreateRoom("p1", "房间", 4)
		if err := m.LeaveRoom(room.ID, "p2"); !errors.Is(err, ErrNotInRoom) {
			t.Fatalf("期望 ErrNotInRoom, 实际 %v", err)
		}
	})
}

func TestFindRoomByPlayer(t *testing.T) {
	m := NewRoomManager()
	room, _ := m.CreateRoom("p1", "房间", 4)
	_, _ = m.JoinRoom(room.ID, "p2")

	got, err := m.FindRoomByPlayer("p2")
	if err != nil {
		t.Fatalf("FindRoomByPlayer 返回错误: %v", err)
	}
	if got.ID != room.ID {
		t.Errorf("房间 ID 不匹配: %q != %q", got.ID, room.ID)
	}

	if _, err := m.FindRoomByPlayer("nobody"); !errors.Is(err, ErrNotInRoom) {
		t.Fatalf("期望 ErrNotInRoom, 实际 %v", err)
	}
}
