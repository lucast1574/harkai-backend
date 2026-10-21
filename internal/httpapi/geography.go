package httpapi

import (
	"context"
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/districtgeo"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (a *API) geographicDistrict(r *http.Request, _ core.User) (Result, error) {
	var d core.GeographicDistrict
	var err error
	var detectedCenter *[2]float64
	query := r.URL.Query()
	if id := query.Get("ubigeo"); id != "" {
		d, err = districtgeo.ByID(id)
	} else if query.Get("auto") == "1" && query.Get("lat") == "" && query.Get("lng") == "" {
		d, detectedCenter, err = resolveAutoDistrict(r)
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
	if detectedCenter != nil {
		centerLat, centerLng = detectedCenter[0], detectedCenter[1]
	}
	return geographyResponse(r, map[string]any{
		"district":  d,
		"area":      area,
		"districts": neighbors,
		"center":    map[string]float64{"latitude": centerLat, "longitude": centerLng},
	})
}

func isCoveredDistrict(id string) bool {
	return len(id) >= 4 && (id[:4] == "1501" || id[:4] == "0701" || id[:4] == "1301")
}

func resolveAutoDistrict(r *http.Request) (core.GeographicDistrict, *[2]float64, error) {
	for _, pair := range [][2]string{
		{"CF-IPLatitude", "CF-IPLongitude"},
		{"X-Vercel-IP-Latitude", "X-Vercel-IP-Longitude"},
		{"X-Geo-Latitude", "X-Geo-Longitude"},
	} {
		lat, latErr := strconv.ParseFloat(r.Header.Get(pair[0]), 64)
		lng, lngErr := strconv.ParseFloat(r.Header.Get(pair[1]), 64)
		if latErr == nil && lngErr == nil {
			if d, err := districtgeo.At(lat, lng); err == nil && isCoveredDistrict(d.ID) {
				pt := [2]float64{lat, lng}
				return d, &pt, nil
			}
		}
	}
	for _, key := range []string{"CF-IPCity", "X-Vercel-IP-City", "X-Geo-City"} {
		city := strings.TrimSpace(r.Header.Get(key))
		if strings.EqualFold(city, "Trujillo") {
			d, err := districtgeo.ByID("130101")
			return d, nil, err
		}
		if strings.EqualFold(city, "Callao") {
			d, err := districtgeo.ByID("070101")
			return d, nil, err
		}
		if strings.EqualFold(city, "Lima") {
			d, err := districtgeo.ByID("150101")
			return d, nil, err
		}
	}
	if d, pt, ok := lookupClientIPGeo(r); ok {
		return d, pt, nil
	}
	d, err := districtgeo.ByID("150101")
	return d, nil, err
}

func clientIP(r *http.Request) string {
	for _, header := range []string{"CF-Connecting-IP", "X-Real-IP", "X-Forwarded-For"} {
		raw := strings.TrimSpace(r.Header.Get(header))
		if raw == "" {
			continue
		}
		first := strings.TrimSpace(strings.Split(raw, ",")[0])
		if ip := net.ParseIP(first); ip != nil && !ip.IsLoopback() && !ip.IsPrivate() {
			return ip.String()
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		if ip := net.ParseIP(host); ip != nil && !ip.IsLoopback() && !ip.IsPrivate() {
			return ip.String()
		}
	}
	return ""
}

func lookupClientIPGeo(r *http.Request) (core.GeographicDistrict, *[2]float64, bool) {
	ip := clientIP(r)
	ctx, cancel := context.WithTimeout(r.Context(), 2500*time.Millisecond)
	defer cancel()

	freeIPURL := "https://freeipapi.com/api/json/"
	if ip != "" {
		freeIPURL += url.PathEscape(ip)
	}
	if req, err := http.NewRequestWithContext(ctx, http.MethodGet, freeIPURL, nil); err == nil {
		if resp, err := http.DefaultClient.Do(req); err == nil {
			var payload struct {
				CityName   string  `json:"cityName"`
				RegionName string  `json:"regionName"`
				Latitude   float64 `json:"latitude"`
				Longitude  float64 `json:"longitude"`
			}
			decodeErr := json.NewDecoder(resp.Body).Decode(&payload)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK && decodeErr == nil && (payload.Latitude != 0 || payload.Longitude != 0) {
				if d, err := districtgeo.At(payload.Latitude, payload.Longitude); err == nil && isCoveredDistrict(d.ID) {
					pt := [2]float64{payload.Latitude, payload.Longitude}
					return d, &pt, true
				}
				if strings.Contains(strings.ToLower(payload.CityName), "trujillo") || strings.Contains(strings.ToLower(payload.RegionName), "libertad") {
					if d, err := districtgeo.ByID("130101"); err == nil {
						return d, nil, true
					}
				}
			}
		}
	}

	ipwhoURL := "https://ipwho.is/"
	if ip != "" {
		ipwhoURL += url.PathEscape(ip)
	}
	if req, err := http.NewRequestWithContext(ctx, http.MethodGet, ipwhoURL, nil); err == nil {
		if resp, err := http.DefaultClient.Do(req); err == nil {
			var payload struct {
				Success   bool    `json:"success"`
				City      string  `json:"city"`
				Region    string  `json:"region"`
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			}
			decodeErr := json.NewDecoder(resp.Body).Decode(&payload)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK && decodeErr == nil && payload.Success {
				if d, err := districtgeo.At(payload.Latitude, payload.Longitude); err == nil && isCoveredDistrict(d.ID) {
					pt := [2]float64{payload.Latitude, payload.Longitude}
					return d, &pt, true
				}
				if strings.Contains(strings.ToLower(payload.City), "trujillo") || strings.Contains(strings.ToLower(payload.Region), "libertad") {
					if d, err := districtgeo.ByID("130101"); err == nil {
						return d, nil, true
					}
				}
			}
		}
	}
	return core.GeographicDistrict{}, nil, false
}
