package httpapi_test

import (
	"bytes"
	"context"
	"github.com/lucast1574/harkai-backend/internal/application"
	"github.com/lucast1574/harkai-backend/internal/core"
	"image"
	"image/png"
	"net/http/httptest"
	"testing"
)

func TestRulesCannotVerifyReports(t *testing.T) {
	api := fixture(t)
	auth := registered(t, api, "rules@example.com")
	request(t, api, "POST", "/v1/analysis/text", "", map[string]string{"text": "Hay un incendio"}, 401)
	response := request(t, api, "POST", "/v1/analysis/text", auth.AccessToken, map[string]string{"text": "Hay un incendio en una casa"}, 200)
	if !bytes.Contains(response.Body.Bytes(), []byte(`"requires_confirmation":true`)) || !bytes.Contains(response.Body.Bytes(), []byte(`"suggested_type":"fire"`)) {
		t.Fatal("unexpected analysis contract")
	}
	request(t, api, "POST", "/v1/analysis/text", auth.AccessToken, map[string]string{"text": "Hay un incendio", "verified": "true"}, 400)
	request(t, api, "POST", "/v1/analysis/text", auth.AccessToken, map[string]string{"text": "x"}, 400)
}

func TestRejectedAndUnavailableImagesAreNeverStored(t *testing.T) {
	api := fixture(t)
	auth := registered(t, api, "blockedphoto@example.com")
	for _, failure := range []error{core.ErrImageRejected, core.ErrUnavailable} {
		api.Service.ModerateImage = func(_ context.Context, _ []byte) (string, error) { return "", failure }
		stored := false
		original := api.Service.Media
		photos := api.Service.Photos
		api.Service.Media = rejectingStore{MediaStore: original, stored: &stored}
		api.Service.Photos = rejectingPhotos{PhotoStore: photos, stored: &stored}
		var payload bytes.Buffer
		_ = png.Encode(&payload, image.NewRGBA(image.Rect(0, 0, 32, 32)))
		r := httptest.NewRequest("POST", "/v1/media", &payload)
		r.Header.Set("Content-Type", "image/png")
		r.Header.Set("Authorization", "Bearer "+auth.AccessToken)
		w := httptest.NewRecorder()
		api.ServeHTTP(w, r)
		if (failure == core.ErrImageRejected && w.Code != 400) || (failure == core.ErrUnavailable && w.Code != 503) || stored {
			t.Fatal("failed moderation reached storage")
		}
		api.Service.Media = original
		api.Service.Photos = photos
	}
}

type rejectingStore struct {
	application.MediaStore
	stored *bool
}

func (s rejectingStore) CreateMedia(ctx context.Context, media core.Media) (core.Media, error) {
	*s.stored = true
	return s.MediaStore.CreateMedia(ctx, media)
}

type rejectingPhotos struct {
	application.PhotoStore
	stored *bool
}

func (s rejectingPhotos) Put(ctx context.Context, id string, data []byte) error {
	*s.stored = true
	return s.PhotoStore.Put(ctx, id, data)
}
