package httpapi_test

import (
	"context"
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/core"
	"testing"
)

func TestDirectoryRequiresAdminAndDoesNotInventNumbers(t *testing.T) {
	api := fixture(t)
	user := registered(t, api, "citizen@example.com")
	request(t, api, "GET", "/v1/emergency-contacts?country=PE&city=trujillo", "", nil, 404)
	contacts := core.EmergencyContacts{Country: "PE", City: "Trujillo", Numbers: map[string]string{"firefighters": "123456"}}
	request(t, api, "PUT", "/v1/admin/emergency-contacts", user.AccessToken, contacts, 403)
	hash, err := api.Service.HashPassword("safe-admin-password-123")
	if err != nil {
		t.Fatal(err)
	}
	_, err = api.Service.Users.CreateUser(context.Background(), core.User{ID: api.Service.ID(), Email: "admin@example.com", Name: "Admin", Role: "admin", PasswordHash: hash, Preferences: core.DefaultPreferences(), CreatedAt: api.Service.Now()})
	if err != nil {
		t.Fatal(err)
	}
	login := request(t, api, "POST", "/v1/auth/login", "", core.Credentials{Email: "admin@example.com", Password: "safe-admin-password-123"}, 200)
	var auth core.AuthResult
	_ = json.Unmarshal(login.Body.Bytes(), &auth)
	request(t, api, "PUT", "/v1/admin/emergency-contacts", auth.AccessToken, contacts, 200)
	request(t, api, "GET", "/v1/emergency-contacts?country=pe&city=trujillo", "", nil, 200)
	contacts.Numbers = map[string]string{"medical": "https://malicious.example"}
	request(t, api, "PUT", "/v1/admin/emergency-contacts", auth.AccessToken, contacts, 400)
}

func TestOpenAPIIsValidAndServed(t *testing.T) {
	api := fixture(t)
	response := request(t, api, "GET", "/openapi.json", "", nil, 200)
	var spec map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &spec); err != nil {
		t.Fatal(err)
	}
	if spec["openapi"] != "3.1.0" {
		t.Fatal("unexpected specification")
	}
	paths := spec["paths"].(map[string]any)
	for _, path := range []string{"/v1/auth/register", "/v1/incidents", "/v1/media", "/v1/me/preferences", "/v1/alerts", "/v1/emergency-contacts"} {
		if paths[path] == nil {
			t.Errorf("undocumented %s", path)
		}
	}
}
