package codextools

import (
	"bytes"
	"encoding/json"
	"io"
)

// strictJSON preserves number spelling and array order while rejecting duplicate
// keys and excessive nesting before schema compilation or argument validation.
func strictJSON(raw []byte, limit int) (any, error) {
	if len(raw) == 0 || len(raw) > limit {
		return nil, ErrInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value func(int) (any, error)
	value = func(depth int) (any, error) {
		if depth > 64 {
			return nil, ErrInvalid
		}
		token, err := decoder.Token()
		if err != nil {
			return nil, ErrInvalid
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return token, nil
		}
		switch delimiter {
		case '{':
			m := map[string]any{}
			for decoder.More() {
				key, err := decoder.Token()
				if err != nil {
					return nil, ErrInvalid
				}
				s, ok := key.(string)
				if !ok {
					return nil, ErrInvalid
				}
				if _, exists := m[s]; exists {
					return nil, ErrInvalid
				}
				v, err := value(depth + 1)
				if err != nil {
					return nil, err
				}
				m[s] = v
			}
			last, err := decoder.Token()
			if err != nil || last != json.Delim('}') {
				return nil, ErrInvalid
			}
			return m, nil
		case '[':
			items := []any{}
			for decoder.More() {
				v, err := value(depth + 1)
				if err != nil {
					return nil, err
				}
				items = append(items, v)
			}
			last, err := decoder.Token()
			if err != nil || last != json.Delim(']') {
				return nil, ErrInvalid
			}
			return items, nil
		default:
			return nil, ErrInvalid
		}
	}
	result, err := value(0)
	if err != nil {
		return nil, err
	}
	if _, err = decoder.Token(); err != io.EOF {
		return nil, ErrInvalid
	}
	return result, nil
}
