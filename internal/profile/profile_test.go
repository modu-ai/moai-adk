package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetBaseDir_Default(t *testing.T) {
	BaseDirOverride = ""
	dir := GetBaseDir()
	if dir == "" || dir == "." {
		// Only fail if HOME is set (CI environments may not have it)
		if os.Getenv("HOME") != "" {
			t.Error("GetBaseDir should return a valid path when HOME is set")
		}
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".moai", "claude-profiles")
	if dir != expected {
		t.Errorf("GetBaseDir() = %q, want %q", dir, expected)
	}
}

func TestGetBaseDir_Override(t *testing.T) {
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()

	BaseDirOverride = "/tmp/test-profiles"
	dir := GetBaseDir()
	if dir != "/tmp/test-profiles" {
		t.Errorf("GetBaseDir() = %q, want /tmp/test-profiles", dir)
	}
}

func TestGetCurrentName_Default(t *testing.T) {
	// Isolate to a clean base dir so GetCurrentName's ledger fallback (via
	// ResolveLaunchProfile) cannot read the real ~/.moai/claude-profiles/
	// launch.yaml, which on the maintainer's machine records a named profile.
	// A clean tmpDir has no launch.yaml → ResolveLaunchProfile returns "" →
	// GetCurrentName returns "default", preserving this test's intent.
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	t.Setenv("CLAUDE_CONFIG_DIR", "")
	name := GetCurrentName()
	if name != "default" {
		t.Errorf("GetCurrentName() = %q, want %q", name, "default")
	}
}

func TestGetCurrentName_WithProfile(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(tmpDir, "work"))
	name := GetCurrentName()
	if name != "work" {
		t.Errorf("GetCurrentName() = %q, want %q", name, "work")
	}
}

func TestGetCurrentName_UnrelatedPath(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", "/some/random/path")
	name := GetCurrentName()
	if name != "/some/random/path" {
		t.Errorf("GetCurrentName() = %q, want raw path", name)
	}
}

// TestGetCurrentName_LedgerFallback verifies that when CLAUDE_CONFIG_DIR is
// unset (the common `moai web` case — the cc/glm/cg launchers only set it when
// spawning Claude Code), GetCurrentName consults the launch.yaml ledger and
// returns the last-used named profile if its directory still exists, rather
// than blindly returning "default". This keeps `moai web`'s displayed profile
// in sync with the profile a bare `moai cc` would actually launch.
func TestGetCurrentName_LedgerFallback(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Named profile directory must exist (stale-record guard).
	if err := os.Mkdir(filepath.Join(tmpDir, "moai-adk"), 0o755); err != nil {
		t.Fatalf("Mkdir(moai-adk): %v", err)
	}
	// Ledger pointing at the named profile.
	if err := os.WriteFile(filepath.Join(tmpDir, "launch.yaml"), []byte("last_profile: moai-adk\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(launch.yaml): %v", err)
	}

	t.Setenv("CLAUDE_CONFIG_DIR", "")

	got := GetCurrentName()
	if got != "moai-adk" {
		t.Errorf("GetCurrentName() = %q, want %q (ledger fallback)", got, "moai-adk")
	}
}

func TestList_DefaultOnly(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	t.Setenv("CLAUDE_CONFIG_DIR", "")

	entries := List()
	if len(entries) != 1 {
		t.Fatalf("List() returned %d entries, want 1", len(entries))
	}
	if entries[0].Name != "default" {
		t.Errorf("entries[0].Name = %q, want %q", entries[0].Name, "default")
	}
	if !entries[0].Current {
		t.Error("default should be current when CLAUDE_CONFIG_DIR is empty")
	}
}

func TestList_WithProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Create profile directories
	if err := os.MkdirAll(filepath.Join(tmpDir, "work"), 0755); err != nil {
		t.Fatalf("MkdirAll(work): %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "personal"), 0755); err != nil {
		t.Fatalf("MkdirAll(personal): %v", err)
	}
	// Create a file (should be ignored)
	if err := os.WriteFile(filepath.Join(tmpDir, "notes.txt"), []byte("ignored"), 0644); err != nil {
		t.Fatalf("WriteFile(notes.txt): %v", err)
	}

	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(tmpDir, "work"))

	entries := List()
	if len(entries) != 3 {
		t.Fatalf("List() returned %d entries, want 3", len(entries))
	}

	// Check that work is marked as current
	found := false
	for _, e := range entries {
		if e.Name == "work" && e.Current {
			found = true
		}
		if e.Name == "default" && e.Current {
			t.Error("default should not be current when work is active")
		}
	}
	if !found {
		t.Error("work profile should be marked as current")
	}
}

func TestEnsureDir_Default(t *testing.T) {
	err := EnsureDir("default")
	if err != nil {
		t.Errorf("EnsureDir(default) should be no-op, got: %v", err)
	}
}

func TestEnsureDir_Empty(t *testing.T) {
	err := EnsureDir("")
	if err != nil {
		t.Errorf("EnsureDir('') should be no-op, got: %v", err)
	}
}

func TestEnsureDir_CreatesAndSetsEnv(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	t.Setenv("CLAUDE_CONFIG_DIR", "")

	err := EnsureDir("myprofile")
	if err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}

	// Check directory was created
	profileDir := filepath.Join(tmpDir, "myprofile")
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		t.Error("profile directory should be created")
	}

	// Check env var was set
	if os.Getenv("CLAUDE_CONFIG_DIR") != profileDir {
		t.Errorf("CLAUDE_CONFIG_DIR = %q, want %q", os.Getenv("CLAUDE_CONFIG_DIR"), profileDir)
	}
}

