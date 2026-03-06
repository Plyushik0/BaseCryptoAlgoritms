package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

var ip = [64]byte{
	58, 50, 42, 34, 26, 18, 10, 2,
	60, 52, 44, 36, 28, 20, 12, 4,
	62, 54, 46, 38, 30, 22, 14, 6,
	64, 56, 48, 40, 32, 24, 16, 8,
	57, 49, 41, 33, 25, 17, 9, 1,
	59, 51, 43, 35, 27, 19, 11, 3,
	61, 53, 45, 37, 29, 21, 13, 5,
	63, 55, 47, 39, 31, 23, 15, 7,
}

var ipinv = [64]byte{
	40, 8, 48, 16, 56, 24, 64, 32,
	39, 7, 47, 15, 55, 23, 63, 31,
	38, 6, 46, 14, 54, 22, 62, 30,
	37, 5, 45, 13, 53, 21, 61, 29,
	36, 4, 44, 12, 52, 20, 60, 28,
	35, 3, 43, 11, 51, 19, 59, 27,
	34, 2, 42, 10, 50, 18, 58, 26,
	33, 1, 41, 9, 49, 17, 57, 25,
}

var expansion = [48]byte{
	32, 1, 2, 3, 4, 5,
	4, 5, 6, 7, 8, 9,
	8, 9, 10, 11, 12, 13,
	12, 13, 14, 15, 16, 17,
	16, 17, 18, 19, 20, 21,
	20, 21, 22, 23, 24, 25,
	24, 25, 26, 27, 28, 29,
	28, 29, 30, 31, 32, 1,
}

var pbox = [32]byte{
	16, 7, 20, 21, 29, 12, 28, 17,
	1, 15, 23, 26, 5, 18, 31, 10,
	2, 8, 24, 14, 32, 27, 3, 9,
	19, 13, 30, 6, 22, 11, 4, 25,
}

var pc1 = [56]byte{
	57, 49, 41, 33, 25, 17, 9,
	1, 58, 50, 42, 34, 26, 18,
	10, 2, 59, 51, 43, 35, 27,
	19, 11, 3, 60, 52, 44, 36,
	63, 55, 47, 39, 31, 23, 15,
	7, 62, 54, 46, 38, 30, 22,
	14, 6, 61, 53, 45, 37, 29,
	21, 13, 5, 28, 20, 12, 4,
}

var pc2 = [48]byte{
	14, 17, 11, 24, 1, 5,
	3, 28, 15, 6, 21, 10,
	23, 19, 12, 4, 26, 8,
	16, 7, 27, 20, 13, 2,
	41, 52, 31, 37, 47, 55,
	30, 40, 51, 45, 33, 48,
	44, 49, 39, 56, 34, 53,
	46, 42, 50, 36, 29, 32,
}

var rot = [16]byte{1, 1, 2, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 2, 2, 1}

