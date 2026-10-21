package httpapi

import (
	"compress/gzip"
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/application"
	"github.com/lucast1574/harkai-backend/internal/core"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
)

func TestGeographicEndpointAndValidation(t *testing.T) {
	api := New(application.Service{}, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	r := httptest.NewRequest("GET", "/v1/geo/district?lat=-12.083&lng=-76.992&scope=city", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	api.ServeHTTP(w, r)
	if w.Code != 200 || w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("response %d %s", w.Code, w.Body.String())
	}
	reader, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var body struct {
		District  core.GeographicDistrict
		Area      core.GeographicArea
		Districts []map[string]string
		Center    map[string]float64
	}
	if err = json.NewDecoder(reader).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.District.ID != "150130" || body.Area.Scope != "city" || len(body.Districts) != 50 || body.Center["latitude"] == 0 {
		t.Fatalf("wrong resolution: %s %s %d %+v", body.District.ID, body.Area.Scope, len(body.Districts), body.Center)
	}
	autoReq := httptest.NewRequest("GET", "/v1/geo/district?auto=1", nil)
	autoReq.Header.Set("CF-IPCity", "Trujillo")
	w = httptest.NewRecorder()
	api.ServeHTTP(w, autoReq)
	if w.Code != 200 {
		t.Fatalf("auto resolution failed: %d", w.Code)
	}
	for _, query := range []string{"lat=NaN&lng=-77", "ubigeo=bad", "ubigeo=150130&scope=bogus", "lat=-12"} {
		w = httptest.NewRecorder()
		api.ServeHTTP(w, httptest.NewRequest("GET", "/v1/geo/district?"+query, nil))
		if w.Code != 400 {
			t.Fatalf("invalid %s returned %d", query, w.Code)
		}
	}
	q, err := incidentQuery(httptest.NewRequest("GET", "/?lat=-12&lng=-77&ubigeo=150130&scope=city", nil), true)
	if err != nil || q.Boundary == nil || q.Boundary.Type != "MultiPolygon" {
		t.Fatal("query did not select geographic city boundary", err)
	}
}
