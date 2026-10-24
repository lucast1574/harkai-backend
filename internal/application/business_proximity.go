package application

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"math"
	"time"
)

func (s Service) VisitBusinesses(ctx context.Context, u core.User, d core.Device, lat, lng, accuracy float64, observed time.Time) ([]core.BusinessNotice, error) {
	if s.Businesses == nil || !u.Preferences.BusinessNotifications || u.Preferences.Notifications == "none" {
		return []core.BusinessNotice{}, nil
	}
	now := s.Now()
	if accuracy < 0 || accuracy > 100 || math.IsNaN(accuracy) || math.IsInf(accuracy, 0) || observed.Before(now.Add(-2*time.Minute)) || observed.After(now.Add(30*time.Second)) {
		return []core.BusinessNotice{}, nil
	}
	page, err := s.Businesses.ListBusinesses(ctx, core.IncidentQuery{Latitude: lat, Longitude: lng, RadiusMeters: 1000, Limit: 100}, "", false, now)
	if err != nil {
		return nil, err
	}
	eligible := map[string]bool{}
	for _, p := range page.Items {
		if p.OwnerID == u.ID || !p.Public(now) || core.DistanceMeters(lat, lng, p.Latitude, p.Longitude) > float64(p.RadiusMeters) {
			continue
		}
		eligible[p.ID] = true
		notice := core.BusinessNotice{ID: s.HashToken(u.ID + ":" + p.ID), UserID: u.ID, DeviceID: d.ID, SessionID: d.SessionID, PointID: p.ID, Name: p.Name, Pending: true, CreatedAt: now, NextEligibleAt: now.Add(24 * time.Hour)}
		if err = s.Businesses.QueueBusinessNotice(ctx, notice); err != nil {
			return nil, err
		}
	}
	pending, err := s.Businesses.BusinessNotices(ctx, u.ID, d.SessionID, d.ID, now)
	if err != nil {
		return nil, err
	}
	result := []core.BusinessNotice{}
	for _, notice := range pending {
		if eligible[notice.PointID] {
			result = append(result, notice)
		}
	}
	return result, nil
}
func (s Service) AckBusinessNotice(ctx context.Context, u core.User, access, id string) error {
	if s.Businesses == nil || len(id) != 64 {
		return core.ErrInvalid
	}
	session, err := s.Sessions.SessionByAccess(ctx, s.HashToken(access), s.Now())
	if err != nil {
		return err
	}
	return s.Businesses.AckBusinessNotice(ctx, u.ID, session.ID, id)
}
