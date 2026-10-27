package application

import (
	"context"
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
	"slices"
	"time"
)

func (s Service) Report(ctx context.Context, user core.User, input core.IncidentInput) (core.Incident, error) {
	if input.Category != "" || input.Type == "crime" {
		category := input.Category
		if category == "" {
			category = "crime"
		}
		kind, err := core.ResolveReportType(input.Description, category)
		if input.ManualType {
			kind, err = core.ManualReportType(input.Type, category)
		}
		if err != nil {
			return core.Incident{}, err
		}
		input.Type = kind
	}
	now := s.Now()
	incident, err := core.BuildIncident(input, s.ID(), user.ID, now)
	if err != nil {
		return core.Incident{}, err
	}
	if err = s.ValidateMedia(ctx, user.ID, input.MediaID); err != nil {
		return core.Incident{}, err
	}
	if err = s.Limits.Allow(ctx, "report:"+user.ID, 20, time.Hour, now); err != nil {
		return core.Incident{}, err
	}
	if err = s.Incidents.ClaimCooldown(ctx, user.ID, input.Type, now); err != nil {
		return core.Incident{}, err
	}
	if s.DistrictAt != nil {
		if district, e := s.DistrictAt(incident.Latitude, incident.Longitude); e == nil {
			incident.District, incident.City, incident.Country = district.Name, district.Province, "Perú"
		}
	}
	if s.PushEnabled {
		incident.Push = &core.PushJob{State: "pending", NextAt: now}
	}
	incident, err = s.Incidents.CreateIncident(ctx, incident)
	incident.ViewerIsAuthor = true
	return core.PublicIncident(incident), err
}

func (s Service) FindIncidents(ctx context.Context, q core.IncidentQuery) (core.IncidentPage, error) {
	q, err := core.ValidateQuery(q)
	if err != nil {
		return core.IncidentPage{}, err
	}
	page, err := s.Incidents.ListIncidents(ctx, q, s.Now())
	for n := range page.Items {
		page.Items[n] = core.PublicIncident(page.Items[n])
	}
	return page, err
}

func (s Service) Incident(ctx context.Context, id string, user core.User) (core.Incident, error) {
	i, err := s.Incidents.IncidentByID(ctx, id)
	if err != nil {
		return core.Incident{}, err
	}
	if i.Status == "hidden" && user.ID != i.UserID && user.Role != "admin" {
		return core.Incident{}, core.ErrNotFound
	}
	i.ViewerIsAuthor = user.ID != "" && user.ID == i.UserID
	i.ViewerHasConfirmed = user.ID != "" && (i.VerifiedBy == user.ID || slices.Contains(i.ConfirmedBy, user.ID))
	return s.PublicationIncident(i), nil
}

func (s Service) ChangeStatus(ctx context.Context, user core.User, id, status string) (core.Incident, error) {
	if !slices.Contains([]string{"active", "resolved", "hidden"}, status) {
		return core.Incident{}, fmt.Errorf("%w: status must be active, resolved or hidden", core.ErrInvalid)
	}
	if status == "active" {
		current, err := s.Incidents.IncidentByID(ctx, id)
		if err != nil {
			return core.Incident{}, err
		}
		if !core.MapIncidentActive(current, s.Now()) {
			return core.Incident{}, core.ErrConflict
		}
		if err = s.Limits.Allow(ctx, "incident-update:"+user.ID, 6, time.Hour, s.Now()); err != nil {
			return core.Incident{}, err
		}
	}
	owner := user.ID
	if user.Role == "admin" {
		owner = ""
	}
	i, err := s.Incidents.SetIncidentStatus(ctx, id, owner, status, s.Now())
	i.ViewerIsAuthor = user.ID != "" && user.ID == i.UserID
	return core.PublicIncident(i), err
}

func (s Service) Alerts(ctx context.Context, user core.User, q core.IncidentQuery) (core.IncidentPage, error) {
	q.RadiusMeters = core.ProximityBands()[0]
	q.Type = ""
	q.IncludeInactive = false
	q.OwnerID = ""
	return s.FindIncidents(ctx, q)
}
