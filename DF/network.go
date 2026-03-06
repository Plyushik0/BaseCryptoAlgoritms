package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

func SendJSON(conn net.Conn, msg interface{}) bool {
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("Ошибка сериализации JSON: %v\n", err)
		return false
	}

	length := uint32(len(data))
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, length)

	if _, err := conn.Write(lengthBuf); err != nil {
		fmt.Printf("Ошибка отправки длины сообщения: %v\n", err)
		return false
	}
	if _, err := conn.Write(data); err != nil {
		fmt.Printf("Ошибка отправки сообщения: %v\n", err)
		return false
	}
	fmt.Printf("Отправлено JSON: %s\n", string(data))
	return true
}

func ReceiveJSON(conn net.Conn) (interface{}, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(conn, lengthBuf)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("соединение закрыто: %w", err)
		}
		return nil, fmt.Errorf("ошибка чтения длины сообщения: %w", err)
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	msg := make([]byte, length)
	_, err = io.ReadFull(conn, msg)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("соединение закрыто при чтении данных: %w", err)
		}
		return nil, fmt.Errorf("ошибка чтения сообщения: %w", err)
	}

	fmt.Printf("Получено JSON: %s\n", string(msg))

	var result interface{}
	if err := json.Unmarshal(msg, &result); err != nil {
		return nil, fmt.Errorf("ошибка десериализации JSON: %w", err)
	}
	return result, nil
}