package mongostore_test

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/application"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/security"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	"strings"
	"testing"
	"time"
)

func TestInterruptedRotationRecoversAcrossServiceRestart(t *testing.T) {
	store := testutil.Store(t)
	now := time.Now().UTC()
	ctx := context.Background()
	_, err := store.CreateUser(ctx, core.User{ID: "owner", Name: "Test", Email: "session-recovery@example.test", Role: "user", Preferences: core.DefaultPreferences()})
	if err != nil {
		t.Fatal(err)
	}
	old := strings.Repeat("a", 43)
	nonce := strings.Repeat("b", 43)
	if err = store.CreateSession(ctx, core.Session{ID: "session", UserID: "owner", AccessHash: "old-access", RefreshHash: security.HashToken(old), AccessExpiresAt: now, ExpiresAt: now.Add(24 * time.Hour)}); err != nil {
		t.Fatal(err)
	}
	service := func() application.Service {
		return application.Service{Sessions: store, Users: store, Now: func() time.Time { return now }, HashToken: security.HashToken, DeriveSessionTokens: security.DeriveSessionTokens}
	}
	first, err := service().RefreshWithRotation(ctx, old, nonce)
	if err != nil {
		t.Fatal(err)
	}
	// The client lost the first response and returns hours later after restart.
	now = now.Add(2 * time.Hour)
	recovered, err := service().RefreshWithRotation(ctx, old, nonce)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.AccessToken != first.AccessToken || recovered.RefreshToken != first.RefreshToken || recovered.User.ID != "owner" {
		t.Fatal("recovery changed identity or tokens")
	}
	if _, err = store.SessionByAccess(ctx, security.HashToken(recovered.AccessToken), now); err != nil {
		t.Fatal("recovered access was expired", err)
	}
	if _, err = service().RefreshWithRotation(ctx, old, strings.Repeat("c", 43)); !errors.Is(err, core.ErrUnauthorized) {
		t.Fatal("wrong rotation ID was accepted", err)
	}
	if _, err = service().RefreshWithRotation(ctx, recovered.RefreshToken, strings.Repeat("d", 43)); err != nil {
		t.Fatal(err)
	}
	if _, err = service().RefreshWithRotation(ctx, old, nonce); !errors.Is(err, core.ErrUnauthorized) {
		t.Fatal("superseded recovery survived next rotation", err)
	}
	now = now.Add(48 * time.Hour)
	if _, err = service().RefreshWithRotation(ctx, recovered.RefreshToken, strings.Repeat("d", 43)); !errors.Is(err, core.ErrUnauthorized) {
		t.Fatal("expired session recovered", err)
	}
}
