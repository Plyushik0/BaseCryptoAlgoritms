package main

import (
	"encoding/binary"
)

const (
	BlockSize = 64
)

var (
	pi [256]byte = [256]byte{
		0xfc, 0xee, 0xdd, 0x11, 0xcf, 0x6e, 0x31, 0x16,
		0xfb, 0xc4, 0xfa, 0xda, 0x23, 0xc5, 0x04, 0x4d,
		0xe9, 0x77, 0xf0, 0xdb, 0x93, 0x2e, 0x99, 0xba,
		0x17, 0x36, 0xf1, 0xbb, 0x14, 0xcd, 0x5f, 0xc1,
		0xf9, 0x18, 0x65, 0x5a, 0xe2, 0x5c, 0xef, 0x21,
		0x81, 0x1c, 0x3c, 0x42, 0x8b, 0x01, 0x8e, 0x4f,
		0x05, 0x84, 0x02, 0xae, 0xe3, 0x6a, 0x8f, 0xa0,
		0x06, 0x0b, 0xed, 0x98, 0x7f, 0xd4, 0xd3, 0x1f,
		0xeb, 0x34, 0x2c, 0x51, 0xea, 0xc8, 0x48, 0xab,
		0xf2, 0x2a, 0x68, 0xa2, 0xfd, 0x3a, 0xce, 0xcc,
		0xb5, 0x70, 0x0e, 0x56, 0x08, 0x0c, 0x76, 0x12,
		0xbf, 0x72, 0x13, 0x47, 0x9c, 0xb7, 0x5d, 0x87,
		0x15, 0xa1, 0x96, 0x29, 0x10, 0x7b, 0x9a, 0xc7,
		0xf3, 0x91, 0x78, 0x6f, 0x9d, 0x9e, 0xb2, 0xb1,
		0x32, 0x75, 0x19, 0x3d, 0xff, 0x35, 0x8a, 0x7e,
		0x6d, 0x54, 0xc6, 0x80, 0xc3, 0xbd, 0x0d, 0x57,
		0xdf, 0xf5, 0x24, 0xa9, 0x3e, 0xa8, 0x43, 0xc9,
		0xd7, 0x79, 0xd6, 0xf6, 0x7c, 0x22, 0xb9, 0x03,
		0xe0, 0x0f, 0xec, 0xde, 0x7a, 0x94, 0xb0, 0xbc,
		0xdc, 0xe8, 0x28, 0x50, 0x4e, 0x33, 0x0a, 0x4a,
		0xa7, 0x97, 0x60, 0x73, 0x1e, 0x00, 0x62, 0x44,
		0x1a, 0xb8, 0x38, 0x82, 0x64, 0x9f, 0x26, 0x41,
		0xad, 0x45, 0x46, 0x92, 0x27, 0x5e, 0x55, 0x2f,
		0x8c, 0xa3, 0xa5, 0x7d, 0x69, 0xd5, 0x95, 0x3b,
		0x07, 0x58, 0xb3, 0x40, 0x86, 0xac, 0x1d, 0xf7,
		0x30, 0x37, 0x6b, 0xe4, 0x88, 0xd9, 0xe7, 0x89,
		0xe1, 0x1b, 0x83, 0x49, 0x4c, 0x3f, 0xf8, 0xfe,
		0x8d, 0x53, 0xaa, 0x90, 0xca, 0xd8, 0x85, 0x61,
		0x20, 0x71, 0x67, 0xa4, 0x2d, 0x2b, 0x09, 0x5b,
		0xcb, 0x9b, 0x25, 0xd0, 0xbe, 0xe5, 0x6c, 0x52,
		0x59, 0xa6, 0x74, 0xd2, 0xe6, 0xf4, 0xb4, 0xc0,
		0xd1, 0x66, 0xaf, 0xc2, 0x39, 0x4b, 0x63, 0xb6,
	}
	tau [64]int = [64]int{
		0x00, 0x08, 0x10, 0x18, 0x20, 0x28, 0x30, 0x38,
		0x01, 0x09, 0x11, 0x19, 0x21, 0x29, 0x31, 0x39,
		0x02, 0x0a, 0x12, 0x1a, 0x22, 0x2a, 0x32, 0x3a,
		0x03, 0x0b, 0x13, 0x1b, 0x23, 0x2b, 0x33, 0x3b,
		0x04, 0x0c, 0x14, 0x1c, 0x24, 0x2c, 0x34, 0x3c,
		0x05, 0x0d, 0x15, 0x1d, 0x25, 0x2d, 0x35, 0x3d,
		0x06, 0x0e, 0x16, 0x1e, 0x26, 0x2e, 0x36, 0x3e,
		0x07, 0x0f, 0x17, 0x1f, 0x27, 0x2f, 0x37, 0x3f,
	}
	c [12][BlockSize]byte = [12][BlockSize]byte{
		[BlockSize]byte{
			0x07, 0x45, 0xa6, 0xf2, 0x59, 0x65, 0x80, 0xdd,
			0x23, 0x4d, 0x74, 0xcc, 0x36, 0x74, 0x76, 0x05,
			0x15, 0xd3, 0x60, 0xa4, 0x08, 0x2a, 0x42, 0xa2,
			0x01, 0x69, 0x67, 0x92, 0x91, 0xe0, 0x7c, 0x4b,
			0xfc, 0xc4, 0x85, 0x75, 0x8d, 0xb8, 0x4e, 0x71,
			0x16, 0xd0, 0x45, 0x2e, 0x43, 0x76, 0x6a, 0x2f,
			0x1f, 0x7c, 0x65, 0xc0, 0x81, 0x2f, 0xcb, 0xeb,
			0xe9, 0xda, 0xca, 0x1e, 0xda, 0x5b, 0x08, 0xb1,
		},
		[BlockSize]byte{
			0xb7, 0x9b, 0xb1, 0x21, 0x70, 0x04, 0x79, 0xe6,
			0x56, 0xcd, 0xcb, 0xd7, 0x1b, 0xa2, 0xdd, 0x55,
			0xca, 0xa7, 0x0a, 0xdb, 0xc2, 0x61, 0xb5, 0x5c,
			0x58, 0x99, 0xd6, 0x12, 0x6b, 0x17, 0xb5, 0x9a,
			0x31, 0x01, 0xb5, 0x16, 0x0f, 0x5e, 0xd5, 0x61,
			0x98, 0x2b, 0x23, 0x0a, 0x72, 0xea, 0xfe, 0xf3,
			0xd7, 0xb5, 0x70, 0x0f, 0x46, 0x9d, 0xe3, 0x4f,
			0x1a, 0x2f, 0x9d, 0xa9, 0x8a, 0xb5, 0xa3, 0x6f,
		},
		[BlockSize]byte{
			0xb2, 0x0a, 0xba, 0x0a, 0xf5, 0x96, 0x1e, 0x99,
			0x31, 0xdb, 0x7a, 0x86, 0x43, 0xf4, 0xb6, 0xc2,
			0x09, 0xdb, 0x62, 0x60, 0x37, 0x3a, 0xc9, 0xc1,
			0xb1, 0x9e, 0x35, 0x90, 0xe4, 0x0f, 0xe2, 0xd3,
			0x7b, 0x7b, 0x29, 0xb1, 0x14, 0x75, 0xea, 0xf2,
			0x8b, 0x1f, 0x9c, 0x52, 0x5f, 0x5e, 0xf1, 0x06,
			0x35, 0x84, 0x3d, 0x6a, 0x28, 0xfc, 0x39, 0x0a,
			0xc7, 0x2f, 0xce, 0x2b, 0xac, 0xdc, 0x74, 0xf5,
		},
		[BlockSize]byte{
			0x2e, 0xd1, 0xe3, 0x84, 0xbc, 0xbe, 0x0c, 0x22,
			0xf1, 0x37, 0xe8, 0x93, 0xa1, 0xea, 0x53, 0x34,
			0xbe, 0x03, 0x52, 0x93, 0x33, 0x13, 0xb7, 0xd8,
			0x75, 0xd6, 0x03, 0xed, 0x82, 0x2c, 0xd7, 0xa9,
			0x3f, 0x35, 0x5e, 0x68, 0xad, 0x1c, 0x72, 0x9d,
			0x7d, 0x3c, 0x5c, 0x33, 0x7e, 0x85, 0x8e, 0x48,
			0xdd, 0xe4, 0x71, 0x5d, 0xa0, 0xe1, 0x48, 0xf9,
			0xd2, 0x66, 0x15, 0xe8, 0xb3, 0xdf, 0x1f, 0xef,
		},
		[BlockSize]byte{
			0x57, 0xfe, 0x6c, 0x7c, 0xfd, 0x58, 0x17, 0x60,
			0xf5, 0x63, 0xea, 0xa9, 0x7e, 0xa2, 0x56, 0x7a,
			0x16, 0x1a, 0x27, 0x23, 0xb7, 0x00, 0xff, 0xdf,
			0xa3, 0xf5, 0x3a, 0x25, 0x47, 0x17, 0xcd, 0xbf,
			0xbd, 0xff, 0x0f, 0x80, 0xd7, 0x35, 0x9e, 0x35,
			0x4a, 0x10, 0x86, 0x16, 0x1f, 0x1c, 0x15, 0x7f,
			0x63, 0x23, 0xa9, 0x6c, 0x0c, 0x41, 0x3f, 0x9a,
			0x99, 0x47, 0x47, 0xad, 0xac, 0x6b, 0xea, 0x4b,
		},
		[BlockSize]byte{
			0x6e, 0x7d, 0x64, 0x46, 0x7a, 0x40, 0x68, 0xfa,
			0x35, 0x4f, 0x90, 0x36, 0x72, 0xc5, 0x71, 0xbf,
			0xb6, 0xc6, 0xbe, 0xc2, 0x66, 0x1f, 0xf2, 0x0a,
			0xb4, 0xb7, 0x9a, 0x1c, 0xb7, 0xa6, 0xfa, 0xcf,
			0xc6, 0x8e, 0xf0, 0x9a, 0xb4, 0x9a, 0x7f, 0x18,
			0x6c, 0xa4, 0x42, 0x51, 0xf9, 0xc4, 0x66, 0x2d,
			0xc0, 0x39, 0x30, 0x7a, 0x3b, 0xc3, 0xa4, 0x6f,
			0xd9, 0xd3, 0x3a, 0x1d, 0xae, 0xae, 0x4f, 0xae,
		},
		[BlockSize]byte{
			0x93, 0xd4, 0x14, 0x3a, 0x4d, 0x56, 0x86, 0x88,
			0xf3, 0x4a, 0x3c, 0xa2, 0x4c, 0x45, 0x17, 0x35,
			0x04, 0x05, 0x4a, 0x28, 0x83, 0x69, 0x47, 0x06,
			0x37, 0x2c, 0x82, 0x2d, 0xc5, 0xab, 0x92, 0x09,
			0xc9, 0x93, 0x7a, 0x19, 0x33, 0x3e, 0x47, 0xd3,
			0xc9, 0x87, 0xbf, 0xe6, 0xc7, 0xc6, 0x9e, 0x39,
			0x54, 0x09, 0x24, 0xbf, 0xfe, 0x86, 0xac, 0x51,
			0xec, 0xc5, 0xaa, 0xee, 0x16, 0x0e, 0xc7, 0xf4,
		},
		[BlockSize]byte{
			0x1e, 0xe7, 0x02, 0xbf, 0xd4, 0x0d, 0x7f, 0xa4,
			0xd9, 0xa8, 0x51, 0x59, 0x35, 0xc2, 0xac, 0x36,
			0x2f, 0xc4, 0xa5, 0xd1, 0x2b, 0x8d, 0xd1, 0x69,
			0x90, 0x06, 0x9b, 0x92, 0xcb, 0x2b, 0x89, 0xf4,
			0x9a, 0xc4, 0xdb, 0x4d, 0x3b, 0x44, 0xb4, 0x89,
			0x1e, 0xde, 0x36, 0x9c, 0x71, 0xf8, 0xb7, 0x4e,
			0x41, 0x41, 0x6e, 0x0c, 0x02, 0xaa, 0xe7, 0x03,
			0xa7, 0xc9, 0x93, 0x4d, 0x42, 0x5b, 0x1f, 0x9b,
		},
		[BlockSize]byte{
			0xdb, 0x5a, 0x23, 0x83, 0x51, 0x44, 0x61, 0x72,
			0x60, 0x2a, 0x1f, 0xcb, 0x92, 0xdc, 0x38, 0x0e,
			0x54, 0x9c, 0x07, 0xa6, 0x9a, 0x8a, 0x2b, 0x7b,
			0xb1, 0xce, 0xb2, 0xdb, 0x0b, 0x44, 0x0a, 0x80,
			0x84, 0x09, 0x0d, 0xe0, 0xb7, 0x55, 0xd9, 0x3c,
			0x24, 0x42, 0x89, 0x25, 0x1b, 0x3a, 0x7d, 0x3a,
			0xde, 0x5f, 0x16, 0xec, 0xd8, 0x9a, 0x4c, 0x94,
			0x9b, 0x22, 0x31, 0x16, 0x54, 0x5a, 0x8f, 0x37,
		},
		[BlockSize]byte{
			0xed, 0x9c, 0x45, 0x98, 0xfb, 0xc7, 0xb4, 0x74,
			0xc3, 0xb6, 0x3b, 0x15, 0xd1, 0xfa, 0x98, 0x36,
			0xf4, 0x52, 0x76, 0x3b, 0x30, 0x6c, 0x1e, 0x7a,
			0x4b, 0x33, 0x69, 0xaf, 0x02, 0x67, 0xe7, 0x9f,
			0x03, 0x61, 0x33, 0x1b, 0x8a, 0xe1, 0xff, 0x1f,
			0xdb, 0x78, 0x8a, 0xff, 0x1c, 0xe7, 0x41, 0x89,
			0xf3, 0xf3, 0xe4, 0xb2, 0x48, 0xe5, 0x2a, 0x38,
			0x52, 0x6f, 0x05, 0x80, 0xa6, 0xde, 0xbe, 0xab,
		},
		[BlockSize]byte{
			0x1b, 0x2d, 0xf3, 0x81, 0xcd, 0xa4, 0xca, 0x6b,
			0x5d, 0xd8, 0x6f, 0xc0, 0x4a, 0x59, 0xa2, 0xde,
			0x98, 0x6e, 0x47, 0x7d, 0x1d, 0xcd, 0xba, 0xef,
			0xca, 0xb9, 0x48, 0xea, 0xef, 0x71, 0x1d, 0x8a,
			0x79, 0x66, 0x84, 0x14, 0x21, 0x80, 0x01, 0x20,
			0x61, 0x07, 0xab, 0xeb, 0xbb, 0x6b, 0xfa, 0xd8,
			0x94, 0xfe, 0x5a, 0x63, 0xcd, 0xc6, 0x02, 0x30,
			0xfb, 0x89, 0xc8, 0xef, 0xd0, 0x9e, 0xcd, 0x7b,
		},
		[BlockSize]byte{
			0x20, 0xd7, 0x1b, 0xf1, 0x4a, 0x92, 0xbc, 0x48,
			0x99, 0x1b, 0xb2, 0xd9, 0xd5, 0x17, 0xf4, 0xfa,
			0x52, 0x28, 0xe1, 0x88, 0xaa, 0xa4, 0x1d, 0xe7,
			0x86, 0xcc, 0x91, 0x18, 0x9d, 0xef, 0x80, 0x5d,
			0x9b, 0x9f, 0x21, 0x30, 0xd4, 0x12, 0x20, 0xf8,
			0x77, 0x1d, 0xdf, 0xbc, 0x32, 0x3c, 0xa4, 0xcd,
			0x7a, 0xb1, 0x49, 0x04, 0xb0, 0x80, 0x13, 0xd2,
			0xba, 0x31, 0x16, 0xf1, 0x67, 0xe7, 0x8e, 0x37,
		},
	}
	a [64]uint64 
)

