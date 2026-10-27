package core

import (
	"strings"
	"testing"
)

func TestConversationalReports(t *testing.T) {
	cases := []struct{ text, want string }{
		{"Aquí están extorsionando", "extortion"},
		{"Bueno estoy en la esquina de mi casa y no se que hacer aqui estan extorsionando al vecino y le estan pidiendo plata", "extortion"},
		{"No puedo salir porque estan extorsionando a los comerciantes", "extortion"},
		{"Aqui estan extorcionando al de la bodega", "extortion"},
		{"Les cobran los cupos a las tiendas", "extortion"},
		{"Estan pidiendo cupos a los vecinos", "extortion"},
		{"La estan chantajeando", "extortion"},
		{"Se llevaron a la fuerza a una mujer, la secuestraron", "kidnapping"},
		{"Se la llevaron a la fuerza", "kidnapping"},
		{"Lo tienen retenido contra su voluntad", "kidnapping"},
		{"Asesinaron a un vecino de la cuadra", "homicide"},
		{"Hay dos cadaveres en la esquina", "homicide"},
		{"Mataron a una persona afuera del mercado", "homicide"},
		{"No puedo acercarme porque se esta quemando la casa", "fire"},
		{"Se quemo una tienda", "fire"},
		{"Las llamas salen de la ventana", "fire"},
		{"No se que paso pero hay humo negro", "fire"},
		{"Un carro choco contra un poste", "crash"},
		{"Chocaron dos motos", "crash"},
		{"Se volco un camion", "crash"},
		{"Hubo un accidente vehicular en la avenida", "crash"},
		{"Acaban de arrancharon mi cartera", "theft"},
		{"Un carterista me quito el celular", "theft"},
		{"Me robo el telefono", "theft"},
		{"Se estan peleando en la calle", "violence"},
		{"Estan golpeando a una persona", "violence"},
		{"Hay disparos de arma de fuego", "violence"},
		{"Abrieron fuego en la avenida", "violence"},
		{"Esta inconsciente en la vereda", "emergency"},
		{"No reacciona, manden ayuda", "emergency"},
		{"Se esta ahogando una persona", "emergency"},
		{"Encontre un cachorro en la calle", "pet"},
		{"Mi gata esta desaparecida desde ayer", "pet"},
		{"Se perdio mi mascota", "pet"},
		{"Estoy probando un robot en la calle", ""},
		{"Me llamas cuando llegues a casa", ""},
		{"No hay incendio en la calle", ""},
		{"No estan extorsionando a nadie", ""},
		{"No ha sido reportado ningun robo", ""},
		{"Sin personas heridas en la zona", ""},
		{"No hubo robo pero aqui estan extorsionando", "extortion"},
		{"No es incendio sino un choque de carros", "crash"},
	}
	for _, tc := range cases {
		t.Run(tc.text, func(t *testing.T) {
			got, err := AnalyzeText(tc.text, "")
			if err != nil || got.SuggestedType != tc.want || !got.RequiresConfirmation {
				t.Fatalf("got %+v, %v; want %q", got, err, tc.want)
			}
		})
	}
}

func TestLongCrimeDescriptionStillRecognizesExtortion(t *testing.T) {
	text := strings.Repeat("estoy en la avenida cerca del mercado y quiero contar lo que vi ", 20) + "no se que hacer porque aqui estan extorsionando a la tienda"
	got, err := ResolveReportType(text, "crime")
	if err != nil || got != "extortion" {
		t.Fatalf("got %q, %v", got, err)
	}
}
