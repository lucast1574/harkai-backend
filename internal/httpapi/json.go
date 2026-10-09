package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
	"io"
	"mime"
	"net/http"
)

func decode(r *http.Request, dst any) error {
	kind, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || kind != "application/json" {
		return fmt.Errorf("%w: Content-Type must be application/json", core.ErrInvalid)
	}
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err = d.Decode(dst); err != nil {
		return fmt.Errorf("%w: malformed or unknown JSON fields", core.ErrInvalid)
	}
	if err = d.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("%w: one JSON object required", core.ErrInvalid)
	}
	return nil
}

func writeResult(w http.ResponseWriter, result Result) error {
	for key, values := range result.Headers {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}
	if result.Status == 0 {
		result.Status = 200
	}
	if result.Bytes != nil {
		w.WriteHeader(result.Status)
		_, err := w.Write(result.Bytes)
		return err
	}
	if result.Status == 204 {
		w.WriteHeader(204)
		return nil
	}
	data, err := json.Marshal(result.Body)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(result.Status)
	_, err = w.Write(append(data, '\n'))
	return err
}

func writeError(w http.ResponseWriter, err error) error {
	status, code, message := 500, "internal_error", "An internal error occurred"
	switch {
	case errors.Is(err, core.ErrImageRejected):
		status, code, message = 400, "image_rejected", "Image did not pass moderation"
	case errors.Is(err, core.ErrInvalid):
		status, code, message = 400, "invalid_input", err.Error()
	case errors.Is(err, core.ErrUnauthorized):
		status, code, message = 401, "unauthorized", "Authentication required"
	case errors.Is(err, core.ErrForbidden):
		status, code, message = 403, "forbidden", "Access denied"
	case errors.Is(err, core.ErrNotFound):
		status, code, message = 404, "not_found", "Resource not found"
	case errors.Is(err, core.ErrConflict):
		status, code, message = 409, "conflict", "Resource already exists"
	case errors.Is(err, core.ErrRateLimited):
		status, code, message = 429, "rate_limited", "Please try again later"
		w.Header().Set("Retry-After", "60")
	case errors.Is(err, core.ErrUnavailable):
		status, code, message = 503, "unavailable", "Service unavailable"
	}
	return writeResult(w, Result{Status: status, Body: map[string]any{"error": map[string]string{"code": code, "message": message}}})
}
