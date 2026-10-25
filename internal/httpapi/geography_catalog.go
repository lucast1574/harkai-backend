package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/districtgeo"
	"net/http"
)

func (a *API) geographicCatalog(r *http.Request, _ core.User) (Result, error) {
	items, err := districtgeo.Options()
	if err != nil {
		return Result{}, err
	}
	result, err := geographyResponse(r, map[string]any{"items": items, "source_url": districtgeo.Source, "reference_year": 2020})
	if result.Headers != nil {
		result.Headers.Set("Cache-Control", "public, max-age=86400")
	}
	return result, err
}
