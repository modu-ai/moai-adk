package opaque

import (
	"bytes"
	"encoding/json"
	"io"
	"unicode/utf8"
)

// Kept local to avoid a translate -> opaque -> translate dependency during codec integration.
func object(raw []byte) (map[string]json.RawMessage, error) {
	if len(raw) > MaxEnvelopeBytes || !utf8.Valid(raw) || !unicodeEscapes(raw) {
		return nil, errInvalid
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := walk(d, 0); err != nil {
		return nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, errInvalid
	}
	var m map[string]json.RawMessage
	if json.Unmarshal(raw, &m) != nil || m == nil {
		return nil, errInvalid
	}
	return m, nil
}
func walk(d *json.Decoder, depth int) error {
	if depth > 64 {
		return errInvalid
	}
	t, err := d.Token()
	if err != nil {
		return errInvalid
	}
	switch t {
	case json.Delim('{'):
		seen := map[string]bool{}
		for d.More() {
			k, err := d.Token()
			if err != nil {
				return errInvalid
			}
			key, ok := k.(string)
			if !ok || seen[key] {
				return errInvalid
			}
			seen[key] = true
			if err := walk(d, depth+1); err != nil {
				return err
			}
		}
		t, err = d.Token()
		if err != nil || t != json.Delim('}') {
			return errInvalid
		}
	case json.Delim('['):
		for d.More() {
			if err := walk(d, depth+1); err != nil {
				return err
			}
		}
		t, err = d.Token()
		if err != nil || t != json.Delim(']') {
			return errInvalid
		}
	default:
		if _, ok := t.(json.Delim); ok {
			return errInvalid
		}
	}
	return nil
}
