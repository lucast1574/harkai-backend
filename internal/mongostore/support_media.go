package mongostore

import "github.com/lucast1574/harkai-backend/internal/core"

var supportMedia, supportMediaError = supportMediaCatalog()

func supportMediaCatalog() (map[string]core.SupportPlace, error) {
	_, places, err := ReadSupportSeed()
	catalog := make(map[string]core.SupportPlace, len(places))
	for _, place := range places {
		catalog[place.ID] = place
	}
	return catalog, err
}

// Fill missing source media without overwriting curated database contacts.
func EnrichSupportPlaces(places []core.SupportPlace) ([]core.SupportPlace, error) {
	if supportMediaError != nil {
		return nil, supportMediaError
	}
	for i := range places {
		source, exists := supportMedia[places[i].ID]
		if !exists {
			continue
		}
		if places[i].Phone == "" && len(places[i].Phones) == 0 {
			places[i].Phone, places[i].Phones = source.Phone, source.Phones
		}
		if len(places[i].PhotoURLs) == 0 {
			places[i].PhotoURLs = source.PhotoURLs
		}
	}
	return places, nil
}
