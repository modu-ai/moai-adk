package translate

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
)

// ImageSourceURL validates a public Messages image block and returns the image
// representation accepted by both Responses and App Server. It performs no IO.
func ImageSourceURL(raw json.RawMessage) (string, error) {
	var block map[string]any
	if json.Unmarshal(raw, &block) != nil {
		return "", errors.New("invalid image block")
	}
	if err := keys(block, "type", "source", "cache_control"); err != nil {
		return "", err
	}
	if block["type"] != "image" {
		return "", errors.New("invalid image type")
	}
	if err := cacheHint(block); err != nil {
		return "", err
	}
	source, ok := block["source"].(map[string]any)
	if !ok {
		return "", errors.New("missing image source")
	}
	switch source["type"] {
	case "base64":
		if err := keys(source, "type", "media_type", "data"); err != nil {
			return "", err
		}
		mime, ok := source["media_type"].(string)
		if !ok {
			return "", errors.New("invalid image media type")
		}
		switch mime {
		case "image/png", "image/jpeg", "image/gif", "image/webp":
		default:
			return "", errors.New("unsupported image media type")
		}
		data, ok := source["data"].(string)
		if !ok || len(data) == 0 || len(data) > 8<<20 {
			return "", errors.New("invalid image data size")
		}
		decoded, err := base64.StdEncoding.Strict().DecodeString(data)
		if err != nil || len(decoded) == 0 {
			return "", errors.New("invalid base64 image")
		}
		return "data:" + mime + ";base64," + data, nil
	case "url":
		if err := keys(source, "type", "url"); err != nil {
			return "", err
		}
		value, ok := source["url"].(string)
		if !ok || len(value) > 8192 {
			return "", errors.New("invalid image URL")
		}
		u, err := url.Parse(value)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" || strings.ContainsAny(value, "\r\n") {
			return "", errors.New("image URL must be HTTPS without credentials")
		}
		return value, nil
	default:
		return "", errors.New("unsupported image source")
	}
}
