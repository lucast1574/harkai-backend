package push_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/lucast1574/harkai-backend/internal/core"
)

func TestConcurrentClaimsAndExpiredLeaseAreFenced(t *testing.T) {
	s, _, now := fixture(t)
	i := enqueue(t, s, now)
	ctx := context.Background()
	var wg sync.WaitGroup
	results := make(chan error, 12)
	for range 12 {
		wg.Go(func() {
			_, err := s.ClaimPush(ctx, "first-lease", now)
			results <- err
		})
	}
	wg.Wait()
	close(results)
	winners := 0
	for err := range results {
		if err == nil {
			winners++
		} else if !errors.Is(err, core.ErrNotFound) {
			t.Fatal(err)
		}
	}
	if winners != 1 {
		t.Fatalf("concurrent claims: %d winners", winners)
	}
	if _, err := s.ClaimPush(ctx, "replacement-lease", now.Add(91*time.Second)); err != nil {
		t.Fatal("abandoned job was not reclaimable", err)
	}
	if err := s.AdvancePush(ctx, i.ID, "first-lease", "wrong", true, now); !errors.Is(err, core.ErrConflict) {
		t.Fatal("stale worker changed the replacement lease")
	}
}
