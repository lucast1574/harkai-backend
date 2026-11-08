package core

import (
	"encoding/json"
	"os"
	"testing"
)

func TestTextVariations(t *testing.T) {
	for _, name := range []string{"text_variations", "text_additional", "text_regressions"} {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile("testdata/" + name + ".json")
			if err != nil {
				t.Fatal(err)
			}
			var cases []struct{ Text, Type, Variation string }
			if err := json.Unmarshal(raw, &cases); err != nil {
				t.Fatal(err)
			}
			correct, manual, rejected, wrong := 0, 0, 0, 0
			for _, test := range cases {
				got, err := AnalyzeText(test.Text, "")
				if err != nil {
					t.Fatal(err)
				}
				if got.SuggestedType != "" && got.SuggestedType != test.Type {
					wrong++
					t.Errorf("%q: got %q; expected %q or manual", test.Text, got.SuggestedType, test.Type)
				} else if test.Type == "" {
					rejected++
				} else if got.SuggestedType == "" {
					manual++
				} else {
					correct++
				}
			}
			t.Logf("%d automatic correct; %d manual; %d non-incidents rejected; %d incorrect", correct, manual, rejected, wrong)
		})
	}
}

func BenchmarkShortReportClassification(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := AnalyzeText("Hay gente tomando cerveza en la calle", "crime"); err != nil {
			b.Fatal(err)
		}
	}
}
