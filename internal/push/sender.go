package push

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io"
	"net/http"
	"strconv"
	"time"
)

var ErrDeviceGone = errors.New("notification device unavailable")

type RetryError struct{ Delay time.Duration }

func (e RetryError) Error() string { return "notification delivery must be retried" }

type Sender struct {
	client   *http.Client
	endpoint string
}

func New(ctx context.Context, project string, raw []byte) (*Sender, error) {
	if !validCredentials(project, raw) {
		return nil, errors.New("invalid FCM credential configuration")
	}
	credentials, err := google.JWTConfigFromJSON(raw, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return nil, errors.New("invalid FCM credential configuration")
	}
	transport := &http.Client{Timeout: 10 * time.Second, CheckRedirect: rejectRedirect}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, transport)
	client := oauth2.NewClient(ctx, credentials.TokenSource(ctx))
	client.Timeout = 10 * time.Second
	client.CheckRedirect = rejectRedirect
	return &Sender{client: client, endpoint: "https://fcm.googleapis.com/v1/projects/" + project + "/messages:send"}, nil
}
func rejectRedirect(_ *http.Request, _ []*http.Request) error {
	return errors.New("notification redirect rejected")
}
func (s *Sender) Send(ctx context.Context, i core.Incident, d core.Device, now time.Time) error {
	if !core.ValidID(i.ID) || d.Token == "" {
		return core.ErrInvalid
	}
	ttl := 3600
	if i.ExpiresAt != nil {
		ttl = min(ttl, int(i.ExpiresAt.Sub(now).Seconds()))
	}
	if ttl < 1 {
		return nil
	}
	data := map[string]string{"report_id": i.ID, "verified": strconv.FormatBool(i.Verified)}
	message := map[string]any{"token": d.Token, "data": data, "notification": map[string]string{"title": "Un reporte cerca de tu zona", "body": "Información de la comunidad. Revisa su estado y recomendaciones."}, "android": map[string]any{"ttl": fmt.Sprintf("%ds", ttl), "notification": map[string]string{"tag": i.ID, "channel_id": "harkai_reports"}}, "webpush": map[string]any{"headers": map[string]string{"TTL": strconv.Itoa(ttl)}, "notification": map[string]any{"tag": i.ID, "renotify": false}, "fcm_options": map[string]string{"link": "https://panel.harkai.lat/incidents/" + i.ID}}}
	body, err := json.Marshal(map[string]any{"message": message})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := s.client.Do(req)
	if err != nil {
		return RetryError{Delay: time.Minute}
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 8193))
	if err != nil || len(raw) > 8192 {
		return RetryError{Delay: time.Minute}
	}
	if response.StatusCode == 200 {
		return nil
	}
	var result struct {
		Error struct {
			Details []struct {
				Type string `json:"@type"`
				Code string `json:"errorCode"`
			} `json:"details"`
		} `json:"error"`
	}
	if json.Unmarshal(raw, &result) == nil {
		for _, detail := range result.Error.Details {
			if detail.Type == "type.googleapis.com/google.firebase.fcm.v1.FcmError" && (detail.Code == "UNREGISTERED" || detail.Code == "SENDER_ID_MISMATCH") {
				return ErrDeviceGone
			}
		}
	}
	delay := time.Minute
	if n, err := strconv.Atoi(response.Header.Get("Retry-After")); err == nil && n > 0 {
		delay = time.Duration(min(n, 3600)) * time.Second
	}
	return RetryError{Delay: delay}
}
