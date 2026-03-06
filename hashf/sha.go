package hashf


import (
    "encoding/binary"
	"math/bits"
)

var (
	sha256K = [64]uint32{
		0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5,
		0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
		0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3,
		0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
		0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc,
		0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
		0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7,
		0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
		0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13,
		0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
		0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3,
		0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
		0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5,
		0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
		0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208,
		0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
	}

	sha512K = [80]uint64{
		0x428a2f98d728ae22, 0x7137449123ef65cd, 0xb5c0fbcfec4d3b2f, 0xe9b5dba58189dbbc,
		0x3956c25bf348b538, 0x59f111f1b605d019, 0x923f82a4af194f9b, 0xab1c5ed5da6d8118,
		0xd807aa98a3030242, 0x12835b0145706fbe, 0x243185be4ee4b28c, 0x550c7dc3d5ffb4e2,
		0x72be5d74f27b896f, 0x80deb1fe3b1696b1, 0x9bdc06a725c71235, 0xc19bf174cf692694,
		0xe49b69c19ef14ad2, 0xefbe4786384f25e3, 0x0fc19dc68b8cd5b5, 0x240ca1cc77ac9c65,
		0x2de92c6f592b0275, 0x4a7484aa6ea6e483, 0x5cb0a9dcbd41fbd4, 0x76f988da831153b5,
		0x983e5152ee66dfab, 0xa831c66d2db43210, 0xb00327c898fb213f, 0xbf597fc7beef0ee4,
		0xc6e00bf33da88fc2, 0xd5a79147930aa725, 0x06ca6351e003826f, 0x142929670a0e6e70,
		0x27b70a8546d22ffc, 0x2e1b21385c26c926, 0x4d2c6dfc5ac42aed, 0x53380d139d95b3df,
		0x650a73548baf63de, 0x766a0abb3c77b2a8, 0x81c2c92e47edaee6, 0x92722c851482353b,
		0xa2bfe8a14cf10364, 0xa81a664bbc423001, 0xc24b8b70d0f89791, 0xc76c51a30654be30,
		0xd192e819d6ef5218, 0xd69906245565a910, 0xf40e35855771202a, 0x106aa07032bbd1b8,
		0x19a4c116b8d2d0c8, 0x1e376c085141ab53, 0x2748774cdf8eeb99, 0x34b0bcb5e19b48a8,
		0x391c0cb3c5c95a63, 0x4ed8aa4ae3418acb, 0x5b9cca4f7763e373, 0x682e6ff3d6b2b8a3,
		0x748f82ee5defb2fc, 0x78a5636f43172f60, 0x84c87814a1f0ab72, 0x8cc702081a6439ec,
		0x90befffa23631e28, 0xa4506cebde82bde9, 0xbef9a3f7b2c67915, 0xc67178f2e372532b,
		0xca273eceea26619c, 0xd186b8c721c0c207, 0xeada7dd6cde0eb1e, 0xf57d4f7fee6ed178,
		0x06f067aa72176fba, 0x0a637dc5a2c898a6, 0x113f9804bef90dae, 0x1b710b35131c471b,
		0x28db77f523047d84, 0x32caab7b40c72493, 0x3c9ebe0a15c9bebc, 0x431d67c49c100d4c,
		0x4cc5d4becb3e42b6, 0x597f299cfc657e2a, 0x5fcb6fab3ad6faec, 0x6c44198c4a475817,
	}

	sha256H = [8]uint32{
		0x6a09e667,
		0xbb67ae85,
		0x3c6ef372,
		0xa54ff53a,
		0x510e527f,
		0x9b05688c,
		0x1f83d9ab,
		0x5be0cd19,
	}

	sha512H = [8]uint64{
		0x6a09e667f3bcc908,
		0xbb67ae8584caa73b,
		0x3c6ef372fe94f82b,
		0xa54ff53a5f1d36f1,
		0x510e527fade682d1,
		0x9b05688c2b3e6c1f,
		0x1f83d9abfb41bd6b,
		0x5be0cd19137e2179,
	}
)

func ch256(x, y, z uint32) uint32 { return (x & y) ^ (^x & z) }
func maj256(x, y, z uint32) uint32 { return (x & y) ^ (x & z) ^ (y & z) }
func sigma0_256(x uint32) uint32 { return bits.RotateLeft32(x, -2) ^ bits.RotateLeft32(x, -13) ^ bits.RotateLeft32(x, -22) } //циклический сдвиг
func sigma1_256(x uint32) uint32 { return bits.RotateLeft32(x, -6) ^ bits.RotateLeft32(x, -11) ^ bits.RotateLeft32(x, -25) }
func gamma0_256(x uint32) uint32 { return bits.RotateLeft32(x, -7) ^ bits.RotateLeft32(x, -18) ^ (x >> 3) }
func gamma1_256(x uint32) uint32 { return bits.RotateLeft32(x, -17) ^ bits.RotateLeft32(x, -19) ^ (x >> 10) }

