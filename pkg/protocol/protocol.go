// Package protocol 定义客户端与服务器之间的消息 ID、错误码与数据结构
package protocol

// 消息 ID 常量，范围分配见 docs/PROTOCOL.md
const (
	// MsgIDLoginRequest 登录请求
	MsgIDLoginRequest uint32 = 1001
	// MsgIDLoginResponse 登录响应
	MsgIDLoginResponse uint32 = 1002
	// MsgIDRegisterRequest 注册请求
	MsgIDRegisterRequest uint32 = 1003
	// MsgIDRegisterResponse 注册响应
	MsgIDRegisterResponse uint32 = 1004
	// MsgIDHeartbeatRequest 心跳请求
	MsgIDHeartbeatRequest uint32 = 9001
	// MsgIDHeartbeatResponse 心跳响应
	MsgIDHeartbeatResponse uint32 = 9002
	// MsgIDCreateRoomRequest 创建房间请求
	MsgIDCreateRoomRequest uint32 = 2001
	// MsgIDCreateRoomResponse 创建房间响应
	MsgIDCreateRoomResponse uint32 = 2002
	// MsgIDJoinRoomRequest 加入房间请求
	MsgIDJoinRoomRequest uint32 = 2003
	// MsgIDJoinRoomResponse 加入房间响应
	MsgIDJoinRoomResponse uint32 = 2004
	// MsgIDLeaveRoomRequest 离开房间请求
	MsgIDLeaveRoomRequest uint32 = 2005
	// MsgIDLeaveRoomResponse 离开房间响应
	MsgIDLeaveRoomResponse uint32 = 2006
	// MsgIDMoveSyncRequest 移动同步请求
	MsgIDMoveSyncRequest uint32 = 3001
	// MsgIDStateSyncBroadcast 状态同步广播（服务端 → 同房间玩家）
	MsgIDStateSyncBroadcast uint32 = 3002
)

// 业务错误码，与 docs/PROTOCOL.md 第 5 节保持一致
const (
	// CodeOK 成功
	CodeOK int = 0
	// CodeInvalidCredentials 用户名或密码错误
	CodeInvalidCredentials int = 1
	// CodeBadRequest 请求格式错误
	CodeBadRequest int = 2
	// CodeRoomNotFound 房间不存在
	CodeRoomNotFound int = 3
	// CodeRoomFull 房间已满
	CodeRoomFull int = 4
	// CodeAlreadyInRoom 玩家已在房间中
	CodeAlreadyInRoom int = 5
	// CodeInvalidToken Token 无效或过期
	CodeInvalidToken int = 6
	// CodeUsernameTaken 用户名已被占用
	CodeUsernameTaken int = 7
)

// LoginRequest 登录/注册请求体（MsgID=1001/1003）
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

// CreateRoomRequest 创建房间请求体（MsgID=2001）
type CreateRoomRequest struct {
	RoomName   string `json:"roomName"`
	MaxPlayers int    `json:"maxPlayers"`
}

// CreateRoomResponse 创建房间响应体（MsgID=2002）
type CreateRoomResponse struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	RoomID     string `json:"roomId,omitempty"`
	RoomName   string `json:"roomName,omitempty"`
	MaxPlayers int    `json:"maxPlayers,omitempty"`
}

// JoinRoomRequest 加入房间请求体（MsgID=2003）
type JoinRoomRequest struct {
	RoomID string `json:"roomId"`
}

// JoinRoomResponse 加入房间响应体（MsgID=2004）
type JoinRoomResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	RoomID  string   `json:"roomId,omitempty"`
	Players []string `json:"players,omitempty"`
}

// LeaveRoomRequest 离开房间请求体（MsgID=2005）
type LeaveRoomRequest struct {
	RoomID string `json:"roomId"`
}

// LeaveRoomResponse 离开房间响应体（MsgID=2006）
type LeaveRoomResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MoveSyncRequest 移动同步请求体（MsgID=3001），客户端上报自身位置与速度
type MoveSyncRequest struct {
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	VX float64 `json:"vx"`
	VY float64 `json:"vy"`
}

// StateSync 状态同步广播体（MsgID=3002），服务端转发给同房间其他玩家
type StateSync struct {
	PlayerID string  `json:"playerId"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	VX       float64 `json:"vx"`
	VY       float64 `json:"vy"`
}

// ErrorResponse 通用错误响应体，用于无专用响应消息的场景（如状态同步失败）
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
