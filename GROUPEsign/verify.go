package main

import (
    "encoding/json"
    "fmt"
    "math/big"
    "os"
    "encoding/hex"
    "sample-app/mylib"
    "strings"
)

func verifyGroupSignature(documentPath, signaturePath string) error {
    var publicParams PublicParams
    var cadesSignature map[string]interface{}
    messageBytes, err := os.ReadFile(documentPath)
    if err != nil {
        return fmt.Errorf("ошибка чтения документа: %v", err)
    }
    if err := loadJSON("group_public_params.json", &publicParams); err != nil {
        return fmt.Errorf("ошибка загрузки публичных параметров: %v", err)
    }
    if err := loadJSON(signaturePath, &cadesSignature); err != nil {
        return fmt.Errorf("ошибка загрузки подписи: %v", err)
    }

    p := publicParams.P
    a := publicParams.A
    L := publicParams.L
 
    e := publicParams.E
    TSPublicKey := publicParams.TSPublicKey
    hashAlgoHM := publicParams.HashAlgoHM
    hashAlgoForE := publicParams.HashAlgoForECalc

    signerInfos, _ := cadesSignature["SignerInfos"].(map[string]interface{})
    hashAlgoInSig := signerInfos["DigestAlgorithmIdentifier"].(string)
    if !strings.EqualFold(hashAlgoInSig, hashAlgoHM) {
        fmt.Printf("Предупреждение: Алгоритм хеширования в подписи (%s) не совпадает с ожидаемым (%s)\n", hashAlgoInSig, hashAlgoHM)
    }

    sigHex := signerInfos["SignatureValue"].(string)
    sigBytes, _ := hex.DecodeString(sigHex)
    var sig map[string]*big.Int
    if err := json.Unmarshal(sigBytes, &sig); err != nil {
        return fmt.Errorf("ошибка декодирования подписи: %v", err)
    }
    U := sig["U"]
    E := sig["E"]
    S := sig["S"]
    if U == nil || E == nil || S == nil {
        return fmt.Errorf("неполная подпись")
    }

    fmt.Println("Извлечена подпись (U, E, S)")
    fmt.Printf("U: %s\n", U.String())
    fmt.Printf("L: %s\n", L.String())
    fmt.Printf("S: %s\n", S.String())
    fmt.Printf("a: %s\n", a.String())
    fmt.Printf("p: %s\n", p.String())
    fmt.Printf("e: %s\n", e.String())
    fmt.Printf("Ожидаемый алгоритм хеширования h(M): %s\n", hashAlgoHM)
    fmt.Printf("Ожидаемый алгоритм хеширования h(M || R || U): %s\n", hashAlgoForE)
    fmt.Printf("MessageBytes: %x\n", messageBytes)

    // H = h(M)
    H := hashMessage(messageBytes)
    fmt.Printf("Вычислено H = h(M): %s\n", H.String())

    // R_' = (U * L)^(-1) * a^(S^e mod q) mod p
    UL := new(big.Int).Mul(U, L)
    UL.Mod(UL, p)
    ULInv, err := modInverse(UL, p)
    if err != nil {
        fmt.Printf("Ошибка вычисления обратного элемента: %v\n", err)
        return err
    }
    aPowS := mylib.FastPowMod(a, S, p) 
    ULInvPowE := mylib.FastPowMod(ULInv, E, p)
    R_tilde := new(big.Int).Mul(aPowS, ULInvPowE)
    R_tilde.Mod(R_tilde, p)
    fmt.Printf("R_tilde: %s\n", R_tilde.String())

    // E_' = h(M || R_tilde || U)
    E_tilde := hashValuesForE(messageBytes, R_tilde, U)
    fmt.Printf("E_tilde: %s\n", E_tilde.String())

    // E_' == E
    signatureValid := E_tilde.Cmp(E) == 0
    fmt.Printf("Проверка E_tilde == E: %v == %v -> %v\n", E_tilde, E, signatureValid)
    if signatureValid {
        fmt.Println("Подпись верна")
    } else {
        fmt.Println("Подпись не верна")
    }

    
    timestampValid := true
    var unsignedAttrs map[string]interface{}
    var attrs map[string]interface{}
    if ua, exists := signerInfos["UnsignedAttributes OPTIONAL"].(map[string]interface{}); exists {
        unsignedAttrs = ua
        fmt.Println("Найдена временная метка. Проверка...")
        attrs, _ = unsignedAttrs["SET_OF_AttributeValue"].(map[string]interface{})
        tsTimestamp := attrs["timestamp"].(string)
        tsSignature, _ := new(big.Int).SetString(attrs["ts_signature"].(string), 10)
        tsPubKey := attrs["ts_public_key"].(map[string]interface{})
        tsHashHex := attrs["hash"].(string)

        tsN, _ := new(big.Int).SetString(tsPubKey["n"].(string), 10)
        tsE, _ := new(big.Int).SetString(tsPubKey["e"].(string), 10)
        if tsN.Cmp(TSPublicKey.N) != 0 || tsE.Cmp(TSPublicKey.E) != 0 {
            fmt.Println("Предупреждение: Ключ TS не соответствует ожидаемому")
            timestampValid = false
        }

        if tsHashHex != sigHex {
            fmt.Println("Ошибка: Хэш в метке не совпадает с подписью")
            timestampValid = false
        } else {
            tsData := append(sigBytes, []byte(tsTimestamp)...)
            tsHash := hashMessage(tsData)
            tsComputedHash := mylib.FastPowMod(tsSignature, tsE, tsN)
            timestampValid = tsHash.Cmp(tsComputedHash) == 0
            fmt.Printf("Проверка TS: %t\n", timestampValid)
        }
    } else {
        fmt.Println("Временная метка отсутствует.")
    }

    overallValid := signatureValid && timestampValid
    fmt.Printf("\nРезультат проверки подписи:\n")
    fmt.Printf("Групповая подпись: %s\n", map[bool]string{true: "верна", false: "неверна"}[overallValid])
    fmt.Printf("Проверка схемы: %t\n", signatureValid)
    if _, exists := signerInfos["UnsignedAttributes OPTIONAL"]; exists {
        fmt.Printf("Проверка временной метки: %t\n", timestampValid)
    }
    fmt.Printf("Алгоритм хеширования h(M): %s\n", hashAlgoHM)
    fmt.Printf("Алгоритм хеширования h(M || R || U): %s\n", hashAlgoForE)
    fmt.Println("Алгоритм: Group Signature (RSA для лидера и TS)")
    fmt.Printf("Идентификатор: %s\n", signerInfos["SignerIdentifier"].(string))
    if unsignedAttrs != nil {
        fmt.Printf("Время: %s\n", attrs["timestamp"].(string))
    } else {
        fmt.Printf("Время: N/A\n")
    }

    return nil
}