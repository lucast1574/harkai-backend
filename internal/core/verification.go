package core

import (
	"fmt"
	"time"
)

func ValidateConfirmation(incident Incident, userID string, now time.Time) error {
	if userID == "" {
		return ErrUnauthorized
	}
	if incident.UserID == userID {
		return fmt.Errorf("%w: author cannot confirm own report", ErrForbidden)
	}
	if incident.Verified || incident.Status != "active" || (incident.ExpiresAt != nil && !incident.ExpiresAt.After(now)) {
		return fmt.Errorf("%w: report cannot be confirmed", ErrConflict)
	}
	return nil
}
