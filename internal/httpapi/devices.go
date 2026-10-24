package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
	"strings"
	"time"
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
		Latitude   *float64  `json:"latitude"`
		Accuracy   *float64  `json:"accuracy,omitempty"`
		ObservedAt time.Time `json:"observed_at,omitempty"`
		Longitude  *float64  `json:"longitude"`
	}
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	result, err := a.Service.MoveDevice(r.Context(), u, strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), r.PathValue("deviceID"), input.Latitude, input.Longitude)
	if err != nil {
		return Result{}, err
	}
	if input.Accuracy != nil {
		notices, e := a.Service.VisitBusinesses(r.Context(), u, result, *input.Latitude, *input.Longitude, *input.Accuracy, input.ObservedAt)
		if e == nil {
			return Result{Body: struct {
				core.Device
				Notices []core.BusinessNotice `json:"business_notices"`
			}{result, notices}}, nil
		}
	}
	return Result{Body: result}, nil
}
