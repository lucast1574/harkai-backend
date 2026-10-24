package core

import (
	"fmt"
	"math"
	"time"
)

// App-owned policy: owners and visitors cannot choose notification distances.
func ProximityBands() []int { return []int{500, 250, 50, 15} }
func ProximityStage(distance, accuracy float64) int {
	if math.IsNaN(distance) || math.IsInf(distance, 0) || distance < 0 || math.IsNaN(accuracy) || math.IsInf(accuracy, 0) || accuracy < 0 || accuracy > 100 {
		return 0
	}
	stage := 0
	for _, band := range ProximityBands() {
		if distance+accuracy <= float64(band) {
			stage = band
		}
	}
	return stage
}
func MapIncidentActive(i Incident, now time.Time) bool {
	return i.Status == "active" && !i.CreatedAt.After(now) && now.Sub(i.CreatedAt) < 3*time.Hour && (i.ExpiresAt == nil || i.ExpiresAt.After(now))
}
func ProximityCopy(kind, name, offer, code string, stage int) (string, string) {
	if kind == "business" {
		title := "Un negocio cerca de ti"
		if stage <= 50 {
			title = "Estás muy cerca de " + name
		}
		if stage <= 15 {
			title = "Llegaste cerca de " + name
		}
		body := fmt.Sprintf("%s · a menos de %d m. Promocionado.", name, stage)
		if offer != "" {
			body += " " + offer
		}
		if code != "" {
			body += " Código: " + code
		}
		return title, body
	}
	title := fmt.Sprintf("Un reporte a menos de %d m", stage)
	prefix := "Información vecinal sin verificar. "
	if stage <= 50 {
		title = "Un reporte muy cerca de ti"
	}
	if stage <= 15 {
		title = "Atención a tu entorno"
	}
	switch kind {
	case "theft", "violence", "extortion", "kidnapping", "homicide":
		if stage <= 50 {
			return title, prefix + "Se reportó un incidente de seguridad muy cerca. Evita acercarte y busca un lugar seguro."
		}
		return title, prefix + "Se reportó un incidente de seguridad en tu zona. Mantente atento y evita el área."
	case "fire":
		return title, prefix + "Hay un incendio reportado cerca. Mantente lejos del humo y deja paso a emergencias."
	case "emergency":
		return title, prefix + "Hay una emergencia reportada cerca. Deja libre el acceso a los equipos de ayuda."
	case "crash":
		return title, prefix + "Hay un accidente reportado cerca. Reduce la velocidad y evita acercarte."
	case "pet":
		return "Una mascota reportada cerca", prefix + "Hay una mascota perdida o encontrada cerca de tu ubicación. Revisa la publicación si puedes ayudar."
	default:
		return title, prefix + "Revisa la información compartida por los vecinos y su estado actual."
	}
}
