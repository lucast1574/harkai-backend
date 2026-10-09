package push

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/core"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestPayloadDoesNotRevealReportContentOrLocation(t *testing.T) {
	now := time.Now().UTC()
	expiry := now.Add(35 * time.Second)
	s := Sender{endpoint: "https://fcm.example.test/send", client: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		raw, _ := io.ReadAll(r.Body)
		for _, secret := range []string{"private description", "author-id", "phone-contact", "-77.0428"} {
			if strings.Contains(string(raw), secret) {
				t.Fatalf("private information in push: %s", secret)
			}
		}
		var body struct {
			Message struct {
				Data    map[string]string `json:"data"`
				Android struct {
					TTL string `json:"ttl"`
				} `json:"android"`
			} `json:"message"`
		}
		_ = json.Unmarshal(raw, &body)
		if body.Message.Data["verified"] != "false" || body.Message.Data["report_id"] != "123456789012345678901234" || body.Message.Android.TTL != "35s" {
			t.Fatalf("bad payload: %s", raw)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"name":"accepted"}`)), Header: make(http.Header)}, nil
	})}}
	err := s.Send(context.Background(), core.Incident{ID: "123456789012345678901234", Description: "private description", ContactInfo: "phone-contact", UserID: "author-id", Longitude: -77.0428, ExpiresAt: &expiry}, core.Device{Token: "private-fcm-token"}, now)
	if err != nil {
		t.Fatal(err)
	}
}
func TestOnlyFCMDeviceErrorsRevokeTokens(t *testing.T) {
	for _, tc := range []struct {
		body string
		gone bool
	}{{`{"error":{"details":[{"@type":"type.googleapis.com/google.firebase.fcm.v1.FcmError","errorCode":"UNREGISTERED"}]}}`, true}, {`{"error":{"status":"NOT_FOUND"}}`, false}, {`{"error":{"status":"PERMISSION_DENIED"}}`, false}, {`not-json`, false}} {
		s := Sender{endpoint: "https://fcm.example.test/send", client: &http.Client{Transport: transport(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(tc.body)), Header: http.Header{"Retry-After": []string{"120"}}}, nil
		})}}
		err := s.Send(context.Background(), core.Incident{ID: "123456789012345678901234"}, core.Device{Token: "token"}, time.Now())
		if errors.Is(err, ErrDeviceGone) != tc.gone {
			t.Fatalf("wrong token revocation: %v", err)
		}
		if !tc.gone {
			var retry RetryError
			if !errors.As(err, &retry) || retry.Delay != 2*time.Minute {
				t.Fatal("retry missing")
			}
		}
	}
}

func TestConfigurationCheckCannotDeliverAMessage(t *testing.T) {
	for _, status := range []int{200, 403} {
		s := Sender{endpoint: "https://fcm.example.test/send", client: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["validate_only"] != true {
				t.Fatal("configuration check could deliver a notification")
			}
			return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader("private provider error"))}, nil
		})}}
		err := s.Validate(context.Background())
		if (err == nil) != (status == 200) || (err != nil && strings.Contains(err.Error(), "private")) {
			t.Fatal("configuration permission failure ignored or disclosed")
		}
	}
}
