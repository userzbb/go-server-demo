package logger

import "testing"

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{name: "默认配置", cfg: Config{Service: "test"}},
		{name: "JSON 格式", cfg: Config{Level: "debug", Format: "json", Service: "test"}},
		{name: "非法级别", cfg: Config{Level: "verbose"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := New(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("期望返回错误，实际成功")
				}
				return
			}
			if err != nil {
				t.Fatalf("New 返回错误: %v", err)
			}
			if l == nil {
				t.Fatal("期望返回非空 logger")
			}
			l.Infof("测试日志")
		})
	}
}
