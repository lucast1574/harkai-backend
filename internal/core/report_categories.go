package core

import "fmt"

type ReportCategory struct {
	ID            string     `json:"id"`
	Label         string     `json:"label"`
	Subcategories []Category `json:"subcategories"`
}

func CrimeTypes() []string {
	return []string{"extortion", "kidnapping", "homicide", "theft", "violence"}
}

func IsCrime(kind string) bool {
	for _, crime := range CrimeTypes() {
		if kind == crime {
			return true
		}
	}
	return false
}

func ReportCategories() []ReportCategory {
	groups := []ReportCategory{{ID: "crime", Label: "Crimen", Subcategories: []Category{}}}
	for _, category := range Categories() {
		if IsCrime(category.ID) {
			groups[0].Subcategories = append(groups[0].Subcategories, category)
		} else {
			groups = append(groups, ReportCategory{ID: category.ID, Label: category.Label, Subcategories: []Category{}})
		}
	}
	return groups
}

// The server selects a crime subtype; clients only select the report category.
func ResolveReportType(text, category string) (string, error) {
	analysis, err := AnalyzeText(text, category)
	if err != nil {
		return "", err
	}
	if analysis.SuggestedType == "" {
		return "", fmt.Errorf("%w: describe con más detalle qué ocurrió para identificar la categoría", ErrInvalid)
	}
	if category == "crime" && !IsCrime(analysis.SuggestedType) || category != "crime" && analysis.SuggestedType != category {
		return "", fmt.Errorf("%w: la descripción no corresponde a la categoría elegida", ErrInvalid)
	}
	return analysis.SuggestedType, nil
}

// Manual corrections remain inside the selected category and never verify a report.
func ManualReportType(kind, category string) (string, error) {
	if category == "crime" && IsCrime(kind) {
		return kind, nil
	}
	if category != "crime" && category == kind {
		if _, ok := CategoryFor(kind); ok {
			return kind, nil
		}
	}
	return "", fmt.Errorf("%w: manual category mismatch", ErrInvalid)
}
