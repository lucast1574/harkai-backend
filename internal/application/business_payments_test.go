package application_test

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/application"
	"github.com/lucast1574/harkai-backend/internal/core"
	"testing"
	"time"
)

type memoryCheckout struct {
	record core.BusinessCheckout
	saved  int
}

func (m *memoryCheckout) BusinessCheckout(_ context.Context, c core.BusinessCheckout) (core.BusinessCheckout, error) {
	if m.record.Key == "" {
		m.record = c
	}
	return m.record, nil
}
func (m *memoryCheckout) BusinessCheckoutByID(context.Context, string) (core.BusinessCheckout, error) {
	return m.record, nil
}
func (m *memoryCheckout) SaveBusinessCheckout(_ context.Context, _ core.BusinessCheckout, id string) error {
	m.record.SessionID = id
	m.saved++
	return nil
}
func (m *memoryCheckout) RotateBusinessCheckout(_ context.Context, old core.BusinessCheckout, key string) error {
	m.record.Key = key
	m.record.SessionID = ""
	return nil
}

type checkoutProvider struct {
	calls  int
	expire bool
	keys   []string
}

func (p *checkoutProvider) Checkout(_ context.Context, _ core.BusinessPoint, r core.BusinessCheckout) (core.CheckoutSession, error) {
	p.calls++
	p.keys = append(p.keys, r.Key)
	if p.expire && p.calls == 1 {
		return core.CheckoutSession{ID: "cs_expired", Status: "expired"}, nil
	}
	return core.CheckoutSession{ID: "cs_open", Status: "open", URL: "https://checkout.stripe.com/test"}, nil
}
func TestCheckoutRequiresOwnerApprovalAndNoExistingAccess(t *testing.T) {
	id := "507f1f77bcf86cd799439011"
	for _, tc := range []struct {
		owner, moderation, billing string
		want                       error
	}{
		{"other", "approved", "pending", core.ErrForbidden},
		{"owner", "pending", "pending", core.ErrConflict},
		{"owner", "rejected", "pending", core.ErrConflict},
		{"owner", "approved", "active", core.ErrConflict},
		{"owner", "approved", "refunded", core.ErrConflict},
		{"owner", "approved", "pending", nil},
	} {
		store := &businessMemory{point: core.BusinessPoint{ID: id, OwnerID: tc.owner, Moderation: tc.moderation, BillingStatus: tc.billing}}
		provider := &checkoutProvider{}
		records := &memoryCheckout{}
		s := application.Service{Businesses: store, Checkout: provider, Checkouts: records, Limits: businessLimits{}, ID: func() string { return "attempt" }, Now: time.Now}
		_, err := s.CheckoutBusiness(context.Background(), core.User{ID: "owner"}, id)
		if !errors.Is(err, tc.want) {
			t.Fatalf("%+v: %v", tc, err)
		}
		if tc.want != nil && provider.calls != 0 {
			t.Fatal("charged an ineligible business")
		}
		if tc.want == nil {
			if _, err = s.CheckoutBusiness(context.Background(), core.User{ID: "owner"}, id); err != nil {
				t.Fatal(err)
			}
			if len(provider.keys) != 2 || provider.keys[0] != provider.keys[1] || records.record.SessionID != "cs_open" {
				t.Fatal("retry lost its session identity")
			}
		}
	}
}
func TestOnlyExpiredCheckoutRotatesAttempt(t *testing.T) {
	store := &businessMemory{point: core.BusinessPoint{ID: "507f1f77bcf86cd799439011", OwnerID: "owner", Moderation: "approved", BillingStatus: "pending"}}
	p := &checkoutProvider{expire: true}
	m := &memoryCheckout{}
	sequence := 0
	s := application.Service{Businesses: store, Checkout: p, Checkouts: m, Limits: businessLimits{}, Now: time.Now, ID: func() string { sequence++; return time.Unix(int64(sequence), 0).String() }}
	r, err := s.CheckoutBusiness(context.Background(), core.User{ID: "owner"}, store.point.ID)
	if err != nil || r.Status != "open" || len(p.keys) != 2 || p.keys[0] == p.keys[1] {
		t.Fatalf("%+v %v", r, err)
	}
}
