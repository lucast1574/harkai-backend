package httpapi_test

import (
	"bytes"
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/httpapi"
	"github.com/lucast1574/harkai-backend/internal/runtime"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fixture(t *testing.T) *httpapi.API {
	t.Helper()
	store := testutil.Store(t)
	service, err := runtime.NewService(store, testutil.Photos(t))
	if err != nil {
		t.Fatal(err)
	}
	return httpapi.New(service, store.Ping, []string{"https://harkai.lat"}, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func request(t *testing.T, api http.Handler, method, path, token string, input any, status int) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	if input != nil {
		if err := json.NewEncoder(&body).Encode(input); err != nil {
			t.Fatal(err)
		}
	}
	r := httptest.NewRequest(method, path, &body)
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	api.ServeHTTP(w, r)
	if w.Code != status {
		t.Fatalf("%s %s: expected %d got %d: %s", method, path, status, w.Code, w.Body.String())
	}
	return w
}

func registered(t *testing.T, api http.Handler, email string) core.AuthResult {
	t.Helper()
	w := request(t, api, "POST", "/v1/auth/register", "", core.Credentials{Email: email, Name: "Test User", Password: "safe-test-password-123"}, 201)
	var result core.AuthResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.User.Role != "user" || result.User.PasswordHash != "" {
		t.Fatal("registration privilege or credential leak")
	}
	return result
}

func TestAuthLifecycleAndContract(t *testing.T) {
	api := fixture(t)
	auth := registered(t, api, "one@example.com")
	request(t, api, "GET", "/v1/me", auth.AccessToken, nil, 200)
	request(t, api, "POST", "/v1/auth/login", "", core.Credentials{Email: "one@example.com", Password: "wrong"}, 401)
	request(t, api, "POST", "/v1/auth/register", "", map[string]any{"email": "two@example.com", "name": "User", "password": "safe-test-password-123", "role": "admin"}, 400)
	refreshed := request(t, api, "POST", "/v1/auth/refresh", "", map[string]string{"refresh_token": auth.RefreshToken}, 200)
	var next core.AuthResult
	_ = json.Unmarshal(refreshed.Body.Bytes(), &next)
	request(t, api, "GET", "/v1/me", auth.AccessToken, nil, 401)
	request(t, api, "POST", "/v1/auth/refresh", "", map[string]string{"refresh_token": auth.RefreshToken}, 401)
	request(t, api, "POST", "/v1/auth/logout", next.AccessToken, nil, 204)
	request(t, api, "GET", "/v1/me", next.AccessToken, nil, 401)
	request(t, api, "POST", "/v1/auth/refresh", "", map[string]string{"refresh_token": next.RefreshToken}, 401)
}

func TestReportGeoOwnershipAndPreferences(t *testing.T) {
	api := fixture(t)
	owner := registered(t, api, "owner@example.com")
	other := registered(t, api, "other@example.com")
	request(t, api, "POST", "/v1/incidents", "", nil, 401)
	request(t, api, "POST", "/v1/incidents", owner.AccessToken, map[string]any{"type": "fire", "description": "Visible fire"}, 400)
	report := request(t, api, "POST", "/v1/incidents", owner.AccessToken, core.IncidentInput{Type: "fire", Description: "Visible fire reported", Latitude: -8.1, Longitude: -79.0}, 201)
	var incident core.Incident
	_ = json.Unmarshal(report.Body.Bytes(), &incident)
	request(t, api, "POST", "/v1/incidents", owner.AccessToken, core.IncidentInput{Type: "fire", Description: "Duplicate report", Latitude: -8.1, Longitude: -79.0}, 429)
	w := request(t, api, "GET", "/v1/incidents?lat=-8.1&lng=-79&radius_meters=500", "", nil, 200)
	var page core.IncidentPage
	_ = json.Unmarshal(w.Body.Bytes(), &page)
	if len(page.Items) != 1 || page.Items[0].ID != incident.ID {
		t.Fatal("nearby report missing")
	}
	w = request(t, api, "GET", "/v1/incidents?lat=40&lng=20&radius_meters=500", "", nil, 200)
	_ = json.Unmarshal(w.Body.Bytes(), &page)
	if len(page.Items) != 0 {
		t.Fatal("distant incident leaked into geo query")
	}
	request(t, api, "PATCH", "/v1/incidents/"+incident.ID+"/status", other.AccessToken, map[string]string{"status": "hidden"}, 404)
	request(t, api, "GET", "/v1/admin/incidents?lat=-8.1&lng=-79", other.AccessToken, nil, 403)
	p := core.DefaultPreferences()
	p.Notifications = "none"
	request(t, api, "PUT", "/v1/me/preferences", other.AccessToken, p, 200)
	w = request(t, api, "GET", "/v1/alerts?lat=-8.1&lng=-79", other.AccessToken, nil, 200)
	_ = json.Unmarshal(w.Body.Bytes(), &page)
	if len(page.Items) != 0 {
		t.Fatal("none notifications ignored")
	}
	request(t, api, "PATCH", "/v1/incidents/"+incident.ID+"/status", owner.AccessToken, map[string]string{"status": "hidden"}, 200)
	request(t, api, "GET", "/v1/incidents/"+incident.ID, "", nil, 404)
	request(t, api, "GET", "/v1/incidents/"+incident.ID, owner.AccessToken, nil, 200)
}

func TestStrictJSONAndCORS(t *testing.T) {
	api := fixture(t)
	for _, body := range []string{`{"email":"a@example.com","name":"Test","password":"safe-password-123"} {}`, `{"unknown":true}`} {
		r := httptest.NewRequest("POST", "/v1/auth/register", bytes.NewBufferString(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		api.ServeHTTP(w, r)
		if w.Code != 400 {
			t.Fatalf("JSON accepted: %d %s", w.Code, w.Body.String())
		}
	}
	r := httptest.NewRequest("GET", "/v1/meta", nil)
	r.Header.Set("Origin", "https://evil.example")
	w := httptest.NewRecorder()
	api.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("untrusted origin accepted")
	}
	r = httptest.NewRequest("OPTIONS", "/v1/me", nil)
	r.Header.Set("Origin", "https://harkai.lat")
	w = httptest.NewRecorder()
	api.ServeHTTP(w, r)
	if w.Code != 204 || w.Header().Get("Access-Control-Allow-Origin") != "https://harkai.lat" {
		t.Fatal("valid preflight failed")
	}
}
