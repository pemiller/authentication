package helpers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GenerateAuthCode creates a random auth code
func GenerateAuthCode() string {
	token := uuid.New().String()
	return strings.TrimRight(base64.URLEncoding.EncodeToString([]byte(token)), "=")
}

// GenerateAccessToken creates a sha of an auth code and timestamp
func GenerateAccessToken(code string) string {
	value := fmt.Sprintf("%v+%v", code, time.Now().UTC().UnixNano())
	hasher := sha256.New()
	hasher.Write([]byte(value))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
