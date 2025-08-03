package logging

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashEmail(email string) string {
	h := sha256.Sum256([]byte(email))
	return hex.EncodeToString(h[:])
}
