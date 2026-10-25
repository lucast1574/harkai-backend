package httpapi_test

import (
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/core"
	"strings"
	"testing"
)

func TestForumDistrictCityPrivacyModerationAndOwnership(t *testing.T) {
	api := fixture(t)
	author := registered(t, api, "forum-writer@example.test")
	other := registered(t, api, "forum-other@example.test")
	request(t, api, "POST", "/v1/forum", "", map[string]string{"body": "Pregunta vecinal", "district_id": "150130"}, 401)
	request(t, api, "POST", "/v1/forum", author.AccessToken, map[string]string{"body": "puta", "district_id": "150130"}, 400)
	request(t, api, "POST", "/v1/forum", author.AccessToken, map[string]string{"body": "Pregunta vecinal", "district_id": "999999"}, 400)
	w := request(t, api, "POST", "/v1/forum", author.AccessToken, map[string]string{"body": "¿Cómo podemos mejorar el parque?", "district_id": "150130"}, 201)
	var thread core.ForumThread
	_ = json.Unmarshal(w.Body.Bytes(), &thread)
	if !thread.CanDelete || thread.District != "San Borja" || strings.Contains(w.Body.String(), author.User.ID) || strings.Contains(w.Body.String(), author.User.Name) || strings.Contains(w.Body.String(), author.User.Email) {
		t.Fatal("author identity exposed", w.Body.String())
	}
	request(t, api, "POST", "/v1/forum", other.AccessToken, map[string]string{"body": "Consulta desde otro distrito", "district_id": "150122"}, 201)
	request(t, api, "POST", "/v1/forum", other.AccessToken, map[string]string{"body": "Consulta en Trujillo", "district_id": "130101"}, 201)
	w = request(t, api, "GET", "/v1/forum?ubigeo=150130", "", nil, 200)
	if !strings.Contains(w.Body.String(), thread.ID) || strings.Contains(w.Body.String(), "otro distrito") {
		t.Fatal("district isolation", w.Body.String())
	}
	w = request(t, api, "GET", "/v1/forum?ubigeo=150130&scope=city", "", nil, 200)
	if !strings.Contains(w.Body.String(), "otro distrito") || strings.Contains(w.Body.String(), "Consulta en Trujillo") {
		t.Fatal("province isolation", w.Body.String())
	}
	request(t, api, "POST", "/v1/forum", other.AccessToken, map[string]string{"body": "Consulta desde Huaral", "district_id": "150601"}, 201)
	request(t, api, "POST", "/v1/forum", other.AccessToken, map[string]string{"body": "Consulta desde Callao", "district_id": "070101"}, 201)
	for _, scope := range []string{"province", "department", "city"} {
		w = request(t, api, "GET", "/v1/forum?ubigeo=150130&scope="+scope, "", nil, 200)
		hasHuaral := strings.Contains(w.Body.String(), "Consulta desde Huaral")
		hasCallao := strings.Contains(w.Body.String(), "Consulta desde Callao")
		if hasHuaral != (scope == "department") || hasCallao != (scope == "city") || strings.Contains(w.Body.String(), "Consulta en Trujillo") {
			t.Fatalf("scope %s isolation: %s", scope, w.Body.String())
		}
	}
	request(t, api, "GET", "/v1/forum?ubigeo=150130&scope=all", "", nil, 400)
	request(t, api, "GET", "/v1/forum?ubigeo=150130&cursor=invalid", "", nil, 400)
	path := "/v1/forum/" + thread.ID + "/comments"
	request(t, api, "POST", path, other.AccessToken, map[string]string{"body": "p.u.t.a"}, 400)
	w = request(t, api, "POST", path, other.AccessToken, map[string]string{"body": "Podemos organizar una reunión vecinal"}, 201)
	var comment core.Comment
	_ = json.Unmarshal(w.Body.Bytes(), &comment)
	if comment.AuthorName != "Vecino anónimo" || !comment.CanDelete || strings.Contains(w.Body.String(), other.User.Name) {
		t.Fatal("reply privacy")
	}
	request(t, api, "DELETE", path+"/"+comment.ID, author.AccessToken, nil, 404)
	request(t, api, "DELETE", path+"/"+comment.ID, other.AccessToken, nil, 204)
	request(t, api, "DELETE", "/v1/forum/"+thread.ID, other.AccessToken, nil, 404)
	request(t, api, "DELETE", "/v1/forum/"+thread.ID, author.AccessToken, nil, 204)
	request(t, api, "GET", path, "", nil, 404)
	w = request(t, api, "GET", "/v1/forum?ubigeo=150130", "", nil, 200)
	if strings.Contains(w.Body.String(), thread.ID) {
		t.Fatal("deleted thread remains public")
	}
}
