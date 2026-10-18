//go:build !windows

package cli

// migrate_ownership_unregsymlink_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 (card
// t1865, sync-audit finding F1; REQ-OWR-005, AC-OWR-08 family): the verdict
// function must not read bytes from outside the project through a symlink.
//
// The common-asset walk hands the verdict function every non-directory entry,
// and a symlink is one. os.ReadFile follows it, so an UNREGISTERED symlink that
// points at a regular file in a profile tree (skills/ or agents/) had that
// file's bytes read before any root-boundary test ran, even with the removal
// switch off. REQ-OWR-005 says nothing in the update path opens or reads a
// profile root; the source scan and the differential canary of AC-OWR-08 cannot
// see this route (no fixture put a profile-pointing symlink under a common
// asset root, and nothing assembled a path at run time).
//
// Expected behaviour, pinned here (chunk 13, leader disposition B: the guarded
// surface is reduced): an unregistered path with a symlink among its components is
// kept unread, class `unregistered`, reason `symlink-escape`, with a row in the
// remaining list — whether the link points inside or outside the project. The one
// baseline that CHANGED by that decision: a symlink to an in-project regular file
// used to be judged by its target's bytes (chunk 6 measured it) and is now kept
// unread. Dangling links and loops still fail their read (`read-failed`), and
// ordinary files are judged by content as ever.
//
// The registered-file path is NOT covered here: it read before the boundary
// check at the base commit as well, and its pinned expectation (class managed +
// `symlink-escape`, TestOwnershipVerdictSymlinkEscape) needs the target's bytes
// to establish the class. progress.md records it as a disclosed debt.

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

const (
	usDir        = ".agents/skills/moai-unreg-symlink"
	usToProfile  = usDir + "/to-profile" // symlink to a regular file under a profile skills tree
	usChained    = usDir + "/chained"    // symlink to an in-project symlink that leaves the project
	usHop        = usDir + "/hop"        // the in-project link of the chain
	usToInside   = usDir + "/to-inside"  // symlink to a regular file inside the project
	usInsideFile = usDir + "/inside.md"  // that file (unregistered)
	usDangling   = usDir + "/to-nothing" // symlink to nothing
	usProfileKey = "synthetic-profile-private-content\n"
	usInsideKey  = "# an unregistered file inside the project\n"
)

type unregSymlinkWorld struct {
	root, home  string
	profileFile string // absolute path of the profile-shaped regular file
}

