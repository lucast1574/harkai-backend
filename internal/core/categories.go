package core

type Category struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Group    string   `json:"group"`
	Critical bool     `json:"critical"`
	Advice   []string `json:"advice"`
}

func Categories() []Category {
	return []Category{
		{"fire", "Incendio", "emergency", true, []string{"Aléjate del humo y las llamas. No ingreses al lugar.", "Contacta a los bomberos desde un lugar seguro."}},
		{"crash", "Accidente vial", "emergency", true, []string{"Mantente fuera de la vía y evita nuevos accidentes.", "Solicita ayuda médica. No muevas a personas heridas salvo peligro inmediato."}},
		{"emergency", "Emergencia médica", "emergency", true, []string{"Contacta a un servicio de emergencias y sigue sus indicaciones.", "Describe el lugar y el estado de la persona sin publicar su identidad."}},
		{"theft", "Robo", "security", true, []string{"Busca un lugar seguro y evita confrontar a los responsables.", "Contacta a la policía. No publiques nombres ni acusaciones personales."}},
		{"extortion", "Extorsión o amenazas", "security", true, []string{"No confrontes ni acuerdes encuentros con quien amenaza.", "Conserva mensajes y pruebas en privado. Contacta a la policía desde un lugar seguro.", "No publiques teléfonos, direcciones, placas ni identidades de las víctimas."}},
		{"kidnapping", "Posible secuestro", "security", true, []string{"Contacta de inmediato a la policía desde un lugar seguro.", "No intentes intervenir ni difundas detalles que puedan poner a la víctima en riesgo."}},
		{"violence", "Agresión o violencia", "security", true, []string{"Aléjate y busca ayuda; evita intervenir si te expone al peligro.", "No publiques imágenes que identifiquen a víctimas ni personas vulnerables."}},
		{"homicide", "Hecho violento grave", "security", true, []string{"Aléjate del lugar y contacta a la policía o emergencias.", "No te acerques, alteres la escena ni publiques imágenes de víctimas."}},
		{"public_drinking", "Consumo de alcohol en vía pública", "security", false, []string{"Describe lo que observaste y la ubicación, sin identificar ni acusar a personas.", "Evita confrontaciones. Si hay una situación de riesgo, busca ayuda desde un lugar seguro."}},
		{"public_drug_use", "Posible consumo de drogas en vía pública", "security", false, []string{"Describe la conducta que observaste y el lugar, sin asumir la identidad ni la condición de las personas.", "Mantén distancia y evita confrontaciones."}},
		{"public_disorder", "Alteración del orden público", "security", false, []string{"Describe los hechos y cómo afectan al espacio público, sin publicar identidades.", "Evita confrontaciones y busca un lugar seguro si la situación escala."}},
		{"pet", "Mascota perdida o encontrada", "community", false, []string{"Describe rasgos, zona y cuándo ocurrió.", "Comparte un contacto solo si aceptas hacerlo público."}},
		{"place", "Lugar de ayuda", "community", false, []string{"Describe qué ayuda ofrece el lugar y verifica su ubicación.", "No lo presentes como un servicio oficial si no lo es."}},
	}
}

func CategoryFor(id string) (Category, bool) {
	for _, category := range Categories() {
		if category.ID == id {
			return category, true
		}
	}
	return Category{}, false
}
