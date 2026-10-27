package core

import (
	"regexp"
	"strings"
)

// Limit negation to a grammatical event phrase, not any "no" in four words.
// "No se que hacer, aqui estan extorsionando" still describes an extortion.
var eventNegation = regexp.MustCompile(`\b(no|nunca|sin) (?:hay |hubo |habia |han |ha |he |se |esta |estan |estaba |estaban |es |era |fue |fueron |ocurrio |existe |existen |ningun |ninguna |un |una |el |la |los |las |evidencia de |signos de |personas? |sido |siendo |reportado |reportes? de ){0,6}$`)
var clauseBoundary = regexp.MustCompile(`[.,;!?:\n]|\b(pero|sino|aunque|ahora)\b`)

func negatedMatch(text string, start int) bool {
	prefix := text[:start]
	boundaries := clauseBoundary.FindAllStringIndex(prefix, -1)
	if len(boundaries) > 0 {
		prefix = prefix[boundaries[len(boundaries)-1][1]:]
	}
	return eventNegation.MatchString(strings.TrimLeft(prefix, " "))
}

func contextualMismatch(text, kind string, start, end int) bool {
	if kind != "fire" {
		return false
	}
	term := text[start:end]
	before := strings.TrimSpace(text[:start])
	if term == "fuego" {
		for _, suffix := range []string{"arma de", "armas de", "abrieron", "abrio", "cruce de", "intercambio de"} {
			if strings.HasSuffix(before, suffix) {
				return true
			}
		}
	}
	// "Me llamas" and animal llamas do not describe a fire.
	if term == "llamas" && strings.HasSuffix(before, "me") {
		return true
	}
	return false
}
