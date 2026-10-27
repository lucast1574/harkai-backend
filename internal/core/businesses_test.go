package core

import (
	"math"
	"testing"
	"time"
)

func TestBusinessRequiresReviewAndOnePayment(t *testing.T) {
	now := time.Now().UTC()
	in := BusinessInput{Name: "Clínica vecinal", Kind: "health", Description: "Atención odontológica", Latitude: -12, Longitude: -77, RadiusMeters: 250}
	p, err := BuildBusiness(in, "123456789012345678901234", "owner", now)
	if err != nil {
		t.Fatal(err)
	}
	if p.Public(now) {
		t.Fatal("unpaid draft is public")
	}
	p.Moderation = "approved"
	if p.Public(now) {
		t.Fatal("review minted billing rights")
	}
	p.BillingStatus = "active"
	p.PaidAt = &now
	if !p.Public(now) {
		t.Fatal("paid approved point hidden")
	}
	p.Enabled = false
	if p.Public(now) {
		t.Fatal("paused point public")
	}
	p.Enabled = true
	if !p.Public(now.Add(365 * 24 * time.Hour)) {
		t.Fatal("one-time payment unexpectedly expired")
	}
	p.BillingStatus = "past_due"
	if p.Public(now) {
		t.Fatal("past due point public")
	}
	in.Latitude = math.NaN()
	if _, err = BuildBusiness(in, "id", "owner", now); err == nil {
		t.Fatal("invalid coordinate")
	}
}
func TestEveryIncidentHasThreeHourMapLifetimeWithoutPhoto(t *testing.T) {
	now := time.Now().UTC()
	for _, kind := range IncidentTypes() {
		i, err := BuildIncident(IncidentInput{Type: kind, Description: "Información de la comunidad", Latitude: -12, Longitude: -77}, "123456789012345678901234", "user", now)
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		if i.ExpiresAt == nil || !i.ExpiresAt.Equal(now.Add(AlertLifetime)) {
			t.Fatalf("%s has wrong lifetime", kind)
		}
	}
}