func unregSymlinkFixture(t *testing.T) unregSymlinkWorld {
	t.Helper()
	root, home := ownershipFixtureBase(t) // switch on inside this fixture only
	profileFile := filepath.Join(t.TempDir(), "profiles", "work", "skills", "canary", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(profileFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(profileFile, []byte(usProfileKey), 0o600); err != nil {
		t.Fatal(err)
	}
	abs := func(rel string) string { return filepath.Join(root, filepath.FromSlash(rel)) }
	ownAddUnregistered(t, root, usInsideFile, []byte(usInsideKey))
	if err := os.Symlink(profileFile, abs(usToProfile)); err != nil {
		t.Fatalf("os.Symlink(%s) failed: %v", abs(usToProfile), err)
	}
	if err := os.Symlink(profileFile, abs(usHop)); err != nil {
		t.Fatalf("os.Symlink(%s) failed: %v", abs(usHop), err)
	}
	if err := os.Symlink("hop", abs(usChained)); err != nil {
		t.Fatalf("os.Symlink(%s) failed: %v", abs(usChained), err)
	}
	if err := os.Symlink("inside.md", abs(usToInside)); err != nil {
		t.Fatalf("os.Symlink(%s) failed: %v", abs(usToInside), err)
	}
	if err := os.Symlink("does-not-exist", abs(usDangling)); err != nil {
		t.Fatalf("os.Symlink(%s) failed: %v", abs(usDangling), err)
	}
	return unregSymlinkWorld{root: root, home: home, profileFile: profileFile}
}

func TestUnregisteredSymlinkOutsideProjectIsNotRead(t *testing.T) {
	w := unregSymlinkFixture(t)
	abs := func(rel string) string { return filepath.Join(w.root, filepath.FromSlash(rel)) }
	rootResolved, err := filepath.EvalSymlinks(w.root)
	if err != nil {
		t.Fatal(err)
	}

	// Verdict level: the function under test, with every removal input satisfied
	// so that nothing but the symlink decides the outcome.
	for _, enabled := range []bool{false, true} {
		env := ownershipEnv{
			entry:                func(string) (*manifest.FileEntry, bool) { return nil, false },
			evidence:             fakeProvenance{},
			counterpartConfirmed: func(string) bool { return true },
			notInCatalog:         func(string) bool { return false },
			removalEnabled:       enabled,
			rootResolved:         rootResolved,
		}
		for _, rel := range []string{usToProfile, usChained} {
			v := classifyOwnership(env, rel, abs(rel))
			if len(v.data) != 0 || bytes.Contains(v.data, []byte(strings.TrimSpace(usProfileKey))) {
				t.Errorf("switch=%v %s: the verdict function read %d bytes from outside the project", enabled, rel, len(v.data))
			}
			if v.class != ownershipUnregistered || v.keep != keepSymlinkEscape || v.removalEligible {
				t.Errorf("switch=%v %s: class=%q keep=%q eligible=%v, want unregistered/%q/false", enabled, rel, v.class, v.keep, v.removalEligible, keepSymlinkEscape)
			}
		}
		// The changed baseline: an in-project link is kept unread as well.
		in := classifyOwnership(env, usToInside, abs(usToInside))
		if in.class != ownershipUnregistered || in.keep != keepSymlinkEscape || len(in.data) != 0 {
			t.Errorf("switch=%v in-project symlink: class=%q keep=%q data=%q, want unregistered/%q/no bytes", enabled, in.class, in.keep, in.data, keepSymlinkEscape)
		}
		// Control: the in-project regular file itself is judged by content.
		plain := classifyOwnership(env, usInsideFile, abs(usInsideFile))
		if plain.class != ownershipUnregistered || plain.keep != keepNoProvenanceRecord || string(plain.data) != usInsideKey {
			t.Errorf("switch=%v in-project file: class=%q keep=%q data=%q, want unregistered/%q/its bytes", enabled, plain.class, plain.keep, plain.data, keepNoProvenanceRecord)
		}
		dang := classifyOwnership(env, usDangling, abs(usDangling))
		if dang.class != ownershipUnregistered || dang.keep != keepReadFailed || dang.data != nil {
			t.Errorf("switch=%v dangling symlink: class=%q keep=%q data=%v, want unregistered/%q/nil", enabled, dang.class, dang.keep, dang.data, keepReadFailed)
		}
	}

	// Flow level: the real walk. A wrapper records the bytes each verdict carried
	// (a read leaves them in v.data), then hands the verdict on unchanged.
	carried := map[string]int{}
	swapClassifier(t, func(env ownershipEnv, rel, p string) ownershipVerdict {
		v := classifyOwnership(env, rel, p)
		carried[rel] = len(v.data)
		return v
	})
	rc := &migrationRunContext{userAssetsInstalled: true}
	removed, err := migrateProjectCommonAssets(w.root, w.home, true, nil, func(string, ...interface{}) {}, rc)
	if err != nil {
		t.Fatalf("migration: %v", err)
	}
	if !slices.Equal(removed, []string{migIdenticalSkill}) {
		t.Errorf("removed %v, want exactly [%s] (the reach control: the same walk could remove)", removed, migIdenticalSkill)
	}
	for _, rel := range []string{usToProfile, usChained} {
		r, ok := findRemaining(rc, rel)
		switch {
		case !ok:
			t.Errorf("%s has no row in the remaining list", rel)
		case r.Class != ownershipUnregistered || r.Reason != keepSymlinkEscape:
			t.Errorf("%s: row class=%q reason=%q, want unregistered/%q", rel, r.Class, r.Reason, keepSymlinkEscape)
		}
		if n, seen := carried[rel]; !seen {
			t.Errorf("%s was never classified: the no-read check below would prove nothing", rel)
		} else if n != 0 {
			t.Errorf("%s: the verdict carried %d bytes read through the symlink", rel, n)
		}
	}
	// Nothing outside the project, and no link, was touched.
	if got, err := os.ReadFile(w.profileFile); err != nil || string(got) != usProfileKey {
		t.Errorf("the profile-shaped file changed (%v)", err)
	}
	for _, rel := range []string{usToProfile, usChained, usHop, usToInside, usDangling} {
		if info, err := os.Lstat(abs(rel)); err != nil || info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("%s is no longer a symlink (%v)", rel, err)
		}
	}
	// The in-project link is kept unread too (changed baseline); its target is judged as itself.
	if r, ok := findRemaining(rc, usToInside); !ok || r.Reason != keepSymlinkEscape {
		t.Errorf("%s: row %+v (found=%v), want reason %q", usToInside, r, ok, keepSymlinkEscape)
	}
	if r, ok := findRemaining(rc, usInsideFile); !ok || r.Reason != keepNoProvenanceRecord {
		t.Errorf("%s: row %+v (found=%v), want reason %q", usInsideFile, r, ok, keepNoProvenanceRecord)
	}
	if r, ok := findRemaining(rc, usDangling); !ok || r.Reason != keepReadFailed {
		t.Errorf("%s: row %+v (found=%v), want reason %q", usDangling, r, ok, keepReadFailed)
	}
}
