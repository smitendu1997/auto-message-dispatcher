package utils

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/smitendu1997/auto-message-dispatcher/utils/db"
	"github.com/smitendu1997/auto-message-dispatcher/utils/redis"
)

type Connections struct {
	DB    *db.MySQLDB
	Redis *redis.RedisClient
}

type ResponseJSON struct {
	Code  string      `json:"code"`
	Msg   string      `json:"msg"`
	Model interface{} `json:"model"`
}

func ResponseWithModel(code string, msg string, model interface{}) ResponseJSON {
	return ResponseJSON{
		Code:  code,
		Msg:   msg,
		Model: model,
	}
}

func SHA256Hash(input string) string {
	// Create a new SHA256 hash
	hasher := sha256.New()

	// Write the input string to the hash
	hasher.Write([]byte(input))

	// Get the hash result as a byte slice
	hashBytes := hasher.Sum(nil)

	// Convert the byte slice to a hex string
	hashString := hex.EncodeToString(hashBytes)

	return hashString
}

func InArraySlice(array []string, value string) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}
