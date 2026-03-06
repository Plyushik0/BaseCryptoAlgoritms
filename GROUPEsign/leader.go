package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"sample-app/mylib"
)

type SigningState struct {
    MessageHash    *big.Int
    MessageBytes   []byte
    NumParticipants int
    MembersData    map[int]map[string]*big.Int
    UProduct       *big.Int
    CombinedS      *big.Int
    LeaderE        *big.Int
    LeaderT        *big.Int
    LeaderRProduct *big.Int
    SigningStarted bool
    TSResponse     map[string]interface{}
    HashAlgoHM     string
    Lambdas        map[int]*big.Int
}

func startLeaderServer() error {
    var publicParams PublicParams
    var leaderPrivateKeys LeaderPrivateKeys
    if err := loadJSON("group_public_params.json", &publicParams); err != nil {
        return fmt.Errorf("ошибка загрузки публичных параметров: %v", err)
    }
    if err := loadJSON("leader_private_keys.json", &leaderPrivateKeys); err != nil {
        return fmt.Errorf("ошибка загрузки ключей лидера: %v", err)
    }

    p := publicParams.P
    q := publicParams.Q
    a := publicParams.A
    z := leaderPrivateKeys.Z
    d := leaderPrivateKeys.D
    e := publicParams.E
    signingState := &SigningState{
        Lambdas: make(map[int]*big.Int),
    }

    listener, err := net.Listen("tcp", "localhost:10000")
    if err != nil {
        return fmt.Errorf("ошибка запуска сервера: %v", err)
    }
    defer listener.Close()
    fmt.Println("Сервер лидера запущен на localhost:10000")

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Printf("Ошибка принятия соединения: %v\n", err)
            continue
        }
        go handleLeaderConnection(conn, signingState, publicParams, z, p, q, a, d, e)
    }
}

