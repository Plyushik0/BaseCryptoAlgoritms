package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

var sbox = [256]byte{
	99, 124, 119, 123, 242, 107, 111, 197, 48, 1, 103, 43, 254, 215, 171, 118,
	202, 130, 201, 125, 250, 89, 71, 240, 173, 212, 162, 175, 156, 164, 114, 192,
	183, 253, 147, 38, 54, 63, 247, 204, 52, 165, 229, 241, 113, 216, 49, 21,
	4, 199, 35, 195, 24, 150, 5, 154, 7, 18, 128, 226, 235, 39, 178, 117,
	9, 131, 44, 26, 27, 110, 90, 160, 82, 59, 214, 179, 41, 227, 47, 132,
	83, 209, 0, 237, 32, 252, 177, 91, 106, 203, 190, 57, 74, 76, 88, 207,
	208, 239, 170, 251, 67, 77, 51, 133, 69, 249, 2, 127, 80, 60, 159, 168,
	81, 163, 64, 143, 146, 157, 56, 245, 188, 182, 218, 33, 16, 255, 243, 210,
	205, 12, 19, 236, 95, 151, 68, 23, 196, 167, 126, 61, 100, 93, 25, 115,
	96, 129, 79, 220, 34, 42, 144, 136, 70, 238, 184, 20, 222, 94, 11, 219,
	224, 50, 58, 10, 73, 6, 36, 92, 194, 211, 172, 98, 145, 149, 228, 121,
	231, 200, 55, 109, 141, 213, 78, 169, 108, 86, 244, 234, 101, 122, 174, 8,
	186, 120, 37, 46, 28, 166, 180, 198, 232, 221, 116, 31, 75, 189, 139, 138,
	112, 62, 181, 102, 72, 3, 246, 14, 97, 53, 87, 185, 134, 193, 29, 158,
	225, 248, 152, 17, 105, 217, 142, 148, 155, 30, 135, 233, 206, 85, 40, 223,
	140, 161, 137, 13, 191, 230, 66, 104, 65, 153, 45, 15, 176, 84, 187, 22,
}

var invSbox = [256]byte{
	82, 9, 106, 213, 48, 54, 165, 56, 191, 64, 163, 158, 129, 243, 215, 251,
	124, 227, 57, 130, 155, 47, 255, 135, 52, 142, 67, 68, 196, 222, 233, 203,
	84, 123, 148, 50, 166, 194, 35, 61, 238, 76, 149, 11, 66, 250, 195, 78,
	8, 46, 161, 102, 40, 217, 36, 178, 118, 91, 162, 73, 109, 139, 209, 37,
	114, 248, 246, 100, 134, 104, 152, 22, 212, 164, 92, 204, 93, 101, 182, 146,
	108, 112, 72, 80, 253, 237, 185, 218, 94, 21, 70, 87, 167, 141, 157, 132,
	144, 216, 171, 0, 140, 188, 211, 10, 247, 228, 88, 5, 184, 179, 69, 6,
	208, 44, 30, 143, 202, 63, 15, 2, 193, 175, 189, 3, 1, 19, 138, 107,
	58, 145, 17, 65, 79, 103, 220, 234, 151, 242, 207, 206, 240, 180, 230, 115,
	150, 172, 116, 34, 231, 173, 53, 133, 226, 249, 55, 232, 28, 117, 223, 110,
	71, 241, 26, 113, 29, 41, 197, 137, 111, 183, 98, 14, 170, 24, 190, 27,
	252, 86, 62, 75, 198, 210, 121, 32, 154, 219, 192, 254, 120, 205, 90, 244,
	31, 221, 168, 51, 136, 7, 199, 49, 177, 18, 16, 89, 39, 128, 236, 95,
	96, 81, 127, 169, 25, 181, 74, 13, 45, 229, 122, 159, 147, 201, 156, 239,
	160, 224, 59, 77, 174, 42, 245, 176, 200, 235, 187, 60, 131, 83, 153, 97,
	23, 43, 4, 126, 186, 119, 214, 38, 225, 105, 20, 99, 85, 33, 12, 125,
}

var rcon = [10]uint32{
	16777216,   // 0x01000000
	33554432,   // 0x02000000
	67108864,   // 0x04000000
	134217728,  // 0x08000000
	268435456,  // 0x10000000
	536870912,  // 0x20000000
	1073741824, // 0x40000000
	2147483648, // 0x80000000
	452984832,  // 0x1b000000
	905969664,  // 0x36000000
}

type AES128 struct {
	roundKeys [11][4]uint32
}

