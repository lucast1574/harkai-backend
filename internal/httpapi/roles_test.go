package httpapi_test

import (
	"context"
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"testing"
)

func TestRolePermissionsApplyImmediately(t *testing.T) {
	api := fixture(t)
	admin := registered(t, api, "admin@example.com")
	gov := registered(t, api, "gov@example.com")
	user := registered(t, api, "user@example.com")
	store := api.Service.Users.(*mongostore.Store)
	// Explicit fixture bootstrap only; there is no public bootstrap or register-role endpoint.
	if _, err := store.DB.Collection("users").UpdateOne(context.Background(), bson.M{"_id": admin.User.ID}, bson.M{"$set": bson.M{"role": "admin"}}); err != nil {
		t.Fatal(err)
	}
	area := "?lat=-12&lng=-77&radius_meters=500"
	request(t, api, "GET", "/v1/gov/incidents"+area, gov.AccessToken, nil, 403)
	request(t, api, "PATCH", "/v1/admin/users/"+gov.User.ID+"/role", user.AccessToken, map[string]string{"role": "gov"}, 403)
	request(t, api, "PATCH", "/v1/admin/users/"+gov.User.ID+"/role", admin.AccessToken, map[string]string{"role": "gov"}, 200)
	request(t, api, "GET", "/v1/gov/incidents"+area, gov.AccessToken, nil, 200)
	request(t, api, "GET", "/v1/admin/incidents"+area, gov.AccessToken, nil, 403)
	request(t, api, "GET", "/v1/admin/users", gov.AccessToken, nil, 403)
	request(t, api, "PATCH", "/v1/admin/users/"+user.User.ID+"/role", gov.AccessToken, map[string]string{"role": "admin"}, 403)
	request(t, api, "PATCH", "/v1/admin/users/"+admin.User.ID+"/role", admin.AccessToken, map[string]string{"role": "user"}, 403)
	w := request(t, api, "GET", "/v1/admin/users?limit=1", admin.AccessToken, nil, 200)
	var page core.UserPage
	if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.NextCursor == "" || strings.Contains(w.Body.String(), "password_hash") || strings.Contains(w.Body.String(), "safe-test-password") {
		t.Fatal("user pagination or credential exposure")
	}
	request(t, api, "PATCH", "/v1/admin/users/"+gov.User.ID+"/role", admin.AccessToken, map[string]string{"role": "user"}, 200)
	request(t, api, "GET", "/v1/gov/incidents"+area, gov.AccessToken, nil, 403)
}

func TestReportConfirmationAndUnverifiedAlerts(t *testing.T) {
	api := fixture(t)
	author := registered(t, api, "author@example.com")
	other := registered(t, api, "observer@example.com")
	w := request(t, api, "POST", "/v1/incidents", author.AccessToken, core.IncidentInput{Type: "fire", Description: "Incendio visible en el mercado", Latitude: -12, Longitude: -77}, 201)
	var i core.Incident
	if err := json.Unmarshal(w.Body.Bytes(), &i); err != nil {
		t.Fatal(err)
	}
	w = request(t, api, "GET", "/v1/alerts?lat=-12&lng=-77", other.AccessToken, nil, 200)
	var page core.IncidentPage
	if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].Verified {
		t.Fatal("unverified report missing from alerts")
	}
	path := "/v1/incidents/" + i.ID + "/confirm"
	request(t, api, "POST", path, "", nil, 401)
	request(t, api, "POST", path, author.AccessToken, nil, 403)
	w = request(t, api, "POST", path, other.AccessToken, nil, 200)
	if err := json.Unmarshal(w.Body.Bytes(), &i); err != nil {
		t.Fatal(err)
	}
	if !i.Verified || i.Verification != "community" || i.VerifiedBy != "" || i.VerifiedAt == nil {
		t.Fatal("community confirmation contract incorrect")
	}
	request(t, api, "POST", path, other.AccessToken, nil, 409)
}
