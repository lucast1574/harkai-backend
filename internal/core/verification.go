package core

import (
	"fmt"
	"slices"
	"time"
)

func ValidateConfirmation(incident Incident, userID string, now time.Time) error {
	if userID == "" {
		return ErrUnauthorized
	}
	if incident.UserID == userID {
		return fmt.Errorf("%w: author cannot confirm own report", ErrForbidden)
	}
	if (incident.VerifiedBy == userID || slices.Contains(incident.ConfirmedBy, userID)) || !MapIncidentActive(incident, now) {
		return fmt.Errorf("%w: report cannot be confirmed", ErrConflict)
	}
	return nil
}
