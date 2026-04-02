package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// sign computes the HMAC-SHA256 signature of queryString using secretKey
// and returns the hex-encoded result.
func sign(queryString, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(queryString))
	return hex.EncodeToString(mac.Sum(nil))
}
