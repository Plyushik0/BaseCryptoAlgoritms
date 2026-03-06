package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
)


func sendJSON(conn net.Conn, data interface{}) error {
	serialized, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("ошибка сериализации JSON: %v", err)
	}
	if err := binary.Write(conn, binary.BigEndian, uint32(len(serialized))); err != nil {
		return fmt.Errorf("ошибка отправки длины сообщения: %v", err)
	}
	_, err = conn.Write(serialized)
	return err
}

func receiveJSON(conn net.Conn) (map[string]interface{}, error) {
	var msgLen uint32
	if err := binary.Read(conn, binary.BigEndian, &msgLen); err != nil {
		return nil, fmt.Errorf("ошибка чтения длины сообщения: %v", err)
	}
	data := make([]byte, msgLen)
	n, err := conn.Read(data)
	if err != nil || n != int(msgLen) {
		return nil, fmt.Errorf("ошибка чтения сообщения: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования JSON: %v", err)
	}
	return result, nil
}