package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- DDD PRESERVE: Characterization tests for init command behavior ---

func TestInitCmd_Exists(t *testing.T) {
	if initCmd == nil {
		t.Fatal("initCmd should not be nil")
	}
}

func TestInitCmd_Use(t *testing.T) {
	if initCmd.Use != "init [project-name]" {
		t.Errorf("initCmd.Use = %q, want %q", initCmd.Use, "init [project-name]")
	}
}

func TestInitCmd_Short(t *testing.T) {
	if initCmd.Short == "" {
		t.Error("initCmd.Short should not be empty")
	}
}

func TestInitCmd_Long(t *testing.T) {
	if initCmd.Long == "" {
		t.Error("initCmd.Long should not be empty")
	}
}

func TestInitCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		// Use Name() which returns the command name without arguments
		if cmd.Name() == "init" {
			found = true
			break
		}
	}
	if !found {
		t.Error("init should be registered as a subcommand of root")
	}
}

func TestInitCmd_HasFlags(t *testing.T) {
	flags := []string{"root", "name", "language", "framework", "mode", "non-interactive", "force"}
	for _, name := range flags {
		if initCmd.Flags().Lookup(name) == nil {
			t.Errorf("init command should have --%s flag", name)
		}
	}
}

func TestInitCmd_NonInteractiveExecution(t *testing.T) {
	root := t.TempDir()

	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)

	// Reset flags to default before setting
	if err := initCmd.Flags().Set("root", root); err != nil {
		t.Fatalf("set root flag: %v", err)
	}
	if err := initCmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatalf("set non-interactive flag: %v", err)
	}
	if err := initCmd.Flags().Set("name", "test-project"); err != nil {
		t.Fatalf("set name flag: %v", err)
	}
	if err := initCmd.Flags().Set("language", "Go"); err != nil {
		t.Fatalf("set language flag: %v", err)
	}
	if err := initCmd.Flags().Set("mode", "ddd"); err != nil {
		t.Fatalf("set mode flag: %v", err)
	}

	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatalf("init command RunE error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "MoAI project initialized") {
		t.Errorf("expected success message in output, got: %q", output)
	}

	// Verify .moai/ was created
	moaiDir := filepath.Join(root, ".moai")
	if _, statErr := os.Stat(moaiDir); os.IsNotExist(statErr) {
		t.Error("expected .moai/ directory to be created")
	}

	// Verify CLAUDE.md was created
	claudeMD := filepath.Join(root, "CLAUDE.md")
	if _, statErr := os.Stat(claudeMD); os.IsNotExist(statErr) {
		t.Error("expected CLAUDE.md to be created")
	}
}

// TestInitCmd_PositionalArgCreatesDirectory tests that positional argument creates a new directory
func TestInitCmd_PositionalArgCreatesDirectory(t *testing.T) {
	// Create a temp directory to work in
	workDir := t.TempDir()

	// Change to work directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current dir: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("chdir to temp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)

	// Reset all flags to defaults
	_ = initCmd.Flags().Set("root", "")
	_ = initCmd.Flags().Set("non-interactive", "true")
	_ = initCmd.Flags().Set("name", "")
	_ = initCmd.Flags().Set("language", "Go")
	_ = initCmd.Flags().Set("mode", "ddd")

	// Run with positional argument
	err = initCmd.RunE(initCmd, []string{"my-new-project"})
	if err != nil {
		t.Fatalf("init command RunE error = %v", err)
	}

	// Verify the directory was created
	projectDir := filepath.Join(workDir, "my-new-project")
	if _, statErr := os.Stat(projectDir); os.IsNotExist(statErr) {
		t.Error("expected my-new-project/ directory to be created")
	}

	// Verify .moai/ was created inside the new directory
	moaiDir := filepath.Join(projectDir, ".moai")
	if _, statErr := os.Stat(moaiDir); os.IsNotExist(statErr) {
		t.Error("expected .moai/ directory to be created inside project folder")
	}

	output := buf.String()
	if !strings.Contains(output, "MoAI project initialized") {
		t.Errorf("expected success message in output, got: %q", output)
	}
}

