package core

import "testing"

func TestRegistrationAndPreferences(t *testing.T) {
	input, err := ValidateRegistration(Credentials{Email: "  TEST@EXAMPLE.COM ", Name: "  Lucas  ", Password: "a-long-enough-password"})
	if err != nil || input.Email != "test@example.com" || input.Name != "Lucas" {
		t.Fatalf("normalization: %+v %v", input, err)
	}
	for _, email := range []string{"not-an-email", "John <john@example.com>", ""} {
		if _, err := NormalizeEmail(email); err == nil {
			t.Errorf("accepted %s", email)
		}
	}
	if _, err := ValidateRegistration(Credentials{Email: "a@example.com", Name: "Test", Password: "short"}); err == nil {
		t.Fatal("short password")
	}
	p := DefaultPreferences()
	p.Notifications = "none"
	if _, err := ValidatePreferences(p); err != nil {
		t.Fatal(err)
	}
	p.RadiusMeters = 999999
	if _, err := ValidatePreferences(p); err == nil {
		t.Fatal("invalid preference bounds")
	}
}
