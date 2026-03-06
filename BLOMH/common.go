package main

import (
	"encoding/json"
	"fmt"
	"net"
	"bytes"

)

const (
	DEFAULT_PRIME   = 19913 
	DEFAULT_DEGREE_M = 2    
	TA_HOST         = "127.0.0.1" 
	TA_PORT         = 9999  
	BUFFER_SIZE     = 4096  
)

// значение полинома с коэффициентами coeffs в точке xVal по модулю prime
func polyEval(coeffs []int, xVal, prime int) int {
	res := 0
	xPowerI := 1
	for _, coeff := range coeffs {
		term := (coeff * xPowerI) % prime
		res = (res + term) % prime
		xPowerI = (xPowerI * xVal) % prime
	}
	return res
}


func sendJSONMessage(conn net.Conn, data interface{}) error {
	message, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("[Socket Send Error] JSON Marshal: %v", err)
	}
	message = append(message, '\n')
	_, err = conn.Write(message)
	if err != nil {
		return fmt.Errorf("[Socket Send Error] Write to socket %v: %v", conn.RemoteAddr(), err)
	}
	return nil
}


func receiveJSONMessage(conn net.Conn) (interface{}, error) {
	buffer := make([]byte, BUFFER_SIZE)
	var data []byte
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				return nil, fmt.Errorf("[Socket Receive Error] Connection closed by peer on socket %v. Buffer: %s", conn.RemoteAddr(), string(data))
			}
			return nil, fmt.Errorf("[Socket Receive Error] Read from socket %v: %v", conn.RemoteAddr(), err)
		}
		data = append(data, buffer[:n]...)
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			message := data[:i]
			var result interface{}
			if err := json.Unmarshal(message, &result); err != nil {
				return nil, fmt.Errorf("[Socket Receive Error] JSON Decode Error: %v. Received: %s", err, string(message[:min(200, len(message))]))
			}
			return result, nil
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}