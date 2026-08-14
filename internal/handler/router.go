// Package handler 提供消息路由与业务处理器
package handler

import (
	"fmt"

	"omega-server/pkg/network"
)

// Handler 处理一条消息；返回 error 表示内部技术故障（由网关记录日志），
// 业务失败通过协议错误码写入响应，不返回 error
type Handler func(sess *network.Session, body []byte) error

// Router 按消息 ID 分发请求到对应处理器
type Router struct {
	handlers map[uint32]Handler
}

// NewRouter 创建空路由表
func NewRouter() *Router {
	return &Router{handlers: make(map[uint32]Handler)}
}

// Register 注册消息处理器，重复注册会覆盖
func (r *Router) Register(msgID uint32, h Handler) {
	r.handlers[msgID] = h
}

// Dispatch 按消息 ID 分发，未注册的消息返回错误
func (r *Router) Dispatch(sess *network.Session, msgID uint32, body []byte) error {
	h, ok := r.handlers[msgID]
	if !ok {
		return fmt.Errorf("未注册的消息 ID: %d", msgID)
	}
	return h(sess, body)
}
