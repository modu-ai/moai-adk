package cli

// doctor_settings_defaultmode_test.go — card t1247.
//
// The check owns exactly one dead value: permissions.defaultMode=
// "bypassPermissions" in project/local scope. The cases below fix that
// boundary — acceptEdits/default/absent files stay OK, and "auto" (subject
// to the same scope rule but out of card t1247's scope) also stays OK so
// widening the set later is a deliberate test change, not a silent drift.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// writeSettingsDefaultModeFixture writes a settings file carrying the given
// defaultMode under <root>/.claude/<name>. mode == "" omits the key entirely.
func writeSettingsDefaultModeFixture(t *testing.T, root, name, mode string) {
	t.Helper()
	dir := filepath.Join(root, defs.ClaudeDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	perms := "{}"
	if mode != "" {
		perms = `{"defaultMode":` + `"` + mode + `"}`
	}
	body := `{"permissions":` + perms + `}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestDoctorSettingsDefaultMode(t *testing.T) {
	t.Run("project settings bypassPermissions warns and names the file", func(t *testing.T) {
		root := t.TempDir()
		writeSettingsDefaultModeFixture(t, root, defs.SettingsJSON, "bypassPermissions")

		got := checkSettingsDefaultMode(root, false)
		if got.Name != settingsDefaultModeCheckName {
			t.Errorf("Name = %q, want %q (moai doctor --check filters by exact name equality)", got.Name, settingsDefaultModeCheckName)
		}
		if got.Status != uikit.CheckWarn {
			t.Errorf("project bypass: status = %q, want %q", got.Status, uikit.CheckWarn)
		}
		if !strings.Contains(got.Message, defs.SettingsJSON) {
			t.Errorf("message should name the offending file %q, got %q", defs.SettingsJSON, got.Message)
		}
	})

	t.Run("local settings bypassPermissions warns and names the file", func(t *testing.T) {
		root := t.TempDir()
		writeSettingsDefaultModeFixture(t, root, defs.SettingsLocalJSON, "bypassPermissions")

		got := checkSettingsDefaultMode(root, false)
		if got.Status != uikit.CheckWarn {
			t.Errorf("local bypass: status = %q, want %q", got.Status, uikit.CheckWarn)
		}
		if !strings.Contains(got.Message, defs.SettingsLocalJSON) {
			t.Errorf("message should name the offending file %q, got %q", defs.SettingsLocalJSON, got.Message)
		}
	})

	t.Run("both files offending names both", func(t *testing.T) {
		root := t.TempDir()
		writeSettingsDefaultModeFixture(t, root, defs.SettingsJSON, "bypassPermissions")
		writeSettingsDefaultModeFixture(t, root, defs.SettingsLocalJSON, "bypassPermissions")

		got := checkSettingsDefaultMode(root, false)
		if got.Status != uikit.CheckWarn {
			t.Errorf("both bypass: status = %q, want %q", got.Status, uikit.CheckWarn)
		}
		if !strings.Contains(got.Message, defs.SettingsJSON) || !strings.Contains(got.Message, defs.SettingsLocalJSON) {
			t.Errorf("message should name both files, got %q", got.Message)
		}
	})

	t.Run("acceptEdits is OK in project/local scope", func(t *testing.T) {
		root := t.TempDir()
		writeSettingsDefaultModeFixture(t, root, defs.SettingsJSON, "acceptEdits")

		if got := checkSettingsDefaultMode(root, false); got.Status != uikit.CheckOK {
			t.Errorf("acceptEdits: status = %q, want %q", got.Status, uikit.CheckOK)
		}
	})

	t.Run("auto stays OK — detection is scoped to bypassPermissions (card t1247)", func(t *testing.T) {
		root := t.TempDir()
		writeSettingsDefaultModeFixture(t, root, defs.SettingsJSON, "auto")

		if got := checkSettingsDefaultMode(root, false); got.Status != uikit.CheckOK {
			t.Errorf("auto: status = %q, want %q (widening the flagged set is a deliberate decision, not silent drift)", got.Status, uikit.CheckOK)
		}
	})

	t.Run("absent defaultMode key is OK", func(t *testing.T) {
		root := t.TempDir()
		writeSettingsDefaultModeFixture(t, root, defs.SettingsJSON, "")

		if got := checkSettingsDefaultMode(root, false); got.Status != uikit.CheckOK {
			t.Errorf("absent key: status = %q, want %q", got.Status, uikit.CheckOK)
		}
	})

	t.Run("absent .claude directory is OK", func(t *testing.T) {
		root := t.TempDir()

		got := checkSettingsDefaultMode(root, false)
		if got.Status != uikit.CheckOK {
			t.Errorf("absent dir: status = %q, want %q", got.Status, uikit.CheckOK)
		}
	})

	t.Run("malformed JSON is skipped, not failed", func(t *testing.T) {
		root := t.TempDir()
		dir := filepath.Join(root, defs.ClaudeDir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, defs.SettingsJSON), []byte("{not json"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}

		if got := checkSettingsDefaultMode(root, false); got.Status != uikit.CheckOK {
			t.Errorf("malformed: status = %q, want %q (malformed settings are another check's surface)", got.Status, uikit.CheckOK)
		}
	})

	t.Run("check is registered as a claude-surface check for codex-only downgrade", func(t *testing.T) {
		if !claudeSurfaceCheckNames[settingsDefaultModeCheckName] {
			t.Errorf("%q must be in claudeSurfaceCheckNames so a codex-only project downgrades it (REQ-IH-011)", settingsDefaultModeCheckName)
		}
	})
}
