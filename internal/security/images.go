package security

import (
	"bytes"
	"fmt"
	"github.com/lucast1574/harkai-backend/internal/core"
	"image"
	"image/jpeg"
	_ "image/png"
)

const MaxImageBytes = 5 << 20

func NormalizeImage(raw []byte) ([]byte, error) {
	if len(raw) == 0 || len(raw) > MaxImageBytes {
		return nil, fmt.Errorf("%w: image must be at most 5 MiB", core.ErrInvalid)
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || (format != "jpeg" && format != "png") {
		return nil, fmt.Errorf("%w: JPEG or PNG required", core.ErrInvalid)
	}
	if config.Width < 1 || config.Height < 1 || config.Width > 4096 || config.Height > 4096 || config.Width*config.Height > 12000000 {
		return nil, fmt.Errorf("%w: image dimensions exceed limit", core.ErrInvalid)
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("%w: corrupt image", core.ErrInvalid)
	}
	var output bytes.Buffer
	if err = jpeg.Encode(&output, img, &jpeg.Options{Quality: 85}); err != nil {
		return nil, err
	}
	if output.Len() > MaxImageBytes {
		return nil, fmt.Errorf("%w: encoded image exceeds limit", core.ErrInvalid)
	}
	return output.Bytes(), nil
}
