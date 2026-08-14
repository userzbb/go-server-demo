package network

import (
    "encoding/binary"
    "errors"
)

// Encode 打包消息：4字节总长度 + 4字节消息ID + body
func Encode(msgID uint32, body []byte) []byte {
    totalLen := 4 + 4 + len(body)
    buf := make([]byte, totalLen)
    
    binary.BigEndian.PutUint32(buf[0:4], uint32(totalLen))
    binary.BigEndian.PutUint32(buf[4:8], msgID)
    copy(buf[8:], body)
    
    return buf
}

// Decode 拆包：从完整包数据中解析出消息ID和body
// 参数data必须是完整的一个包（包含长度前缀）
// 返回: msgID, body, error
func Decode(data []byte) (uint32, []byte, error) {
    if len(data) < 8 {
        return 0, nil, errors.New("数据包太短，至少需要8字节")
    }
    
    totalLen := binary.BigEndian.Uint32(data[0:4])
    if int(totalLen) != len(data) {
        return 0, nil, errors.New("包长度不匹配")
    }
    
    msgID := binary.BigEndian.Uint32(data[4:8])
    body := data[8:]
    
    return msgID, body, nil
}
