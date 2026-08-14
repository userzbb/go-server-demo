package network

// Logger 是网络层使用的日志接口，便于注入 Zap 等结构化日志器
type Logger interface {
	Printf(format string, args ...any)
}
