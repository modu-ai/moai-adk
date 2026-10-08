// fix_round3_addendum_test.go — addendum items 1-3 (card t1509 review-fix
// round 3 addendum).
package userassets

import (
	"os"
	"path/filepath"
	"testing"
)

// Item 2: an unknown root slug in a manifest key refuses deletion — the
// former fallthrough resolved to an EMPTY root dir and os.Remove deleted
// against the CWD (external sentinel deleted, Removed:1 reproduced).
func TestFR3A2_UnknownSlugDeletionRefused(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	if err := os.MkdirAll(MoaiHome(f.home), 0o755); err != nil {
		t.Fatal(err)
	}
	// An external sentinel in the CWD-adjacent area + a manifest record
	// with an unknown slug.
	outside := t.TempDir()
	sentinel := filepath.Join(outside, "sentinel.md")
	if err := os.WriteFile(sentinel, []byte("external\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, _ := Load(ManifestPath(f.home))
	m.Files["unknown-root/outside/sentinel.md"] = FileEntry{
		SHA256: sha256Hex([]byte("external\n")), Bundle: "core",
		InstalledAt: "t0", MoaiVersion: "v0",
	}

	in := f.installer(t)
	res, err := in.PruneUnselected(m)
	if err != nil {
		t.Fatal(err)
	}
	if got := readBytes(t, sentinel); string(got) != "external\n" {
		t.Errorf("Item 2: external sentinel DELETED via unknown slug: %q", got)
	}
	if len(res.Failures) == 0 {
		t.Error("Item 2: refusal not reported as a failure")
	}
}
