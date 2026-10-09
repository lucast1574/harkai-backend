package push

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"strings"
)

func validCredentials(project string, raw []byte) bool {
	var v struct {
		Type     string `json:"type"`
		Project  string `json:"project_id"`
		TokenURI string `json:"token_uri"`
		Email    string `json:"client_email"`
		Key      string `json:"private_key"`
	}
	if len(raw) > 65536 || project == "" || strings.ContainsAny(project, "/?#") || json.Unmarshal(raw, &v) != nil || v.Type != "service_account" || v.Project != project || v.TokenURI != "https://oauth2.googleapis.com/token" || !strings.HasSuffix(v.Email, "@"+project+".iam.gserviceaccount.com") {
		return false
	}
	block, _ := pem.Decode([]byte(v.Key))
	if block == nil {
		return false
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return false
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	return ok && rsaKey.N.BitLen() >= 2048
}
