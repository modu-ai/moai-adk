package config

import (
	"os"
	"path/filepath"
	"testing"
)

// loader_crosssession_test.go — the crosssession.yaml section (cross-session
// messaging preferences the moai launchers translate into injected --settings).
//
// The section is a partial direct-read (consumed by internal/cli launchers on
// demand, not by the Loader.Load chain) — the lsp/mx/security pattern — so its
// entry point is the standalone LoadCrossSessionConfig, mirroring the
// loadHandoffSection partial-override contract: a yaml specifying a subset of
// keys retains the construction-time default for the omitted keys rather than
// collapsing them to zero-values.

// TestLoadCrossSessionConfigDefaults verifies the absent-file contract: a
// project without crosssession.yaml gets the neutral defaults (inbound unset
// so Claude Code's per-message ladder decides; machine isolation OFF so
// cross-machine messages need no approval — the documented default posture).
func TestLoadCrossSessionConfigDefaults(t *testing.T) {
	dir := t.TempDir() // no crosssession.yaml

	got, err := LoadCrossSessionConfig(dir)
	if err != nil {
		t.Fatalf("LoadCrossSessionConfig(%q) error: %v", dir, err)
	}
	if got.Inbound != "" {
		t.Errorf("Inbound = %q, want \"\" (unset → runtime per-message default)", got.Inbound)
	}
	if got.IsolateMachines {
		t.Errorf("IsolateMachines = true, want false (no-approval default)")
	}
	if got.DialogExpiry != "" {
		t.Errorf("DialogExpiry = %q, want \"\" (unset → Claude Code 5m default)", got.DialogExpiry)
	}
}

// TestLoadCrossSessionConfigPartialOverride verifies the partial-override
// contract: a file specifying only `inbound` keeps the defaults for
// isolate_machines and dialog_expiry.
func TestLoadCrossSessionConfigPartialOverride(t *testing.T) {
	dir := t.TempDir()
	writeCrossSessionYAML(t, dir, "crosssession:\n  inbound: accept\n")

	got, err := LoadCrossSessionConfig(dir)
	if err != nil {
		t.Fatalf("LoadCrossSessionConfig error: %v", err)
	}
	if got.Inbound != "accept" {
		t.Errorf("Inbound = %q, want \"accept\"", got.Inbound)
	}
	if got.IsolateMachines {
		t.Errorf("IsolateMachines = true, want false (omitted key keeps default)")
	}
	if got.DialogExpiry != "" {
		t.Errorf("DialogExpiry = %q, want \"\" (omitted key keeps default)", got.DialogExpiry)
	}
}

// TestLoadCrossSessionConfigFull verifies every key round-trips.
func TestLoadCrossSessionConfigFull(t *testing.T) {
	dir := t.TempDir()
	writeCrossSessionYAML(t, dir, "crosssession:\n  inbound: refuse\n  isolate_machines: true\n  dialog_expiry: never\n")

	got, err := LoadCrossSessionConfig(dir)
	if err != nil {
		t.Fatalf("LoadCrossSessionConfig error: %v", err)
	}
	if got.Inbound != "refuse" {
		t.Errorf("Inbound = %q, want \"refuse\"", got.Inbound)
	}
	if !got.IsolateMachines {
		t.Errorf("IsolateMachines = false, want true")
	}
	if got.DialogExpiry != "never" {
		t.Errorf("DialogExpiry = %q, want \"never\"", got.DialogExpiry)
	}
}

// TestCrossSessionConfigDefaultsFactory verifies the Defaults factory matches
// the template-shipped crosssession.yaml (neutral posture).
func TestCrossSessionConfigDefaultsFactory(t *testing.T) {
	got := NewDefaultCrossSessionConfig()
	if got.Inbound != "" || got.IsolateMachines || got.DialogExpiry != "" {
		t.Errorf("NewDefaultCrossSessionConfig() = %+v, want the neutral defaults", got)
	}
}

// TestCrossSessionRegistryEntry pins the yaml↔struct parity registry: the
// section file has a Go struct, so it is a registry entry (not an audit
// exception).
func TestCrossSessionRegistryEntry(t *testing.T) {
	if got := yamlToStructRegistry["crosssession"]; got != "CrossSessionConfig" {
		t.Errorf("yamlToStructRegistry[\"crosssession\"] = %q, want \"CrossSessionConfig\"", got)
	}
}

func writeCrossSessionYAML(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "crosssession.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write crosssession.yaml: %v", err)
	}
}
