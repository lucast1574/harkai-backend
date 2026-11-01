package core

import (
	"strings"
	"testing"
)

func TestDevicePlatforms(t *testing.T) {
	for _, platform := range []string{"android", "ios", "web", "unsupported"} {
		_, err := ValidateDevice(DeviceInput{Token: strings.Repeat("a", 40), Platform: platform, Latitude: -12, Longitude: -77})
		if (err == nil) != (platform != "unsupported") {
			t.Errorf("platform %s: %v", platform, err)
		}
	}
}
