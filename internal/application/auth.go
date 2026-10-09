package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
	"time"
)

func (s Service) Register(ctx context.Context, input core.Credentials) (core.AuthResult, error) {
	input, err := core.ValidateRegistration(input)
	if err != nil {
		return core.AuthResult{}, err
	}
	hash, err := s.HashPassword(input.Password)
	if err != nil {
		return core.AuthResult{}, err
	}
	user, err := s.Users.CreateUser(ctx, core.User{ID: s.ID(), Email: input.Email, Name: input.Name, Role: "user", PasswordHash: hash, Preferences: core.DefaultPreferences(), CreatedAt: s.Now()})
	if err != nil {
		return core.AuthResult{}, err
	}
	return s.newSession(ctx, user)
}

func (s Service) Login(ctx context.Context, input core.Credentials) (core.AuthResult, error) {
	email, err := core.NormalizeEmail(input.Email)
	if err != nil || len(input.Password) > 72 {
		return core.AuthResult{}, core.ErrUnauthorized
	}
	user, err := s.Users.UserByEmail(ctx, email)
	if errors.Is(err, core.ErrNotFound) {
		_ = s.CheckPassword(s.DummyPasswordHash, input.Password)
		return core.AuthResult{}, core.ErrUnauthorized
	}
	if err != nil {
		return core.AuthResult{}, err
	}
	if !s.CheckPassword(user.PasswordHash, input.Password) {
		return core.AuthResult{}, core.ErrUnauthorized
	}
	return s.newSession(ctx, user)
}

func (s Service) Authenticate(ctx context.Context, token string) (core.User, error) {
	if len(token) != 43 {
		return core.User{}, core.ErrUnauthorized
	}
	session, err := s.Sessions.SessionByAccess(ctx, s.HashToken(token), s.Now())
	if err != nil {
		return core.User{}, err
	}
	user, err := s.Users.UserByID(ctx, session.UserID)
	if errors.Is(err, core.ErrNotFound) {
		return core.User{}, core.ErrUnauthorized
	}
	return user, err
}

func (s Service) Logout(ctx context.Context, token string) error {
	return s.Sessions.DeleteSession(ctx, s.HashToken(token))
}

func (s Service) newSession(ctx context.Context, user core.User) (core.AuthResult, error) {
	result, session, err := s.issueTokens(user)
	if err != nil {
		return core.AuthResult{}, err
	}
	if err = s.Sessions.CreateSession(ctx, session); err != nil {
		return core.AuthResult{}, err
	}
	return result, nil
}

func (s Service) issueTokens(user core.User) (core.AuthResult, core.Session, error) {
	access, err := s.Token()
	if err != nil {
		return core.AuthResult{}, core.Session{}, err
	}
	refresh, err := s.Token()
	if err != nil {
		return core.AuthResult{}, core.Session{}, err
	}
	now := s.Now()
	session := core.Session{ID: s.ID(), UserID: user.ID, AccessHash: s.HashToken(access), RefreshHash: s.HashToken(refresh), AccessExpiresAt: now.Add(15 * time.Minute), ExpiresAt: now.Add(30 * 24 * time.Hour)}
	return core.AuthResult{User: user, AccessToken: access, RefreshToken: refresh, TokenType: "Bearer", ExpiresIn: 900}, session, nil
}

func (s Service) Refresh(ctx context.Context, token string) (core.AuthResult, error) {
	if len(token) != 43 {
		return core.AuthResult{}, core.ErrUnauthorized
	}
	result, next, err := s.issueTokens(core.User{})
	if err != nil {
		return core.AuthResult{}, err
	}
	previous, err := s.Sessions.RotateSession(ctx, s.HashToken(token), next, s.Now())
	if err != nil {
		return core.AuthResult{}, err
	}
	user, err := s.Users.UserByID(ctx, previous.UserID)
	if err != nil {
		return core.AuthResult{}, fmt.Errorf("refresh user: %w", err)
	}
	result.User = user
	return result, nil
}
