package core

import (
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

type Preferences struct {
	BusinessNotifications bool   `json:"business_notifications" bson:"business_notifications"`
	Notifications         string `json:"notifications" bson:"notifications"`
	RadiusMeters          int    `json:"radius_meters" bson:"radius_meters"`
	Language              string `json:"language" bson:"language"`
	Theme                 string `json:"theme" bson:"theme"`
}

type User struct {
	PhotoURL      string      `json:"photo_url,omitempty" bson:"photo_url,omitempty"`
	GoogleSubject string      `json:"-" bson:"google_subject,omitempty"`
	ID            string      `json:"id" bson:"_id"`
	Email         string      `json:"email" bson:"email"`
	Name          string      `json:"name" bson:"name"`
	Role          string      `json:"role" bson:"role"`
	PasswordHash  string      `json:"-" bson:"password_hash"`
	Preferences   Preferences `json:"preferences" bson:"preferences"`
	CreatedAt     time.Time   `json:"created_at" bson:"created_at"`
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
}

func NormalizeEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Address != value || len(value) > 254 {
		return "", fmt.Errorf("%w: email", ErrInvalid)
	}
	return value, nil
}

func ValidateRegistration(input Credentials) (Credentials, error) {
	email, err := NormalizeEmail(input.Email)
	if err != nil {
		return Credentials{}, err
	}
	input.Email = email
	input.Name = strings.TrimSpace(input.Name)
	if utf8.RuneCountInString(input.Name) < 2 || utf8.RuneCountInString(input.Name) > 80 {
		return Credentials{}, fmt.Errorf("%w: name must contain 2 to 80 characters", ErrInvalid)
	}
	if len(input.Password) < 12 || len(input.Password) > 72 {
		return Credentials{}, fmt.Errorf("%w: password must contain 12 to 72 bytes", ErrInvalid)
	}
	return input, nil
}

func DefaultPreferences() Preferences {
	return Preferences{Notifications: "all", RadiusMeters: 500, Language: "es", Theme: "system"}
}

func ValidatePreferences(p Preferences) (Preferences, error) {
	if p.Notifications != "all" && p.Notifications != "critical" && p.Notifications != "none" {
		return Preferences{}, fmt.Errorf("%w: notifications", ErrInvalid)
	}
	if p.RadiusMeters < 100 || p.RadiusMeters > 10000 {
		return Preferences{}, fmt.Errorf("%w: radius_meters must be 100 to 10000", ErrInvalid)
	}
	if p.Language != "es" && p.Language != "en" {
		return Preferences{}, fmt.Errorf("%w: language", ErrInvalid)
	}
	if p.Theme != "system" && p.Theme != "light" && p.Theme != "dark" {
		return Preferences{}, fmt.Errorf("%w: theme", ErrInvalid)
	}
	return p, nil
}
