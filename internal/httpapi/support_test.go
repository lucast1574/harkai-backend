package httpapi_test

import (
	"context"
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/httpapi"
	"github.com/lucast1574/harkai-backend/internal/runtime"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	"log/slog"
	"testing"
)

func TestPublicSupportRoutes(t *testing.T) {
	store := testutil.Store(t)
	if err := store.SeedSupport(context.Background()); err != nil {
		t.Fatal(err)
	}
	service, err := runtime.NewService(store, nil)
	if err != nil {
		t.Fatal(err)
	}
	api := httpapi.New(service, store.Ping, []string{"https://panel.harkai.lat"}, slog.Default())
	for city, want := range map[string]int{"lima": 50, "trujillo": 12} {
		w := request(t, api, "GET", "/v1/support/directory?city="+city, "", nil, 200)
		var result struct {
			Items []core.SupportDistrict `json:"items"`
		}
		if err = json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if len(result.Items) != want {
			t.Fatal("wrong district count")
		}
	}
	request(t, api, "GET", "/v1/support/directory?city=unknown", "", nil, 400)
	request(t, api, "GET", "/v1/support/places?lat=-8.1116&lng=-79.0287&radius_meters=50000", "", nil, 200)
	for _, q := range []string{"", "?lat=NaN&lng=0", "?lat=100&lng=0", "?lat=0&lng=0&radius_meters=100001"} {
		request(t, api, "GET", "/v1/support/places"+q, "", nil, 400)
	}
	w := request(t, api, "GET", "/v1/support/places?lat=40&lng=20", "", nil, 200)
	var result struct {
		Items []core.SupportPlace `json:"items"`
	}
	if err = json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.Items == nil || len(result.Items) != 0 {
		t.Fatal("empty result contract")
	}
}
