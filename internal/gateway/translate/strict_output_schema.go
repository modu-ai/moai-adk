package translate

import (
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"
	"unicode/utf8"
)

// strictOutputSchema copies schema nodes, never interpreting enum/const data. The
// strict upstream contract requires all declared properties and closed objects.
func strictOutputSchema(value any, depth int) (map[string]any, error) {
	root, ok := value.(map[string]any)
	if !ok || root["type"] != "object" || root["anyOf"] != nil {
		return nil, errors.New("output schema root must be an object without anyOf")
	}
	budget := outputSchemaBudget{root: root}
	if err := budget.check(root, 0, depth); err != nil {
		return nil, err
	}
	return normalizeOutputSchema(root, depth)
}

type outputSchemaBudget struct {
	root                     map[string]any
	properties, chars, enums int
}

func (b *outputSchemaBudget) check(value any, nesting, depth int) error {
	s, ok := value.(map[string]any)
	if !ok || depth > 64 {
		return errors.New("invalid output schema structure")
	}
	container := s["type"] == "object" || s["type"] == "array"
	if types, ok := s["type"].([]any); ok {
		for _, typ := range types {
			if typ == "object" || typ == "array" {
				container = true
			}
		}
	}
	if container {
		nesting++
	}
	if nesting > 10 {
		return errors.New("output schema nesting limit")
	}
	hasType := func(want string) bool {
		if s["type"] == want {
			return true
		}
		if types, ok := s["type"].([]any); ok {
			for _, typ := range types {
				if typ == want {
					return true
				}
			}
		}
		return false
	}
	if hasType("array") && s["items"] == nil {
		return errors.New("output schema array requires items")
	}
	for key, v := range s {
		want := ""
		switch key {
		case "properties", "required", "additionalProperties":
			want = "object"
		case "items", "minItems", "maxItems":
			want = "array"
		case "pattern", "format":
			want = "string"
		case "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf":
			want = "number"
		}
		compatible := hasType(want) || (want == "number" && hasType("integer"))
		if want != "" && s["type"] != nil && !compatible {
			return errors.New("output schema keyword incompatible with type")
		}
		switch key {
		case "type", "required", "additionalProperties":
		case "title", "description":
			if _, ok := v.(string); !ok {
				return errors.New("invalid output schema annotation")
			}
		case "properties", "$defs":
			entries, ok := v.(map[string]any)
			if !ok {
				return errors.New("invalid output schema entries")
			}
			if key == "properties" {
				b.properties += len(entries)
			}
			for name, child := range entries {
				b.chars += utf8.RuneCountInString(name)
				childNesting := nesting
				if key != "properties" {
					childNesting = 0
				}
				if err := b.check(child, childNesting, depth+1); err != nil {
					return err
				}
			}
		case "items":
			if err := b.check(v, nesting, depth+1); err != nil {
				return err
			}
		case "anyOf":
			items, ok := v.([]any)
			if !ok || len(items) == 0 {
				return errors.New("invalid output schema alternatives")
			}
			for _, child := range items {
				if err := b.check(child, nesting, depth+1); err != nil {
					return err
				}
			}
		case "$ref":
			ref, ok := v.(string)
			if !ok || !b.validReference(ref) {
				return errors.New("invalid output schema reference")
			}
		case "enum":
			items, ok := v.([]any)
			if !ok || len(items) == 0 {
				return errors.New("invalid output schema enum")
			}
			b.enums += len(items)
			chars := 0
			seen := map[string]bool{}
			for _, item := range items {
				raw, err := json.Marshal(item)
				if err != nil || seen[string(raw)] {
					return errors.New("invalid output schema enum value")
				}
				seen[string(raw)] = true
				chars += outputSchemaDataChars(item)
			}
			b.chars += chars
			if len(items) > 250 && chars > 15000 {
				return errors.New("output schema enum string limit")
			}
		case "const":
			b.chars += outputSchemaDataChars(v)
		case "pattern":
			if _, ok := v.(string); !ok {
				return errors.New("invalid output schema pattern")
			}
		case "format":
			switch v {
			case "date-time", "time", "date", "duration", "email", "hostname", "ipv4", "ipv6", "uuid":
			default:
				return errors.New("unsupported output schema format")
			}
		case "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf", "minItems", "maxItems":
			n, ok := v.(float64)
			if !ok || math.IsNaN(n) || math.IsInf(n, 0) {
				return errors.New("invalid output schema numeric restriction")
			}
			if key == "multipleOf" && n <= 0 {
				return errors.New("invalid output schema multipleOf")
			}
			if (key == "minItems" || key == "maxItems") && (n < 0 || math.Trunc(n) != n) {
				return errors.New("invalid output schema array restriction")
			}
		default:
			return errors.New("unsupported output schema keyword")
		}
	}
	if v, exists := s["additionalProperties"]; exists && v != false {
		return errors.New("output schema open objects are unsupported")
	}
	if b.properties > 5000 || b.chars > 120000 || b.enums > 1000 {
		return errors.New("output schema size limit")
	}
	return nil
}

