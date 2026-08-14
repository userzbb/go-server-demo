// Package logger 提供基于 Zap 的结构化日志封装，业务代码禁止直接使用标准库 log
package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config 日志初始化参数
type Config struct {
	Level   string // debug / info / warn / error，空值默认为 info
	Format  string // console / json，空值默认为 console
	Service string // 服务名，自动附加到每条日志
}

// New 根据配置创建 SugaredLogger，输出到 stdout
func New(cfg Config) (*zap.SugaredLogger, error) {
	level := cfg.Level
	if level == "" {
		level = "info"
	}
	lv, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("解析日志级别 %q: %w", cfg.Level, err)
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), lv)
	z := zap.New(core, zap.AddCaller())
	if cfg.Service != "" {
		z = z.Named(cfg.Service)
	}
	return z.Sugar(), nil
}
