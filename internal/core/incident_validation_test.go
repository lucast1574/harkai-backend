package core

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"
)

func TestIncidentContractAndExpiry(t *testing.T) {
	now := time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)
	for _, kind := range IncidentTypes() {
		t.Run(kind, func(t *testing.T) {
			input := IncidentInput{Type: kind, Description: "Reporte de prueba", Latitude: -8.1, Longitude: -79.0, MediaID: "image", ContactInfo: "123456789"}
			end := now.Add(48 * time.Hour)
			if kind == "event" {
				input.EndsAt = &end
			}
			i, err := BuildIncident(input, "id", "owner", now)
			if err != nil {
				t.Fatal(err)
			}
			if i.Type != kind || i.Verified || i.Location.Coordinates[0] != -79.0 || i.Location.Coordinates[1] != -8.1 {
				t.Fatalf("invalid contract: %+v", i)
			}
			if i.ExpiresAt == nil || !i.ExpiresAt.Equal(now.Add(3*time.Hour)) {
				t.Fatal("all alert markers must expire after three hours")
			}
			data, _ := json.Marshal(PublicIncident(i))
			var output map[string]any
			_ = json.Unmarshal(data, &output)
			if _, exists := output["contact_info"]; exists {
				t.Fatal("private contact leaked")
			}
			if _, exists := output["user_id"]; exists {
				t.Fatal("owner leaked")
			}
		})
	}
}

func TestRejectInvalidIncidents(t *testing.T) {
	base := IncidentInput{Type: "fire", Description: "A fire is visible", Latitude: 0, Longitude: 0}
	cases := []IncidentInput{base, base, base, base, base}
	cases[0].Type = "0"
	cases[1].Latitude = math.NaN()
	cases[2].Longitude = 181
	cases[3].Description = "a"
	cases[4].Type = "event"
	for _, input := range cases {
		if _, err := BuildIncident(input, "id", "user", time.Now()); !errors.Is(err, ErrInvalid) {
			t.Errorf("expected invalid for %+v, got %v", input, err)
		}
	}
}

func TestQueryBounds(t *testing.T) {
	q := IncidentQuery{RadiusMeters: 500, Limit: 50}
	if _, err := ValidateQuery(q); err != nil {
		t.Fatal(err)
	}
	q.RadiusMeters = 100001
	if _, err := ValidateQuery(q); !errors.Is(err, ErrInvalid) {
		t.Fatal("unbounded geo query accepted")
	}
	q.RadiusMeters = 500
	q.Cursor = "not-a-valid-cursor"
	if _, err := ValidateQuery(q); !errors.Is(err, ErrInvalid) {
		t.Fatal("invalid cursor accepted")
	}
}
