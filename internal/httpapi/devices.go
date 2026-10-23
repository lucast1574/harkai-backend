package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
	"strings"
)

func (a *API) device(r *http.Request, u core.User) (Result, error) {
	var input core.DeviceInput
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	result, err := a.Service.RegisterDevice(r.Context(), u, strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), input)
	return Result{Body: result}, err
}
func (a *API) deleteDevice(r *http.Request, u core.User) (Result, error) {
	return Result{Status: 204}, a.Service.RemoveDevice(r.Context(), u, r.PathValue("deviceID"))
}

func (a *API) deviceLocation(r *http.Request, u core.User) (Result, error) {
	var input struct {
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	}
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	result, err := a.Service.MoveDevice(r.Context(), u, strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), r.PathValue("deviceID"), input.Latitude, input.Longitude)
	return Result{Body: result}, err
}