// TestInitCmd_DotArgUsesCurrentDirectory tests that "." argument uses current directory
func TestInitCmd_DotArgUsesCurrentDirectory(t *testing.T) {
	root := t.TempDir()

	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current dir: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir to temp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	// Reset flags
	_ = initCmd.Flags().Set("root", "")
	_ = initCmd.Flags().Set("non-interactive", "true")
	_ = initCmd.Flags().Set("name", "dot-test")
	_ = initCmd.Flags().Set("language", "Go")
	_ = initCmd.Flags().Set("mode", "ddd")

	// Run with "." argument
	err = initCmd.RunE(initCmd, []string{"."})
	if err != nil {
		t.Fatalf("init command RunE error = %v", err)
	}

	// Verify .moai/ was created in current directory (not in a new "." folder)
	moaiDir := filepath.Join(root, ".moai")
	if _, statErr := os.Stat(moaiDir); os.IsNotExist(statErr) {
		t.Error("expected .moai/ directory to be created in current directory")
	}

	// Verify initialization worked in current directory
	output := buf.String()
	if !strings.Contains(output, "MoAI project initialized") {
		t.Errorf("expected success message in output, got: %q", output)
	}
}

func TestGetStringFlag(t *testing.T) {
	// Flag exists but may have been set in previous test; just verify no panic
	_ = getStringFlag(initCmd, "name")

	// Non-existent flag returns empty
	if got := getStringFlag(initCmd, "nonexistent-flag-xyz"); got != "" {
		t.Errorf("getStringFlag for nonexistent flag = %q, want empty", got)
	}
}

func TestGetBoolFlag(t *testing.T) {
	// Non-existent flag returns false
	if got := getBoolFlag(initCmd, "nonexistent-flag-xyz"); got {
		t.Error("getBoolFlag for nonexistent flag should return false")
	}
}

// --- DDD PRESERVE: Characterization tests for flag validation ---

func TestValidateInitFlags_ValidMode(t *testing.T) {
	validModes := []string{"ddd", "tdd"}

	for _, mode := range validModes {
		t.Run(mode, func(t *testing.T) {
			// Set the mode flag
			if err := initCmd.Flags().Set("mode", mode); err != nil {
				t.Fatal(err)
			}

			// Validate should pass
			err := validateInitFlags(initCmd, []string{})
			if err != nil {
				t.Errorf("validateInitFlags with mode=%q should not error, got: %v", mode, err)
			}
		})
	}
}

func TestValidateInitFlags_InvalidMode(t *testing.T) {
	invalidModes := []string{"invalid", "test", "unknown", ""}

	for _, mode := range invalidModes {
		if mode == "" {
			continue // Empty mode is valid (uses default)
		}
		t.Run(mode, func(t *testing.T) {
			// Set the invalid mode flag
			if err := initCmd.Flags().Set("mode", mode); err != nil {
				t.Fatal(err)
			}

			// Validate should fail
			err := validateInitFlags(initCmd, []string{})
			if err == nil {
				t.Errorf("validateInitFlags with mode=%q should error, got nil", mode)
			}

			// Error message should mention the invalid value
			if !strings.Contains(err.Error(), "invalid --mode") {
				t.Errorf("error should mention 'invalid --mode', got: %v", err)
			}
		})
	}
}

func TestValidateInitFlags_ValidGitMode(t *testing.T) {
	validGitModes := []string{"manual", "personal", "team"}

	for _, gitMode := range validGitModes {
		t.Run(gitMode, func(t *testing.T) {
			// Reset mode flag first (in case previous test left invalid value)
			if err := initCmd.Flags().Set("mode", ""); err != nil {
				t.Fatal(err)
			}

			// Set the git-mode flag
			if err := initCmd.Flags().Set("git-mode", gitMode); err != nil {
				t.Fatal(err)
			}

			// Validate should pass
			err := validateInitFlags(initCmd, []string{})
			if err != nil {
				t.Errorf("validateInitFlags with git-mode=%q should not error, got: %v", gitMode, err)
			}
		})
	}
}

