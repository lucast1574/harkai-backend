package mongostore_test

import (
	"context"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/districtgeo"
	"github.com/lucast1574/harkai-backend/internal/mongostore"
	"github.com/lucast1574/harkai-backend/internal/testutil"
	"testing"
	"time"
)

func TestDistrictAndCityGeofenceUsesCoordinates(t *testing.T) {
	store := testutil.Store(t)
	ctx := context.Background()
	now := time.Now().UTC()
	for _, p := range []struct{ lat, lng float64 }{{-12.083, -76.992}, {-12.0464, -77.0428}, {-8.1116, -79.0287}} {
		i, err := core.BuildIncident(core.IncidentInput{Type: "fire", Description: "Isolated geographic integration fixture", Latitude: p.lat, Longitude: p.lng, District: "Untrusted label"}, mongostore.NewID(), "test", now)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = store.CreateIncident(ctx, i); err != nil {
			t.Fatal(err)
		}
	}
	for _, scope := range []struct {
		name  string
		count int
	}{{"district", 1}, {"city", 2}} {
		a, err := districtgeo.Area("150130", scope.name)
		if err != nil {
			t.Fatal(err)
		}
		q := core.IncidentQuery{Boundary: &a.Geometry, Latitude: -12.083, Longitude: -76.992, RadiusMeters: 500, Limit: 100}
		page, err := store.ListIncidents(ctx, q, now)
		if err != nil || len(page.Items) != scope.count {
			t.Fatalf("%s: got %d: %v", scope.name, len(page.Items), err)
		}
	}
}
