package httpapi_test

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	"testing"
	"time"
)

func TestAccountDeletionRequiresReauthenticationAndRevokesAccess(t *testing.T) {
	api := fixture(t)
	auth := registered(t, api, "delete-me@example.test")
	token := auth.AccessToken
	request(t, api, "GET", "/v1/me/account-deletion-options", token, nil, 200)
	request(t, api, "DELETE", "/v1/me", token, map[string]string{"confirmation": auth.User.Email, "password": "wrong"}, 403)
	request(t, api, "GET", "/v1/me", token, nil, 200)
	// Reuse the fixture's documented registration password.
	request(t, api, "DELETE", "/v1/me", token, map[string]string{"confirmation": auth.User.Email, "password": "safe-test-password-123"}, 202)
	request(t, api, "GET", "/v1/me", token, nil, 401)
}

func TestAccountDeletionPurgesStoredPhotosAndKeepsOtherAccounts(t *testing.T) {
	api := fixture(t)
	auth := registered(t, api, "purge@example.test")
	other := registered(t, api, "keep@example.test")
	photos := testutil.Photos(t)
	ctx := context.Background()
	store := api.Service.AccountDeletion
	mediaID := "0123456789abcdef01234567"
	if err := photos.Put(ctx, mediaID, []byte("test photo bytes")); err != nil {
		t.Fatal(err)
	}
	if _, err := api.Service.Media.CreateMedia(ctx, core.Media{ID: mediaID, UserID: auth.User.ID, Approved: true}); err != nil {
		t.Fatal(err)
	}
	// Capture the injected clock so the cleanup grace period is deterministic.
	now := api.Service.Now()
	api.Service.Now = func() time.Time { return now }
	if err := api.Service.DeleteMyAccount(ctx, auth.User, auth.User.Email, "safe-test-password-123", ""); err == nil {
		t.Fatal("the public response must never expose a password hash")
	}
	request(t, api, "DELETE", "/v1/me", auth.AccessToken, map[string]string{"confirmation": auth.User.Email, "password": "safe-test-password-123"}, 202)
	if _, err := store.ClaimAccountDeletion(ctx, now); !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("cleanup must allow in-flight requests to finish: %v", err)
	}
	now = now.Add(61 * time.Second)
	if err := api.Service.ProcessAccountDeletion(ctx, photos); err != nil {
		t.Fatal(err)
	}
	request(t, api, "GET", "/v1/me", auth.AccessToken, nil, 401)
	request(t, api, "GET", "/v1/me", other.AccessToken, nil, 200)
	if _, err := photos.Get(ctx, mediaID); err == nil {
		t.Fatal("photo object survived deletion")
	}
	if _, err := api.Service.Media.MediaByID(ctx, mediaID); !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("photo metadata survived: %v", err)
	}
	registered(t, api, "purge@example.test")
}
