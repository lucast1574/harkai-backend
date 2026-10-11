package core

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Boundaries avoid substring false positives such as computadora and diputado.
// Separators and common character substitutions catch basic spelling evasion.
var communityLanguage = compileCommunityLanguage()

func compileCommunityLanguage() *regexp.Regexp {
	words := []string{"puta", "putas", "puto", "putos", "mierda", "mierdas", "carajo", "carajos", "huevon", "huevona", "huevones", "huevonas", "weon", "weones", "cojudo", "cojuda", "cojudos", "cojudas", "chucha", "conchatumadre", "conchadetumadre", "ctm", "csm", "cabron", "cabrones", "pendejo", "pendeja", "joder", "jodete", "fuck", "fucking", "bitch", "asshole", "porno", "pornografia", "pornografico", "pornografica", "xxx", "masturbate"}
	patterns := make([]string, 0, len(words))
	for _, word := range words {
		letters := strings.Split(word, "")
		patterns = append(patterns, strings.Join(letters, `[\s\W_]*`))
	}
	return regexp.MustCompile(`\b(?:` + strings.Join(patterns, "|") + `)\b`)
}

func CommunityTextAllowed(body string) bool {
	text := strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, NormalizeText(body))
	text = strings.NewReplacer("0", "o", "1", "i", "3", "e", "4", "a", "5", "s", "7", "t", "@", "a", "$", "s").Replace(text)
	return !communityLanguage.MatchString(text)
}

func ValidateCommunityText(body string) (string, error) {
	body = strings.TrimSpace(body)
	if utf8.RuneCountInString(body) < 2 || utf8.RuneCountInString(body) > 1000 || strings.ContainsRune(body, '\x00') {
		return "", ErrInvalid
	}
	if !CommunityTextAllowed(body) {
		return "", ErrContentRejected
	}
	return body, nil
}
