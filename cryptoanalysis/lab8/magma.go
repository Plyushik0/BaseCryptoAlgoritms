package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

var pi = [8][16]uint8{
	{12, 4, 6, 2, 10, 5, 11, 9, 14, 8, 13, 7, 0, 3, 15, 1},
	{6, 8, 2, 3, 9, 10, 5, 12, 1, 14, 4, 7, 11, 13, 0, 15},
	{11, 3, 5, 8, 2, 15, 10, 13, 14, 1, 7, 4, 12, 9, 6, 0},
	{12, 8, 2, 1, 13, 4, 15, 6, 7, 0, 10, 5, 3, 14, 9, 11},
	{7, 15, 5, 10, 8, 1, 6, 13, 0, 9, 3, 14, 11, 4, 2, 12},
	{5, 13, 15, 6, 9, 2, 12, 10, 11, 7, 8, 1, 4, 3, 14, 0},
	{8, 14, 2, 5, 6, 9, 1, 12, 15, 4, 11, 0, 13, 10, 3, 7},
	{1, 7, 14, 13, 0, 5, 8, 3, 4, 15, 10, 6, 9, 12, 11, 2},
}

type Magma struct {
	roundKeys [32]uint32
	pi        [8][16]uint8
}

func NewMagma(key []byte) *Magma {
	m := &Magma{pi: pi}

	var subkeys [8]uint32
	for i := 0; i < 8; i++ {
		offset := i * 4
		subkeys[i] = uint32(key[offset])<<24 |
			uint32(key[offset+1])<<16 |
			uint32(key[offset+2])<<8 |
			uint32(key[offset+3])
	}

	for i := 0; i < 3; i++ {
		copy(m.roundKeys[i*8:(i+1)*8], subkeys[:])
	}
	for i := 0; i < 8; i++ {
		m.roundKeys[24+i] = subkeys[7-i]
	}

	return m
}

func (m *Magma) t(value uint32) uint32 {
	var result uint32
	for i := 0; i < 8; i++ {
		part := (value >> (4 * i)) & 0xF
		sub := uint32(m.pi[i][part])
		result |= sub << (4 * i)
	}
	return result
}

func (m *Magma) g(a, k uint32) uint32 {
	temp := a + k
	sub := m.t(temp)
	return (sub << 11) | (sub >> 21)
}

func (m *Magma) Encrypt(block [8]byte) [8]byte {
	L := uint32(block[0])<<24 | uint32(block[1])<<16 | uint32(block[2])<<8 | uint32(block[3])
	R := uint32(block[4])<<24 | uint32(block[5])<<16 | uint32(block[6])<<8 | uint32(block[7])

	for i := 0; i < 32; i++ {
		temp := R
		R = L ^ m.g(R, m.roundKeys[i])
		L = temp
	}

	var out [8]byte
	out[0] = byte(R >> 24)
	out[1] = byte(R >> 16)
	out[2] = byte(R >> 8)
	out[3] = byte(R)
	out[4] = byte(L >> 24)
	out[5] = byte(L >> 16)
	out[6] = byte(L >> 8)
	out[7] = byte(L)

	return out
}

func (m *Magma) Decrypt(block [8]byte) [8]byte {
	a0 := uint32(block[0])<<24 | uint32(block[1])<<16 | uint32(block[2])<<8 | uint32(block[3])
	a1 := uint32(block[4])<<24 | uint32(block[5])<<16 | uint32(block[6])<<8 | uint32(block[7])

	for i := 31; i >= 0; i-- {
		temp := a1
		a1 = a0 ^ m.g(a1, m.roundKeys[i])
		a0 = temp
	}

	var out [8]byte
	out[0] = byte(a1 >> 24)
	out[1] = byte(a1 >> 16)
	out[2] = byte(a1 >> 8)
	out[3] = byte(a1)
	out[4] = byte(a0 >> 24)
	out[5] = byte(a0 >> 16)
	out[6] = byte(a0 >> 8)
	out[7] = byte(a0)

	return out
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("Неверная длина")
	}
	padding := int(data[length-1])
	if padding > length || padding > 8 {
		return nil, fmt.Errorf("Неверный паддинг")
	}
	for i := length - padding; i < length; i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("Неверный паддинг")
		}
	}
	return data[:length-padding], nil
}

