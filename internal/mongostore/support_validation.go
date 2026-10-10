package mongostore

import (
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
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
	if d.Country != "PE" || (d.City != "lima" && d.City != "trujillo") || strings.TrimSpace(d.District) == "" || len(d.Contacts) < 1 {
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
	if p.ID == "" || p.Name == "" || p.Address == "" || p.District == "" || !supportSource(p.SourceURL) || (p.City != "lima" && p.City != "trujillo") {
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
	if p.Kind != "hospital" && p.Kind != "health_center" && p.Kind != "support" {
		return fmt.Errorf("invalid place kind")
	}
	_, err = time.Parse("2006-01-02", p.ReviewedAt)
	return err
}
