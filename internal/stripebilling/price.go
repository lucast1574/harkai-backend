package stripebilling

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/url"
)

// The shared Stripe account also sells other products. A misconfigured price
// must fail closed before opening checkout, even if the currency matches.
func (c *Client) validatePrice(ctx context.Context) error {
	var p struct {
		ID       string `json:"id"`
		Product  string `json:"product"`
		Currency string `json:"currency"`
		Type     string `json:"type"`
		Amount   int64  `json:"unit_amount"`
		Active   bool   `json:"active"`
	}
	if err := c.request(ctx, "GET", "prices/"+url.PathEscape(c.Config.Price), nil, "", &p); err != nil {
		return err
	}
	if p.ID != c.Config.Price || p.Product != "prod_harkai_map_listing" || p.Currency != "usd" || p.Amount != 500 || p.Type != "one_time" || !p.Active {
		return core.ErrUnavailable
	}
	return nil
}
