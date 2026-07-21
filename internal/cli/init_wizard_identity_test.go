package cli

// Focused coverage for the init wizard identity/locale wiring: the wizard's
// conversation_language and user_name answers must flow into opts (driving
// template deployment of language.yaml / user.yaml) and win over the profile
// fallback. This exercises runInit through the injectable runWizardFn seam.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

// TestInitWizardIdentityPersisted asserts the wizard's conversation_language and
// user_name answers land in the generated config files.
func TestInitWizardIdentityPersisted(t *testing.T) {
	// Isolate HOME so no real profile leaks in and the profile auto-prompt path
	// stays inert (stdin is not a TTY in tests → the huh.Confirm never runs).
	t.Setenv("HOME", t.TempDir())

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string, _, _ bool) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{
			ConversationLang: "ja",
			UserName:         "WizardName",
			ProjectName:      "ident-proj",
			ModelPolicy:      "high",
			ReportFormat:     "html+md",
			GitMode:          "manual",
		}, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })

	projectDir := filepath.Join(t.TempDir(), "ident-proj")
	cmd := newInitTestCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)

	if err := runInit(cmd, []string{projectDir}); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	langPath := filepath.Join(projectDir, ".moai", "config", "sections", "language.yaml")
	langData, err := os.ReadFile(langPath)
	if err != nil {
		t.Fatalf("read language.yaml: %v", err)
	}
	if !bytes.Contains(langData, []byte(`conversation_language: "ja"`)) {
		t.Errorf("language.yaml must carry the wizard locale 'ja', got:\n%s", langData)
	}

	userPath := filepath.Join(projectDir, ".moai", "config", "sections", "user.yaml")
	userData, err := os.ReadFile(userPath)
	if err != nil {
		t.Fatalf("read user.yaml: %v", err)
	}
	if !bytes.Contains(userData, []byte("WizardName")) {
		t.Errorf("user.yaml must carry the wizard user name 'WizardName', got:\n%s", userData)
	}
}
