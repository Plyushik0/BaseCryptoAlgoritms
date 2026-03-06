package hashf

import (
	"bytes"
)

func HMAC(hashFunc func([]byte, int) []byte, blockSize int, key, message []byte, hashSize int) []byte {
	if len(key) > blockSize {
		key = hashFunc(key, hashSize)
	}
	if len(key) < blockSize {
		key = append(key, bytes.Repeat([]byte{0x00}, blockSize-len(key))...)
	}

	oKeyPad := make([]byte, blockSize)
	iKeyPad := make([]byte, blockSize)
	for i := 0; i < blockSize; i++ {
		oKeyPad[i] = key[i] ^ 0x5c
		iKeyPad[i] = key[i] ^ 0x36
	}

	inner := hashFunc(append(iKeyPad, message...), hashSize)
	return hashFunc(append(oKeyPad, inner...), hashSize)
}


func HMACStreebog256(key, message []byte) []byte {
	return HMAC(StreebogHash, 64, key, message, 32)
}


func HMACStreebog512(key, message []byte) []byte {
	return HMAC(StreebogHash, 64, key, message, 64)
}


func HMACSHA256(key, message []byte) []byte {
	return HMAC(
		func(msg []byte, _ int) []byte {
			h := Sha256(msg)
			return h[:]
		},
		64, key, message, 32)
}


func HMACSHA512(key, message []byte) []byte {
	return HMAC(
		func(msg []byte, _ int) []byte {
			h := Sha512(msg)
			return h[:]
		},
		128, key, message, 64)
}