var sboxes = [8][4][16]byte{
	{{14, 4, 13, 1, 2, 15, 11, 8, 3, 10, 6, 12, 5, 9, 0, 7},
		{0, 15, 7, 4, 14, 2, 13, 1, 10, 6, 12, 11, 9, 5, 3, 8},
		{4, 1, 14, 8, 13, 6, 2, 11, 15, 12, 9, 7, 3, 10, 5, 0},
		{15, 12, 8, 2, 4, 9, 1, 7, 5, 11, 3, 14, 10, 0, 6, 13}},
	{{15, 1, 8, 14, 6, 11, 3, 4, 9, 7, 2, 13, 12, 0, 5, 10},
		{3, 13, 4, 7, 15, 2, 8, 14, 12, 0, 1, 10, 6, 9, 11, 5},
		{0, 14, 7, 11, 10, 4, 13, 1, 5, 8, 12, 6, 9, 3, 2, 15},
		{13, 8, 10, 1, 3, 15, 4, 2, 11, 6, 7, 12, 0, 5, 14, 9}},
	{{10, 0, 9, 14, 6, 3, 15, 5, 1, 13, 12, 7, 11, 4, 2, 8},
		{13, 7, 0, 9, 3, 4, 6, 10, 2, 8, 5, 14, 12, 11, 15, 1},
		{13, 6, 4, 9, 8, 15, 3, 0, 11, 1, 2, 12, 5, 10, 14, 7},
		{1, 10, 13, 0, 6, 9, 8, 7, 4, 15, 14, 3, 11, 5, 2, 12}},
	{{7, 13, 14, 3, 0, 6, 9, 10, 1, 2, 8, 5, 11, 12, 4, 15},
		{13, 8, 11, 5, 6, 15, 0, 3, 4, 7, 2, 12, 1, 10, 14, 9},
		{10, 6, 9, 0, 12, 11, 7, 13, 15, 1, 3, 14, 5, 2, 8, 4},
		{3, 15, 0, 6, 10, 1, 13, 8, 9, 4, 5, 11, 12, 7, 2, 14}},
	{{2, 12, 4, 1, 7, 10, 11, 6, 8, 5, 3, 15, 13, 0, 14, 9},
		{14, 11, 2, 12, 4, 7, 13, 1, 5, 0, 15, 10, 3, 9, 8, 6},
		{4, 2, 1, 11, 10, 13, 7, 8, 15, 9, 12, 5, 6, 3, 0, 14},
		{11, 8, 12, 7, 1, 14, 2, 13, 6, 15, 0, 9, 10, 4, 5, 3}},
	{{12, 1, 10, 15, 9, 2, 6, 8, 0, 13, 3, 4, 14, 7, 5, 11},
		{10, 15, 4, 2, 7, 12, 9, 5, 6, 1, 13, 14, 0, 11, 3, 8},
		{9, 14, 15, 5, 2, 8, 12, 3, 7, 0, 4, 10, 1, 13, 11, 6},
		{4, 3, 2, 12, 9, 5, 15, 10, 11, 14, 1, 7, 6, 0, 8, 13}},
	{{4, 11, 2, 14, 15, 0, 8, 13, 3, 12, 9, 7, 5, 10, 6, 1},
		{13, 0, 11, 7, 4, 9, 1, 10, 14, 3, 5, 12, 2, 15, 8, 6},
		{1, 4, 11, 13, 12, 3, 7, 14, 10, 15, 6, 8, 0, 5, 9, 2},
		{6, 11, 13, 8, 1, 4, 10, 7, 9, 5, 0, 15, 14, 2, 3, 12}},
	{{13, 2, 8, 4, 6, 15, 11, 1, 10, 9, 3, 14, 5, 0, 12, 7},
		{1, 15, 13, 8, 10, 3, 7, 4, 12, 5, 6, 11, 0, 14, 9, 2},
		{7, 11, 4, 1, 9, 12, 14, 2, 0, 6, 10, 13, 15, 3, 5, 8},
		{2, 1, 14, 7, 4, 10, 8, 13, 15, 12, 9, 0, 3, 5, 6, 11}},
}

type DES struct {
	subkeys [16][]byte
}

func NewDES(key []byte) *DES {
	d := &DES{}
	d.subkeys = generateSubkeys(key)
	return d
}

func (d *DES) Encrypt(block [8]byte) [8]byte {
	lr := permute(block[:], ip[:])
	L := lr[:4]
	R := lr[4:]

	for i := 0; i < 16; i++ {
		fOut := f(R, d.subkeys[i])
		newR := Xor(L, fOut)
		L = R
		R = newR
	}

	rl := append(R, L...)
	cipher := permute(rl, ipinv[:])

	var out [8]byte
	copy(out[:], cipher)
	return out
}

func (d *DES) Decrypt(block [8]byte) [8]byte {
	lr := permute(block[:], ip[:])
	L := lr[:4]
	R := lr[4:]

	for i := 15; i >= 0; i-- {
		fOut := f(R, d.subkeys[i])
		newR := Xor(L, fOut)
		L = R
		R = newR
	}

	rl := append(R, L...)
	plain := permute(rl, ipinv[:])

	var out [8]byte
	copy(out[:], plain)
	return out
}
func getBit(block []byte, pos int) byte {
	byteIdx := (pos - 1) / 8
	bitIdx := 7 - ((pos - 1) % 8)
	return (block[byteIdx] >> bitIdx) & 1
}

func setBit(out []byte, pos int, value byte) {
	byteIdx := (pos - 1) / 8
	bitIdx := 7 - ((pos - 1) % 8)
	if value != 0 {
		out[byteIdx] |= 1 << bitIdx
	}
}

func permute(input []byte, table []byte) []byte {
	bitCount := len(table)
	byteCount := (bitCount + 7) / 8
	out := make([]byte, byteCount)
	for i := 0; i < bitCount; i++ {
		pos := int(table[i])
		bit := getBit(input, pos)
		byteIdx := i / 8
		bitIdx := 7 - (i % 8)
		if bit == 1 {
			out[byteIdx] |= 1 << bitIdx
		}
	}
	return out
}

