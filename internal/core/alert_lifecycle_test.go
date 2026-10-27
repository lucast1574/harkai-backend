package core

import (
	"testing"
	"time"
)

func TestNinetyMinuteMapBoundaryAndArchive(t *testing.T) {
	now := time.Now().UTC()
	legacyEnd := now.Add(3 * time.Hour)
	i := Incident{ID: "report", CreatedAt: now, Status: "active", ExpiresAt: &legacyEnd, Description: "historical conversation stays"}
	if !MapIncidentActive(i, now.Add(AlertLifetime-time.Nanosecond)) || MapIncidentActive(i, now.Add(AlertLifetime)) {
		t.Fatal("90-minute boundary")
	}
	public := PublicIncident(i)
	if !public.ExpiresAt.Equal(now.Add(AlertLifetime)) || public.Description != i.Description || public.ID != i.ID {
		t.Fatal("legacy expiry or archive payload")
	}
}

func TestManualReportCorrectionStaysWithinCategory(t *testing.T) {
	for _, kind := range CrimeTypes() {
		if got, err := ManualReportType(kind, "crime"); err != nil || got != kind {
			t.Fatal(kind, err)
		}
	}
	if _, err := ManualReportType("fire", "crime"); err == nil {
		t.Fatal("cross category correction")
	}
	if _, err := ManualReportType("unknown", "unknown"); err == nil {
		t.Fatal("unknown type")
	}
	if got, err := ManualReportType("pet", "pet"); err != nil || got != "pet" {
		t.Fatal(got, err)
	}
}
