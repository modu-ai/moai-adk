package cli

import (
	"os"
	"slices"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestParseCompanionLabelRecognizesWithoutConsuming is the load-bearing property
// of the `--name` recognition: moai learns the label, and claude still receives
// the flag. Consuming it would name the session nothing at all.
func TestParseCompanionLabelRecognizesWithoutConsuming(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
		want string
	}{
		{"long form", []string{"--name", "plan-tjlgt1"}, "plan-tjlgt1"},
		{"short form", []string{"-n", "run-tjlgt1"}, "run-tjlgt1"},
		{"long equals form", []string{"--name=review-tjlgt1"}, "review-tjlgt1"},
		{"short equals form", []string{"-n=sync-tjlgt1"}, "sync-tjlgt1"},
		{"among other flags", []string{"-p", "dev", "--name", "sync-abc123", "--verbose"}, "sync-abc123"},

		{"absent", []string{"-p", "dev"}, ""},
		{"no value", []string{"--name"}, ""},
		{"non-companion name", []string{"--name", "oauth-migration"}, ""},
		{"unknown role", []string{"--name", "lead-tjlgt1"}, ""},
		{"empty args", nil, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			before := slices.Clone(c.args)

			got, ok := parseCompanionLabel(c.args)
			if got != c.want || ok != (c.want != "") {
				t.Errorf("parseCompanionLabel(%v) = (%q, %v), want (%q, %v)",
					before, got, ok, c.want, c.want != "")
			}
			if !slices.Equal(c.args, before) {
				t.Errorf("args mutated: %v -> %v (--name must reach claude unchanged)", before, c.args)
			}
		})
	}
}

// TestParseCompanionLabelStopsAtPassThroughMarker asserts the `--` discipline
// shared with parseKanbanFlag, stripSpawnFlag, parseProfileFlag and
// normalizeWorktreeFlag: nothing past the marker is read.
func TestParseCompanionLabelStopsAtPassThroughMarker(t *testing.T) {
	t.Parallel()

	if got, ok := parseCompanionLabel([]string{"--", "--name", "plan-tjlgt1"}); ok {
		t.Errorf("read %q from beyond the pass-through marker", got)
	}
	// A label BEFORE the marker is still recognized.
	if got, ok := parseCompanionLabel([]string{"--name", "plan-tjlgt1", "--", "-n", "run-zzz"}); !ok || got != "plan-tjlgt1" {
		t.Errorf("parseCompanionLabel = (%q, %v), want (\"plan-tjlgt1\", true)", got, ok)
	}
}

// TestEnterKanbanModeSetsRunID asserts the lead mints a run id, and that the
// restore returns every variable to its PRIOR PRESENCE — an unset variable is
// unset again, not set to "".
//
// Non-parallel by construction: os.Setenv mutates process-global state.
func TestEnterKanbanModeSetsRunID(t *testing.T) {
	for _, key := range []string{config.EnvMoaiKanban, config.EnvMoaiKanbanID, config.EnvMoaiKanbanSpec} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}

	restore := enterKanbanMode("SPEC-EXAMPLE-001", "")

	runID := os.Getenv(config.EnvMoaiKanbanID)
	if runID == "" {
		t.Fatalf("%s not set by enterKanbanMode", config.EnvMoaiKanbanID)
	}
	if _, _, ok := kanban.SplitCompanionLabel(kanban.CompanionLabel("plan", runID)); !ok {
		t.Errorf("run id %q does not produce a parseable companion label", runID)
	}
	if os.Getenv(config.EnvMoaiKanban) != "1" {
		t.Errorf("%s not set", config.EnvMoaiKanban)
	}

	restore()
	for _, key := range []string{config.EnvMoaiKanban, config.EnvMoaiKanbanID, config.EnvMoaiKanbanSpec} {
		if _, present := os.LookupEnv(key); present {
			t.Errorf("%s still present after restore (prior presence not restored)", key)
		}
	}
}

// TestEnterKanbanCompanionModeSetsLabelNotKanban is the separation this whole
// change rests on: a companion takes the raised block cap but must NOT be seeded
// with the chain, or four sessions each drive the whole chain.
func TestEnterKanbanCompanionModeSetsLabelNotKanban(t *testing.T) {
	for _, key := range []string{config.EnvMoaiKanban, config.EnvMoaiKanbanID, config.EnvMoaiKanbanLabel} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}

	restore := enterKanbanCompanionMode("review-tjlgt1")

	if got := os.Getenv(config.EnvMoaiKanbanLabel); got != "review-tjlgt1" {
		t.Errorf("%s = %q, want %q", config.EnvMoaiKanbanLabel, got, "review-tjlgt1")
	}
	if got := os.Getenv(config.EnvMoaiKanbanID); got != "tjlgt1" {
		t.Errorf("%s = %q, want %q (derived from the label, never carried separately)",
			config.EnvMoaiKanbanID, got, "tjlgt1")
	}
	if _, present := os.LookupEnv(config.EnvMoaiKanban); present {
		t.Errorf("%s must NOT be set on a companion (it seeds the chain)", config.EnvMoaiKanban)
	}

	restore()
	for _, key := range []string{config.EnvMoaiKanbanID, config.EnvMoaiKanbanLabel} {
		if _, present := os.LookupEnv(key); present {
			t.Errorf("%s still present after restore", key)
		}
	}
}
