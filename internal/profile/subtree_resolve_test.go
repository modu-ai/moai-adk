package profile

import (
	"os"
	"path/filepath"
	"testing"
)

// Subtree-aware resolution (SPEC-STATUSLINE-PROFILE-RESPECT-001 REQ-006..008).
// The ledger registers exact paths; a fresh worktree has no entry of its own,
// so before this change a session inside a registered project's subtree
// launched on the anonymous/fallback profile even though the enclosing project
// WAS registered — the "폴백 프로필" of issue #1675. These tests pin that the
// miss path of ResolveLaunchProfileForProject walks the session directory's
// real path-segment ancestors and resolves to the deepest registered one.

// subtreeSandbox wires BaseDirOverride, creates the named profile directories,
// and returns the base.
func subtreeSandbox(t *testing.T, profiles ...string) string {
	t.Helper()
	base := pmSandboxBase(t)
	for _, p := range profiles {
		pmMkProfile(t, base, p)
	}
	return base
}

// subtreeMkdirAll creates dir (a real subtree, so the ancestor walk's Stat
// checks can succeed) and fails the test if it cannot.
func subtreeMkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create %q: %v", dir, err)
	}
}

// subtreeLedger writes a projects map from ordered (path, profile) pairs so
// each test's ledger is readable at a glance. Keys go through
// normalizeProjectKey because that is what the WRITE side records — a ledger
// keyed by a raw /var spelling would never match the read side's resolved
// /private/var walk on macOS (the same asymmetry normalizeProjectKey exists
// to prevent).
func subtreeLedger(t *testing.T, base string, entries ...[2]string) {
	t.Helper()
	body := "projects:\n"
	for _, e := range entries {
		body += "  " + normalizeProjectKey(e[0]) + ": " + e[1] + "\n"
	}
	pmWriteLedger(t, base, body)
}

// TestResolveLaunchProfileForProject_SubtreeWorktreeResolves (AC-006, REQ-006):
// a session directory inside a registered project's subtree — a worktree path
// with no entry of its own — resolves to that project's profile.
func TestResolveLaunchProfileForProject_SubtreeWorktreeResolves(t *testing.T) {
	base := subtreeSandbox(t, "alpha")

	proj := filepath.Join(t.TempDir(), "proj")
	session := filepath.Join(proj, ".claude", "worktrees", "t999")
	subtreeMkdirAll(t, session)
	subtreeLedger(t, base, [2]string{proj, "alpha"})

	if got := ResolveLaunchProfileForProject(session, ""); got != "alpha" {
		t.Errorf("subtree session resolved to %q, want %q", got, "alpha")
	}
}

// TestResolveLaunchProfileForProject_DeepestRegisteredAncestorWins (AC-007,
// REQ-007): with two registered projects nested inside one another, the deeper
// registration serves the deeper session.
func TestResolveLaunchProfileForProject_DeepestRegisteredAncestorWins(t *testing.T) {
	base := subtreeSandbox(t, "alpha", "beta")

	proj := filepath.Join(t.TempDir(), "proj")
	inner := filepath.Join(proj, "sub", "inner")
	subtreeMkdirAll(t, inner)
	subtreeLedger(t, base,
		[2]string{proj, "alpha"},
		[2]string{filepath.Join(proj, "sub"), "beta"},
	)

	if got := ResolveLaunchProfileForProject(inner, ""); got != "beta" {
		t.Errorf("nested subtree resolved to %q, want %q (deepest registered ancestor)", got, "beta")
	}

	// A session directly under the outer root still takes the outer profile:
	// exact-match semantics for the registered root itself are unchanged.
	outer := filepath.Join(proj, "other")
	subtreeMkdirAll(t, outer)
	if got := ResolveLaunchProfileForProject(proj, ""); got != "alpha" {
		t.Errorf("registered root resolved to %q, want %q (exact match preserved)", got, "alpha")
	}
}

// TestResolveLaunchProfileForProject_LexicalPrefixIsNotAnAncestor (AC-008,
// REQ-008): a directory sharing a string prefix with a registration but on a
// different path segment (`/proj` vs `/proj-other`) must NOT resolve to it.
// The guard is the filepath.Dir chain, never prefix matching.
func TestResolveLaunchProfileForProject_LexicalPrefixIsNotAnAncestor(t *testing.T) {
	base := subtreeSandbox(t, "alpha")

	tmp := t.TempDir()
	proj := filepath.Join(tmp, "proj")
	other := filepath.Join(tmp, "proj-other")
	subtreeMkdirAll(t, proj)
	subtreeMkdirAll(t, other)
	subtreeLedger(t, base, [2]string{proj, "alpha"})

	if got := ResolveLaunchProfileForProject(other, ""); got != "" {
		t.Errorf("sibling with shared prefix resolved to %q, want \"\" — a lexical prefix is not a subtree", got)
	}
}

// TestResolveLaunchProfileForProject_SubtreeMissFallsThrough (miss-path
// characterization): a session directory registered nowhere and inside nothing
// registered keeps returning "" — the default-profile semantics that predate
// the walk.
func TestResolveLaunchProfileForProject_SubtreeMissFallsThrough(t *testing.T) {
	base := subtreeSandbox(t, "alpha")

	proj := filepath.Join(t.TempDir(), "registered")
	nowhere := filepath.Join(t.TempDir(), "elsewhere", "deep", "down")
	subtreeMkdirAll(t, proj)
	subtreeMkdirAll(t, nowhere)
	subtreeLedger(t, base, [2]string{proj, "alpha"})

	if got := ResolveLaunchProfileForProject(nowhere, ""); got != "" {
		t.Errorf("unregistered directory resolved to %q, want \"\" (default semantics)", got)
	}
}

// TestResolveLaunchProfileForProject_StaleAncestorSkippedNotDeadEnd (§D.2 edge
// case): a registered ancestor whose profile directory has vanished must not
// dead-end the walk — resolution continues outward to the next registered
// ancestor, and a session under a stale registered path itself (the deleted
// worktree case) resolves through the enclosing live project.
func TestResolveLaunchProfileForProject_StaleAncestorSkippedNotDeadEnd(t *testing.T) {
	base := subtreeSandbox(t, "alpha") // note: no "ghost" profile dir created

	proj := filepath.Join(t.TempDir(), "proj")
	session := filepath.Join(proj, "dead-worktree", "inner")
	subtreeMkdirAll(t, session)
	subtreeLedger(t, base,
		[2]string{filepath.Join(proj, "dead-worktree"), "ghost"}, // profile dir absent
		[2]string{proj, "alpha"},
	)

	if got := ResolveLaunchProfileForProject(session, ""); got != "alpha" {
		t.Errorf("session under a stale registration resolved to %q, want %q (walk continues past the unusable entry)", got, "alpha")
	}
}

// TestResolveLaunchProfileForProject_ExactMatchBeatsSubtree pins the
// precedence the walk must not disturb: when the session directory has an
// entry of its own, that entry wins over any registered ancestor.
func TestResolveLaunchProfileForProject_ExactMatchBeatsSubtree(t *testing.T) {
	base := subtreeSandbox(t, "alpha", "own")

	proj := filepath.Join(t.TempDir(), "proj")
	session := filepath.Join(proj, "sub")
	subtreeMkdirAll(t, session)
	subtreeLedger(t, base,
		[2]string{proj, "alpha"},
		[2]string{session, "own"},
	)

	if got := ResolveLaunchProfileForProject(session, ""); got != "own" {
		t.Errorf("exact entry resolved to %q, want %q — exact match outranks the ancestor walk", got, "own")
	}
}
