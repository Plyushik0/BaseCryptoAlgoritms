package cipher

import (

	"encoding/json"
	"sample-app/mylib"
	"math/big"
	"os"
	"crypto/rand"
	"bytes"
)




func SaveToFile(filename string, data interface{}) error {
	file, _ := json.MarshalIndent(data, "", "  ")
	return os.WriteFile(filename, file, 0644)
}

func LoadFromFile(filename string, data interface{}) error {
	file, _ := os.ReadFile(filename)
	return json.Unmarshal(file, data)
}


func GenerateRSAKeys(bitSize int, pubFile, privFile string) (e, d, n, fn *big.Int) {
    p := mylib.GeneratePrimeBitK(bitSize, 20)
    q := mylib.GeneratePrimeBitK(bitSize, 20)

    n = new(big.Int).Mul(p, q)
    fn = new(big.Int).Mul(new(big.Int).Sub(p, big.NewInt(1)), new(big.Int).Sub(q, big.NewInt(1)))

    for {
        e, _ = rand.Int(rand.Reader, fn)
        nod, _, _ := mylib.ExtendedGCD(e, fn)
        if e.Cmp(big.NewInt(1)) > 0 && nod.Cmp(big.NewInt(1)) == 0 {
            break
        }
    }

    d = mylib.SloveFirstDegreeComparsion(e, big.NewInt(1), fn)
    ok := map[string]interface{}{
        "SubjectPublicKeyInfo": map[string]interface{}{
            "publicExponent": e.String(),
            "N":              n.String(),
        },
        "PKCS10CertRequest":   0,
        "Certificate":         0,
        "PKCS7CertChain-PKCS": 0,
    }
    SaveToFile(pubFile, ok)

    sk := map[string]interface{}{
        "privateExponent": d.String(),
        "prime1":          new(big.Int).Sub(n, big.NewInt(1)).String(),
        "prime2":          new(big.Int).Sub(n, big.NewInt(1)).String(),
        "exponent1":       new(big.Int).Mod(d, new(big.Int).Sub(n, big.NewInt(1))).String(),
        "exponent2":       new(big.Int).Mod(d, new(big.Int).Sub(n, big.NewInt(1))).String(),
    }
    SaveToFile(privFile, sk)
    return e, d, n, fn
}

// PKCS 7
func EncryptRSA(message string, e, n *big.Int) []*big.Int {
    M := []byte(message)
    blockSize := 254

    // PKCS#7 padding
    padding := blockSize - (len(M) % blockSize)
    if padding == 0 {
        padding = blockSize
    }
    padtext := bytes.Repeat([]byte{byte(padding)}, padding)
    paddedMessage := append(M, padtext...)

    var C []*big.Int
    for i := 0; i < len(paddedMessage); i += blockSize {
        block := paddedMessage[i : i+blockSize]
        m := new(big.Int).SetBytes(block)
        c := mylib.FastPowMod(m, e, n)
        C = append(C, c)
    }
    return C
}

func DecryptRSA(C []*big.Int, d, n *big.Int) string {
    var decryptedBytes []byte
    for _, c := range C {
        m := mylib.FastPowMod(c, d, n)
        block := m.Bytes()
        // Ensure block is blockSize bytes (prepend zeros if needed)
        if len(block) < 254 {
            padded := make([]byte, 254-len(block))
            block = append(padded, block...)
        }
        decryptedBytes = append(decryptedBytes, block...)
    }
    if len(decryptedBytes) == 0 {
        return ""
    }
    padding := int(decryptedBytes[len(decryptedBytes)-1])
    if padding == 0 || padding > 254 {
        return "" // Некорректный padding
    }
    // Проверяем корректность всех паддинг-байтов
    for i := 0; i < padding; i++ {
        if decryptedBytes[len(decryptedBytes)-1-i] != byte(padding) {
            return "" // Некорректный padding
        }
    }
    return string(decryptedBytes[:len(decryptedBytes)-padding])
}