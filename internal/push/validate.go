package push

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
)

// Validate checks the project's send permission without delivering a message.
func (s *Sender) Validate(ctx context.Context) error {
	body := []byte(`{"validate_only":true,"message":{"topic":"harkai-configuration-check","data":{"check":"startup"}}}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return errors.New("FCM configuration validation failed")
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := s.client.Do(req)
	if err != nil {
		return errors.New("FCM configuration validation failed")
	}
	defer response.Body.Close()
	_, err = io.Copy(io.Discard, io.LimitReader(response.Body, 8192))
	if err != nil || response.StatusCode != http.StatusOK {
		return errors.New("FCM configuration validation failed")
	}
	return nil
}
