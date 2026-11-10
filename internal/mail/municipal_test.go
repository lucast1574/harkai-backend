package mail

import (
	"strings"
	"testing"
)

func TestMunicipalNotificationEscapesContactAndKeepsAction(t *testing.T) {
	content, err := Render("general", "municipal_inquiry", TemplateData{
		Name: "<script>alert(1)</script>", Title: "Acceso municipal · Prueba", ContactEmail: "contact@example.com",
		Message: "Municipalidad: prueba\n<img src=x onerror=alert(1)>", ActionURL: "https://panel.harkai.lat/dashboard/admin?section=municipal",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(content.HTML, "<script>") || strings.Contains(content.HTML, "<img src=x") {
		t.Fatal("unescaped inquiry in email")
	}
	for _, value := range []string{"contact@example.com", "Abrir solicitudes municipales", "section=municipal"} {
		if !strings.Contains(content.HTML, value) || !strings.Contains(content.Text, value) {
			t.Fatal("email missing context", value)
		}
	}
}
