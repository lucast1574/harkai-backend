package config

import "testing"

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