func NewAES128(key []byte) *AES128 {
	var a AES128
	var w [44]uint32

	for i := 0; i < 4; i++ {
		w[i] = uint32(key[4*i])<<24 | uint32(key[4*i+1])<<16 |
			uint32(key[4*i+2])<<8 | uint32(key[4*i+3])
	}

	for i := 4; i < 44; i++ {
		temp := w[i-1]
		if i%4 == 0 {
			temp = subWord(rotWord(temp)) ^ rcon[(i/4)-1]
		}
		w[i] = w[i-4] ^ temp
	}

	for i := 0; i < 11; i++ {
		a.roundKeys[i] = [4]uint32{w[4*i], w[4*i+1], w[4*i+2], w[4*i+3]}
	}

	return &a
}

func (a *AES128) Encrypt(block [16]byte) [16]byte {
	var state [4][4]byte
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			state[j][i] = block[i*4+j]
		}
	}

	addRoundKey(&state, a.roundKeys[0])

	for round := 1; round < 10; round++ {
		subBytes(&state)
		shiftRows(&state)
		mixColumns(&state)
		addRoundKey(&state, a.roundKeys[round])
	}

	subBytes(&state)
	shiftRows(&state)
	addRoundKey(&state, a.roundKeys[10])

	var out [16]byte
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			out[i*4+j] = state[j][i]
		}
	}
	return out
}

func (a *AES128) Decrypt(block [16]byte) [16]byte {
	var state [4][4]byte
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			state[j][i] = block[i*4+j]
		}
	}

	addRoundKey(&state, a.roundKeys[10])

	for round := 9; round > 0; round-- {
		invShiftRows(&state)
		invSubBytes(&state)
		addRoundKey(&state, a.roundKeys[round])
		invMixColumns(&state)
	}

	invShiftRows(&state)
	invSubBytes(&state)
	addRoundKey(&state, a.roundKeys[0])

	var out [16]byte
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			out[i*4+j] = state[j][i]
		}
	}
	return out
}

func rotWord(x uint32) uint32 {
	return x<<8 | x>>24
}

func subWord(x uint32) uint32 {
	return uint32(sbox[x>>24])<<24 |
		uint32(sbox[(x>>16)&0xff])<<16 |
		uint32(sbox[(x>>8)&0xff])<<8 |
		uint32(sbox[x&0xff])
}

func subBytes(s *[4][4]byte) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			s[i][j] = sbox[s[i][j]]
		}
	}
}

func invSubBytes(s *[4][4]byte) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			s[i][j] = invSbox[s[i][j]]
		}
	}
}

func shiftRows(s *[4][4]byte) {
	s[1][0], s[1][1], s[1][2], s[1][3] = s[1][1], s[1][2], s[1][3], s[1][0]
	s[2][0], s[2][1], s[2][2], s[2][3] = s[2][2], s[2][3], s[2][0], s[2][1]
	s[3][0], s[3][1], s[3][2], s[3][3] = s[3][3], s[3][0], s[3][1], s[3][2]
}

func invShiftRows(s *[4][4]byte) {
	s[1][0], s[1][1], s[1][2], s[1][3] = s[1][3], s[1][0], s[1][1], s[1][2]
	s[2][0], s[2][1], s[2][2], s[2][3] = s[2][2], s[2][3], s[2][0], s[2][1]
	s[3][0], s[3][1], s[3][2], s[3][3] = s[3][1], s[3][2], s[3][3], s[3][0]
}

// x^8 + x^4 + x^3 + x + 1
// Число a = x + 1 (0b11).
// Число b = x + 1 (0b11).
// Проверка младшего бита числа b (0b11). Он ненулевой
// p (0b11) = p (0b00) XOR a (0b11)
// Старший бит числа a (0b11) ненулевой
// Сдвигаем число a влево
// a (0b10) = a (0b11) << 1
// Так как число a у нас переполнилось то производим взятие остатка по модулю
// a (0b01) = a (0b10) XOR неприводимый многочлен (0b1b)
// Сдвигаем число b вправо.
// b (0b01) = b (0b11) >> 1
// ...
// Повторяем пока b не станет равным нулю
// x = 0b10
func gfMul(a, b byte) byte {
	p := byte(0)
	for b != 0 {
		if b&1 != 0 {
			p ^= a
		}
		hi := a & 0x80
		a <<= 1
		if hi != 0 {
			a ^= 0x1b
		}
		b >>= 1
	}
	return p
}

// [2 3 1 1]
// [1 2 3 1]
// [1 1 2 3]
// [3 1 1 2]

func mixColumns(s *[4][4]byte) {
	for i := 0; i < 4; i++ {
		a0 := s[0][i]
		a1 := s[1][i]
		a2 := s[2][i]
		a3 := s[3][i]

		s[0][i] = gfMul(2, a0) ^ gfMul(3, a1) ^ a2 ^ a3
		s[1][i] = a0 ^ gfMul(2, a1) ^ gfMul(3, a2) ^ a3
		s[2][i] = a0 ^ a1 ^ gfMul(2, a2) ^ gfMul(3, a3)
		s[3][i] = gfMul(3, a0) ^ a1 ^ a2 ^ gfMul(2, a3)
	}
}