func ch512(x, y, z uint64) uint64 { return (x & y) ^ (^x & z) }
func maj512(x, y, z uint64) uint64 { return (x & y) ^ (x & z) ^ (y & z) }
func sigma0_512(x uint64) uint64 { return bits.RotateLeft64(x, -28) ^ bits.RotateLeft64(x, -34) ^ bits.RotateLeft64(x, -39) }
func sigma1_512(x uint64) uint64 { return bits.RotateLeft64(x, -14) ^ bits.RotateLeft64(x, -18) ^ bits.RotateLeft64(x, -41) }
func gamma0_512(x uint64) uint64 { return bits.RotateLeft64(x, -1) ^ bits.RotateLeft64(x, -8) ^ (x >> 7) }
func gamma1_512(x uint64) uint64 { return bits.RotateLeft64(x, -19) ^ bits.RotateLeft64(x, -61) ^ (x >> 6) }

func Sha256(message []byte) [32]byte {
	m := make([]byte, len(message))
	copy(m, message)
	m = append(m, 0x80)
	for len(m)%64 != 56 {
		m = append(m, 0x00)
	}
	length := uint64(len(message)) * 8
	lengthBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(lengthBytes, length)
	m = append(m, lengthBytes...)

	h := sha256H

	for chunkStart := 0; chunkStart < len(m); chunkStart += 64 {
		var w [64]uint32


		for i := 0; i < 16; i++ {
			w[i] = binary.BigEndian.Uint32(m[chunkStart+i*4:])
		}

		for i := 16; i < 64; i++ {
			w[i] = w[i-16] + gamma0_256(w[i-15]) + w[i-7] + gamma1_256(w[i-2])
		}

		a, b, c, d, e, f, g, hh := h[0], h[1], h[2], h[3], h[4], h[5], h[6], h[7]

		for i := 0; i < 64; i++ {
			temp1 := hh + sigma1_256(e) + ch256(e, f, g) + sha256K[i] + w[i]
			temp2 := sigma0_256(a) + maj256(a, b, c)

			hh = g
			g = f
			f = e
			e = d + temp1
			d = c
			c = b
			b = a
			a = temp1 + temp2
		}

		h[0] += a
		h[1] += b
		h[2] += c
		h[3] += d
		h[4] += e
		h[5] += f
		h[6] += g
		h[7] += hh
	}

	var digest [32]byte
	for i, s := range h {
		binary.BigEndian.PutUint32(digest[i*4:], s)
	}
	return digest
}

func Sha512(message []byte) [64]byte {
	m := make([]byte, len(message))
	copy(m, message)
	m = append(m, 0x80)
	for len(m)%128 != 112 {
		m = append(m, 0x00)
	}
	length := uint64(len(message)) * 8
	lengthBytes := make([]byte, 16)
	binary.BigEndian.PutUint64(lengthBytes[8:], length) 
	m = append(m, lengthBytes...)

	h := sha512H 

	for chunkStart := 0; chunkStart < len(m); chunkStart += 128 {
		var w [80]uint64


		for i := 0; i < 16; i++ {
			w[i] = binary.BigEndian.Uint64(m[chunkStart+i*8:])
		}


		for i := 16; i < 80; i++ {
			w[i] = w[i-16] + gamma0_512(w[i-15]) + w[i-7] + gamma1_512(w[i-2])
		}

		a, b, c, d, e, f, g, hh := h[0], h[1], h[2], h[3], h[4], h[5], h[6], h[7]

		for i := 0; i < 80; i++ {
			temp1 := hh + sigma1_512(e) + ch512(e, f, g) + sha512K[i] + w[i]
			temp2 := sigma0_512(a) + maj512(a, b, c)

			hh = g
			g = f
			f = e
			e = d + temp1
			d = c
			c = b
			b = a
			a = temp1 + temp2
		}

		h[0] += a
		h[1] += b
		h[2] += c
		h[3] += d
		h[4] += e
		h[5] += f
		h[6] += g
		h[7] += hh
	}

	var digest [64]byte
	for i, s := range h {
		binary.BigEndian.PutUint64(digest[i*8:], s)
	}
	return digest
}