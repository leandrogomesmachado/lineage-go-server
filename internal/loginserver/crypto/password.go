package crypto

import (
	"crypto/sha256"
	"encoding/base64"
)

func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func VerifyPassword(password, hash string) bool {
	return HashPassword(password) == hash
}
