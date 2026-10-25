package districtgeo

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"sort"
)

type DistrictOption struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Department string  `json:"department"`
	Province   string  `json:"province"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

func Options() ([]DistrictOption, error) {
	if err := ready(); err != nil {
		return nil, err
	}
	items := make([]DistrictOption, 0, len(catalog))
	seen := map[string]bool{}
	for _, e := range catalog {
		d := e.district
		if seen[d.ID] {
			continue
		}
		seen[d.ID] = true
		items = append(items, DistrictOption{d.ID, d.Name, d.Department, d.Province, e.center[0], e.center[1]})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

// ValidProvince accepts only a province actually present in the reference catalog.
func ValidProvince(id string) bool {
	if len(id) != 4 {
		return false
	}
	if err := ready(); err != nil {
		return false
	}
	for _, e := range catalog {
		if e.district.ID[:4] == id {
			return true
		}
	}
	return false
}
func NationalDirectoryCity(id string) (string, error) {
	if len(id) != 6 {
		return "", core.ErrInvalid
	}
	if _, err := ByID(id); err != nil {
		return "", err
	}
	if metropolitan(id) {
		return "lima", nil
	}
	if id[:4] == "1301" {
		return "trujillo", nil
	}
	return id[:4], nil
}
