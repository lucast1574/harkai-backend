package application

import (
	"context"
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
	"strings"
)

func (s Service) ManageRole(ctx context.Context, actor core.User, targetID, role string) (core.User, error) {
	if err := core.ValidateRoleAssignment(actor, targetID, role); err != nil {
		return core.User{}, err
	}
	return s.Users.AssignRole(ctx, targetID, actor.ID, role, s.Now())
}

func (s Service) AdminUsers(ctx context.Context, actor core.User, cursor string, limit int) (core.UserPage, error) {
	if actor.Role != "admin" {
		return core.UserPage{}, core.ErrForbidden
	}
	if limit < 1 || limit > 100 || (cursor != "" && (len(cursor) != 24 || strings.Trim(cursor, "0123456789abcdef") != "")) {
		return core.UserPage{}, fmt.Errorf("%w: cursor or limit", core.ErrInvalid)
	}
	return s.Users.ListUsers(ctx, cursor, limit)
}
