package mongostore_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCommunityConfirmationHasOneWinner(t *testing.T) {
	store := testutil.Store(t)
	now := time.Now().UTC()
	expires := now.Add(time.Hour)
	i := core.Incident{ID: "report", UserID: "author", Status: "active", ExpiresAt: &expires, Location: core.Point{Type: "Point", Coordinates: []float64{-77, -12}}}
	if _, err := store.CreateIncident(context.Background(), i); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConfirmIncident(context.Background(), i.ID, i.UserID, now); !errors.Is(err, core.ErrConflict) {
		t.Fatal("author confirmation accepted", err)
	}
	var wins atomic.Int32
	var wg sync.WaitGroup
	for n := range 20 {
		wg.Go(func() {
			confirmed, err := store.ConfirmIncident(context.Background(), i.ID, fmt.Sprint("user-", n), now)
			if err == nil {
				wins.Add(1)
				if !confirmed.Verified || confirmed.Verification != "community" || confirmed.VerifiedAt == nil || confirmed.VerifiedBy == "" {
					t.Error("confirmation metadata missing")
				}
			} else if !errors.Is(err, core.ErrConflict) {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatalf("%d winners, want one", wins.Load())
	}
}
