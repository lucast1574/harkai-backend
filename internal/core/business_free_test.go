package core

import (
	"testing"
	"time"
)

func TestFreeBusinessStillRequiresReviewAndEnabled(t *testing.T) {
	now := time.Now()
	p := BusinessPoint{PaymentMode: "free", BillingStatus: "active", Enabled: true, Moderation: "pending"}
	if p.Public(now) {
		t.Fatal("unreviewed free point is public")
	}
	p.Moderation = "approved"
	if !p.Public(now) || !p.Public(now.AddDate(1, 0, 0)) {
		t.Fatal("approved free point requires payment")
	}
	p.Enabled = false
	if p.Public(now) {
		t.Fatal("paused free point is public")
	}
	p.Enabled = true
	p.BillingStatus = "refunded"
	if p.Public(now) {
		t.Fatal("inactive entitlement is public")
	}
}
