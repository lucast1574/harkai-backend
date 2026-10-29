package mongostore

import "testing"

func TestEmergencySupportSeed(t *testing.T) {
	_, places, err := ReadSupportSeed()
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, p := range places {
		if p.Kind != "police" && p.Kind != "fire_station" {
			continue
		}
		counts[p.Kind]++
		if p.DistrictID == "" || p.Phone != "" || len(p.Phones) != 0 || p.LocationSourceURL != p.SourceURL {
			t.Fatalf("incomplete provenance or invented contact: %s", p.ID)
		}
		invalid := p
		invalid.Kind = "unknown"
		if ValidateSupportPlace(invalid) == nil {
			t.Fatal("unknown kind accepted")
		}
	}
	if counts["police"] != 1332 || counts["fire_station"] != 210 {
		t.Fatalf("incomplete official snapshot: %v", counts)
	}
}
