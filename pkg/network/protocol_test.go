package network

import "testing"

func TestEncode(t *testing.T) {
	msgID := uint32(1001)
	body := []byte("hello game")

	packet := Encode(msgID, body)

	// 期望包结构：4字节总长度 + 4字节消息ID + body
	// 总长度 = 4 + 4 + len(body) = 4 + 4 + 10 = 18
	expectedLen := 4 + 4 + len(body)
	if len(packet) != expectedLen {
		t.Errorf("包长度错误: 期望 %d, 实际 %d", expectedLen, len(packet))
	}
}

func TestDecode(t *testing.T) {
	msgID := uint32(2002)
	body := []byte("test data")
	packet := Encode(msgID, body)

	// 目前 Decode 还未实现，我们只验证它能否返回正确值
	// 先写调用，后面实现后再验证
	decodedID, decodedBody, err := Decode(packet)
	if err != nil {
		t.Fatalf("Decode 返回错误: %v", err)
	}
	if decodedID != msgID {
		t.Errorf("消息ID不匹配: 期望 %d, 实际 %d", msgID, decodedID)
	}
	if string(decodedBody) != string(body) {
		t.Errorf("消息体不匹配: 期望 %s, 实际 %s", body, decodedBody)
	}
}
