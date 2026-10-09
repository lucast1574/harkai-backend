package application

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"time"
)

type Devices interface {
	SaveDevice(context.Context, core.Device) (core.Device, error)
	DeleteDevice(context.Context, string, string) error
}

func (s Service) RegisterDevice(ctx context.Context, u core.User, access string, input core.DeviceInput) (core.Device, error) {
	if !s.PushEnabled || s.Devices == nil {
		return core.Device{}, core.ErrUnavailable
	}
	input, err := core.ValidateDevice(input)
	if err != nil {
		return core.Device{}, err
	}
	session, err := s.Sessions.SessionByAccess(ctx, s.HashToken(access), s.Now())
	if err != nil {
		return core.Device{}, err
	}
	if session.UserID != u.ID {
		return core.Device{}, core.ErrForbidden
	}
	if err = s.Limits.Allow(ctx, "device:"+u.ID, 30, time.Hour, s.Now()); err != nil {
		return core.Device{}, err
	}
	d := core.Device{ID: s.HashToken(input.Token), Token: input.Token, UserID: u.ID, SessionID: session.ID, Generation: s.ID(), Platform: input.Platform, Location: core.Point{Type: "Point", Coordinates: []float64{input.Longitude, input.Latitude}}, UpdatedAt: s.Now(), ExpiresAt: session.ExpiresAt}
	return s.Devices.SaveDevice(ctx, d)
}
func (s Service) RemoveDevice(ctx context.Context, u core.User, id string) error {
	if s.Devices == nil {
		return core.ErrUnavailable
	}
	if len(id) != 64 {
		return core.ErrInvalid
	}
	return s.Devices.DeleteDevice(ctx, u.ID, id)
}
