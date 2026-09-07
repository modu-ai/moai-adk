package kanban

// temp_origin_test.go — SPEC-TODO-HOME-TEMP-GUARD-001 M1: the temporary-origin
// discriminant's own tests. Producing acceptance criteria: AC-THG-002 (the two
// symlink spellings classify alike), AC-THG-004 (fail-open), AC-THG-006 (the
// boundary is component-wise), plus the REQ-THG-009 seam's two M1 exit clauses
// — that a test can stub the temp-root set at all, and that the stub is read at
// CALL time rather than snapshotted at package init.
//
// Every test here overrides the package-global TempRootsFn seam, so none of
// them run in parallel — the same constraint the HomeDirFn tests carry.

import (
	"os"
	"path/filepath"
	"testing"
)

// stubTempRoots points the package's temp-root seam at roots for the test's
// duration. It is the REQ-THG-009 escape from this repository's [HARD]
// isolation discipline: every fixture lives under t.TempDir(), which is by
// definition inside os.TempDir(), so without this a "non-temporary base"
// fixture cannot exist.
func stubTempRoots(t *testing.T, roots ...string) {
	t.Helper()
	orig := TempRootsFn
	TempRootsFn = func() []string { return roots }
	t.Cleanup(func() { TempRootsFn = orig })
}

// TestTempOrigin_SymlinkSpellingEquivalence — AC-THG-002.
//
// The unresolved spelling (what os.TempDir() reports: /var/folders/... on
// macOS) and the resolved spelling (/private/var/folders/...) name the SAME
// directory, so the discriminant must classify them alike and name the same
// matched root. A one-sided comparison — normalizing the base but not the root,
// or strings.HasPrefix(abs, os.TempDir()) — fails on exactly one of the two and
// fails SILENTLY.
func TestTempOrigin_SymlinkSpellingEquivalence(t *testing.T) {
	unresolved := t.TempDir()
	resolved, err := filepath.EvalSymlinks(unresolved)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", unresolved, err)
	}
	// Recorded, not asserted: on a platform where the temp tree sits behind no
	// symlink the two spellings coincide and this test degenerates to a single
	// case. The log says which cell this run measured.
	if resolved == unresolved {
		t.Logf("spellings coincide on this platform (%s): %q", os.Getenv("GOOS")+runtimeNote(), unresolved)
	} else {
		t.Logf("spellings differ: unresolved=%q resolved=%q", unresolved, resolved)
	}

	reasonUnresolved, tempUnresolved := TempOriginReason(unresolved)
	if !tempUnresolved {
		t.Errorf("unresolved spelling %q classified NOT temporary", unresolved)
	}
	reasonResolved, tempResolved := TempOriginReason(resolved)
	if !tempResolved {
		t.Errorf("resolved spelling %q classified NOT temporary", resolved)
	}
	if reasonUnresolved != reasonResolved {
		t.Errorf("the two spellings named different temp roots: %q vs %q",
			reasonUnresolved, reasonResolved)
	}
	if reasonUnresolved == "" {
		t.Errorf("a temporary verdict must name the matched root, got empty reason")
	}
}

// TestTempOrigin_FailsOpenOnUnresolvable — AC-THG-004.
//
// Inputs the discriminant cannot resolve are reported NOT temporary, and no
// input panics or propagates an error. The asymmetry is the point: a false
// negative keeps today's behaviour, a false positive withdraws a real
// project's queue.
func TestTempOrigin_FailsOpenOnUnresolvable(t *testing.T) {
	nonTemp := filepath.Join(string(filepath.Separator), "moai-t536-does-not-exist", "deep", "path")

	cases := []struct {
		name string
		base string
	}{
		{"empty", ""},
		{"whitespace only", "   "},
		{"non-existent absolute path outside every temp root", nonTemp},
		{"non-existent relative path outside every temp root", filepath.Join("..", "..", "moai-t536-nope")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reason, isTemp := TempOriginReason(tc.base)
			if isTemp {
				t.Errorf("base %q classified temporary (reason %q); fail-open requires NOT temporary",
					tc.base, reason)
			}
		})
	}

	// A root the seam hands over that cannot be resolved must be skipped rather
	// than crash the classification or flip its direction.
	t.Run("unusable roots in the injected set", func(t *testing.T) {
		stubTempRoots(t, "", "   ", filepath.Join(t.TempDir(), "never-created"))
		base := t.TempDir()
		if reason, isTemp := TempOriginReason(base); isTemp {
			t.Errorf("base %q matched an unusable root set (reason %q)", base, reason)
		}
	})
}

