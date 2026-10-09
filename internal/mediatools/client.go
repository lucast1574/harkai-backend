package mediatools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/core"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	base, key string
	http      *http.Client
}

func New(base, key string) (*Client, error) {
	u, err := url.Parse(base)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") || len(key) < 32 {
		return nil, errors.New("invalid media worker configuration")
	}
	return &Client{base: strings.TrimRight(base, "/"), key: key, http: &http.Client{Timeout: 35 * time.Second, CheckRedirect: rejectRedirect}}, nil
}
func rejectRedirect(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }

func (c *Client) call(ctx context.Context, path string, raw []byte, dst any) error {
	req, err := http.NewRequestWithContext(ctx, "POST", c.base+path, bytes.NewReader(raw))
	if err != nil {
		return core.ErrUnavailable
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Content-Type", "application/octet-stream")
	response, err := c.http.Do(req)
	if err != nil {
		return core.ErrUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode == 400 {
		return core.ErrInvalid
	}
	if response.StatusCode != 200 {
		return core.ErrUnavailable
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 64*1024+1))
	if err != nil || len(data) > 64*1024 || json.Unmarshal(data, dst) != nil {
		return core.ErrUnavailable
	}
	return nil
}

func (c *Client) Moderate(ctx context.Context, raw []byte) (string, error) {
	var result struct {
		Approved bool   `json:"approved"`
		Policy   string `json:"policy"`
		Reason   string `json:"reason"`
	}
	if err := c.call(ctx, "/moderate", raw, &result); err != nil {
		return "", err
	}
	if result.Policy != "explicit-v1" {
		return "", core.ErrUnavailable
	}
	if !result.Approved {
		return "", core.ErrImageRejected
	}
	if result.Reason != "approved" {
		return "", core.ErrUnavailable
	}
	return result.Policy, nil
}

func (c *Client) Transcribe(ctx context.Context, raw []byte) (string, error) {
	var result struct {
		Text                 string `json:"text"`
		Source               string `json:"source"`
		RequiresConfirmation bool   `json:"requires_confirmation"`
	}
	if err := c.call(ctx, "/transcribe", raw, &result); err != nil {
		return "", err
	}
	if result.Source != "local_stt" || !result.RequiresConfirmation {
		return "", core.ErrUnavailable
	}
	return result.Text, nil
}

func (c *Client) Source(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.base+"/source", nil)
	if err != nil {
		return nil, core.ErrUnavailable
	}
	response, err := c.http.Do(req)
	if err != nil {
		return nil, core.ErrUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, core.ErrUnavailable
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 128*1024+1))
	if err != nil || len(data) > 128*1024 {
		return nil, core.ErrUnavailable
	}
	return data, nil
}