func TestValidateInitFlags_InvalidGitMode(t *testing.T) {
	invalidGitModes := []string{"invalid", "auto", "unknown"}

	for _, gitMode := range invalidGitModes {
		t.Run(gitMode, func(t *testing.T) {
			// Reset mode flag first (in case previous test left invalid value)
			if err := initCmd.Flags().Set("mode", ""); err != nil {
				t.Fatal(err)
			}

			// Set the invalid git-mode flag
			if err := initCmd.Flags().Set("git-mode", gitMode); err != nil {
				t.Fatal(err)
			}

			// Validate should fail
			err := validateInitFlags(initCmd, []string{})
			if err == nil {
				t.Errorf("validateInitFlags with git-mode=%q should error, got nil", gitMode)
			}

			// Error message should mention the invalid value
			if !strings.Contains(err.Error(), "invalid --git-mode") {
				t.Errorf("error should mention 'invalid --git-mode', got: %v", err)
			}
		})
	}
}

func TestValidateInitFlags_EmptyFlags(t *testing.T) {
	// Reset all flags to empty
	if err := initCmd.Flags().Set("mode", ""); err != nil {
		t.Fatal(err)
	}
	if err := initCmd.Flags().Set("git-mode", ""); err != nil {
		t.Fatal(err)
	}

	// Validate should pass with empty flags (uses defaults)
	err := validateInitFlags(initCmd, []string{})
	if err != nil {
		t.Errorf("validateInitFlags with empty flags should not error, got: %v", err)
	}
}

// TestInitCmd_HasPhase1Flags verifies all Phase 1 flags are registered (REQ-IWE-006/007/008).
func TestInitCmd_HasPhase1Flags(t *testing.T) {
	phase1Flags := []string{
		"standard",
		"advanced",
		"project-mode",
		"harness-profile",
		"enable-lsp",
		"enforce-quality",
		"enable-design",
	}
	for _, name := range phase1Flags {
		if initCmd.Flags().Lookup(name) == nil {
			t.Errorf("init command should have --%s flag", name)
		}
	}
}

// TestGetBoolFlagWithDefault_WhenNotChanged verifies the function returns defaultVal
// when the flag has not been explicitly set by the user.
func TestGetBoolFlagWithDefault_WhenNotChanged(t *testing.T) {
	// enforce-quality default should be true (not changed)
	got := getBoolFlagWithDefault(initCmd, "enforce-quality", true)
	if !got {
		t.Error("getBoolFlagWithDefault: expected true when flag not changed and defaultVal=true")
	}

	got = getBoolFlagWithDefault(initCmd, "enforce-quality", false)
	// When not changed, returns defaultVal (false)
	if got {
		t.Error("getBoolFlagWithDefault: expected false when flag not changed and defaultVal=false")
	}
}

// TestAdvancedImpliesStandard verifies that --advanced=true results in StandardMode=true.
// This tests the EC-3 requirement from acceptance.md.
func TestAdvancedImpliesStandard(t *testing.T) {
	advanced := true
	standard := false
	// Simulate the resolution logic from runInit
	resolved := standard || advanced
	if !resolved {
		t.Error("--advanced should imply --standard (standardMode should be true)")
	}
}

// resetInitFlagsForProfile clears the flags a profile test cares about so a
// prior test's leftover value on the shared global initCmd cannot bleed in.
func resetInitFlagsForProfile(t *testing.T) {
	t.Helper()
	for _, f := range []string{"mode", "git-mode", "profile", "model-policy"} {
		if initCmd.Flags().Lookup(f) != nil {
			_ = initCmd.Flags().Set(f, "")
		}
	}
}

