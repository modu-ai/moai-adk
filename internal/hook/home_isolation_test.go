package hook

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// TestResolveCaptureMemoryDir_DoesNotEscapeToRealHome guards the second of the
// two memory-dir derivations that join the real home with a project-derived
// slug.
//
// The shape is the one that leaked 514,385 permanent directories into a
// developer's ~/.claude/projects from the sibling internal/cli/preference
// package: a test isolates the project half (CWD / CLAUDE_PROJECT_DIR) into
// t.TempDir() but leaves the home half resolving to the real home, and because
// every t.TempDir() name carries a fresh random suffix, each run deposits one
// more directory that nothing ever removes. Nothing in a green test run reports
// it; the only symptom is a home directory that grows until reading it stalls.
//
// resolveCaptureMemoryDir reads os.UserHomeDir() inside the function rather than
// receiving the home as a parameter, which is exactly why it needs a guard: the
// caller cannot inject an isolated home, so the isolation has to come from the
// environment and a test that forgets it leaks silently.
//
// Falsifiability: deleting the t.Setenv("HOME", tmp) line below makes this test
// fail, because the derivation then points inside the real home.
func TestResolveCaptureMemoryDir_DoesNotEscapeToRealHome(t *testing.T) {
	// Cannot run in parallel: t.Setenv mutates process-wide state.
	realHome, err := os.UserHomeDir() // captured BEFORE HOME is overridden
	if err != nil {
		t.Fatalf("os.UserHomeDir(): %v", err)
	}

	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)
	t.Setenv("HOME", tmp)
	// os.UserHomeDir reads USERPROFILE on Windows and HOME elsewhere, so both
	// have to be redirected for the isolation to hold on every platform.
	// Setting only HOME leaves the derivation pointing at the real home on
	// Windows — precisely the leak this test exists to catch.
	t.Setenv("USERPROFILE", tmp)

	got, err := resolveCaptureMemoryDir(&HookInput{CWD: tmp})
	if err != nil {
		t.Fatalf("resolveCaptureMemoryDir: %v", err)
	}

	realProjects := filepath.Join(realHome, ".claude", "projects")
	rel, err := filepath.Rel(realProjects, got)
	if err == nil && rel != ".." && !filepath.IsAbs(rel) &&
		len(rel) > 0 && rel[0] != '.' {
		t.Errorf("resolveCaptureMemoryDir() = %q\n"+
			"that path is inside the real home's project store (%q), so every run of "+
			"this package would deposit one more permanent directory there",
			got, realProjects)
	}
}

// homeJoinSiteAllowlist pins exactly WHICH files build a path of the form
// <home>/.claude/projects/<slug>. It is an explicit set, not a bare count:
// a bare count told the reader that something moved but not whether the list
// had followed the code — the release incident where internal/cli/tokens.go
// became the fifth site and the pin still said 4 shipped a deterministic red
// that nobody had context to close. Set equality turns red in BOTH sync-loss
// directions: a site added to the code without a row here, and a row here
// whose site was removed or renamed.
//
// Each site combines a real home with a project-derived name. The write-shaped
// sites each carry a behavioural guard — TestResolveCaptureMemoryDir_DoesNotEscapeToRealHome
// in internal/hook, TestDecayScan_DoesNotWriteToRealHome in internal/cli/preference,
// and, for the `moai memory` / `moai migrate profiles` derivations,
// TestMemoryCandidateStores_DoNotEscapeToRealHome and
// TestDefaultMemoryStore_DoesNotEscapeToRealHome in internal/cli. The
// tokens.go entry is a glob READ (transcript lookup), so the write-leak guard
// class does not apply to it as written — it stays pinned anyway so that a
// future edit turning the read into a write re-opens this decision instead of
// passing silently.
var homeJoinSiteAllowlist = []string{
	"internal/cli/memory.go",
	"internal/cli/migrate_profiles.go",
	"internal/cli/preference/cmd.go",
	"internal/cli/tokens.go", // glob READ of the real home (session transcript lookup)
	"internal/hook/session_end.go",
}

// TestHomeJoinSiteCountIsPinned asserts the tree's home-join sites match
// homeJoinSiteAllowlist exactly. See that allowlist's comment for the
// per-site guard mapping and the count→set rationale.
func TestHomeJoinSiteCountIsPinned(t *testing.T) {

	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repo root %s has no go.mod: %v", root, err)
	}

	var sites []string
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // unreadable subtree is not this guard's concern
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "node_modules", "worktrees", ".moai":
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if filepath.Ext(name) != ".go" || len(name) > 8 && name[len(name)-8:] == "_test.go" {
			return nil
		}
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		if containsHomeProjectsJoin(string(body)) {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			// Normalize to slash form so the allowlist compares identically on
			// every platform CI runs on (filepath.Rel yields OS separators).
			sites = append(sites, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repo: %v", err)
	}

	want := make(map[string]bool, len(homeJoinSiteAllowlist))
	for _, s := range homeJoinSiteAllowlist {
		want[s] = true
	}
	var stale, unlisted []string
	for _, s := range homeJoinSiteAllowlist {
		if !containsSite(sites, s) {
			stale = append(stale, s)
		}
	}
	for _, s := range sites {
		if !want[s] {
			unlisted = append(unlisted, s)
		}
	}
	if len(stale) > 0 || len(unlisted) > 0 {
		sort.Strings(stale)
		sort.Strings(unlisted)
		t.Errorf("home-join sites out of sync with homeJoinSiteAllowlist:\n"+
			"  unlisted (code has them, allowlist does not — add a row, and a leak guard if the site writes): %v\n"+
			"  stale (allowlist has them, code does not — update the allowlist): %v\n\n"+
			"Each such site combines a REAL home with a project-derived slug, so a test that "+
			"isolates only the project half deposits a permanent directory in the developer's "+
			"own home on every run. See TestResolveCaptureMemoryDir_DoesNotEscapeToRealHome "+
			"for the guard shape.",
			unlisted, stale)
	}
}

// containsSite reports whether sites contains s.
func containsSite(sites []string, s string) bool {
	for _, x := range sites {
		if x == s {
			return true
		}
	}
	return false
}

// containsHomeProjectsJoin reports whether src builds a .claude/projects path
// via filepath.Join. It matches the join argument spelling rather than the
// resulting path, because the path never appears as a literal.
func containsHomeProjectsJoin(src string) bool {
	const needle = `".claude", "projects"`
	for i := 0; i+len(needle) <= len(src); i++ {
		if src[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
