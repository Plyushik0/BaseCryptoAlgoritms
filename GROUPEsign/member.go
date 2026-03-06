package main

import (
    "crypto/rand"
    "encoding/json"
    "fmt"
    "math/big"
    "net"
    "os"
    "sample-app/mylib"
    "time"
)

type MemberStep2Request struct {
    Type        string `json:"type"`
    MemberIndex int    `json:"member_index"`
    T           string `json:"t"`
    R           string `json:"R"`
}

type MemberStep4Request struct {
    Type        string `json:"type"`
    MemberIndex int    `json:"member_index"`
    S           string `json:"S"`
}

type LeaderStep3Response struct {
    Type   string   `json:"type"`
    E      *big.Int `json:"E"`
    Lambda *big.Int `json:"lambda"`
}

type SigningCompleteResponse struct {
    Status         string                 `json:"status"`
    GroupSignature map[string]*big.Int    `json:"group_signature"`
    Timestamp      map[string]interface{} `json:"timestamp_response"`
}

func memberSignDocument(memberID int, documentPath string, numParticipants int) error {
    var memberSecret MemberKey
    var publicParams PublicParams
    if err := loadJSON(fmt.Sprintf("member_%d_secret.json", memberID), &memberSecret); err != nil {
        return fmt.Errorf("ошибка загрузки ключей участника: %v", err)
    }
    if err := loadJSON(memberSecret.GroupPublicParamsFile, &publicParams); err != nil {
        return fmt.Errorf("ошибка загрузки публичных параметров: %v", err)
    }
    messageBytes, err := os.ReadFile(documentPath)
    if err != nil {
        return fmt.Errorf("ошибка чтения документа: %v", err)
    }

    kJ := memberSecret.K
    p := publicParams.P
    q := publicParams.Q
    a := publicParams.A
    PJ := mylib.FastPowMod(a, kJ, p)
    fmt.Printf("Участник %d: k_j=%s, P_j=%s\n", memberID, kJ.String(), PJ.String())

    memberIdx := -1
    for idx, pk := range publicParams.PublicKeyDirectory {
        if pk.P.Cmp(PJ) == 0 {
            memberIdx = idx
            break
        }
    }
    if memberIdx == -1 {
        return fmt.Errorf("публичный ключ участника не найден")
    }
    fmt.Printf("Участник %d (Индекс %d) начинает процесс подписи.\n", memberID, memberIdx)

    H := hashMessage(messageBytes)
    fmt.Printf("Вычислен хеш сообщения: H=%s\n", H.String())

    if memberID == 1 {
        fmt.Println("Инициация процесса подписи лидером.")
        conn, err := net.Dial("tcp", "localhost:10000")
        if err != nil {
            return fmt.Errorf("ошибка подключения к лидеру: %v", err)
        }
        defer conn.Close()
        initRequest := map[string]interface{}{
            "type":              "init_signing",
            "message_hash":      H.String(),
            "message_bytes_hex": fmt.Sprintf("%x", messageBytes),
            "num_participants":  numParticipants,
            "hash_algo_hm":      "sha256",
        }
        fmt.Printf("Отправка init_signing: %v\n", initRequest)
        if err := json.NewEncoder(conn).Encode(initRequest); err != nil {
            return fmt.Errorf("ошибка отправки запроса инициации: %v", err)
        }
        var response map[string]string
        if err := json.NewDecoder(conn).Decode(&response); err != nil {
            return fmt.Errorf("ошибка получения ответа: %v", err)
        }
        fmt.Printf("Получен ответ init_signing: %v\n", response)
        if response["status"] != "signing_initialized" {
            return fmt.Errorf("ошибка инициации: %s", response["error"])
        }
        fmt.Println("Лидер подтвердил инициацию.")
        time.Sleep(500 * time.Millisecond)
    } else {
        time.Sleep(2 * time.Second) 
        fmt.Println("Ожидание инициации участником 1...")
        time.Sleep(2 * time.Second)
    }

    tI, _ := rand.Int(rand.Reader, q)
    R_I := mylib.FastPowMod(a, tI, p)
    fmt.Printf("Участник %d: t_i=%s, R_i=%s\n", memberID, tI.String(), R_I.String())

    conn, err := net.Dial("tcp", "localhost:10000")
    if err != nil {
        return fmt.Errorf("ошибка подключения для шага 2: %v", err)
    }
    defer conn.Close()
    step2Request := MemberStep2Request{
        Type:        "member_step_2",
        MemberIndex: memberIdx,
        T:           tI.String(),
        R:           R_I.String(),
    }
    fmt.Printf("Отправка шага 2: %v\n", step2Request)
    if err := json.NewEncoder(conn).Encode(step2Request); err != nil {
        return fmt.Errorf("ошибка отправки шага 2: %v", err)
    }
    conn.SetReadDeadline(time.Now().Add(10 * time.Second))
    var response map[string]interface{}
    if err := json.NewDecoder(conn).Decode(&response); err != nil {
        fmt.Printf("Ошибка получения ответа: %v\n", err)
        return fmt.Errorf("ошибка получения ответа шага 2: %v", err)
    }
    fmt.Printf("Получен ответ шага 2: %v\n", response)

    // Check the status from the leader's response
    status, _ := response["status"].(string)
    if !(status == "step_2_received" || status == "step_3_ready" || response["type"] == "leader_step_3")  {
        return fmt.Errorf("ошибка шага 2: %s", response["error"])
    }
    fmt.Println("Лидер подтвердил получение данных шага 2.")

    fmt.Println("Ожидание E от лидера...")
    var E, lambdaI *big.Int
    for retries := 0; retries < 10; retries++ {
        conn, _ := net.Dial("tcp", "localhost:10000")
       
        request := map[string]interface{}{
            "type":         "get_E",
            "member_index": float64(memberIdx),
        }
        if err := json.NewEncoder(conn).Encode(request); err != nil {
            conn.Close()
            time.Sleep(1 * time.Second)
            continue
        }
        var genericResp map[string]interface{}
        decoder := json.NewDecoder(conn)
        if err := decoder.Decode(&genericResp); err != nil {
            conn.Close()
            fmt.Printf("Ошибка декодирования ответа get_E: %v. Повтор через 1 сек...\n", err)
            time.Sleep(1 * time.Second) // Wait before retrying

            continue
        }
        conn.Close()
        
        if status, ok := genericResp["status"].(string); ok && status == "E_not_ready"  {
            // Leader indicated E is not ready yet, wait and retry
            fmt.Printf("E ещё не готово. Повтор через 1 сек...\n")
            time.Sleep(1 * time.Second) // Wait before retrying
            continue
        }
        if respType, ok := genericResp["type"].(string); ok && respType == "leader_step_3" {
            // Попытка извлечь E и Lambda из genericResp
            // Это упрощенный вариант, в идеале нужно использовать json.Unmarshal снова с байтами ответа
            if eStr, eOk := genericResp["E"].(string); eOk {
                E, _ = new(big.Int).SetString(eStr, 10)
            }
            if lStr, lOk := genericResp["Lambda"].(string); lOk {
                lambdaI, _ = new(big.Int).SetString(lStr, 10)
            }

            if E != nil && lambdaI != nil {
                fmt.Printf("Получено E от лидера: %s\n", E.String())
                fmt.Printf("Получено lambda_%d: %s\n", memberIdx, lambdaI.String())
                break
            }
        }

        fmt.Printf("Неожиданный ответ от лидера на запрос get_E: %v. Повтор через 1 сек...\n", genericResp)
        time.Sleep(1 * time.Second) // Fallback retry
    }
    if E == nil || lambdaI == nil {
        return fmt.Errorf("не удалось получить E или lambda_i после 10 попыток")
    }

    fmt.Printf("Участник %d выполняет шаг 4.\n", memberID)
    kJLambdaE := new(big.Int).Mul(kJ, new(big.Int).Mul(lambdaI, E))
    S_I := new(big.Int).Add(tI, kJLambdaE)
    S_I.Mod(S_I, q)
    fmt.Printf("Вычисление S_%d для участника %d\n", memberIdx, memberID)
    fmt.Printf("t_%d: %s\n", memberIdx, tI.String())
    fmt.Printf("k_j: %s\n", kJ.String())
    fmt.Printf("lambda_%d: %s\n", memberIdx, lambdaI.String())
    fmt.Printf("E: %s\n", E.String())
    fmt.Printf("k_j * lambda_%d * E: %s\n", memberIdx, kJLambdaE.String())
    fmt.Printf("S_%d: %s\n", memberIdx, S_I.String())

    conn, err = net.Dial("tcp", "localhost:10000")
    if err != nil {
        return fmt.Errorf("ошибка подключения для шага 4: %v", err)
    }
    defer conn.Close()
    step4Request := MemberStep4Request{
        Type:        "member_step_4",
        MemberIndex: memberIdx,
        S:           S_I.String(),
    }
    fmt.Printf("Отправка шага 4: %v\n", step4Request)
    if err := json.NewEncoder(conn).Encode(step4Request); err != nil {
        return fmt.Errorf("ошибка отправки шага 4: %v", err)
    }
    var finalResp SigningCompleteResponse
    if err := json.NewDecoder(conn).Decode(&finalResp); err != nil {
        return fmt.Errorf("ошибка получения ответа шага 4: %v", err)
    }

    if finalResp.Status == "step_4_received" {
        fmt.Println("Лидер подтвердил получение данных шага 4. Ожидание финальной подписи...")
        return nil
    } else if finalResp.Status == "signing_complete" {
        fmt.Println("Лидер завершил процесс подписи.")
        cadesSignature := map[string]interface{}{
            "CMSVersion":                 1,
            "DigestAlgorithmIdentifiers": "sha256",
            "EncapsulatedContentInfo": map[string]interface{}{
                "ContentType":  "text/plain",
                "OCTET_STRING": fmt.Sprintf("%x", messageBytes),
            },
            "CertificateSet OPTIONAL":        nil,
            "RevocationInfoChoices OPTIONAL": nil,
            "SignerInfos": map[string]interface{}{
                "CMSVersion":                   1,
                "SignerIdentifier":             "Group Member",
                "DigestAlgorithmIdentifier":    "sha256",
                "SignatureAlgorithmIdentifier": "GroupSignature",
                "SignatureValue": func() string {
                    b, err := json.Marshal(finalResp.GroupSignature)
                    if err != nil {
                        return ""
                    }
                    return fmt.Sprintf("%x", b)
                }(),
                "UnsignedAttributes OPTIONAL": nil,
            },
        }

        if finalResp.Timestamp != nil {
            if finalResp.Timestamp["error"] != nil {
                fmt.Printf("Ошибка TS: %v\n", finalResp.Timestamp["error"])
            } else {
                fmt.Println("Добавление временной метки.")
                cadesSignature["SignerInfos"].(map[string]interface{})["UnsignedAttributes OPTIONAL"] = map[string]interface{}{
                    "OBJECT_IDENTIFIER": "signature-time-stamp",
                    "SET_OF_AttributeValue": map[string]interface{}{
                        "hash": func() string {
                            b, err := json.Marshal(finalResp.GroupSignature)
                            if err != nil {
                                return ""
                            }
                            return fmt.Sprintf("%x", b)
                        }(),
                        "timestamp":     finalResp.Timestamp["timestamp"],
                        "ts_signature":  finalResp.Timestamp["ts_signature"],
                        "ts_public_key": finalResp.Timestamp["ts_public_key"],
                    },
                }
            }
        } else {
            fmt.Println("TS ответ отсутствует")
        }

        if err := saveJSON("group_signature.cades", cadesSignature); err != nil {
            return fmt.Errorf("ошибка сохранения подписи: %v", err)
        }
        fmt.Println("Групповая подпись сохранена в 'group_signature.cades'")
    }
    return nil
}

func loadJSON(filename string, v interface{}) error {
    data, err := os.ReadFile(filename)
    if err != nil {
        return err
    }
    return json.Unmarshal(data, v)
}