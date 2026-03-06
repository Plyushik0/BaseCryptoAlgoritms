package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"sample-app/hashf"
	"sample-app/mylib"
)


type TimeStampRequest struct {
	Hash    []byte   `json:"hash"`
	S       []byte   `json:"s"` // битовая строка s
	T       *big.Int `json:"t"` // t = r * prod(a_i^s_i) mod n
	U       *big.Int `json:"u"` // u = r^2 mod n
	HashAlg string   `json:"hash_alg"`
}

type TimeStampResponse struct {
	Timestamp  int64    `json:"timestamp"`
	ServerS    []byte   `json:"server_s"`
	ServerT    *big.Int `json:"server_t"`
	ServerU    *big.Int `json:"server_u"`
	ServerCert string   `json:"server_cert"`
}


type UnsignedAttributes struct {
	ObjectIdentifier string   `json:"object_identifier"`
	Attributes       []string `json:"attributes"`
}

type SignerInfo struct {
	CMSVersion                int    `json:"cms_version"`
	SignerIdentifier          string `json:"signer_identifier"`
	DigestAlgorithmIdentifier string `json:"digest_algorithm_identifier"`
	SignatureAlgorithm        string `json:"signature_algorithm"`
	SignatureValue            string `json:"signature_value"`
	TimeStamp                 string `json:"timestamp"`
}

type PKCS7Signature struct {
	CMSVersion                int                `json:"cms_version"`
	DigestAlgorithmIdentifier string             `json:"digest_algorithm_identifier"`
	SignerInfos               SignerInfo         `json:"signer_infos"`
	UnsignedAttributes        UnsignedAttributes `json:"unsigned_attributes"`
	Hash                      []byte             `json:"hash,omitempty"`
	S                         []byte             `json:"s,omitempty"`
	T                         *big.Int           `json:"t,omitempty"`
	U                         *big.Int           `json:"u,omitempty"`
	ServerS                   []byte             `json:"server_s,omitempty"`
	ServerT                   *big.Int           `json:"server_t,omitempty"`
	ServerU                   *big.Int           `json:"server_u,omitempty"`
	TimeStamp                 int64              `json:"timestamp,omitempty"`
	ServerCert                string             `json:"server_cert,omitempty"`
}





func LoadFiatShamirPubKey(path string) (B []*big.Int, n *big.Int, err error) {
	var file map[string]interface{}
	if err = loadFromFile(path, &file); err != nil {
		return
	}
	bArr := file["b"].([]interface{})
	B = make([]*big.Int, len(bArr))
	for i, v := range bArr {
		B[i] = new(big.Int)
		B[i].SetString(v.(string), 10)
	}
	n = new(big.Int)
	n.SetString(file["n"].(string), 10)
	return
}


func LoadFiatShamirPrivKey(path string) (A []*big.Int, p, q, n *big.Int, err error) {
	var file map[string]interface{}
	if err = loadFromFile(path, &file); err != nil {
		return
	}
	aArr := file["a"].([]interface{})
	A = make([]*big.Int, len(aArr))
	for i, v := range aArr {
		A[i] = new(big.Int)
		A[i].SetString(v.(string), 10)
	}
	p = new(big.Int)
	p.SetString(file["p"].(string), 10)
	q = new(big.Int)
	q.SetString(file["q"].(string), 10)
	n = new(big.Int)
	n.SetString(file["n"].(string), 10)
	return
}


func signDataForTimeStampFiat(
	data []byte,
	A []*big.Int,
	n *big.Int,
	hashAlg string,
) (hash []byte, s []byte, t *big.Int, u *big.Int, usedHashAlg string, err error) {
	m := len(A)

	var r *big.Int
	for {
		r, _ = rand.Int(rand.Reader, n)
		if r.Cmp(big.NewInt(1)) > 0 && r.Cmp(new(big.Int).Sub(n, big.NewInt(1))) < 0 {
			break
		}
	}

	u = mylib.FastPowMod(r, big.NewInt(2), n)

	var h []byte
	switch hashAlg {
	case "sha256", "SHA256":
		tmp := hashf.Sha256(append(data, u.Bytes()...))
		h = tmp[:]
		usedHashAlg = "SHA256"
	case "streebog256", "GOST256":
		h = hashf.StreebogHash(append(data, u.Bytes()...), 32)
		usedHashAlg = "GOST256"
	default:
		err = fmt.Errorf("unknown hash algorithm")
		return
	}
	hash = h
	sBits := make([]uint8, m)
	for i := 0; i < m; i++ {
		sBits[i] = (h[i/8] >> (uint(i) % 8)) & 1
	}
	s = h[:(m+7)/8]

	t = new(big.Int).Set(r)
	for i := 0; i < m; i++ {
		if sBits[i] == 1 {
			t.Mul(t, A[i])
			t.Mod(t, n)
		}
	}

	return hash, s, t, u, usedHashAlg, nil
}
func equalBits(a, b []byte, m int) bool {
    for i := 0; i < m; i++ {
        bitA := (a[i/8] >> (uint(i) % 8)) & 1
        bitB := (b[i/8] >> (uint(i) % 8)) & 1
        if bitA != bitB {
            return false
        }
    }
    return true
}

func verifySignatureFiat(
	data []byte,
	s []byte,
	t *big.Int,
	u *big.Int,
	B []*big.Int,
	n *big.Int,
	hashAlg string,
) bool {
	m := len(B)

	w := mylib.FastPowMod(t, big.NewInt(2), n)
	for i := 0; i < m; i++ {
		sBit := (s[i/8] >> (uint(i) % 8)) & 1
		if sBit == 1 {
			w.Mul(w, B[i])
			w.Mod(w, n)
		}
	}

	var h []byte
	switch hashAlg {
	case "sha256", "SHA256":
		tmp := hashf.Sha256(data)
		h = tmp[:]
	case "streebog256", "GOST256":
		h = hashf.StreebogHash(data, 32)
	default:
		return false
	}
	sPrime := h[:(m+7)/8]
	fmt.Printf("w (server) = %x\n", w.Bytes())
	fmt.Printf("h (server) = %x\n", h)
	fmt.Printf("sPrime (server) = %x\n", sPrime)

	return equalBits(s, sPrime, m)
}



func loadFromFile(filename string, data interface{}) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, data)
}
func serializeTimeStampRequest(req *TimeStampRequest) ([]byte, error) {
	return json.Marshal(req)
}
func deserializeTimeStampRequest(data []byte) (*TimeStampRequest, error) {
	var req TimeStampRequest
	err := json.Unmarshal(data, &req)
	return &req, err
}
func serializeTimeStampResponse(resp *TimeStampResponse) ([]byte, error) {
	return json.Marshal(resp)
}
func deserializeTimeStampResponse(data []byte) (*TimeStampResponse, error) {
	var resp TimeStampResponse
	err := json.Unmarshal(data, &resp)
	return &resp, err
}
func serializePKCS7Signature(sig *PKCS7Signature) ([]byte, error) {
	return json.Marshal(sig)
}
func deserializePKCS7Signature(data []byte) (*PKCS7Signature, error) {
	var sig PKCS7Signature
	err := json.Unmarshal(data, &sig)
	return &sig, err
}
