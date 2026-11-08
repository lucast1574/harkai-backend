package core

import (
	"regexp"
)

// Only unambiguous common misspellings are normalized. General edit-distance
// correction can turn valid words such as "cobrando" into an invented robbery.
var reportWord = regexp.MustCompile(`[a-z]+`)
var reportSpellings = map[string]string{
	"insendio": "incendio", "incendoo": "incendio",
	"axidente": "accidente", "accidnte": "accidente",
	"rrobando": "robando", "rrobaron": "robaron",
	"estorcion": "extorsion", "estorcionadores": "extorsionadores",
	"extorcionadores": "extorsionadores", "secuestrndo": "secuestrando",
	"golpiando": "golpeando", "cadaber": "cadaver",
	"desmallo": "desmayo", "veviendo": "bebiendo", "alcol": "alcohol",
	"mariuana": "marihuana", "parke": "parque", "parqe": "parque",
	"destrosos": "destrozos",
	"chela":     "cerveza", "chelas": "cerveza", "chupando": "bebiendo",
	"pisco": "licor", "ron": "licor", "trago": "licor",
	"celu": "celular", "chibolo": "menor", "chibola": "menor",
	"firulais": "perro", "felino": "gato", "can": "perro",
	"trompadas": "punetazos", "balazos": "disparos",
	"canchita": "cancha", "loza": "cancha", "losa": "cancha",
	"estupefacientes": "drogas", "jatear": "dormir",
}

// This is inference-only. Original descriptions are always stored unchanged.
func NormalizeReportText(text string) string {
	return reportWord.ReplaceAllStringFunc(NormalizeText(text), correctReportWord)
}

func correctReportWord(word string) string {
	if replacement, ok := reportSpellings[word]; ok {
		return replacement
	}
	return word
}
