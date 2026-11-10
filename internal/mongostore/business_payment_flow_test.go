package mongostore_test

import (
	"context"
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/application"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"github.com/lucast1574/harkai-backend/internal/stripebilling"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	stripe "github.com/stripe/stripe-go/v87"
	"github.com/stripe/stripe-go/v87/webhook"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWebPaymentGatesMobileManagementAndRefundsSurviveRetries(t *testing.T) {
	store := testutil.Store(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/prices/") {
			_, _ = w.Write([]byte(`{"id":"price_harkai","product":"prod_harkai_map_listing","unit_amount":500,"currency":"usd","type":"one_time","active":true}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":"cs_harkai","status":"open","url":"https://checkout.stripe.com/c/pay/test"}`))
	}))
	defer provider.Close()
	billing := stripebilling.New(stripebilling.Config{Secret: "sk_test_key", WebhookSecret: "whsec_test", Price: "price_harkai"})
	billing.BaseURL = provider.URL + "/"
	s := application.Service{Businesses: store, Checkouts: store, Checkout: billing, Billing: billing, Limits: store, ID: mongostore.NewID, Now: func() time.Time { return now }}
	u := core.User{ID: "owner"}
	admin := core.User{ID: "admin", Role: "admin"}
	in := core.BusinessInput{Name: "Negocio", Kind: "restaurant", Description: "Comida del barrio", Latitude: -12, Longitude: -77}
	p, err := s.SaveBusiness(ctx, u, "", in)
	if err != nil {
		t.Fatal(err)
	}
	if p.BillingStatus != "pending" || p.PaymentMode != "one_time" {
		t.Fatal("new registration bypassed payment")
	}
	if _, err = s.ReviewBusiness(ctx, admin, p.ID, "approved"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CheckoutBusiness(ctx, u, p.ID); err != nil {
		t.Fatal(err)
	}
	p, _ = store.BusinessByID(ctx, p.ID)
	if p.Public(now) || p.BillingStatus != "pending" {
		t.Fatal("checkout redirect activated access")
	}
	object := map[string]any{"id": "cs_harkai", "mode": "payment", "payment_status": "paid", "currency": "usd", "amount_total": 500, "payment_intent": "pi_harkai", "metadata": map[string]string{"app": "harkai", "point_id": p.ID, "owner_id": u.ID}}
	paid, paidSignature := flowEvent(t, "evt_paid", "checkout.session.completed", now, object)
	if err = s.BusinessBillingEvent(ctx, paid, "forged"); err == nil {
		t.Fatal("unsigned event activated access")
	}
	if err = s.BusinessBillingEvent(ctx, paid, paidSignature); err != nil {
		t.Fatal(err)
	}
	if err = s.BusinessBillingEvent(ctx, paid, paidSignature); err != nil {
		t.Fatal("retry failed", err)
	}
	p, _ = store.BusinessByID(ctx, p.ID)
	if !p.Public(now) {
		t.Fatal("paid approved listing not public")
	}
	in.Name = "Nombre actualizado"
	p, err = s.SaveBusiness(ctx, u, p.ID, in)
	if err != nil || p.BillingStatus != "active" || p.PaymentID != "pi_harkai" || p.Moderation != "pending" {
		t.Fatal("mobile edit lost access", err)
	}
	refund := map[string]any{"payment_intent": "pi_harkai", "refunded": true, "amount": 500, "currency": "usd", "metadata": object["metadata"]}
	raw, sig := flowEvent(t, "evt_refund", "charge.refunded", now.Add(time.Second), refund)
	now = now.Add(2 * time.Second)
	if err = s.BusinessBillingEvent(ctx, raw, sig); err != nil {
		t.Fatal(err)
	}
	if err = s.BusinessBillingEvent(ctx, paid, paidSignature); err != nil {
		t.Fatal("stale event not acknowledged", err)
	}
	p, _ = store.BusinessByID(ctx, p.ID)
	if p.BillingStatus != "refunded" || p.Public(now) {
		t.Fatal("old paid event restored refunded access")
	}
}
func flowEvent(t *testing.T, id, kind string, at time.Time, object any) ([]byte, string) {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"id": id, "object": "event", "type": kind, "api_version": stripe.APIVersion, "created": at.Unix(), "data": map[string]any{"object": object}})
	if err != nil {
		t.Fatal(err)
	}
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: raw, Secret: "whsec_test"})
	return raw, signed.Header
}
