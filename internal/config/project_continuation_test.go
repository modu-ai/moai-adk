package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeProjectContinuationWorkflowYAML writes a workflow.yaml carrying body
// under the .moai/config/sections/ tree of a fresh temp project and returns the
// .moai directory (the argument Loader.Load takes). It mirrors
// writeTodoWorkflowYAML rather than sharing it, so a later change to the todo
// helper cannot silently retarget these cases.
func writeProjectContinuationWorkflowYAML(t *testing.T, body string) string {
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

// TestValidProjectContinuations pins the closed set to exactly three tokens in
// order (AC-PCK-001 / REQ-PCK-001). It asserts the members rather than the
// length: a length check would pass for any three tokens, which is the mutant
// AC-PCK-001 exists to exclude.
func TestValidProjectContinuations(t *testing.T) {
	got := ValidProjectContinuations()
	want := []string{"none", "card", "pipeline"}

	if len(got) != len(want) {
		t.Fatalf("ValidProjectContinuations() length = %d, want %d (got %v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ValidProjectContinuations()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestProjectContinuation covers the resolution contract: absent ⇒ card with no
// unmatched report (REQ-PCK-002 / AC-PCK-002), an unmatched token ⇒ card WITH
// the offending value reported (REQ-PCK-003 / AC-PCK-003), and each in-domain
// token resolving to itself.
func TestProjectContinuation(t *testing.T) {
	cases := []struct {
		name          string
		body          string
		wantValue     string
		wantUnmatched string
	}{
		{
			name:          "project block absent",
			body:          "workflow:\n    default_mode: \"\"\n",
			wantValue:     ProjectContinuationCard,
			wantUnmatched: "",
		},
		{
			name:          "continuation key absent inside a present project block",
			body:          "workflow:\n    project: {}\n",
			wantValue:     ProjectContinuationCard,
			wantUnmatched: "",
		},
		{
			name:          "explicit none",
			body:          "workflow:\n    project:\n        continuation: none\n",
			wantValue:     ProjectContinuationNone,
			wantUnmatched: "",
		},
		{
			name:          "explicit card",
			body:          "workflow:\n    project:\n        continuation: card\n",
			wantValue:     ProjectContinuationCard,
			wantUnmatched: "",
		},
		{
			name:          "explicit pipeline",
			body:          "workflow:\n    project:\n        continuation: pipeline\n",
			wantValue:     ProjectContinuationPipeline,
			wantUnmatched: "",
		},
		{
			name:          "unmatched value falls back to card and is reported",
			body:          "workflow:\n    project:\n        continuation: pipelien\n",
			wantValue:     ProjectContinuationCard,
			wantUnmatched: "pipelien",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			moaiDir := writeProjectContinuationWorkflowYAML(t, tc.body)
			cfg, err := NewLoader().Load(moaiDir)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			value, unmatched := cfg.ProjectContinuation()
			if value != tc.wantValue {
				t.Errorf("ProjectContinuation() value = %q, want %q", value, tc.wantValue)
			}
			if unmatched != tc.wantUnmatched {
				t.Errorf("ProjectContinuation() unmatched = %q, want %q", unmatched, tc.wantUnmatched)
			}
		})
	}
}

// TestProjectContinuationNilReceiver pins the nil-receiver half of AC-PCK-002:
// a caller that could not load config resolves to card with nothing to report,
// and does not panic.
func TestProjectContinuationNilReceiver(t *testing.T) {
	var cfg *Config
	value, unmatched := cfg.ProjectContinuation()
	if value != ProjectContinuationCard {
		t.Errorf("nil *Config ProjectContinuation() value = %q, want %q", value, ProjectContinuationCard)
	}
	if unmatched != "" {
		t.Errorf("nil *Config ProjectContinuation() unmatched = %q, want empty", unmatched)
	}
}

// TestProjectContinuationForRoot covers the project-root convenience form,
// including the two failure paths that must resolve to card rather than error.
func TestProjectContinuationForRoot(t *testing.T) {
	t.Run("empty root", func(t *testing.T) {
		value, unmatched := ProjectContinuationForRoot("")
		if value != ProjectContinuationCard || unmatched != "" {
			t.Errorf("ProjectContinuationForRoot(\"\") = (%q, %q), want (%q, \"\")", value, unmatched, ProjectContinuationCard)
		}
	})

	t.Run("missing .moai tree", func(t *testing.T) {
		value, unmatched := ProjectContinuationForRoot(t.TempDir())
		if value != ProjectContinuationCard || unmatched != "" {
			t.Errorf("ProjectContinuationForRoot(missing) = (%q, %q), want (%q, \"\")", value, unmatched, ProjectContinuationCard)
		}
	})

	t.Run("explicit pipeline resolves through the root form", func(t *testing.T) {
		moaiDir := writeProjectContinuationWorkflowYAML(t, "workflow:\n    project:\n        continuation: pipeline\n")
		value, unmatched := ProjectContinuationForRoot(filepath.Dir(moaiDir))
		if value != ProjectContinuationPipeline || unmatched != "" {
			t.Errorf("ProjectContinuationForRoot(pipeline) = (%q, %q), want (%q, \"\")", value, unmatched, ProjectContinuationPipeline)
		}
	})
}
