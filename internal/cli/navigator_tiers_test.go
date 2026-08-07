package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNavigatorTiers_EmitsOverlayWhenProjectPopulated verifies the CLI surface
// for the 4-tier overlay step end-to-end: a populated project root yields
// tiers.json with the six required top-level keys (provenance + 4 tier slices
// + tier_edges). Mirrors TestNavigatorSync_EmitsNavGraphWhenInputsPresent.
func TestNavigatorTiers_EmitsOverlayWhenProjectPopulated(t *testing.T) {
	root := t.TempDir()
	// A minimal authored module tree so Tier 1 has at least one blueprint.
	navTiersWriteFile(t, filepath.Join(root, ".moai", "project", "blueprint", "module_tree.json"),
		`{"modules":[{"package_path":"internal/x","display_name":"X","layer":"domain","responsibility":"x","depends_on":[]}]}`)
	navTiersWriteFile(t, filepath.Join(root, "internal", "x", "x.go"), "package x\nfunc F() {}\n")

	if err := runNavigatorTiers(root); err != nil {
		t.Fatalf("runNavigatorTiers error: %v", err)
	}
	out := filepath.Join(root, ".moai", "project", "navigator", "tiers.json")
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read tiers.json: %v", err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{
		"provenance",
		"tier0_contracts",
		"tier1_blueprints",
		"tier2_decisions",
		"tier3_symbols",
		"tier_edges",
	} {
		if _, ok := doc[key]; !ok {
			t.Errorf("tiers.json missing key %q", key)
		}
	}
}

// TestNavigatorTiers_AbsenceIsGraceful verifies REQ-NS3-020 fail-open at the
// CLI surface: when every input is absent, runNavigatorTiers returns nil
// (exit 0) and tiers.json is STILL emitted (overlay layer emits an empty
// overlay rather than aborting — mirrors the package-level Enrich contract).
func TestNavigatorTiers_AbsenceIsGraceful(t *testing.T) {
	root := t.TempDir()
	if err := runNavigatorTiers(root); err != nil {
		t.Fatalf("runNavigatorTiers returned non-nil under all-absent: %v", err)
	}
	out := filepath.Join(root, ".moai", "project", "navigator", "tiers.json")
	if _, err := os.Stat(out); err != nil {
		t.Errorf("tiers.json not emitted under all-absent (fail-open contract): %v", err)
	}
}

// TestNavigatorTiers_HiddenRegistration verifies the subcommand is registered
// on the root and is Hidden (NOT surfaced on `moai --help`). Mirrors the
// navigator-sync Hidden contract.
func TestNavigatorTiers_HiddenRegistration(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "navigator-tiers" {
			if !cmd.Hidden {
				t.Errorf("navigator-tiers is not Hidden (must mirror navigator-sync)")
			}
			return
		}
	}
	t.Errorf("navigator-tiers not registered on rootCmd")
}

// TestNavigatorTiers_NoAskUserQuestion is the C-HRA-008 / REQ-PGN-012
// subagent-boundary static guard: this CLI file MUST NOT call AskUserQuestion
// or mcp__askuser__* (canonical guard pattern per internal/cli/CLAUDE.md).
func TestNavigatorTiers_NoAskUserQuestion(t *testing.T) {
	b, err := os.ReadFile("navigator_tiers.go")
	if err != nil {
		t.Fatal(err)
	}
	body := string(b)
	for _, frag := range []string{"AskUserQuestion", "mcp__askuser"} {
		if strings.Contains(body, frag) {
			t.Errorf("navigator_tiers.go: forbidden %q reference (subagent boundary)", frag)
		}
	}
}

func navTiersWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
