package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"sample-app/cipher"
	"sample-app/mylib"
)

// PubKey представляет публичный ключ участника
type PubKey struct {
	Index int      `json:"index"`
	P     *big.Int `json:"big_P"`
}

// TSPubKey представляет публичный ключ службы временных меток
type TSPubKey struct {
	N *big.Int `json:"n"`
	E *big.Int `json:"e"`
}

// TSPrivateKey представляет приватный ключ службы временных меток
type TSPrivateKey struct {
	P *big.Int `json:"bigP"`
	Q *big.Int `json:"bigQ"`
	D *big.Int `json:"d"`
}

// MemberKey представляет секретный ключ участника
type MemberKey struct {
	ID                   int   `json:"id"`
	K                    *big.Int `json:"bigK"`
	GroupPublicParamsFile string   `json:"group_public_params_file"`
}

// PublicParams структура для публичных параметров группы
type PublicParams struct {
	P                    *big.Int       `json:"p"`
	Q                    *big.Int       `json:"q"`
	A                    *big.Int       `json:"a"`
	L                    *big.Int       `json:"L"`
	LeaderRSAE           *big.Int       `json:"leader_rsa_e"`
	LeaderRSAN           *big.Int       `json:"leader_rsa_n"`
	PublicKeyDirectory   []PubKey       `json:"public_key_directory"`
	TSPublicKey          TSPubKey       `json:"ts_public_key"`
	HashAlgoForECalc     string         `json:"hash_algo_for_E_calc"`
	HashAlgoHM           string         `json:"hash_algo_hm"`
	ModulusPBitLength    int            `json:"modulus_p_bit_length"`
	E                    *big.Int       `json:"e"` // Промежуточный ключ
}

// LeaderPrivateKeys структура для секретных ключей лидера
type LeaderPrivateKeys struct {
	Z           *big.Int `json:"z"`
	LeaderRSAD  *big.Int `json:"leader_rsa_d"`
	LeaderRSAP1 *big.Int `json:"leader_rsa_p1"`
	LeaderRSAP2 *big.Int `json:"leader_rsa_p2"`
	D           *big.Int `json:"d"` // Промежуточный ключ
}

// saveJSON сериализует данные в JSON и сохраняет в файл
func saveJSON(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("ошибка создания файла %s: %v", filename, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("ошибка сериализации JSON в %s: %v", filename, err)
	}
	return nil
}

