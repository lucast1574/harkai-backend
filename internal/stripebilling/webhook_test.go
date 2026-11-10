package stripebilling

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/core"
	stripe "github.com/stripe/stripe-go/v87"
	"github.com/stripe/stripe-go/v87/webhook"
	"testing"
	"time"
)

func signedEvent(t *testing.T, typ string, object any) ([]byte, string) {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"id": "evt_harkai", "object": "event", "type": typ, "api_version": stripe.APIVersion, "created": time.Now().Unix(), "data": map[string]any{"object": object}})
	if err != nil {
		t.Fatal(err)
	}
	s := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: raw, Secret: "whsec_test"})
	return raw, s.Header
}
func TestOnlySignedPaidHarkaiPaymentsGrantAccess(t *testing.T) {
	c := New(Config{WebhookSecret: "whsec_test"})
	valid := map[string]any{"id": "cs_harkai", "mode": "payment", "payment_status": "paid", "currency": "usd", "amount_total": 500, "payment_intent": "pi_harkai", "metadata": map[string]string{"app": "harkai", "point_id": "507f1f77bcf86cd799439011", "owner_id": "owner"}}
	for _, tc := range []struct {
		name, key string
		value     any
		want      error
	}{
		{"valid", "", nil, nil}, {"unpaid", "payment_status", "unpaid", core.ErrInvalid}, {"wrong amount", "amount_total", 1, core.ErrInvalid}, {"currency", "currency", "eur", core.ErrInvalid}, {"subscription", "mode", "subscription", core.ErrInvalid}, {"other app", "metadata", map[string]string{"app": "nexode"}, core.ErrNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := map[string]any{}
			for k, v := range valid {
				obj[k] = v
			}
			if tc.key != "" {
				obj[tc.key] = tc.value
			}
			raw, sig := signedEvent(t, "checkout.session.completed", obj)
			e, err := c.VerifyBusinessEntitlement(context.Background(), raw, sig)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
			if err == nil {
				if err = core.ValidateBusinessEntitlement(e, time.Now()); err != nil {
					t.Fatal(err)
				}
			}
			if _, err = c.VerifyBusinessEntitlement(context.Background(), append(append([]byte{}, raw...), ' '), sig); err == nil {
				t.Fatal("tampering accepted")
			}
		})
	}
}
func TestFullRefundRevokesAndPartialRefundDoesNot(t *testing.T) {
	c := New(Config{WebhookSecret: "whsec_test"})
	for _, full := range []bool{true, false} {
		raw, sig := signedEvent(t, "charge.refunded", map[string]any{"refunded": full, "currency": "usd", "amount": 500, "payment_intent": "pi_harkai", "metadata": map[string]string{"app": "harkai", "point_id": "507f1f77bcf86cd799439011", "owner_id": "owner"}})
		e, err := c.VerifyBusinessEntitlement(context.Background(), raw, sig)
		if full && (err != nil || e.Status != "refunded") {
			t.Fatalf("full refund: %v %+v", err, e)
		}
		if !full && !errors.Is(err, core.ErrNotFound) {
			t.Fatal("partial refund changed access")
		}
	}
}