// Reference validation does not expand targets: recursive definitions are valid
// and must not turn the finite preflight into unbounded recursion or I/O.
func (b *outputSchemaBudget) validReference(ref string) bool {
	if ref == "#" {
		return true
	}
	if !strings.HasPrefix(ref, "#/") {
		return false
	}
	var current any = b.root
	for _, part := range strings.Split(ref[2:], "/") {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		m, ok := current.(map[string]any)
		if !ok {
			return false
		}
		current, ok = m[part]
		if !ok {
			return false
		}
	}
	_, ok := current.(map[string]any)
	return ok
}

func outputSchemaDataChars(value any) int {
	switch v := value.(type) {
	case string:
		return utf8.RuneCountInString(v)
	case []any:
		n := 0
		for _, item := range v {
			n += outputSchemaDataChars(item)
		}
		return n
	case map[string]any:
		n := 0
		for key, item := range v {
			n += utf8.RuneCountInString(key) + outputSchemaDataChars(item)
		}
		return n
	}
	return 0
}

func normalizeOutputSchema(value any, depth int) (map[string]any, error) {
	schema, ok := value.(map[string]any)
	if !ok || depth > 64 {
		return nil, errors.New("invalid output schema structure")
	}
	if value, exists := schema["type"]; exists {
		types, array := value.([]any)
		if !array {
			types = []any{value}
		}
		if len(types) == 0 {
			return nil, errors.New("invalid output schema type")
		}
		seen := map[string]bool{}
		for _, value := range types {
			typ, ok := value.(string)
			if !ok || seen[typ] {
				return nil, errors.New("invalid output schema type")
			}
			switch typ {
			case "object", "array", "string", "number", "integer", "boolean", "null":
				seen[typ] = true
			default:
				return nil, errors.New("invalid output schema type")
			}
		}
	}
	if value, exists := schema["additionalProperties"]; exists {
		switch value.(type) {
		case bool, map[string]any:
		default:
			return nil, errors.New("invalid output schema additionalProperties")
		}
	}
	out := make(map[string]any, len(schema))
	for key, value := range schema {
		out[key] = value
	}
	for _, key := range []string{"properties", "$defs", "definitions"} {
		if value, exists := schema[key]; exists {
			entries, ok := value.(map[string]any)
			if !ok {
				return nil, errors.New("invalid output schema entries")
			}
			cloned := make(map[string]any, len(entries))
			for name, child := range entries {
				normalized, err := normalizeOutputSchema(child, depth+1)
				if err != nil {
					return nil, err
				}
				cloned[name] = normalized
			}
			out[key] = cloned
		}
	}
	if value, exists := schema["items"]; exists {
		child, err := normalizeOutputSchema(value, depth+1)
		if err != nil {
			return nil, err
		}
		out["items"] = child
	}
	for _, key := range []string{"anyOf", "allOf", "oneOf"} {
		if value, exists := schema[key]; exists {
			items, ok := value.([]any)
			if !ok || len(items) == 0 {
				return nil, errors.New("invalid output schema alternatives")
			}
			cloned := make([]any, len(items))
			for i, child := range items {
				normalized, err := normalizeOutputSchema(child, depth+1)
				if err != nil {
					return nil, err
				}
				cloned[i] = normalized
			}
			out[key] = cloned
		}
	}
	object := schema["type"] == "object"
	if types, ok := schema["type"].([]any); ok {
		for _, typ := range types {
			if typ == "object" {
				object = true
			}
		}
	}
	properties, hasProperties := out["properties"].(map[string]any)
	if object || hasProperties {
		if !hasProperties {
			properties = map[string]any{}
			out["properties"] = properties
		}
		required := make([]any, 0, len(properties))
		seen := map[string]bool{}
		if value, exists := schema["required"]; exists {
			items, ok := value.([]any)
			if !ok {
				return nil, errors.New("invalid output schema required")
			}
			for _, item := range items {
				name, ok := item.(string)
				if !ok || seen[name] {
					return nil, errors.New("invalid output schema required")
				}
				if _, ok := properties[name]; !ok {
					return nil, errors.New("unknown output schema required property")
				}
				seen[name] = true
				required = append(required, name)
			}
		}
		missing := make([]string, 0, len(properties))
		for name := range properties {
			if !seen[name] {
				missing = append(missing, name)
			}
		}
		sort.Strings(missing)
		for _, name := range missing {
			required = append(required, name)
		}
		out["required"] = required
		out["additionalProperties"] = false
	}
	return out, nil
}
