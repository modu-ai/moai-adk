package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// v2-era settings.json fixture: the 12 retired deny entries (Write/Grep/Glob x
// 4 protected paths) that the v3 template no longer ships, plus the surviving
// Read/Edit entries, a user-custom deny entry, and unrelated keys that must
// round-trip untouched (SPEC-CLIFIX-CRITICAL-001 map[string]any precedent).
const v2SettingsFixture = `{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "outputStyle": "MoAI-Easy",
  "permissions": {
    "allow": [
      "Bash(go build:*)"
    ],
    "deny": [
      "Read(./secrets/**)",
      "Read(~/.ssh/**)",
      "Read(~/.aws/**)",
      "Read(~/.config/gcloud/**)",
      "Edit(./secrets/**)",
      "Edit(~/.ssh/**)",
      "Edit(~/.aws/**)",
      "Edit(~/.config/gcloud/**)",
      "Write(./secrets/**)",
      "Write(~/.ssh/**)",
      "Write(~/.aws/**)",
      "Write(~/.config/gcloud/**)",
      "Grep(./secrets/**)",
      "Grep(~/.ssh/**)",
      "Grep(~/.aws/**)",
      "Grep(~/.config/gcloud/**)",
      "Glob(./secrets/**)",
      "Glob(~/.ssh/**)",
      "Glob(~/.aws/**)",
      "Glob(~/.config/gcloud/**)",
      "Write(./my-custom-protected/**)"
    ]
  },
  "env": {
    "MOAI_CONFIG_SOURCE": "template"
  }
}`

func writeSettingsFixture(t *testing.T, root, content string) string {
	t.Helper()
	dir := filepath.Join(root, ".claude")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func readDenyList(t *testing.T, path string) ([]string, map[string]any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	m := make(map[string]any)
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	perms, _ := m["permissions"].(map[string]any)
	rawDeny, _ := perms["deny"].([]any)
	deny := make([]string, 0, len(rawDeny))
	for _, e := range rawDeny {
		if s, ok := e.(string); ok {
			deny = append(deny, s)
		}
	}
	return deny, m
}

func TestStripRetiredV2DenyEntries_RemovesRetiredKeepsCustom(t *testing.T) {
	root := t.TempDir()
	path := writeSettingsFixture(t, root, v2SettingsFixture)

	var out bytes.Buffer
	if err := stripRetiredV2DenyEntries(root, &out); err != nil {
		t.Fatalf("stripRetiredV2DenyEntries: %v", err)
	}

	deny, m := readDenyList(t, path)

	// All 12 retired entries removed.
	for _, retired := range retiredV2DenyEntries {
		for _, e := range deny {
			if e == retired {
				t.Errorf("retired entry %q still present after migration", retired)
			}
		}
	}

	// Surviving v3 entries + user-custom entry preserved, in order.
	want := []string{
		"Read(./secrets/**)",
		"Read(~/.ssh/**)",
		"Read(~/.aws/**)",
		"Read(~/.config/gcloud/**)",
		"Edit(./secrets/**)",
		"Edit(~/.ssh/**)",
		"Edit(~/.aws/**)",
		"Edit(~/.config/gcloud/**)",
		"Write(./my-custom-protected/**)",
	}
	if len(deny) != len(want) {
		t.Fatalf("deny list length = %d, want %d; got %v", len(deny), len(want), deny)
	}
	for i, w := range want {
		if deny[i] != w {
			t.Errorf("deny[%d] = %q, want %q", i, deny[i], w)
		}
	}

	// Unknown top-level keys and sibling permission keys survive byte-semantically.
	if m["outputStyle"] != "MoAI-Easy" {
		t.Errorf("outputStyle not preserved: %v", m["outputStyle"])
	}
	if m["$schema"] == nil {
		t.Error("$schema key wiped")
	}
	env, _ := m["env"].(map[string]any)
	if env["MOAI_CONFIG_SOURCE"] != "template" {
		t.Errorf("env.MOAI_CONFIG_SOURCE not preserved: %v", env)
	}
	perms, _ := m["permissions"].(map[string]any)
	allow, _ := perms["allow"].([]any)
	if len(allow) != 1 || allow[0] != "Bash(go build:*)" {
		t.Errorf("permissions.allow not preserved: %v", allow)
	}
}

func TestStripRetiredV2DenyEntries_Idempotent(t *testing.T) {
	root := t.TempDir()
	path := writeSettingsFixture(t, root, v2SettingsFixture)

	var out bytes.Buffer
	if err := stripRetiredV2DenyEntries(root, &out); err != nil {
		t.Fatalf("first run: %v", err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	info1, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := stripRetiredV2DenyEntries(root, &out); err != nil {
		t.Fatalf("second run: %v", err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Error("second run changed file content; migration is not idempotent")
	}
	info2, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info2.ModTime().Equal(info1.ModTime()) {
		t.Error("second run rewrote the file (mtime changed); expected no-op")
	}
}

func TestStripRetiredV2DenyEntries_V3CleanUntouched(t *testing.T) {
	// A v3-clean settings.json (no retired entries) must not be rewritten.
	const v3Clean = `{
  "permissions": {
    "deny": [
      "Read(./secrets/**)",
      "Edit(./secrets/**)"
    ]
  }
}`
	root := t.TempDir()
	path := writeSettingsFixture(t, root, v3Clean)

	var out bytes.Buffer
	if err := stripRetiredV2DenyEntries(root, &out); err != nil {
		t.Fatalf("stripRetiredV2DenyEntries: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != v3Clean {
		t.Errorf("v3-clean settings.json was rewritten:\n%s", got)
	}
}

func TestStripRetiredV2DenyEntries_MissingFileNoop(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	if err := stripRetiredV2DenyEntries(root, &out); err != nil {
		t.Fatalf("missing settings.json should be a no-op, got error: %v", err)
	}
}

func TestStripRetiredV2DenyEntries_NoPermissionsKeyNoop(t *testing.T) {
	const noPerms = `{
  "outputStyle": "MoAI-Easy"
}`
	root := t.TempDir()
	path := writeSettingsFixture(t, root, noPerms)
	var out bytes.Buffer
	if err := stripRetiredV2DenyEntries(root, &out); err != nil {
		t.Fatalf("no-permissions settings.json should be a no-op, got error: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != noPerms {
		t.Errorf("no-permissions settings.json was rewritten:\n%s", got)
	}
}
