package cli

// update_global_settings_noop_test.go — card t1703: moai init/update must not
// rewrite ~/.claude/settings.json when the cleanup changes nothing.
//
// ensureGlobalSettingsEnv deletes the moai-managed PATH key and then re-adds
// the current value, so a file that already carried it was flagged as changed
// on every run and re-marshaled: same meaning, different bytes (key order,
// indent, trailing newline) and a fresh mtime. These tests run against a
// throwaway HOME only; the real ~/.claude is never read or written.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/template"
)

// globalSettingsHome points userHomeDirFn and HOME at a fresh temp directory
// and returns the settings path inside it.
func globalSettingsHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	orig := userHomeDirFn
	userHomeDirFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDirFn = orig })
	t.Setenv("HOME", home)
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, "gitconfig"))
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatalf("create the temp .claude dir: %v", err)
	}
	return filepath.Join(home, ".claude", "settings.json")
}

// TestEnsureGlobalSettingsEnvLeavesAnUpToDateFileAlone writes a settings file
// whose meaning already equals the cleanup's result but whose bytes differ
// from what json.MarshalIndent would emit (four-space indent, unsorted keys,
// no trailing newline), backdates it, and requires the call to leave both the
// bytes and the mtime untouched.
func TestEnsureGlobalSettingsEnvLeavesAnUpToDateFileAlone(t *testing.T) {
	path := globalSettingsHome(t)
	pathJSON, err := json.Marshal(template.BuildSmartPATH())
	if err != nil {
		t.Fatalf("marshal the smart PATH: %v", err)
	}
	seed := []byte("{\n" +
		"    \"zeta\": 1,\n" +
		"    \"env\": {\n" +
		"        \"USER_KEY\": \"x\",\n" +
		"        \"PATH\": " + string(pathJSON) + ",\n" +
		"        \"CLAUDE_DISABLE_PATH_WARNING\": \"1\",\n" +
		"        \"CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS\": \"1\"\n" +
		"    },\n" +
		"    \"alpha\": true\n" +
		"}")
	if err := os.WriteFile(path, seed, 0o644); err != nil {
		t.Fatalf("seed the global settings: %v", err)
	}
	past := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(path, past, past); err != nil {
		t.Fatalf("backdate the global settings: %v", err)
	}

	if err := ensureGlobalSettingsEnv(); err != nil {
		t.Fatalf("ensureGlobalSettingsEnv: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back the global settings: %v", err)
	}
	if !bytes.Equal(got, seed) {
		t.Errorf("an unchanged global settings file was rewritten:\n--- before ---\n%s\n--- after ---\n%s", seed, got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat the global settings: %v", err)
	}
	if !info.ModTime().Equal(past) {
		t.Errorf("mtime moved from %s to %s although nothing changed", past, info.ModTime())
	}
}

// TestEnsureGlobalSettingsEnvStillWritesRealChanges keeps the fix from
// over-suppressing: a file missing the default env keys must still gain them.
func TestEnsureGlobalSettingsEnvStillWritesRealChanges(t *testing.T) {
	path := globalSettingsHome(t)
	if err := os.WriteFile(path, []byte("{\"env\":{\"USER_KEY\":\"x\"}}"), 0o644); err != nil {
		t.Fatalf("seed the global settings: %v", err)
	}

	if err := ensureGlobalSettingsEnv(); err != nil {
		t.Fatalf("ensureGlobalSettingsEnv: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back the global settings: %v", err)
	}
	var got struct {
		Env map[string]string `json:"env"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("parse the rewritten settings: %v", err)
	}
	for _, key := range []string{"PATH", "CLAUDE_DISABLE_PATH_WARNING", "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS", "USER_KEY"} {
		if _, ok := got.Env[key]; !ok {
			t.Errorf("env key %q missing after a cleanup that had real changes to make", key)
		}
	}
}