func TestDelete_DefaultProfile(t *testing.T) {
	err := Delete("default")
	if err == nil {
		t.Error("Delete(default) should return error")
	}
}

func TestDelete_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	err := Delete("nonexistent")
	if err == nil {
		t.Error("Delete(nonexistent) should return error")
	}
}

func TestDelete_Success(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	t.Setenv("CLAUDE_CONFIG_DIR", "")

	// Create the profile
	profileDir := filepath.Join(tmpDir, "testprofile")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		t.Fatalf("MkdirAll(testprofile): %v", err)
	}

	err := Delete("testprofile")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, err := os.Stat(profileDir); !os.IsNotExist(err) {
		t.Error("profile directory should be deleted")
	}
}

// TestResolveLaunchProfile_EmptyInputReturnsLastUsed verifies that when no -p
// flag is given (profileName==""), ResolveLaunchProfile falls back to the
// last_profile recorded in launch.yaml, and that the resolved profile yields
// real preferences.
func TestResolveLaunchProfile_EmptyInputReturnsLastUsed(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Empty base preferences (the default profile is unconfigured).
	if err := os.WriteFile(filepath.Join(tmpDir, "preferences.yaml"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Named profile with real preferences.
	namedDir := filepath.Join(tmpDir, "moai-adk")
	if err := os.MkdirAll(namedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(namedDir, "preferences.yaml"), []byte("model: opus[1m]\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// launch.yaml ledger pointing to the named profile.
	if err := os.WriteFile(filepath.Join(tmpDir, "launch.yaml"), []byte("last_profile: moai-adk\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CLAUDE_CONFIG_DIR", "")

	resolved := ResolveLaunchProfile("")
	if resolved != "moai-adk" {
		t.Fatalf("ResolveLaunchProfile('') = %q, want %q", resolved, "moai-adk")
	}

	// The resolved profile must yield real preferences.
	prefs, err := ReadPreferences(resolved)
	if err != nil {
		t.Fatalf("ReadPreferences(%q) error: %v", resolved, err)
	}
	if prefs.Model != "opus[1m]" {
		t.Errorf("ReadPreferences(%q).Model = %q, want %q", resolved, prefs.Model, "opus[1m]")
	}
}

// TestResolveLaunchProfile_ExplicitFlagWins verifies that an explicit profileName
// (from -p) is returned as-is, never overridden by the ledger.
func TestResolveLaunchProfile_ExplicitFlagWins(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	if err := os.MkdirAll(filepath.Join(tmpDir, "moai-adk"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "launch.yaml"), []byte("last_profile: moai-adk\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := ResolveLaunchProfile("work")
	if got != "work" {
		t.Errorf("ResolveLaunchProfile('work') = %q, want %q", got, "work")
	}
}

// TestResolveLaunchProfile_StaleRecordIgnored verifies that a last_profile entry
// whose directory does not exist is ignored (returns "" = default semantics).
func TestResolveLaunchProfile_StaleRecordIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// launch.yaml points to a profile whose directory does NOT exist.
	if err := os.WriteFile(filepath.Join(tmpDir, "launch.yaml"), []byte("last_profile: ghost\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := ResolveLaunchProfile("")
	if got != "" {
		t.Errorf("ResolveLaunchProfile('') = %q, want '' (stale record ignored)", got)
	}
}

// TestResolveLaunchProfile_OptOutEnv verifies that MOAI_NO_PROFILE_FALLBACK=1
// disables the fallback even when a valid last_profile ledger exists.
func TestResolveLaunchProfile_OptOutEnv(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Valid ledger + named profile directory.
	if err := os.MkdirAll(filepath.Join(tmpDir, "moai-adk"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "launch.yaml"), []byte("last_profile: moai-adk\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MOAI_NO_PROFILE_FALLBACK", "1")

	got := ResolveLaunchProfile("")
	if got != "" {
		t.Errorf("ResolveLaunchProfile('') with opt-out = %q, want ''", got)
	}
}

// TestRecordLastUsedProfile_RejectsDefaultAndEmpty verifies that only NAMED
// profiles are recorded — "" and "default" are refused.
func TestRecordLastUsedProfile_RejectsDefaultAndEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	if err := RecordLastUsedProfile("default"); err == nil {
		t.Error("RecordLastUsedProfile('default') should return error, got nil")
	}
	if err := RecordLastUsedProfile(""); err == nil {
		t.Error("RecordLastUsedProfile('') should return error, got nil")
	}
}

// TestRecordLastUsedProfile_PreservesLegacyKeys verifies the read-modify-write
// contract: pre-existing launch.yaml keys (e.g. legacy model:/bypass:) are
// preserved alongside the new last_profile key.
func TestRecordLastUsedProfile_PreservesLegacyKeys(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Pre-existing launch.yaml with legacy keys.
	legacy := "model: claude-opus-4-6\nbypass: true\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "launch.yaml"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RecordLastUsedProfile("work"); err != nil {
		t.Fatalf("RecordLastUsedProfile('work') error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "launch.yaml"))
	if err != nil {
		t.Fatalf("read launch.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "last_profile: work") {
		t.Errorf("launch.yaml missing 'last_profile: work'\n--- content ---\n%s", content)
	}
	if !strings.Contains(content, "model: claude-opus-4-6") {
		t.Errorf("launch.yaml lost legacy 'model' key\n--- content ---\n%s", content)
	}
	if !strings.Contains(content, "bypass: true") {
		t.Errorf("launch.yaml lost legacy 'bypass' key\n--- content ---\n%s", content)
	}
}
