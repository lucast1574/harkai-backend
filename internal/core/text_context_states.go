package core

import "regexp"

var deniedPayment = regexp.MustCompile(`\bno (?:me |nos |le |les |te )?(exigen|exige|piden|pide|cobran|cobra) (dinero|plata|un pago|una cuota)\b`)
var ownedPet = regexp.MustCompile(`\b(mi|nuestro|nuestra) (gato|gata|perro|perra|mascota|conejo|coneja)\b`)
var returnedPet = regexp.MustCompile(`\b(volvio|regreso|retorno)\b`)
var recoveredPet = regexp.MustCompile(`\bya (lo |la |los |las )?(encontramos|encontre|hallamos|halle|recuperamos|recupere)\b`)

// Separate a denied or resolved state from a new event. Negated return verbs
// ("mi perro no volvió") still describe a missing pet, not a resolved report.
func TextStateExclusions(text string) []string {
	excluded := []string{}
	if deniedPayment.MatchString(text) {
		excluded = append(excluded, "extortion")
	}
	if ownedPet.MatchString(text) {
		if recoveredPet.MatchString(text) {
			excluded = append(excluded, "pet")
		}
		for _, match := range returnedPet.FindAllStringIndex(text, -1) {
			if !negatedMatch(text, match[0]) {
				excluded = append(excluded, "pet")
				break
			}
		}
	}
	return excluded
}
