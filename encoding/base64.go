package main



const base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

func encodeBase64(input []byte) string {
	var result string
	padding := 0

	for i := 0; i < len(input); i += 3 {
		var chunk [3]byte
		copy(chunk[:], input[i:min(i+3, len(input))])

		if len(input[i:]) < 3 {
			padding = 3 - len(input[i:])
		}

		result += string(base64Table[chunk[0]>>2])                        // сдвигаем на 2 бита вправо, оставляем 6 бит
		result += string(base64Table[((chunk[0]&0x03)<<4)|(chunk[1]>>4)]) // берем 2 младших бита (0x03) первого байта,
		// сдвигаем их влево, берём 4 старших бита второго байта, объединяем их
		if padding < 2 {
			result += string(base64Table[((chunk[1]&0x0F)<<2)|(chunk[2]>>6)]) // берём 4 младших бита (0x0F) второго байта,
			// сдвигаем их влево, берём 2 старших бита третьего байта, объединяем их
		} else {
			result += "="
		}
		if padding < 1 {
			result += string(base64Table[chunk[2]&0x3F]) // берём 6 младших бит третьего байта
		} else {
			result += "="
		}
	}

	return result
}

func decodeBase64(input string) []byte {
	var result []byte
	var temp [4]byte

	revTable := make(map[byte]int)
	for i := 0; i < len(base64Table); i++ {
		revTable[base64Table[i]] = i
	}

	for i := 0; i < len(input); i += 4 {
		for j := 0; j < 4; j++ {
			if i+j < len(input) && input[i+j] != '=' {
				temp[j] = byte(revTable[input[i+j]])
			} else {
				temp[j] = 0
			}
		}

		result = append(result, (temp[0]<<2)|(temp[1]>>4)) // Сдвигаем 6 бит из первого символа влево на 2 позиции
		// и 4 бита из второго символа вправо на 4 позиции, объединяем их
		if input[i+2] != '=' {
			result = append(result, (temp[1]<<4)|(temp[2]>>2)) // Сдвигаем 4 бита из второго символа влево на 4 позиции
			// и 2 бита из третьего символа вправо на 2 позиции, объединяем их
		}
		if input[i+3] != '=' {
			result = append(result, (temp[2]<<6)|temp[3]) // двигаем 2 младших бита из третьего символа влево на 6 позиций
			// Берём все 6 бит из четвёртого символа, объединяем их
		}

	}

	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
