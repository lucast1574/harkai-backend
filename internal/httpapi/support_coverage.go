package httpapi

import (
	"errors"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/districtgeo"
	"net/http"
	"strconv"
)

// A small response allows mobile to check coverage without downloading boundaries.
func (a *API) supportCoverage(r *http.Request, _ core.User) (Result, error) {
	lat, latErr := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, lngErr := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	if latErr != nil || lngErr != nil {
		return Result{}, core.ErrInvalid
	}
	district, err := districtgeo.At(lat, lng)
	if errors.Is(err, core.ErrNotFound) {
		return Result{Body: map[string]any{"covered": false}}, nil
	}
	if err != nil {
		return Result{}, err
	}
	province := district.ID[:4]
	covered := districtgeo.ValidProvince(province)
	return Result{Body: map[string]any{"covered": covered, "district": district.Name,
		"district_id": district.ID, "province": district.Province, "province_id": province, "department": district.Department, "national": true}}, nil
}
