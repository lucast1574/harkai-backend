package core

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type TextAnalysis struct {
	SuggestedType        string   `json:"suggested_type,omitempty"`
	Candidates           []string `json:"candidates"`
	MatchedTerms         []string `json:"matched_terms"`
	RequiresConfirmation bool     `json:"requires_confirmation"`
	Reason               string   `json:"reason"`
	Advice               []string `json:"advice"`
	Source               string   `json:"source"`
}

func NormalizeText(text string) string {
	return strings.Join(strings.Fields(strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u").Replace(strings.ToLower(text))), " ")
}

func AnalyzeText(text, selected string) (TextAnalysis, error) {
	if utf8.RuneCountInString(strings.TrimSpace(text)) < 5 || utf8.RuneCountInString(text) > 2000 {
		return TextAnalysis{}, fmt.Errorf("%w: text must contain 5 to 2000 characters", ErrInvalid)
	}
	if selected != "" && selected != "crime" {
		if _, ok := CategoryFor(selected); !ok {
			return TextAnalysis{}, fmt.Errorf("%w: selected_type", ErrInvalid)
		}
	}
	result := TextAnalysis{Candidates: []string{}, MatchedTerms: []string{}, RequiresConfirmation: true, Source: "rules", Advice: []string{"Confirma la categoría y revisa el texto antes de publicar."}, Reason: "No encontré una coincidencia clara. Puedes elegir la categoría manualmente."}
	text = NormalizeText(text)
	for _, rule := range textRules {
		for _, match := range rule.pattern.FindAllStringIndex(text, -1) {
			if negatedMatch(text, match[0]) || contextualMismatch(text, rule.kind, match[0], match[1]) {
				continue
			}
			result.Candidates = append(result.Candidates, rule.kind)
			result.MatchedTerms = append(result.MatchedTerms, text[match[0]:match[1]])
			break
		}
	}
	if len(result.Candidates) == 1 {
		result.SuggestedType = result.Candidates[0]
		result.Reason = "Encontré palabras relacionadas con esta categoría. Es una sugerencia, no una verificación de los hechos."
	} else if len(result.Candidates) > 1 {
		result.Reason = "El texto menciona más de una categoría. Confirma cuál describe el hecho principal."
	}
	if selected != "" {
		result.SuggestedType = ""
		for _, candidate := range result.Candidates {
			if candidate == selected || selected == "crime" && IsCrime(candidate) {
				result.SuggestedType = candidate
				result.Reason = "Categoría identificada por las reglas de texto. Revisa tu descripción antes de publicar."
				break
			}
		}
		if result.SuggestedType == "" {
			result.Reason = "Describe con más detalle lo ocurrido para identificar la categoría seleccionada."
		}
	}
	adviceType := result.SuggestedType
	if selected != "" && selected != "crime" {
		adviceType = selected
	}
	if category, ok := CategoryFor(adviceType); ok {
		result.Advice = category.Advice
	}
	return result, nil
}
