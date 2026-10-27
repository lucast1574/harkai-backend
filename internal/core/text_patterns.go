package core

import "regexp"

type textRule struct {
	kind    string
	pattern *regexp.Regexp
}

// Prefer explicit event words and verb forms over broad stems (e.g. robot != robo).
var textRules = []textRule{
	{"extortion", regexp.MustCompile(`\b(extor[sc]ion[a-z]*|chantaj[a-z]*|cobr(o|os|ando|an|aron|ar|aban) (de |un |el |los )?cupos?|pid(en|iendo|ieron) (un |el |los )?cupos?|exig(en|iendo|ieron) (un |el |los )?cupos?|pag(o|os|ando|ar) (de |un |el |los )?cupos?|amenaz[a-z]* (de muerte|por telefono|por whatsapp))\b`)},
	{"kidnapping", regexp.MustCompile(`\b(secuestr[a-z]*|rapt(ar|aron|ando|ado|ada|o)|privad[oa] de (su |la )?libertad|se (lo|la|los|las) llevaron a la fuerza|retenid[oa] contra su voluntad)\b`)},
	{"homicide", regexp.MustCompile(`\b(asesin(ato|atos|aron|ado|ada|ados|adas|ando|o|a)|homicidi[a-z]*|sicari[a-z]*|persona (asesinada|muerta)|cadaver(es)?|mataron (a |al |a la )?(un |una )?(persona|vecino|vecina|senor|senora|hombre|mujer))\b`)},
	{"fire", regexp.MustCompile(`\b(incendi[a-z]*|fuego|llamas|quemand(o|ose)|se (quema|quemo|esta quemando)|humo (negro|denso)|explosion(es)?)\b`)},
	{"crash", regexp.MustCompile(`\b(cho(que|ques|co|caron|cando)|colisi[a-z]*|atropell[a-z]*|accidente (vial|de transito|de auto|de moto|vehicular)|volcadur[a-z]*|volc(o|aron|ado) (un |una )?(auto|carro|camion|bus|moto))\b`)},
	{"theft", regexp.MustCompile(`\b(rob(o|os|a|an|ar|aron|ando|ado|ada|ados|adas|aba|aban)|asalt(o|os|an|aron|ando|ado|ada)|hurt(o|os|aron|ando)|arrebataron|arrancharon|carterist[a-z]*|me (robo|robaron|quitaron) (el|la|mi))\b`)},
	{"violence", regexp.MustCompile(`\b(agresi[a-z]*|agred[a-z]*|golpe(ando|aron|a|an)|violencia|dispar[a-z]*|balacera|tiroteo|violacion|apunalaron|apunaland[o]|pelea|peleando|maltrat(o|ando|an)|amenaz[a-z]*|abrieron fuego|abrio fuego|intercambio de disparos)\b`)},
	{"emergency", regexp.MustCompile(`\b(desmay[a-z]*|no (respira|puede respirar|reacciona)|herid[oa]s?|sangrand[o]|convulsion[a-z]*|emergencia medica|ambulancia|inconsciente|paro cardiaco|ataque (al corazon|cardiaco)|se (esta )?ahogando)\b`)},
	{"pet", regexp.MustCompile(`\b(mascotas?|perr[oa]s?|gat[oa]s?|cachorr[oa]s?)\b.*\b(perdid[a-z]*|perdio|perdi|extraviad[a-z]*|encontrad[a-z]*|encontre|encontramos|desaparecid[a-z]*)\b|\b(perdid[a-z]*|perdio|perdi|extraviad[a-z]*|encontrad[a-z]*|encontre|encontramos|desaparecid[a-z]*)\b.*\b(mascotas?|perr[oa]s?|gat[oa]s?|cachorr[oa]s?)\b`)},
	{"place", regexp.MustCompile(`\b(lugar de ayuda|puesto de auxilio|refugio|centro de apoyo|registrar (un |una )?(lugar|negocio))\b`)},
}
