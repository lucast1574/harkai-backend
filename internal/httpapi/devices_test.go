package httpapi_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/lucast1574/harkai-backend/internal/core"
)

func TestDeviceRegistrationIsPrivateAndSessionBound(t *testing.T) {
	api := fixture(t)
	a := registered(t, api, "push-one@example.test")
	b := registered(t, api, "push-two@example.test")
	input := core.DeviceInput{Token: "private-device-token-for-tests", Platform: "web", Latitude: -12.04, Longitude: -77.04}
	request(t, api, "PUT", "/v1/me/devices", "", input, 401)
	request(t, api, "PUT", "/v1/me/devices", a.AccessToken, input, 503)
	api.Service.PushEnabled = true
	w := request(t, api, "PUT", "/v1/me/devices", a.AccessToken, input, 200)
	var d core.Device
	if err := json.Unmarshal(w.Body.Bytes(), &d); err != nil || len(d.ID) != 64 {
		t.Fatal("invalid public device contract")
	}
	for _, private := range []string{input.Token, a.User.ID, "session_id", "generation", "location"} {
		if strings.Contains(w.Body.String(), private) {
			t.Fatal("private device data leaked")
		}
	}
	request(t, api, "PUT", "/v1/me/devices", b.AccessToken, input, 409)
	request(t, api, "DELETE", "/v1/me/devices/"+d.ID, b.AccessToken, nil, 204)
	request(t, api, "PUT", "/v1/me/devices", b.AccessToken, input, 409)
	request(t, api, "DELETE", "/v1/me/devices/"+d.ID, a.AccessToken, nil, 204)
	request(t, api, "PUT", "/v1/me/devices", b.AccessToken, input, 200)
	request(t, api, "POST", "/v1/auth/logout", b.AccessToken, nil, 204)
	request(t, api, "PUT", "/v1/me/devices", b.AccessToken, input, 401)
}
