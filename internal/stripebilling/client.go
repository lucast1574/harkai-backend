package stripebilling

import (
	"context"
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/core"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct{ Secret, WebhookSecret, Price, SuccessURL, CancelURL string }
type Client struct {
	Config  Config
	HTTP    *http.Client
	BaseURL string
}

func New(c Config) *Client {
	return &Client{Config: c, HTTP: &http.Client{Timeout: 20 * time.Second}, BaseURL: "https://api.stripe.com/v1/"}
}
func (c *Client) request(ctx context.Context, method, path string, form url.Values, key string, out any) error {
	r, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return core.ErrUnavailable
	}
	r.SetBasicAuth(c.Config.Secret, "")
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Stripe-Version", "2026-09-30.endive")
	if key != "" {
		r.Header.Set("Idempotency-Key", key)
	}
	resp, err := c.HTTP.Do(r)
	if err != nil {
		return core.ErrUnavailable
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return core.ErrUnavailable
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out); err != nil {
		return core.ErrUnavailable
	}
	return nil
}

type session struct {
	ID            string            `json:"id"`
	URL           string            `json:"url"`
	Status        string            `json:"status"`
	PaymentStatus string            `json:"payment_status"`
	Mode          string            `json:"mode"`
	Currency      string            `json:"currency"`
	Amount        int64             `json:"amount_total"`
	PaymentIntent string            `json:"payment_intent"`
	Metadata      map[string]string `json:"metadata"`
}

func (c *Client) Checkout(ctx context.Context, p core.BusinessPoint, record core.BusinessCheckout) (core.CheckoutSession, error) {
	var result session
	var err error
	if record.SessionID != "" {
		err = c.request(ctx, "GET", "checkout/sessions/"+url.PathEscape(record.SessionID), nil, "", &result)
	} else {
		if err = c.validatePrice(ctx); err != nil {
			return core.CheckoutSession{}, err
		}
		// Stripe may discard idempotency keys after 24 hours. Fail closed if a
		// process died before persisting its session; never blindly create again.
		if time.Since(record.CreatedAt) > 23*time.Hour {
			return core.CheckoutSession{}, core.ErrUnavailable
		}
		v := url.Values{"mode": {"payment"}, "line_items[0][price]": {c.Config.Price}, "line_items[0][quantity]": {"1"}, "success_url": {c.Config.SuccessURL}, "cancel_url": {c.Config.CancelURL}, "client_reference_id": {p.ID}, "payment_method_types[0]": {"card"}}
		for k, value := range map[string]string{"app": "harkai", "point_id": p.ID, "owner_id": p.OwnerID} {
			v.Set("metadata["+k+"]", value)
			v.Set("payment_intent_data[metadata]["+k+"]", value)
		}
		err = c.request(ctx, "POST", "checkout/sessions", v, "harkai-map-"+record.Key, &result)
	}
	if err != nil {
		return core.CheckoutSession{}, err
	}
	if result.Status == "complete" {
		return core.CheckoutSession{ID: result.ID, Status: "processing"}, nil
	}
	if result.Status != "expired" {
		u, e := url.Parse(result.URL)
		if e != nil || u.Scheme != "https" || u.Host != "checkout.stripe.com" {
			return core.CheckoutSession{}, core.ErrUnavailable
		}
	}
	return core.CheckoutSession{ID: result.ID, URL: result.URL, Status: result.Status}, nil
}