func invMixColumns(s *[4][4]byte) {
	for i := 0; i < 4; i++ {
		a0 := s[0][i]
		a1 := s[1][i]
		a2 := s[2][i]
		a3 := s[3][i]

		s[0][i] = gfMul(14, a0) ^ gfMul(11, a1) ^ gfMul(13, a2) ^ gfMul(9, a3)
		s[1][i] = gfMul(9, a0) ^ gfMul(14, a1) ^ gfMul(11, a2) ^ gfMul(13, a3)
		s[2][i] = gfMul(13, a0) ^ gfMul(9, a1) ^ gfMul(14, a2) ^ gfMul(11, a3)
		s[3][i] = gfMul(11, a0) ^ gfMul(13, a1) ^ gfMul(9, a2) ^ gfMul(14, a3)
	}
}

func addRoundKey(s *[4][4]byte, rk [4]uint32) {
	for i := 0; i < 4; i++ {
		w := rk[i]
		s[0][i] ^= byte(w >> 24)
		s[1][i] ^= byte(w >> 16)
		s[2][i] ^= byte(w >> 8)
		s[3][i] ^= byte(w)
	}
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
	if padding > length || padding > 16 {
		return nil, fmt.Errorf("Неверный паддинг")
	}
	for i := length - padding; i < length; i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("Неверный паддин")
		}
	}
	return data[:length-padding], nil
}

func (a *AES128) EncryptECB(plaintext []byte) []byte {
	plaintext = pkcs7Pad(plaintext, 16)
	ciphertext := make([]byte, len(plaintext))

	for i := 0; i < len(plaintext); i += 16 {
		block := [16]byte{}
		copy(block[:], plaintext[i:i+16])
		enc := a.Encrypt(block)
		copy(ciphertext[i:], enc[:])
	}
	return ciphertext
}

func (a *AES128) DecryptECB(ciphertext []byte) ([]byte, error) {
	plain := make([]byte, len(ciphertext))

	for i := 0; i < len(ciphertext); i += 16 {
		block := [16]byte{}
		copy(block[:], ciphertext[i:i+16])
		dec := a.Decrypt(block)
		copy(plain[i:], dec[:])
	}
	return pkcs7Unpad(plain)
}

func (a *AES128) EncryptCBC(plaintext []byte, iv []byte) []byte {
	plaintext = pkcs7Pad(plaintext, 16)
	ciphertext := make([]byte, len(plaintext)+16)
	copy(ciphertext[:16], iv)

	prev := iv
	for i := 0; i < len(plaintext); i += 16 {
		block := [16]byte{}
		copy(block[:], plaintext[i:i+16])

		for j := 0; j < 16; j++ {
			block[j] ^= prev[j]
		}

		enc := a.Encrypt(block)
		copy(ciphertext[16+i:16+i+16], enc[:])
		prev = enc[:]
	}
	return ciphertext
}

func (a *AES128) DecryptCBC(ciphertext []byte) ([]byte, error) {
	iv := make([]byte, 16)
	copy(iv, ciphertext[:16])
	data := ciphertext[16:]

	plain := make([]byte, len(data))
	prev := iv

	for i := 0; i < len(data); i += 16 {
		block := [16]byte{}
		copy(block[:], data[i:i+16])
		dec := a.Decrypt(block)
		for j := 0; j < 16; j++ {
			plain[i+j] = dec[j] ^ prev[j]
		}
		prev = data[i : i+16]
	}

	return pkcs7Unpad(plain)
}

func (a *AES128) EncryptCFB(plaintext []byte, iv []byte) []byte {
	ciphertext := make([]byte, 16+len(plaintext))
	copy(ciphertext[:16], iv)

	Ci := iv
	for i := 0; i < len(plaintext); i += 16 {
		chunkSize := 16
		if i+16 > len(plaintext) {
			chunkSize = len(plaintext) - i
		}
		gamma := a.Encrypt([16]byte(Ci))
		for j := 0; j < chunkSize; j++ {
			ciphertext[16+i+j] = plaintext[i+j] ^ gamma[j]
		}
		copy(Ci, ciphertext[16+i:16+i+chunkSize])
	}
	return ciphertext
}

func (a *AES128) DecryptCFB(ciphertext []byte) []byte {
	if len(ciphertext) < 16 {
		return nil
	}
	iv := ciphertext[:16]
	data := ciphertext[16:]

	plain := make([]byte, len(data))
	Ci := iv

	for i := 0; i < len(data); i += 16 {
		chunkSize := 16
		if i+16 > len(data) {
			chunkSize = len(data) - i
		}
		gamma := a.Encrypt([16]byte(Ci))
		for j := 0; j < chunkSize; j++ {
			plain[i+j] = data[i+j] ^ gamma[j]
		}
		copy(Ci, data[i:i+chunkSize])
	}
	return plain
}

