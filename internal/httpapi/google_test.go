package httpapi_test

import (
	"context"
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/core"
	"testing"
)

func TestGoogleIdentityDoesNotTakeOverPasswordAccount(t *testing.T) {
	api := fixture(t)
	existing := registered(t, api, "owned@example.com")
	api.Service.VerifyGoogle = func(_ context.Context, token string) (core.GoogleIdentity, error) {
		if token == "invalid" {
			return core.GoogleIdentity{}, core.ErrUnauthorized
		}
		email := "google@example.com"
		if token == "collision" {
			email = "owned@example.com"
		}
		return core.GoogleIdentity{Subject: token, Email: email, Name: "Google User"}, nil
	}
	request(t, api, "POST", "/v1/auth/google", "", map[string]string{"id_token": "invalid"}, 401)
	request(t, api, "POST", "/v1/auth/google", "", map[string]string{"id_token": "collision"}, 409)
	request(t, api, "GET", "/v1/me", existing.AccessToken, nil, 200)
	first := request(t, api, "POST", "/v1/auth/google", "", map[string]string{"id_token": "subject-one"}, 200)
	var a, b core.AuthResult
	if err := json.Unmarshal(first.Body.Bytes(), &a); err != nil {
		t.Fatal(err)
	}
	second := request(t, api, "POST", "/v1/auth/google", "", map[string]string{"id_token": "subject-one"}, 200)
	if err := json.Unmarshal(second.Body.Bytes(), &b); err != nil {
		t.Fatal(err)
	}
	if a.User.ID != b.User.ID || a.User.Role != "user" || a.User.GoogleSubject != "" || a.User.PasswordHash != "" {
		t.Fatal("Google identity did not map to a safe backend account")
	}
	request(t, api, "GET", "/v1/me", b.AccessToken, nil, 200)
	request(t, api, "POST", "/v1/auth/google", "", map[string]string{"id_token": "subject-one", "role": "admin"}, 400)
}
