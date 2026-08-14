package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"omega-server/internal/service"
	"omega-server/pkg/network"
	"omega-server/pkg/protocol"
)

// LoginHandler 处理登录请求（MsgID=1001 → 1002）
type LoginHandler struct {
	auth    *service.AuthService
	onLogin func(sess *network.Session, playerID string)
}

// NewLoginHandler 创建登录处理器
func NewLoginHandler(auth *service.AuthService) *LoginHandler {
	return &LoginHandler{auth: auth}
}

// SetOnLogin 设置登录成功回调，用于会话绑定玩家身份等收尾工作
func (h *LoginHandler) SetOnLogin(fn func(sess *network.Session, playerID string)) {
	h.onLogin = fn
}

// Handle 解析登录请求并回发登录响应
func (h *LoginHandler) Handle(sess *network.Session, body []byte) error {
	var req protocol.LoginRequest
	if err := json.Unmarshal(body, &req); err != nil || req.Username == "" || req.Password == "" {
		return reply(sess, protocol.MsgIDLoginResponse, protocol.LoginResponse{
			Code:    protocol.CodeBadRequest,
			Message: "请求格式错误",
		})
	}

	token, player, err := h.auth.Login(context.Background(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			return reply(sess, protocol.MsgIDLoginResponse, protocol.LoginResponse{
				Code:    protocol.CodeInvalidCredentials,
				Message: err.Error(),
			})
		case errors.Is(err, service.ErrBadRequest):
			return reply(sess, protocol.MsgIDLoginResponse, protocol.LoginResponse{
				Code:    protocol.CodeBadRequest,
				Message: err.Error(),
			})
		default:
			return fmt.Errorf("登录失败: %w", err)
		}
	}

	if h.onLogin != nil {
		h.onLogin(sess, player.ID)
	}
	return reply(sess, protocol.MsgIDLoginResponse, protocol.LoginResponse{
		Code:    protocol.CodeOK,
		Message: "登录成功",
		Token:   token,
		Player:  player,
	})
}

// RegisterHandler 处理注册请求（MsgID=1003 → 1004）
type RegisterHandler struct {
	auth *service.AuthService
}

// NewRegisterHandler 创建注册处理器
func NewRegisterHandler(auth *service.AuthService) *RegisterHandler {
	return &RegisterHandler{auth: auth}
}

// Handle 解析注册请求并回发注册响应
func (h *RegisterHandler) Handle(sess *network.Session, body []byte) error {
	var req protocol.LoginRequest
	if err := json.Unmarshal(body, &req); err != nil || req.Username == "" || req.Password == "" {
		return reply(sess, protocol.MsgIDRegisterResponse, protocol.LoginResponse{
			Code:    protocol.CodeBadRequest,
			Message: "请求格式错误",
		})
	}

	player, err := h.auth.Register(context.Background(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUsernameTaken):
			return reply(sess, protocol.MsgIDRegisterResponse, protocol.LoginResponse{
				Code:    protocol.CodeUsernameTaken,
				Message: err.Error(),
			})
		case errors.Is(err, service.ErrBadRequest):
			return reply(sess, protocol.MsgIDRegisterResponse, protocol.LoginResponse{
				Code:    protocol.CodeBadRequest,
				Message: err.Error(),
			})
		default:
			return fmt.Errorf("注册失败: %w", err)
		}
	}

	return reply(sess, protocol.MsgIDRegisterResponse, protocol.LoginResponse{
		Code:    protocol.CodeOK,
		Message: "注册成功",
		Player:  player,
	})
}
