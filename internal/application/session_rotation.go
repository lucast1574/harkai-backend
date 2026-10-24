package application

import (
	"context"
	"encoding/base64"
	"github.com/lucast1574/harkai-backend/internal/core"
	"time"
)

// RefreshWithRotation keeps legacy clients compatible and makes a persisted
// client rotation ID safe to retry after the process or connection is lost.
func (s Service) RefreshWithRotation(ctx context.Context, token, rotationID string) (core.AuthResult, error) {
	if rotationID == "" {
		return s.Refresh(ctx, token)
	}
	nonce, err := base64.RawURLEncoding.DecodeString(rotationID)
	if err != nil || len(nonce) != 32 || len(rotationID) != 43 {
		return core.AuthResult{}, core.ErrInvalid
	}
	if len(token) != 43 {
		return core.AuthResult{}, core.ErrUnauthorized
	}
	if s.DeriveSessionTokens == nil {
		return core.AuthResult{}, core.ErrUnavailable
	}
	access, refresh := s.DeriveSessionTokens(token, rotationID)
	now := s.Now()
	next := core.Session{AccessHash: s.HashToken(access), RefreshHash: s.HashToken(refresh), AccessExpiresAt: now.Add(15 * time.Minute), RotationID: rotationID}
	previous, err := s.Sessions.RotateSession(ctx, s.HashToken(token), next, now)
	if err != nil {
		return core.AuthResult{}, err
	}
	user, err := s.Users.UserByID(ctx, previous.UserID)
	if err != nil {
		return core.AuthResult{}, err
	}
	return core.AuthResult{User: user, AccessToken: access, RefreshToken: refresh, TokenType: "Bearer", ExpiresIn: 900}, nil
}
