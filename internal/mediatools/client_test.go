package mediatools

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestModerationFailsClosed(t *testing.T) {
	cases := []struct {
		name, body string
		status     int
		want       error
	}{
		{"approved", `{"approved":true,"policy":"explicit-v1","reason":"approved"}`, 200, nil},
		{"rejected", `{"approved":false,"policy":"explicit-v1","reason":"explicit_content"}`, 200, core.ErrImageRejected},
		{"unknown policy", `{"approved":true,"policy":"other","reason":"approved"}`, 200, core.ErrUnavailable},
		{"missing reason", `{"approved":true,"policy":"explicit-v1"}`, 200, core.ErrUnavailable},
		{"bad JSON", `broken`, 200, core.ErrUnavailable},
		{"busy", `{}`, 503, core.ErrUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer test-worker-key-32-characters-long" {
					t.Error("missing authorization")
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			client, err := New(server.URL, "test-worker-key-32-characters-long")
			if err != nil {
				t.Fatal(err)
			}
			policy, err := client.Moderate(context.Background(), []byte("safe test payload"))
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v got %v", tc.want, err)
			}
			if err == nil && policy != "explicit-v1" {
				t.Fatal("bad policy")
			}
		})
	}
}

func TestWorkerRejectsRedirectAndBadConfig(t *testing.T) {
	for _, address := range []string{"file:///tmp/worker", "http://user:pass@worker", "http://worker?key=secret"} {
		if _, err := New(address, "test-worker-key-32-characters-long"); err == nil {
			t.Fatal("unsafe worker config")
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "https://example.invalid", 302) }))
	defer server.Close()
	client, _ := New(server.URL, "test-worker-key-32-characters-long")
	if _, err := client.Moderate(context.Background(), []byte("test")); !errors.Is(err, core.ErrUnavailable) {
		t.Fatal("redirect must deny review")
	}
}
