package core

import (
	"fmt"
	"regexp"
	"strings"
)

type EmergencyContacts struct {
	ID      string            `json:"-" bson:"_id"`
	Country string            `json:"country" bson:"country"`
	City    string            `json:"city" bson:"city"`
	Numbers map[string]string `json:"numbers" bson:"numbers"`
}

func Region(country, city string) (string, string, error) {
	country = strings.ToUpper(strings.TrimSpace(country))
	city = strings.ToLower(strings.TrimSpace(city))
	if len(country) != 2 || country[0] < 'A' || country[0] > 'Z' || country[1] < 'A' || country[1] > 'Z' || len(city) < 2 || len(city) > 100 {
		return "", "", fmt.Errorf("%w: country ISO code and city required", ErrInvalid)
	}
	return country, city, nil
}

func ValidateContacts(c EmergencyContacts) (EmergencyContacts, error) {
	country, city, err := Region(c.Country, c.City)
	if err != nil {
		return c, err
	}
	if len(c.Numbers) < 1 || len(c.Numbers) > 4 {
		return c, fmt.Errorf("%w: emergency numbers required", ErrInvalid)
	}
	valid := regexp.MustCompile(`^\+?[0-9 ()-]{3,25}$`)
	for key, value := range c.Numbers {
		if key != "police" && key != "firefighters" && key != "medical" && key != "municipal" {
			return c, fmt.Errorf("%w: unknown emergency service", ErrInvalid)
		}
		if !valid.MatchString(value) || len(strings.Trim(value, "+ ()-")) < 3 {
			return c, fmt.Errorf("%w: invalid phone number", ErrInvalid)
		}
	}
	c.ID = country + ":" + city
	c.Country = country
	c.City = city
	return c, nil
}
