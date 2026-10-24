package mongostore_test

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	"testing"
	"time"
)

func TestProximityProgressOnlyInwardAndAcknowledgementFenced(t *testing.T) {
	s := testutil.Store(t)
	ctx := context.Background()
	now := time.Now().UTC()
	i, err := core.BuildIncident(core.IncidentInput{Type: "fire", Description: "Fuego reportado", Latitude: -12, Longitude: -77}, mongostore.NewID(), "author", now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateIncident(ctx, i); err != nil {
		t.Fatal(err)
	}
	n := core.BusinessNotice{ID: "stable-notice", UserID: "visitor", SessionID: "session", DeviceID: "device", PointID: i.ID, Kind: "incident", Stage: 500, Pending: true, CreatedAt: now, NextEligibleAt: now.Add(24 * time.Hour)}
	if err = s.QueueBusinessNotice(ctx, n); err != nil {
		t.Fatal(err)
	}
	n.Stage = 250
	if err = s.QueueBusinessNotice(ctx, n); err != nil {
		t.Fatal(err)
	}
	if err = s.AckProximityNotice(ctx, n.UserID, n.SessionID, n.ID, 500); err != nil {
		t.Fatal(err)
	}
	items, err := s.BusinessNotices(ctx, n.UserID, n.SessionID, n.DeviceID, now)
	if err != nil || len(items) != 1 || items[0].Stage != 250 {
		t.Fatal("stale ack hid a closer warning", items, err)
	}
	if err = s.AckProximityNotice(ctx, n.UserID, n.SessionID, n.ID, 250); err != nil {
		t.Fatal(err)
	}
	n.Stage = 500
	if err = s.QueueBusinessNotice(ctx, n); err != nil {
		t.Fatal(err)
	}
	items, err = s.BusinessNotices(ctx, n.UserID, n.SessionID, n.DeviceID, now)
	if err != nil || len(items) != 0 {
		t.Fatal("walking out repeated a warning", err)
	}
	n.Stage = 50
	if err = s.QueueBusinessNotice(ctx, n); err != nil {
		t.Fatal(err)
	}
	n.Stage = 15
	if err = s.QueueBusinessNotice(ctx, n); err != nil {
		t.Fatal(err)
	}
	items, err = s.BusinessNotices(ctx, n.UserID, n.SessionID, n.DeviceID, now.Add(4*time.Hour))
	if err != nil || len(items) != 0 {
		t.Fatal("expired incident delivered", err)
	}
}