func init() {
	as := [64][]byte{
		[]byte{0x8e, 0x20, 0xfa, 0xa7, 0x2b, 0xa0, 0xb4, 0x70},
		[]byte{0x47, 0x10, 0x7d, 0xdd, 0x9b, 0x50, 0x5a, 0x38},
		[]byte{0xad, 0x08, 0xb0, 0xe0, 0xc3, 0x28, 0x2d, 0x1c},
		[]byte{0xd8, 0x04, 0x58, 0x70, 0xef, 0x14, 0x98, 0x0e},
		[]byte{0x6c, 0x02, 0x2c, 0x38, 0xf9, 0x0a, 0x4c, 0x07},
		[]byte{0x36, 0x01, 0x16, 0x1c, 0xf2, 0x05, 0x26, 0x8d},
		[]byte{0x1b, 0x8e, 0x0b, 0x0e, 0x79, 0x8c, 0x13, 0xc8},
		[]byte{0x83, 0x47, 0x8b, 0x07, 0xb2, 0x46, 0x87, 0x64},
		[]byte{0xa0, 0x11, 0xd3, 0x80, 0x81, 0x8e, 0x8f, 0x40},
		[]byte{0x50, 0x86, 0xe7, 0x40, 0xce, 0x47, 0xc9, 0x20},
		[]byte{0x28, 0x43, 0xfd, 0x20, 0x67, 0xad, 0xea, 0x10},
		[]byte{0x14, 0xaf, 0xf0, 0x10, 0xbd, 0xd8, 0x75, 0x08},
		[]byte{0x0a, 0xd9, 0x78, 0x08, 0xd0, 0x6c, 0xb4, 0x04},
		[]byte{0x05, 0xe2, 0x3c, 0x04, 0x68, 0x36, 0x5a, 0x02},
		[]byte{0x8c, 0x71, 0x1e, 0x02, 0x34, 0x1b, 0x2d, 0x01},
		[]byte{0x46, 0xb6, 0x0f, 0x01, 0x1a, 0x83, 0x98, 0x8e},
		[]byte{0x90, 0xda, 0xb5, 0x2a, 0x38, 0x7a, 0xe7, 0x6f},
		[]byte{0x48, 0x6d, 0xd4, 0x15, 0x1c, 0x3d, 0xfd, 0xb9},
		[]byte{0x24, 0xb8, 0x6a, 0x84, 0x0e, 0x90, 0xf0, 0xd2},
		[]byte{0x12, 0x5c, 0x35, 0x42, 0x07, 0x48, 0x78, 0x69},
		[]byte{0x09, 0x2e, 0x94, 0x21, 0x8d, 0x24, 0x3c, 0xba},
		[]byte{0x8a, 0x17, 0x4a, 0x9e, 0xc8, 0x12, 0x1e, 0x5d},
		[]byte{0x45, 0x85, 0x25, 0x4f, 0x64, 0x09, 0x0f, 0xa0},
		[]byte{0xac, 0xcc, 0x9c, 0xa9, 0x32, 0x8a, 0x89, 0x50},
		[]byte{0x9d, 0x4d, 0xf0, 0x5d, 0x5f, 0x66, 0x14, 0x51},
		[]byte{0xc0, 0xa8, 0x78, 0xa0, 0xa1, 0x33, 0x0a, 0xa6},
		[]byte{0x60, 0x54, 0x3c, 0x50, 0xde, 0x97, 0x05, 0x53},
		[]byte{0x30, 0x2a, 0x1e, 0x28, 0x6f, 0xc5, 0x8c, 0xa7},
		[]byte{0x18, 0x15, 0x0f, 0x14, 0xb9, 0xec, 0x46, 0xdd},
		[]byte{0x0c, 0x84, 0x89, 0x0a, 0xd2, 0x76, 0x23, 0xe0},
		[]byte{0x06, 0x42, 0xca, 0x05, 0x69, 0x3b, 0x9f, 0x70},
		[]byte{0x03, 0x21, 0x65, 0x8c, 0xba, 0x93, 0xc1, 0x38},
		[]byte{0x86, 0x27, 0x5d, 0xf0, 0x9c, 0xe8, 0xaa, 0xa8},
		[]byte{0x43, 0x9d, 0xa0, 0x78, 0x4e, 0x74, 0x55, 0x54},
		[]byte{0xaf, 0xc0, 0x50, 0x3c, 0x27, 0x3a, 0xa4, 0x2a},
		[]byte{0xd9, 0x60, 0x28, 0x1e, 0x9d, 0x1d, 0x52, 0x15},
		[]byte{0xe2, 0x30, 0x14, 0x0f, 0xc0, 0x80, 0x29, 0x84},
		[]byte{0x71, 0x18, 0x0a, 0x89, 0x60, 0x40, 0x9a, 0x42},
		[]byte{0xb6, 0x0c, 0x05, 0xca, 0x30, 0x20, 0x4d, 0x21},
		[]byte{0x5b, 0x06, 0x8c, 0x65, 0x18, 0x10, 0xa8, 0x9e},
		[]byte{0x45, 0x6c, 0x34, 0x88, 0x7a, 0x38, 0x05, 0xb9},
		[]byte{0xac, 0x36, 0x1a, 0x44, 0x3d, 0x1c, 0x8c, 0xd2},
		[]byte{0x56, 0x1b, 0x0d, 0x22, 0x90, 0x0e, 0x46, 0x69},
		[]byte{0x2b, 0x83, 0x88, 0x11, 0x48, 0x07, 0x23, 0xba},
		[]byte{0x9b, 0xcf, 0x44, 0x86, 0x24, 0x8d, 0x9f, 0x5d},
		[]byte{0xc3, 0xe9, 0x22, 0x43, 0x12, 0xc8, 0xc1, 0xa0},
		[]byte{0xef, 0xfa, 0x11, 0xaf, 0x09, 0x64, 0xee, 0x50},
		[]byte{0xf9, 0x7d, 0x86, 0xd9, 0x8a, 0x32, 0x77, 0x28},
		[]byte{0xe4, 0xfa, 0x20, 0x54, 0xa8, 0x0b, 0x32, 0x9c},
		[]byte{0x72, 0x7d, 0x10, 0x2a, 0x54, 0x8b, 0x19, 0x4e},
		[]byte{0x39, 0xb0, 0x08, 0x15, 0x2a, 0xcb, 0x82, 0x27},
		[]byte{0x92, 0x58, 0x04, 0x84, 0x15, 0xeb, 0x41, 0x9d},
		[]byte{0x49, 0x2c, 0x02, 0x42, 0x84, 0xfb, 0xae, 0xc0},
		[]byte{0xaa, 0x16, 0x01, 0x21, 0x42, 0xf3, 0x57, 0x60},
		[]byte{0x55, 0x0b, 0x8e, 0x9e, 0x21, 0xf7, 0xa5, 0x30},
		[]byte{0xa4, 0x8b, 0x47, 0x4f, 0x9e, 0xf5, 0xdc, 0x18},
		[]byte{0x70, 0xa6, 0xa5, 0x6e, 0x24, 0x40, 0x59, 0x8e},
		[]byte{0x38, 0x53, 0xdc, 0x37, 0x12, 0x20, 0xa2, 0x47},
		[]byte{0x1c, 0xa7, 0x6e, 0x95, 0x09, 0x10, 0x51, 0xad},
		[]byte{0x0e, 0xdd, 0x37, 0xc4, 0x8a, 0x08, 0xa6, 0xd8},
		[]byte{0x07, 0xe0, 0x95, 0x62, 0x45, 0x04, 0x53, 0x6c},
		[]byte{0x8d, 0x70, 0xc4, 0x31, 0xac, 0x02, 0xa7, 0x36},
		[]byte{0xc8, 0x38, 0x62, 0x96, 0x56, 0x01, 0xdd, 0x1b},
		[]byte{0x64, 0x1c, 0x31, 0x4b, 0x2b, 0x8e, 0xe0, 0x83},
	}
	for i := 0; i < 64; i++ {
		a[i] = binary.BigEndian.Uint64(as[i])
	}
}

