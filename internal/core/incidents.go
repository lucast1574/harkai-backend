package core

import "time"

type Point struct {
	Type        string    `json:"-" bson:"type"`
	Coordinates []float64 `json:"-" bson:"coordinates"`
}

type Incident struct {
	ID           string     `json:"id" bson:"_id"`
	UserID       string     `json:"-" bson:"user_id"`
	Type         string     `json:"type" bson:"type"`
	Description  string     `json:"description" bson:"description"`
	Latitude     float64    `json:"latitude" bson:"latitude"`
	Longitude    float64    `json:"longitude" bson:"longitude"`
	Location     Point      `json:"-" bson:"location"`
	Status       string     `json:"status" bson:"status"`
	MediaID      string     `json:"media_id,omitempty" bson:"media_id,omitempty"`
	ContactInfo  string     `json:"contact_info,omitempty" bson:"contact_info,omitempty"`
	ShareContact bool       `json:"-" bson:"share_contact"`
	District     string     `json:"district,omitempty" bson:"district,omitempty"`
	City         string     `json:"city,omitempty" bson:"city,omitempty"`
	Country      string     `json:"country,omitempty" bson:"country,omitempty"`
	Verified     bool       `json:"verified" bson:"verified"`
	CreatedAt    time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" bson:"updated_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
}

type IncidentInput struct {
	Type         string     `json:"type"`
	Description  string     `json:"description"`
	Latitude     float64    `json:"latitude"`
	Longitude    float64    `json:"longitude"`
	MediaID      string     `json:"media_id,omitempty"`
	ContactInfo  string     `json:"contact_info,omitempty"`
	ShareContact bool       `json:"share_contact"`
	District     string     `json:"district,omitempty"`
	City         string     `json:"city,omitempty"`
	Country      string     `json:"country,omitempty"`
	EndsAt       *time.Time `json:"ends_at,omitempty"`
}

type IncidentQuery struct {
	Latitude        float64
	Longitude       float64
	RadiusMeters    int
	Type            string
	Cursor          string
	Limit           int
	After           *time.Time
	OwnerID         string
	IncludeInactive bool
}

type IncidentPage struct {
	Items      []Incident `json:"items"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

func IncidentTypes() []string {
	return []string{"fire", "crash", "theft", "pet", "emergency", "place", "event"}
}

func PublicIncident(i Incident) Incident {
	if !i.ShareContact {
		i.ContactInfo = ""
	}
	return i
}

func CriticalType(t string) bool { return t == "fire" || t == "theft" || t == "emergency" }
