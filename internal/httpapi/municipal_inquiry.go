package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
	"time"
)

func (a *API) municipalInquiry(r *http.Request, _ core.User) (Result, error) {
	if err := a.authLimit(r, "municipal-ip:"+clientIP(r), 10, time.Hour); err != nil {
		return Result{}, err
	}
	var input core.MunicipalInquiry
	if err := decodeSized(r, &input, 16384); err != nil {
		return Result{}, err
	}
	ticket, err := a.Service.CreateMunicipalInquiry(r.Context(), input)
	return Result{Status: 201, Body: map[string]string{"id": ticket.ID, "status": "received"}}, err
}
