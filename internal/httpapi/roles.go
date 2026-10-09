package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
	"strconv"
)

func (a *API) confirmReport(r *http.Request, user core.User) (Result, error) {
	i, err := a.Service.ConfirmReport(r.Context(), user, r.PathValue("id"))
	return Result{Body: i}, err
}

func (a *API) adminUsers(r *http.Request, user core.User) (Result, error) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Result{}, core.ErrInvalid
		}
		limit = parsed
	}
	page, err := a.Service.AdminUsers(r.Context(), user, r.URL.Query().Get("cursor"), limit)
	return Result{Body: page}, err
}

func (a *API) assignRole(r *http.Request, user core.User) (Result, error) {
	if user.Role != "admin" {
		return Result{}, core.ErrForbidden
	}
	var input struct {
		Role string `json:"role"`
	}
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	target, err := a.Service.ManageRole(r.Context(), user, r.PathValue("id"), input.Role)
	return Result{Body: target}, err
}

func (a *API) govIncidents(r *http.Request, user core.User) (Result, error) {
	if !core.CanViewInstitutional(user.Role) {
		return Result{}, core.ErrForbidden
	}
	q, err := incidentQuery(r, true)
	if err != nil {
		return Result{}, err
	}
	q.IncludeInactive = true
	page, err := a.Service.FindIncidents(r.Context(), q)
	return Result{Body: page}, err
}
