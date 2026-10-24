package application

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"math"
	"time"
)

func (s Service) VisitBusinesses(ctx context.Context, u core.User, d core.Device, lat, lng, accuracy float64, observed time.Time) ([]core.BusinessNotice, error) {
	return s.VisitNearbyPoints(ctx, u, d, lat, lng, accuracy, observed, false)
}
func (s Service) VisitNearbyPoints(ctx context.Context, u core.User, d core.Device, lat, lng, accuracy float64, observed time.Time, alerts bool) ([]core.BusinessNotice, error) {
	if s.Businesses == nil {
		return []core.BusinessNotice{}, nil
	}
	now := s.Now()
	if accuracy < 0 || accuracy > 100 || math.IsNaN(accuracy) || math.IsInf(accuracy, 0) || observed.Before(now.Add(-2*time.Minute)) || observed.After(now.Add(30*time.Second)) {
		return []core.BusinessNotice{}, nil
	}
	q := core.IncidentQuery{Latitude: lat, Longitude: lng, RadiusMeters: core.ProximityBands()[0], Limit: 100}
	eligible := map[string]int{}
	// Safety events are queued before promotional points.
	if alerts && s.Incidents != nil {
		page, err := s.Incidents.ListIncidents(ctx, q, now)
		if err != nil {
			return nil, err
		}
		for _, i := range page.Items {
			if i.UserID == u.ID || !core.MapIncidentActive(i, now) {
				continue
			}
			stage := core.ProximityStage(core.DistanceMeters(lat, lng, i.Latitude, i.Longitude), accuracy)
			if stage == 0 {
				continue
			}
			eligible[i.ID] = stage
			title, body := core.ProximityCopy(i.Type, "", "", "", stage)
			if err = s.queueProximity(ctx, u, d, i.ID, "incident", i.Type, title, body, stage, now); err != nil {
				return nil, err
			}
		}
	}
	page, err := s.Businesses.ListBusinesses(ctx, q, "", false, now)
	if err != nil {
		return nil, err
	}
	for _, p := range page.Items {
		if p.OwnerID == u.ID || !p.Public(now) {
			continue
		}
		stage := core.ProximityStage(core.DistanceMeters(lat, lng, p.Latitude, p.Longitude), accuracy)
		if stage == 0 {
			continue
		}
		eligible[p.ID] = stage
		title, body := core.ProximityCopy("business", p.Name, p.Offer, p.DiscountCode, stage)
		if err = s.queueProximity(ctx, u, d, p.ID, "business", p.Name, title, body, stage, now); err != nil {
			return nil, err
		}
	}
	pending, err := s.Businesses.BusinessNotices(ctx, u.ID, d.SessionID, d.ID, now)
	if err != nil {
		return nil, err
	}
	result := []core.BusinessNotice{}
	for _, n := range pending {
		if stage := eligible[n.PointID]; stage > 0 && (n.Stage == 0 || n.Stage >= stage) {
			result = append(result, n)
		}
	}
	return result, nil
}
func (s Service) queueProximity(ctx context.Context, u core.User, d core.Device, point, kind, name, title, body string, stage int, now time.Time) error {
	return s.Businesses.QueueBusinessNotice(ctx, core.BusinessNotice{ID: s.HashToken(u.ID + ":" + kind + ":" + point), UserID: u.ID, DeviceID: d.ID, SessionID: d.SessionID, PointID: point, Kind: kind, Name: name, Stage: stage, Title: title, Body: body, Pending: true, CreatedAt: now, NextEligibleAt: now.Add(24 * time.Hour)})
}
func (s Service) AckBusinessNotice(ctx context.Context, u core.User, access, id string) error {
	return s.AckProximityNotice(ctx, u, access, id, 0)
}
func (s Service) AckProximityNotice(ctx context.Context, u core.User, access, id string, stage int) error {
	if s.Businesses == nil || len(id) != 64 {
		return core.ErrInvalid
	}
	if stage != 0 && stage != 500 && stage != 250 && stage != 50 && stage != 15 {
		return core.ErrInvalid
	}
	session, err := s.Sessions.SessionByAccess(ctx, s.HashToken(access), s.Now())
	if err != nil {
		return err
	}
	return s.Businesses.AckProximityNotice(ctx, u.ID, session.ID, id, stage)
}
