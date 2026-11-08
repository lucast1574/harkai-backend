package core

import (
	"regexp"
	"slices"
)

var physicalViolence = regexp.MustCompile(`\b(golpe[a-z]*|dispar[a-z]*|pelea|peleando|patadas|punetazos|pateando|cachetada|agresi[a-z]*)\b`)
var separateEvents = regexp.MustCompile(`\b(ademas|tambien|al otro lado|local de al lado|otra persona)\b`)

// A financial threat is extortion; an explicitly lethal attack is homicide.
// Keep distinct incidents ambiguous instead of choosing the first keyword.
func PrincipalEventCandidates(text string, candidates []string) []string {
	if slices.Contains(candidates, "extortion") && !physicalViolence.MatchString(text) {
		candidates = withoutCandidate(candidates, "violence")
	}
	if slices.Contains(candidates, "homicide") && !separateEvents.MatchString(text) {
		candidates = withoutCandidate(candidates, "violence")
	}
	return candidates
}

func withoutCandidate(candidates []string, kind string) []string {
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate != kind {
			result = append(result, candidate)
		}
	}
	return result
}
