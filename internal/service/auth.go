// Package service 实现核心业务逻辑，包括注册、登录鉴权等
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"omega-server/internal/model"
	"omega-server/internal/repository"
	"omega-server/pkg/protocol"
)

// 认证业务错误，handler 据此映射协议错误码
var (
	// ErrInvalidCredentials 用户名或密码错误
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	// ErrUsernameTaken 用户名已被占用
	ErrUsernameTaken = errors.New("用户名已被占用")
	// ErrBadRequest 请求参数不合法
	ErrBadRequest = errors.New("请求参数不合法")
	// ErrInvalidToken 令牌无效或已过期
	ErrInvalidToken = errors.New("令牌无效或已过期")
)

// AuthService 处理玩家注册与登录鉴权
type AuthService struct {
	repo      repository.PlayerRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

// NewAuthService 创建认证服务，jwtSecret 必须来自环境变量
func NewAuthService(repo repository.PlayerRepository, jwtSecret string, tokenTTL time.Duration) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  tokenTTL,
	}
}

// Register 注册新玩家，返回玩家摘要信息
func (s *AuthService) Register(ctx context.Context, username, password string) (*protocol.PlayerInfo, error) {
	if username == "" || password == "" {
		return nil, ErrBadRequest
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("生成密码哈希: %w", err)
	}

	player := &model.Player{
		Username:     username,
		PasswordHash: string(hash),
		Nickname:     username,
		Level:        1,
	}
	created, err := s.repo.Create(ctx, player)
	if errors.Is(err, repository.ErrUsernameExists) {
		return nil, ErrUsernameTaken
	}
	if err != nil {
		return nil, fmt.Errorf("创建玩家: %w", err)
	}
	return toPlayerInfo(created), nil
}

// Login 校验凭证并签发 JWT，返回令牌与玩家摘要信息
func (s *AuthService) Login(ctx context.Context, username, password string) (string, *protocol.PlayerInfo, error) {
	if username == "" || password == "" {
		return "", nil, ErrBadRequest
	}

	player, err := s.repo.GetByUsername(ctx, username)
	if errors.Is(err, repository.ErrNotFound) {
		return "", nil, ErrInvalidCredentials
	}
	if err != nil {
		return "", nil, fmt.Errorf("查询玩家: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(player.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(player.ID)
	if err != nil {
		return "", nil, err
	}

	if err := s.repo.UpdateLastLogin(ctx, player.ID); err != nil {
		return "", nil, fmt.Errorf("更新登录时间: %w", err)
	}
	return token, toPlayerInfo(player), nil
}

// VerifyToken 校验 JWT 并返回玩家 ID，失败返回 ErrInvalidToken
func (s *AuthService) VerifyToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名算法: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}
	subject, err := claims.GetSubject()
	if err != nil {
		return "", ErrInvalidToken
	}
	return subject, nil
}

// generateToken 签发 HS256 签名 JWT，subject 为玩家 ID
func (s *AuthService) generateToken(playerID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   playerID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("签发令牌: %w", err)
	}
	return signed, nil
}

// toPlayerInfo 将玩家实体转换为协议摘要信息
func toPlayerInfo(p *model.Player) *protocol.PlayerInfo {
	return &protocol.PlayerInfo{
		ID:       p.ID,
		Username: p.Username,
		Nickname: p.Nickname,
		Level:    p.Level,
	}
}
