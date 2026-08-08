package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// execution_modes_test.go pins the workflow.execution_mode closed set against
// the evidence that defines it, rather than against an inference.
//
// Background: the set was previously written as {auto, solo, team}. Nothing in
// the Go tree reads ExecutionMode, so the value's meaning is carried by prose —
// and the prose names a fourth mode. Once the console converted the field to a
// closed-set widget, a value outside the declared set became unsavable, so an
// omission here is not cosmetic: it silently removes a documented mode from the
// user's reach.

// TestExecutionModePinsMatchModeDefaults is the load-bearing assertion: the
// pinnable execution modes are exactly the keys of harness.mode_defaults.
//
// The two are the same concept seen from two sides — mode_defaults maps "each
// execution mode" to a harness level, and execution_mode picks which of those
// modes is in force (or `auto`, which defers the pick). If they drift, either
// the console refuses a mode the harness config knows about, or it offers one
// the harness cannot resolve a level for.
func TestExecutionModePinsMatchModeDefaults(t *testing.T) {
	pins := ValidExecutionModePins()
	if len(pins) == 0 {
		t.Fatal("ValidExecutionModePins() is empty")
	}

	got := map[string]bool{}
	for _, p := range pins {
		got[p] = true
	}
	// The shipped harness.yaml is the artifact users actually receive; read its
	// mode_defaults keys rather than restating them here.
	for _, key := range shippedModeDefaultKeys(t) {
		if !got[key] {
			t.Errorf("harness.yaml declares mode_defaults.%s but it is not a pinnable execution mode — a user cannot select it", key)
		}
	}
	if len(pins) != len(shippedModeDefaultKeys(t)) {
		t.Errorf("pin set %v and harness.yaml mode_defaults %v differ in size", pins, shippedModeDefaultKeys(t))
	}
}

// TestValidExecutionModesIsAutoPlusPins fixes the composition: `auto` (defer to
// harness auto-selection) plus every pinnable mode.
func TestValidExecutionModesIsAutoPlusPins(t *testing.T) {
	modes := ValidExecutionModes()
	if len(modes) == 0 || modes[0] != "auto" {
		t.Fatalf("ValidExecutionModes() = %v, want `auto` first (the defer-to-harness default)", modes)
	}
	if len(modes) != len(ValidExecutionModePins())+1 {
		t.Errorf("ValidExecutionModes() = %v, want auto + pins %v", modes, ValidExecutionModePins())
	}
	for _, pin := range ValidExecutionModePins() {
		found := false
		for _, m := range modes {
			if m == pin {
				found = true
			}
		}
		if !found {
			t.Errorf("pin %q is missing from ValidExecutionModes() = %v", pin, modes)
		}
	}
}

// TestExecutionModeDefaultIsInSet guards the config default against the set the
// console enforces: a default the widget would refuse is a trap.
func TestExecutionModeDefaultIsInSet(t *testing.T) {
	def := NewDefaultConfig()
	got := def.Workflow.ExecutionMode
	for _, m := range ValidExecutionModes() {
		if m == got {
			return
		}
	}
	t.Errorf("default execution_mode %q is outside ValidExecutionModes() = %v", got, ValidExecutionModes())
}

// shippedModeDefaultKeys reads the mode_defaults keys from the repo's
// harness.yaml. It parses the block by indentation rather than unmarshalling
// the whole harness config, so the test stays readable and does not depend on
// the loader it is meant to be independent of.
func shippedModeDefaultKeys(t *testing.T) []string {
	t.Helper()
	// Walk up from the package dir to the repo root (the dir holding go.mod).
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate the repo root (no go.mod found walking up)")
		}
		dir = parent
	}
	data, err := os.ReadFile(filepath.Join(dir, ".moai", "config", "sections", "harness.yaml"))
	if err != nil {
		t.Skipf("harness.yaml not readable (%v) — this assertion needs the shipped artifact", err)
	}

	var keys []string
	inBlock := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if trimmed == "mode_defaults:" {
			inBlock = true
			continue
		}
		if !inBlock {
			continue
		}
		// The block ends at the first line indented no deeper than `mode_defaults:`.
		if indentOf(line) <= 2 {
			break
		}
		key, _, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		keys = append(keys, strings.TrimSpace(key))
	}
	if len(keys) == 0 {
		t.Fatal("no mode_defaults keys parsed from harness.yaml — the comparison would be vacuous")
	}
	return keys
}

func indentOf(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}
