package httpapi_test

import (
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/core"
	"strings"
	"testing"
)

func TestCommentsOwnershipAndArchivedReport(t *testing.T) {
	api := fixture(t)
	a := registered(t, api, "writer@example.test")
	b := registered(t, api, "reader@example.test")
	w := request(t, api, "POST", "/v1/incidents", a.AccessToken, core.IncidentInput{Type: "fire", Description: "Fuego observado en la zona", Latitude: -12.04, Longitude: -77.04, District: "Lima"}, 201)
	var report core.Incident
	_ = json.Unmarshal(w.Body.Bytes(), &report)
	path := "/v1/incidents/" + report.ID + "/comments"
	request(t, api, "POST", path, "", map[string]string{"body": "Lo observé también"}, 401)
	request(t, api, "POST", path, a.AccessToken, map[string]string{"body": "x"}, 400)
	w = request(t, api, "POST", path, a.AccessToken, map[string]string{"body": "Las personas ya se alejaron"}, 201)
	var comment core.Comment
	_ = json.Unmarshal(w.Body.Bytes(), &comment)
	if !comment.CanDelete || strings.Contains(w.Body.String(), a.User.ID) || strings.Contains(w.Body.String(), a.User.Email) {
		t.Fatal("ownership or privacy")
	}
	request(t, api, "DELETE", path+"/"+comment.ID, b.AccessToken, nil, 404)
	request(t, api, "PATCH", "/v1/incidents/"+report.ID+"/status", a.AccessToken, map[string]string{"status": "resolved"}, 200)
	request(t, api, "POST", path, b.AccessToken, map[string]string{"body": "Gracias por actualizar"}, 201)
	w = request(t, api, "GET", "/v1/archive/incidents?lat=-12.04&lng=-77.04&district=Lima", "", nil, 200)
	if !strings.Contains(w.Body.String(), report.ID) {
		t.Fatal("resolved missing from archive")
	}
	w = request(t, api, "GET", "/v1/incidents?lat=-12.04&lng=-77.04", "", nil, 200)
	if strings.Contains(w.Body.String(), report.ID) {
		t.Fatal("resolved in active map")
	}
	request(t, api, "DELETE", path+"/"+comment.ID, a.AccessToken, nil, 204)
	w = request(t, api, "GET", path, "", nil, 200)
	if strings.Contains(w.Body.String(), "Las personas ya se alejaron") {
		t.Fatal("deleted text exposed")
	}
	request(t, api, "PATCH", "/v1/incidents/"+report.ID+"/status", a.AccessToken, map[string]string{"status": "hidden"}, 200)
	request(t, api, "GET", path, "", nil, 404)
	w = request(t, api, "GET", "/v1/archive/incidents?lat=-12.04&lng=-77.04", "", nil, 200)
	if strings.Contains(w.Body.String(), report.ID) {
		t.Fatal("hidden in archive")
	}
	w = request(t, api, "GET", "/v1/archive/districts?lat=-12.04&lng=-77.04", "", nil, 200)
	if strings.Contains(w.Body.String(), "Lima") {
		t.Fatal("hidden district exposed")
	}
}
