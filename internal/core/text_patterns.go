package core

import "regexp"

type textRule struct {
	kind    string
	pattern *regexp.Regexp
}

// Prefer explicit event words and verb forms over broad stems (e.g. robot != robo).
var textRules = []textRule{
	{"extortion", regexp.MustCompile(`\b(extor[sc]ion[a-z]*|cobran seguridad a la fuerza|pid(en|e) (plata|dinero) para no (hacer|matar)|exig(en|e) (dinero|un pago)[^.!?;\n]{0,50}(fotos|difundir|publicar)|exig(en|e)[^.!?;\n]{0,40}soles para no divulgar|paga o vamos por|chantaj[a-z]*|cobr(o|os|ando|an|aron|ar|aban) (de |un |el |los )?cupos?|pid(en|iendo|ieron) (un |el |los )?cupos?|exig(en|iendo|ieron) (un |el |los )?cupos?|pag(o|os|ando|ar) (de |un |el |los )?cupos?|amenaz[a-z]* (de muerte|por telefono|por whatsapp))\b`)},
	{"kidnapping", regexp.MustCompile(`\b(secuestr[a-z]*|rapt(ar|aron|ando|ado|ada|o)|privad[oa] de (su |la )?libertad|se (lo|la|los|las) llevaron a la fuerza|(se llevaron|subieron|metieron|arrastraron)[^.!?;\n]{0,60}(a la fuerza|por la fuerza|contra su voluntad)|retenid[oa] contra su voluntad)\b`)},
	{"homicide", regexp.MustCompile(`\b(asesin(ato|atos|aron|ado|ada|ados|adas|ando|o|a)|restos humanos|sin vida|quitaron la vida|ultimaron|balear y matar|homicidi[a-z]*|sicari[a-z]*|persona (asesinada|muerta)|cadaver(es)?|mataron (a |al |a la )?(un |una )?(persona|vecino|vecina|senor|senora|hombre|mujer))\b`)},
	{"fire", regexp.MustCompile(`\b(incendi[a-z]*|fuego|llamas|quemand(o|ose)|ardiendo|agarro candela|se (ta |esta )?prendiendo|se (quema|quemo|esta quemando)|humo (negro|denso)|explosion(es)?)\b`)},
	{"crash", regexp.MustCompile(`\b(cho(que|ques|co|caron|cando)|impact(o|aron)|arroll(o|aron)|accident(o|aron)|se dieron de frente|perdio los frenos|colisi[a-z]*|atropell[a-z]*|accidente (vial|de transito|de auto|de moto|vehicular)|volcadur[a-z]*|volc(o|aron|ado) (un |una )?(auto|carro|camion|bus|moto))\b`)},
	{"theft", regexp.MustCompile(`\b(rob(o|os|a|an|ar|aron|ando|ado|ada|ados|adas|aba|aban)|cogote[a-z]*|vaciaron la caja|forzaron la puerta[^.!?;\n]{0,60}se llevaron|me jalaron el celular|se (llevaron|quedaron con)[^.!?;\n]{0,40}(mochila|bicicleta|billetera|celular|tele|computadora)|asalt(o|os|an|aron|ando|ado|ada)|hurt(o|os|aron|ando)|arrebataron|arrancharon|carterist[a-z]*|me (robo|robaron|quitaron) (el|la|mi))\b`)},
	{"violence", regexp.MustCompile(`\b(agresi[a-z]*|patadas|punetazos|pateando|cachetada|se estan atacando|hay[^.!?;\n]{0,30}atacandose|le estan sacando la mugre|agred[a-z]*|golpe(ando|aron|a|an)|violencia|dispar[a-z]*|balacera|tiroteo|violacion|apunalaron|apunaland[o]|pelea|peleando|maltrat(o|ando|an)|amenaz[a-z]*|abrieron fuego|abrio fuego|intercambio de disparos)\b`)},
	{"emergency", regexp.MustCompile(`\b(desmay[a-z]*|no despierta|se desplomo|perdio el conocimiento|desvanecido|asfixiando|espuma por la boca|dificultades para respirar|no puede tomar aire|no (respira|puede respirar|reacciona)|herid[oa]s?|sangrand[o]|convulsion[a-z]*|emergencia medica|ambulancia|inconsciente|paro cardiaco|ataque (al corazon|cardiaco)|se (esta )?ahogando)\b`)},
	{"public_drinking", regexp.MustCompile(`\b((tomando|bebiendo|beber|consumiendo|consumo de|empinando botellas de) (alcohol|cerveza|licor|trago[s]?|bebidas alcoholicas)[^.!?;\n]{0,60}\b(calle|via publica|parque|vereda|plaza|avenida|cancha|paradero|puente peatonal|callejon)|((tomando|bebiendo) en (la calle|la via publica|el parque|la vereda|la plaza)))\b`)},
	{"public_drug_use", regexp.MustCompile(`\b(fumando (marihuana|pasta basica|(un )?porros?)|metiendo pasta|jalando (coca|cocaina)|se drogan|cigarro de marihuana|consumiendo (drogas|marihuana|cocaina|pasta basica)|consumo de (drogas|marihuana|cocaina|pasta basica)|drogandose|fumon(es)?)[^.!?;\n]{0,60}\b(calle|via publica|parque|vereda|plaza|avenida|cancha|paradero|puente peatonal|callejon)\b`)},
	{"public_disorder", regexp.MustCompile(`\b(alteracion del orden publico|destrozos|rompiendo las bancas|destruyendo las senales|rompiendo los postes|destrozando los juegos|destrozando las senales|arrancaron los tachos|parlantes? a full|no (dejan|podemos|pueden) dormir|disturbios|vandalismo|vandalizando|destrozando (bancas|bancos|senales|mobiliario publico)|ruidos molestos|musica a todo volumen)\b`)},
	{"pet", regexp.MustCompile(`\b(mascotas?|perr[oa]s?|gat[oa]s?|cachorr[oa]s?|conej[oa]s?)\b.*\b(perdid[a-z]*|perdio|perdi|extraviad[a-z]*|encontrad[a-z]*|encontre|encontramos|desaparecid[a-z]*|escapo|no aparece|no vuelve|no regresa|busco|buscandola|recogi|halle)\b|\b(perdid[a-z]*|perdio|perdi|extraviad[a-z]*|encontrad[a-z]*|encontre|encontramos|desaparecid[a-z]*|escapo|no aparece|no vuelve|no regresa|busco|buscandola|recogi|halle)\b.*\b(mascotas?|perr[oa]s?|gat[oa]s?|cachorr[oa]s?|conej[oa]s?)\b`)},
	{"place", regexp.MustCompile(`\b(lugar de ayuda|puesto de auxilio|refugio|centro de apoyo|registrar (un |una )?(lugar|negocio))\b`)},
}
