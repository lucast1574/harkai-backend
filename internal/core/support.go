package core

// Support resources are directory records, never incidents or push jobs.
type SupportContact struct {
	Label      string `json:"label" bson:"label"`
	Phone      string `json:"phone" bson:"phone"`
	Scope      string `json:"scope" bson:"scope"`
	SourceURL  string `json:"source_url" bson:"source_url"`
	ReviewedAt string `json:"reviewed_at" bson:"reviewed_at"`
}
type SupportDistrict struct {
	DistrictID string           `json:"district_id,omitempty" bson:"district_id,omitempty"`
	Department string           `json:"department,omitempty" bson:"department,omitempty"`
	Province   string           `json:"province,omitempty" bson:"province,omitempty"`
	ID         string           `json:"-" bson:"_id"`
	Country    string           `json:"country" bson:"country"`
	City       string           `json:"city" bson:"city"`
	District   string           `json:"district" bson:"district"`
	Contacts   []SupportContact `json:"contacts" bson:"contacts"`
}
type SupportPlace struct {
	Department        string   `json:"department,omitempty" bson:"department,omitempty"`
	Province          string   `json:"province,omitempty" bson:"province,omitempty"`
	ID                string   `json:"id" bson:"_id"`
	RegistryID        string   `json:"registry_id,omitempty" bson:"registry_id,omitempty"`
	DistrictID        string   `json:"district_id,omitempty" bson:"district_id,omitempty"`
	Institution       string   `json:"institution,omitempty" bson:"institution,omitempty"`
	Classification    string   `json:"classification,omitempty" bson:"classification,omitempty"`
	Category          string   `json:"category,omitempty" bson:"category,omitempty"`
	SourceUpdatedAt   string   `json:"source_updated_at,omitempty" bson:"source_updated_at,omitempty"`
	Name              string   `json:"name" bson:"name"`
	Kind              string   `json:"kind" bson:"kind"`
	City              string   `json:"city" bson:"city"`
	District          string   `json:"district" bson:"district"`
	Address           string   `json:"address" bson:"address"`
	Latitude          float64  `json:"latitude" bson:"latitude"`
	Longitude         float64  `json:"longitude" bson:"longitude"`
	Location          Point    `json:"-" bson:"location"`
	Phone             string   `json:"phone,omitempty" bson:"phone,omitempty"`
	Phones            []string `json:"phones,omitempty" bson:"phones,omitempty"`
	PhotoURLs         []string `json:"photo_urls,omitempty" bson:"photo_urls,omitempty"`
	SourceURL         string   `json:"source_url" bson:"source_url"`
	LocationSourceURL string   `json:"location_source_url" bson:"location_source_url"`
	ReviewedAt        string   `json:"reviewed_at" bson:"reviewed_at"`
}
