package mongostore_test

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"testing"
)

func TestSupportMediaEnrichmentPreservesCuratedContacts(t *testing.T) {
	places, err := mongostore.EnrichSupportPlaces([]core.SupportPlace{
		{ID: "renipress-00009116"},
		{ID: "renipress-00009116", Phone: "+5115555555", PhotoURLs: []string{"https://example.com/curated.jpg"}},
		{ID: "unknown"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if places[0].Phone != "+5112190000" || len(places[0].PhotoURLs) != 3 {
		t.Fatalf("missing verified registry media: %+v", places[0])
	}
	if places[1].Phone != "+5115555555" || places[1].PhotoURLs[0] != "https://example.com/curated.jpg" {
		t.Fatal("curated record overwritten")
	}
	if places[2].Phone != "" || len(places[2].PhotoURLs) != 0 {
		t.Fatal("invented contact")
	}
}
