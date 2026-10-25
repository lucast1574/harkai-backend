package districtgeo

import (
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
)

func Area(id, scope string) (core.GeographicArea, error) {
	d, err := ByID(id)
	if err != nil {
		return core.GeographicArea{}, err
	}
	if scope == "" {
		scope = "district"
	}
	if scope != "district" && scope != "city" && scope != "province" && scope != "department" {
		return core.GeographicArea{}, fmt.Errorf("%w: invalid geographic scope", core.ErrInvalid)
	}
	a := core.GeographicArea{ID: d.ID, Name: d.Name, Province: d.Province, Department: d.Department, Scope: scope, SourceURL: d.SourceURL, ReferenceYear: d.ReferenceYear, Geometry: core.AreaGeometry{Type: "Polygon", Coordinates: d.Geometry.Coordinates}}
	polygons := [][][][]float64{}
	for _, e := range catalog {
		if (scope == "city" && sameCity(e.district.ID, id)) || (scope == "province" && e.district.ID[:4] == id[:4]) || (scope == "department" && e.district.ID[:2] == id[:2]) || (scope == "district" && e.district.ID == id) {
			polygons = append(polygons, e.district.Geometry.Coordinates)
		}
	}
	if len(polygons) > 1 {
		a.Geometry = core.AreaGeometry{Type: "MultiPolygon", Coordinates: polygons}
	}
	if scope == "city" {
		a.Name = d.Province
		if metropolitan(id) {
			a.Name = "Lima y Callao"
		}
	}
	if scope == "province" {
		a.Name = d.Province
	}
	if scope == "department" {
		a.Name = d.Department
	}
	return a, nil
}

// Lima and Callao share the metropolitan exploration area; districts keep their IDs.
func metropolitan(id string) bool {
	return len(id) >= 4 && (id[:4] == "1501" || id[:4] == "0701")
}
func sameCity(left, right string) bool {
	return len(left) >= 4 && len(right) >= 4 && (left[:4] == right[:4] || (metropolitan(left) && metropolitan(right)))
}
