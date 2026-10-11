package districtgeo

import (
	"encoding/json"
	"testing"
	"time"
)

func TestReferenceDistrictsAndCity(t *testing.T) {
	started := time.Now()
	for _, p := range []struct {
		lat, lng float64
		id       string
		count    int
	}{{-12.083, -76.992, "150130", 50}, {-8.1116, -79.0287, "130101", 11}} {
		d, err := At(p.lat, p.lng)
		if err != nil || d.ID != p.id {
			t.Fatalf("district %+v: %v", d, err)
		}
		neighbors, err := Province(d.ID)
		if err != nil || len(neighbors) != p.count {
			t.Fatalf("neighbors %d: %v", len(neighbors), err)
		}
		a, err := Area(d.ID, "city")
		if err != nil || a.Geometry.Type != "MultiPolygon" || (a.Name != d.Province && a.Name != "Lima y Callao") {
			t.Fatalf("city: %+v %v", a, err)
		}
		b, _ := json.Marshal(a)
		t.Logf("%s geometry response: %d bytes", a.Name, len(b))
	}
	t.Logf("cold catalog+lookups: %s", time.Since(started))
	if _, err := ByID("bad"); err == nil {
		t.Fatal("invalid ID accepted")
	}
	if _, err := At(0, 0); err == nil {
		t.Fatal("outside Peru accepted")
	}
	if _, err := Area("150130", "bogus"); err == nil {
		t.Fatal("invalid scope accepted")
	}
}
func TestPolygonHolesAndEdges(t *testing.T) {
	rings := [][][]float64{{{0, 0}, {4, 0}, {4, 4}, {0, 4}, {0, 0}}, {{1, 1}, {2, 1}, {2, 2}, {1, 2}, {1, 1}}}
	if !contains(rings, 3, 3) || contains(rings, 1.5, 1.5) || !contains(rings, 0, 2) || contains(rings, 5, 5) {
		t.Fatal("hole/edge inclusion incorrect")
	}
}
func BenchmarkDistrictLookup(b *testing.B) {
	_, _ = At(-12.083, -76.992)
	b.ResetTimer()
	for range b.N {
		_, _ = At(-12.083, -76.992)
	}
}

func TestLimaCallaoMetropolitanCoverage(t *testing.T) {
	for _, id := range []string{"150130", "070101", "070107"} {
		districts, err := Province(id)
		if err != nil || len(districts) != 50 {
			t.Fatalf("metro %s: %d %v", id, len(districts), err)
		}
		seen := map[string]bool{}
		for _, district := range districts {
			seen[district["id"]] = true
		}
		for _, callao := range []string{"070101", "070102", "070103", "070104", "070105", "070106", "070107"} {
			if !seen[callao] {
				t.Fatalf("missing Callao district %s", callao)
			}
		}
		area, err := Area(id, "city")
		if err != nil || area.Name != "Lima y Callao" {
			t.Fatalf("metro area: %+v %v", area, err)
		}
		district, err := Area(id, "district")
		if err != nil || district.ID != id || district.Scope != "district" {
			t.Fatalf("district lost: %+v %v", district, err)
		}
	}
}
