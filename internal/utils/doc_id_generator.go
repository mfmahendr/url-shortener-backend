package utils

import (
    "crypto/sha256"
    "encoding/base64"
)

func GenerateDocID(input string) string {
    hash := sha256.Sum256([]byte(input))
    return base64.RawURLEncoding.EncodeToString(hash[:])
}
