package cli

// init_docs_guard_test.go — AC-021 (a) (SPEC-INIT-SHRINK-001 REQ-021): the
// static grep-guard over the user-facing init surfaces. A static check
// proves the tokens are present, not that behavior holds — the behavior
// arms live in init_mode_test.go (AC-001..AC-007) beside it.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitDocsDescribeThinDeploy asserts the init flag help and the success
// card name the thin deploy and both paths: a plugin-mode user can tell
// where skills and commands live, and an opt-out user can find --no-plugin.
func TestInitDocsDescribeThinDeploy(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("init.go"))
	if err != nil {
		t.Fatalf("read init.go: %v", err)
	}
	text := string(src)

	// The --no-plugin flag help names the full local payload AND the
	// plugin-mode default.
	noPluginHelp := extractFlagHelp(t, text, `"no-plugin"`)
	for _, want := range []string{"FULL local payload", "ride the moai plugin"} {
		if !strings.Contains(noPluginHelp, want) {
			t.Errorf("--no-plugin flag help missing %q:\n%s", want, noPluginHelp)
		}
	}

	// The --all flag help names the local full deploy.
	allHelp := extractFlagHelp(t, text, `"all"`)
	if !strings.Contains(allHelp, "full local deploy") {
		t.Errorf("--all flag help does not name the local full deploy:\n%s", allHelp)
	}

	// The success card names the deploy mode and both paths (the builder
	// carries the mode-aware line; asserted live by init_mode_test.go).
	card, err := os.ReadFile(filepath.Join("init_warnings.go"))
	if err != nil {
		t.Fatalf("read init_warnings.go: %v", err)
	}
	cardText := string(card)
	for _, want := range []string{
		"Deploy mode: plugin",
		"--no-plugin for a full local deploy",
		"Deploy mode: local",
	} {
		if !strings.Contains(cardText, want) {
			t.Errorf("success card missing %q", want)
		}
	}
}

// extractFlagHelp returns the help string of one flag registration line.
func extractFlagHelp(t *testing.T, src, flagLiteral string) string {
	t.Helper()
	for _, line := range strings.Split(src, "\n") {
		if strings.Contains(line, "Flags()") && strings.Contains(line, flagLiteral) {
			return line
		}
	}
	t.Fatalf("flag registration for %s not found", flagLiteral)
	return ""
}
