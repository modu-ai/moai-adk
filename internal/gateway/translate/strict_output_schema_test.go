package translate

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestGPTStrictOutputSchemaNormalizesNestedOptionalProperties(t *testing.T) {
	body := []byte(`{"max_tokens":100,"messages":[{"role":"user","content":"check"}],"output_config":{"format":{"type":"json_schema","schema":{"type":"object","properties":{"ok":{"type":"boolean"},"reason":{"type":"string"},"impossible":{"type":"boolean"},"rows":{"type":"array","items":{"anyOf":[{"type":"object","properties":{"z":{"type":"string"},"a":{"type":"number"}}},{"type":"null"}]}}},"required":["ok","reason"],"additionalProperties":false}}}}`)
	raw, _, err := Request("gpt-5.6-luna", body, Limits{PolicyProfile: PolicyGPTNative})
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	schema := out["text"].(map[string]any)["format"].(map[string]any)["schema"].(map[string]any)
	if !reflect.DeepEqual(schema["required"], []any{"ok", "reason", "impossible", "rows"}) {
		t.Errorf("strict root required=%v", schema["required"])
	}
	nested := schema["properties"].(map[string]any)["rows"].(map[string]any)["items"].(map[string]any)["anyOf"].([]any)[0].(map[string]any)
	if !reflect.DeepEqual(nested["required"], []any{"a", "z"}) || nested["additionalProperties"] != false {
		t.Errorf("nested schema=%v", nested)
	}
	for i := 0; i < 5; i++ {
		again, _, err := Request("gpt-5.6-luna", body, Limits{PolicyProfile: PolicyGPTNative})
		if err != nil || string(raw) != string(again) {
			t.Fatal("non-deterministic projection", err)
		}
	}
}

func TestGPTStrictOutputPreflightBounds(t *testing.T) {
	object := func(p map[string]any) map[string]any { return map[string]any{"type": "object", "properties": p} }
	propertySchema := func(n int) map[string]any {
		p := map[string]any{}
		for i := 0; i < n; i++ {
			p[fmt.Sprintf("p%d", i)] = map[string]any{"type": "string"}
		}
		return object(p)
	}
	nested := func(n int) map[string]any {
		s := object(map[string]any{})
		for i := 1; i < n; i++ {
			s = object(map[string]any{"x": s})
		}
		return s
	}
	enum := func(n, size int) map[string]any {
		v := make([]any, n)
		for i := range v {
			v[i] = fmt.Sprintf("%04d", i) + strings.Repeat("a", size-4)
		}
		return object(map[string]any{"x": map[string]any{"type": "string", "enum": v}})
	}
	for _, tc := range []struct {
		name   string
		schema map[string]any
		valid  bool
	}{
		{"properties5000", propertySchema(5000), true}, {"properties5001", propertySchema(5001), false},
		{"depth10", nested(10), true}, {"depth11", nested(11), false},
		{"chars120000", object(map[string]any{strings.Repeat("한", 120000): map[string]any{"type": "string"}}), true},
		{"chars120001", object(map[string]any{strings.Repeat("한", 120001): map[string]any{"type": "string"}}), false},
		{"enum1000", enum(1000, 4), true}, {"enum1001", enum(1001, 4), false},
		{"enum250Long", enum(250, 61), true}, {"enum251Long", enum(251, 60), false},
		{"rootArray", map[string]any{"type": "array", "items": map[string]any{"type": "string"}}, false},
		{"rootAnyOf", map[string]any{"type": "object", "anyOf": []any{object(map[string]any{})}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := strictOutputSchema(tc.schema, 0)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v err=%v", tc.valid, err)
			}
		})
	}
}

func TestGPTStrictOutputPreflightKeywordsAndData(t *testing.T) {
	for _, child := range []string{`{"type":"string","items":{"type":"string"}}`, `{"type":"number","pattern":"x"}`, `{"type":"string","properties":{}}`, `{"type":"string","required":[]}`, `{"type":"array"}`, `{"type":"string","minimum":0}`} {
		var value any
		if err := json.Unmarshal([]byte(child), &value); err != nil {
			t.Fatal(err)
		}
		if _, err := strictOutputSchema(map[string]any{"type": "object", "properties": map[string]any{"x": value}}, 0); err == nil {
			t.Errorf("type-incompatible schema accepted: %s", child)
		}
	}
	for _, key := range []string{"allOf", "oneOf", "not", "dependentRequired", "dependentSchemas", "if", "then", "else", "patternProperties", "propertyNames", "unevaluatedProperties", "contains", "uniqueItems", "prefixItems", "default", "examples"} {
		t.Run(key, func(t *testing.T) {
			s := map[string]any{"type": "object", "properties": map[string]any{}, key: []any{map[string]any{"type": "object"}}}
			if _, err := strictOutputSchema(s, 0); err == nil {
				t.Fatal("unsupported keyword accepted")
			}
		})
	}
	for _, raw := range []string{`{"type":"object","additionalProperties":true}`, `{"type":"object","additionalProperties":{"type":"string"}}`, `{"type":"object","properties":{"x":{"$ref":"https://example.invalid/schema"}}}`, `{"type":"object","properties":{"x":{"$ref":"#/$defs/missing"}}}`, `{"type":"object","properties":{"x":{"type":"string","format":"unsupported"}}}`} {
		var s any
		if err := json.Unmarshal([]byte(raw), &s); err != nil {
			t.Fatal(err)
		}
		if _, err := strictOutputSchema(s, 0); err == nil {
			t.Errorf("invalid schema accepted: %s", raw)
		}
	}
	raw := `{"type":"object","properties":{"node":{"$ref":"#/$defs/node"},"n":{"type":"number","minimum":0,"exclusiveMaximum":10,"multipleOf":0.5},"s":{"type":"string","pattern":"^[a-z]+$","format":"email"},"list":{"type":"array","minItems":1,"maxItems":4,"items":{"type":"string"}}},"$defs":{"node":{"anyOf":[{"type":"null"},{"type":"object","properties":{"next":{"$ref":"#/$defs/node"}}}]}},"const":{"allOf":"data not schema"}} `
	var s any
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	normalized, err := strictOutputSchema(s, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(normalized["const"], s.(map[string]any)["const"]) {
		t.Fatal("const data interpreted as schema")
	}
}

func TestGPTStrictOutputSchemaRejectsMalformedStructure(t *testing.T) {
	for _, schema := range []string{`{"type":7}`, `{"type":["object",7]}`, `{"type":[]}`, `{"type":"bogus"}`, `{"type":"object","additionalProperties":"false"}`} {
		var input any
		if err := json.Unmarshal([]byte(schema), &input); err != nil {
			t.Fatal(err)
		}
		if _, err := strictOutputSchema(input, 0); err == nil {
			t.Errorf("malformed schema accepted: %s", schema)
		}
	}
	for _, schema := range []string{`{"type":"object","properties":[]}`, `{"type":"object","properties":{"x":true}}`, `{"type":"object","properties":{},"required":["missing"]}`, `{"type":"object","properties":{},"required":"x"}`, `{"type":"object","properties":{"x":{"type":"array","items":7}}}`, `{"type":"object","properties":{"x":{"anyOf":{}}}}`} {
		body := []byte(`{"max_tokens":100,"messages":[{"role":"user","content":"check"}],"output_config":{"format":{"type":"json_schema","schema":` + schema + `}}}`)
		if _, _, err := Request("gpt-5.6-luna", body, Limits{PolicyProfile: PolicyGPTNative}); err == nil {
			t.Errorf("malformed schema accepted: %s", schema)
		}
	}
}
