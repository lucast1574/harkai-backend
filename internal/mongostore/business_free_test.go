package mongostore_test

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/application"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"github.com/lucast1574/harkai-backend/internal/security"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	"testing"
	"time"
)

func TestFreeBusinessVisibleAndNotifiedOnlyAfterReview(t *testing.T) {
	store := testutil.Store(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	p, err := core.BuildBusiness(core.BusinessInput{Name: "Café vecinal", Kind: "restaurant", Description: "Café y pan artesanal", Latitude: -12, Longitude: -77}, mongostore.NewID(), "owner", now)
	if err != nil {
		t.Fatal(err)
	}
	p.PaymentMode, p.BillingStatus = "free", "active"
	if _, err = store.CreateBusiness(ctx, p); err != nil {
		t.Fatal(err)
	}
	s := application.Service{Businesses: store, Now: func() time.Time { return now }, HashToken: security.HashToken}
	u := core.User{ID: "visitor", Preferences: core.DefaultPreferences()}
	d := core.Device{ID: "device", UserID: u.ID, SessionID: "session"}
	notices, err := s.VisitBusinesses(ctx, u, d, -12, -77, 10, now)
	if err != nil || len(notices) != 0 {
		t.Fatal("unreviewed business notified", err)
	}
	if _, err = store.ModerateBusiness(ctx, p.ID, "approved", now); err != nil {
		t.Fatal(err)
	}
	notices, err = s.VisitBusinesses(ctx, u, d, -12, -77, 10, now)
	if err != nil || len(notices) != 1 {
		t.Fatal("free approved business not delivered", err)
	}
	if _, err = store.SetBusinessEnabled(ctx, p.ID, p.OwnerID, false, now); err != nil {
		t.Fatal(err)
	}
	notices, err = s.VisitBusinesses(ctx, u, d, -12, -77, 10, now)
	if err != nil || len(notices) != 0 {
		t.Fatal("paused business delivered", err)
	}
}
