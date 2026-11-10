package stripebilling

import (
	"context"
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCheckoutUsesOnlyServerPriceAndReusesExistingSession(t *testing.T) {
	creates, gets := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != "sk_test_key" || p != "" {
			t.Error("missing private auth")
		}
		switch {
		case strings.HasPrefix(r.URL.Path, "/prices/"):
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "price_fixed", "product": "prod_harkai_map_listing", "unit_amount": 500, "currency": "usd", "type": "one_time", "active": true})
		case r.Method == "POST":
			creates++
			_ = r.ParseForm()
			if r.Form.Get("line_items[0][price]") != "price_fixed" || r.Form.Get("mode") != "payment" || r.Form.Get("metadata[owner_id]") != "owner" || r.Header.Get("Idempotency-Key") != "harkai-map-attempt" {
				t.Error("unbound checkout request")
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "cs_saved", "status": "open", "url": "https://checkout.stripe.com/c/pay/test"})
		default:
			gets++
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "cs_saved", "status": "complete"})
		}
	}))
	defer server.Close()
	c := New(Config{Secret: "sk_test_key", Price: "price_fixed"})
	c.BaseURL = server.URL + "/"
	p := core.BusinessPoint{ID: "507f1f77bcf86cd799439011", OwnerID: "owner"}
	r := core.BusinessCheckout{Key: "attempt", CreatedAt: time.Now()}
	result, err := c.Checkout(context.Background(), p, r)
	if err != nil || result.ID != "cs_saved" || creates != 1 {
		t.Fatalf("%+v %v", result, err)
	}
	r.SessionID = result.ID
	result, err = c.Checkout(context.Background(), p, r)
	if err != nil || result.Status != "processing" || result.URL != "" || creates != 1 || gets != 1 {
		t.Fatalf("duplicate charge risk: %+v %v", result, err)
	}
}
func TestMismatchedPriceCannotCreateCheckout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatal("created checkout with a foreign product")
		}
		_, _ = w.Write([]byte(`{"id":"foreign","product":"prod_nexode","unit_amount":500,"currency":"usd","type":"one_time","active":true}`))
	}))
	defer server.Close()
	c := New(Config{Price: "foreign"})
	c.BaseURL = server.URL + "/"
	if _, err := c.Checkout(context.Background(), core.BusinessPoint{}, core.BusinessCheckout{Key: "attempt", CreatedAt: time.Now()}); err == nil {
		t.Fatal("foreign price accepted")
	}
}
