package core

import "testing"

func TestRuleAnalysisSuggestsWithoutVerifying(t *testing.T) {
	for text, kind := range map[string]string{"Hay un incendio en el mercado": "fire", "Me están extorsionando por WhatsApp": "extortion", "Secuestraron a una persona": "kidnapping", "Una persona no respira": "emergency", "Encontré un perro perdido": "pet"} {
		result, err := AnalyzeText(text, "")
		if err != nil || result.SuggestedType != kind || !result.RequiresConfirmation || result.Source != "rules" || len(result.Advice) == 0 {
			t.Errorf("%s: %+v %v", text, result, err)
		}
	}
	result, err := AnalyzeText("No hay incendio, hay un choque de autos", "")
	if err != nil || result.SuggestedType != "crash" {
		t.Fatal("negation ignored", result, err)
	}
	result, err = AnalyzeText("Hay fuego y una persona herida", "")
	if err != nil || result.SuggestedType != "" || len(result.Candidates) != 2 {
		t.Fatal("ambiguous report auto selected", result, err)
	}
}
