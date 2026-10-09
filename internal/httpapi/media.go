package httpapi

import (
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/security"
	"io"
	"mime"
	"net/http"
)

func (a *API) upload(r *http.Request, user core.User) (Result, error) {
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || (contentType != "image/jpeg" && contentType != "image/png") {
		return Result{}, fmt.Errorf("%w: image/jpeg or image/png body required", core.ErrInvalid)
	}
	r.Body = http.MaxBytesReader(nil, r.Body, security.MaxImageBytes)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return Result{}, fmt.Errorf("%w: image exceeds limit", core.ErrInvalid)
	}
	media, err := a.Service.Upload(r.Context(), user.ID, raw)
	return Result{Status: 201, Body: media}, err
}

func (a *API) download(r *http.Request, user core.User) (Result, error) {
	media, err := a.Service.Download(r.Context(), r.PathValue("id"), user)
	if err != nil {
		return Result{}, err
	}
	return Result{Bytes: media.Data, Headers: http.Header{"Content-Type": {media.ContentType}, "Content-Disposition": {"inline; filename=\"incident.jpg\""}}}, nil
}
