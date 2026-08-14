package handler

import (
	"sync"

	"omega-server/pkg/network"
)

// SessionRegistry 维护玩家 ID → 会话的映射，用于向房间成员广播消息
type SessionRegistry struct {
	mu       sync.RWMutex
	sessions map[string]*network.Session
}

// NewSessionRegistry 创建会话注册表
func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{sessions: make(map[string]*network.Session)}
}

// Register 登录成功后注册玩家会话，重复注册覆盖旧会话
func (r *SessionRegistry) Register(playerID string, sess *network.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[playerID] = sess
}

// Remove 连接断开时注销玩家会话
func (r *SessionRegistry) Remove(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, playerID)
}

// Get 按玩家 ID 获取会话
func (r *SessionRegistry) Get(playerID string) (*network.Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sess, ok := r.sessions[playerID]
	return sess, ok
}

// Broadcast 向一批玩家发送消息，exclude 指定的玩家被跳过（通常为消息来源）
func (r *SessionRegistry) Broadcast(playerIDs []string, exclude string, msgID uint32, body []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, pid := range playerIDs {
		if pid == exclude {
			continue
		}
		if sess, ok := r.sessions[pid]; ok {
			sess.SendMessage(msgID, body)
		}
	}
}
