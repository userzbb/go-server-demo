package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"omega-server/internal/model"
	"omega-server/internal/repository"
)

// fakePlayerRepo 内存版玩家仓库，仅用于单元测试
type fakePlayerRepo struct {
	players map[string]*model.Player
	nextID  int
}

func newFakePlayerRepo() *fakePlayerRepo {
	return &fakePlayerRepo{players: map[string]*model.Player{}}
}

func (f *fakePlayerRepo) Create(_ context.Context, p *model.Player) (*model.Player, error) {
	if _, ok := f.players[p.Username]; ok {
		return nil, repository.ErrUsernameExists
	}
	f.nextID++
	p.ID = fmt.Sprintf("player-%d", f.nextID)
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	f.players[p.Username] = p
	return p, nil
}

func (f *fakePlayerRepo) GetByUsername(_ context.Context, username string) (*model.Player, error) {
	p, ok := f.players[username]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return p, nil
}

func (f *fakePlayerRepo) UpdateLastLogin(_ context.Context, _ string) error {
	return nil
}

func newTestAuthService(repo repository.PlayerRepository) *AuthService {
	return NewAuthService(repo, "test-secret", time.Hour)
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		password  string
		preCreate bool
		wantErr   error
	}{
		{name: "正常注册", username: "player1", password: "123456"},
		{name: "空用户名", username: "", password: "123456", wantErr: ErrBadRequest},
		{name: "空密码", username: "player1", password: "", wantErr: ErrBadRequest},
		{name: "重复用户名", username: "player1", password: "654321", preCreate: true, wantErr: ErrUsernameTaken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestAuthService(newFakePlayerRepo())
			if tt.preCreate {
				if _, err := svc.Register(context.Background(), "player1", "123456"); err != nil {
					t.Fatalf("准备数据失败: %v", err)
				}
			}

			info, err := svc.Register(context.Background(), tt.username, tt.password)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("期望错误 %v, 实际 %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Register 返回错误: %v", err)
			}
			if info.Username != tt.username {
				t.Errorf("用户名不匹配: %q", info.Username)
			}
			if info.Level != 1 {
				t.Errorf("初始等级应为 1, 实际 %d", info.Level)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	setup := func(t *testing.T) *AuthService {
		t.Helper()
		svc := newTestAuthService(newFakePlayerRepo())
		if _, err := svc.Register(context.Background(), "player1", "123456"); err != nil {
			t.Fatalf("准备数据失败: %v", err)
		}
		return svc
	}

	tests := []struct {
		name     string
		username string
		password string
		wantErr  error
	}{
		{name: "正确凭证", username: "player1", password: "123456"},
		{name: "密码错误", username: "player1", password: "wrong-pass", wantErr: ErrInvalidCredentials},
		{name: "用户不存在", username: "nobody", password: "123456", wantErr: ErrInvalidCredentials},
		{name: "空参数", username: "", password: "", wantErr: ErrBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := setup(t)

			token, info, err := svc.Login(context.Background(), tt.username, tt.password)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("期望错误 %v, 实际 %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Login 返回错误: %v", err)
			}
			if token == "" {
				t.Error("期望返回非空 token")
			}
			if info.Username != tt.username {
				t.Errorf("玩家信息不匹配: %q", info.Username)
			}
		})
	}
}

func TestVerifyToken(t *testing.T) {
	svc := newTestAuthService(newFakePlayerRepo())
	if _, err := svc.Register(context.Background(), "player1", "123456"); err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	token, info, err := svc.Login(context.Background(), "player1", "123456")
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	t.Run("有效令牌", func(t *testing.T) {
		playerID, err := svc.VerifyToken(token)
		if err != nil {
			t.Fatalf("VerifyToken 返回错误: %v", err)
		}
		if playerID != info.ID {
			t.Errorf("玩家 ID 不匹配: %q != %q", playerID, info.ID)
		}
	})

	t.Run("无效令牌", func(t *testing.T) {
		if _, err := svc.VerifyToken("invalid-token"); !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("期望 ErrInvalidToken, 实际 %v", err)
		}
	})

	t.Run("篡改令牌", func(t *testing.T) {
		tampered := token[:len(token)-4] + "AAAA"
		if _, err := svc.VerifyToken(tampered); !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("期望 ErrInvalidToken, 实际 %v", err)
		}
	})
}
