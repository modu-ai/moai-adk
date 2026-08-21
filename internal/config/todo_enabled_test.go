package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTodoWorkflowYAML writes a workflow.yaml carrying body under the
// .moai/config/sections/ tree of a fresh temp project and returns the .moai
// directory (the argument Loader.Load takes).
func writeTodoWorkflowYAML(t *testing.T, body string) string {
	t.Helper()
	moaiDir := filepath.Join(t.TempDir(), ".moai")
	sections := filepath.Join(moaiDir, "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if body != "" {
		if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(body), 0o644); err != nil {
			t.Fatalf("write workflow.yaml: %v", err)
		}
	}
	return moaiDir
}

// TestTodoEnabled covers the four key-interpretation cases of REQ-1
// (AC-T-001). The first case proves the *bool requirement: an ABSENT key is
// not the same as an explicit false. The fourth case travels the
// loadWorkflowSection section-wide fallback path (loader.go warns and returns,
// leaving the construction-time defaults in place) — this SPEC neither created
// nor changes that blast radius; it only records that the todo key lands on
// the enabled side of it.
func TestTodoEnabled(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "todo block absent",
			body: "workflow:\n    default_mode: \"\"\n",
			want: true,
		},
		{
			name: "explicit false",
			body: "workflow:\n    todo:\n        enabled: false\n",
			want: false,
		},
		{
			name: "explicit true",
			body: "workflow:\n    todo:\n        enabled: true\n",
			want: true,
		},
		{
			name: "non-bool value falls back to the section defaults",
			body: "workflow:\n    todo:\n        enabled: maybe\n",
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			moaiDir := writeTodoWorkflowYAML(t, tc.body)
			cfg, err := NewLoader().Load(moaiDir)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if got := cfg.TodoEnabled(); got != tc.want {
				t.Fatalf("TodoEnabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestTodoEnabledNilConfig pins the nil-receiver fail-open: a caller that
// could not load config gets the enabled default rather than a panic.
func TestTodoEnabledNilConfig(t *testing.T) {
	var cfg *Config
	if !cfg.TodoEnabled() {
		t.Fatal("nil *Config TodoEnabled() = false, want true (fail-open)")
	}
}

// TestTodoEnabledForRoot exercises the project-root convenience reader the
// runtime surfaces call: an absent tree and an unreadable one both resolve to
// enabled; an explicit false resolves to disabled.
func TestTodoEnabledForRoot(t *testing.T) {
	if !TodoEnabledForRoot("") {
		t.Fatal("empty root = disabled, want enabled (fail-open)")
	}
	if !TodoEnabledForRoot(t.TempDir()) {
		t.Fatal("root without .moai = disabled, want enabled (fail-open)")
	}

	moaiDir := writeTodoWorkflowYAML(t, "workflow:\n    todo:\n        enabled: false\n")
	root := filepath.Dir(moaiDir)
	if TodoEnabledForRoot(root) {
		t.Fatal("root with todo.enabled=false = enabled, want disabled")
	}
}
