package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
	"gopkg.in/yaml.v3"
)

// --- helpers (SPEC-PROFILE-MEMORY-001 REQ-PM-021) ---

// lpSandbox points the profile base at a throwaway directory, fixes the project
// root, and stubs the launch seam. It returns the profile base and the project
// root. Every seam is restored via t.Cleanup.
func lpSandbox(t *testing.T) (base string, root string) {
	t.Helper()

	base = t.TempDir()
	origBase := profile.BaseDirOverride
	profile.BaseDirOverride = base
	t.Cleanup(func() { profile.BaseDirOverride = origBase })

	root = t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("stage project root: %v", err)
	}
	origRoot := findProjectRootFn
	findProjectRootFn = func() (string, error) { return root, nil }
	t.Cleanup(func() { findProjectRootFn = origRoot })

	// EnsureDir mutates CLAUDE_CONFIG_DIR; t.Setenv restores it after the test.
	t.Setenv("CLAUDE_CONFIG_DIR", "")

	return base, root
}

// lpStubLaunch replaces launchClaudeFunc with a recorder and returns a pointer
// to the profile name it was handed.
func lpStubLaunch(t *testing.T) (calls *int, gotProfile *string) {
	t.Helper()
	orig := launchClaudeFunc
	t.Cleanup(func() { launchClaudeFunc = orig })

	calls = new(int)
	gotProfile = new(string)
	launchClaudeFunc = func(p string, _ []string) error {
		*calls++
		*gotProfile = p
		return nil
	}
	return calls, gotProfile
}

func lpProjectKey(t *testing.T, root string) string {
	t.Helper()
	if resolved, err := filepath.EvalSymlinks(filepath.Clean(root)); err == nil {
		return resolved
	}
	return filepath.Clean(root)
}

func lpMkProfile(t *testing.T, base, name string) string {
	t.Helper()
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create profile dir %q: %v", name, err)
	}
	return dir
}

func lpWriteLedger(t *testing.T, base, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(base, "launch.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write ledger: %v", err)
	}
}

func lpReadLedger(t *testing.T, base string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(base, "launch.yaml"))
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	m := map[string]any{}
	if err := yaml.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal ledger: %v", err)
	}
	return m
}

func lpLedgerProjectEntry(t *testing.T, base, root string) string {
	t.Helper()
	ledger := lpReadLedger(t, base)
	projects, ok := ledger["projects"].(map[string]any)
	if !ok {
		return ""
	}
	s, _ := projects[lpProjectKey(t, root)].(string)
	return s
}

// --- AC-PM-017 (REQ-PM-003, launch layer) ---

// TestUnifiedLaunch_UsesProjectScopedResolution is the judge of the iter-1
// audit's critical defect: wiring the WRITE side to the projects map while
// leaving the launcher reading the global-only resolver produces a key that is
// written and never read. A function-level test cannot see that, because it
// calls the resolver directly; only a launch-level test can.
func TestUnifiedLaunch_UsesProjectScopedResolution(t *testing.T) {
	base, root := lpSandbox(t)
	lpMkProfile(t, base, "proj-one")
	lpMkProfile(t, base, "global-one")
	lpWriteLedger(t, base, "last_profile: global-one\nprojects:\n  "+lpProjectKey(t, root)+": proj-one\n")

	_, gotProfile := lpStubLaunch(t)

	if err := unifiedLaunch("", "claude", nil); err != nil {
		t.Fatalf("unifiedLaunch: %v", err)
	}

	if *gotProfile != "proj-one" {
		t.Errorf("launchClaude received profile %q, want proj-one; %q means the launcher "+
			"is still resolving through the global-only path", *gotProfile, "global-one")
	}
}

// --- AC-PM-009 (REQ-PM-012, 013, 015) ---

