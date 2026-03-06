package main

import "math/big"

var (
	
	PRIME = big.NewInt(2305843009213693951)
	
	HOST = "127.0.0.1"
	PORT = 65432

	DEFAULT_N = int64(5)
	DEFAULT_T = int64(3)
	
	SHARES_DIR    = "shares"
	METADATA_FILE = SHARES_DIR + "/metadata.json"
)