// TestTempOrigin_ComponentBoundary — AC-THG-006.
//
// The comparison is by path component: a sibling that merely shares the root's
// string prefix is outside, while the root itself and everything beneath it is
// inside.
func TestTempOrigin_ComponentBoundary(t *testing.T) {
	root := t.TempDir()
	stubTempRoots(t, root)

	if _, isTemp := TempOriginReason(root); !isTemp {
		t.Errorf("the temp root itself %q classified NOT temporary", root)
	}
	child := filepath.Join(root, "a", "b")
	if _, isTemp := TempOriginReason(child); !isTemp {
		t.Errorf("a path beneath the temp root %q classified NOT temporary", child)
	}
	sibling := root + "foo" // shares the prefix, is not beneath it
	if reason, isTemp := TempOriginReason(sibling); isTemp {
		t.Errorf("sibling %q classified temporary against root %q (reason %q) — "+
			"raw string-prefix matching", sibling, root, reason)
	}

	// The literal shape spec.md §4 names, measured against the production root
	// set rather than a stub.
	t.Run("production roots: /tmpfoo is not inside /tmp", func(t *testing.T) {
		if reason, isTemp := TempOriginReason(filepath.Join(string(filepath.Separator), "tmpfoo")); isTemp {
			t.Errorf("/tmpfoo classified temporary (reason %q)", reason)
		}
	})
}

// TestTempRootsSeam_ReadAtCallTime — REQ-THG-009, and the M1 exit condition's
// third clause.
//
// The verdict on ONE unchanged base must change when the stub is installed. An
// implementation that snapshotted the root set at package-init time would
// ignore the stub silently, keeping the verdict identical — and that single
// shape is what would make a whole family of later acceptance assertions
// vacuous, since a stub that is not read cannot make anything RED.
func TestTempRootsSeam_ReadAtCallTime(t *testing.T) {
	base := t.TempDir() // by definition under os.TempDir()

	reasonBefore, tempBefore := TempOriginReason(base)
	if !tempBefore {
		t.Fatalf("precondition: %q must be temporary under the production root set (reason %q)",
			base, reasonBefore)
	}

	// A root set that does not contain base: the seam, if read, must flip the
	// verdict to non-temporary.
	elsewhere := filepath.Join(t.TempDir(), "isolated-root")
	stubTempRoots(t, elsewhere)

	reasonAfter, tempAfter := TempOriginReason(base)
	if tempAfter {
		t.Fatalf("the stub was ignored: %q still temporary (reason %q) after the root set "+
			"was replaced with %q — the set is being read at init time, not at call time",
			base, reasonAfter, elsewhere)
	}
	t.Logf("call-time evidence for base %q: before stub (isTemp=%v reason=%q) -> "+
		"after stub (isTemp=%v reason=%q)", base, tempBefore, reasonBefore, tempAfter, reasonAfter)

	// The other direction: a base the production set would never match becomes
	// temporary when the injected set names its parent. Only a call-time read
	// can produce both flips.
	outside := filepath.Join(string(filepath.Separator), "moai-t536-synthetic-root", "project")
	stubTempRoots(t, filepath.Join(string(filepath.Separator), "moai-t536-synthetic-root"))
	if reason, isTemp := TempOriginReason(outside); !isTemp {
		t.Errorf("injected root did not classify %q as temporary (reason %q)", outside, reason)
	}
}

// runtimeNote keeps the spelling log honest when GOOS is not in the
// environment (it usually is not at test time).
func runtimeNote() string {
	if os.Getenv("GOOS") == "" {
		return "GOOS unset in env; see go env GOOS"
	}
	return ""
}