type Hash struct {
	size int
	buf  []byte
	n    uint64
	hsh  *[BlockSize]byte
	chk  *[BlockSize]byte
	tmp  *[BlockSize]byte
}
func StreebogHash(data []byte, size int) []byte {
    h := New(size / 8) 
    h.Write(data)      
    return h.Sum(nil) 
}


func New(size int) *Hash {
	if size != 32 && size != 64 {
		panic("размер должен быть 32 или 64")
	}
	h := Hash{
		size: size,
		hsh:  new([BlockSize]byte),
		chk:  new([BlockSize]byte),
		tmp:  new([BlockSize]byte),
	}
	h.Reset()
	return &h
}

func (h *Hash) Reset() {
	h.n = 0
	h.buf = nil
	for i := 0; i < BlockSize; i++ {
		h.chk[i] = 0
		if h.size == 32 {
			h.hsh[i] = 1
		} else {
			h.hsh[i] = 0
		}
	}
}

func (h *Hash) BlockSize() int {
	return BlockSize
}

func (h *Hash) Size() int {
	return h.size
}

func (h *Hash) Write(data []byte) (int) {
	h.buf = append(h.buf, data...)
	for len(h.buf) >= BlockSize {
		copy(h.tmp[:], h.buf[:BlockSize])
		h.hsh = g(h.n, h.hsh, h.tmp) // обрабатываем блок
		h.chk = add512bit(h.chk, h.tmp) // обновляем контрольную сумму
		h.n += BlockSize * 8 // увеличиваем счетчик обработанных битов
		h.buf = h.buf[BlockSize:]
	}
	return len(data)
}

