package core

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

func ValidateCoordinates(lat, lng float64) error {
	if math.IsNaN(lat) || math.IsNaN(lng) || math.IsInf(lat, 0) || math.IsInf(lng, 0) || lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return fmt.Errorf("%w: coordinates", ErrInvalid)
	}
	return nil
}

func BuildIncident(input IncidentInput, id, userID string, now time.Time) (Incident, error) {
	if !slices.Contains(IncidentTypes(), input.Type) {
		return Incident{}, fmt.Errorf("%w: type", ErrInvalid)
	}
	if err := ValidateCoordinates(input.Latitude, input.Longitude); err != nil {
		return Incident{}, err
	}
	desc := strings.TrimSpace(input.Description)
	if utf8.RuneCountInString(desc) < 5 || utf8.RuneCountInString(desc) > 2000 {
		return Incident{}, fmt.Errorf("%w: description must contain 5 to 2000 characters", ErrInvalid)
	}
	if len(input.ContactInfo) > 40 || len(input.City) > 100 || len(input.District) > 100 || len(input.Country) > 100 {
		return Incident{}, fmt.Errorf("%w: contact or address too long", ErrInvalid)
	}
	if input.Type == "pet" || input.Type == "place" {
		if len(strings.TrimSpace(input.ContactInfo)) > 0 && len(strings.TrimSpace(input.ContactInfo)) < 6 {
			return Incident{}, fmt.Errorf("%w: contact_info must contain at least six characters", ErrInvalid)
		}
	}
	expiry, err := IncidentExpiry(input.Type, input.EndsAt, now)
	if err != nil {
		return Incident{}, err
	}
	return Incident{ID: id, UserID: userID, Type: input.Type, Description: desc, Latitude: input.Latitude, Longitude: input.Longitude, Location: Point{Type: "Point", Coordinates: []float64{input.Longitude, input.Latitude}}, Status: "active", MediaID: input.MediaID, ContactInfo: strings.TrimSpace(input.ContactInfo), ShareContact: input.ShareContact, District: strings.TrimSpace(input.District), City: strings.TrimSpace(input.City), Country: strings.TrimSpace(input.Country), CreatedAt: now, UpdatedAt: now, ExpiresAt: expiry}, nil
}

func IncidentExpiry(kind string, endsAt *time.Time, now time.Time) (*time.Time, error) {
	if endsAt != nil {
		return nil, fmt.Errorf("%w: ends_at is no longer supported", ErrInvalid)
	}
	duration := 3 * time.Hour
	t := now.Add(duration)
	return &t, nil
}

func ValidateQuery(q IncidentQuery) (IncidentQuery, error) {
	q.District = strings.TrimSpace(q.District)
	if len(q.District) > 100 || (q.After != nil && q.Before != nil && !q.Before.After(*q.After)) {
		return q, fmt.Errorf("%w: district or date range", ErrInvalid)
	}
	if err := ValidateCoordinates(q.Latitude, q.Longitude); err != nil {
		return q, err
	}
	if q.RadiusMeters < 100 || q.RadiusMeters > 100000 {
		return q, fmt.Errorf("%w: radius_meters must be 100 to 100000", ErrInvalid)
	}
	if q.Limit < 1 || q.Limit > 100 {
		return q, fmt.Errorf("%w: limit must be 1 to 100", ErrInvalid)
	}
	if q.Type != "" && q.Type != "crime" && !slices.Contains(IncidentTypes(), q.Type) {
		return q, fmt.Errorf("%w: type", ErrInvalid)
	}
	if q.Cursor != "" && (len(q.Cursor) != 24 || strings.Trim(q.Cursor, "0123456789abcdef") != "") {
		return q, fmt.Errorf("%w: cursor", ErrInvalid)
	}
	return q, nil
}
