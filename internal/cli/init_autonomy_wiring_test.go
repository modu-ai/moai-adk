package cli

// SPEC-INIT-WIZARD-REPAIR-001 chain ① runInit wiring tests (M1): the wizard's
// autonomy-tier answer and the --autonomy-tier flag both reach
// opts.AutonomyTier, and the ApplyAutonomyTierBundle consumer runs after the
// initializer returns — with the USER-scope write landing in a temp HOME and
// splicing only the permissions block (spec.md §4 key-scoped constraint).

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
)

// runInitForAutonomyAtHome runs a full init in a temp project with HOME pinned
// to homeDir (the caller owns the t.Setenv). A non-nil wizardResult swaps the
// isInteractiveStdin + runWizardFn seams so the interactive branch runs
// without a TTY.
func runInitForAutonomyAtHome(t *testing.T, homeDir string, wizResult *wizard.WizardResult, flags map[string]string) (projectDir string) {
	t.Helper()
	projectDir, _ = runInitForAutonomyAtHomeCapturingOut(t, homeDir, wizResult, flags)
	return projectDir
}

// runInitForAutonomyAtHomeCapturingOut is runInitForAutonomyAtHome plus the
// captured stdout (SPEC-INIT-HARNESS-PROMPT-001 M2.0). The buffer was already
// built and bound here; it was simply discarded, which left the .mcp.json
// provisioning announcement — the only user-visible signal of whether the
// ensure-entry call ran (acceptance.md §A.1) — unobservable from a test.
//
// It is a SIBLING rather than a signature change: runInitForAutonomyAtHome
// delegates to it, so the 11 existing invocation sites are untouched AND no
// setup is duplicated. The plan.md M2.0 default was "sibling unless it would
// duplicate meaningful setup"; delegation satisfies both halves at once.
func runInitForAutonomyAtHomeCapturingOut(t *testing.T, homeDir string, wizResult *wizard.WizardResult, flags map[string]string) (projectDir, stdout string) {
	t.Helper()

	if wizResult != nil {
		origInteractive := isInteractiveStdin
		isInteractiveStdin = func() bool { return true }
		t.Cleanup(func() { isInteractiveStdin = origInteractive })

		origDeps := deps
		deps = nil
		t.Cleanup(func() { deps = origDeps })

		origWizard := runWizardFn
		runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) { return wizResult, nil }
		t.Cleanup(func() { runWizardFn = origWizard })
	}

	projectDir = filepath.Join(t.TempDir(), "tier-proj")
	cmd := newInitTestCmd()
	for name, val := range flags {
		if err := cmd.Flags().Set(name, val); err != nil {
			t.Fatalf("set --%s=%s: %v", name, val, err)
		}
	}
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)

	if err := runInit(cmd, []string{projectDir}); err != nil {
		t.Fatalf("runInit: %v (stderr: %s)", err, errBuf.String())
	}
	return projectDir, out.String()
}

// runInitForAutonomy runs a full init in a temp project. HOME is isolated via
// t.Setenv so the USER-scope settings write (and ensureGlobalSettingsEnv)
// always lands under a temp dir — never the real ~/.claude.
func runInitForAutonomy(t *testing.T, wizResult *wizard.WizardResult, flags map[string]string) (projectDir, homeDir string) {
	t.Helper()
	homeDir = t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")
	return runInitForAutonomyAtHome(t, homeDir, wizResult, flags), homeDir
}

// userSettingsDefaultMode reads permissions.defaultMode from the USER-scope
// settings.json under homeDir. Returns "" when the file or key is absent.
func userSettingsDefaultMode(t *testing.T, homeDir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(homeDir, ".claude", "settings.json"))
	if err != nil {
		return ""
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse USER settings.json: %v\nraw: %s", err, data)
	}
	perms, _ := doc["permissions"].(map[string]any)
	mode, _ := perms["defaultMode"].(string)
	return mode
}

