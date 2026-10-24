package core

import (
	"net/url"
	"strings"
	"time"
)

type GoogleIdentity struct {
	Subject, Email, Name, PhotoURL string
	AuthAt                         time.Time
}

func GooglePhotoURL(value string) string {
	parsed, err := url.Parse(value)
	if err != nil || len(value) > 2048 || parsed.Scheme != "https" || !strings.HasSuffix(parsed.Hostname(), ".googleusercontent.com") {
		return ""
	}
	return value
}
