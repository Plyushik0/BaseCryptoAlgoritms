package main

const (
	PARAMS_FILE                 = "shared_params.json"
	SPEKE_PARAMS_FILE           = "speke_params.json"
	HOST                        = "localhost"
	PORT                        = 12345 
	SPEKE_PORT                  = 65432
	PRIME_BITS                  = 1024 
	PRIME_BITS_SPEKE            = 1024
	USER_A_OTP_STATE_DIR        = "user_a_otp_states"
	USER_B_OTP_SERVER_STATE_DIR = "user_b_otp_server_states"
	OTP_HASH_FUNCTION           = "sha256"
	SPEKE_PASSWORD_HASH_FUNCTION = "sha512"
)