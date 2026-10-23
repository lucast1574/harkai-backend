package application

import (
	"context"
	"time"

	"github.com/lucast1574/harkai-backend/internal/core"
)

// Moving a device preserves its token and generation, including in-flight jobs.
func (s Service) MoveDevice(ctx context.Context, u core.User, access, id string, lat, lng *float64) (core.Device, error) {
	if !s.PushEnabled || s.Devices == nil {
		return core.Device{}, core.ErrUnavailable
	}
	if len(id) != 64 || lat == nil || lng == nil {
		return core.Device{}, core.ErrInvalid
	}
	if err := core.ValidateCoordinates(*lat, *lng); err != nil {
		return core.Device{}, err
	}
	now := s.Now()
	session, err := s.Sessions.SessionByAccess(ctx, s.HashToken(access), now)
	if err != nil {
		return core.Device{}, err
	}
	if session.UserID != u.ID {
		return core.Device{}, core.ErrForbidden
	}
	if err = s.Limits.Allow(ctx, "device-location:"+u.ID, 360, time.Hour, now); err != nil {
		return core.Device{}, err
	}
	point := core.Point{Type: "Point", Coordinates: []float64{*lng, *lat}}
	return s.Devices.MoveDevice(ctx, u.ID, session.ID, id, point, now, session.ExpiresAt)
}
