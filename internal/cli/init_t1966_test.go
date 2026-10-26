package cli

// Card t1966 — `moai init` wizard/flag regression tests (items 1, 2, 7).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// Item 1: the already-initialized probe must run BEFORE the wizard, so an
// initialized project gets the `moai update` guidance before any question.
func TestRunInit_AlreadyInitializedProbeRunsBeforeWizard(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	var calls int
	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		calls++
		return &wizard.WizardResult{}, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "config"), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := newInitTestCmd()
	err := runInit(cmd, []string{projectDir})
	if err == nil {
		t.Fatal("runInit on an initialized project: want the already-initialized error, got nil")
	}
	if !strings.Contains(err.Error(), "moai update") {
		t.Errorf("error lacks the `moai update` guidance: %v", err)
	}
	if calls != 0 {
		t.Errorf("wizard ran %d time(s) before the already-initialized probe, want 0", calls)
	}
}

// Item 2: --bundles catalog validation lives in validateInitFlags (PreRunE).
func TestValidateInitFlags_UnknownBundleFailsEarly(t *testing.T) {
	cmd := newInitTestCmd()
	if err := cmd.Flags().Set("bundles", "no-such-bundle-t1966"); err != nil {
		t.Fatal(err)
	}
	err := validateInitFlags(cmd, nil)
	if err == nil {
		t.Fatal("validateInitFlags accepted an unknown bundle name")
	}
	if !strings.Contains(err.Error(), "no-such-bundle-t1966") || !strings.Contains(err.Error(), "--bundles") {
		t.Errorf("error should name the flag and the unknown bundle: %v", err)
	}
}

func TestValidateInitFlags_KnownBundleAccepted(t *testing.T) {
	cmd := newInitTestCmd()
	if err := cmd.Flags().Set("bundles", "backend"); err != nil {
		t.Fatal(err)
	}
	if err := validateInitFlags(cmd, nil); err != nil {
		t.Errorf("known bundle rejected: %v", err)
	}
}

// Item 5: development_mode is not a wizard question on the init path (TDD is
// the silent default; --mode is the only override), so a wizard result field
// must not reach the persisted config. The former dead mapping let it through.
func TestRunInit_WizardDevelopmentModeDoesNotReachConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{ProjectName: "devmode-proj", DevelopmentMode: "ddd", GitMode: "manual"}, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	projectDir := filepath.Join(t.TempDir(), "devmode-proj")
	if err := runInit(newInitTestCmd(), []string{projectDir}); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(projectDir, ".moai", "config", "sections", "quality.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var mode string
	for _, line := range strings.Split(string(data), "\n") {
		if k, v, ok := strings.Cut(strings.TrimSpace(line), ":"); ok && k == "development_mode" {
			mode = strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	if mode != "tdd" {
		t.Errorf("quality.yaml development_mode = %q, want %q (silent default; wizard result must not set it)", mode, "tdd")
	}
}

// Item 7: the --enable-lsp registration default must match its help text
// ("default: true"), like --enforce-quality.
func TestInitCmd_EnableLSPRegisteredDefaultMatchesHelp(t *testing.T) {
	f := initCmd.Flags().Lookup("enable-lsp")
	if f == nil {
		t.Fatal("--enable-lsp not registered")
	}
	if f.DefValue != "true" {
		t.Errorf("--enable-lsp DefValue = %q, want \"true\" (help says default: true)", f.DefValue)
	}
}

// Item 3 follow-up (coordinator P2): a profile already holds a user name; the
// wizard answer is an explicit clear. The cleared name must reach user.yaml —
// runInit and syncInitProfile must not re-save the profile's name.
func TestRunInit_ClearedUserNameDoesNotResaveProfileName(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := profile.WritePreferences(profile.GetCurrentName(), profile.ProfilePreferences{UserName: "Profile Alice"}); err != nil {
		t.Fatalf("seed profile: %v", err)
	}

	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{UserNameCleared: true, GitMode: "manual"}, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	projectDir := filepath.Join(t.TempDir(), "cleared-name-proj")
	if err := runInit(newInitTestCmd(), []string{projectDir}); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(projectDir, ".moai", "config", "sections", "user.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "Profile Alice") {
		t.Errorf("user.yaml re-saved the profile name after the wizard cleared it:\n%s", data)
	}
}
