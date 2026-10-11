package core

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCommunityLanguageBoundariesAndEvasion(t *testing.T) {
	for _, body := range []string{"Eres una puta", "P.U.T.4", "huevón", "h u e v o n", "p\u200buta", "pornografía", "FUCK", "concha de tu madre", "cojudo", "puta\u0301"} {
		if CommunityTextAllowed(body) {
			t.Errorf("accepted disallowed wording %q", body)
		}
	}
	for _, body := range []string{"La computadora no funciona", "El diputado llegó al distrito", "Necesito ayuda por violencia sexual", "Reportaron una violación", "Mi madre necesita una ambulancia", "Una pregunta para el municipio"} {
		if !CommunityTextAllowed(body) {
			t.Errorf("false positive %q", body)
		}
	}
}
func TestAnonymousCommentsIncludingLegacy(t *testing.T) {
	user := User{ID: "private-user", Name: "Nombre privado", Role: "user"}
	c, err := BuildComment("Hay información útil", "507f1f77bcf86cd799439011", "507f1f77bcf86cd799439012", user, time.Now())
	if err != nil || c.AuthorName != "Vecino anónimo" {
		t.Fatal(c, err)
	}
	c.AuthorName = user.Name
	c = PublicComment(c, user)
	data, _ := json.Marshal(c)
	if !c.CanDelete || strings.Contains(string(data), user.ID) || strings.Contains(string(data), user.Name) {
		t.Fatal(string(data))
	}
	c.Body = "Eres un cojudo"
	if PublicComment(c, user).Body == c.Body {
		t.Fatal("legacy language leaked")
	}
}
