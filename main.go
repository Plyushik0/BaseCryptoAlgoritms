package main

import (
	// "encoding/json"
	//"bufio"
	"fmt"
	"math/big"
	//"os"
	// "crypto/rand"
	// "sample-app/mylib"
)

func main() {
	for {

		var option int
		fmt.Println("1 - криптосистема RSA")
		fmt.Println("2 - криптосистема Рабина")
		fmt.Println("3 - криптосистема Эль-Гамаля")
		fmt.Println("ctrl + c - для выхода")
		fmt.Scan(&option)

		switch option {

		case 1:
			var e, d, n *big.Int
			var C []*big.Int
			for {
				fmt.Println("Выберите действие:")
				fmt.Println("1. Генерация параметров RSA")
				fmt.Println("2. Шифрование сообщения")
				fmt.Println("3. Расшифрование сообщения")
				fmt.Println("4. Выход")

				var choice int
				fmt.Scan(&choice)

				switch choice {
				case 1:

					e, d, n, _ = generateRSAKeys(1024)

					// Сохранение открытого ключа в формате PKCS 8
					ok := map[string]interface{}{
						"SubjectPublicKeyInfo": map[string]interface{}{
							"publicExponent": e.String(),
							"N":              n.String(),
						},
						"PKCS10CertRequest":   0,
						"Certificate":         0,
						"PKCS7CertChain-PKCS": 0,
					}
					saveToFile("OpenKey.json", ok)

					// Сохранение закрытого ключа в формате PKCS 12
					sk := map[string]interface{}{
						"privateExponent": d.String(),
						"prime1":          new(big.Int).Sub(n, big.NewInt(1)).String(),
						"prime2":          new(big.Int).Sub(n, big.NewInt(1)).String(),
						"exponent1":       new(big.Int).Mod(d, new(big.Int).Sub(n, big.NewInt(1))).String(),
						"exponent2":       new(big.Int).Mod(d, new(big.Int).Sub(n, big.NewInt(1))).String(),
					}
					saveToFile("SecretKey.json", sk)

					fmt.Println("Параметры RSA успешно сгенерированы и сохранены в файлы")

				case 2:
					fmt.Print("Сообщение для шифрования: ")
					message := "Hello, RSA!"
					fmt.Println(message)
					C = encryptRSA(message, e, n)
					encryptedData := map[string]interface{}{
						"encryptedMessage": C,
						"Version": 0,
						"ContentType": "text",
						"ContentEncryptionAlgorithmIdentifier": "rsaEncryption",
						"OPTIONAL": nil,
					}
					saveToFile("encrypted_message.json", encryptedData)
					fmt.Println("Сообщение успешно зашифровано и сохранено в файл")

				case 3:
					var encryptedData map[string][]*big.Int
					loadFromFile("encrypted_message.json", &encryptedData)
					C := encryptedData["encryptedMessage"]
					decryptedMessage := decryptRSA(C, d, n)
					fmt.Println("Расшифрованное сообщение:", decryptedMessage)

				case 4:
					fmt.Println("Выход из программы.")
					return

				default:
					fmt.Println("Неверный выбор. Пожалуйста, выберите действие от 1 до 4.")
				}
			}
		case 2:
			var p, q, n *big.Int
			var ciphertext, encMess *big.Int

			for {
				fmt.Println("Выберите действие:")
				fmt.Println("1. Генерация параметров (p, q, n)")
				fmt.Println("2. Шифрование сообщения")
				fmt.Println("3. Расшифрование сообщения")
				fmt.Println("4. Выход")

				var choice int
				fmt.Scan(&choice)

				switch choice {
				case 1:

					p, q, n = GenerateParameters(256)
					fmt.Println("Сгенерированные параметры:")
					fmt.Println("p:", p)
					fmt.Println("q:", q)
					fmt.Println("n:", n)

				case 2:

					fmt.Print("Сообщение для шифрования: ")
					messageText := "Hello, Rabin!"
					fmt.Println(messageText)
					ciphertext = new(big.Int).SetBytes([]byte(messageText))
					fmt.Println("Числовое представление:", ciphertext)
					encMess = encryptRabin(ciphertext, n)
					fmt.Println("Зашифрованное сообщение (число):", encMess)

				case 3:

					r1, r2, r3, r4 := decryptRabin(encMess, p, q)

					fmt.Println("Возможные исходные сообщения (числа):")
					fmt.Println(r1)
					fmt.Println(r2)
					fmt.Println(r3)
					fmt.Println(r4)

					fmt.Println("Возможные исходные сообщения (текст):")
					fmt.Println(string(r1.Bytes()))
					fmt.Println(string(r2.Bytes()))
					fmt.Println(string(r3.Bytes()))
					fmt.Println(string(r4.Bytes()))

				case 4:
					fmt.Println("Выход из программы.")
					return

				default:
					fmt.Println("Неверный выбор. Пожалуйста, выберите действие от 1 до 4.")
				}
			}

		case 3:
			var p, α, β, a *big.Int
			var c1, c2, ciphertext *big.Int

			for {
				fmt.Println("\nВыберите действие:")
				fmt.Println("1. Сгенерировать ключи")
				fmt.Println("2. Зашифровать текст")
				fmt.Println("3. Расшифровать текст")
				fmt.Println("4. Выход")

				var choice int
				fmt.Scan(&choice)

				switch choice {
				case 1:
					p, α, β, a = generateKeys(256)

					fmt.Println("Параметры шифрования:")
					fmt.Println("Закрытый ключ a:", a)
					fmt.Println("p:", p)
					fmt.Println("α:", α)
					fmt.Println("β:", β)

				case 2:
					fmt.Print("Сообщение для шифрования: ")
					text := "Hello, El Gamal!"
					fmt.Println(text)
					ciphertext = new(big.Int).SetBytes([]byte(text))
					c1, c2 = encryptELG(ciphertext, p, α, β)
					fmt.Println("Числовое представление:")
					fmt.Println("c1:", c1)
					fmt.Println("c2:", c2)
				case 3:
					decryptedText := decryptELG(c1, c2, a, p)
					fmt.Println("Расшифрованный текст:", string(decryptedText.Bytes()))
				case 4:
					fmt.Println("Выход из программы.")
					return
				default:
					fmt.Println("Неверный выбор. Пожалуйста, выберите действие от 1 до 4.")
				}
			}

		default:
			fmt.Println("Неверный вариант")
		}
	}

}
