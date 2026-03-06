package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"sample-app/hashf"
	"sample-app/mylib"
	"time"
)


type TimeStampRequest struct {
	Hash    []byte `json:"hash"`
	Gamma   []byte `json:"gamma"` // γ
	Delta   []byte `json:"delta"` // δ
	HashAlg string `json:"hash_alg"`
}

type TimeStampResponse struct {
	Timestamp   int64  `json:"timestamp"`
	ServerGamma []byte `json:"server_gamma"` // γ центра
	ServerDelta []byte `json:"server_delta"` // δ центра
	ServerCert  string `json:"server_cert"`
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
	Gamma                     []byte             `json:"gamma,omitempty"`
	Delta                     []byte             `json:"delta,omitempty"`
	ServerGamma               []byte             `json:"server_gamma,omitempty"`
	ServerDelta               []byte             `json:"server_delta,omitempty"`
	TimeStamp                 int64              `json:"timestamp,omitempty"`
	ServerCert                string             `json:"server_cert,omitempty"`
}


func signDataForTimeStampElGamal(
	data []byte,
	p, alpha, a *big.Int,
	hashAlg string,
) (hash []byte, gamma, delta []byte, usedHashAlg string, err error) {
	var m *big.Int
	switch hashAlg {
	case "sha256":
		h := hashf.Sha256(data)
		hash = h[:]
		usedHashAlg = "SHA256"
		m = new(big.Int).SetBytes(hash)
	case "streebog256":
		hash = hashf.StreebogHash(data, 32)
		usedHashAlg = "GOST256"
		m = new(big.Int).SetBytes(hash)
	default:
		return nil, nil, nil, "", fmt.Errorf("unknown hash algorithm")
	}
	
	gammaVal, deltaVal, err := elgamalSign(m, p, alpha, a)
	if err != nil {
		return nil, nil, nil, "", err
	}
	return hash, gammaVal.Bytes(), deltaVal.Bytes(), usedHashAlg, nil
}


func verifySignatureElGamal(hash, gammaBytes, deltaBytes []byte, p, alpha, beta *big.Int) bool {
	m := new(big.Int).SetBytes(hash)
	gamma := new(big.Int).SetBytes(gammaBytes)
	delta := new(big.Int).SetBytes(deltaBytes)
	return elgamalVerify(m, gamma, delta, p, alpha, beta)
}


func serverSignElGamal(hash []byte, hashAlg string, p, alpha, a *big.Int) (serverGamma, serverDelta []byte, timestamp int64, err error) {
	timestamp = time.Now().Unix()
	dataToSign := append(hash, []byte(time.Unix(timestamp, 0).UTC().Format("060102150405Z"))...)

	var m *big.Int
	switch hashAlg {
	case "sha256", "SHA256":
		h := hashf.Sha256(dataToSign)
		m = new(big.Int).SetBytes(h[:])
	case "streebog256", "GOST256":
		m = new(big.Int).SetBytes(hashf.StreebogHash(dataToSign, 32))
	default:
		h := hashf.Sha256(dataToSign)
		m = new(big.Int).SetBytes(h[:])
	}
	gamma, delta, err := elgamalSign(m, p, alpha, a)
	if err != nil {
		return nil, nil, 0, err
	}
	return gamma.Bytes(), delta.Bytes(), timestamp, nil
}


func verifyCenterSignatureElGamal(hash []byte, timestamp int64, serverGamma, serverDelta []byte, hashAlg string, p, alpha, beta *big.Int) bool {
	dataToCheck := append(hash, []byte(time.Unix(timestamp, 0).UTC().Format("060102150405Z"))...)

	var m *big.Int
	switch hashAlg {
	case "sha256", "SHA256":
		h := hashf.Sha256(dataToCheck)
		m = new(big.Int).SetBytes(h[:])
	case "streebog256", "GOST256":
		m = new(big.Int).SetBytes(hashf.StreebogHash(dataToCheck, 32))
	default:
		h := hashf.Sha256(dataToCheck)
		m = new(big.Int).SetBytes(h[:])
	}
	gamma := new(big.Int).SetBytes(serverGamma)
	delta := new(big.Int).SetBytes(serverDelta)
	return elgamalVerify(m, gamma, delta, p, alpha, beta)
}


func elgamalSign(m, p, alpha, a *big.Int) (*big.Int, *big.Int, error) {
	one := big.NewInt(1)
	pm1 := new(big.Int).Sub(p, one)
	var r, gamma, delta *big.Int
	var err error

	for {
		r, err = randCoprime(pm1)
		if err != nil {
			return nil, nil, err
		}
		gamma = mylib.FastPowMod(alpha, r, p)
		gcd, rInv, _ := mylib.ExtendedGCD(r, pm1)
		if gcd.Cmp(big.NewInt(1)) != 0 {
    		continue 
		}
		rInv.Mod(rInv, pm1)
		if rInv == nil {
			continue
		}
		delta = new(big.Int).Mul(a, gamma)
		delta.Sub(m, delta)
		delta.Mul(delta, rInv)
		delta.Mod(delta, pm1)
		if delta.Sign() != 0 {
			break
		}
	}
	return gamma, delta, nil
}


func elgamalVerify(m, gamma, delta, p, alpha, beta *big.Int) bool {
	one := big.NewInt(1)
	pm1 := new(big.Int).Sub(p, one)
	if gamma.Cmp(one) < 0 || gamma.Cmp(pm1) > 0 {
		return false
	}
	left := mylib.FastPowMod(beta, gamma, p)
	left.Mul(left, mylib.FastPowMod(gamma, delta, p))
	left.Mod(left, p)
	right := mylib.FastPowMod(alpha, m, p)
	return left.Cmp(right) == 0
}

func randCoprime(n *big.Int) (*big.Int, error) {
	one := big.NewInt(1)
    for {
        r, err := randInt(one, n)
        if err != nil {
            return nil, err
        }
        gcd, _, _ := mylib.ExtendedGCD(r, n)
        if gcd.Cmp(one) == 0 {
            return r, nil
        }
    }
}

func randInt(min, max *big.Int) (*big.Int, error) {
	diff := new(big.Int).Sub(max, min)
	b := diff.Bytes()
	if len(b) == 0 {
		b = []byte{0}
	}
	randBytes := make([]byte, len(b))
	_, err := rand.Read(randBytes)
	if err != nil {
		return nil, err
	}
	r := new(big.Int).SetBytes(randBytes)
	r.Mod(r, diff)
	r.Add(r, min)
	return r, nil
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
