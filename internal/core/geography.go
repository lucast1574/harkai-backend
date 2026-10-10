package core

type DistrictGeometry struct {
	Type        string        `json:"type" bson:"type"`
	Coordinates [][][]float64 `json:"coordinates" bson:"coordinates"`
}
type GeographicDistrict struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Province      string           `json:"province"`
	Department    string           `json:"department"`
	Geometry      DistrictGeometry `json:"geometry"`
	SourceURL     string           `json:"source_url"`
	ReferenceYear int              `json:"reference_year"`
}

// AreaGeometry supports a district or the collection of districts in a province.
type AreaGeometry struct {
	Type        string `json:"type" bson:"type"`
	Coordinates any    `json:"coordinates" bson:"coordinates"`
}
type GeographicArea struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Province      string       `json:"province"`
	Department    string       `json:"department"`
	Scope         string       `json:"scope"`
	Geometry      AreaGeometry `json:"geometry"`
	SourceURL     string       `json:"source_url"`
	ReferenceYear int          `json:"reference_year"`
}
