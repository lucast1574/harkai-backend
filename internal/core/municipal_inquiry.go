package core

import (
	"strings"
	"unicode/utf8"
)

type MunicipalInquiry struct {
	RequestID    string `json:"requestId"`
	Municipality string `json:"municipality"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Message      string `json:"message"`
	Consent      bool   `json:"consent"`
}

func ValidateMunicipalInquiry(in MunicipalInquiry) (MunicipalInquiry, error) {
	fields := []*string{&in.Municipality, &in.Name, &in.Role, &in.Phone, &in.Message}
	limits := [][2]int{{3, 120}, {2, 100}, {2, 100}, {0, 40}, {10, 3000}}
	for i, field := range fields {
		*field = strings.TrimSpace(*field)
		count := utf8.RuneCountInString(*field)
		if !utf8.ValidString(*field) || count < limits[i][0] || count > limits[i][1] || strings.ContainsRune(*field, '\x00') {
			return in, ErrInvalid
		}
		if i < 4 && strings.ContainsAny(*field, "\r\n") {
			return in, ErrInvalid
		}
	}
	email, err := NormalizeEmail(in.Email)
	if err != nil || !in.Consent {
		return in, ErrInvalid
	}
	in.Email = email
	return in, nil
}