// TestUnifiedLaunch_FirstTimeNewProfileIsRecorded is the judge of the ordering
// invariant: the directory must be created BEFORE the ledger write, otherwise
// the REQ-PM-011 existence guard rejects every first-time `-p <new>` launch and
// the profile is silently never remembered.
func TestUnifiedLaunch_FirstTimeNewProfileIsRecorded(t *testing.T) {
	base, root := lpSandbox(t)

	profileDir := filepath.Join(base, "brand-new")
	if _, err := os.Stat(profileDir); !os.IsNotExist(err) {
		t.Fatalf("precondition: %q must not exist, stat err = %v", profileDir, err)
	}

	// Snapshot the state as observed AT the moment launchClaude is called, so
	// the assertions cannot be satisfied by anything happening afterwards.
	var dirExistedAtLaunch bool
	var lastAtLaunch, projectAtLaunch string
	origLaunch := launchClaudeFunc
	t.Cleanup(func() { launchClaudeFunc = origLaunch })
	launched := 0
	launchClaudeFunc = func(_ string, _ []string) error {
		launched++
		info, err := os.Stat(profileDir)
		dirExistedAtLaunch = err == nil && info.IsDir()
		if data, err := os.ReadFile(filepath.Join(base, "launch.yaml")); err == nil {
			m := map[string]any{}
			if yaml.Unmarshal(data, &m) == nil {
				lastAtLaunch, _ = m["last_profile"].(string)
				if projects, ok := m["projects"].(map[string]any); ok {
					projectAtLaunch, _ = projects[lpProjectKey(t, root)].(string)
				}
			}
		}
		return nil
	}

	if err := unifiedLaunch("brand-new", "claude", nil); err != nil {
		t.Fatalf("unifiedLaunch: %v", err)
	}
	if launched != 1 {
		t.Fatalf("launchClaude called %d times, want 1", launched)
	}

	if !dirExistedAtLaunch {
		t.Error("(a) profile directory did not exist when launchClaude ran")
	}
	if lastAtLaunch != "brand-new" {
		t.Errorf("(b) last_profile at launch = %q, want brand-new", lastAtLaunch)
	}
	if projectAtLaunch != "brand-new" {
		t.Errorf("(b) project-scoped entry at launch = %q, want brand-new", projectAtLaunch)
	}

	// And the same holds after the launch returns.
	if got := lpReadLedger(t, base)["last_profile"]; got != "brand-new" {
		t.Errorf("last_profile = %v, want brand-new", got)
	}
	if got := lpLedgerProjectEntry(t, base, root); got != "brand-new" {
		t.Errorf("projects[root] = %q, want brand-new", got)
	}
}

// --- AC-PM-010 (REQ-PM-014) ---

// TestUnifiedLaunch_RecordFailureDoesNotBlockLaunch injects the failure through
// the recordLastProfileFn seam rather than file permissions: a chmod-based
// recipe is void under root and unjudgeable on Windows CI (constraint C5).
func TestUnifiedLaunch_RecordFailureDoesNotBlockLaunch(t *testing.T) {
	base, _ := lpSandbox(t)
	lpMkProfile(t, base, "work")

	origRecord := recordLastProfileFn
	t.Cleanup(func() { recordLastProfileFn = origRecord })
	recordLastProfileFn = func(_, _ string) error {
		return errors.New("injected ledger failure")
	}

	var stderr bytes.Buffer
	origStderr := launcherStderr
	t.Cleanup(func() { launcherStderr = origStderr })
	launcherStderr = &stderr

	calls, _ := lpStubLaunch(t)

	if err := unifiedLaunch("work", "claude", nil); err != nil {
		t.Errorf("(a) unifiedLaunch returned %v, want nil — a ledger write failure must not block the launch", err)
	}
	if *calls != 1 {
		t.Errorf("(b) launchClaude called %d times, want 1", *calls)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("failed to record")) {
		t.Errorf("(c) injected stderr does not carry the warning; got:\n%s", stderr.String())
	}
}
