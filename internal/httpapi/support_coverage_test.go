package httpapi

import (
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/application"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
)

func TestSupportCoverageIsSeparateFromEmptyReports(t *testing.T) {
	api := New(application.Service{}, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, test := range []struct {
		query    string
		covered  bool
		province string
	}{
		{"lat=-12.083&lng=-76.992", true, "1501"},
		{"lat=-12.061&lng=-77.139", true, "0701"},
		{"lat=-8.1116&lng=-79.0287", true, "1301"},
		{"lat=-16.398&lng=-71.536", false, "0401"},
	} {
		w := httptest.NewRecorder()
		api.ServeHTTP(w, httptest.NewRequest("GET", "/v1/support/coverage?"+test.query, nil))
		var body struct {
			Covered    bool
			ProvinceID string `json:"province_id"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if w.Code != 200 || body.Covered != test.covered || body.ProvinceID != test.province {
			t.Fatalf("coverage %s: %d %+v", test.query, w.Code, body)
		}
	}
	for _, query := range []string{"lat=NaN&lng=-77", "lat=-12", "lat=95&lng=-77"} {
		w := httptest.NewRecorder()
		api.ServeHTTP(w, httptest.NewRequest("GET", "/v1/support/coverage?"+query, nil))
		if w.Code != 400 {
			t.Fatalf("invalid coordinates %s: %d", query, w.Code)
		}
	}
}
