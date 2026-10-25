package mongostore

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/districtgeo"
)

// Outside curated municipalities, publish only national lines, never invented local numbers.
func nationalSupportDirectory(curated []core.SupportDistrict) ([]core.SupportDistrict, error) {
	options, err := districtgeo.Options()
	if err != nil {
		return nil, err
	}
	national := []core.SupportContact{}
	if len(curated) > 0 {
		for _, c := range curated[0].Contacts {
			if c.Scope == "nacional" {
				national = append(national, c)
			}
		}
	}
	for _, d := range options {
		city, err := districtgeo.NationalDirectoryCity(d.ID)
		if err != nil {
			return nil, err
		}
		if city == "lima" || city == "trujillo" {
			continue
		}
		curated = append(curated, core.SupportDistrict{Country: "PE", City: city, District: d.Name, DistrictID: d.ID, Department: d.Department, Province: d.Province, Contacts: national})
	}
	return curated, nil
}
