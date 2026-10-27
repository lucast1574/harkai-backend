package application

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"time"
)

func (s Service) ConfirmReport(ctx context.Context, user core.User, id string) (core.Incident, error) {
	i, err := s.Incidents.IncidentByID(ctx, id)
	if err != nil {
		return core.Incident{}, err
	}
	now := s.Now()
	if err = core.ValidateConfirmation(i, user.ID, now); err != nil {
		return core.Incident{}, err
	}
	if err = s.Limits.Allow(ctx, "confirmation:"+user.ID, 30, time.Hour, now); err != nil {
		return core.Incident{}, err
	}
	i, err = s.Incidents.ConfirmIncident(ctx, id, user.ID, now)
	i.ViewerHasConfirmed = err == nil
	return core.PublicIncident(i), err
}
