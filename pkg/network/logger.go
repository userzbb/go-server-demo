// Package network 提供游戏服务器的网络层功能，包括 TCP 连接管理、消息编解码和会话控制
package network

// Logger 是网络层使用的日志接口，zap.SugaredLogger 天然满足该接口
type Logger interface {
	// Infof 记录一条普通级别日志
	Infof(format string, args ...any)
	// Errorf 记录一条错误级别日志
	Errorf(format string, args ...any)
}