func (h *Hash) Sum(in []byte) []byte {
	buf := new([BlockSize]byte)
	copy(h.tmp[:], buf[:])
	copy(buf[:], h.buf[:])
	buf[len(h.buf)] = 1 // добавляем 1 в конец блока
	hsh := g(h.n, h.hsh, buf) // обрабатываем последний блок
	binary.LittleEndian.PutUint64(h.tmp[:], h.n+uint64(len(h.buf))*8) // добавляем длину сообщения в битах
	hsh = g(0, hsh, h.tmp)
	hsh = g(0, hsh, add512bit(h.chk, buf)) // добавляем контрольную сумму
	if h.size == 32 {
		return append(in, hsh[BlockSize/2:]...)
	}
	return append(in, hsh[:]...)
}

func add512bit(chk, data *[BlockSize]byte) *[BlockSize]byte {
	var ss uint16
	r := new([BlockSize]byte)
	for i := 0; i < BlockSize; i++ {
		ss = uint16(chk[i]) + uint16(data[i]) + (ss >> 8) // скаладываются 2 байта и добавляется перенос от предыдущей опреации
		r[i] = byte(0xFF & ss) // оставляем только младшие 8 бит
	}
	return r
}

// основная функция сжатия
// n - номер блока
// hsh - текущее состояние хэша
func g(n uint64, hsh, data *[BlockSize]byte) *[BlockSize]byte {
	ns := make([]byte, 8)
	binary.LittleEndian.PutUint64(ns, n) // LittleEndian - порядок байт, в котором младший байт идет первым
	r := new([BlockSize]byte)
	for i := 0; i < 8; i++ {
		r[i] = hsh[i] ^ ns[i] // XOR первых 8 байт состояния хэша с номером блока
	}
	copy(r[8:], hsh[8:])
	return blockXor(blockXor(e(l(ps(r)), data), hsh), data)
}

