package main

const base32Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

func encodeBase32(input []byte) string {
	var result string
	padding := 0

	for i := 0; i < len(input); i += 5 {
		var chunk [5]byte
		copy(chunk[:], input[i:min(i+5, len(input))])

		if len(input[i:]) < 5 {
			padding = 5 - len(input[i:])
		}

		result += string(base32Table[chunk[0]>>3]) // Берём 5 старших бит первого байта
		result += string(base32Table[((chunk[0]&0x07)<<2)|(chunk[1]>>6)]) // 3 младших бита первого байта 
		// + 2 старших бита второго
		if padding < 4 {
			result += string(base32Table[(chunk[1]>>1)&0x1F]) // 5 бит из второго байта
		} else {
			result += "="
		}
		if padding < 3 {
			result += string(base32Table[((chunk[1]&0x01)<<4)|(chunk[2]>>4)]) // 1 младший бит второго байта 
			// + 4 старших бита третьего
		} else {
			result += "="
		}
		if padding < 2 {
			result += string(base32Table[((chunk[2]&0x0F)<<1)|(chunk[3]>>7)]) // 4 младших бита третьего байта 
			// + 1 старший бит четвёртого
		} else {
			result += "="
		}
		if padding < 1 {
			result += string(base32Table[(chunk[3]>>2)&0x1F]) // 5 бит из четвёртого байта
			result += string(base32Table[((chunk[3]&0x03)<<3)|(chunk[4]>>5)]) // 2 младших бита четвёртого байта 
			// + 3 старших бита пятого
			result += string(base32Table[chunk[4]&0x1F]) // 5 младших бит пятого байта
		} else {
			result += "="
		}
	}

	return result
}

func decodeBase32(input string) []byte {
	var result []byte
	var temp [8]byte

	revTable := make(map[byte]int)
	for i := 0; i < len(base32Table); i++ {
		revTable[base32Table[i]] = i
	}


	for i := 0; i < len(input); i += 8 {
		for j := 0; j < 8; j++ {
			if i+j < len(input) && input[i+j] != '=' {
				temp[j] = byte(revTable[input[i+j]])
			} else {
				temp[j] = 0
			}
		}


		result = append(result, (temp[0]<<3)|(temp[1]>>2)) // 5 бит из первого символа + 3 старших бита второго
		if i+2 < len(input) && input[i+2] != '=' {
			result = append(result, (temp[1]<<6)|(temp[2]<<1)|(temp[3]>>4)) // 2 младших бита второго + 5 бит из третьего
		}
		if i+4 < len(input) && input[i+4] != '=' {
			result = append(result, (temp[3]<<4)|(temp[4]>>1)) // 4 младших бита третьего + 4 старших бита четвёртого
		}
		if i+5 < len(input) && input[i+5] != '=' {
			result = append(result, (temp[4]<<7)|(temp[5]<<2)|(temp[6]>>3)) // 1 младший бит четвёртого + 5 бит из пятого
		}
		if i+7 < len(input) && input[i+7] != '=' {
			result = append(result, (temp[6]<<5)|temp[7]) // 3 младших бита шестого + 5 бит из седьмого
		}
	}

	return result
}