func handleLeaderConnection(conn net.Conn, state *SigningState, params PublicParams, z, p, q, a, d, e *big.Int) {
    defer conn.Close()
    data := make([]byte, 16384)
    n, err := conn.Read(data)
    if err != nil {
        fmt.Printf("Ошибка чтения данных: %v\n", err)
        return
    }

    var request map[string]interface{}
    if err := json.Unmarshal(data[:n], &request); err != nil {
        conn.Write([]byte(`{"error":"Неверный JSON"}`))
        return
    }

    reqType, _ := request["type"].(string)
    fmt.Printf("Получен запрос: %s\n", reqType)

    switch reqType {
    case "init_signing":
        messageHash, _ := new(big.Int).SetString(request["message_hash"].(string), 10)
        numParticipants := int(request["num_participants"].(float64))
        messageBytes, _ := hex.DecodeString(request["message_bytes_hex"].(string))
        hashAlgoHM := request["hash_algo_hm"].(string)

        *state = SigningState{
            MessageHash:     messageHash,
            MessageBytes:    messageBytes,
            NumParticipants: numParticipants,
            MembersData:     make(map[int]map[string]*big.Int),
            UProduct:        big.NewInt(1),
            CombinedS:       big.NewInt(0),
            LeaderRProduct:  big.NewInt(1),
            SigningStarted:  true,
            HashAlgoHM:      hashAlgoHM,
            Lambdas:         make(map[int]*big.Int),
        }
        fmt.Printf("init_signing: MessageBytes = %x\n", messageBytes)

        // Вычисление U и lambda_i
        for _, pk := range params.PublicKeyDirectory {
            lambdaI := computeLambda(state.MessageBytes, pk.P, d)
            state.Lambdas[pk.Index] = lambdaI
            state.UProduct.Mul(state.UProduct, mylib.FastPowMod(pk.P, lambdaI, p)).Mod(state.UProduct, p)
            fmt.Printf("lambda_%d: %s\n", pk.Index, lambdaI.String())
            fmt.Printf("P_%d: %s\n", pk.Index, pk.P.String())
            fmt.Printf("UProduct (после P_%d): %s\n", pk.Index, state.UProduct.String())
        }
        fmt.Printf("UProduct: %s\n", state.UProduct.String())
        fmt.Printf("Процесс подписи инициализирован для %d участников.\n", numParticipants)
        conn.Write([]byte(`{"status":"signing_initialized"}`))

    case "member_step_2":
        memberIdx := int(request["member_index"].(float64))
        t, _ := new(big.Int).SetString(request["t"].(string), 10)
        R, _ := new(big.Int).SetString(request["R"].(string), 10)

        if !state.SigningStarted || state.NumParticipants == 0 {
            conn.Write([]byte(`{"error":"Процесс подписи не инициализирован"}`))
            return
        }
        if memberIdx < 0 || memberIdx >= len(params.PublicKeyDirectory) {
            conn.Write([]byte(fmt.Sprintf(`{"error":"Неверный индекс участника: %d"}`, memberIdx)))
            return
        }

        state.MembersData[memberIdx] = map[string]*big.Int{"t": t, "R": R}
        state.LeaderRProduct.Mul(state.LeaderRProduct, R).Mod(state.LeaderRProduct, p)
        fmt.Printf("Получены данные шага 2 от участника %d. Всего: %d/%d\n", memberIdx, len(state.MembersData), state.NumParticipants)
        fmt.Printf("t_%d: %s\n", memberIdx, t.String())
        fmt.Printf("R_%d: %s\n", memberIdx, R.String())
        fmt.Printf("LeaderRProduct: %s\n", state.LeaderRProduct.String())

        if len(state.MembersData) == state.NumParticipants {
            fmt.Println("Все участники отправили данные шага 2. Выполняется шаг 3.")
            T, _ := rand.Int(rand.Reader, q)
            RPrime := mylib.FastPowMod(a, T, p)
            RCombined := new(big.Int).Mul(RPrime, state.LeaderRProduct)
            RCombined.Mod(RCombined, p)
            UCombined := state.UProduct
            fmt.Printf("T: %s\n", T.String())
            fmt.Printf("RPrime: %s\n", RPrime.String())
            fmt.Printf("RProduct: %s\n", state.LeaderRProduct.String())
            fmt.Printf("RCombined: %s\n", RCombined.String())
            fmt.Printf("UCombined: %s\n", UCombined.String())
            fmt.Printf("MessageBytes: %x\n", state.MessageBytes)

            E := hashValuesForE(state.MessageBytes, RCombined, UCombined)
            fmt.Printf("E: %s\n", E.String())
            fmt.Printf("Input to hashValuesForE: MessageBytes=%x, RCombined=%s, UCombined=%s\n", state.MessageBytes, RCombined.String(), UCombined.String())
            state.LeaderE = E
            state.LeaderT = T

        // Все участники отправили данные шага 2. E и T вычислены.
            // Отправляем статус, что шаг 3 завершен. Участники запросят E отдельно.
            conn.Write([]byte(`{"status":"step_3_ready"}`)) // Используем json.NewEncoder неявно через Write
            fmt.Println("Лидер вычислил E и T. Шаг 3 завершен.")

   
        } else {
            conn.Write([]byte(`{"status":"step_2_received"}`))
        }

    case "get_E":
        if !state.SigningStarted || state.LeaderE == nil {
            conn.Write([]byte(`{"error":"E ещё не вычислено"}`))
             // Send a status indicating E is not ready, member should retry
             conn.Write([]byte(`{"status":"E_not_ready"}`)) // Using json.NewEncoder неявно через Write
             return // Don't return error, just indicate not ready
        }
        memberIdx, ok := request["member_index"].(float64)
        if !ok {
            conn.Write([]byte(`{"error":"Отсутствует member_index"}`))
            return
        }
        idx := int(memberIdx)
        if _, exists := state.Lambdas[idx]; !exists {
            conn.Write([]byte(fmt.Sprintf(`{"error":"lambda_%d не найдено"}`, idx)))
            return
        }
        responseForMember := map[string]interface{}{
            "type":   "leader_step_3",
            "E":      state.LeaderE.String(),
            "Lambda": state.Lambdas[idx].String(),

        }
        json.NewEncoder(conn).Encode(responseForMember)
        fmt.Printf("Отправлено E и lambda_%d: %s\n", idx, state.Lambdas[idx].String())

    case "member_step_4":
        memberIdx := int(request["member_index"].(float64))
        S, _ := new(big.Int).SetString(request["S"].(string), 10)

        if !state.SigningStarted || state.LeaderE == nil {
            conn.Write([]byte(`{"error":"Процесс подписи не готов"}`))
            return
        }
        if _, exists := state.MembersData[memberIdx]; !exists {
            conn.Write([]byte(`{"error":"Шаг 2 не выполнен для этого участника"}`))
            return
        }

        // Проверка частичной подписи
        pk := params.PublicKeyDirectory[memberIdx].P
        lambdaI := state.Lambdas[memberIdx]
        pkInv, err := modInverse(pk, p)
        if err != nil {
            fmt.Printf("Ошибка вычисления обратного P_%d: %v\n", memberIdx, err)
            conn.Write([]byte(`{"error":"Ошибка вычисления обратного элемента"}`))
            return
        }
        lambdaE := new(big.Int).Mul(lambdaI, state.LeaderE)
        left := mylib.FastPowMod(pkInv, lambdaE, p)
        right := mylib.FastPowMod(a, S, p)
        RiComputed := new(big.Int).Mul(left, right).Mod(new(big.Int).Mul(left, right), p)
        fmt.Printf("Проверка S_%d для участника %d\n", memberIdx, memberIdx)
        fmt.Printf("P_%d: %s\n", memberIdx, pk.String())
        fmt.Printf("lambda_%d: %s\n", memberIdx, lambdaI.String())
        fmt.Printf("E: %s\n", state.LeaderE.String())
        fmt.Printf("lambda_%d * E: %s\n", memberIdx, lambdaE.String())
        fmt.Printf("P_%d^{-1}: %s\n", memberIdx, pkInv.String())
        fmt.Printf("Left: %s\n", left.String())
        fmt.Printf("Right: %s\n", right.String())
        fmt.Printf("RiComputed: %s\n", RiComputed.String())
        fmt.Printf("R_%d (ожидаемое): %s\n", memberIdx, state.MembersData[memberIdx]["R"].String())
        if RiComputed.Cmp(state.MembersData[memberIdx]["R"]) != 0 {
            fmt.Printf("Частичная подпись участника %d не прошла проверку\n", memberIdx)
            conn.Write([]byte(`{"error":"Неверная частичная подпись"}`))
            return
        }

        state.MembersData[memberIdx]["S"] = S
        state.CombinedS.Add(state.CombinedS, S).Mod(state.CombinedS, q)
        fmt.Printf("Получены данные шага 4 от участника %d. Всего: %d/%d\n", memberIdx, len(state.MembersData), state.NumParticipants)
        fmt.Printf("S_%d: %s\n", memberIdx, S.String())
        fmt.Printf("CombinedS: %s\n", state.CombinedS.String())

        if countS := len(state.MembersData); countS == state.NumParticipants {
            fmt.Println("Все участники отправили данные шага 4. Выполняется шаг 5.")
            T := state.LeaderT
            E := state.LeaderE
            SPrime := new(big.Int).Add(T, new(big.Int).Mul(z, E))
            SPrime.Mod(SPrime, q)
            SFinal := new(big.Int).Add(SPrime, state.CombinedS)
            SFinal.Mod(SFinal, q)
            UFinal := state.UProduct

            groupSignature := map[string]*big.Int{"U": UFinal, "E": E, "S": SFinal}
            fmt.Printf("Финальная подпись: %v\n", groupSignature)

            tsResponse, err := requestTimestamp(groupSignature)
            if err != nil {
                fmt.Printf("Ошибка получения временной метки: %v\n", err)
                tsResponse = map[string]interface{}{"error": err.Error()}
            }

            response := SigningCompleteResponse{
                Status:         "signing_complete",
                GroupSignature: groupSignature,
                Timestamp:      tsResponse,
            }
            json.NewEncoder(conn).Encode(response)
            *state = SigningState{} // Сброс состояния
            fmt.Println("Процесс подписи завершён, состояние сброшено.")
        } else {
            conn.Write([]byte(`{"status":"step_4_received"}`))
        }

    default:
        conn.Write([]byte(fmt.Sprintf(`{"error":"Неизвестный тип запроса: %s"}`, reqType)))
    }
}

func requestTimestamp(signature map[string]*big.Int) (map[string]interface{}, error) {
    conn, err := net.Dial("tcp", "localhost:9999")
    if err != nil {
        return nil, fmt.Errorf("ошибка подключения к TS: %v", err)
    }
    defer conn.Close()
    sigJSON, err := json.Marshal(signature)
    if err != nil {
        return nil, fmt.Errorf("ошибка маршалинга подписи: %v", err)
    }
    request := map[string]string{"group_signature_encoded": hex.EncodeToString(sigJSON)}
    fmt.Printf("Отправка TS запроса: %v\n", request)
    if err := json.NewEncoder(conn).Encode(request); err != nil {
        return nil, fmt.Errorf("ошибка отправки запроса: %v", err)
    }
    var response map[string]interface{}
    if err := json.NewDecoder(conn).Decode(&response); err != nil {
        return nil, fmt.Errorf("ошибка получения ответа: %v", err)
    }
    fmt.Printf("Получен TS ответ: %v\n", response)
    return response, nil
}