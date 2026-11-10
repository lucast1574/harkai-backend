package stripebilling

import (
	"context"
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/stripe/stripe-go/v87/webhook"
	"strings"
	"time"
)

func (c *Client) VerifyBusinessEntitlement(ctx context.Context, raw []byte, signature string) (core.BusinessEntitlement, error) {
	e, err := webhook.ConstructEvent(raw, signature, c.Config.WebhookSecret)
	if err != nil {
		return core.BusinessEntitlement{}, core.ErrInvalid
	}
	if e.Livemode != strings.HasPrefix(c.Config.Secret, "sk_live_") {
		return core.BusinessEntitlement{}, core.ErrInvalid
	}
	if e.Type != "checkout.session.completed" && e.Type != "charge.refunded" {
		return core.BusinessEntitlement{}, core.ErrNotFound
	}
	var payment string
	var checkoutID string
	var metadata map[string]string
	status := "active"
	paid := time.Unix(e.Created, 0).UTC()
	if e.Type == "checkout.session.completed" {
		var s session
		if json.Unmarshal(e.Data.Raw, &s) != nil {
			return core.BusinessEntitlement{}, core.ErrInvalid
		}
		if s.Metadata["app"] != "harkai" {
			return core.BusinessEntitlement{}, core.ErrNotFound
		}
		if json.Unmarshal(e.Data.Raw, &s) != nil || s.Mode != "payment" || s.PaymentStatus != "paid" || s.Currency != "usd" || s.Amount != 500 || s.PaymentIntent == "" {
			return core.BusinessEntitlement{}, core.ErrInvalid
		}
		payment, metadata = s.PaymentIntent, s.Metadata
		checkoutID = s.ID
		if checkoutID == "" {
			return core.BusinessEntitlement{}, core.ErrInvalid
		}
	} else {
		var charge struct {
			PaymentIntent string            `json:"payment_intent"`
			Refunded      bool              `json:"refunded"`
			Metadata      map[string]string `json:"metadata"`
			Currency      string            `json:"currency"`
			Amount        int64             `json:"amount"`
		}
		if json.Unmarshal(e.Data.Raw, &charge) != nil {
			return core.BusinessEntitlement{}, core.ErrInvalid
		}
		if !charge.Refunded {
			return core.BusinessEntitlement{}, core.ErrNotFound
		}
		if charge.Currency != "usd" || charge.Amount != 500 {
			return core.BusinessEntitlement{}, core.ErrNotFound
		}
		payment, metadata, status = charge.PaymentIntent, charge.Metadata, "refunded"
	}
	// Other Nexode products share this account, but never grant Harkai access.
	if metadata["app"] != "harkai" {
		return core.BusinessEntitlement{}, core.ErrNotFound
	}
	if payment == "" {
		return core.BusinessEntitlement{}, core.ErrInvalid
	}
	return core.BusinessEntitlement{CheckoutID: checkoutID, PointID: metadata["point_id"], OwnerID: metadata["owner_id"], PaymentID: payment, EventID: e.ID, EventAt: paid, PaidAt: &paid, Mode: "one_time", Status: status}, nil
}
