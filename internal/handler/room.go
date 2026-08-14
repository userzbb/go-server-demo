package handler

import (
	"encoding/json"
	"errors"
	"fmt"

	"omega-server/internal/service"
	"omega-server/pkg/network"
	"omega-server/pkg/protocol"
)

// RoomHandler 处理房间操作（创建/加入/离开）
type RoomHandler struct {
	rooms    *service.RoomManager
	registry *SessionRegistry
}

// NewRoomHandler 创建房间处理器
func NewRoomHandler(rooms *service.RoomManager, registry *SessionRegistry) *RoomHandler {
	return &RoomHandler{rooms: rooms, registry: registry}
}

// HandleCreate 处理创建房间请求（MsgID=2001 → 2002）
func (h *RoomHandler) HandleCreate(sess *network.Session, body []byte) error {
	playerID := sess.PlayerID()
	if playerID == "" {
		return reply(sess, protocol.MsgIDCreateRoomResponse, protocol.CreateRoomResponse{
			Code:    protocol.CodeInvalidToken,
			Message: "请先登录",
		})
	}

	var req protocol.CreateRoomRequest
	if err := json.Unmarshal(body, &req); err != nil || req.RoomName == "" {
		return reply(sess, protocol.MsgIDCreateRoomResponse, protocol.CreateRoomResponse{
			Code:    protocol.CodeBadRequest,
			Message: "请求格式错误",
		})
	}

	room, err := h.rooms.CreateRoom(playerID, req.RoomName, req.MaxPlayers)
	if errors.Is(err, service.ErrBadRequest) {
		return reply(sess, protocol.MsgIDCreateRoomResponse, protocol.CreateRoomResponse{
			Code:    protocol.CodeBadRequest,
			Message: "房间名或人数不合法（人数 2-100）",
		})
	}
	if err != nil {
		return fmt.Errorf("创建房间: %w", err)
	}

	h.registry.Register(playerID, sess)
	return reply(sess, protocol.MsgIDCreateRoomResponse, protocol.CreateRoomResponse{
		Code:       protocol.CodeOK,
		Message:    "创建成功",
		RoomID:     room.ID,
		RoomName:   room.Name,
		MaxPlayers: room.MaxPlayers,
	})
}

// HandleJoin 处理加入房间请求（MsgID=2003 → 2004）
func (h *RoomHandler) HandleJoin(sess *network.Session, body []byte) error {
	playerID := sess.PlayerID()
	if playerID == "" {
		return reply(sess, protocol.MsgIDJoinRoomResponse, protocol.JoinRoomResponse{
			Code:    protocol.CodeInvalidToken,
			Message: "请先登录",
		})
	}

	var req protocol.JoinRoomRequest
	if err := json.Unmarshal(body, &req); err != nil || req.RoomID == "" {
		return reply(sess, protocol.MsgIDJoinRoomResponse, protocol.JoinRoomResponse{
			Code:    protocol.CodeBadRequest,
			Message: "请求格式错误",
		})
	}

	room, err := h.rooms.JoinRoom(req.RoomID, playerID)
	switch {
	case errors.Is(err, service.ErrRoomNotFound):
		return reply(sess, protocol.MsgIDJoinRoomResponse, protocol.JoinRoomResponse{
			Code:    protocol.CodeRoomNotFound,
			Message: err.Error(),
		})
	case errors.Is(err, service.ErrRoomFull):
		return reply(sess, protocol.MsgIDJoinRoomResponse, protocol.JoinRoomResponse{
			Code:    protocol.CodeRoomFull,
			Message: err.Error(),
		})
	case errors.Is(err, service.ErrAlreadyInRoom):
		return reply(sess, protocol.MsgIDJoinRoomResponse, protocol.JoinRoomResponse{
			Code:    protocol.CodeAlreadyInRoom,
			Message: err.Error(),
		})
	case err != nil:
		return fmt.Errorf("加入房间: %w", err)
	}

	h.registry.Register(playerID, sess)
	return reply(sess, protocol.MsgIDJoinRoomResponse, protocol.JoinRoomResponse{
		Code:    protocol.CodeOK,
		Message: "加入成功",
		RoomID:  room.ID,
		Players: room.Players,
	})
}

// HandleLeave 处理离开房间请求（MsgID=2005 → 2006）
func (h *RoomHandler) HandleLeave(sess *network.Session, body []byte) error {
	playerID := sess.PlayerID()
	if playerID == "" {
		return reply(sess, protocol.MsgIDLeaveRoomResponse, protocol.LeaveRoomResponse{
			Code:    protocol.CodeInvalidToken,
			Message: "请先登录",
		})
	}

	var req protocol.LeaveRoomRequest
	if err := json.Unmarshal(body, &req); err != nil || req.RoomID == "" {
		return reply(sess, protocol.MsgIDLeaveRoomResponse, protocol.LeaveRoomResponse{
			Code:    protocol.CodeBadRequest,
			Message: "请求格式错误",
		})
	}

	err := h.rooms.LeaveRoom(req.RoomID, playerID)
	switch {
	case errors.Is(err, service.ErrRoomNotFound):
		return reply(sess, protocol.MsgIDLeaveRoomResponse, protocol.LeaveRoomResponse{
			Code:    protocol.CodeRoomNotFound,
			Message: err.Error(),
		})
	case errors.Is(err, service.ErrNotInRoom):
		return reply(sess, protocol.MsgIDLeaveRoomResponse, protocol.LeaveRoomResponse{
			Code:    protocol.CodeRoomNotFound,
			Message: err.Error(),
		})
	case err != nil:
		return fmt.Errorf("离开房间: %w", err)
	}

	return reply(sess, protocol.MsgIDLeaveRoomResponse, protocol.LeaveRoomResponse{
		Code:    protocol.CodeOK,
		Message: "离开成功",
	})
}