func Xor(a, b []byte) []byte {
	n := len(a)
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = a[i] ^ b[i]
	}
	return out
}

func generateSubkeys(key []byte) [16][]byte {
	key56 := permute(key, pc1[:]) // 56

	var c, d uint32
	for i := 0; i < 28; i++ {
		c = (c << 1) | uint32(getBit(key56, i+1))
	}
	for i := 0; i < 28; i++ {
		d = (d << 1) | uint32(getBit(key56, i+29))
	}

	var subkeys [16][]byte

	for i := 0; i < 16; i++ {
		c = ((c << uint32(rot[i])) | (c >> (28 - uint32(rot[i])))) & 0x0FFFFFFF
		d = ((d << uint32(rot[i])) | (d >> (28 - uint32(rot[i])))) & 0x0FFFFFFF

		combined := make([]byte, 7)
		for j := 0; j < 28; j++ {
			setBit(combined, j+1, byte((c>>(27-j))&1))
		}
		for j := 0; j < 28; j++ {
			setBit(combined, j+29, byte((d>>(27-j))&1))
		}

		subkeys[i] = permute(combined, pc2[:]) // 48
	}

	return subkeys
}

func f(right []byte, subkey []byte) []byte {
	expanded := permute(right, expansion[:]) // 48
	xored := Xor(expanded, subkey)

	output := make([]byte, 4)
	for i := 0; i < 8; i++ {
		blockStart := i * 6
		b1 := getBit(xored, blockStart+1)
		b6 := getBit(xored, blockStart+6)
		row := int(b1<<1 | b6)

		col := int(getBit(xored, blockStart+2)<<3 |
			getBit(xored, blockStart+3)<<2 |
			getBit(xored, blockStart+4)<<1 |
			getBit(xored, blockStart+5))

		val := sboxes[i][row][col]

		for j := 0; j < 4; j++ {
			bit := (val >> (3 - j)) & 1
			setBit(output, i*4+j+1, byte(bit))
		}
	}

	return permute(output, pbox[:]) // 32
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
			return nil, fmt.Errorf("Неверный паддин")
		}
	}
	return data[:length-padding], nil
}

func (d *DES) EncryptECB(plaintext []byte) []byte {
	plaintext = pkcs7Pad(plaintext, 8)
	ciphertext := make([]byte, len(plaintext))

	for i := 0; i < len(plaintext); i += 8 {
		block := [8]byte{}
		copy(block[:], plaintext[i:i+8])
		enc := d.Encrypt(block)
		copy(ciphertext[i:], enc[:])
	}
	return ciphertext
}

func (d *DES) DecryptECB(ciphertext []byte) ([]byte, error) {
	plain := make([]byte, len(ciphertext))

	for i := 0; i < len(ciphertext); i += 8 {
		block := [8]byte{}
		copy(block[:], ciphertext[i:i+8])
		dec := d.Decrypt(block)
		copy(plain[i:], dec[:])
	}
	return pkcs7Unpad(plain)
}

func (d *DES) EncryptCBC(plaintext []byte, iv []byte) []byte {
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
		enc := d.Encrypt(block)
		copy(ciphertext[8+i:8+i+8], enc[:])
		prev = enc[:]
	}
	return ciphertext
}

func (d *DES) DecryptCBC(ciphertext []byte) ([]byte, error) {
	iv := ciphertext[:8]
	copy(iv, ciphertext[:8])
	data := ciphertext[8:]

	plain := make([]byte, len(data))
	prev := iv

	for i := 0; i < len(data); i += 8 {
		block := [8]byte{}
		copy(block[:], data[i:i+8])
		dec := d.Decrypt(block)
		for j := 0; j < 8; j++ {
			plain[i+j] = dec[j] ^ prev[j]
		}
		prev = data[i : i+8]
	}
	return pkcs7Unpad(plain)
}

func (d *DES) EncryptCFB(plaintext []byte, iv []byte) []byte {
	ciphertext := make([]byte, 8+len(plaintext))
	copy(ciphertext[:8], iv)

	Ci := iv
	for i := 0; i < len(plaintext); i += 8 {
		chunkSize := 8
		if i+8 > len(plaintext) {
			chunkSize = len(plaintext) - i
		}
		gamma := d.Encrypt([8]byte(Ci))
		for j := 0; j < chunkSize; j++ {
			ciphertext[8+i+j] = plaintext[i+j] ^ gamma[j]
		}
		copy(Ci, ciphertext[8+i:8+i+chunkSize])
	}
	return ciphertext
}

