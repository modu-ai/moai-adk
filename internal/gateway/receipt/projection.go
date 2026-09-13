package receipt

import (
	"bytes"
	"crypto/sha256"
	"encoding"
	"encoding/json"
	"io"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Boundary struct{ Prefix, Previous Digest }

// CanonicalPrefixes accepts the message array after semantic content and opaque
// codec validation. It does not authenticate thinking blocks. decodeID must
// strictly validate/recover both tool-use and tool-result IDs. Top-level policy
// is deliberately absent from this API; system-role messages are retained.
func CanonicalPrefixes(raw []byte, decodeID func(string) (string, error)) ([]Boundary, error) {
	if len(raw) > MaxBytes {
		return nil, ErrLimit
	}
	v, e := decodeStrict(raw)
	if e != nil {
		return nil, e
	}
	messages, ok := v.([]any)
	if !ok {
		return nil, ErrInvalid
	}
	history := sha256.New()
	_, _ = history.Write([]byte{'['})
	out := []Boundary{}
	previous := Digest{}
	for _, v := range messages {
		msg, ok := v.(map[string]any)
		if !ok || len(msg) != 2 {
			return nil, ErrInvalid
		}
		role, ok := msg["role"].(string)
		if !ok || (role != "assistant" && role != "user" && role != "system") {
			return nil, ErrInvalid
		}
		content, e := normalizeContent(msg["content"], decodeID)
		if e != nil {
			return nil, e
		}
		var encoded bytes.Buffer
		canonical(&encoded, map[string]any{"role": role, "content": content})
		_, _ = history.Write(encoded.Bytes())
		if role == "assistant" {
			// SHA-256 documents binary state cloning. Snapshotting avoids
			// rehashing every previous message at each assistant boundary.
			state, _ := history.(encoding.BinaryMarshaler).MarshalBinary()
			snapshot := sha256.New()
			_ = snapshot.(encoding.BinaryUnmarshaler).UnmarshalBinary(state)
			_, _ = snapshot.Write([]byte{']'})
			var prefix Digest
			copy(prefix[:], snapshot.Sum(nil))
			out = append(out, Boundary{prefix, previous})
			previous = prefix
		}
	}
	return out, nil
}
func normalizeContent(v any, decodeID func(string) (string, error)) (any, error) {
	return normalizeContentPosition(v, decodeID, false)
}

func normalizeContentPosition(v any, decodeID func(string) (string, error), toolResult bool) (any, error) {
	if s, ok := v.(string); ok {
		return []any{map[string]any{"type": "text", "text": s}}, nil
	}
	blocks, ok := v.([]any)
	if !ok {
		return nil, ErrInvalid
	}
	result := []any{}
	for _, v := range blocks {
		block, ok := v.(map[string]any)
		if !ok {
			return nil, ErrInvalid
		}
		typ, ok := block["type"].(string)
		if !ok {
			return nil, ErrInvalid
		}
		switch typ {
		case "thinking", "redacted_thinking":
			continue
		case "tool_reference":
			name, valid := block["tool_name"].(string)
			if !toolResult || !valid || name == "" || len(block) != 2 {
				return nil, ErrInvalid
			}
		case "text", "image", "document", "tool_use", "tool_result":
		default:
			return nil, ErrInvalid
		}
		delete(block, "cache_control")
		if typ == "text" {
			if _, ok := block["text"].(string); !ok {
				return nil, ErrInvalid
			}
		}
		key := ""
		if typ == "tool_use" {
			key = "id"
		}
		if typ == "tool_result" {
			key = "tool_use_id"
		}
		if key != "" {
			id, ok := block[key].(string)
			if !ok || id == "" || decodeID == nil {
				return nil, ErrInvalid
			}
			id, e := decodeID(id)
			if e != nil || id == "" {
				return nil, ErrInvalid
			}
			block[key] = id
		}
		if typ == "tool_result" {
			if content, ok := block["content"]; ok {
				var e error
				block["content"], e = normalizeContentPosition(content, decodeID, true)
				if e != nil {
					return nil, e
				}
			}
		}
		result = append(result, block)
	}
	return result, nil
}

// Length-prefixed typed encoding avoids number/string and delimiter collisions.
// Rational normalization preserves number meaning without float64 rounding.
func canonical(b *bytes.Buffer, v any) {
	switch v := v.(type) {
	case nil:
		b.WriteByte('n')
	case bool:
		if v {
			b.WriteByte('t')
		} else {
			b.WriteByte('f')
		}
	case string:
		b.WriteByte('s')
		b.WriteString(strconv.Itoa(len(v)))
		b.WriteByte(':')
		b.WriteString(v)
	case json.Number:
		r := new(big.Rat)
		r.SetString(string(v))
		b.WriteByte('d')
		b.WriteString(r.RatString())
		b.WriteByte(';')
	case []any:
		b.WriteByte('[')
		for _, x := range v {
			canonical(b, x)
		}
		b.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteByte('{')
		for _, k := range keys {
			canonical(b, k)
			canonical(b, v[k])
		}
		b.WriteByte('}')
	}
}
func decodeStrict(raw []byte) (any, error) {
	if !utf8.Valid(raw) || !pairedSurrogates(raw) {
		return nil, ErrInvalid
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	v, e := readValue(dec, 0)
	if e != nil {
		return nil, e
	}
	if _, e = dec.Token(); e != io.EOF {
		return nil, ErrInvalid
	}
	return v, nil
}
func readValue(d *json.Decoder, depth int) (any, error) {
	if depth > 64 {
		return nil, ErrInvalid
	}
	t, e := d.Token()
	if e != nil {
		return nil, ErrInvalid
	}
	switch t := t.(type) {
	case json.Delim:
		if t == '{' {
			m := map[string]any{}
			for d.More() {
				key, e := d.Token()
				if e != nil {
					return nil, ErrInvalid
				}
				k, ok := key.(string)
				if !ok {
					return nil, ErrInvalid
				}
				if _, exists := m[k]; exists {
					return nil, ErrInvalid
				}
				v, e := readValue(d, depth+1)
				if e != nil {
					return nil, e
				}
				m[k] = v
			}
			end, e := d.Token()
			if e != nil || end != json.Delim('}') {
				return nil, ErrInvalid
			}
			return m, nil
		}
		if t == '[' {
			a := []any{}
			for d.More() {
				v, e := readValue(d, depth+1)
				if e != nil {
					return nil, e
				}
				a = append(a, v)
			}
			end, e := d.Token()
			if e != nil || end != json.Delim(']') {
				return nil, ErrInvalid
			}
			return a, nil
		}
		return nil, ErrInvalid
	case json.Number:
		// Bound exponent/digit work before arbitrary-precision normalization.
		s := string(t)
		if len(s) > 1024 {
			return nil, ErrInvalid
		}
		if i := strings.IndexAny(s, "eE"); i >= 0 {
			ex, e := strconv.Atoi(s[i+1:])
			if e != nil || ex > 1024 || ex < -1024 {
				return nil, ErrInvalid
			}
		}
		if _, ok := new(big.Rat).SetString(s); !ok {
			return nil, ErrInvalid
		}
		return t, nil
	default:
		return t, nil
	}
}

// encoding/json substitutes replacement runes for lone UTF-16 escapes. Reject
// those before decoding so distinct malformed texts cannot share a digest.
func pairedSurrogates(raw []byte) bool {
	inString := false
	for i := 0; i < len(raw); i++ {
		if raw[i] == '"' {
			inString = !inString
			continue
		}
		if !inString || raw[i] != 92 {
			continue
		}
		i++
		if i >= len(raw) {
			return false
		}
		if raw[i] != 'u' {
			continue
		}
		if i+4 >= len(raw) {
			return false
		}
		n, e := strconv.ParseUint(string(raw[i+1:i+5]), 16, 16)
		if e != nil {
			return false
		}
		i += 4
		if n >= 0xdc00 && n <= 0xdfff {
			return false
		}
		if n >= 0xd800 && n <= 0xdbff {
			if i+6 >= len(raw) || raw[i+1] != 92 || raw[i+2] != 'u' {
				return false
			}
			low, e := strconv.ParseUint(string(raw[i+3:i+7]), 16, 16)
			if e != nil || low < 0xdc00 || low > 0xdfff {
				return false
			}
			i += 6
		}
	}
	return true
}
