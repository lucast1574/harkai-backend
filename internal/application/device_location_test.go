package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lucast1574/harkai-backend/internal/application"
	"github.com/lucast1574/harkai-backend/internal/core"
)

type movingDevices struct {
	application.Devices
	calls             int
	user, session, id string
	point             core.Point
}

func (d *movingDevices) MoveDevice(_ context.Context, user, session, id string, point core.Point, _, _ time.Time) (core.Device, error) {
	d.calls++
	d.user, d.session, d.id, d.point = user, session, id, point
	return core.Device{ID: id, Location: point}, nil
}

type deviceSessions struct {
	application.Sessions
	owner string
}

func (s deviceSessions) SessionByAccess(context.Context, string, time.Time) (core.Session, error) {
	return core.Session{ID: "session", UserID: s.owner, ExpiresAt: time.Now().Add(time.Hour)}, nil
}

type locationLimits struct {
	limit  int
	denied bool
}

func (l *locationLimits) Allow(_ context.Context, key string, count int, window time.Duration, _ time.Time) error {
	if key != "device-location:user" || window != time.Hour {
		return core.ErrInvalid
	}
	l.limit = count
	if l.denied {
		return core.ErrForbidden
	}
	return nil
}

func TestMovingDeviceUsesAuthenticatedSessionAndCoordinates(t *testing.T) {
	devices, limits := &movingDevices{}, &locationLimits{}
	s := application.Service{Devices: devices, PushEnabled: true, Sessions: deviceSessions{owner: "user"}, Limits: limits, Now: time.Now, HashToken: func(s string) string { return s }}
	lat, lng := -12.04, -77.04
	id := strings.Repeat("a", 64)
	_, err := s.MoveDevice(context.Background(), core.User{ID: "user"}, "access", id, &lat, &lng)
	if err != nil || devices.calls != 1 || devices.user != "user" || devices.session != "session" || devices.id != id || devices.point.Coordinates[0] != lng || limits.limit != 360 {
		t.Fatalf("incorrect authenticated location update: %v", err)
	}
	for _, test := range []struct {
		name     string
		lat, lng *float64
		owner    string
		denied   bool
	}{
		{"missing latitude", nil, &lng, "user", false},
		{"foreign session", &lat, &lng, "other", false},
		{"rate limited", &lat, &lng, "user", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			s.Sessions = deviceSessions{owner: test.owner}
			limits.denied = test.denied
			_, err := s.MoveDevice(context.Background(), core.User{ID: "user"}, "access", id, test.lat, test.lng)
			if err == nil || devices.calls != 1 {
				t.Fatal("invalid update reached persistence")
			}
		})
	}
	lat = 91
	_, err = s.MoveDevice(context.Background(), core.User{ID: "user"}, "access", id, &lat, &lng)
	if !errors.Is(err, core.ErrInvalid) || devices.calls != 1 {
		t.Fatal("invalid coordinates accepted")
	}
}
