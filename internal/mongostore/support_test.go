package mongostore_test

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	"go.mongodb.org/mongo-driver/v2/bson"
	"testing"
)

func TestSupportSeedCoverageAndProvenance(t *testing.T) {
	districts, places, err := mongostore.ReadSupportSeed()
	if err != nil {
		t.Fatal(err)
	}
	cities := map[string]int{}
	seen := map[string]bool{}
	for _, d := range districts {
		if seen[d.ID] {
			t.Fatal("duplicate district", d.District)
		}
		seen[d.ID] = true
		cities[d.City]++
		national := map[string]bool{}
		for _, c := range d.Contacts {
			if c.Scope == "nacional" {
				national[c.Phone] = true
			}
		}
		for _, phone := range []string{"105", "116", "106", "100", "113"} {
			if !national[phone] {
				t.Fatalf("missing national contact %s in %s", phone, d.District)
			}
		}
	}
	if cities["lima"] != 43 || cities["trujillo"] != 12 {
		t.Fatal("district coverage", cities)
	}
	seen = map[string]bool{}
	for _, p := range places {
		if seen[p.ID] {
			t.Fatal("duplicate place", p.ID)
		}
		seen[p.ID] = true
		if p.LocationSourceURL == "" || len(p.Location.Coordinates) != 2 || p.Location.Coordinates[0] != p.Longitude {
			t.Fatal("coordinate provenance", p.ID)
		}
	}
}

func TestSupportSeedIdempotenceAndGeoIsolation(t *testing.T) {
	ctx := context.Background()
	store := testutil.Store(t)
	if err := store.SeedSupport(ctx); err != nil {
		t.Fatal(err)
	}
	_, err := store.DB.Collection("support_directory").UpdateOne(ctx, bson.M{"city": "lima", "district": "Lima"}, bson.M{"$set": bson.M{"contacts.0.label": "Edited by administrator"}})
	if err != nil {
		t.Fatal(err)
	}
	if err = store.SeedSupport(ctx); err != nil {
		t.Fatal(err)
	}
	for city, want := range map[string]int{"lima": 43, "trujillo": 12} {
		districts, err := store.SupportDirectory(ctx, city)
		if err != nil || len(districts) != want {
			t.Fatalf("%s: count=%d error=%v", city, len(districts), err)
		}
		for _, d := range districts {
			if d.City != city {
				t.Fatal("other city leaked")
			}
			if city == "lima" && d.District == "Lima" && d.Contacts[0].Label != "Edited by administrator" {
				t.Fatal("seed overwrote edit")
			}
		}
	}
	for _, q := range []core.IncidentQuery{{Latitude: -12.0464, Longitude: -77.0428, RadiusMeters: 50000}, {Latitude: -8.1116, Longitude: -79.0287, RadiusMeters: 50000}} {
		places, err := store.SupportPlaces(ctx, q)
		if err != nil || len(places) == 0 {
			t.Fatalf("geo results %v %v", places, err)
		}
		want := "lima"
		if q.Latitude > -10 {
			want = "trujillo"
		}
		for _, p := range places {
			if p.City != want {
				t.Fatal("far place returned", p.ID)
			}
		}
	}
	places, err := store.SupportPlaces(ctx, core.IncidentQuery{Latitude: 40, Longitude: 20, RadiusMeters: 5000})
	if err != nil || len(places) != 0 {
		t.Fatalf("far query %v %v", places, err)
	}
	for _, collection := range []string{"incidents", "push_jobs"} {
		count, err := store.DB.Collection(collection).CountDocuments(ctx, bson.M{})
		if err != nil || count != 0 {
			t.Fatalf("directory created %s: %d %v", collection, count, err)
		}
	}
}
