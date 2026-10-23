package core

import "testing"

func TestQuickReportCategoryClassification(t *testing.T) {
	for _, test := range []struct{ text, category, want string }{
		{"Me robaron el celular en la calle", "crime", "theft"},
		{"Están cobrando cupos a la tienda", "crime", "extortion"},
		{"Se lo llevaron a la fuerza", "crime", "kidnapping"},
		{"Hay una persona asesinada en la esquina", "crime", "homicide"},
		{"Están golpeando a una persona", "crime", "violence"},
		{"Hay fuego y una persona herida", "fire", "fire"},
		{"Se perdió mi perro marrón", "pet", "pet"},
	} {
		got, err := ResolveReportType(test.text, test.category)
		if err != nil || got != test.want {
			t.Fatalf("%s: got %s, %v", test.text, got, err)
		}
	}
	for _, test := range [][2]string{{"Me robaron mi celular", "fire"}, {"Hay un incendio", "crime"}, {"No hubo robo en la calle", "crime"}, {"Estoy caminando por aquí", "crime"}} {
		if _, err := ResolveReportType(test[0], test[1]); err == nil {
			t.Fatalf("accepted mismatch: %v", test)
		}
	}
}
