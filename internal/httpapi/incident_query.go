package httpapi

import (
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
	"strconv"
	"time"
)

func incidentQuery(r *http.Request, geoRequired bool) (core.IncidentQuery, error) {
	values := r.URL.Query()
	q := core.IncidentQuery{RadiusMeters: 5000, Limit: 50, Type: values.Get("type"), Cursor: values.Get("cursor")}
	if geoRequired && (values.Get("lat") == "" || values.Get("lng") == "") {
		return q, fmt.Errorf("%w: lat and lng required", core.ErrInvalid)
	}
	for _, p := range []struct {
		key string
		dst *float64
	}{{"lat", &q.Latitude}, {"lng", &q.Longitude}} {
		if value := values.Get(p.key); value != "" {
			parsed, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return q, fmt.Errorf("%w: %s", core.ErrInvalid, p.key)
			}
			*p.dst = parsed
		}
	}
	for _, p := range []struct {
		key string
		dst *int
	}{{"radius_meters", &q.RadiusMeters}, {"limit", &q.Limit}} {
		if value := values.Get(p.key); value != "" {
			parsed, err := strconv.Atoi(value)
			if err != nil {
				return q, fmt.Errorf("%w: %s", core.ErrInvalid, p.key)
			}
			*p.dst = parsed
		}
	}
	if value := values.Get("after"); value != "" {
		t, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			return q, fmt.Errorf("%w: after must be RFC3339", core.ErrInvalid)
		}
		q.After = &t
	}
	return core.ValidateQuery(q)
}