func (m *Magma) EncryptECB(plaintext []byte) []byte {
	plaintext = pkcs7Pad(plaintext, 8)
	ciphertext := make([]byte, len(plaintext))

	for i := 0; i < len(plaintext); i += 8 {
		block := [8]byte{}
		copy(block[:], plaintext[i:i+8])
		enc := m.Encrypt(block)
		copy(ciphertext[i:], enc[:])
	}
	return ciphertext
}

func (m *Magma) DecryptECB(ciphertext []byte) ([]byte, error) {
	plain := make([]byte, len(ciphertext))

	for i := 0; i < len(ciphertext); i += 8 {
		block := [8]byte{}
		copy(block[:], ciphertext[i:i+8])
		dec := m.Decrypt(block)
		copy(plain[i:], dec[:])
	}
	return pkcs7Unpad(plain)
}

func (m *Magma) EncryptCBC(plaintext []byte, iv []byte) []byte {
	plaintext = pkcs7Pad(plaintext, 8)
	ciphertext := make([]byte, 8+len(plaintext))
	copy(ciphertext[:8], iv)

	prev := iv
	for i := 0; i < len(plaintext); i += 8 {
		block := [8]byte{}
		copy(block[:], plaintext[i:i+8])

		for j := 0; j < 8; j++ {
			block[j] ^= prev[j]
		}
		enc := m.Encrypt(block)
		copy(ciphertext[8+i:8+i+8], enc[:])
		prev = enc[:]
	}
	return ciphertext
}

func (m *Magma) DecryptCBC(ciphertext []byte) ([]byte, error) {
	iv := ciphertext[:8]
	copy(iv, ciphertext[:8])
	data := ciphertext[8:]

	plain := make([]byte, len(data))
	prev := iv

	for i := 0; i < len(data); i += 8 {
		block := [8]byte{}
		copy(block[:], data[i:i+8])
		dec := m.Decrypt(block)
		for j := 0; j < 8; j++ {
			plain[i+j] = dec[j] ^ prev[j]
		}
		prev = data[i : i+8]
	}
	return pkcs7Unpad(plain)
}

func (m *Magma) EncryptCFB(plaintext []byte, iv []byte) []byte {
	ciphertext := make([]byte, 8+len(plaintext))
	copy(ciphertext[:8], iv)

	Ci := iv
	for i := 0; i < len(plaintext); i += 8 {
		chunkSize := 8
		if i+8 > len(plaintext) {
			chunkSize = len(plaintext) - i
		}
		gamma := m.Encrypt([8]byte(Ci))
		for j := 0; j < chunkSize; j++ {
			ciphertext[8+i+j] = plaintext[i+j] ^ gamma[j]
		}
		copy(Ci, ciphertext[8+i:8+i+chunkSize])
	}
	return ciphertext
}

func (m *Magma) DecryptCFB(ciphertext []byte) []byte {
	if len(ciphertext) < 8 {
		return nil
	}
	iv := ciphertext[:8]
	data := ciphertext[8:]

	plain := make([]byte, len(data))
	Ci := iv

	for i := 0; i < len(data); i += 8 {
		chunkSize := 8
		if i+8 > len(data) {
			chunkSize = len(data) - i
		}
		gamma := m.Encrypt([8]byte(Ci))
		for j := 0; j < chunkSize; j++ {
			plain[i+j] = data[i+j] ^ gamma[j]
		}
		copy(Ci, data[i:i+chunkSize])
	}
	return plain
}

func (m *Magma) EncryptOFB(plaintext []byte, iv []byte) []byte {

	ciphertext := make([]byte, 8+len(plaintext))
	copy(ciphertext[:8], iv)

	Ci := iv
	for i := 0; i < len(plaintext); i += 8 {
		gamma := m.Encrypt([8]byte(Ci))
		copy(Ci, gamma[:])

		chunkSize := 8
		if i+8 > len(plaintext) {
			chunkSize = len(plaintext) - i
		}
		for j := 0; j < chunkSize; j++ {
			ciphertext[8+i+j] = plaintext[i+j] ^ gamma[j]
		}
	}
	return ciphertext
}

