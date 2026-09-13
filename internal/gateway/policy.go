package gateway

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/url"
)

var errBodyTooLarge = errors.New("gateway body too large")

func readBounded(r io.Reader, limit int64) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, errBodyTooLarge
	}
	return b, nil
}
func readRequestBody(r *http.Request, limit int64) ([]byte, error) {
	raw, err := readBounded(r.Body, limit)
	if err != nil {
		return nil, err
	}
	switch r.Header.Get("Content-Encoding") {
	case "":
		return raw, nil
	case "gzip":
		z, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil, err
		}
		defer func() { _ = z.Close() }() // gzip over in-memory bytes; Close carries no flush to lose
		return readBounded(z, limit)
	default:
		return nil, errors.New("unsupported content encoding")
	}
}

func pathPolicy(r *http.Request) (int, string) {
	if r.URL.EscapedPath() != r.URL.Path {
		return 404, "not_found_error"
	}
	method := "POST"
	switch r.URL.Path {
	case "/v1/messages", "/v1/messages/count_tokens":
	case "/v1/models":
		method = "GET"
	default:
		return 404, "not_found_error"
	}
	if r.Method != method {
		return 405, "invalid_request_error"
	}
	if r.URL.RawQuery != "" {
		query, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil || len(query) != 1 {
			return 400, "invalid_request_error"
		}
		key, value := "beta", "true"
		if r.URL.Path == "/v1/models" {
			key, value = "limit", "1000"
		}
		values := query[key]
		if len(values) != 1 || values[0] != value {
			return 400, "invalid_request_error"
		}
	}
	return 0, ""
}
