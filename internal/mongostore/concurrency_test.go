package mongostore_test

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimitsAreAtomicAcrossRequests(t *testing.T) {
	store := testutil.Store(t)
	now := time.Now().UTC()
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for range 30 {
		wg.Go(func() {
			err := store.Allow(context.Background(), "same-account", 6, time.Minute, now)
			if err == nil {
				allowed.Add(1)
			} else if !errors.Is(err, core.ErrRateLimited) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
	wg.Wait()
	if allowed.Load() != 6 {
		t.Fatalf("allowed %d requests instead of 6", allowed.Load())
	}
}

func TestCooldownHasOneWinner(t *testing.T) {
	store := testutil.Store(t)
	now := time.Now().UTC()
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			err := store.ClaimCooldown(context.Background(), "user", "fire", now)
			if err == nil {
				allowed.Add(1)
			} else if !errors.Is(err, core.ErrRateLimited) {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
	wg.Wait()
	if allowed.Load() != 1 {
		t.Fatalf("%d winners", allowed.Load())
	}
	if err := store.ClaimCooldown(context.Background(), "user", "fire", now.Add(time.Hour)); err != nil {
		t.Fatal("cooldown boundary should allow report", err)
	}
}

func TestRefreshRotationHasOneWinner(t *testing.T) {
	store := testutil.Store(t)
	now := time.Now().UTC()
	initial := core.Session{ID: "session", UserID: "user", AccessHash: "access", RefreshHash: "refresh", AccessExpiresAt: now.Add(time.Minute), ExpiresAt: now.Add(time.Hour)}
	if err := store.CreateSession(context.Background(), initial); err != nil {
		t.Fatal(err)
	}
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			previous, err := store.RotateSession(context.Background(), "refresh", core.Session{AccessHash: "next-access", RefreshHash: "next-refresh", AccessExpiresAt: now.Add(time.Minute)}, now)
			if err == nil {
				allowed.Add(1)
				if previous.UserID != "user" {
					t.Error("rotation lost owner")
				}
			} else if !errors.Is(err, core.ErrUnauthorized) {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if allowed.Load() != 1 {
		t.Fatalf("%d refresh winners", allowed.Load())
	}
	if _, err := store.SessionByAccess(context.Background(), "access", now); !errors.Is(err, core.ErrUnauthorized) {
		t.Fatal("old access token survived rotation")
	}
}
