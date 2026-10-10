package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
)

func (a *API) supportDirectory(r *http.Request, _ core.User) (Result, error) {
	items, err := a.Service.SupportDirectory(r.Context(), r.URL.Query().Get("city"))
	return Result{Body: map[string]any{"items": items}}, err
}
func (a *API) supportPlaces(r *http.Request, _ core.User) (Result, error) {
	query, err := incidentQuery(r, true)
	if err != nil {
		return Result{}, err
	}
	items, err := a.Service.SupportPlaces(r.Context(), query)
	return Result{Body: map[string]any{"items": items}}, err
}
