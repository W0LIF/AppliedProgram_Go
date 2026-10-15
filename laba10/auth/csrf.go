package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func NewCSRFToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
