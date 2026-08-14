package handler

import (
	"encoding/json"
	"time"

	"omega-server/pkg/network"
	"omega-server/pkg/protocol"
)

// HeartbeatHandler 处理心跳请求（MsgID=9001 → 9002）
type HeartbeatHandler struct{}

// NewHeartbeatHandler 创建心跳处理器
func NewHeartbeatHandler() *HeartbeatHandler {
	return &HeartbeatHandler{}
}

// Handle 回发心跳响应，附上服务器当前时间
func (h *HeartbeatHandler) Handle(sess *network.Session, body []byte) error {
	var req protocol.HeartbeatRequest
	// 心跳容忍解析失败，客户端时间戳缺失时按 0 处理
	_ = json.Unmarshal(body, &req)

	resp := protocol.HeartbeatResponse{
		Timestamp:  req.Timestamp,
		ServerTime: time.Now().Unix(),
	}
	return reply(sess, protocol.MsgIDHeartbeatResponse, resp)
}
