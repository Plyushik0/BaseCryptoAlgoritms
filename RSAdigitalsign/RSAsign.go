package main

import (
    "encoding/json"
    "fmt"
    "math/big"

    "sample-app/hashf"
    "sample-app/mylib"
    "time"
)



type TimeStampRequest struct {
    Hash      []byte `json:"hash"`
    Signature []byte `json:"signature"`
    HashAlg   string `json:"hash_alg"`
}

type TimeStampResponse struct {
    Timestamp   int64  `json:"timestamp"`
    ServerSign  []byte `json:"server_sign"`
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
    Hash                      []byte            `json:"hash,omitempty"`
    Signature                 []byte            `json:"signature,omitempty"`
    ServerSign                []byte            `json:"server_sign,omitempty"`
    TimeStamp                 int64             `json:"timestamp,omitempty"`
    ServerCert                string            `json:"server_cert,omitempty"`
}


func signDataForTimeStamp(
    data []byte,
    priv *big.Int,
    n *big.Int,
    hashAlg string,
) (hash []byte, signature []byte, usedHashAlg string, err error) {
    switch hashAlg {
    case "sha256":
        h := hashf.Sha256(data)
        hash = h[:]
        usedHashAlg = "SHA256"
    case "streebog256":
        hash = hashf.StreebogHash(data, 32)
        usedHashAlg = "GOST256"
    default:
        return nil, nil, "", fmt.Errorf("unknown hash algorithm")
    }
    m := new(big.Int).SetBytes(hash)
    signature = mylib.FastPowMod(m, priv, n).Bytes()
    return hash, signature, usedHashAlg, nil
}


func verifySignature(hash, signature []byte, pub *big.Int, n *big.Int) bool {
    m := new(big.Int).SetBytes(signature)
    hashRecovered := mylib.FastPowMod(m, pub, n).Bytes()
    return string(hashRecovered) == string(hash)
}


func serverSign(hash []byte, hashAlg string, serverPriv *big.Int, serverN *big.Int) (serverSignature []byte, timestamp int64) {
    timestamp = time.Now().Unix()
    dataToSign := append(hash, []byte(time.Unix(timestamp, 0).UTC().Format("060102150405Z"))...)

    var hashForSign []byte
    switch hashAlg {
    case "sha256", "SHA256":
        h := hashf.Sha256(dataToSign)
        hashForSign = h[:]
    case "streebog256", "GOST256":
        hashForSign = hashf.StreebogHash(dataToSign, 32)
    default:
        h := hashf.Sha256(dataToSign)
        hashForSign = h[:]
    }

    m := new(big.Int).SetBytes(hashForSign)
    serverSignature = mylib.FastPowMod(m, serverPriv, serverN).Bytes()
    return serverSignature, timestamp
}


func verifyCenterSignature(hash []byte, timestamp int64, serverSign []byte, hashAlg string, serverPub *big.Int, serverN *big.Int) bool {
    dataToCheck := append(hash, []byte(time.Unix(timestamp, 0).UTC().Format("060102150405Z"))...)

    var hashForCheck []byte
    switch hashAlg {
    case "sha256", "SHA256":
        h := hashf.Sha256(dataToCheck)
        hashForCheck = h[:]
    case "streebog256", "GOST256":
        hashForCheck = hashf.StreebogHash(dataToCheck, 32)
    default:
        h := hashf.Sha256(dataToCheck)
        hashForCheck = h[:]
    }

    m := new(big.Int).SetBytes(serverSign)
    hashRecovered := mylib.FastPowMod(m, serverPub, serverN).Bytes()
    return string(hashRecovered) == string(hashForCheck)
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