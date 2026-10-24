package core

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

func BuildBusiness(in BusinessInput, id, owner string, now time.Time) (BusinessPoint, error) {
	in.Name, in.Description = strings.TrimSpace(in.Name), strings.TrimSpace(in.Description)
	if utf8.RuneCountInString(in.Name) < 2 || len(in.Name) > 200 || utf8.RuneCountInString(in.Description) < 5 || utf8.RuneCountInString(in.Description) > 2000 || !slices.Contains(BusinessKinds(), in.Kind) {
		return BusinessPoint{}, fmt.Errorf("%w: business name, description or kind", ErrInvalid)
	}
	if err := ValidateCoordinates(in.Latitude, in.Longitude); err != nil {
		return BusinessPoint{}, err
	}
	if in.RadiusMeters < 100 || in.RadiusMeters > 1000 || len(in.Subtype) > 80 || len(in.Address) > 300 || len(in.City) > 100 || len(in.District) > 100 || len(in.Phone) > 30 || len(in.Website) > 300 {
		return BusinessPoint{}, fmt.Errorf("%w: business details or notification radius", ErrInvalid)
	}
	if in.Website != "" {
		u, err := url.Parse(in.Website)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
			return BusinessPoint{}, fmt.Errorf("%w: website must use HTTPS", ErrInvalid)
		}
	}
	for _, c := range in.Phone {
		if !strings.ContainsRune("+0123456789 ()-", c) {
			return BusinessPoint{}, fmt.Errorf("%w: phone", ErrInvalid)
		}
	}
	return BusinessPoint{ID: id, OwnerID: owner, Name: in.Name, Kind: in.Kind, Subtype: strings.TrimSpace(in.Subtype), Description: in.Description, Address: strings.TrimSpace(in.Address), City: strings.TrimSpace(in.City), District: strings.TrimSpace(in.District), Latitude: in.Latitude, Longitude: in.Longitude, Location: Point{Type: "Point", Coordinates: []float64{in.Longitude, in.Latitude}}, RadiusMeters: in.RadiusMeters, MediaID: in.MediaID, Phone: in.Phone, Website: in.Website, Moderation: "pending", BillingStatus: "pending", Enabled: true, CreatedAt: now, UpdatedAt: now}, nil
}
func ValidateBusinessEntitlement(e BusinessEntitlement, now time.Time) error {
	if !ValidID(e.PointID) || e.OwnerID == "" || e.EventID == "" || len(e.EventID) > 200 || e.SubscriptionID == "" || len(e.SubscriptionID) > 200 || e.EventAt.IsZero() || e.EventAt.After(now.Add(time.Minute)) {
		return ErrInvalid
	}
	if !slices.Contains([]string{"active", "past_due", "canceled", "expired"}, e.Status) {
		return ErrInvalid
	}
	if e.Status == "active" && (!e.PaidUntil.After(now) || e.PaidUntil.After(now.Add(370*24*time.Hour))) {
		return ErrInvalid
	}
	return nil
}
