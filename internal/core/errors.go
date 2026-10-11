package core

import "errors"

var (
	ErrContentRejected = errors.New("community language rejected")
	ErrInvalid         = errors.New("invalid input")
	ErrImageRejected   = errors.New("image rejected")
	ErrUnauthorized    = errors.New("authentication required")
	ErrForbidden       = errors.New("access denied")
	ErrNotFound        = errors.New("resource not found")
	ErrConflict        = errors.New("resource conflict")
	ErrRateLimited     = errors.New("too many requests")
	ErrUnavailable     = errors.New("service unavailable")
)
