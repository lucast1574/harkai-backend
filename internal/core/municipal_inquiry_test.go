package core

import (
	"strings"
	"testing"
)

func TestMunicipalContactValidation(t *testing.T) {
	in := MunicipalInquiry{Municipality: " Municipalidad de prueba ", Name: " Ana Pérez ", Role: " Gestión local ", Email: "ana@example.com", Message: "Nos gustaría conocer el acceso municipal.", Consent: true}
	got, err := ValidateMunicipalInquiry(in)
	if err != nil || got.Name != "Ana Pérez" || got.Municipality != "Municipalidad de prueba" {
		t.Fatal(got, err)
	}
	for _, update := range []func(*MunicipalInquiry){
		func(x *MunicipalInquiry) { x.Email = "not-an-email" },
		func(x *MunicipalInquiry) { x.Consent = false },
		func(x *MunicipalInquiry) { x.Message = "short" },
		func(x *MunicipalInquiry) { x.Message = strings.Repeat("a", 3001) },
		func(x *MunicipalInquiry) { x.Municipality = "" },
	} {
		draft := in
		update(&draft)
		if _, err = ValidateMunicipalInquiry(draft); err == nil {
			t.Fatal("accepted invalid inquiry")
		}
	}
}