func generateGroupParameters(numMembers int) error {
	fmt.Println("Генерация параметров группы...")

	// Генерация простых чисел p и q
	kBits := 512
	var p, q *big.Int
	for {
		q = mylib.GeneratePrimeBitK(kBits/2, 100)
		mStart := new(big.Int).Lsh(big.NewInt(1), uint(kBits-1))
		mStart.Div(mStart, new(big.Int).Mul(big.NewInt(2), q))
		if mStart.Bit(0) == 0 {
			mStart.Add(mStart, big.NewInt(1))
		}
		m := new(big.Int).Set(mStart)
		for i := 0; i < 1000; i++ {
			// p = 2*m*q + 1
			p = new(big.Int).Mul(big.NewInt(2), m)
			p.Mul(p, q)
			p.Add(p, big.NewInt(1))
			if p.BitLen() > kBits {
				break
			}
			if p.ProbablyPrime(100) {
				fmt.Printf("Сгенерированы p (%d бит) и q (%d бит)\n", p.BitLen(), q.BitLen())
				break
			}
			m.Add(m, big.NewInt(2))
		}
		isPPrime := mylib.MillerRabinPrimalityTest(p, 100) // Оставляем без обработки ошибок по запросу
		// if err != nil { return fmt.Errorf("ошибка проверки простоты p: %v", err) }
		if isPPrime {
			break
		}
	}

	// Генерация генератора a порядка q
	m := new(big.Int).Div(new(big.Int).Sub(p, big.NewInt(1)), q)
	var a *big.Int
	for {
		
		g, _ := rand.Int(rand.Reader, new(big.Int).Sub(p, big.NewInt(1)))
		g.Add(g, big.NewInt(2))
		a = mylib.FastPowMod(g, m, p)
		if a.Cmp(big.NewInt(1)) != 0 {
			break
		}
	}
	fmt.Println("Сгенерированы p, q, a.")

	// Генерация ключей лидера: z и L = a^z mod p
	z, _ := rand.Int(rand.Reader, q)
	L := mylib.FastPowMod(a, z, p)
	fmt.Println("Сгенерирован DL-ключ лидера (z, L).")

	// Генерация RSA-ключей лидера
	rsaKBits := 512
	eL, dL, nL, _ := cipher.GenerateRSAKeys(rsaKBits/2, "leader_public.json", "leader_private_temp.json")
	var p1, p2 *big.Int
	leaderPriv := make(map[string]interface{})
	if err := cipher.LoadFromFile("leader_private_temp.json", &leaderPriv); err == nil {
		p1, _ = new(big.Int).SetString(leaderPriv["prime1"].(string), 10)
		p2, _ = new(big.Int).SetString(leaderPriv["prime2"].(string), 10)
	}
	fmt.Println("Сгенерирован внутренний RSA-ключ лидера (e, n, d).")

	// Генерация промежуточных ключей e, d
	e := mylib.GeneratePrimeBitK(8, 50) // Простое число ~8 бит
	fn := q                             // FN = N = q
	d := new(big.Int)
	_, x, _ := mylib.ExtendedGCD(e, fn)
	d.Mod(x, fn)
	if d.Cmp(big.NewInt(0)) < 0 {
		d.Add(d, fn)
	}
	fmt.Println("Сгенерированы промежуточные ключи e, d.")

	// Генерация ключей участников
	memberKeys := make([]MemberKey, numMembers)
	publicKeys := make([]PubKey, numMembers)
	for j := 0; j < numMembers; j++ {
		kJ, _ := rand.Int(rand.Reader, q)
		PJ := mylib.FastPowMod(a, kJ, p)
		memberKeys[j] = MemberKey{
			ID:                   j + 1,
			K:                    kJ,
			GroupPublicParamsFile: "group_public_params.json",
		}
		publicKeys[j] = PubKey{Index: j, P: PJ}
	}
	fmt.Printf("Сгенерированы ключи для %d участников.\n", numMembers)

	// Перемешивание публичных ключей
	for i := len(publicKeys) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		publicKeys[i], publicKeys[j.Int64()] = publicKeys[j.Int64()], publicKeys[i]
	}
	for i := range publicKeys {
		publicKeys[i].Index = i
	}
	fmt.Println("Список публичных ключей перемешан.")

	// Генерация ключей службы временных меток
	tsKBits := 512
	eTS, dTS, nTS, _ := cipher.GenerateRSAKeys(tsKBits/2, "ts_public.json", "ts_private_temp.json")
	var tsP, tsQ *big.Int
	tsPriv := make(map[string]interface{})
	if err := cipher.LoadFromFile("ts_private_temp.json", &tsPriv); err == nil {
		tsP, _ = new(big.Int).SetString(tsPriv["prime1"].(string), 10)
		tsQ, _ = new(big.Int).SetString(tsPriv["prime2"].(string), 10)
	}

	tsPublicKey := TSPubKey{N: nTS, E: eTS}
	tsPrivateKey := TSPrivateKey{P: tsP, Q: tsQ, D: dTS}
	fmt.Println("Сгенерированы ключи службы временных меток.")

	// Сборка публичных параметров
	groupPublicParams := PublicParams{
		P:                  p,
		Q:                  q,
		A:                  a,
		L:                  L,
		LeaderRSAE:         eL,
		LeaderRSAN:         nL,
		PublicKeyDirectory: publicKeys,
		TSPublicKey:        tsPublicKey,
		HashAlgoForECalc:   "sha256",
		HashAlgoHM:         "sha256",
		ModulusPBitLength:  p.BitLen(),
		E:                  e,
	}

	// Сборка секретных ключей лидера
	leaderPrivateKeys := LeaderPrivateKeys{
		Z:           z,
		LeaderRSAD:  dL,
		LeaderRSAP1: p1,
		LeaderRSAP2: p2,
		D:           d,
	}

	// Сохранение файлов
	if err := saveJSON("group_public_params.json", groupPublicParams); err != nil {
		return err
	}
	if err := saveJSON("leader_private_keys.json", leaderPrivateKeys); err != nil {
		return err
	}
	if err := saveJSON("ts_private_key.json", tsPrivateKey); err != nil {
		return err
	}
	for _, mk := range memberKeys {
		if err := saveJSON(fmt.Sprintf("member_%d_secret.json", mk.ID), mk); err != nil {
			return err
		}
	}

	fmt.Println("\nГенерация завершена. Созданы файлы:")
	fmt.Println("- group_public_params.json")
	fmt.Println("- leader_private_keys.json")
	fmt.Println("- ts_private_key.json")
	for j := 1; j <= numMembers; j++ {
		fmt.Printf("- member_%d_secret.json\n", j)
	}
	return nil
}