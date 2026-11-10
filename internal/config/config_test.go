package config

import "testing"

func TestStripeConfigurationFailsClosedWhenIncomplete(t *testing.T) {
	env := map[string]string{"MONGODB_URI": "mongodb://localhost", "S3_ENDPOINT": "localhost:9000", "S3_ACCESS_KEY": "test", "S3_SECRET_KEY": "testsecret", "STRIPE_SECRET_KEY": "sk_test_example"}
	get := func(k string) string { return env[k] }
	if _, err := Parse(get); err == nil {
		t.Fatal("partial credentials silently enabled free registrations")
	}
	env["STRIPE_WEBHOOK_SECRET"] = "whsec_test"
	env["STRIPE_PRICE_ID"] = "price_test"
	env["STRIPE_SUCCESS_URL"] = "https://panel.harkai.lat/dashboard/my-businesses?payment=processing"
	env["STRIPE_CANCEL_URL"] = "https://panel.harkai.lat/dashboard/my-businesses?payment=canceled"
	if _, err := Parse(get); err != nil {
		t.Fatal(err)
	}
	env["STRIPE_SUCCESS_URL"] = "https://attacker.example"
	if _, err := Parse(get); err == nil {
		t.Fatal("foreign checkout return origin accepted")
	}
}

func TestConfigurationValidation(t *testing.T) {
	env := map[string]string{"MONGODB_URI": "mongodb://localhost", "S3_ENDPOINT": "localhost:9000", "S3_ACCESS_KEY": "test", "S3_SECRET_KEY": "testsecret"}
	get := func(k string) string { return env[k] }
	if _, err := Parse(get); err != nil {
		t.Fatal(err)
	}
	env["CORS_ORIGINS"] = "*"
	if _, err := Parse(get); err == nil {
		t.Fatal("wildcard origin accepted")
	}
	env["CORS_ORIGINS"] = "https://harkai.lat/path"
	if _, err := Parse(get); err == nil {
		t.Fatal("non-origin accepted")
	}
	env["CORS_ORIGINS"] = "https://harkai.lat"
	env["PORT"] = "-1"
	if _, err := Parse(get); err == nil {
		t.Fatal("invalid port accepted")
	}
}