// TestRunInit_WizardAutonomyTierAppliesBundle asserts AC-001 (re-scoped by
// SPEC-USER-SETTINGS-PROTECT-001, decision Q1): an interactive init whose
// injected wizard result carries AutonomyTier=automatic (flag absent) deploys
// the tier bundle through the USER-scope writer — which PRESERVES an existing
// permissions.defaultMode (a present value means the user owns the setting;
// REQ-002) and writes the tier default only when the key is absent (REQ-001).
// Both dispositions are KEY-SCOPED splices that preserve unrelated
// pre-existing keys verbatim (spec.md §4).
func TestRunInit_WizardAutonomyTierAppliesBundle(t *testing.T) {
	wiz := &wizard.WizardResult{AutonomyTier: config.AutonomyTierAutomatic}

	t.Run("existing-defaultMode-preserved", func(t *testing.T) {
		homeDir := t.TempDir()
		t.Setenv("HOME", homeDir)
		t.Setenv("MOAI_SANDBOX_PROOF", "")
		t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

		// Pre-existing USER settings carrying an unrelated key, an env block,
		// and the user's own defaultMode: the Q1 keep-rule preserves the
		// user's value — the tier default is NOT written over it.
		userSettings := `{
  "teammateMode": "glm",
  "env": {
    "MOAI_FLAG": "keep-me"
  },
  "permissions": {
    "defaultMode": "default"
  }
}
`
		if err := os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		userPath := filepath.Join(homeDir, ".claude", "settings.json")
		if err := os.WriteFile(userPath, []byte(userSettings), 0o644); err != nil {
			t.Fatal(err)
		}

		runInitForAutonomyAtHomeCapturingOut(t, homeDir, wiz, nil)

		// Q1 keep-rule: the user's defaultMode survives the tier application.
		if mode := userSettingsDefaultMode(t, homeDir); mode != "default" {
			t.Errorf("USER permissions.defaultMode = %q, want %q preserved (Q1 keep-rule: the user owns an existing defaultMode; SPEC-USER-SETTINGS-PROTECT-001)", mode, "default")
		}

		// Key-scoped splice (spec.md §4): the unrelated pre-existing key and
		// the env block survive the write verbatim — whole-file overwrite is
		// prohibited.
		data, err := os.ReadFile(userPath)
		if err != nil {
			t.Fatalf("read USER settings.json: %v", err)
		}
		if !bytes.Contains(data, []byte(`"teammateMode": "glm"`)) {
			t.Errorf("unrelated pre-existing key teammateMode was not preserved verbatim; got:\n%s", data)
		}
		if !bytes.Contains(data, []byte(`"MOAI_FLAG": "keep-me"`)) {
			t.Errorf("unrelated env block was not preserved verbatim; got:\n%s", data)
		}
	})

	t.Run("absent-defaultMode-applies-tier", func(t *testing.T) {
		homeDir := t.TempDir()
		t.Setenv("HOME", homeDir)
		t.Setenv("MOAI_SANDBOX_PROOF", "")
		t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

		// USER settings WITHOUT a defaultMode key (the permissions block has
		// only an allow list): the tier default IS written here, and the
		// user's allow entries survive the write (REQ-001).
		userSettings := `{
  "teammateMode": "glm",
  "permissions": {
    "allow": ["Bash(make:*)"]
  }
}
`
		if err := os.MkdirAll(filepath.Join(homeDir, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		userPath := filepath.Join(homeDir, ".claude", "settings.json")
		if err := os.WriteFile(userPath, []byte(userSettings), 0o644); err != nil {
			t.Fatal(err)
		}

		runInitForAutonomyAtHomeCapturingOut(t, homeDir, wiz, nil)

		// The tier bundle wrote defaultMode=auto into the USER scope (the
		// key was absent — the tier default applies).
		if mode := userSettingsDefaultMode(t, homeDir); mode != "auto" {
			t.Errorf("USER permissions.defaultMode = %q, want %q (tier automatic applied on the absent-key path)", mode, "auto")
		}
		data, err := os.ReadFile(userPath)
		if err != nil {
			t.Fatalf("read USER settings.json: %v", err)
		}
		if !bytes.Contains(data, []byte(`"Bash(make:*)"`)) {
			t.Errorf("the user's allow entry was not preserved across the tier write; got:\n%s", data)
		}
		if !bytes.Contains(data, []byte(`"teammateMode": "glm"`)) {
			t.Errorf("unrelated pre-existing key teammateMode was not preserved verbatim; got:\n%s", data)
		}
	})
}

// TestRunInit_FlagAutonomyTierNonInteractive asserts AC-003: a non-interactive
// init with an explicit --autonomy-tier=automatic (no TTY, no wizard) still
// reaches the bundle apply — the validated flag value is no longer discarded.
func TestRunInit_FlagAutonomyTierNonInteractive(t *testing.T) {
	projectDir, homeDir := runInitForAutonomy(t, nil, map[string]string{
		"autonomy-tier": config.AutonomyTierAutomatic,
	})
	if mode := userSettingsDefaultMode(t, homeDir); mode != "auto" {
		t.Errorf("USER permissions.defaultMode = %q, want %q (flag automatic, non-interactive)", mode, "auto")
	}
	_ = projectDir
}

// TestRunInit_SemiAutoAndEmptyAreBoundedDelta asserts REQ-004 of
// SPEC-AUT-PERMMODES-001 (re-scoped from AC-004's zero delta): wizard answers
// of "" and "semi-auto" (flag absent) produce the BOUNDED delta — the USER
// settings.json carries permissions.defaultMode="acceptEdits", the two runs
// stay byte-identical to each other, and the PROJECT settings.json are
// byte-equal across runs.
func TestRunInit_SemiAutoAndEmptyAreBoundedDelta(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")
	userPath := filepath.Join(homeDir, ".claude", "settings.json")

	projEmpty := runInitForAutonomyAtHome(t, homeDir, &wizard.WizardResult{AutonomyTier: ""}, nil)
	afterEmpty := readSettingsFileOrNull(t, userPath)

	projSemi := runInitForAutonomyAtHome(t, homeDir, &wizard.WizardResult{AutonomyTier: config.AutonomyTierSemiAuto}, nil)
	afterSemi := readSettingsFileOrNull(t, userPath)

	if !bytes.Equal(afterEmpty, afterSemi) {
		t.Errorf("USER settings.json differs between empty and semi-auto wizard answers (bounded delta violated):\n--empty--\n%s\n--semi-auto--\n%s", afterEmpty, afterSemi)
	}
	// The bounded delta: both selections MUST write USER
	// permissions.defaultMode="acceptEdits" — and that is the ONLY permission
	// key either selection may add.
	for name, got := range map[string][]byte{"empty": afterEmpty, "semi-auto": afterSemi} {
		if got == nil {
			t.Errorf("%s selection must write USER settings.json (bounded delta requires the defaultMode record); got no file", name)
			continue
		}
		var doc map[string]any
		if err := json.Unmarshal(got, &doc); err != nil {
			t.Fatalf("parse USER settings.json (%s): %v", name, err)
		}
		perms, ok := doc["permissions"].(map[string]any)
		if !ok {
			t.Errorf("%s selection must write a permissions block (bounded delta); got:\n%s", name, got)
			continue
		}
		if mode, _ := perms["defaultMode"].(string); mode != "acceptEdits" {
			t.Errorf("%s selection must write permissions.defaultMode=acceptEdits (bounded delta); got %q:\n%s", name, mode, got)
		}
	}

	projA := readSettingsFile(t, filepath.Join(projEmpty, ".claude", "settings.json"))
	projB := readSettingsFile(t, filepath.Join(projSemi, ".claude", "settings.json"))
	if !bytes.Equal(projA, projB) {
		t.Errorf("PROJECT settings.json differs between empty and semi-auto wizard answers:\n--empty--\n%s\n--semi-auto--\n%s", projA, projB)
	}
}

// TestRunInit_FlagFullyAutonomousWithoutProofDowngrades asserts AC-002: an
// explicit --autonomy-tier=fully-autonomous with no sandbox proof is gated
// down to automatic — the USER scope receives the automatic defaultMode and
// the downgrade advisory lands in the project's autonomy-downgrade log.
func TestRunInit_FlagFullyAutonomousWithoutProofDowngrades(t *testing.T) {
	projectDir, homeDir := runInitForAutonomy(t, nil, map[string]string{
		"autonomy-tier": config.AutonomyTierFullyAutonomous,
	})

	if mode := userSettingsDefaultMode(t, homeDir); mode != "auto" {
		t.Errorf("USER permissions.defaultMode = %q, want %q (downgraded to automatic)", mode, "auto")
	}
	advPath := filepath.Join(projectDir, ".moai", "logs", "autonomy-downgrade.log")
	adv, err := os.ReadFile(advPath)
	if err != nil {
		t.Fatalf("downgrade advisory log missing (%v): %s", err, advPath)
	}
	if !bytes.Contains(adv, []byte(config.AutonomyTierFullyAutonomous)) || !bytes.Contains(adv, []byte(config.AutonomyTierAutomatic)) {
		t.Errorf("advisory must name the %s→%s downgrade; got:\n%s", config.AutonomyTierFullyAutonomous, config.AutonomyTierAutomatic, adv)
	}
}

// readSettingsFile is a fatal-on-missing settings reader (the deployer always
// renders the PROJECT settings.json, so absence is a defect).
func readSettingsFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

// readSettingsFileOrNull reads a settings file, returning nil when absent.
func readSettingsFileOrNull(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}
