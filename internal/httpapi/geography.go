package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/districtgeo"
	"net/http"
	"strconv"
)

func (a *API) geographicDistrict(r *http.Request, _ core.User) (Result, error) {
	var d core.GeographicDistrict
	var err error
	if id := r.URL.Query().Get("ubigeo"); id != "" {
		d, err = districtgeo.ByID(id)
	} else {
		lat, latError := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
		lng, lngError := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
		if latError != nil || lngError != nil {
			return Result{}, core.ErrInvalid
		}
		d, err = districtgeo.At(lat, lng)
	}
	if err != nil {
		return Result{}, err
	}
	area, err := districtgeo.Area(d.ID, r.URL.Query().Get("scope"))
	if err != nil {
		return Result{}, err
	}
	neighbors, err := districtgeo.Province(d.ID)
	if err != nil {
		return Result{}, err
	}
	return geographyResponse(r, map[string]any{"district": d, "area": area, "districts": neighbors})
}
