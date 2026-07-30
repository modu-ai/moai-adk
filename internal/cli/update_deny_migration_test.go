package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

// runUpdateInFixture chdirs into root, runs the real `moai update` command with
// a mocked update checker, and returns the combined output. Mirrors the
// TestRunUpdate_ThreeRunIdempotency_V3Project harness: updateCmd is a
// package-level cobra command, so callers must not run in parallel.
func runUpdateInFixture(t *testing.T, root string) string {
	t.Helper()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir fixture: %v", err)
	}

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = &Dependencies{UpdateChecker: &mockUpdateChecker{}}

	var buf bytes.Buffer
	updateCmd.SetOut(&buf)
	updateCmd.SetErr(&buf)
	updateCmd.SetContext(context.Background())
	if err := updateCmd.Flags().Set("check", "false"); err != nil {
		t.Fatalf("set check flag: %v", err)
	}
	if err := updateCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatalf("set yes flag: %v", err)
	}

	if runErr := updateCmd.RunE(updateCmd, []string{}); runErr != nil {
		// Template sync on a minimal fixture may warn; the filesystem
		// assertions in the caller are the load-bearing invariants.
		t.Logf("runUpdate returned (non-fatal): %v", runErr)
	}
	return buf.String()
}

// writeV3ProjectFixture lays down the minimum markers that make detectV2Fingerprint
// return IsV2:false — a v3 system.yaml, no .agency/, no deprecated paths — so
// `moai update` routes to the plain v3 file-level sync rather than to
// runCleanReinstall. This is the exact shape that previously never received the
// deny-rule migration.
func writeV3ProjectFixture(t *testing.T, root string) {
	t.Helper()
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n    version: v3.0.1\n")
	writeTestFile(t, root, ".moai/config/sections/language.yaml",
		"language:\n    conversation_language: ko\n")
}

// TestRunUpdate_V3Path_StripsRetiredDenyEntries pins the reachability of
// stripRetiredV2DenyEntries from the plain v3 update path (issue #1101
// follow-up). Before the fix the call site existed ONLY inside
// runCleanReinstall, which opens on a v2 fingerprint — so a project already on
// v3 kept the retired Write/Grep/Glob entries through every `moai update` and
// Claude Code warned about them on every session start.
//
// Falsifiability: disabling the runUpdate call site makes this test fail
// (verified by running it with the call stubbed out).
//
// Coverage note — what this fixture does and does NOT exercise: under the test
// binary's version the fixture takes the full file-level sync branch, so the
// version-match short-circuit (`syncSkipped` early return) is NOT covered here.
// That branch is why the call site sits BEFORE the short-circuit rather than in
// the post-sync section; the placement is asserted by the call site's own
// comment in update.go, not by this test.
func TestRunUpdate_V3Path_StripsRetiredDenyEntries(t *testing.T) {
	root := t.TempDir()
	writeV3ProjectFixture(t, root)
	settingsPath := writeSettingsFixture(t, root, v2SettingsFixture)

	out := runUpdateInFixture(t, root)

	deny, m := readDenyList(t, settingsPath)

	for _, retired := range retiredV2DenyEntries {
		if slices.Contains(deny, retired) {
			t.Errorf("retired deny entry %q survived the v3 update path; deny=%v\n--- output ---\n%s",
				retired, deny, out)
		}
	}

	// The surviving rules must be untouched: the 8 template Read/Edit entries
	// plus the user's own custom entry (exact-match strip only).
	mustKeep := []string{
		"Read(./secrets/**)", "Read(~/.ssh/**)", "Read(~/.aws/**)", "Read(~/.config/gcloud/**)",
		"Edit(./secrets/**)", "Edit(~/.ssh/**)", "Edit(~/.aws/**)", "Edit(~/.config/gcloud/**)",
		"Write(./my-custom-protected/**)",
	}
	for _, want := range mustKeep {
		if !slices.Contains(deny, want) {
			t.Errorf("deny entry %q was removed but must be preserved; deny=%v", want, deny)
		}
	}

	// Unknown top-level keys round-trip (SPEC-CLIFIX-CRITICAL-001 precedent).
	if m["outputStyle"] != "MoAI-Easy" {
		t.Errorf("outputStyle not preserved through the v3 update path: %v", m["outputStyle"])
	}

	// The migration must not claim a clean reinstall happened on the v3 path,
	// and must not collide with the AC-CRR-009(c) "[clean-reinstall] Removed"
	// assertion in update_clean_install_test.go.
	if !strings.Contains(out, "[settings] Removed") {
		t.Errorf("expected the neutral-prefix migration log on the v3 path; output:\n%s", out)
	}
	if strings.Contains(out, "[clean-reinstall] Removed") {
		t.Errorf("deny migration emitted a clean-reinstall label on the v3 path; output:\n%s", out)
	}
}

// TestRunUpdate_V3Path_CleanSettingsUntouched pins the no-op half of the
// contract: a v3 project whose settings.json carries no retired entries must
// come out byte-identical, so adding the call site to the always-run path
// cannot churn an already-clean file.
func TestRunUpdate_V3Path_CleanSettingsUntouched(t *testing.T) {
	const v3Clean = `{
  "outputStyle": "MoAI-Easy",
  "permissions": {
    "deny": [
      "Read(./secrets/**)",
      "Edit(./secrets/**)"
    ]
  }
}`
	root := t.TempDir()
	writeV3ProjectFixture(t, root)
	settingsPath := writeSettingsFixture(t, root, v3Clean)

	out := runUpdateInFixture(t, root)

	got, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != v3Clean {
		t.Errorf("clean v3 settings.json was rewritten by the update path:\n--- got ---\n%s\n--- want ---\n%s",
			got, v3Clean)
	}
	if strings.Contains(out, "[settings] Removed") {
		t.Errorf("migration logged a removal on a clean file; output:\n%s", out)
	}
}
