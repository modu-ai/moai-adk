package toolpolicy

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// PermissionsBlock is the Claude Code settings.json permissions sub-object.
// Only the three list fields + defaultMode are modeled; additional permission
// keys Claude Code may introduce are preserved by the parse-modify-serialize
// strategy (Raw map), never dropped.
type PermissionsBlock struct {
	DefaultMode string            `json:"defaultMode,omitempty"`
	Allow       []string          `json:"allow"`
	Ask         []string          `json:"ask,omitempty"`
	Deny        []string          `json:"deny,omitempty"`
	Raw         map[string]json.RawMessage `json:"-"`
}

// permissionsRegion is the parsed region boundaries of a settings file's
// "permissions": { ... } object. Start/End are byte offsets into the original
// file text spanning exactly the permissions object (including the key and
// enclosing braces).
type permissionsRegion struct {
	start int // byte offset of the opening quote of "permissions"
	end   int // byte offset AFTER the closing brace of the object
}

// locatePermissionsRegion finds the byte span of the "permissions": { ... }
// object in a settings file body via brace-depth matching. It does NOT parse
// the file as JSON — it walks the raw text — so it works equally on pure JSON
// (.claude/settings.json) and on mixed JSON + Go-template directives
// (.tmpl files that contain {{jsonEscape .SmartPATH}} in the env block).
//
// The matching is string-literal-aware: braces inside JSON string values
// (e.g., a deny pattern containing "}") are not counted as depth changes.
// This is the load-bearing safety property for raw-text region replacement
// on the .tmpl target (AC-TPS-014 — region replacement must not be fooled by
// braces inside string literals).
func locatePermissionsRegion(body []byte) (permissionsRegion, error) {
	idx := indexOfPermissionsKey(body)
	if idx < 0 {
		return permissionsRegion{}, fmt.Errorf("permissions block not found: no \"permissions\": key in body")
	}
	// Find the opening brace of the permissions object value.
	open := idx
	for open < len(body) && body[open] != '{' {
		open++
	}
	if open >= len(body) {
		return permissionsRegion{}, fmt.Errorf("permissions block malformed: no opening brace after \"permissions\": key")
	}
	// Walk from the opening brace, tracking depth, skipping string literals.
	depth := 0
	inString := false
	escaped := false
	for i := open; i < len(body); i++ {
		c := body[i]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' && inString {
			escaped = true
			continue
		}
		if c == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		switch c {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return permissionsRegion{start: idx, end: i + 1}, nil
			}
		}
	}
	return permissionsRegion{}, fmt.Errorf("permissions block malformed: unmatched braces (depth=%d at EOF)", depth)
}

// indexOfPermissionsKey returns the byte index of the opening quote of the
// "permissions" JSON key, or -1 if not found. It scans for the literal
// `"permissions"` token followed by optional whitespace and a colon.
func indexOfPermissionsKey(body []byte) int {
	needle := []byte(`"permissions"`)
	// Search for the key as a standalone token (preceded by { or , or
	// whitespace) so we do not match a substring of a longer key.
	for i := 0; i+len(needle) <= len(body); i++ {
		if !bytesEqual(body[i:i+len(needle)], needle) {
			continue
		}
		// Verify preceding char is a JSON delimiter/whitespace.
		if i > 0 {
			prev := body[i-1]
			if prev != '{' && prev != ',' && !isSpaceByte(prev) {
				continue
			}
		}
		// Verify following char is whitespace or colon.
		j := i + len(needle)
		for j < len(body) && isSpaceByte(body[j]) {
			j++
		}
		if j < len(body) && body[j] == ':' {
			return i
		}
	}
	return -1
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func isSpaceByte(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// extractPermissions parses the permissions object out of a settings file body
// and unmarshals it into a PermissionsBlock. Used by both the JSON and the
// .tmpl codegen paths to read the existing block before regenerating it.
func extractPermissions(body []byte) (*PermissionsBlock, error) {
	region, err := locatePermissionsRegion(body)
	if err != nil {
		return nil, err
	}
	// The region spans `"permissions": { ... }`. Extract just the object value
	// (from the opening brace to the closing brace).
	objStart := region.start
	for objStart < region.end && body[objStart] != '{' {
		objStart++
	}
	objBytes := body[objStart:region.end]

	// First unmarshal into a raw map to capture any extra keys (defaultMode,
	// additionalDirectories, etc.) so the round-trip preserves them.
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(objBytes, &raw); err != nil {
		return nil, fmt.Errorf("permissions object parse: %w", err)
	}
	block := &PermissionsBlock{Raw: raw}
	if v, ok := raw["defaultMode"]; ok {
		_ = json.Unmarshal(v, &block.DefaultMode)
	}
	block.Allow = extractStringList(raw, "allow")
	block.Ask = extractStringList(raw, "ask")
	block.Deny = extractStringList(raw, "deny")
	return block, nil
}

func extractStringList(raw map[string]json.RawMessage, key string) []string {
	v, ok := raw[key]
	if !ok {
		return nil
	}
	var out []string
	_ = json.Unmarshal(v, &out)
	return out
}

// renderPermissionsObject serializes a PermissionsBlock back to an indented
// JSON object literal (without the "permissions": key). Key order is
// canonical: defaultMode, allow, ask, deny, then any extra Raw keys sorted
// alphabetically. This fixed ordering is load-bearing for AC-TPS-013
// (idempotency — byte-identical output across runs).
func renderPermissionsObject(block *PermissionsBlock) ([]byte, error) {
	var sb strings.Builder
	sb.WriteString("{\n")
	first := true

	writeKey := func(key string, value string) {
		if !first {
			sb.WriteString(",\n")
		}
		first = false
		sb.WriteString("    ")
		sb.WriteString(quoteJSON(key))
		sb.WriteString(": ")
		sb.WriteString(value)
	}

	if block.DefaultMode != "" {
		writeKey("defaultMode", quoteJSON(block.DefaultMode))
	}
	if len(block.Allow) > 0 {
		writeKey("allow", renderStringList(block.Allow))
	}
	if len(block.Ask) > 0 {
		writeKey("ask", renderStringList(block.Ask))
	}
	if len(block.Deny) > 0 {
		writeKey("deny", renderStringList(block.Deny))
	}

	// Preserve any extra keys that were in the original Raw map (excluding
	// the four we already serialized). Sorted for determinism.
	consumed := map[string]bool{
		"defaultMode": true,
		"allow":       true,
		"ask":         true,
		"deny":        true,
	}
	var extraKeys []string
	for k := range block.Raw {
		if !consumed[k] {
			extraKeys = append(extraKeys, k)
		}
	}
	sort.Strings(extraKeys)
	for _, k := range extraKeys {
		writeKey(k, string(block.Raw[k]))
	}

	sb.WriteString("\n  }")
	return []byte(sb.String()), nil
}

// renderStringList produces a JSON array literal of quoted strings, indented
// with 6 spaces per element (matching the surrounding 4-space permissions
// object indentation + 2 for the array contents).
func renderStringList(items []string) string {
	var sb strings.Builder
	sb.WriteString("[\n")
	for i, s := range items {
		sb.WriteString("      ")
		sb.WriteString(quoteJSON(s))
		if i < len(items)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n    ")
	}
	sb.WriteString("]")
	return sb.String()
}

// quoteJSON returns a JSON-quoted string literal for s.
func quoteJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
