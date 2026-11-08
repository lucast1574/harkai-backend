package core

import "testing"

func TestInferenceSpellingNeverInventsEventsFromValidWords(t *testing.T) {
	for _, text := range []string{"Están cobrando cupos", "Estoy probando un robot", "Mariana cocina mientras Roberto está grabando", "El incentivo fue aprobado", "Estamos rodando una película"} {
		if got := NormalizeReportText(text); got != NormalizeText(text) {
			t.Fatalf("changed valid words: %q -> %q", text, got)
		}
	}
	if got := NormalizeReportText("Hay un insendio en el parke"); got != "hay un incendio en el parque" {
		t.Fatal(got)
	}
}
