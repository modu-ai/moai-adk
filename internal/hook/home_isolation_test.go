package hook

import (
	"os"
	"path/filepath"
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

// TestHomeJoinSiteCountIsPinned pins how many places in the tree build a path of
// the form <home>/.claude/projects/<slug>.
//
// Each such site is a potential silent leak: it combines a real home with a
// project-derived name, so any test that isolates only the project half writes
// into the developer's actual home. The two known sites each carry a behavioural
// guard — this test in internal/hook, and TestDecayScan_DoesNotWriteToRealHome in
// internal/cli/preference. A third site added without one would reintroduce the
// class with nothing to report it, which is what this count exists to prevent.
//
// This is a count, not an allowlist of paths, so it stays readable; the failure
// message carries the located sites so the reader sees which one is new.
func TestHomeJoinSiteCountIsPinned(t *testing.T) {
	const wantSites = 2

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
			sites = append(sites, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repo: %v", err)
	}

	if len(sites) != wantSites {
		t.Errorf("found %d file(s) joining a home dir with .claude/projects, want %d:\n  %v\n\n"+
			"Each such site combines a REAL home with a project-derived slug, so a test that "+
			"isolates only the project half deposits a permanent directory in the developer's "+
			"own home on every run. If you added a site, add a leak guard for it (see "+
			"TestResolveCaptureMemoryDir_DoesNotEscapeToRealHome) and raise wantSites. If you "+
			"removed one, lower wantSites.",
			len(sites), wantSites, sites)
	}
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
