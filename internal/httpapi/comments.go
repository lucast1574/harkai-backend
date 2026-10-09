package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
)

func (a *API) comments(r *http.Request, u core.User) (Result, error) {
	page, err := a.Service.CommentThread(r.Context(), u, r.PathValue("id"), r.URL.Query().Get("cursor"))
	return Result{Body: page}, err
}
func (a *API) comment(r *http.Request, u core.User) (Result, error) {
	var input struct {
		Body string `json:"body"`
	}
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	comment, err := a.Service.AddComment(r.Context(), u, r.PathValue("id"), input.Body)
	return Result{Status: 201, Body: comment}, err
}
func (a *API) deleteComment(r *http.Request, u core.User) (Result, error) {
	return Result{Status: 204}, a.Service.RemoveComment(r.Context(), u, r.PathValue("id"), r.PathValue("commentID"))
}
func (a *API) archive(r *http.Request, _ core.User) (Result, error) {
	q, err := incidentQuery(r, true)
	if err != nil {
		return Result{}, err
	}
	q.IncludeInactive = true
	q.PublicOnly = true
	page, err := a.Service.FindIncidents(r.Context(), q)
	return Result{Body: page}, err
}

func (a *API) districts(r *http.Request, _ core.User) (Result, error) {
	q, err := incidentQuery(r, true)
	if err != nil {
		return Result{}, err
	}
	items, err := a.Service.Districts.ArchiveDistricts(r.Context(), q, a.Service.Now())
	return Result{Body: map[string]any{"items": items}}, err
}
