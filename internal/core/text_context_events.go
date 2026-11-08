package core

import (
	"regexp"
	"strings"
)

var eventExercise = regexp.MustCompile(`\b(simulacro|simulacion|prevencion|prevenir|pelicula|serie|novela|disfraz) (de |del |sobre |un |una |el |la |los |las ){0,4}$`)
var virtualEvent = regexp.MustCompile(`\ben (el |un |este )?(videojuego|juego|pelicula|serie) (me |nos |le |les |lo |la |estan |han |habian |habia |se ){0,5}$`)
var personReference = regexp.MustCompile(`\b(persona|menor|chic[oa]|senor[a]?|mujer|hombre|dueno|duena|muchach[oa]|vecin[oa])\b`)
var noiseEvidence = regexp.MustCompile(`\b(musica|parlantes?|ruido[a-z]*|bulla|fiesta|escandalo)\b`)

// These are descriptions of another context, not claims of a real incident.
// Match the grammatical event phrase rather than banning words anywhere in text.
func nonIncidentContext(text, kind string, start, end int) bool {
	before := strings.TrimSpace(text[:start]) + " "
	after := strings.TrimSpace(text[end:])
	term := text[start:end]
	if eventExercise.MatchString(before) || virtualEvent.MatchString(before) {
		return true
	}
	if kind == "kidnapping" && (strings.HasPrefix(term, "subieron") || strings.HasPrefix(term, "metieron") || strings.HasPrefix(term, "arrastraron") || strings.HasPrefix(term, "se llevaron")) && !personReference.MatchString(term) {
		return true
	}
	if kind == "public_disorder" && strings.HasPrefix(term, "no ") && !noiseEvidence.MatchString(text) {
		return true
	}
	if kind == "violence" && strings.Contains(term, "sacando la mugre") && (strings.HasPrefix(after, "al carro") || strings.HasPrefix(after, "a la cocina") || strings.HasPrefix(after, "al piso")) {
		return true
	}
	if kind == "fire" && text[start:end] == "llamas" {
		return strings.HasPrefix(after, "pastando") || strings.HasPrefix(after, "en el zoologico")
	}
	if kind == "crash" && text[start:end] == "choque" {
		return strings.HasPrefix(after, "de opiniones") || strings.HasPrefix(after, "de ideas")
	}
	if kind == "theft" {
		return strings.HasPrefix(after, "idea") || strings.HasPrefix(after, "la idea") || strings.HasPrefix(after, "mi idea") || strings.HasPrefix(after, "el corazon")
	}
	return false
}
