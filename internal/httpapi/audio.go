package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"io"
	"mime"
	"net/http"
	"slices"
	"time"
)

func (a *API) analyzeAudio(r *http.Request, user core.User) (Result, error) {
	if a.Service.TranscribeAudio == nil {
		return Result{}, core.ErrUnavailable
	}
	if err := a.Service.Limits.Allow(r.Context(), "audio:"+user.ID, 10, time.Hour, a.Service.Now()); err != nil {
		return Result{}, err
	}
	kind, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || !slices.Contains([]string{"audio/wav", "audio/x-wav", "audio/webm", "audio/ogg", "audio/mpeg", "audio/mp4"}, kind) {
		return Result{}, core.ErrInvalid
	}
	raw, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, 4<<20))
	if err != nil || len(raw) == 0 {
		return Result{}, core.ErrInvalid
	}
	text, err := a.Service.TranscribeAudio(r.Context(), raw)
	if err != nil {
		return Result{}, err
	}
	analysis, err := core.AnalyzeText(text, "")
	return Result{Body: map[string]any{"text": text, "analysis": analysis, "requires_confirmation": true}}, err
}

func (a *API) mediaSource(r *http.Request, _ core.User) (Result, error) {
	if a.Service.MediaSource == nil {
		return Result{}, core.ErrUnavailable
	}
	data, err := a.Service.MediaSource(r.Context())
	return Result{Bytes: data, Headers: http.Header{"Content-Type": {"application/zip"}, "Content-Disposition": {"attachment; filename=\"harkai-media-worker-source.zip\""}}}, err
}
