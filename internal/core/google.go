package core

import (
	"net/url"
	"strings"
)

type GoogleIdentity struct{ Subject, Email, Name, PhotoURL string }

func GooglePhotoURL(value string) string {
	parsed, err := url.Parse(value)
	if err != nil || len(value) > 2048 || parsed.Scheme != "https" || !strings.HasSuffix(parsed.Hostname(), ".googleusercontent.com") {
		return ""
	}
	return value
}
