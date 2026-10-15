package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

const hookShellSettingsWithBash = `{
  "hooks": {
    "PreToolUse": [
      {"matcher": "Write|Edit|Bash", "hooks": [
        {"type": "command", "command": "bash", "args": ["-c", "run pre-tool"]}
      ]}
    ],
    "PermissionRequest": [
      {"hooks": [{"type": "command", "command": "bash", "args": ["-c", "run perm"]}]}
    ],
    "Stop": [
      {"hooks": [{"type": "command", "command": "moai", "args": ["hook", "stop"]}]}
    ]
  }
}`

const hookShellSettingsNoBash = `{
  "hooks": {
    "Stop": [
      {"hooks": [{"type": "command", "command": "moai", "args": ["hook", "stop"]}]}
    ]
  }
}`

func writeHookShellSettings(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), []byte(body), 0o644); err != nil {
		t.Fatalf("write settings: %v", err)
	}
	return dir
}

func lookPathMissing(string) (string, error) { return "", errors.New("not found") }
func lookPathFound(string) (string, error)   { return `C:\Git\bin\bash.exe`, nil }

func TestHookShellWindows_BashMissing_Warns(t *testing.T) {
	dir := writeHookShellSettings(t, hookShellSettingsWithBash)
	c := checkHookShellWindows(dir, "windows", lookPathMissing, true)
	if c.Name != hookShellWindowsCheckName {
		t.Errorf("Name = %q, want %q", c.Name, hookShellWindowsCheckName)
	}
	if c.Status != uikit.CheckWarn {
		t.Fatalf("Status = %v, want CheckWarn (msg=%q)", c.Status, c.Message)
	}
	for _, want := range []string{"2 ", "PreToolUse", "PermissionRequest", "2.1.299", "Git for Windows"} {
		if !strings.Contains(c.Message, want) {
			t.Errorf("Message %q lacks %q", c.Message, want)
		}
	}
	if !strings.Contains(c.Detail, "not measured") {
		t.Errorf("Detail %q must say the effect was not measured on Windows", c.Detail)
	}
}

func TestHookShellWindows_BashPresent_OK(t *testing.T) {
	dir := writeHookShellSettings(t, hookShellSettingsWithBash)
	c := checkHookShellWindows(dir, "windows", lookPathFound, false)
	if c.Status != uikit.CheckOK {
		t.Fatalf("Status = %v, want CheckOK (msg=%q)", c.Status, c.Message)
	}
}

func TestHookShellWindows_NoBashHooks_OK(t *testing.T) {
	dir := writeHookShellSettings(t, hookShellSettingsNoBash)
	c := checkHookShellWindows(dir, "windows", lookPathMissing, false)
	if c.Status != uikit.CheckOK {
		t.Fatalf("Status = %v, want CheckOK (msg=%q)", c.Status, c.Message)
	}
}

func TestHookShellWindows_NonWindows_OK(t *testing.T) {
	// No settings file at all: a non-windows host must not even look.
	c := checkHookShellWindows(t.TempDir(), "darwin", lookPathMissing, false)
	if c.Status != uikit.CheckOK {
		t.Fatalf("Status = %v, want CheckOK (msg=%q)", c.Status, c.Message)
	}
	if !strings.Contains(c.Message, "not applicable") {
		t.Errorf("Message = %q, want 'not applicable'", c.Message)
	}
}

func TestHookShellWindows_MissingSettings_Warns(t *testing.T) {
	c := checkHookShellWindows(t.TempDir(), "windows", lookPathMissing, false)
	if c.Status != uikit.CheckWarn {
		t.Fatalf("Status = %v, want CheckWarn (msg=%q)", c.Status, c.Message)
	}
	if !strings.HasPrefix(c.Message, "not checked:") {
		t.Errorf("Message = %q, want 'not checked:' prefix", c.Message)
	}
}

func TestHookShellWindows_RegisteredAsClaudeSurface(t *testing.T) {
	if !claudeSurfaceCheckNames[hookShellWindowsCheckName] {
		t.Errorf("%q must be in claudeSurfaceCheckNames", hookShellWindowsCheckName)
	}
}
