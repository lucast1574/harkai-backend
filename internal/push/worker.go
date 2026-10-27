package push

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/core"
	"log/slog"
	"time"
)

type Store interface {
	ClaimPush(context.Context, string, time.Time) (core.Incident, error)
	AdvancePush(context.Context, string, string, string, bool, time.Time) error
	PushDevices(context.Context, core.Incident, string, time.Time) ([]core.Device, error)
	LiveDevice(context.Context, core.Device, time.Time) (core.User, error)
	IncidentByID(context.Context, string) (core.Incident, error)
	DisableDevice(context.Context, core.Device) error
}
type Delivery interface {
	Send(context.Context, core.Incident, core.Device, time.Time) error
}
type Worker struct {
	Store  Store
	Sender Delivery
	Now    func() time.Time
	ID     func() string
}

func (w Worker) Run(ctx context.Context, logger *slog.Logger) error {
	timer := time.NewTicker(5 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			cycle, cancel := context.WithTimeout(ctx, 25*time.Second)
			err := w.Tick(cycle)
			cancel()
			if err != nil && !errors.Is(err, core.ErrNotFound) && !errors.Is(err, context.Canceled) {
				logger.Warn("notification delivery delayed")
			}
		}
	}
}
func (w Worker) Tick(ctx context.Context) error {
	lease := w.ID()
	i, err := w.Store.ClaimPush(ctx, lease, w.Now())
	if err != nil {
		return err
	}
	devices, err := w.Store.PushDevices(ctx, i, i.Push.Cursor, w.Now())
	if err != nil {
		return err
	}
	cursor := i.Push.Cursor
	for _, d := range devices {
		current, err := w.Store.IncidentByID(ctx, i.ID)
		if err != nil {
			return err
		}
		if !core.MapIncidentActive(current, w.Now()) {
			return w.Store.AdvancePush(ctx, i.ID, lease, cursor, true, w.Now())
		}
		user, err := w.Store.LiveDevice(ctx, d, w.Now())
		if err != nil && !errors.Is(err, core.ErrNotFound) {
			return err
		}
		if err == nil && core.PushEligible(current, d, user, w.Now()) {
			err = w.Sender.Send(ctx, current, d, w.Now())
			if errors.Is(err, ErrDeviceGone) {
				if err = w.Store.DisableDevice(ctx, d); err != nil {
					return err
				}
			} else if err != nil {
				delay := time.Duration(min(i.Push.Attempts, 10)) * time.Minute
				var retry RetryError
				if errors.As(err, &retry) {
					delay = max(delay, retry.Delay)
				}
				if updateErr := w.Store.AdvancePush(ctx, i.ID, lease, cursor, false, w.Now().Add(delay)); updateErr != nil {
					return updateErr
				}
				return err
			}
		}
		cursor = d.ID
		if err = w.Store.AdvancePush(ctx, i.ID, lease, cursor, false, w.Now().Add(90*time.Second)); err != nil {
			return err
		}
	}
	return w.Store.AdvancePush(ctx, i.ID, lease, cursor, len(devices) < 50, w.Now())
}
