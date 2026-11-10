package mongostore_test

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	"sync"
	"testing"
	"time"
)

func TestConcurrentCheckoutsShareOneAttemptAcrossReplicas(t *testing.T) {
	s := testutil.Store(t)
	ctx := context.Background()
	point := mongostore.NewID()
	var group sync.WaitGroup
	keys := make(chan string, 20)
	for i := 0; i < 20; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			r, err := s.BusinessCheckout(ctx, core.BusinessCheckout{PointID: point, OwnerID: "owner", Key: mongostore.NewID(), CreatedAt: time.Now()})
			if err != nil {
				t.Error(err)
				return
			}
			keys <- r.Key
		}()
	}
	group.Wait()
	close(keys)
	first := ""
	for key := range keys {
		if first == "" {
			first = key
		}
		if key != first {
			t.Fatal("concurrent callers received different payment attempts")
		}
	}
	r, err := s.BusinessCheckoutByID(ctx, point)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.SaveBusinessCheckout(ctx, r, "cs_harkai"); err != nil {
		t.Fatal(err)
	}
	if err = s.RotateBusinessCheckout(ctx, r, "next"); err != nil {
		t.Fatal(err)
	}
	if err = s.SaveBusinessCheckout(ctx, r, "stale_session"); err == nil {
		t.Fatal("stale caller overwrote new attempt")
	}
}
