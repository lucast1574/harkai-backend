package httpapi

import (
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
)

func (a *API) listIncidents(r *http.Request, _ core.User) (Result, error) {
	q, err := incidentQuery(r, true)
	if err != nil {
		return Result{}, err
	}
	page, err := a.Service.FindIncidents(r.Context(), q)
	return Result{Body: page}, err
}

func (a *API) myIncidents(r *http.Request, user core.User) (Result, error) {
	q, err := incidentQuery(r, false)
	if err != nil {
		return Result{}, err
	}
	q.OwnerID = user.ID
	q.IncludeInactive = true
	page, err := a.Service.FindIncidents(r.Context(), q)
	return Result{Body: page}, err
}

func (a *API) adminIncidents(r *http.Request, user core.User) (Result, error) {
	if user.Role != "admin" {
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

func (a *API) report(r *http.Request, user core.User) (Result, error) {
	// Pointers distinguish omitted coordinates from valid zero latitude/longitude.
	var input struct {
		core.IncidentInput
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	}
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	if input.Latitude == nil || input.Longitude == nil {
		return Result{}, fmt.Errorf("%w: latitude and longitude required", core.ErrInvalid)
	}
	input.IncidentInput.Latitude = *input.Latitude
	input.IncidentInput.Longitude = *input.Longitude
	i, err := a.Service.Report(r.Context(), user, input.IncidentInput)
	return Result{Status: 201, Body: i}, err
}

func (a *API) incident(r *http.Request, user core.User) (Result, error) {
	i, err := a.Service.Incident(r.Context(), r.PathValue("id"), user)
	return Result{Body: i}, err
}

func (a *API) incidentStatus(r *http.Request, user core.User) (Result, error) {
	var input struct {
		Status string `json:"status"`
	}
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	i, err := a.Service.ChangeStatus(r.Context(), user, r.PathValue("id"), input.Status)
	return Result{Body: i}, err
}

func (a *API) alerts(r *http.Request, user core.User) (Result, error) {
	q, err := incidentQuery(r, true)
	if err != nil {
		return Result{}, err
	}
	page, err := a.Service.Alerts(r.Context(), user, q)
	return Result{Body: page}, err
}