func e(k, msg *[BlockSize]byte) *[BlockSize]byte {
	for i := 0; i < 12; i++ {
		msg = l(ps(blockXor(k, msg)))
		k = l(ps(blockXor(k, &c[i])))
	}
	return blockXor(k, msg)
}

func blockXor(x, y *[BlockSize]byte) *[BlockSize]byte {
	r := new([BlockSize]byte)
	for i := 0; i < BlockSize; i++ {
		r[i] = x[i] ^ y[i]
	}
	return r
}

func ps(data *[BlockSize]byte) *[BlockSize]byte {
	r := new([BlockSize]byte)
	for i := 0; i < BlockSize; i++ {
		r[tau[i]] = pi[int(data[i])]
	}
	return r
}

func l(data *[BlockSize]byte) *[BlockSize]byte {
	r := new([BlockSize]byte)
    for i := 0; i < 8; i++ {
        val := binary.LittleEndian.Uint64(data[i*8 : (i+1)*8]) //Блок данных делится на 8 частей по 8 байт (64 бита каждая)
        var res64 uint64

        for _, mask := range a {
            if val&(1<<63) != 0 { 
                res64 ^= mask
            }
            val <<= 1 
        }

        binary.LittleEndian.PutUint64(r[i*8:(i+1)*8], res64)
    }

    return r
}