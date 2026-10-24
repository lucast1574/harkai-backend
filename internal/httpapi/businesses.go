package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (a *API) publications(r *http.Request, _ core.User) (Result, error) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, e := strconv.Atoi(raw)
		if e != nil {
			return Result{}, core.ErrInvalid
		}
		limit = n
	}
	page, err := a.Service.PublicPublications(r.Context(), r.URL.Query().Get("cursor"), limit)
	return Result{Body: page}, err
}
func (a *API) businessList(r *http.Request, u core.User) (Result, error) {
	own := strings.HasPrefix(r.URL.Path, "/v1/me/")
	admin := strings.HasPrefix(r.URL.Path, "/v1/admin/")
	q, err := incidentQuery(r, !own && !admin)
	if err != nil {
		return Result{}, err
	}
	page, err := a.Service.FindBusinesses(r.Context(), u, q, own, admin)
	return Result{Body: page}, err
}
func (a *API) business(r *http.Request, u core.User) (Result, error) {
	p, err := a.Service.Business(r.Context(), u, r.PathValue("id"))
	return Result{Body: p}, err
}
func (a *API) saveBusiness(r *http.Request, u core.User) (Result, error) {
	var input struct {
		core.BusinessInput
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	}
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	if input.Latitude == nil || input.Longitude == nil {
		return Result{}, core.ErrInvalid
	}
	input.BusinessInput.Latitude, input.BusinessInput.Longitude = *input.Latitude, *input.Longitude
	p, err := a.Service.SaveBusiness(r.Context(), u, r.PathValue("id"), input.BusinessInput)
	status := 200
	if r.Method == http.MethodPost {
		status = 201
	}
	return Result{Status: status, Body: p}, err
}
func (a *API) moderateBusiness(r *http.Request, u core.User) (Result, error) {
	var in struct {
		Status string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		return Result{}, err
	}
	p, err := a.Service.ReviewBusiness(r.Context(), u, r.PathValue("id"), in.Status)
	return Result{Body: p}, err
}
func (a *API) enableBusiness(r *http.Request, u core.User) (Result, error) {
	var in struct {
		Enabled *bool `json:"enabled"`
	}
	if err := decode(r, &in); err != nil {
		return Result{}, err
	}
	if in.Enabled == nil {
		return Result{}, core.ErrInvalid
	}
	p, err := a.Service.EnableBusiness(r.Context(), u, r.PathValue("id"), *in.Enabled)
	return Result{Body: p}, err
}
func (a *API) businessBilling(r *http.Request, _ core.User) (Result, error) {
	if a.Service.Billing == nil {
		return Result{}, core.ErrUnavailable
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 65537))
	if err != nil || len(raw) > 65536 {
		return Result{}, core.ErrInvalid
	}
	return Result{Status: 204}, a.Service.BusinessBillingEvent(r.Context(), raw, r.Header.Get("Stripe-Signature"))
}
func (a *API) ackBusinessNotice(r *http.Request, u core.User) (Result, error) {
	return Result{Status: 204}, a.Service.AckBusinessNotice(r.Context(), u, strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), r.PathValue("id"))
}

func (a *API) ackProximityNotice(r *http.Request, u core.User) (Result, error) {
	var in struct {
		Stage int `json:"stage"`
	}
	if err := decode(r, &in); err != nil {
		return Result{}, err
	}
	return Result{Status: 204}, a.Service.AckProximityNotice(r.Context(), u, strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), r.PathValue("id"), in.Stage)
}
