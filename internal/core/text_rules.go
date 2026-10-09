package core

import (
	"fmt"
	"regexp"
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

type textRule struct {
	kind    string
	pattern *regexp.Regexp
}

var textRules = []textRule{
	{"extortion", regexp.MustCompile(`\b(extorsion\w*|cobro de cupos|pago de cupos|pagar (un |el )?cupo|amenaz\w* (de muerte|por telefono|por whatsapp)|me (estan |esta )?(extorsionando|amenazando))\b`)},
	{"kidnapping", regexp.MustCompile(`\b(secuestr\w*|raptar\w*|raptaron|privad[oa] de (su |la )?libertad|se lo llevaron a la fuerza)\b`)},
	{"homicide", regexp.MustCompile(`\b(asesinat\w*|asesinaron|homicidi\w*|sicari\w*|persona (asesinada|muerta)|cadaver)\b`)},
	{"fire", regexp.MustCompile(`\b(incendi\w*|fuego|llamas|quemando|humo (negro|denso))\b`)},
	{"crash", regexp.MustCompile(`\b(choqu\w*|colision\w*|atropell\w*|accidente (vial|de transito|de auto)|volcadur\w*)\b`)},
	{"theft", regexp.MustCompile(`\b(rob[oa]\w*|robar\w*|asalt\w*|hurto\w*)\b`)},
	{"violence", regexp.MustCompile(`\b(agresi\w*|agred\w*|golpeando|golpearon|violencia|dispar\w*|balacera|tiroteo|violacion)\b`)},
	{"emergency", regexp.MustCompile(`\b(desmay\w*|no respira|no puede respirar|herid\w*|sangrand\w*|convulsion\w*|emergencia medica|ambulancia)\b`)},
	{"pet", regexp.MustCompile(`\b(mascota|perr[oa]|gat[oa])\b.*\b(perdid\w*|extraviad\w*|encontrad\w*)\b|\b(perdid\w*|extraviad\w*|encontrad\w*)\b.*\b(mascota|perr[oa]|gat[oa])\b`)},
	{"place", regexp.MustCompile(`\b(lugar de ayuda|puesto de auxilio|refugio|centro de apoyo|registrar (un |una )?(lugar|negocio))\b`)},
}

func NormalizeText(text string) string {
	return strings.Join(strings.Fields(strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u").Replace(strings.ToLower(text))), " ")
}

func AnalyzeText(text, selected string) (TextAnalysis, error) {
	if utf8.RuneCountInString(strings.TrimSpace(text)) < 5 || utf8.RuneCountInString(text) > 2000 {
		return TextAnalysis{}, fmt.Errorf("%w: text must contain 5 to 2000 characters", ErrInvalid)
	}
	if selected != "" {
		if _, ok := CategoryFor(selected); !ok {
			return TextAnalysis{}, fmt.Errorf("%w: selected_type", ErrInvalid)
		}
	}
	result := TextAnalysis{Candidates: []string{}, MatchedTerms: []string{}, RequiresConfirmation: true, Source: "rules", Advice: []string{"Confirma la categoría y revisa el texto antes de publicar."}, Reason: "No encontré una coincidencia clara. Puedes elegir la categoría manualmente."}
	text = NormalizeText(text)
	for _, rule := range textRules {
		for _, match := range rule.pattern.FindAllStringIndex(text, -1) {
			if negatedMatch(text, match[0]) {
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
	adviceType := result.SuggestedType
	if selected != "" {
		adviceType = selected
	}
	if category, ok := CategoryFor(adviceType); ok {
		result.Advice = category.Advice
	}
	return result, nil
}

func negatedMatch(text string, start int) bool {
	prefix := text[:start]
	if i := strings.LastIndexAny(prefix, ".,;!?\n"); i >= 0 {
		prefix = prefix[i+1:]
	}
	words := strings.Fields(prefix)
	if len(words) > 4 {
		words = words[len(words)-4:]
	}
	for _, word := range words {
		if word == "no" || word == "sin" || word == "nunca" {
			return true
		}
	}
	return false
}