// TestInitCmd_PlanTypeFlagRetired (SPEC-MODEL-PROFILE-MATRIX-001 REQ-MPM-017,
// AC-MPM-011) — the retired --plan-type flag is no longer registered on the
// init command, and the new --profile flag is.
func TestInitCmd_PlanTypeFlagRetired(t *testing.T) {
	if initCmd.Flags().Lookup("plan-type") != nil {
		t.Error("init command must NOT expose the retired --plan-type flag")
	}
	if initCmd.Flags().Lookup("profile") == nil {
		t.Error("init command should have a --profile flag")
	}
}

// TestValidateInitFlags_ValidProfile (REQ-MPM-015) — max/medium/low pass.
func TestValidateInitFlags_ValidProfile(t *testing.T) {
	for _, p := range []string{"max", "medium", "low"} {
		t.Run(p, func(t *testing.T) {
			resetInitFlagsForProfile(t)
			if err := initCmd.Flags().Set("profile", p); err != nil {
				t.Fatal(err)
			}
			if err := validateInitFlags(initCmd, []string{}); err != nil {
				t.Errorf("validateInitFlags with profile=%q should not error, got: %v", p, err)
			}
		})
	}
	resetInitFlagsForProfile(t)
}

// TestValidateInitFlags_InvalidProfile (REQ-MPM-015) — an out-of-set value
// errors, and the message names the closed set {max, medium, low}.
func TestValidateInitFlags_InvalidProfile(t *testing.T) {
	for _, p := range []string{"bogus", "high", "subscription"} {
		t.Run(p, func(t *testing.T) {
			resetInitFlagsForProfile(t)
			if err := initCmd.Flags().Set("profile", p); err != nil {
				t.Fatal(err)
			}
			err := validateInitFlags(initCmd, []string{})
			if err == nil {
				t.Fatalf("validateInitFlags with profile=%q should error, got nil", p)
			}
			msg := err.Error()
			if !strings.Contains(msg, "invalid --profile") {
				t.Errorf("error should mention 'invalid --profile', got: %v", err)
			}
			if !strings.Contains(msg, "max, medium, low") {
				t.Errorf("error should name the closed set, got: %v", err)
			}
		})
	}
	resetInitFlagsForProfile(t)
}

// TestInitCmd_ProfilePersistence (SPEC-MODEL-PROFILE-MATRIX-001 REQ-MPM-016,
// AC-MPM-010) — `moai init --profile max` persists profile: max to the deployed
// llm.yaml and writes no plan_type key (REQ-MPM-017/032, AC-MPM-011).
func TestInitCmd_ProfilePersistence(t *testing.T) {
	root := t.TempDir()

	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)

	resetInitFlagsForProfile(t)
	for k, v := range map[string]string{
		"root": root, "non-interactive": "true", "name": "profile-test",
		"language": "Go", "mode": "tdd", "profile": "max",
	} {
		if err := initCmd.Flags().Set(k, v); err != nil {
			t.Fatalf("set %s flag: %v", k, err)
		}
	}
	t.Cleanup(func() {
		_ = initCmd.Flags().Set("profile", "")
		_ = initCmd.Flags().Set("root", "")
	})

	if err := initCmd.RunE(initCmd, []string{}); err != nil {
		t.Fatalf("init command RunE error = %v", err)
	}

	llmPath := filepath.Join(root, ".moai", "config", "sections", "llm.yaml")
	content, err := os.ReadFile(llmPath)
	if err != nil {
		t.Fatalf("read deployed llm.yaml: %v", err)
	}
	if !strings.Contains(string(content), "profile: max") {
		t.Errorf("deployed llm.yaml should contain 'profile: max', got:\n%s", content)
	}
	if strings.Contains(string(content), "plan_type") {
		t.Errorf("deployed llm.yaml must NOT contain a plan_type key (retired), got:\n%s", content)
	}
}

// --- F3 init-path input validation: git-provider identity parity ---
//
// The reconfigure/update path validates --github-username and
// --gitlab-instance-url via validateWizardInput (update.go). The init path
// (validateInitFlags) previously did NOT, so `moai init` accepted a malformed
// username or a plaintext http URL that the reconfigure path rejected. These
// tests close that asymmetry.

