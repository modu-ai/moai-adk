package gateway

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestValidationShape(t *testing.T) {
	base := `{"model":"gpt-6-astra","max_tokens":1,"messages":[{"role":"user","content":"arbitrary text"}]}`
	cases := []struct {
		name, field string
		value       any
		want        bool
	}{
		{"absent stream", "", nil, true}, {"false stream", "stream", false, true},
		{"true stream", "stream", true, false}, {"string stream", "stream", "false", false}, {"null stream", "stream", nil, false},
		{"two tokens", "max_tokens", 2, false}, {"zero tokens", "max_tokens", 0, false},
		{"empty messages", "messages", []any{}, false}, {"two users", "messages", []any{map[string]any{"role": "user"}, map[string]any{"role": "user"}}, false},
		{"system message", "messages", []any{map[string]any{"role": "system"}}, false}, {"nonempty tools", "tools", []any{map[string]any{"name": "tool"}}, false},
		{"empty tools", "tools", []any{}, true}, {"null tools", "tools", nil, false}, {"object tools", "tools", map[string]any{}, false},
		{"null messages", "messages", nil, false}, {"string tokens", "max_tokens", "1", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var b map[string]any
			_ = json.Unmarshal([]byte(base), &b)
			if tc.field != "" {
				b[tc.field] = tc.value
			}
			raw, _ := json.Marshal(b)
			got, err := IsValidationRequest(raw)
			if err != nil || got != tc.want {
				t.Fatalf("got %v, %v; want %v", got, err, tc.want)
			}
		})
	}
	for _, raw := range []string{`{"max_tokens":1.0,"messages":[{"role":"user"}]}`, `{"max_tokens":1e0,"messages":[{"role":"user"}]}`, `{"messages":[{"role":"user"}]}`} {
		got, err := IsValidationRequest([]byte(raw))
		if err != nil || got {
			t.Fatalf("noninteger/missing tokens recognized: %s (%v)", raw, err)
		}
	}
}

func TestValidationRejectsAmbiguousJSON(t *testing.T) {
	for _, raw := range []string{`{`, `[]`, `null`, `{} {}`, `{"model":"a","model":"b"}`, `{"stream":true,"stream":false}`, `{"messages":[{"role":"system","role":"user"}]}`} {
		got, err := IsValidationRequest([]byte(raw))
		if err == nil || got {
			t.Fatalf("ambiguous JSON accepted: %s", raw)
		}
	}
}

func TestActualCapturedRequests(t *testing.T) {
	paths, err := filepath.Glob("testdata/validation-capture/claude-code-2.1.268/request-*.json")
	if err != nil || len(paths) != 5 {
		t.Fatalf("capture count %d: %v", len(paths), err)
	}
	for _, p := range paths {
		t.Run(filepath.Base(p), func(t *testing.T) {
			raw, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			got, err := IsValidationRequest(raw)
			want := filepath.Base(p) == "request-003.json"
			if err != nil || got != want {
				t.Fatalf("got %v,%v want %v", got, err, want)
			}
		})
	}
}
