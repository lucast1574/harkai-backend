package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
)

func (a *API) contacts(r *http.Request, _ core.User) (Result, error) {
	c, err := a.Service.Contacts(r.Context(), r.URL.Query().Get("country"), r.URL.Query().Get("city"))
	return Result{Body: c}, err
}

func (a *API) saveContacts(r *http.Request, user core.User) (Result, error) {
	if user.Role != "admin" {
		return Result{}, core.ErrForbidden
	}
	var input core.EmergencyContacts
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	c, err := a.Service.SaveContacts(r.Context(), user, input)
	return Result{Body: c}, err
}