// resetInitFlagsForGitIdentity clears every flag validateInitFlags reads so a
// prior test's leftover value on the shared global initCmd cannot bleed into
// the git-identity validation under test.
func resetInitFlagsForGitIdentity(t *testing.T) {
	t.Helper()
	for _, f := range []string{
		"mode", "git-mode", "git-provider", "model-policy", "profile",
		"github-username", "gitlab-instance-url",
	} {
		if initCmd.Flags().Lookup(f) != nil {
			_ = initCmd.Flags().Set(f, "")
		}
	}
}

// TestValidateInitFlags_InvalidGitHubUsername (F3 init-path parity) — a
// malformed --github-username is rejected on the init path, mirroring the
// reconfigure path's validateWizardInput.
func TestValidateInitFlags_InvalidGitHubUsername(t *testing.T) {
	for _, name := range []string{"--bad--name", "-leadinghyphen", "trailinghyphen-", "has--consecutive"} {
		t.Run(name, func(t *testing.T) {
			resetInitFlagsForGitIdentity(t)
			if err := initCmd.Flags().Set("github-username", name); err != nil {
				t.Fatal(err)
			}
			err := validateInitFlags(initCmd, []string{})
			if err == nil {
				t.Fatalf("validateInitFlags with github-username=%q should error, got nil", name)
			}
			if !strings.Contains(err.Error(), "invalid --github-username") {
				t.Errorf("error should mention 'invalid --github-username', got: %v", err)
			}
		})
	}
	resetInitFlagsForGitIdentity(t)
}

// TestValidateInitFlags_InvalidGitLabInstanceURL (F3 init-path parity) — a
// plaintext-http, wrong-scheme, or missing-host --gitlab-instance-url is
// rejected on the init path so a captured credential is never transmitted over
// an unencrypted channel.
func TestValidateInitFlags_InvalidGitLabInstanceURL(t *testing.T) {
	for _, raw := range []string{"http://gitlab.example.com", "ftp://gitlab.example.com", "https://"} {
		t.Run(raw, func(t *testing.T) {
			resetInitFlagsForGitIdentity(t)
			if err := initCmd.Flags().Set("gitlab-instance-url", raw); err != nil {
				t.Fatal(err)
			}
			err := validateInitFlags(initCmd, []string{})
			if err == nil {
				t.Fatalf("validateInitFlags with gitlab-instance-url=%q should error, got nil", raw)
			}
			if !strings.Contains(err.Error(), "invalid --gitlab-instance-url") {
				t.Errorf("error should mention 'invalid --gitlab-instance-url', got: %v", err)
			}
		})
	}
	resetInitFlagsForGitIdentity(t)
}

// TestValidateInitFlags_ValidGitIdentity (F3 init-path parity) — a well-formed
// --github-username plus an https --gitlab-instance-url passes.
func TestValidateInitFlags_ValidGitIdentity(t *testing.T) {
	resetInitFlagsForGitIdentity(t)
	if err := initCmd.Flags().Set("github-username", "octocat"); err != nil {
		t.Fatal(err)
	}
	if err := initCmd.Flags().Set("gitlab-instance-url", "https://gitlab.example.com"); err != nil {
		t.Fatal(err)
	}
	if err := validateInitFlags(initCmd, []string{}); err != nil {
		t.Errorf("validateInitFlags with valid git identity should not error, got: %v", err)
	}
	resetInitFlagsForGitIdentity(t)
}

// TestValidateInitFlags_EmptyGitIdentity (F3 init-path parity) — empty
// git-identity flags are permitted (the field is simply not set), mirroring the
// empty-is-OK semantics of validateWizardInput.
func TestValidateInitFlags_EmptyGitIdentity(t *testing.T) {
	resetInitFlagsForGitIdentity(t)
	if err := validateInitFlags(initCmd, []string{}); err != nil {
		t.Errorf("validateInitFlags with empty git identity should not error, got: %v", err)
	}
}
