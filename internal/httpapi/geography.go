package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/districtgeo"
	"net/http"
	"strconv"
	"strings"
)

func (a *API) geographicDistrict(r *http.Request, _ core.User) (Result, error) {
	var d core.GeographicDistrict
	var err error
	query := r.URL.Query()
	if id := query.Get("ubigeo"); id != "" {
		d, err = districtgeo.ByID(id)
	} else if query.Get("auto") == "1" && query.Get("lat") == "" && query.Get("lng") == "" {
		d, err = resolveAutoDistrict(r)
	} else {
		lat, latError := strconv.ParseFloat(query.Get("lat"), 64)
		lng, lngError := strconv.ParseFloat(query.Get("lng"), 64)
		if latError != nil || lngError != nil {
			return Result{}, core.ErrInvalid
		}
		d, err = districtgeo.At(lat, lng)
	}
	if err != nil {
		return Result{}, err
	}
	area, err := districtgeo.Area(d.ID, query.Get("scope"))
	if err != nil {
		return Result{}, err
	}
	neighbors, err := districtgeo.Province(d.ID)
	if err != nil {
		return Result{}, err
	}
	centerLat, centerLng, _ := districtgeo.Center(d.ID)
	return geographyResponse(r, map[string]any{
		"district":  d,
		"area":      area,
		"districts": neighbors,
		"center":    map[string]float64{"latitude": centerLat, "longitude": centerLng},
	})
}

func resolveAutoDistrict(r *http.Request) (core.GeographicDistrict, error) {
	for _, pair := range [][2]string{
		{"CF-IPLatitude", "CF-IPLongitude"},
		{"X-Vercel-IP-Latitude", "X-Vercel-IP-Longitude"},
		{"X-Geo-Latitude", "X-Geo-Longitude"},
	} {
		lat, latErr := strconv.ParseFloat(r.Header.Get(pair[0]), 64)
		lng, lngErr := strconv.ParseFloat(r.Header.Get(pair[1]), 64)
		if latErr == nil && lngErr == nil {
			if d, err := districtgeo.At(lat, lng); err == nil {
				return d, nil
			}
		}
	}
	for _, key := range []string{"CF-IPCity", "X-Vercel-IP-City", "X-Geo-City"} {
		city := strings.TrimSpace(r.Header.Get(key))
		if strings.EqualFold(city, "Trujillo") {
			return districtgeo.ByID("130101")
		}
		if strings.EqualFold(city, "Callao") {
			return districtgeo.ByID("070101")
		}
	}
	return districtgeo.ByID("150101")
}
