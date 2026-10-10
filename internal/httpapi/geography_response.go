package httpapi

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"strings"
)

func geographyResponse(r *http.Request, body any) (Result, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return Result{}, err
	}
	headers := http.Header{"Content-Type": []string{"application/json; charset=utf-8"}, "Vary": []string{"Accept-Encoding"}}
	for _, encoding := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		if strings.TrimSpace(encoding) != "gzip" {
			continue
		}
		var buffer bytes.Buffer
		writer, _ := gzip.NewWriterLevel(&buffer, gzip.BestSpeed)
		if _, err = writer.Write(data); err != nil {
			return Result{}, err
		}
		if err = writer.Close(); err != nil {
			return Result{}, err
		}
		data = buffer.Bytes()
		headers.Set("Content-Encoding", "gzip")
		break
	}
	return Result{Bytes: data, Headers: headers}, nil
}
