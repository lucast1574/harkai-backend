package districtgeo

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"sort"
	"strings"
	"sync"
)

//go:embed data/peru-districts.geojson.gz
var data []byte
var loadCatalog = sync.OnceValue(load)
var catalog []entry

const Source = "https://geosenace.senace.gob.pe/arcgis/rest/services/EVA_SERVER/Limite_distrital_2025/MapServer/98"

type entry struct {
	district core.GeographicDistrict
	bounds   [4]float64
}

func load() error {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer reader.Close()
	var collection struct {
		Features []struct {
			Properties struct {
				ID         string `json:"IDDIST"`
				Name       string `json:"NOMBDIST"`
				Province   string `json:"NOMBPROV"`
				Department string `json:"NOMBDEP"`
			}
			Geometry core.DistrictGeometry
		}
	}
	if err = json.NewDecoder(reader).Decode(&collection); err != nil {
		return err
	}
	title := cases.Title(language.Spanish)
	for _, f := range collection.Features {
		d := core.GeographicDistrict{ID: f.Properties.ID, Name: title.String(strings.ToLower(f.Properties.Name)), Province: title.String(strings.ToLower(f.Properties.Province)), Department: title.String(strings.ToLower(f.Properties.Department)), Geometry: f.Geometry, SourceURL: Source, ReferenceYear: 2020}
		bounds := [4]float64{180, 90, -180, -90}
		for _, ring := range d.Geometry.Coordinates {
			for _, p := range ring {
				if len(p) != 2 {
					return fmt.Errorf("invalid district coordinates")
				}
				bounds[0] = min(bounds[0], p[0])
				bounds[1] = min(bounds[1], p[1])
				bounds[2] = max(bounds[2], p[0])
				bounds[3] = max(bounds[3], p[1])
			}
		}
		catalog = append(catalog, entry{d, bounds})
	}
	return nil
}
func ready() error { return loadCatalog() }
func ByID(id string) (core.GeographicDistrict, error) {
	if err := ready(); err != nil {
		return core.GeographicDistrict{}, err
	}
	for _, e := range catalog {
		if e.district.ID == id {
			return e.district, nil
		}
	}
	return core.GeographicDistrict{}, fmt.Errorf("%w: unknown geographic district", core.ErrInvalid)
}
func At(latitude, longitude float64) (core.GeographicDistrict, error) {
	if err := core.ValidateCoordinates(latitude, longitude); err != nil {
		return core.GeographicDistrict{}, err
	}
	if err := ready(); err != nil {
		return core.GeographicDistrict{}, err
	}
	for _, e := range catalog {
		b := e.bounds
		if longitude < b[0] || longitude > b[2] || latitude < b[1] || latitude > b[3] {
			continue
		}
		if contains(e.district.Geometry.Coordinates, longitude, latitude) {
			return e.district, nil
		}
	}
	return core.GeographicDistrict{}, core.ErrNotFound
}
func Province(id string) ([]map[string]string, error) {
	if _, err := ByID(id); err != nil {
		return nil, err
	}
	items := []map[string]string{}
	seen := map[string]bool{}
	for _, e := range catalog {
		if strings.HasPrefix(e.district.ID, id[:4]) && !seen[e.district.ID] {
			seen[e.district.ID] = true
			items = append(items, map[string]string{"id": e.district.ID, "name": e.district.Name})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i]["name"] < items[j]["name"] })
	return items, nil
}

// Warm loads the reference catalog before the server accepts requests.
func Warm() error { return ready() }
