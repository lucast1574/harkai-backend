package application

import (
	"context"
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
	"strings"
	"unicode/utf8"
)

func (s Service) SavePreferences(ctx context.Context, userID string, p core.Preferences) (core.User, error) {
	p, err := core.ValidatePreferences(p)
	if err != nil {
		return core.User{}, err
	}
	return s.Users.UpdatePreferences(ctx, userID, p)
}

func (s Service) SaveName(ctx context.Context, userID, name string) (core.User, error) {
	name = strings.TrimSpace(name)
	if utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 80 {
		return core.User{}, fmt.Errorf("%w: name must contain 2 to 80 characters", core.ErrInvalid)
	}
	return s.Users.UpdateName(ctx, userID, name)
}
