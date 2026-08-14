// Package config 负责加载服务配置，支持 YAML 文件 + 环境变量覆盖
package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

// Server 服务器配置
type Server struct {
	Name         string `mapstructure:"name"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// Log 日志配置
type Log struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// Game 游戏逻辑配置
type Game struct {
	TickRate          int `mapstructure:"tick_rate"`
	MaxRoomSize       int `mapstructure:"max_room_size"`
	HeartbeatInterval int `mapstructure:"heartbeat_interval"`
	HeartbeatTimeout  int `mapstructure:"heartbeat_timeout"`
}

// Postgres PostgreSQL 配置
type Postgres struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Database    string `mapstructure:"database"`
	MaxOpenConn int    `mapstructure:"max_open_conns"`
	MaxIdleConn int    `mapstructure:"max_idle_conns"`
}

// MongoDB MongoDB 配置
type MongoDB struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Database    string `mapstructure:"database"`
	MaxPoolSize int    `mapstructure:"max_pool_size"`
}

// Redis Redis 配置
type Redis struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// Config 根配置结构，字段与 configs/dev.yaml 对应
type Config struct {
	Server   Server   `mapstructure:"server"`
	Log      Log      `mapstructure:"log"`
	Game     Game     `mapstructure:"game"`
	Postgres Postgres `mapstructure:"postgres"`
	MongoDB  MongoDB  `mapstructure:"mongodb"`
	Redis    Redis    `mapstructure:"redis"`
	// JWTSecret 仅从环境变量注入，禁止写入配置文件
	JWTSecret string `mapstructure:"jwt_secret"`
}

// envBindings 配置文件键与 .env.example 环境变量的对应关系
var envBindings = map[string]string{
	"server.port":       "SERVER_PORT",
	"log.level":         "LOG_LEVEL",
	"postgres.host":     "PG_HOST",
	"postgres.port":     "PG_PORT",
	"postgres.user":     "PG_USER",
	"postgres.password": "PG_PASSWORD",
	"postgres.database": "PG_DATABASE",
	"mongodb.host":      "MONGO_HOST",
	"mongodb.port":      "MONGO_PORT",
	"mongodb.user":      "MONGO_USER",
	"mongodb.password":  "MONGO_PASSWORD",
	"mongodb.database":  "MONGO_DATABASE",
	"redis.host":        "REDIS_HOST",
	"redis.port":        "REDIS_PORT",
	"redis.password":    "REDIS_PASSWORD",
	"jwt_secret":        "JWT_SECRET",
}

// Load 从指定路径加载配置文件，环境变量优先于文件值
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件 %s: %w", path, err)
	}

	var errs []error
	for key, env := range envBindings {
		if err := v.BindEnv(key, env); err != nil {
			errs = append(errs, fmt.Errorf("绑定环境变量 %s 到配置键 %s: %w", env, key, err))
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置: %w", err)
	}

	// JWT 密钥只允许来自环境变量，文件中的值不作为兜底
	cfg.JWTSecret = v.GetString("jwt_secret")

	return &cfg, nil
}
