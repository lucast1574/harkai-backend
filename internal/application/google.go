package application

import (
	"context"
	"errors"
	"github.com/lucast1574/harkai-backend/internal/core"
	"strings"
	"unicode/utf8"
)

func (s Service) GoogleLogin(ctx context.Context, token string) (core.AuthResult, error) {
	if s.VerifyGoogle == nil {
		return core.AuthResult{}, core.ErrUnavailable
	}
	identity, err := s.VerifyGoogle(ctx, token)
	if err != nil {
		return core.AuthResult{}, err
	}
	user, err := s.Users.UserByGoogle(ctx, identity.Subject)
	if err == nil {
		return s.newSession(ctx, user)
	}
	if !errors.Is(err, core.ErrNotFound) {
		return core.AuthResult{}, err
	}
	email, err := core.NormalizeEmail(identity.Email)
	if err != nil {
		return core.AuthResult{}, core.ErrUnauthorized
	}
	name := strings.TrimSpace(identity.Name)
	if utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 80 {
		name = "Usuario de Google"
	}
	// Do not silently link a password account by email: prove ownership separately.
	user, err = s.Users.CreateUser(ctx, core.User{ID: s.ID(), GoogleSubject: identity.Subject, Email: email, Name: name, Role: "user", Preferences: core.DefaultPreferences(), CreatedAt: s.Now()})
	if errors.Is(err, core.ErrConflict) {
		user, err = s.Users.UserByGoogle(ctx, identity.Subject)
		if errors.Is(err, core.ErrNotFound) {
			return core.AuthResult{}, core.ErrConflict
		}
	}
	if err != nil {
		return core.AuthResult{}, err
	}
	return s.newSession(ctx, user)
}
