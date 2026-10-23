package core

import "testing"

func TestGooglePhotoURL(t *testing.T) {
	for _, value := range []string{"http://lh3.googleusercontent.com/a", "https://googleusercontent.com.evil.test/a", "https://example.com/a", "data:image/png;base64,x", ""} {
		if GooglePhotoURL(value) != "" {
			t.Errorf("accepted unsupported photo %q", value)
		}
	}
	value := "https://lh3.googleusercontent.com/a/photo=s96-c"
	if GooglePhotoURL(value) != value {
		t.Fatal("lost verified Google photo")
	}
}
