package security

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

func TestImagesAreDecodedAndReencoded(t *testing.T) {
	var input bytes.Buffer
	_ = png.Encode(&input, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	data, err := NormalizeImage(input.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || format != "jpeg" {
		t.Fatal("image not normalized to JPEG")
	}
	for _, data := range [][]byte{[]byte("<svg onload='evil()'></svg>"), []byte("not-an-image"), make([]byte, MaxImageBytes+1)} {
		if _, err := NormalizeImage(data); err == nil {
			t.Fatal("invalid image accepted")
		}
	}
}

func TestTokensAreRandomAndHashed(t *testing.T) {
	a, err := Token()
	if err != nil {
		t.Fatal(err)
	}
	b, err := Token()
	if err != nil {
		t.Fatal(err)
	}
	if a == b || len(a) != 43 || HashToken(a) == a || len(HashToken(a)) != 64 {
		t.Fatal("unsafe token")
	}
}
