package handler

import (
	"encoding/json"
	"fmt"

	"omega-server/pkg/network"
)

// reply 将响应结构编码为 JSON 并通过会话回发
func reply(sess *network.Session, msgID uint32, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("编码响应: %w", err)
	}
	sess.SendMessage(msgID, body)
	return nil
}
