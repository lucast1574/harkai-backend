package application

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"strings"
	"time"
)

type AccountDeletionStore interface {
	RequestAccountDeletion(context.Context, string, time.Time) error
	ClaimAccountDeletion(context.Context, time.Time) (string, error)
	AccountDeletionMedia(context.Context, string) ([]string, error)
	PurgeAccountRecords(context.Context, string) error
	ForgetDeletionMedia(context.Context, string, string) error
}
type PhotoEraser interface {
	Delete(context.Context, string) error
}

func (s Service) DeleteMyAccount(ctx context.Context, u core.User, confirmation, password, googleToken string) error {
	if len(password) > 72 || len(googleToken) > 16384 || len(confirmation) > 254 {
		return core.ErrInvalid
	}
	if s.AccountDeletion == nil {
		return core.ErrUnavailable
	}
	if !strings.EqualFold(strings.TrimSpace(confirmation), u.Email) {
		return core.ErrInvalid
	}
	if err := s.Limits.Allow(ctx, "delete-account:"+u.ID, 5, time.Hour, s.Now()); err != nil {
		return err
	}
	authorized := false
	if password != "" && u.PasswordHash != "" {
		authorized = s.CheckPassword(u.PasswordHash, password)
	}
	if googleToken != "" && u.GoogleSubject != "" && s.VerifyGoogle != nil {
		identity, err := s.VerifyGoogle(ctx, googleToken)
		if err != nil {
			return err
		}
		authorized = identity.Subject == u.GoogleSubject && !identity.AuthAt.Before(s.Now().Add(-5*time.Minute)) && !identity.AuthAt.After(s.Now().Add(time.Minute))
	}
	if !authorized {
		return core.ErrForbidden
	}
	return s.AccountDeletion.RequestAccountDeletion(ctx, u.ID, s.Now())
}
func (s Service) ProcessAccountDeletion(ctx context.Context, photos PhotoEraser) error {
	id, err := s.AccountDeletion.ClaimAccountDeletion(ctx, s.Now())
	if err != nil {
		return err
	}
	ids, err := s.AccountDeletion.AccountDeletionMedia(ctx, id)
	if err != nil {
		return err
	}
	for _, mediaID := range ids {
		if err = photos.Delete(ctx, mediaID); err != nil {
			return err
		}
		if err = s.AccountDeletion.ForgetDeletionMedia(ctx, id, mediaID); err != nil {
			return err
		}
	}
	return s.AccountDeletion.PurgeAccountRecords(ctx, id)
}
