package core

import "testing"

func TestQuickReportCategoryClassification(t *testing.T) {
	for _, test := range []struct{ text, category, want string }{
		{"Me robaron el celular en la calle", "crime", "theft"},
		{"Están cobrando cupos a la tienda", "crime", "extortion"},
		{"Se lo llevaron a la fuerza", "crime", "kidnapping"},
		{"Hay una persona asesinada en la esquina", "crime", "homicide"},
		{"Están golpeando a una persona", "crime", "violence"},
		{"Hay personas tomando cerveza en la calle", "crime", "public_drinking"},
		{"Hay gente tomando en la vía pública", "crime", "public_drinking"},
		{"Están fumando marihuana en el parque", "crime", "public_drug_use"},
		{"Hay fumones en la calle", "crime", "public_drug_use"},
		{"Están causando disturbios en la plaza", "crime", "public_disorder"},
		{"Están tomando cerveza y golpeando a una persona en la calle", "crime", "violence"},
		{"Hay fuego y una persona herida", "fire", "fire"},
		{"Incendio", "crime", "fire"},
		{"Me robaron mi celular", "fire", "theft"},
		{"Hay un incendio", "crime", "fire"},
		{"Un perro perdido en el parque", "crash", "pet"},
		{"Se perdió mi perro marrón", "pet", "pet"},
	} {
		got, err := ResolveReportType(test.text, test.category)
		if err != nil || got != test.want {
			t.Fatalf("%s: got %s, %v", test.text, got, err)
		}
	}
	for _, test := range [][2]string{{"No hubo robo en la calle", "crime"}, {"Estoy caminando por aquí", "crime"}, {"Hay un robo y están golpeando a alguien", "crime"}} {
		if _, err := ResolveReportType(test[0], test[1]); err == nil {
			t.Fatalf("accepted mismatch: %v", test)
		}
	}
}

func TestPublicConductRemainsANonCriticalCrimeSubtype(t *testing.T) {
	for _, kind := range []string{"public_drinking", "public_drug_use", "public_disorder"} {
		if !IsCrime(kind) || CriticalType(kind) {
			t.Fatalf("%s must be a noncritical crime subtype", kind)
		}
		if got, err := ManualReportType(kind, "crime"); err != nil || got != kind {
			t.Fatalf("manual subtype %s: %s, %v", kind, got, err)
		}
		incident := PublicIncident(Incident{Type: kind})
		if incident.Category != "crime" || incident.Subcategory != kind {
			t.Fatalf("incorrect public grouping: %+v", incident)
		}
		found := false
		for _, group := range ReportCategories() {
			if group.ID == kind {
				t.Fatalf("%s exposed as a top-level category", kind)
			}
			for _, subtype := range group.Subcategories {
				if subtype.ID == kind && group.ID == "crime" {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("%s missing from crime metadata", kind)
		}
	}
}
