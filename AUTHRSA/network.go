package main

import (
    "encoding/binary"
    "encoding/json"
    "net"
    "io"
	"fmt"
)

func SendJSON(conn net.Conn, msg interface{}) bool {
    data, err := json.Marshal(msg)
    if err != nil {
        println("Ошибка сериализации JSON:", err.Error())
        return false
    }

    length := uint32(len(data))
    lengthBuf := make([]byte, 4)
    binary.BigEndian.PutUint32(lengthBuf, length)

    if _, err := conn.Write(lengthBuf); err != nil {
        println("Ошибка отправки длины сообщения:", err.Error())
        return false
    }
    if _, err := conn.Write(data); err != nil {
        println("Ошибка отправки сообщения:", err.Error())
        return false
    }
    return true
}

func ReceiveJSON(conn net.Conn, v interface{}) error {
    lengthBuf := make([]byte, 4)
    _, err := io.ReadFull(conn, lengthBuf)
    if err != nil {
        if err == io.EOF {
            return fmt.Errorf("соединение закрыто: %w", err)
        }
        return fmt.Errorf("ошибка чтения длины сообщения: %w", err)
    }
    length := binary.BigEndian.Uint32(lengthBuf)

    msg := make([]byte, length)
    _, err = io.ReadFull(conn, msg)
    if err != nil {
        if err == io.EOF {
            return fmt.Errorf("соединение закрыто при чтении данных: %w", err)
        }
        return fmt.Errorf("ошибка чтения сообщения: %w", err)
    }

    if err := json.Unmarshal(msg, v); err != nil {
        return fmt.Errorf("ошибка десериализации JSON: %w", err)
    }
    return nil
}