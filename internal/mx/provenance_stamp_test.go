package mx

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// stampGit runs git in the provenance stamp fixture repo.
func stampGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var errBuf strings.Builder
	cmd.Stderr = &errBuf
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v\nstderr: %s", args, err, errBuf.String())
	}
	return strings.TrimSpace(string(out))
}

// newStampFixture creates a committed git repo with one described-source file
// and returns (root, headSHA, parentSHA).
func newStampFixture(t *testing.T) (string, string, string) {
	t.Helper()
	root := t.TempDir()
	stampGit(t, root, "init", "-q")
	stampGit(t, root, "config", "user.email", "fixture@example.com")
	stampGit(t, root, "config", "user.name", "Fixture")
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "a.go"), []byte("package a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stampGit(t, root, "add", "-A")
	stampGit(t, root, "commit", "-q", "-m", "base")
	parent := stampGit(t, root, "rev-parse", "HEAD")
	if err := os.WriteFile(filepath.Join(root, "internal", "b.go"), []byte("package a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stampGit(t, root, "add", "-A")
	stampGit(t, root, "commit", "-q", "-m", "second")
	head := stampGit(t, root, "rev-parse", "HEAD")
	return root, head, parent
}

// dirtyStampFixture returns a fixture root whose described sources carry one
// uncommitted change versus HEAD.
func dirtyStampFixture(t *testing.T) string {
	t.Helper()
	root, _, _ := newStampFixture(t)
	if err := os.WriteFile(filepath.Join(root, "internal", "a.go"), []byte("package a // dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// testdataDirtyStampFixture returns a fixture root whose only uncommitted
// change is a committed testdata fixture file edited in place — no
// described-worthy change at all under REQ-GFC-002.
func testdataDirtyStampFixture(t *testing.T) string {
	t.Helper()
	root, _, _ := newStampFixture(t)
	tdDir := filepath.Join(root, "internal", "testdata")
	if err := os.MkdirAll(tdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join(tdDir, "f.txt")
	if err := os.WriteFile(fixture, []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stampGit(t, root, "add", "internal/testdata/f.txt")
	stampGit(t, root, "commit", "-q", "-m", "testdata fixture")
	if err := os.WriteFile(fixture, []byte("v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestResolveCommit covers REQ-SR-005's resolution contract: full shas, short
// revs, and ref names resolve to the full sha; unresolvable input errors
// without naming an absolute local path.
func TestResolveCommit(t *testing.T) {
	root, head, parent := newStampFixture(t)

	cases := []struct {
		name string
		rev  string
		want string
	}{
		{"full sha", head, head},
		{"short rev", head[:7], head},
		{"ref name", "HEAD", head},
		{"named parent sha", parent, parent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveCommit(root, tc.rev)
			if err != nil {
				t.Fatalf("ResolveCommit(%q): %v", tc.rev, err)
			}
			if got != tc.want {
				t.Errorf("ResolveCommit(%q) = %q, want %q", tc.rev, got, tc.want)
			}
		})
	}

	if _, err := ResolveCommit(root, "deadbeef42"); err == nil {
		t.Fatal("ResolveCommit(deadbeef42) must fail on an unresolvable revision")
	} else if strings.Contains(err.Error(), root) {
		t.Errorf("unresolvable-revision error leaks the absolute local path: %v", err)
	}
}

// TestStampCodemaps_ExplicitCommit pins REQ-SR-005: an explicit full sha on a
// clean described-source tree is recorded verbatim, dirty stays false, and the
// detected HEAD does NOT override the named commit.
func TestStampCodemaps_ExplicitCommit(t *testing.T) {
	root, head, parent := newStampFixture(t)

	pv, err := StampCodemaps(root, parent)
	if err != nil {
		t.Fatalf("StampCodemaps explicit: %v", err)
	}
	if pv.CommitSHA != parent {
		t.Errorf("commit_sha = %q, want the named parent %q (not HEAD %q)", pv.CommitSHA, parent, head)
	}
	if pv.Dirty {
		t.Errorf("clean tree with --commit must not set dirty")
	}
	if pv.SchemaVersion != ProvenanceSchemaVersion {
		t.Errorf("schema_version = %d, want %d", pv.SchemaVersion, ProvenanceSchemaVersion)
	}
	if pv.ContentFingerprint != "" {
		t.Errorf("clean anchor must not carry a content fingerprint")
	}
}

// TestStampCodemaps_ExplicitCommitDirtyRejected pins REQ-SR-006 at the mx
// boundary: a named commit plus a dirty described-source tree are mutually
// exclusive honesty claims — the entry rejects before any install path.
func TestStampCodemaps_ExplicitCommitDirtyRejected(t *testing.T) {
	root := dirtyStampFixture(t)
	_, head, _ := newStampFixture(t) // any valid sha shape works; reuse helper

	pv, err := StampCodemaps(root, head)
	if err == nil {
		t.Fatal("explicit commit + dirty tree must be rejected")
	}
	if pv != nil {
		t.Errorf("rejected stamp must return a nil provenance, got %+v", pv)
	}
	if strings.Contains(err.Error(), root) {
		t.Errorf("mixed-anchor rejection leaks the absolute local path: %v", err)
	}
	if !strings.Contains(err.Error(), "dirty") && !strings.Contains(err.Error(), "uncommitted") {
		t.Errorf("mixed-anchor rejection must name the conflict, got: %v", err)
	}
}

// TestStampCodemaps_DefaultPathUnchanged is the flagless characterization
// (REQ-SR-007 / AC-SP-008): absent an explicit commit, the anchor semantics
// are exactly the pre-SPEC behavior — clean tree records HEAD; dirty tree
// records dirty + 64-hex aggregate fingerprint with no commit anchor.
func TestStampCodemaps_DefaultPathUnchanged(t *testing.T) {
	root, head, _ := newStampFixture(t)

	pv, err := StampCodemaps(root, "")
	if err != nil {
		t.Fatalf("default path must not error on a clean tree: %v", err)
	}
	if pv.CommitSHA != head || pv.Dirty {
		t.Errorf("clean default anchor = (%q, dirty=%v), want (%q, false)", pv.CommitSHA, pv.Dirty, head)
	}
	want := "provenance: tree=" + root + " commit=" + head[:12]
	if got := pv.Describe(); got != want {
		t.Errorf("clean Describe = %q, want %q", got, want)
	}

	dirty := dirtyStampFixture(t)
	dpv, err := StampCodemaps(dirty, "")
	if err != nil {
		t.Fatalf("default path must not error on a dirty tree: %v", err)
	}
	if dpv.CommitSHA != "" || !dpv.Dirty {
		t.Errorf("dirty default anchor = (%q, dirty=%v), want empty commit + dirty", dpv.CommitSHA, dpv.Dirty)
	}
	if len(dpv.ContentFingerprint) != 64 {
		t.Errorf("dirty anchor fingerprint len = %d, want 64 hex chars", len(dpv.ContentFingerprint))
	}
	if want, got := 3, len(dpv.DescribedRoots); got != want {
		t.Errorf("described_roots len = %d, want %d (%v)", got, want, dpv.DescribedRoots)
	}
}

// TestStampCodemaps_ExplicitCommitAllowsTestdataOnlyDirty closes the residual
// SPEC-GRAPH-FRESHNESS-CADENCE-001 deferred at v0.2.1 (card t327): the anchor
// gate must read the REQ-GFC-002 predicate the codemaps fingerprint already
// applies, so a tree dirty only under a testdata directory carries no anchor
// contradiction and the --commit merge-base anchor is not refused.
func TestStampCodemaps_ExplicitCommitAllowsTestdataOnlyDirty(t *testing.T) {
	root := testdataDirtyStampFixture(t)
	_, head, _ := newStampFixture(t)

	pv, err := StampCodemaps(root, head)
	if err != nil {
		t.Fatalf("testdata-only dirty must not refuse the --commit anchor: %v", err)
	}
	if pv.Dirty {
		t.Errorf("testdata-only dirty must not set dirty")
	}
	if pv.CommitSHA != head {
		t.Errorf("commit_sha = %q, want the named commit %q", pv.CommitSHA, head)
	}
}

// TestStampCodemaps_ExplicitCommitRejectsUntrackedDescribedSource guards the
// --untracked-files=all half of the same fix: a new untracked described source
// inside a fresh directory must still read dirty (the default porcelain view
// collapses it to "dir/", whose bare directory path the predicate would
// reject).
func TestStampCodemaps_ExplicitCommitRejectsUntrackedDescribedSource(t *testing.T) {
	root, head, _ := newStampFixture(t)
	sub := filepath.Join(root, "internal", "newpkg")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "c.go"), []byte("package newpkg\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := StampCodemaps(root, head); err == nil {
		t.Fatal("untracked described source must keep refusing the --commit anchor")
	}
}

// TestStampCodemaps_DefaultPathTestdataOnlyDirtyRecordsCommit extends the
// flagless characterization the same way: a testdata-only dirty tree records
// the commit anchor, not a dirty content fingerprint.
func TestStampCodemaps_DefaultPathTestdataOnlyDirtyRecordsCommit(t *testing.T) {
	root, _, _ := newStampFixture(t)
	tdDir := filepath.Join(root, "internal", "testdata")
	if err := os.MkdirAll(tdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join(tdDir, "f.txt")
	if err := os.WriteFile(fixture, []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stampGit(t, root, "add", "internal/testdata/f.txt")
	stampGit(t, root, "commit", "-q", "-m", "testdata fixture")
	stampedHead := stampGit(t, root, "rev-parse", "HEAD")
	if err := os.WriteFile(fixture, []byte("v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	pv, err := StampCodemaps(root, "")
	if err != nil {
		t.Fatalf("flagless path must not error: %v", err)
	}
	if pv.CommitSHA != stampedHead || pv.Dirty {
		t.Errorf("testdata-only dirty default anchor = (%q, dirty=%v), want (%q, false)", pv.CommitSHA, pv.Dirty, stampedHead)
	}
	if pv.ContentFingerprint != "" {
		t.Errorf("commit anchor must not carry a content fingerprint, got %q", pv.ContentFingerprint)
	}
}
