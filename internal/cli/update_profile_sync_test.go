package cli

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// REQ-RC30-002 / AC-RC30-21: the profile sync of `moai update` resolves the
// project's profile from the launch ledger (projects[<root>] in launch.yaml,
// under profile.GetBaseDir()), not from the environment alone. CLAUDE_CONFIG_DIR
// is unset here, so GetCurrentName() can only answer "default"; the ledger maps
// this project to the "work" profile, whose conversation language is "ja".
func TestProfileSyncResolution(t *testing.T) {
	base, root := lpSandbox(t)
	writeTestFile(t, root, ".moai/config/sections/system.yaml", "moai:\n  version: v9.9.9-test\n")
	lpMkProfile(t, base, "work")
	if err := profile.WritePreferences("work", profile.ProfilePreferences{ConversationLang: "ja"}); err != nil {
		t.Fatalf("write work preferences: %v", err)
	}
	lpWriteLedger(t, base, "projects:\n  "+lpProjectKey(t, root)+": work\n")
	// Fixture check: the ledger resolves this project to "work" through the
	// project-scoped resolver, so a failure below is the update's resolution,
	// not a broken ledger.
	if got := profile.GetCurrentNameForProject(root); got != "work" {
		t.Fatalf("fixture: launch ledger does not resolve the work profile for the project: got %q", got)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })
	if err := os.Chdir(root); err != nil {
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

	// The run error is not asserted: the judged result is the language the
	// profile sync wrote into the project config.
	_ = updateCmd.RunE(updateCmd, []string{})

	cfg, err := config.NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("load project config: %v", err)
	}
	if got := cfg.Language.ConversationLanguage; got != "ja" {
		t.Errorf("profile sync did not resolve the project's ledger profile: conversation_language = %q, want %q (work profile)", got, "ja")
	}
}
