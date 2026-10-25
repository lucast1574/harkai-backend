package districtgeo

import "testing"

func TestNationalHierarchy(t *testing.T) {
	items, err := Options()
	if err != nil {
		t.Fatal(err)
	}
	departments, provinces, ids := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, d := range items {
		if ids[d.ID] {
			t.Fatalf("duplicate %s", d.ID)
		}
		ids[d.ID] = true
		departments[d.ID[:2]] = true
		provinces[d.ID[:4]] = true
	}
	if len(departments) != 25 || len(provinces) < 190 || len(items) < 1800 {
		t.Fatalf("incomplete hierarchy: %d/%d/%d", len(departments), len(provinces), len(items))
	}
	for _, id := range []string{"040101", "080101", "160101", "200101"} {
		if !ids[id] || !ValidProvince(id[:4]) {
			t.Fatalf("missing province %s", id)
		}
		for _, scope := range []string{"province", "department"} {
			a, e := Area(id, scope)
			if e != nil || a.Scope != scope || a.Geometry.Type != "MultiPolygon" {
				t.Fatalf("area %s %s: %+v %v", id, scope, a, e)
			}
		}
		city, e := NationalDirectoryCity(id)
		if e != nil || city != id[:4] {
			t.Fatalf("directory %s %v", city, e)
		}
	}
	if ValidProvince("9999") {
		t.Fatal("unknown province accepted")
	}
}