func (d *DES) DecryptCFB(ciphertext []byte) []byte {
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
		gamma := d.Encrypt([8]byte(Ci))
		for j := 0; j < chunkSize; j++ {
			plain[i+j] = data[i+j] ^ gamma[j]
		}
		copy(Ci, data[i:i+chunkSize])
	}
	return plain
}

func (d *DES) EncryptOFB(plaintext []byte, iv []byte) []byte {

	ciphertext := make([]byte, 8+len(plaintext))
	copy(ciphertext[:8], iv)

	Ci := iv
	for i := 0; i < len(plaintext); i += 8 {
		gamma := d.Encrypt([8]byte(Ci))
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

func (d *DES) DecryptOFB(ciphertext []byte) []byte {
	return d.EncryptOFB(ciphertext[8:], ciphertext[:8])[8:]
}

func (d *DES) EncryptCTR(plaintext []byte, iv []byte) []byte {

	ciphertext := make([]byte, len(iv)+len(plaintext))
	copy(ciphertext[:len(iv)], iv)

	counter := make([]byte, 16)
	copy(counter[:len(iv)], iv)

	pos := 0
	for pos < len(plaintext) {
		gammaBlock := d.Encrypt([8]byte(counter))

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

func (d *DES) DecryptCTR(ciphertext []byte) []byte {
	if len(ciphertext) < 8 {
		return nil
	}
	IVLen := len(ciphertext) - (len(ciphertext)/8-1)*8
	if IVLen != 4 && IVLen != 8 {
		IVLen = 8
	}
	iv := ciphertext[:IVLen]
	return d.EncryptCTR(ciphertext[IVLen:], iv)[IVLen:]
}

func EncryptFile(d *DES, mode, inPath, outPath string) error {
	data, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}

	var ciphertext []byte
	var iv []byte

	switch mode {
	case "ECB":
		ciphertext = d.EncryptECB(data)
	case "CBC":
		iv = make([]byte, 8)
		rand.Read(iv)
		ciphertext = d.EncryptCBC(data, iv)
	case "CFB":
		iv = make([]byte, 8)
		rand.Read(iv)
		ciphertext = d.EncryptCFB(data, iv)
	case "OFB":
		iv = make([]byte, 8)
		rand.Read(iv)
		ciphertext = d.EncryptOFB(data, iv)
	case "CTR":
		iv = make([]byte, 8)
		rand.Read(iv)
		ciphertext = d.EncryptCTR(data, iv)
	default:
		return fmt.Errorf("неизвестный режим")
	}

	return os.WriteFile(outPath, ciphertext, 0644)
}

func DecryptFile(d *DES, mode, inPath, outPath string) error {
	data, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}

	var plaintext []byte

	switch mode {
	case "ECB":
		plaintext, err = d.DecryptECB(data)
	case "CBC":
		plaintext, err = d.DecryptCBC(data)
	case "CFB":
		plaintext = d.DecryptCFB(data)
	case "OFB":
		plaintext = d.DecryptOFB(data)
	case "CTR":
		plaintext = d.DecryptCTR(data)
	default:
		return fmt.Errorf("неизвестный режим")
	}

	if err != nil {
		return err
	}
	return os.WriteFile(outPath, plaintext, 0644)
}
func main() {
	keyHex := "ffeeddccbbaa9988"
	key, _ := hex.DecodeString(keyHex)

	d := NewDES(key)

	modes := []string{"ECB", "CBC", "CFB", "OFB", "CTR"}

	for _, mode := range modes {
		in := "1.png"
		enc := fmt.Sprintf("1_%s.enc", mode)
		dec := fmt.Sprintf("1_%s_dec.png", mode)

		if err := EncryptFile(d, mode, in, enc); err != nil {
			fmt.Printf("Ошибка шифрования %s: %v\n", mode, err)
			continue
		}
		if err := DecryptFile(d, mode, enc, dec); err != nil {
			fmt.Printf("Ошибка дешифрования %s: %v\n", mode, err)
		} else {
			fmt.Printf("%s зашифровано и расшифровано успешно\n", mode)
		}
	}
}
