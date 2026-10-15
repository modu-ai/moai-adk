package cli

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/pkg/version"
)

// REQ-RC30-013 / AC-RC30-18: a default `moai update` run prints exactly one
// version banner, and the identity band (REQ-TUXIU-015) is that banner.
func TestUpdateSingleBanner(t *testing.T) {
	const testVersion = "v9.9.9-banner"
	origVersion := version.Version
	version.Version = testVersion
	t.Cleanup(func() { version.Version = origVersion })

	projectRoot := t.TempDir()
	writeTestFile(t, projectRoot, ".moai/config/sections/system.yaml", "moai:\n  version: "+testVersion+"\n")
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}

	origDeps := deps
	t.Cleanup(func() { deps = origDeps })
	deps = &Dependencies{UpdateChecker: &mockUpdateChecker{}}

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)
	updateCmd.SetContext(context.Background())
	if err := updateCmd.Flags().Set("check", "false"); err != nil {
		t.Fatal(err)
	}
	if err := updateCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = updateCmd.Flags().Set("yes", "false") })

	// The template sync may fail in the test environment. The judged output is
	// everything printed before any such failure, so the run error is not asserted.
	_ = updateCmd.RunE(updateCmd, []string{})

	plain := stripSGR(buf.String())
	var versionLines []string
	for _, line := range strings.Split(plain, "\n") {
		if strings.Contains(line, testVersion) {
			versionLines = append(versionLines, line)
		}
	}
	if len(versionLines) != 1 {
		t.Errorf("a run must print exactly one version line, got %d: %q", len(versionLines), versionLines)
	}
	if !strings.Contains(plain, "◆ MoAI-ADK") {
		t.Errorf("the identity band must be kept (REQ-TUXIU-015)")
	}
}