func (a *AES128) EncryptOFB(plaintext []byte, iv []byte) []byte {

	ciphertext := make([]byte, 16+len(plaintext))
	copy(ciphertext[:16], iv)

	Ci := iv
	for i := 0; i < len(plaintext); i += 16 {
		gamma := a.Encrypt([16]byte(Ci))
		copy(Ci, gamma[:])

		chunkSize := 16
		if i+16 > len(plaintext) {
			chunkSize = len(plaintext) - i
		}
		for j := 0; j < chunkSize; j++ {
			ciphertext[16+i+j] = plaintext[i+j] ^ gamma[j]
		}
	}
	return ciphertext
}

func (a *AES128) DecryptOFB(ciphertext []byte) []byte {
	return a.EncryptOFB(ciphertext[16:], ciphertext[:16])[16:]
}

func (a *AES128) EncryptCTR(plaintext []byte, iv []byte) []byte {

	ciphertext := make([]byte, len(iv)+len(plaintext))
	copy(ciphertext[:len(iv)], iv)

	counter := make([]byte, 16)
	copy(counter[:len(iv)], iv)

	pos := 0
	for pos < len(plaintext) {
		gammaBlock := a.Encrypt([16]byte(counter))

		for i := 0; i < 16 && pos < len(plaintext); i++ {
			ciphertext[len(iv)+pos] = plaintext[pos] ^ gammaBlock[i]
			pos++
		}

		for i := len(iv); i < 16; i++ {
			if counter[i] < 0xFF {
				counter[i]++
				break
			}
			counter[i] = 0
			if i == 15 {
				panic("переполнение счётчика")
			}
		}
	}
	return ciphertext
}

func (a *AES128) DecryptCTR(ciphertext []byte) []byte {
	if len(ciphertext) < 12 {
		return nil
	}
	IVlen := len(ciphertext) - (len(ciphertext)/16-1)*16
	if IVlen != 12 && IVlen != 16 {
		IVlen = 16
	}
	iv := ciphertext[:IVlen]
	return a.EncryptCTR(ciphertext[IVlen:], iv)[IVlen:]
}

func EncryptFile(a *AES128, mode, inPath, outPath string) error {
	data, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}

	var ciphertext []byte
	var iv []byte

	switch mode {
	case "ECB":
		ciphertext = a.EncryptECB(data)
	case "CBC":
		iv = make([]byte, 16)
		rand.Read(iv)
		ciphertext = a.EncryptCBC(data, iv)
	case "CFB":
		iv = make([]byte, 16)
		rand.Read(iv)
		ciphertext = a.EncryptCFB(data, iv)
	case "OFB":
		iv = make([]byte, 16)
		rand.Read(iv)
		ciphertext = a.EncryptOFB(data, iv)
	case "CTR":
		iv = make([]byte, 16)
		rand.Read(iv)
		ciphertext = a.EncryptCTR(data, iv)
	default:
		return fmt.Errorf("неизвестный режим")
	}

	return os.WriteFile(outPath, ciphertext, 0644)
}

func DecryptFile(a *AES128, mode, inPath, outPath string) error {
	data, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}

	var plaintext []byte

	switch mode {
	case "ECB":
		plaintext, err = a.DecryptECB(data)
	case "CBC":
		plaintext, err = a.DecryptCBC(data)
	case "CFB":
		plaintext = a.DecryptCFB(data)
	case "OFB":
		plaintext = a.DecryptOFB(data)
	case "CTR":
		plaintext = a.DecryptCTR(data)
	default:
		return fmt.Errorf("неизвестный режим")
	}

	if err != nil {
		return err
	}
	return os.WriteFile(outPath, plaintext, 0644)
}

func main() {
	keyHex := "000102030405060708090a0b0c0d0e0f"
	key, _ := hex.DecodeString(keyHex)

	a := NewAES128(key)

	modes := []string{"ECB", "CBC", "CFB", "OFB", "CTR"}

	for _, mode := range modes {
		in := "1.png"
		enc := fmt.Sprintf("1_%s.enc", mode)
		dec := fmt.Sprintf("1_%s_dec.png", mode)

		if err := EncryptFile(a, mode, in, enc); err != nil {
			fmt.Printf("Ошибка шифрования %s: %v\n", mode, err)
			continue
		}
		if err := DecryptFile(a, mode, enc, dec); err != nil {
			fmt.Printf("Ошибка дешифрования %s: %v\n", mode, err)
		} else {
			fmt.Printf("%s зашифровано и расшифровано успешно\n", mode)
		}
	}
}

//Ci := make([]byte, 16)
//copy(Ci, iv)
