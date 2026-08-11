package cli

// navigator_fix_test.go — M3.2 Hidden-subcommand assertion
// (SPEC-NAVIGATOR-SYNC-005 AC-NS5-011 / REQ-NS5-011). Mirrors
// navigator_tiers_test.go:65-70 (the canonical Hidden-registration guard)
// and adds the subagent-boundary static guard per internal/cli/CLAUDE.md.

import (
	"os"
	"strings"
	"testing"
)

// TestNavigatorFix_HiddenRegistration verifies AC-NS5-011(a): the navigator-fix
// cobra command is registered on rootCmd AND is Hidden (NOT surfaced on
// `moai --help`), mirroring navigator-sync / navigator-route / navigator-tiers.
func TestNavigatorFix_HiddenRegistration(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "navigator-fix" {
			if !cmd.Hidden {
				t.Errorf("navigator-fix is not Hidden (must mirror navigator-sync/route/tiers)")
			}
			return
		}
	}
	t.Errorf("navigator-fix not registered on rootCmd")
}

// TestNavigatorFix_NoAskUserQuestion is the C-HRA-008 / REQ-PGN-012
// subagent-boundary static guard: navigator_fix.go MUST NOT call
// AskUserQuestion or mcp__askuser__* (canonical guard pattern per
// internal/cli/CLAUDE.md, mirroring navigator_tiers_test.go:81-92).
func TestNavigatorFix_NoAskUserQuestion(t *testing.T) {
	b, err := os.ReadFile("navigator_fix.go")
	if err != nil {
		t.Fatal(err)
	}
	body := string(b)
	for _, frag := range []string{"AskUserQuestion", "mcp__askuser"} {
		if strings.Contains(body, frag) {
			t.Errorf("navigator_fix.go: forbidden %q reference (subagent boundary)", frag)
		}
	}
}

// TestNavigatorFix_NotTopLevelMoaiSubcommand verifies AC-NS5-011(b): the
// navigator-fix registration is a Hidden sibling subcommand, NOT a top-level
// user-facing `moai navigator fix` surface. The presence check above plus the
// Hidden flag satisfy this; this test makes the intent explicit by asserting
// there is no top-level `navigator` parent command that would surface fix as a
// user-facing verb.
func TestNavigatorFix_NotTopLevelMoaiSubcommand(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "navigator" {
			t.Errorf("a top-level \"navigator\" parent command exists; navigator-fix must be a Hidden sibling, not a child verb")
		}
	}
}
