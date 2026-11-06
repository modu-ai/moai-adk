package cli

// update_template_sync_harness_gate_test.go — audit N1 regression (card t1957,
// GH #1827): the settings.local.json PATH-refresh gate must consume the
// PRE-DEPLOY harness capture, because the Deploy Templates stage rewrites
// llm.yaml to the template default ("claude"); a post-deploy re-read makes
// the codex-only guard inert and an update would create .claude/** paths in
// a gpt project (SPEC-INIT-HARNESS-001 REQ-IH-005).

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
)

// TestRunTemplateSyncGPTProjectCreatesNoClaudeSettings runs a real template
// sync against a harness=gpt project fixture and asserts the flow leaves no
// .claude/settings.local.json behind. The deploy genuinely rewrites llm.yaml
// to the template default in this scenario, so this test is red exactly when
// the gate re-reads the harness after the deploy instead of using the
// captured pre-deploy value.
func TestRunTemplateSyncGPTProjectCreatesNoClaudeSettings(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("charmbracelet/bubbletea blocks on Windows console ReadConsole in non-TTY CI")
	}
	tmpDir := t.TempDir()

	// A gpt project whose recorded template version forces the sync to run.
	sectionsDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sectionsDir, "llm.yaml"), []byte("llm:\n  harness: gpt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	systemYAML := "system:\n  template_version: \"0.0.0\"\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, "system.yaml"), []byte(systemYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(tmpDir, ".moai", "manifest.json")
	if err := os.WriteFile(manifestPath, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	t.Setenv("HOME", tmpDir)

	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("config", false, "")
	_ = cmd.Flags().Set("yes", "true")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetContext(context.Background())

	if _, err := runTemplateSyncWithProgress(cmd, true); err != nil {
		t.Fatalf("unexpected sync error: %v", err)
	}

	// The committed settings.json keeps SmartPATH-only and the machine-local
	// PATH lives in settings.local.json — but ONLY for Claude-harness
	// projects. A gpt project carries zero .claude/** paths.
	if _, err := os.Stat(filepath.Join(tmpDir, ".claude", "settings.local.json")); !os.IsNotExist(err) {
		t.Errorf("gpt project must not get .claude/settings.local.json from an update (stat err=%v)", err)
	}
}
