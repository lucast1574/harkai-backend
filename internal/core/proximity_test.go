package core

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestGlobalProximityBandsRespectAccuracy(t *testing.T) {
	for _, c := range []struct {
		distance, accuracy float64
		want               int
	}{{510, 1, 0}, {480, 10, 500}, {230, 10, 250}, {40, 5, 50}, {0, 10, 15}, {45, 10, 250}, {0, 60, 250}, {0, 101, 0}, {0, math.NaN(), 0}, {-1, 0, 0}} {
		if got := ProximityStage(c.distance, c.accuracy); got != c.want {
			t.Fatalf("%+v got %d", c, got)
		}
	}
}
func TestPointMessagesStaySpecificAndUnverified(t *testing.T) {
	for _, kind := range IncidentTypes() {
		title, body := ProximityCopy(kind, "", "", "", 50)
		if title == "" || !strings.Contains(body, "sin verificar") {
			t.Fatal(kind, title, body)
		}
		if kind == "pet" && strings.Contains(body, "seguro") {
			t.Fatal("pet should not cause an evacuation claim")
		}
	}
	_, body := ProximityCopy("business", "Café", "15 % en tu visita", "VECINO15", 250)
	if !strings.Contains(body, "Promocionado") || !strings.Contains(body, "VECINO15") {
		t.Fatal(body)
	}
	_, body = ProximityCopy("business", "Café", "", "", 500)
	if strings.Contains(body, "oferta") || strings.Contains(body, "Código") {
		t.Fatal("invented discount", body)
	}
}
func TestOneTimeBillingValidation(t *testing.T) {
	now := time.Now().UTC()
	paid := now.Add(-30 * 24 * time.Hour)
	e := BusinessEntitlement{PointID: "123456789012345678901234", OwnerID: "owner", Mode: "one_time", PaymentID: "pi_test", EventID: "evt_test", Status: "active", EventAt: now, PaidAt: &paid}
	if err := ValidateBusinessEntitlement(e, now); err != nil {
		t.Fatal(err)
	}
	e.Mode = "subscription"
	if ValidateBusinessEntitlement(e, now) == nil {
		t.Fatal("subscription accepted")
	}
	e.Mode = "one_time"
	e.PaymentID = ""
	if ValidateBusinessEntitlement(e, now) == nil {
		t.Fatal("missing payment accepted")
	}
}
