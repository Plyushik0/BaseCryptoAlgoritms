package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

var lc = [16]byte{
	148, 32, 133, 16, 194, 192, 1, 251, 1, 192, 194, 16,
	133, 32, 148, 1,
}

var pi = [256]byte{
	252, 238, 221, 17, 207, 110, 49, 22, 251, 196, 250, 218, 35, 197, 4, 77,
	233, 119, 240, 219, 147, 46, 153, 186, 23, 54, 241, 187, 20, 205, 95, 193,
	249, 24, 101, 90, 226, 92, 239, 33, 129, 28, 60, 66, 139, 1, 142, 79,
	5, 132, 2, 174, 227, 106, 143, 160, 6, 11, 237, 152, 127, 212, 211, 31,
	235, 52, 44, 81, 234, 200, 72, 171, 242, 42, 104, 162, 253, 58, 206, 204,
	181, 112, 14, 86, 8, 12, 118, 18, 191, 114, 19, 71, 156, 183, 93, 135,
	21, 161, 150, 41, 16, 123, 154, 199, 243, 145, 120, 111, 157, 158, 178, 177,
	50, 117, 25, 61, 255, 53, 138, 126, 109, 84, 198, 128, 195, 189, 13, 87,
	223, 245, 36, 169, 62, 168, 67, 201, 215, 121, 214, 246, 124, 34, 185, 3,
	224, 15, 236, 222, 122, 148, 176, 188, 220, 232, 40, 80, 78, 51, 10, 74,
	167, 151, 96, 115, 30, 0, 98, 68, 26, 184, 56, 130, 100, 159, 38, 65,
	173, 69, 70, 146, 39, 94, 85, 47, 140, 163, 165, 125, 105, 213, 149, 59,
	7, 88, 179, 64, 134, 172, 29, 247, 48, 55, 107, 228, 136, 217, 231, 137,
	225, 27, 131, 73, 76, 63, 248, 254, 141, 83, 170, 144, 202, 216, 133, 97,
	32, 113, 103, 164, 45, 43, 9, 91, 203, 155, 37, 208, 190, 229, 108, 82,
	89, 166, 116, 210, 230, 244, 180, 192, 209, 102, 175, 194, 57, 75, 99, 182,
}

var piInv [256]byte

var cBlk [32][16]byte

type Kuznyechik struct {
	roundKeys [10][16]byte
}

func NewKuznyechik(key []byte) *Kuznyechik {
	k := &Kuznyechik{}

	var k0, k1 [16]byte
	copy(k0[:], key[:16])
	copy(k1[:], key[16:])

	k.roundKeys[0] = k0
	k.roundKeys[1] = k1

	for i := 0; i < 4; i++ {
		for j := 0; j < 8; j++ {
			f(&k0, &k1, &cBlk[8*i+j])
		}
		k.roundKeys[2+2*i] = k0
		k.roundKeys[3+2*i] = k1
	}

	return k
}

func gf(a, b byte) byte {
	p := byte(0)
	for b != 0 {
		if b&1 != 0 {
			p ^= a
		}
		aHigh := a & 0x80
		a <<= 1
		if aHigh != 0 {
			a ^= 0xC3
		}
		b >>= 1
	}
	return p
}

func xor(dst, src1, src2 *[16]byte) {
	for i := 0; i < 16; i++ {
		dst[i] = src1[i] ^ src2[i]
	}
}

func r(blk *[16]byte) {
	t := blk[15]
	for i := 14; i >= 0; i-- {
		blk[i+1] = blk[i]
		t ^= gf(blk[i], lc[i])
	}
	blk[0] = t
}

func rInv(blk *[16]byte) {
	t := blk[0]
	for i := 0; i < 15; i++ {
		blk[i] = blk[i+1]
		t ^= gf(blk[i], lc[i])
	}
	blk[15] = t
}

func l(blk *[16]byte) {
	for i := 0; i < 16; i++ {
		r(blk)
	}
}

