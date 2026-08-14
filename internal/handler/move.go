package handler

import (
	"encoding/json"
	"errors"

	"omega-server/internal/service"
	"omega-server/pkg/network"
	"omega-server/pkg/protocol"
)

// MoveHandler 处理移动同步请求（MsgID=3001 → 广播 3002 给同房间其他玩家）
type MoveHandler struct {
	rooms    *service.RoomManager
	registry *SessionRegistry
}

// NewMoveHandler 创建移动同步处理器
func NewMoveHandler(rooms *service.RoomManager, registry *SessionRegistry) *MoveHandler {
	return &MoveHandler{rooms: rooms, registry: registry}
}

// Handle 将玩家的位置状态广播给同房间其他成员
func (h *MoveHandler) Handle(sess *network.Session, body []byte) error {
	playerID := sess.PlayerID()
	if playerID == "" {
		return reply(sess, protocol.MsgIDStateSyncBroadcast, protocol.ErrorResponse{
			Code:    protocol.CodeInvalidToken,
			Message: "请先登录",
		})
	}

	var req protocol.MoveSyncRequest
	// 位置字段缺失时按 0 处理，容忍不完整请求
	_ = json.Unmarshal(body, &req)

	room, err := h.rooms.FindRoomByPlayer(playerID)
	if errors.Is(err, service.ErrNotInRoom) {
		return reply(sess, protocol.MsgIDStateSyncBroadcast, protocol.ErrorResponse{
			Code:    protocol.CodeRoomNotFound,
			Message: "请先加入房间",
		})
	}
	if err != nil {
		return err
	}

	state := protocol.StateSync{
		PlayerID: playerID,
		X:        req.X,
		Y:        req.Y,
		VX:       req.VX,
		VY:       req.VY,
	}
	stateBody, err := json.Marshal(state)
	if err != nil {
		return err
	}

	// 广播给同房间其他成员（排除来源玩家）
	h.registry.Broadcast(room.Players, playerID, protocol.MsgIDStateSyncBroadcast, stateBody)
	return nil
}
