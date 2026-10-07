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

// Public help must describe the supported project/user split, not send users
// back to the retired plugin carrier. Runtime bundle behavior is tested beside it.
func TestInitDocsDescribeUserAssets(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("init.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	noPluginHelp := extractFlagHelp(t, text, `"no-plugin"`)
	if !strings.Contains(noPluginHelp, "Deprecated") {
		t.Errorf("legacy no-plugin flag must disclose retirement: %s", noPluginHelp)
	}
	allHelp := extractFlagHelp(t, text, `"all"`)
	for _, want := range []string{"project harness", "--bundles"} {
		if !strings.Contains(allHelp, want) {
			t.Errorf("--all help missing %q: %s", want, allHelp)
		}
	}
	var notice strings.Builder
	emitSlimModeNotice(&notice)
	for _, want := range []string{"user folders", "--bundles"} {
		if !strings.Contains(notice.String(), want) {
			t.Errorf("notice missing %q: %s", want, notice.String())
		}
	}
	for _, retired := range []string{"ride the moai plugin", "--no-plugin or --all for a full local deploy"} {
		if strings.Contains(notice.String(), retired) {
			t.Errorf("retired plugin guidance remains: %s", notice.String())
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
