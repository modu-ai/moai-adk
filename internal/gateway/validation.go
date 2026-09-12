package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

var ErrInvalidJSON = errors.New("gateway invalid or ambiguous JSON")

// ValidateJSONObject rejects duplicate keys at every depth, invalid UTF-8,
// trailing values and non-object roots before configuration or routing use.
func ValidateJSONObject(body []byte) error {
	if !utf8.Valid(body) {
		return ErrInvalidJSON
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := uniqueJSONValue(dec); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return ErrInvalidJSON
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil || object == nil {
		return ErrInvalidJSON
	}
	return nil
}

// IsValidationRequest recognizes the four structural conditions observed in
// Claude Code 2.1.268. It does not consult model registry or credentials, and
// must run only after HTTP path, method and body-size policy checks.
func IsValidationRequest(body []byte) (bool, error) {
	if err := ValidateJSONObject(body); err != nil {
		return false, err
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(body, &fields)
	if stream, ok := fields["stream"]; ok && !bytes.Equal(bytes.TrimSpace(stream), []byte("false")) {
		return false, nil
	}
	if !bytes.Equal(bytes.TrimSpace(fields["max_tokens"]), []byte("1")) {
		return false, nil
	}
	var messages []map[string]json.RawMessage
	if err := json.Unmarshal(fields["messages"], &messages); err != nil || len(messages) != 1 {
		return false, nil
	}
	var role string
	if err := json.Unmarshal(messages[0]["role"], &role); err != nil || role != "user" {
		return false, nil
	}
	if tools, ok := fields["tools"]; ok {
		var items []json.RawMessage
		if err := json.Unmarshal(tools, &items); err != nil || items == nil || len(items) != 0 {
			return false, nil
		}
	}
	return true, nil
}

// uniqueJSONValue rejects duplicate keys at every object depth, including escaped
// spellings of the same key. This avoids disagreement with downstream parsers.
func uniqueJSONValue(dec *json.Decoder) error {
	token, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for dec.More() {
			key, err := dec.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return errors.New("duplicate object key")
			}
			seen[name] = true
			if err := uniqueJSONValue(dec); err != nil {
				return err
			}
		}
	case '[':
		for dec.More() {
			if err := uniqueJSONValue(dec); err != nil {
				return err
			}
		}
	default:
		return errors.New("unexpected JSON delimiter")
	}
	_, err = dec.Token()
	return err
}