func lInv(blk *[16]byte) {
	for i := 0; i < 16; i++ {
		rInv(blk)
	}
}

func s(blk *[16]byte) {
	for i := 0; i < 16; i++ {
		blk[i] = pi[blk[i]]
	}
}

func sInv(blk *[16]byte) {
	for i := 0; i < 16; i++ {
		blk[i] = piInv[blk[i]]
	}
}

func f(k0, k1 *[16]byte, k *[16]byte) {
	var t [16]byte
	xor(&t, k0, k)
	s(&t)
	l(&t)
	xor(&t, &t, k1)
	*k1 = *k0
	*k0 = t
}

func (k *Kuznyechik) Encrypt(block [16]byte) [16]byte {
	var state [16]byte
	copy(state[:], block[:])

	for i := 0; i < 9; i++ {
		xor(&state, &state, &k.roundKeys[i])
		s(&state)
		l(&state)
	}
	xor(&state, &state, &k.roundKeys[9])

	var out [16]byte
	copy(out[:], state[:])
	return out
}

func (k *Kuznyechik) Decrypt(block [16]byte) [16]byte {
	var state [16]byte
	copy(state[:], block[:])

	xor(&state, &state, &k.roundKeys[9])
	for i := 8; i >= 0; i-- {
		lInv(&state)
		sInv(&state)
		xor(&state, &state, &k.roundKeys[i])
	}

	var out [16]byte
	copy(out[:], state[:])
	return out
}

func init() {
	for i := 0; i < 256; i++ {
		piInv[pi[i]] = byte(i)
	}
	for i := 0; i < 32; i++ {
		cBlk[i][15] = byte(i + 1)
		l(&cBlk[i])
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
			return nil, fmt.Errorf("Неверный паддинг")
		}
	}
	return data[:length-padding], nil
}

func (k *Kuznyechik) EncryptECB(plaintext []byte) []byte {
	plaintext = pkcs7Pad(plaintext, 16)
	ciphertext := make([]byte, len(plaintext))

	for i := 0; i < len(plaintext); i += 16 {
		block := [16]byte{}
		copy(block[:], plaintext[i:i+16])
		enc := k.Encrypt(block)
		copy(ciphertext[i:], enc[:])
	}
	return ciphertext
}

func (k *Kuznyechik) DecryptECB(ciphertext []byte) ([]byte, error) {
	plain := make([]byte, len(ciphertext))

	for i := 0; i < len(ciphertext); i += 16 {
		block := [16]byte{}
		copy(block[:], ciphertext[i:i+16])
		dec := k.Decrypt(block)
		copy(plain[i:], dec[:])
	}
	return pkcs7Unpad(plain)
}

func (k *Kuznyechik) EncryptCBC(plaintext []byte, iv []byte) []byte {
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

		enc := k.Encrypt(block)
		copy(ciphertext[16+i:16+i+16], enc[:])
		prev = enc[:]
	}
	return ciphertext
}

func (k *Kuznyechik) DecryptCBC(ciphertext []byte) ([]byte, error) {

	iv := make([]byte, 16)
	copy(iv, ciphertext[:16])
	data := ciphertext[16:]

	plain := make([]byte, len(data))
	prev := iv

	for i := 0; i < len(data); i += 16 {
		block := [16]byte{}
		copy(block[:], data[i:i+16])
		d := k.Decrypt(block)
		for j := 0; j < 16; j++ {
			plain[i+j] = d[j] ^ prev[j]
		}
		prev = data[i : i+16]
	}

	return pkcs7Unpad(plain)
}

