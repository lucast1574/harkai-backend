package districtgeo

import (
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
	"strings"
)

func Area(id, scope string) (core.GeographicArea, error) {
	d, err := ByID(id)
	if err != nil {
		return core.GeographicArea{}, err
	}
	if scope == "" {
		scope = "district"
	}
	if scope != "district" && scope != "city" {
		return core.GeographicArea{}, fmt.Errorf("%w: invalid geographic scope", core.ErrInvalid)
	}
	a := core.GeographicArea{ID: d.ID, Name: d.Name, Province: d.Province, Department: d.Department, Scope: scope, SourceURL: d.SourceURL, ReferenceYear: d.ReferenceYear, Geometry: core.AreaGeometry{Type: "Polygon", Coordinates: d.Geometry.Coordinates}}
	polygons := [][][][]float64{}
	for _, e := range catalog {
		if (scope == "city" && strings.HasPrefix(e.district.ID, id[:4])) || (scope == "district" && e.district.ID == id) {
			polygons = append(polygons, e.district.Geometry.Coordinates)
		}
	}
	if len(polygons) > 1 {
		a.Geometry = core.AreaGeometry{Type: "MultiPolygon", Coordinates: polygons}
	}
	if scope == "city" {
		a.Name = d.Province
	}
	return a, nil
}
