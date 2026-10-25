package mongostore

import "testing"

func TestNationalSupportSeed(t *testing.T) {
	districts, places, err := ReadSupportSeed()
	if err != nil {
		t.Fatal(err)
	}
	departments := map[string]int{}
	for _, p := range places {
		if p.Department != "" {
			departments[p.Department]++
		}
	}
	if len(departments) != 25 || len(places) < 11000 {
		t.Fatalf("national health coverage: %d departments, %d places", len(departments), len(places))
	}
	for _, city := range []string{"0401", "0801", "1601", "2001"} {
		count := 0
		for _, d := range districts {
			if d.City != city {
				continue
			}
			count++
			if d.DistrictID == "" {
				t.Fatal("missing district ID")
			}
			for _, c := range d.Contacts {
				if c.Scope != "nacional" {
					t.Fatalf("unverified local contact %+v", c)
				}
			}
		}
		if count == 0 {
			t.Fatalf("missing directory %s", city)
		}
	}
}
