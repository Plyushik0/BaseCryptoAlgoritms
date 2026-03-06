package main

import (
	"fmt"
	"math/big"
	"sample-app/hashf"
	"sample-app/mylib"
)

func signDataForTimeStamp(
	data []byte,
	priv *big.Int,
	n *big.Int,
	hashAlg string,
) (hash []byte, signature []byte, usedHashAlg string, err error) {
	if hashAlg != "sha512" {
		return nil, nil, "", fmt.Errorf("поддерживается только sha512")
	}
	h := hashf.Sha512(data)
	hash = h[:]
	usedHashAlg = "SHA512"
	m := new(big.Int).SetBytes(hash)
	signature = mylib.FastPowMod(m, priv, n).Bytes()
	return hash, signature, usedHashAlg, nil
}

func verifySignature(hash, signature []byte, pub *big.Int, n *big.Int) bool {
	m := new(big.Int).SetBytes(signature)
	hashRecovered := mylib.FastPowMod(m, pub, n).Bytes()
	return string(hashRecovered) == string(hash)
}