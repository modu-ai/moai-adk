package cli

import (
	"os"
	"strings"
	"testing"
)

// TestNavigatorRouteHidden verifies the Route subcommand is registered on
// the root and is Hidden (NOT surfaced on `moai --help`). Mirrors the
// navigator-tiers Hidden assertion at navigator_tiers_test.go:66-76
// (AC-NS4-011a, REQ-NS4-011).
func TestNavigatorRouteHidden(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "navigator-route" {
			if !cmd.Hidden {
				t.Errorf("navigator-route is not Hidden (must mirror navigator-sync + navigator-tiers)")
			}
			return
		}
	}
	t.Errorf("navigator-route not registered on rootCmd")
}

// TestNavigatorRoute_NoAskUserQuestion is the C-HRA-008 / REQ-PGN-012
// subagent-boundary static guard: this CLI file MUST NOT call
// AskUserQuestion or mcp__askuser__* (canonical guard pattern per
// internal/cli/CLAUDE.md).
func TestNavigatorRoute_NoAskUserQuestion(t *testing.T) {
	b, err := os.ReadFile("navigator_route.go")
	if err != nil {
		t.Fatal(err)
	}
	body := string(b)
	for _, frag := range []string{"AskUserQuestion", "mcp__askuser"} {
		if strings.Contains(body, frag) {
			t.Errorf("navigator_route.go: forbidden %q reference (subagent boundary)", frag)
		}
	}
}
