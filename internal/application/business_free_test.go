package application_test

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/application"
	"github.com/lucast1574/harkai-backend/internal/core"
	"testing"
	"time"
)

type businessMemory struct {
	application.BusinessStore
	point core.BusinessPoint
}

func (s *businessMemory) CreateBusiness(_ context.Context, p core.BusinessPoint) (core.BusinessPoint, error) {
	s.point = p
	return p, nil
}
func (s *businessMemory) UpdateBusiness(_ context.Context, p core.BusinessPoint) (core.BusinessPoint, error) {
	s.point = p
	return p, nil
}
func (s *businessMemory) BusinessByID(context.Context, string) (core.BusinessPoint, error) {
	return s.point, nil
}

type businessLimits struct{}

func (businessLimits) Allow(context.Context, string, int, time.Duration, time.Time) error { return nil }

type configuredBilling struct{}

func (configuredBilling) VerifyBusinessEntitlement(context.Context, []byte, string) (core.BusinessEntitlement, error) {
	return core.BusinessEntitlement{}, core.ErrInvalid
}

func TestBusinessFreeLaunchAndFutureBilling(t *testing.T) {
	store := &businessMemory{}
	s := application.Service{Businesses: store, Limits: businessLimits{}, Now: time.Now, ID: func() string { return "123456789012345678901234" }}
	in := core.BusinessInput{Name: "Café vecinal", Description: "Café y pan artesanal", Kind: "restaurant", Latitude: -12, Longitude: -77}
	user := core.User{ID: "owner"}
	p, err := s.SaveBusiness(context.Background(), user, "", in)
	if err != nil || p.PaymentMode != "free" || p.BillingStatus != "active" || p.Public(time.Now()) {
		t.Fatalf("free submission: %+v %v", p, err)
	}
	s.Billing = configuredBilling{}
	p, err = s.SaveBusiness(context.Background(), user, "", in)
	if err != nil || p.PaymentMode != "one_time" || p.BillingStatus != "pending" {
		t.Fatalf("billing bypassed: %+v %v", p, err)
	}
	s.Billing = nil
	p, err = s.SaveBusiness(context.Background(), user, p.ID, in)
	if err != nil || p.PaymentMode != "free" {
		t.Fatalf("legacy unpaid draft not upgraded: %+v %v", p, err)
	}
	store.point.BillingStatus = "refunded"
	store.point.PaymentMode = "one_time"
	store.point.BillingEventID = "refund"
	p, err = s.SaveBusiness(context.Background(), user, p.ID, in)
	if err != nil || p.BillingStatus != "refunded" || p.PaymentMode != "one_time" {
		t.Fatal("edit overwrote existing billing", err)
	}
}
