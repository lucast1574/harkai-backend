package httpapi

import (
	_ "embed"
	"encoding/json"
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
)

//go:embed openapi.json
var specification []byte

func (a *API) specification(_ *http.Request, _ core.User) (Result, error) {
	return Result{Body: json.RawMessage(specification)}, nil
}
