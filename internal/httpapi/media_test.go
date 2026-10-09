package httpapi_test

import (
	"bytes"
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/core"
	"image"
	"image/png"
	"net/http/httptest"
	"testing"
)

func TestMediaPublicationAndOwnership(t *testing.T) {
	api := fixture(t)
	owner := registered(t, api, "photo@example.com")
	other := registered(t, api, "viewer@example.com")
	var imageData bytes.Buffer
	_ = png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	r := httptest.NewRequest("POST", "/v1/media", &imageData)
	r.Header.Set("Content-Type", "image/png")
	r.Header.Set("Authorization", "Bearer "+owner.AccessToken)
	w := httptest.NewRecorder()
	api.ServeHTTP(w, r)
	if w.Code != 201 {
		t.Fatalf("upload: %d %s", w.Code, w.Body.String())
	}
	var media core.Media
	_ = json.Unmarshal(w.Body.Bytes(), &media)
	request(t, api, "GET", "/v1/media/"+media.ID, "", nil, 404)
	request(t, api, "GET", "/v1/media/"+media.ID, owner.AccessToken, nil, 200)
	input := core.IncidentInput{Type: "pet", Description: "Lost dog reported", Latitude: -8.1, Longitude: -79, MediaID: media.ID, ContactInfo: "123456789", ShareContact: false}
	request(t, api, "POST", "/v1/incidents", other.AccessToken, input, 403)
	report := request(t, api, "POST", "/v1/incidents", owner.AccessToken, input, 201)
	var i core.Incident
	_ = json.Unmarshal(report.Body.Bytes(), &i)
	if i.ContactInfo != "" {
		t.Fatal("contact published without consent")
	}
	request(t, api, "GET", "/v1/media/"+media.ID, "", nil, 200)
	request(t, api, "PATCH", "/v1/incidents/"+i.ID+"/status", owner.AccessToken, map[string]string{"status": "hidden"}, 200)
	request(t, api, "GET", "/v1/media/"+media.ID, "", nil, 404)
}