func (m *Magma) DecryptOFB(ciphertext []byte) []byte {
	return m.EncryptOFB(ciphertext[8:], ciphertext[:8])[8:]
}

func (m *Magma) EncryptCTR(plaintext []byte, iv []byte) []byte {

	ciphertext := make([]byte, len(iv)+len(plaintext))
	copy(ciphertext[:len(iv)], iv)

	counter := make([]byte, 16)
	copy(counter[:len(iv)], iv)

	pos := 0
	for pos < len(plaintext) {
		gammaBlock := m.Encrypt([8]byte(counter))

		for i := 0; i < 8 && pos < len(plaintext); i++ {
			ciphertext[len(iv)+pos] = plaintext[pos] ^ gammaBlock[i]
			pos++
		}

		for i := len(iv); i < 8; i++ {
			if counter[i] < 0xFF {
				counter[i]++
				break
			}
			counter[i] = 0
			if i == 7 {
				panic("переполнение счётчика")
			}
		}
	}
	return ciphertext
}

func (m *Magma) DecryptCTR(ciphertext []byte) []byte {
	if len(ciphertext) < 8 {
		return nil
	}
	IVLen := len(ciphertext) - (len(ciphertext)/8-1)*8
	if IVLen != 4 && IVLen != 8 {
		IVLen = 8
	}
	iv := ciphertext[:IVLen]
	return m.EncryptCTR(ciphertext[IVLen:], iv)[IVLen:]
}

func EncryptFile(m *Magma, mode, inPath, outPath string) error {
	data, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}

	var ciphertext []byte
	var iv []byte

	switch mode {
	case "ECB":
		ciphertext = m.EncryptECB(data)
	case "CBC":
		iv = make([]byte, 8)
		rand.Read(iv)
		ciphertext = m.EncryptCBC(data, iv)
	case "CFB":
		iv = make([]byte, 8)
		rand.Read(iv)
		ciphertext = m.EncryptCFB(data, iv)
	case "OFB":
		iv = make([]byte, 8)
		rand.Read(iv)
		ciphertext = m.EncryptOFB(data, iv)
	case "CTR":
		iv = make([]byte, 8)
		rand.Read(iv)
		ciphertext = m.EncryptCTR(data, iv)
	default:
		return fmt.Errorf("неизвестный режим")
	}

	return os.WriteFile(outPath, ciphertext, 0644)
}

func DecryptFile(m *Magma, mode, inPath, outPath string) error {
	data, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}

	var plaintext []byte

	switch mode {
	case "ECB":
		plaintext, err = m.DecryptECB(data)
	case "CBC":
		plaintext, err = m.DecryptCBC(data)
	case "CFB":
		plaintext = m.DecryptCFB(data)
	case "OFB":
		plaintext = m.DecryptOFB(data)
	case "CTR":
		plaintext = m.DecryptCTR(data)
	default:
		return fmt.Errorf("unknown mode")
	}

	if err != nil {
		return err
	}
	return os.WriteFile(outPath, plaintext, 0644)
}

func main() {
	keyHex := "ffeeddccbbaa99887766554433221100f0f1f2f3f4f5f6f7f8f9fafbfcfdfeff"
	key, _ := hex.DecodeString(keyHex)

	m := NewMagma(key)

	modes := []string{"ECB", "CBC", "CFB", "OFB", "CTR"}

	for _, mode := range modes {
		in := "1.png"
		enc := fmt.Sprintf("1_%s.enc", mode)
		dec := fmt.Sprintf("1_%s_decrypted.png", mode)

		if err := EncryptFile(m, mode, in, enc); err != nil {
			fmt.Printf("Ошибка шифрования %s: %v\n", mode, err)
			continue
		}
		if err := DecryptFile(m, mode, enc, dec); err != nil {
			fmt.Printf("Ошибка дешифрования %s: %v\n", mode, err)
		} else {
			fmt.Printf("%s зашифровано и расшифровано успешно\n", mode)
		}
	}
}
