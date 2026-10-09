package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
)

func (a *API) me(_ *http.Request, user core.User) (Result, error) { return Result{Body: user}, nil }

func (a *API) profile(r *http.Request, user core.User) (Result, error) {
	var input struct {
		Name string `json:"name"`
	}
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	updated, err := a.Service.SaveName(r.Context(), user.ID, input.Name)
	return Result{Body: updated}, err
}

func (a *API) preferences(r *http.Request, user core.User) (Result, error) {
	var p core.Preferences
	if err := decode(r, &p); err != nil {
		return Result{}, err
	}
	updated, err := a.Service.SavePreferences(r.Context(), user.ID, p)
	return Result{Body: updated}, err
}
