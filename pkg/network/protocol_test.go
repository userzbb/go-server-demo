package network

import "testing"

func TestEncode(t *testing.T) {
	tests := []struct {
		name  string
		msgID uint32
		body  []byte
	}{
		{name: "普通消息", msgID: 1001, body: []byte("hello game")},
		{name: "空消息体", msgID: 2001, body: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := Encode(tt.msgID, tt.body)

			// 期望包结构：4字节总长度 + 4字节消息ID + body
			expectedLen := 8 + len(tt.body)
			if len(packet) != expectedLen {
				t.Errorf("包长度错误: 期望 %d, 实际 %d", expectedLen, len(packet))
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		msgID   uint32
		body    []byte
		wantErr bool
	}{
		{name: "正常消息", data: Encode(2002, []byte("test data")), msgID: 2002, body: []byte("test data")},
		{name: "数据包过短", data: []byte{0x00, 0x00, 0x00, 0x01}, wantErr: true},
		{name: "超长包", data: Encode(1, make([]byte, MaxPacketSize+1)), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decodedID, decodedBody, err := Decode(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Fatal("期望返回错误，实际成功")
				}
				return
			}
			if err != nil {
				t.Fatalf("Decode 返回错误: %v", err)
			}
			if decodedID != tt.msgID {
				t.Errorf("消息ID不匹配: 期望 %d, 实际 %d", tt.msgID, decodedID)
			}
			if string(decodedBody) != string(tt.body) {
				t.Errorf("消息体不匹配: 期望 %q, 实际 %q", tt.body, decodedBody)
			}
		})
	}
}
