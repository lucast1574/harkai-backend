package mongostore

import (
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/districtgeo"
	"math"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var supportPhone = regexp.MustCompile(`^\+?[0-9]{3,15}$`)

func supportSource(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && strings.HasSuffix(u.Hostname(), ".gob.pe")
}
func ValidateSupportDistrict(d core.SupportDistrict) error {
	if d.Country != "PE" || (d.City != "lima" && d.City != "trujillo" && !districtgeo.ValidProvince(d.City)) || strings.TrimSpace(d.District) == "" || len(d.Contacts) < 1 {
		return fmt.Errorf("invalid support district")
	}
	for _, c := range d.Contacts {
		if _, err := time.Parse("2006-01-02", c.ReviewedAt); err != nil {
			return err
		}
		if c.Label == "" || !supportPhone.MatchString(c.Phone) || !supportSource(c.SourceURL) || (c.Scope != "nacional" && c.Scope != "distrital") {
			return fmt.Errorf("invalid support contact")
		}
	}
	return nil
}
func ValidateSupportPlace(p core.SupportPlace) error {
	if p.ID == "" || p.Name == "" || p.Address == "" || p.District == "" || !supportSource(p.SourceURL) || (p.City != "lima" && p.City != "trujillo" && !districtgeo.ValidProvince(p.City)) {
		return fmt.Errorf("invalid support place")
	}
	if math.IsNaN(p.Latitude) || math.IsNaN(p.Longitude) || p.Latitude < -19 || p.Latitude > 0 || p.Longitude < -82 || p.Longitude > -68 {
		return fmt.Errorf("invalid Peru coordinates")
	}
	location, err := url.Parse(p.LocationSourceURL)
	if err != nil || location.Scheme != "https" || location.Hostname() == "" || location.User != nil {
		return fmt.Errorf("invalid location source")
	}
	if p.Phone != "" && !supportPhone.MatchString(p.Phone) {
		return fmt.Errorf("invalid support phone")
	}
	for _, phone := range p.Phones {
		if !supportPhone.MatchString(phone) {
			return fmt.Errorf("invalid support phone")
		}
	}
	for _, raw := range p.PhotoURLs {
		photo, err := url.Parse(raw)
		if err != nil || photo.User != nil || photo.Hostname() == "" {
			return fmt.Errorf("invalid support photo")
		}
		legacy := photo.Scheme == "http" && photo.Host == "app20.susalud.gob.pe:8080" && photo.Path == "/registro-renipress-webapp/ipress.htm" && photo.Query().Get("action") == "downloadFile" && photo.Query().Get("idArchivo") != ""
		if photo.Scheme != "https" && !legacy {
			return fmt.Errorf("invalid support photo")
		}
	}
	if p.Kind != "hospital" && p.Kind != "health_center" && p.Kind != "support" {
		return fmt.Errorf("invalid place kind")
	}
	_, err = time.Parse("2006-01-02", p.ReviewedAt)
	return err
}
