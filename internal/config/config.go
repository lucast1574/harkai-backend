package config

import (
	"errors"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	StripeSecret        string
	StripeWebhookSecret string
	StripePrice         string
	StripeSuccessURL    string
	StripeCancelURL     string
	MunicipalContactTo  string
	ResendAPIKey        string
	MailFrom            string
	FCMKeyBase64        string
	GoogleProject       string
	MediaWorkerURL      string
	MediaWorkerKey      string
	Port                string
	MongoURI            string
	Database            string
	Origins             []string
	S3Endpoint          string
	S3AccessKey         string
	S3SecretKey         string
	S3Bucket            string
	S3Secure            bool
}

func Load() (Config, error) { return Parse(os.Getenv) }

func Parse(get func(string) string) (Config, error) {
	c := Config{Port: value(get, "PORT", "8080"), MongoURI: get("MONGODB_URI"), Database: value(get, "MONGODB_DATABASE", "harkai")}
	c.StripeSecret, c.StripeWebhookSecret, c.StripePrice = get("STRIPE_SECRET_KEY"), get("STRIPE_WEBHOOK_SECRET"), get("STRIPE_PRICE_ID")
	c.StripeSuccessURL, c.StripeCancelURL = get("STRIPE_SUCCESS_URL"), get("STRIPE_CANCEL_URL")
	if c.StripeWebhookSecret != "" || c.StripeSecret != "" || c.StripePrice != "" {
		if c.StripeSecret == "" || c.StripePrice == "" || c.StripeWebhookSecret == "" {
			return c, errors.New("incomplete Stripe configuration")
		}
		for _, raw := range []string{c.StripeSuccessURL, c.StripeCancelURL} {
			u, err := url.Parse(raw)
			if err != nil || u.Scheme != "https" || u.Host != "panel.harkai.lat" || u.User != nil {
				return c, errors.New("invalid Stripe return URL")
			}
		}
	}
	c.ResendAPIKey = get("RESEND_API_KEY")
	c.MailFrom = value(get, "MAIL_FROM", "Harkai <hola@harkai.lat>")
	c.MunicipalContactTo = strings.TrimSpace(get("MUNICIPAL_CONTACT_EMAIL"))
	if c.MunicipalContactTo != "" {
		address, err := mail.ParseAddress(c.MunicipalContactTo)
		if err != nil || address.Address != c.MunicipalContactTo {
			return c, errors.New("invalid MUNICIPAL_CONTACT_EMAIL")
		}
	}
	c.GoogleProject = get("GOOGLE_PROJECT_ID")
	c.FCMKeyBase64 = get("FCM_SERVICE_ACCOUNT_BASE64")
	if len(c.FCMKeyBase64) > 90000 || (c.FCMKeyBase64 != "" && c.GoogleProject == "") {
		return c, errors.New("invalid FCM environment configuration")
	}
	c.MediaWorkerURL, c.MediaWorkerKey = get("MEDIA_WORKER_URL"), get("MEDIA_WORKER_KEY")
	if (c.MediaWorkerURL == "") != (c.MediaWorkerKey == "") {
		return c, errors.New("both media worker URL and key are required")
	}
	c.S3Endpoint, c.S3AccessKey, c.S3SecretKey = get("S3_ENDPOINT"), get("S3_ACCESS_KEY"), get("S3_SECRET_KEY")
	c.S3Bucket = value(get, "S3_BUCKET", "harkai-photos")
	c.S3Secure = get("S3_USE_SSL") == "true"
	if c.S3Endpoint == "" || c.S3AccessKey == "" || c.S3SecretKey == "" {
		return c, errors.New("S3_ENDPOINT, S3_ACCESS_KEY and S3_SECRET_KEY are required")
	}
	if v := get("S3_USE_SSL"); v != "" && v != "true" && v != "false" {
		return c, errors.New("S3_USE_SSL must be true or false")
	}
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return c, errors.New("PORT must be 1 to 65535")
	}
	if c.MongoURI == "" {
		return c, errors.New("MONGODB_URI is required")
	}
	if len(c.Database) > 63 || strings.ContainsAny(c.Database, " /\\.\"$\x00") {
		return c, errors.New("invalid MONGODB_DATABASE")
	}
	raw := value(get, "CORS_ORIGINS", "https://harkai.lat,https://www.harkai.lat")
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSpace(origin)
		u, err := url.Parse(origin)
		if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
			return c, errors.New("CORS_ORIGINS must contain exact origins")
		}
		c.Origins = append(c.Origins, origin)
	}
	return c, nil
}

func value(get func(string) string, key, fallback string) string {
	if x := get(key); x != "" {
		return x
	}
	return fallback
}
