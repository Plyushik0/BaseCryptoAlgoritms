package main

import (
    "sample-app/hashf" 
    "fmt"
    "math/big"
    "sample-app/mylib"

)

func hashMessage(message []byte) *big.Int {
    hash := hashf.Sha256(message)
    return new(big.Int).SetBytes(hash[:])
}

func computeLambda(message []byte, publicKey *big.Int, d *big.Int) *big.Int {
    msgHash := hashf.Sha256(message)
    msgHashStr := fmt.Sprintf("%x", msgHash[:])
    pkStr := publicKey.String()
    innerData := msgHashStr + pkStr + d.String()
    innerHash := hashf.Sha256([]byte(innerData))
    lambda := new(big.Int).SetBytes(innerHash[:])
    return lambda
}

func hashValuesForE(message []byte, R, U *big.Int) *big.Int {
    data := append([]byte{}, message...) 
    data = append(data, []byte(R.String())...); data = append(data, []byte(U.String())...)
    hash := hashf.Sha256(data) 
    return new(big.Int).SetBytes(hash[:])
}

func modInverse(a, m *big.Int) (*big.Int, error) {
    gcd, x, _ := mylib.ExtendedGCD(a, m)
    if gcd.Cmp(big.NewInt(1)) != 0 {
        return nil, fmt.Errorf("модульный обратный не существует")
    }
    return x.Mod(x, m), nil
}