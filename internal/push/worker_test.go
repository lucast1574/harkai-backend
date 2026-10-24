package push_test

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"github.com/lucast1574/harkai-backend/internal/push"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	"testing"
	"time"
)

type delivery struct {
	calls    int
	fail     bool
	verified bool
}

func (d *delivery) Send(_ context.Context, i core.Incident, _ core.Device, _ time.Time) error {
	d.calls++
	d.verified = i.Verified
	if d.fail {
		return push.RetryError{Delay: 2 * time.Minute}
	}
	return nil
}
func fixture(t *testing.T) (*mongostore.Store, core.Device, time.Time) {
	t.Helper()
	s := testutil.Store(t)
	ctx := context.Background()
	now := time.Now().UTC()
	u := core.User{ID: "recipient", Email: "recipient@example.test", Role: "user", Preferences: core.DefaultPreferences()}
	if _, err := s.CreateUser(ctx, u); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateSession(ctx, core.Session{ID: "session", UserID: u.ID, AccessHash: "access", RefreshHash: "refresh", AccessExpiresAt: now.Add(time.Hour), ExpiresAt: now.Add(24 * time.Hour)}); err != nil {
		t.Fatal(err)
	}
	d := core.Device{ID: "device", UserID: u.ID, SessionID: "session", Token: "test-token", Generation: "initial", Platform: "android", Location: core.Point{Type: "Point", Coordinates: []float64{-77.04, -12.04}}, UpdatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(24 * time.Hour)}
	if _, err := s.SaveDevice(ctx, d); err != nil {
		t.Fatal(err)
	}
	return s, d, now
}
func enqueue(t *testing.T, s *mongostore.Store, now time.Time) core.Incident {
	t.Helper()
	i, err := core.BuildIncident(core.IncidentInput{Type: "fire", Description: "A community report", Latitude: -12.04, Longitude: -77.04}, mongostore.NewID(), "author", now)
	if err != nil {
		t.Fatal(err)
	}
	i.Push = &core.PushJob{State: "pending", NextAt: now}
	if _, err = s.CreateIncident(context.Background(), i); err != nil {
		t.Fatal(err)
	}
	return i
}
func TestUnverifiedDeliveryIsDurableAndRetryable(t *testing.T) {
	s, _, now := fixture(t)
	ctx := context.Background()
	i := enqueue(t, s, now)
	sender := &delivery{fail: true}
	worker := push.Worker{Store: s, Sender: sender, Now: func() time.Time { return now }, ID: mongostore.NewID}
	if err := worker.Tick(ctx); err == nil {
		t.Fatal("missing delivery failure")
	}
	current, err := s.IncidentByID(ctx, i.ID)
	if err != nil || current.Push.State != "pending" || current.Push.Cursor != "" {
		t.Fatal("failure lost durable job")
	}
	now = now.Add(3 * time.Minute)
	sender.fail = false
	if err = worker.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	current, _ = s.IncidentByID(ctx, i.ID)
	if sender.calls != 2 || sender.verified || current.Push.State != "done" {
		t.Fatal("unverified report or retry missing")
	}
	if err = worker.Tick(ctx); !errors.Is(err, core.ErrNotFound) {
		t.Fatal("completed job replayed")
	}
}
func TestLogoutAndDeviceChangesSuppressDelivery(t *testing.T) {
	for _, mode := range []string{"logout", "revoked", "moved", "hidden"} {
		t.Run(mode, func(t *testing.T) {
			s, d, now := fixture(t)
			ctx := context.Background()
			i := enqueue(t, s, now)
			switch mode {
			case "logout":
				if err := s.DeleteSession(ctx, "access"); err != nil {
					t.Fatal(err)
				}
			case "revoked":
				if err := s.DeleteDevice(ctx, d.UserID, d.ID); err != nil {
					t.Fatal(err)
				}
			case "moved":
				d.Generation = "next"
				d.Location.Coordinates = []float64{-79, -8}
				d.UpdatedAt = now.Add(time.Second)
				if _, err := s.SaveDevice(ctx, d); err != nil {
					t.Fatal(err)
				}
			case "hidden":
				if _, err := s.SetIncidentStatus(ctx, i.ID, "author", "hidden", now); err != nil {
					t.Fatal(err)
				}
			}
			sender := &delivery{}
			worker := push.Worker{Store: s, Sender: sender, Now: func() time.Time { return now }, ID: mongostore.NewID}
			err := worker.Tick(ctx)
			if err != nil && !errors.Is(err, core.ErrNotFound) {
				t.Fatal(err)
			}
			if sender.calls != 0 {
				t.Fatal("delivery ignored latest preferences/session/status")
			}
		})
	}
}
func TestTokenCannotMoveToAnotherAccountAndRefreshKeepsBinding(t *testing.T) {
	s, d, now := fixture(t)
	ctx := context.Background()
	other := d
	other.UserID = "other"
	if _, err := s.SaveDevice(ctx, other); !errors.Is(err, core.ErrConflict) {
		t.Fatal("token reassigned across accounts")
	}
	next := core.Session{AccessHash: "next-access", RefreshHash: "next-refresh", AccessExpiresAt: now.Add(time.Hour)}
	if _, err := s.RotateSession(ctx, "refresh", next, now); err != nil {
		t.Fatal(err)
	}
	if _, err := s.LiveDevice(ctx, d, now); err != nil {
		t.Fatal("refresh detached notification device")
	}
	if err := s.DeleteSession(ctx, "next-access"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.LiveDevice(ctx, d, now); !errors.Is(err, core.ErrNotFound) {
		t.Fatal("logout did not revoke eligibility")
	}
}
