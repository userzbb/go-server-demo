package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("写入测试配置失败: %v", err)
	}
	return path
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		content string
		env     map[string]string
		want    func(*Config) bool
		wantErr bool
	}{
		{
			name: "加载完整配置",
			content: "server:\n" +
				"  name: omega-test\n" +
				"  host: 0.0.0.0\n" +
				"  port: 8888\n" +
				"postgres:\n" +
				"  host: localhost\n" +
				"  port: 5432\n",
			want: func(c *Config) bool {
				return c.Server.Name == "omega-test" && c.Server.Port == 8888 && c.Postgres.Host == "localhost"
			},
		},
		{
			name:    "环境变量覆盖文件值",
			content: "server:\n  port: 8888\nlog:\n  level: debug\n",
			env:     map[string]string{"SERVER_PORT": "9999", "LOG_LEVEL": "warn"},
			want: func(c *Config) bool {
				return c.Server.Port == 9999 && c.Log.Level == "warn"
			},
		},
		{
			name:    "JWT 密钥来自环境变量",
			content: "server:\n  port: 8888\n",
			env:     map[string]string{"JWT_SECRET": "test-secret"},
			want: func(c *Config) bool {
				return c.JWTSecret == "test-secret"
			},
		},
		{
			name:    "配置文件不存在",
			path:    filepath.Join(t.TempDir(), "missing.yaml"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.path
			if path == "" {
				path = writeTempConfig(t, tt.content)
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, err := Load(path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("期望返回错误，实际成功")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load 返回错误: %v", err)
			}
			if tt.want != nil && !tt.want(cfg) {
				t.Errorf("配置解析结果不符合预期: %+v", cfg)
			}
		})
	}
}
