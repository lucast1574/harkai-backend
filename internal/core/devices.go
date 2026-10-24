package core

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type DeviceInput struct {
	Token     string  `json:"token"`
	Platform  string  `json:"platform"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
type Device struct {
	ID                string    `json:"id" bson:"_id"`
	UserID            string    `json:"-" bson:"user_id"`
	SessionID         string    `json:"-" bson:"session_id"`
	Token             string    `json:"-" bson:"token"`
	Generation        string    `json:"-" bson:"generation"`
	Platform          string    `json:"platform" bson:"platform"`
	Location          Point     `json:"-" bson:"location"`
	UpdatedAt         time.Time `json:"updated_at" bson:"updated_at"`
	LocationUpdatedAt time.Time `json:"location_updated_at" bson:"location_updated_at"`
	ExpiresAt         time.Time `json:"-" bson:"expires_at"`
}

// The job is embedded in the report so publication and queueing are one write.
type PushJob struct {
	State    string    `bson:"state"`
	Cursor   string    `bson:"cursor"`
	Lease    string    `bson:"lease"`
	NextAt   time.Time `bson:"next_at"`
	Attempts int       `bson:"attempts"`
}

func ValidateDevice(v DeviceInput) (DeviceInput, error) {
	if len(v.Token) < 20 || len(v.Token) > 4096 || strings.ContainsAny(v.Token, " \r\n\t") {
		return v, fmt.Errorf("%w: device token", ErrInvalid)
	}
	if v.Platform != "android" && v.Platform != "web" {
		return v, fmt.Errorf("%w: platform", ErrInvalid)
	}
	return v, ValidateCoordinates(v.Latitude, v.Longitude)
}
func PushEligible(i Incident, d Device, u User, now time.Time) bool {
	if u.ID == i.UserID || i.Status != "active" || now.Sub(i.CreatedAt) >= 3*time.Hour || (i.ExpiresAt != nil && !i.ExpiresAt.After(now)) || !d.ExpiresAt.After(now) {
		return false
	}
	if len(d.Location.Coordinates) != 2 {
		return false
	}
	return DistanceMeters(i.Latitude, i.Longitude, d.Location.Coordinates[1], d.Location.Coordinates[0]) <= float64(ProximityBands()[0])
}
func DistanceMeters(a, b, c, d float64) float64 {
	const r = math.Pi / 180
	x, y := (c-a)*r, (d-b)*r
	h := math.Sin(x/2)*math.Sin(x/2) + math.Cos(a*r)*math.Cos(c*r)*math.Sin(y/2)*math.Sin(y/2)
	return 6371008.8 * 2 * math.Asin(math.Sqrt(math.Min(1, h)))
}
