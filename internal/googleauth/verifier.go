package googleauth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"github.com/lucast1574/harkai-backend/internal/core"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Verifier struct {
	project, certURL string
	client           *http.Client
	mu               sync.Mutex
	certificates     map[string]*rsa.PublicKey
	expires          time.Time
}
type claims struct {
	Subject       string `json:"sub"`
	Audience      string `json:"aud"`
	Issuer        string `json:"iss"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	EmailVerified bool   `json:"email_verified"`
	Expires       int64  `json:"exp"`
	Issued        int64  `json:"iat"`
	AuthTime      int64  `json:"auth_time"`
	Firebase      struct {
		Provider string `json:"sign_in_provider"`
	} `json:"firebase"`
}

func New(project string) *Verifier {
	return &Verifier{project: project, certURL: "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com", client: &http.Client{Timeout: 5 * time.Second, CheckRedirect: denyRedirect}}
}
func denyRedirect(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }

func (v *Verifier) Verify(ctx context.Context, token string) (core.GoogleIdentity, error) {
	if len(token) > 16384 {
		return core.GoogleIdentity{}, core.ErrUnauthorized
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return core.GoogleIdentity{}, core.ErrUnauthorized
	}
	var header struct {
		Algorithm string `json:"alg"`
		KeyID     string `json:"kid"`
	}
	if decodePart(parts[0], &header) != nil || header.Algorithm != "RS256" || header.KeyID == "" {
		return core.GoogleIdentity{}, core.ErrUnauthorized
	}
	var c claims
	if decodePart(parts[1], &c) != nil {
		return core.GoogleIdentity{}, core.ErrUnauthorized
	}
	now := time.Now().Unix()
	if c.Audience != v.project || c.Issuer != "https://securetoken.google.com/"+v.project || c.Subject == "" || len(c.Subject) > 128 || c.Expires <= now || c.Issued > now || c.AuthTime > now || c.AuthTime <= 0 || c.Issued <= 0 || c.Firebase.Provider != "google.com" || !c.EmailVerified {
		return core.GoogleIdentity{}, core.ErrUnauthorized
	}
	key, err := v.key(ctx, header.KeyID)
	if err != nil {
		return core.GoogleIdentity{}, err
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return core.GoogleIdentity{}, core.ErrUnauthorized
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature) != nil {
		return core.GoogleIdentity{}, core.ErrUnauthorized
	}
	email, err := core.NormalizeEmail(c.Email)
	if err != nil {
		return core.GoogleIdentity{}, core.ErrUnauthorized
	}
	return core.GoogleIdentity{Subject: c.Subject, Email: email, Name: c.Name}, nil
}
func decodePart(value string, dst any) error {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dst)
}

func (v *Verifier) key(ctx context.Context, id string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if time.Now().Before(v.expires) {
		if key, ok := v.certificates[id]; ok {
			return key, nil
		}
		return nil, core.ErrUnauthorized
	}
	req, err := http.NewRequestWithContext(ctx, "GET", v.certURL, nil)
	if err != nil {
		return nil, core.ErrUnavailable
	}
	response, err := v.client.Do(req)
	if err != nil {
		return nil, core.ErrUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, core.ErrUnavailable
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, 64*1024+1))
	if err != nil || len(raw) > 64*1024 {
		return nil, core.ErrUnavailable
	}
	var encoded map[string]string
	if json.Unmarshal(raw, &encoded) != nil {
		return nil, core.ErrUnavailable
	}
	keys := map[string]*rsa.PublicKey{}
	for kid, text := range encoded {
		block, _ := pem.Decode([]byte(text))
		if block == nil {
			return nil, core.ErrUnavailable
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil || time.Now().After(cert.NotAfter) {
			return nil, core.ErrUnavailable
		}
		key, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, core.ErrUnavailable
		}
		keys[kid] = key
	}
	v.certificates = keys
	v.expires = time.Now().Add(5 * time.Minute)
	key, ok := keys[id]
	if !ok {
		return nil, core.ErrUnauthorized
	}
	return key, nil
}
