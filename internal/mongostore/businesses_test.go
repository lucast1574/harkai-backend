package mongostore_test

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/application"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"github.com/lucast1574/harkai-backend/internal/security"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	"testing"
	"time"
)

func TestBusinessBillingVisibilityAndProximityDelivery(t *testing.T) {
	store := testutil.Store(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	p, err := core.BuildBusiness(core.BusinessInput{Name: "Negocio", Kind: "restaurant", Description: "Almuerzo casero", Latitude: -12, Longitude: -77, RadiusMeters: 250}, mongostore.NewID(), "owner", now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.CreateBusiness(ctx, p); err != nil {
		t.Fatal(err)
	}
	q := core.IncidentQuery{Latitude: -12, Longitude: -77, RadiusMeters: 1000, Limit: 50}
	page, err := store.ListBusinesses(ctx, q, "", false, now)
	if err != nil || len(page.Items) != 0 {
		t.Fatal("draft public", err)
	}
	if _, err = store.ModerateBusiness(ctx, p.ID, "approved", now); err != nil {
		t.Fatal(err)
	}
	e := core.BusinessEntitlement{PointID: p.ID, OwnerID: "owner", EventID: "event1", Mode: "one_time", PaymentID: "payment1", PaidAt: &now, Status: "active", EventAt: now, PaidUntil: now.Add(24 * time.Hour)}
	if err = store.ApplyBusinessEntitlement(ctx, e, now); err != nil {
		t.Fatal(err)
	}
	if err = store.ApplyBusinessEntitlement(ctx, e, now); err != nil {
		t.Fatal("duplicate not idempotent", err)
	}
	e.EventID = "old"
	e.EventAt = now.Add(-time.Second)
	e.Status = "canceled"
	if err = store.ApplyBusinessEntitlement(ctx, e, now); !errors.Is(err, core.ErrConflict) {
		t.Fatal("old event changed entitlement", err)
	}
	e.EventID = "foreign"
	e.EventAt = now.Add(time.Second)
	e.OwnerID = "attacker"
	if err = store.ApplyBusinessEntitlement(ctx, e, now); !errors.Is(err, core.ErrConflict) {
		t.Fatal("foreign owner changed billing", err)
	}
	page, err = store.ListBusinesses(ctx, q, "", false, now)
	if err != nil || len(page.Items) != 1 {
		t.Fatal("paid point absent", err)
	}
	s := application.Service{Businesses: store, Now: func() time.Time { return now }, HashToken: security.HashToken}
	u := core.User{ID: "visitor", Preferences: core.DefaultPreferences()}
	d := core.Device{ID: "device1", UserID: u.ID, SessionID: "session1"}
	notices, err := s.VisitBusinesses(ctx, u, d, -12, -77, 10, now)
	if err != nil || len(notices) != 1 {
		t.Fatal("app-managed business notification missing", err)
	}
	u.Preferences.BusinessNotifications = true
	notices, err = s.VisitBusinesses(ctx, u, d, -12, -77, 101, now)
	if err != nil || len(notices) != 0 {
		t.Fatal("inaccurate GPS delivered", err)
	}
	notices, err = s.VisitBusinesses(ctx, u, d, -12, -77, 10, now)
	if err != nil || len(notices) != 1 {
		t.Fatal("eligible notice missing", err)
	}
	id := notices[0].ID
	notices, err = s.VisitBusinesses(ctx, u, d, -12, -77, 10, now)
	if err != nil || len(notices) != 1 || notices[0].ID != id {
		t.Fatal("lost response not recoverable", err)
	}
	if err = store.AckProximityNotice(ctx, u.ID, "wrong-session", id, 15); err != nil {
		t.Fatal(err)
	}
	notices, _ = store.BusinessNotices(ctx, u.ID, d.SessionID, d.ID, now)
	if len(notices) != 1 {
		t.Fatal("foreign ack removed notice")
	}
	if err = store.AckProximityNotice(ctx, u.ID, d.SessionID, id, 15); err != nil {
		t.Fatal(err)
	}
	notices, err = s.VisitBusinesses(ctx, u, d, -12, -77, 10, now)
	if err != nil || len(notices) != 0 {
		t.Fatal("cooldown failed", err)
	}
	now = now.Add(25 * time.Hour)
	page, err = store.ListBusinesses(ctx, q, "", false, now)
	if err != nil || len(page.Items) != 1 {
		t.Fatal("one-time point unexpectedly expired", err)
	}
	if err = s.BusinessBillingEvent(ctx, []byte(`{}`), "forged"); !errors.Is(err, core.ErrUnavailable) {
		t.Fatal("unconfigured billing accepted", err)
	}
}
func TestPermanentPublicationsAndLegacyMapExpiry(t *testing.T) {
	store := testutil.Store(t)
	ctx := context.Background()
	now := time.Now().UTC()
	for _, kind := range []string{"pet", "place", "fire"} {
		i, err := core.BuildIncident(core.IncidentInput{Type: kind, Description: "Un reporte antiguo", Latitude: -12, Longitude: -77}, mongostore.NewID(), "owner", now.Add(-4*time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		i.ExpiresAt = nil
		if _, err = store.CreateIncident(ctx, i); err != nil {
			t.Fatal(err)
		}
	}
	q := core.IncidentQuery{Latitude: -12, Longitude: -77, RadiusMeters: 1000, Limit: 50, PublicOnly: true}
	page, err := store.ListIncidents(ctx, q, now)
	if err != nil || len(page.Items) != 0 {
		t.Fatal("legacy markers survived three hours", err)
	}
	archive, err := store.ListPublications(ctx, "", 2)
	if err != nil || len(archive.Items) != 2 || archive.NextCursor == "" {
		t.Fatal("archive pagination", err)
	}
	next, err := store.ListPublications(ctx, archive.NextCursor, 2)
	if err != nil || len(next.Items) != 1 {
		t.Fatal("archive lost content", err)
	}
	if _, err = store.SetIncidentStatus(ctx, next.Items[0].ID, "owner", "hidden", now); err != nil {
		t.Fatal(err)
	}
	all, err := store.ListPublications(ctx, "", 100)
	if err != nil || len(all.Items) != 2 {
		t.Fatal("hidden publication public", err)
	}
}
