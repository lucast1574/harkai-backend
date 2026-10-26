package httpapi

import (
	"bytes"
	"encoding/base64"
	"errors"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"

	"github.com/lucast1574/harkai-backend/internal/core"
)

func supportImages(input []string) ([]core.SupportImage, error) {
	if len(input) > 3 {
		return nil, errors.New("attach up to three screenshots")
	}
	images := make([]core.SupportImage, 0, len(input))
	for _, encoded := range input {
		prefix, raw, ok := strings.Cut(encoded, ",")
		if !ok || (prefix != "data:image/png;base64" && prefix != "data:image/jpeg;base64" && prefix != "data:image/webp;base64") || len(raw) > base64.StdEncoding.EncodedLen(core.MaxSupportImageBytes) {
			return nil, errors.New("invalid screenshot")
		}
		data, err := base64.StdEncoding.DecodeString(raw)
		if err != nil || len(data) == 0 || len(data) > core.MaxSupportImageBytes {
			return nil, errors.New("screenshots must be smaller than 1 MiB")
		}
		mime := http.DetectContentType(data)
		if prefix != "data:"+mime+";base64" {
			return nil, errors.New("invalid screenshot type")
		}
		config, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 6000 || config.Height > 6000 || int64(config.Width)*int64(config.Height) > 12000000 {
			return nil, errors.New("invalid screenshot dimensions")
		}
		if _, _, err = image.Decode(bytes.NewReader(data)); err != nil {
			return nil, errors.New("corrupt screenshot")
		}
		images = append(images, core.SupportImage{MIME: mime, Data: data})
	}
	return images, nil
}
