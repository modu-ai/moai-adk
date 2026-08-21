// tomldecodertest_test.go — independent test-side TOML decoder.
//
// Implements, from the TOML v1.0.0 spec, ONLY the grammar subset the emitter
// is allowed to produce: full-line comments, bare keys, basic strings
// ("..." with \" and \\ escapes), multi-line literal strings (three
// apostrophes, with first-newline trim and up to two apostrophes admitted
// before the closing delimiter), and arrays of basic strings.
//
// This file is deliberately independent of writer.go: it exists so a writer
// bug cannot be masked by sharing decode logic with the code that produced
// the bytes. The emitted artifacts are additionally parsed by codex-cli
// itself in the MS2 probe smoke (the real consumer).
package agentemit_test

import (
	"fmt"
	"strings"
)

// decodeTOML parses src into a map keyed by top-level bare key. Values are
// string or []string. Anything outside the emitter's declared grammar is an
// error — this decoder is strict by design.
func decodeTOML(src string) (map[string]any, error) {
	out := map[string]any{}
	lines := strings.Split(src, "\n")
	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			i++
			continue
		}
		eq := strings.Index(line, "=")
		if eq < 0 {
			return nil, fmt.Errorf("line %d: expected key = value, got %q", i+1, line)
		}
		key := strings.TrimSpace(line[:eq])
		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", i+1)
		}
		rest := strings.TrimSpace(line[eq+1:])
		switch {
		case strings.HasPrefix(rest, "'''"):
			value, consumed, err := decodeMultiLineLiteral(lines, i, eq)
			if err != nil {
				return nil, err
			}
			if _, dup := out[key]; dup {
				return nil, fmt.Errorf("line %d: duplicate key %q", i+1, key)
			}
			out[key] = value
			i = consumed + 1 // skip past the closing-delimiter line
		case strings.HasPrefix(rest, `"`):
			value, err := decodeBasicString(rest)
			if err != nil {
				return nil, fmt.Errorf("line %d: %v", i+1, err)
			}
			if _, dup := out[key]; dup {
				return nil, fmt.Errorf("line %d: duplicate key %q", i+1, key)
			}
			out[key] = value
			i++
		case strings.HasPrefix(rest, "["):
			value, err := decodeStringArray(rest)
			if err != nil {
				return nil, fmt.Errorf("line %d: %v", i+1, err)
			}
			if _, dup := out[key]; dup {
				return nil, fmt.Errorf("line %d: duplicate key %q", i+1, key)
			}
			out[key] = value
			i++
		default:
			return nil, fmt.Errorf("line %d: value outside emitter grammar: %q", i+1, rest)
		}
	}
	return out, nil
}

// decodeMultiLineLiteral parses a multi-line literal string starting at
// lines[startLine] where the value begins after "=". Per the spec, a newline
// immediately following the opening delimiter is trimmed, and one or two
// apostrophes may appear immediately before the closing delimiter (they are
// content). Returns the decoded value and the index of the line containing
// the closing delimiter (the caller advances past it).
func decodeMultiLineLiteral(lines []string, startLine, eq int) (string, int, error) {
	// Value text begins after ''' on the start line.
	afterOpen := strings.TrimPrefix(strings.TrimSpace(lines[startLine][eq+1:]), "'''")
	if afterOpen != "" {
		return "", 0, fmt.Errorf("line %d: content on opening delimiter line", startLine+1)
	}
	// First newline after the opening delimiter is trimmed: the content
	// starts at lines[startLine+1].
	var b strings.Builder
	for j := startLine + 1; j < len(lines); j++ {
		line := lines[j]
		closeIdx := strings.Index(line, "'''")
		if closeIdx < 0 {
			b.WriteString(line)
			b.WriteString("\n")
			continue
		}
		// Candidate close. Up to two apostrophes immediately before the
		// closing delimiter are content per spec; three or more are an error.
		adjacent := 0
		for k := closeIdx - 1; k >= 0 && line[k] == '\''; k-- {
			adjacent++
		}
		if adjacent > 2 {
			return "", 0, fmt.Errorf("line %d: three or more apostrophes before closing delimiter", j+1)
		}
		b.WriteString(line[:closeIdx-adjacent])
		trailing := strings.TrimSpace(line[closeIdx+3:])
		if trailing != "" && !strings.HasPrefix(trailing, "#") {
			return "", 0, fmt.Errorf("line %d: trailing content after closing delimiter: %q", j+1, trailing)
		}
		return b.String(), j, nil
	}
	return "", 0, fmt.Errorf("unterminated multi-line literal starting at line %d", startLine+1)
}

// decodeBasicString decodes a single-line basic string with the minimal
// escape set the emitter may legally need (\" and \\). It rejects multi-line
// and control characters.
func decodeBasicString(s string) (string, error) {
	if !strings.HasPrefix(s, `"`) {
		return "", fmt.Errorf("not a basic string: %q", s)
	}
	var b strings.Builder
	i := 1
	for i < len(s) {
		c := s[i]
		switch c {
		case '"':
			trailing := strings.TrimSpace(s[i+1:])
			if trailing != "" && !strings.HasPrefix(trailing, "#") {
				return "", fmt.Errorf("trailing content after string: %q", trailing)
			}
			return b.String(), nil
		case '\\':
			if i+1 >= len(s) {
				return "", fmt.Errorf("dangling escape")
			}
			switch s[i+1] {
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			default:
				return "", fmt.Errorf("unsupported escape \\%c", s[i+1])
			}
			i += 2
		case '\n', '\r':
			return "", fmt.Errorf("raw newline in single-line basic string")
		default:
			if c < 0x20 && c != '\t' {
				return "", fmt.Errorf("control char %q", c)
			}
			b.WriteByte(c)
			i++
		}
	}
	return "", fmt.Errorf("unterminated basic string")
}

// decodeStringArray decodes a single-line array of basic strings, e.g.
// ["moai"].
func decodeStringArray(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[") || !strings.HasSuffix(s, "]") {
		return nil, fmt.Errorf("not an array: %q", s)
	}
	inner := strings.TrimSpace(s[1 : len(s)-1])
	if inner == "" {
		return []string{}, nil
	}
	parts := strings.Split(inner, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v, err := decodeBasicString(strings.TrimSpace(p))
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}
