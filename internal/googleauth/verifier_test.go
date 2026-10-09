package googleauth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lucast1574/harkai-backend/internal/core"
)

func TestOnlyVerifiedGoogleTokensAreAccepted(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	verifier := New("test-project")
	verifier.certificates = map[string]*rsa.PublicKey{"test-key": &key.PublicKey}
	verifier.expires = time.Now().Add(time.Hour)
	valid := func() claims {
		c := claims{Subject: "google-user", Audience: "test-project", Issuer: "https://securetoken.google.com/test-project", Email: "owner@example.com", Name: "Owner", EmailVerified: true, Expires: time.Now().Add(time.Hour).Unix(), Issued: time.Now().Add(-time.Minute).Unix(), AuthTime: time.Now().Add(-time.Minute).Unix()}
		c.Firebase.Provider = "google.com"
		return c
	}
	tests := []struct {
		name           string
		change         func(*claims)
		algorithm, kid string
		ok             bool
	}{
		{"valid", func(*claims) {}, "RS256", "test-key", true},
		{"wrong audience", func(c *claims) { c.Audience = "other-project" }, "RS256", "test-key", false},
		{"wrong issuer", func(c *claims) { c.Issuer = "https://attacker.example" }, "RS256", "test-key", false},
		{"password provider", func(c *claims) { c.Firebase.Provider = "password" }, "RS256", "test-key", false},
		{"unverified email", func(c *claims) { c.EmailVerified = false }, "RS256", "test-key", false},
		{"expired", func(c *claims) { c.Expires = time.Now().Add(-time.Minute).Unix() }, "RS256", "test-key", false},
		{"future issued", func(c *claims) { c.Issued = time.Now().Add(time.Hour).Unix() }, "RS256", "test-key", false},
		{"missing subject", func(c *claims) { c.Subject = "" }, "RS256", "test-key", false},
		{"invalid email", func(c *claims) { c.Email = "invalid" }, "RS256", "test-key", false},
		{"wrong algorithm", func(*claims) {}, "HS256", "test-key", false},
		{"unknown key", func(*claims) {}, "RS256", "unknown", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := valid()
			tc.change(&c)
			token := signTestToken(t, key, c, tc.algorithm, tc.kid)
			identity, err := verifier.Verify(context.Background(), token)
			if tc.ok {
				if err != nil || identity.Subject != c.Subject || identity.Email != c.Email {
					t.Fatalf("valid token rejected: %v", err)
				}
			} else if !errors.Is(err, core.ErrUnauthorized) {
				t.Fatalf("unexpected acceptance: %v", err)
			}
		})
	}
	token := signTestToken(t, key, valid(), "RS256", "test-key")
	parts := strings.Split(token, ".")
	parts[2] = base64.RawURLEncoding.EncodeToString(make([]byte, 256))
	if _, err := verifier.Verify(context.Background(), strings.Join(parts, ".")); !errors.Is(err, core.ErrUnauthorized) {
		t.Fatal("invalid signature accepted")
	}
}

func signTestToken(t *testing.T, key *rsa.PrivateKey, c claims, algorithm, kid string) string {
	t.Helper()
	header, _ := json.Marshal(map[string]string{"alg": algorithm, "kid": kid})
	payload, _ := json.Marshal(c)
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	hash := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	if err != nil {
		t.Fatal(err)
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature)
}