func (k *Kuznyechik) EncryptCFB(plaintext []byte, iv []byte) []byte {
	ciphertext := make([]byte, 16+len(plaintext))
	copy(ciphertext[:16], iv)

	Ci := iv
	for i := 0; i < len(plaintext); i += 16 {
		chunkSize := 16
		if i+16 > len(plaintext) {
			chunkSize = len(plaintext) - i
		}
		gamma := k.Encrypt([16]byte(Ci))
		for j := 0; j < chunkSize; j++ {
			ciphertext[16+i+j] = plaintext[i+j] ^ gamma[j]
		}
		copy(Ci, ciphertext[16+i:16+i+chunkSize])
	}
	return ciphertext
}

func (k *Kuznyechik) DecryptCFB(ciphertext []byte) []byte {
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
		gamma := k.Encrypt([16]byte(Ci))
		for j := 0; j < chunkSize; j++ {
			plain[i+j] = data[i+j] ^ gamma[j]
		}
		copy(Ci, data[i:i+chunkSize])
	}
	return plain
}

func (d *Kuznyechik) EncryptOFB(plaintext []byte, iv []byte) []byte {

	ciphertext := make([]byte, 16+len(plaintext))
	copy(ciphertext[:16], iv)

	Ci := iv
	for i := 0; i < len(plaintext); i += 16 {
		gamma := d.Encrypt([16]byte(Ci))
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

func (k *Kuznyechik) DecryptOFB(ciphertext []byte) []byte {
	return k.EncryptOFB(ciphertext[16:], ciphertext[:16])[16:]
}

func (k *Kuznyechik) EncryptCTR(plaintext []byte, iv []byte) []byte {

	ciphertext := make([]byte, len(iv)+len(plaintext))
	copy(ciphertext[:len(iv)], iv)

	counter := make([]byte, 16)
	copy(counter[:len(iv)], iv)

	pos := 0
	for pos < len(plaintext) {
		gammaBlock := k.Encrypt([16]byte(counter))

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

func (k *Kuznyechik) DecryptCTR(ciphertext []byte) []byte {
	if len(ciphertext) < 12 {
		return nil
	}
	IVLen := len(ciphertext) - (len(ciphertext)/16-1)*16
	if IVLen != 12 && IVLen != 16 {
		IVLen = 16
	}
	iv := ciphertext[:IVLen]
	return k.EncryptCTR(ciphertext[IVLen:], iv)[IVLen:]
}

func EncryptFile(k *Kuznyechik, mode, inPath, outPath string) error {
	data, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}

	var ciphertext []byte
	var iv []byte

	switch mode {
	case "ECB":
		ciphertext = k.EncryptECB(data)
	case "CBC":
		iv = make([]byte, 16)
		rand.Read(iv)
		ciphertext = k.EncryptCBC(data, iv)
	case "CFB":
		iv = make([]byte, 16)
		rand.Read(iv)
		ciphertext = k.EncryptCFB(data, iv)
	case "OFB":
		iv = make([]byte, 16)
		rand.Read(iv)
		ciphertext = k.EncryptOFB(data, iv)
	case "CTR":
		iv = make([]byte, 16)
		rand.Read(iv)
		ciphertext = k.EncryptCTR(data, iv)
	default:
		return fmt.Errorf("неизвестный режим")
	}

	return os.WriteFile(outPath, ciphertext, 0644)
}

func DecryptFile(k *Kuznyechik, mode, inPath, outPath string) error {
	data, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}

	var plaintext []byte

	switch mode {
	case "ECB":
		plaintext, err = k.DecryptECB(data)
	case "CBC":
		plaintext, err = k.DecryptCBC(data)
	case "CFB":
		plaintext = k.DecryptCFB(data)
	case "OFB":
		plaintext = k.DecryptOFB(data)
	case "CTR":
		plaintext = k.DecryptCTR(data)
	default:
		return fmt.Errorf("неизвестный режим")
	}

	if err != nil {
		return err
	}
	return os.WriteFile(outPath, plaintext, 0644)
}

func main() {
	keyHex := "8899aabbccddeeff0011223344556677fedcba98765432100123456789abcdef"
	key, _ := hex.DecodeString(keyHex)
	a := NewKuznyechik(key)

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
