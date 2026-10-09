package mongostore_test

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	"testing"
	"time"
)

func TestExpiryPaginationAndVisibility(t *testing.T) {
	store := testutil.Store(t)
	ctx := context.Background()
	now := time.Now().UTC()
	for n := range 5 {
		i, err := core.BuildIncident(core.IncidentInput{Type: "fire", Description: "Visible incident", Latitude: -8.1, Longitude: -79}, mongostore.NewID(), "user", now)
		if err != nil {
			t.Fatal(err)
		}
		if n == 0 {
			expired := now.Add(-time.Minute)
			i.ExpiresAt = &expired
		}
		if n == 1 {
			i.Status = "hidden"
		}
		if _, err = store.CreateIncident(ctx, i); err != nil {
			t.Fatal(err)
		}
	}
	q := core.IncidentQuery{Latitude: -8.1, Longitude: -79, RadiusMeters: 500, Limit: 2}
	first, err := store.ListIncidents(ctx, q, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Items) != 2 || first.NextCursor == "" {
		t.Fatalf("bad first page: %+v", first)
	}
	q.Cursor = first.NextCursor
	second, err := store.ListIncidents(ctx, q, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Items) != 1 || second.NextCursor != "" {
		t.Fatalf("bad next page: %+v", second)
	}
	for _, i := range second.Items {
		if i.ID == first.Items[0].ID || i.ID == first.Items[1].ID {
			t.Fatal("duplicate across pages")
		}
	}
}
