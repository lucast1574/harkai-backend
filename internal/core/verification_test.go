package core

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCommunityConfirmationRules(t *testing.T) {
	now := time.Now().UTC()
	end := now.Add(time.Hour)
	base := Incident{UserID: "author", Status: "active", ExpiresAt: &end}
	if err := ValidateConfirmation(base, "other", now); err != nil {
		t.Fatal(err)
	}
	if err := ValidateConfirmation(base, "author", now); !errors.Is(err, ErrForbidden) {
		t.Fatal("author self-confirmed", err)
	}
	if err := ValidateConfirmation(base, "", now); !errors.Is(err, ErrUnauthorized) {
		t.Fatal("anonymous confirmation accepted", err)
	}
	for _, state := range []string{"verified", "hidden", "resolved", "expired"} {
		i := base
		switch state {
		case "verified":
			i.Verified = true
		case "expired":
			i.ExpiresAt = &now
		default:
			i.Status = state
		}
		if err := ValidateConfirmation(i, "other", now); !errors.Is(err, ErrConflict) {
			t.Errorf("%s accepted: %v", state, err)
		}
	}
	base.VerifiedBy = "private-approver-identity"
	data, err := json.Marshal(PublicIncident(base))
	if err != nil || strings.Contains(string(data), "private-approver") {
		t.Fatal("approver identity leaked", err)
	}
}

func TestRoleAssignmentBoundaries(t *testing.T) {
	for _, role := range Roles() {
		if err := ValidateRoleAssignment(User{ID: "a", Role: "admin"}, "b", role); err != nil {
			t.Fatal(err)
		}
	}
	for _, role := range []string{"user", "gov", "unknown", ""} {
		if err := ValidateRoleAssignment(User{ID: "a", Role: role}, "b", "admin"); !errors.Is(err, ErrForbidden) {
			t.Fatal("privilege escalation", role)
		}
	}
	if err := ValidateRoleAssignment(User{ID: "a", Role: "admin"}, "a", "user"); !errors.Is(err, ErrForbidden) {
		t.Fatal("self demotion accepted")
	}
	if err := ValidateRoleAssignment(User{ID: "a", Role: "admin"}, "b", "company"); !errors.Is(err, ErrInvalid) {
		t.Fatal("old role accepted")
	}
	if CanViewInstitutional("user") || !CanViewInstitutional("gov") || !CanViewInstitutional("admin") {
		t.Fatal("panel permissions")
	}
}
