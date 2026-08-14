// Package protocol 定义客户端与服务器之间的消息 ID、错误码与数据结构
package protocol

// 消息 ID 常量，范围分配见 docs/PROTOCOL.md
const (
	// MsgIDLoginRequest 登录/注册请求
	MsgIDLoginRequest uint32 = 1001
	// MsgIDLoginResponse 登录/注册响应
	MsgIDLoginResponse uint32 = 1002
	// MsgIDHeartbeatRequest 心跳请求
	MsgIDHeartbeatRequest uint32 = 9001
	// MsgIDHeartbeatResponse 心跳响应
	MsgIDHeartbeatResponse uint32 = 9002
)

// 业务错误码，与 docs/PROTOCOL.md 第 5 节保持一致
const (
	// CodeOK 成功
	CodeOK int = 0
	// CodeInvalidCredentials 用户名或密码错误
	CodeInvalidCredentials int = 1
	// CodeBadRequest 请求格式错误
	CodeBadRequest int = 2
	// CodeInvalidToken Token 无效或过期
	CodeInvalidToken int = 6
)

// LoginRequest 登录/注册请求体（MsgID=1001）
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// PlayerInfo 返回给客户端的玩家摘要信息
type PlayerInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Level    int    `json:"level"`
}

// LoginResponse 登录/注册响应体（MsgID=1002）
type LoginResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Token   string      `json:"token,omitempty"`
	Player  *PlayerInfo `json:"player,omitempty"`
}

// HeartbeatRequest 心跳请求体（MsgID=9001）
type HeartbeatRequest struct {
	Timestamp int64 `json:"timestamp"`
}

// HeartbeatResponse 心跳响应体（MsgID=9002）
type HeartbeatResponse struct {
	Timestamp  int64 `json:"timestamp"`
	ServerTime int64 `json:"server_time"`
}
