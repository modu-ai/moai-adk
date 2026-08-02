package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetBaseDir_Default(t *testing.T) {
	// Checking the home-derived default requires clearing the override, which
	// also disables the package sandbox for every test that runs afterwards.
	// Restoring it is mandatory; TestSandboxSurvivesPackageRun
	// (zz_sandbox_guard_test.go, which sorts last) fails if this is dropped.
	orig := BaseDirOverride
	t.Cleanup(func() { BaseDirOverride = orig })

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

// TestGetCurrentName_GlobalLedgerDoesNotBleed verifies that a global
// last_profile entry in launch.yaml does NOT leak into the project-less
// GetCurrentName() wrapper (which forwards an empty projectRoot).
//
// The global last_profile key is write-only on this binary: resolution is
// project-scoped only, so a caller with no project root (the common `moai web`
// case where CLAUDE_CONFIG_DIR is unset) resolves to "default" regardless of
// what the global ledger says. This is the console-side half of the
// cross-project-bleed regression.
func TestGetCurrentName_GlobalLedgerDoesNotBleed(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Named profile directory exists so the entry is not stale — the point is
	// that even a USABLE global entry is not read.
	if err := os.Mkdir(filepath.Join(tmpDir, "moai-adk"), 0o755); err != nil {
		t.Fatalf("Mkdir(moai-adk): %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "launch.yaml"), []byte("last_profile: moai-adk\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(launch.yaml): %v", err)
	}

	t.Setenv("CLAUDE_CONFIG_DIR", "")

	got := GetCurrentName()
	if got != "default" {
		t.Errorf("GetCurrentName() = %q, want %q (global last_profile must not bleed into project-less resolution)", got, "default")
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

// TestResolveLaunchProfile_EmptyInputIgnoresGlobalLedger verifies that when
// no -p flag is given (profileName==""), ResolveLaunchProfile ("") does NOT
// fall back to the global last_profile recorded in launch.yaml. The global
// read was removed because it let one project's profile bleed into another;
// with an empty projectRoot the wrapper has no project-scoped entry to consult
// either, so it returns "" (default semantics) regardless of the global key.
func TestResolveLaunchProfile_EmptyInputIgnoresGlobalLedger(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Named profile directory exists so the global entry is not stale — the
	// point is that even a USABLE global entry is no longer read.
	namedDir := filepath.Join(tmpDir, "moai-adk")
	if err := os.MkdirAll(namedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(namedDir, "preferences.yaml"), []byte("model: opus[1m]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "launch.yaml"), []byte("last_profile: moai-adk\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CLAUDE_CONFIG_DIR", "")

	resolved := ResolveLaunchProfile("")
	if resolved != "" {
		t.Fatalf("ResolveLaunchProfile('') = %q, want \"\" (global last_profile must not be read for resolution)", resolved)
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

// TestResolveLaunchProfile_StaleRecordIgnored verifies that a last_profile
// entry whose directory does not exist yields "" (default semantics). Since
// the global last_profile key is no longer read for resolution at all, the
// stale-skip guard is moot for this key — this test now pins the stronger
// invariant that the global key simply does not participate in resolution,
// stale or not.
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

// TestResolveLaunchProfile_NoGlobalBleedAcrossProjects pins the user's exact
// bug: a global last_profile recorded for one project must NOT bleed into a
// different project that has no projects[] entry of its own. Resolution is
// project-scoped only after the fix — the global last_profile key is write-only
// on this binary.
//
// Reproduces (and pins the fix for) the on-disk state observed in
// ~/.moai/claude-profiles/launch.yaml:
//
//	last_profile: moai-cowork
//	projects:
//	    /Users/goos/MoAI/moai-cowork: moai-cowork   # no entry for moai-adk-go
//
// where bare `moai glm` in moai-adk-go used to resolve to moai-cowork.
func TestResolveLaunchProfile_NoGlobalBleedAcrossProjects(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	t.Cleanup(func() { BaseDirOverride = orig })
	BaseDirOverride = tmpDir

	// Two real projects, two named profiles. Both profile directories exist so
	// launchCandidateIsUsable never skips them — the test must fail because the
	// global read is gone, not because of a stale guard.
	projA := t.TempDir()
	projB := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "prof-x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "prof-y"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "prof-z"), 0o755); err != nil {
		t.Fatal(err)
	}

	// projects[A] = "prof-x", last_profile = "prof-y" (would have bled to B
	// under the old global fallback).
	ledger := "last_profile: prof-y\nprojects:\n  " + normalizeProjectKey(projA) + ": prof-x\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "launch.yaml"), []byte(ledger), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CLAUDE_CONFIG_DIR", "")

	// (a) The bug: project B is NOT in the projects map → MUST resolve to ""
	//     (default), NOT "prof-y". Under the old code this returned "prof-y".
	if got := ResolveLaunchProfileForProject(projB, ""); got != "" {
		t.Errorf("(a) ResolveLaunchProfileForProject(projB, \"\") = %q, want \"\" (global last_profile must not bleed into a project with no projects[] entry)", got)
	}

	// (b) Project memory still works: project A IS in the map → "prof-x".
	if got := ResolveLaunchProfileForProject(projA, ""); got != "prof-x" {
		t.Errorf("(b) ResolveLaunchProfileForProject(projA, \"\") = %q, want %q (project-scoped memory)", got, "prof-x")
	}

	// (c) Explicit -p always wins even when projects[A] and last_profile disagree.
	if got := ResolveLaunchProfileForProject(projA, "prof-z"); got != "prof-z" {
		t.Errorf("(c) ResolveLaunchProfileForProject(projA, \"prof-z\") = %q, want %q (explicit -p wins)", got, "prof-z")
	}
	if got := ResolveLaunchProfileForProject(projB, "prof-z"); got != "prof-z" {
		t.Errorf("(c') ResolveLaunchProfileForProject(projB, \"prof-z\") = %q, want %q (explicit -p wins)", got, "prof-z")
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

	// The recorder refuses names whose directory does not exist
	// (SPEC-PROFILE-MEMORY-001 REQ-PM-011), so the profile must be staged
	// before recording. TestRecordForProject_RejectsMissingDirectory owns the
	// refusal case; this test stays focused on legacy-key preservation.
	if err := os.MkdirAll(filepath.Join(tmpDir, "work"), 0o755); err != nil {
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
