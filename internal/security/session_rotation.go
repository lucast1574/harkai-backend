package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// DeriveSessionTokens allows recovery without storing any bearer token in MongoDB.
func DeriveSessionTokens(previous, rotationID string) (string, string) {
	access := sessionToken(previous, "harkai/access/v1/"+rotationID)
	refresh := sessionToken(previous, "harkai/refresh/v1/"+rotationID)
	return access, refresh
}

func sessionToken(key, purpose string) string {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(purpose))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
