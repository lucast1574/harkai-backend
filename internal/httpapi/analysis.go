package httpapi

import (
	"github.com/lucast1574/harkai-backend/internal/core"
	"net/http"
	"time"
)

func (a *API) analyzeText(r *http.Request, user core.User) (Result, error) {
	if err := a.Service.Limits.Allow(r.Context(), "analysis:"+user.ID, 60, time.Minute, a.Service.Now()); err != nil {
		return Result{}, err
	}
	var input struct {
		Text         string `json:"text"`
		SelectedType string `json:"selected_type,omitempty"`
	}
	if err := decode(r, &input); err != nil {
		return Result{}, err
	}
	analysis, err := core.AnalyzeText(input.Text, input.SelectedType)
	return Result{Body: analysis}, err
}
