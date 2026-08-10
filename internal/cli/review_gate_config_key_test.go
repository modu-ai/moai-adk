// Package cli — regression tests pinning the config key path both Stop-hook
// review gates read.
//
// Both readers previously unmarshalled a struct whose `codex:` / `multi:` keys
// sat at the TOP LEVEL of `.moai/config/sections/workflow.yaml`. That file
// nests every key under a `workflow:` root (config.workflowFileWrapper, which
// config.Loader.loadWorkflowSection uses, and the shape the shipped template
// deploys), so a user who set the documented
// `workflow.codex.review_gate.enabled` read false forever and the gate could
// never fire. These tests pin the nested path and pin that the flat form is
// NOT honoured — one spelling only; the flat form was simply wrong.
package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// writeGateWorkflowYAML writes body to <tmp>/.moai/config/sections/workflow.yaml
// and returns the project root. An empty body leaves the file absent.
func writeGateWorkflowYAML(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config sections: %v", err)
	}
	if body == "" {
		return dir
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
	return dir
}

// The two shapes under test. `nested` is the deployed shape (a `workflow:`
// root); `flat` is the shape the readers used to accept and must no longer.
const (
	nestedCodexOn  = "workflow:\n    codex:\n        review_gate:\n            enabled: true\n"
	nestedCodexOff = "workflow:\n    codex:\n        review_gate:\n            enabled: false\n"
	flatCodexOn    = "codex:\n    review_gate:\n        enabled: true\n"
	nestedMultiOn  = "workflow:\n    multi:\n        review_gate:\n            enabled: true\n"
	nestedMultiOff = "workflow:\n    multi:\n        review_gate:\n            enabled: false\n"
	flatMultiOn    = "multi:\n    review_gate:\n        enabled: true\n"
)

// TestReviewGateReaders_HonourNestedWorkflowKeyPath is the core regression: the
// nested path (the deployed shape) turns each gate ON, and every other shape —
// including the flat form the readers used to accept — reads OFF (fail-CLOSED).
func TestReviewGateReaders_HonourNestedWorkflowKeyPath(t *testing.T) {
	for _, tc := range []struct {
		name string
		read func(string) bool
		body string
		want bool
	}{
		// codex
		{"codex nested true", readCodexReviewGateEnabled, nestedCodexOn, true},
		{"codex nested false", readCodexReviewGateEnabled, nestedCodexOff, false},
		{"codex flat NOT honoured", readCodexReviewGateEnabled, flatCodexOn, false},
		{"codex sibling gate only", readCodexReviewGateEnabled, nestedMultiOn, false},
		{"codex missing file", readCodexReviewGateEnabled, "", false},
		{"codex malformed yaml", readCodexReviewGateEnabled, "workflow:\n\tcodex: [oops\n", false},
		// multi
		{"multi nested true", readMultiReviewGateEnabled, nestedMultiOn, true},
		{"multi nested false", readMultiReviewGateEnabled, nestedMultiOff, false},
		{"multi flat NOT honoured", readMultiReviewGateEnabled, flatMultiOn, false},
		{"multi sibling gate only", readMultiReviewGateEnabled, nestedCodexOn, false},
		{"multi missing file", readMultiReviewGateEnabled, "", false},
		{"multi malformed yaml", readMultiReviewGateEnabled, "workflow:\n\tmulti: [oops\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeGateWorkflowYAML(t, tc.body)
			if got := tc.read(dir); got != tc.want {
				t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
			}
		})
	}
	if readCodexReviewGateEnabled("") {
		t.Error("codex: empty projectDir ⇒ want false (fail-CLOSED)")
	}
	if readMultiReviewGateEnabled("") {
		t.Error("multi: empty projectDir ⇒ want false (fail-CLOSED)")
	}
}

// TestReviewGateReaders_AgreeWithConfigLoader is the drift guard. Each gate has
// a small hand-rolled reader (deliberate — see the comment at the reader) rather
// than calling config.Loader.Load, so this test pins that the hand-rolled shape
// still decodes to the same value the real loader produces from the same
// document. A schema change that moved the key would fail here.
func TestReviewGateReaders_AgreeWithConfigLoader(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{"both on", "workflow:\n    codex:\n        review_gate:\n            enabled: true\n    multi:\n        review_gate:\n            enabled: true\n"},
		{"both off", "workflow:\n    codex:\n        review_gate:\n            enabled: false\n    multi:\n        review_gate:\n            enabled: false\n"},
		{"codex only", nestedCodexOn},
		{"multi only", nestedMultiOn},
		{"absent", "workflow:\n    execution_mode: auto\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeGateWorkflowYAML(t, tc.body)
			cfg, err := config.NewLoader().Load(filepath.Join(dir, ".moai"))
			if err != nil {
				t.Fatalf("config load: %v", err)
			}
			if got, want := readCodexReviewGateEnabled(dir), cfg.Workflow.Codex.ReviewGate.Enabled; got != want {
				t.Errorf("codex reader = %v, config loader = %v (schema drift)", got, want)
			}
			if got, want := readMultiReviewGateEnabled(dir), cfg.Workflow.Multi.ReviewGate.Enabled; got != want {
				t.Errorf("multi reader = %v, config loader = %v (schema drift)", got, want)
			}
		})
	}
}
