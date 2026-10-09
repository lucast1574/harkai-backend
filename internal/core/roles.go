package core

import (
	"fmt"
	"slices"
)

type UserPage struct {
	Items      []User `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func Roles() []string { return []string{"user", "gov", "admin"} }

func CanViewInstitutional(role string) bool {
	return role == "gov" || role == "admin"
}

func ValidateRoleAssignment(actor User, targetID, role string) error {
	if actor.Role != "admin" {
		return ErrForbidden
	}
	if !slices.Contains(Roles(), role) {
		return fmt.Errorf("%w: role", ErrInvalid)
	}
	if targetID == actor.ID {
		return fmt.Errorf("%w: own role cannot be changed", ErrForbidden)
	}
	return nil
}
