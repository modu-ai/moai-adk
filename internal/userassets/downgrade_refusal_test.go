// downgrade_refusal_test.go — U-11 (SPEC-UPDATE-REPAIR-RC30-001, REQ-RC30-012,
// AC-RC30-12): an older binary never refreshes a user-root file whose recorded
// moai version is newer than the binary's own.
package userassets

import (
	"path/filepath"
	"testing"
	"testing/fstest"
)

// TestUserRootDowngradeRefused judges AC-RC30-12. A first run records the
// claude skill under v3.2.0-rc.30; the shipped copy then changes, so the file
// on disk still matches its recorded hash and the shipped copy would refresh it
// (stateManifestMatch). An older binary (v3.2.0-rc.29) must refuse that
// refresh: the bytes and the record stay as recorded, and the refusal shows up
// as a per-file failure. A newer binary (v3.2.0-rc.31) still refreshes the same
// fixture, so the refusal is the version comparison, not a blanket stop.
func TestUserRootDowngradeRefused(t *testing.T) {
	t.Parallel()
	const relSkill = ".claude/skills/moai-alpha/SKILL.md"
	const manifestKey = "claude-skills/moai-alpha/SKILL.md"
	alphaV2 := []byte("alpha skill body v2\n")

	setup := func(t *testing.T) *fixture {
		t.Helper()
		f := newFixture(t)
		recorded := f.installer(t)
		recorded.MoaiVersion = "v3.2.0-rc.30"
		if _, err := recorded.Install(nil); err != nil {
			t.Fatalf("record run under v3.2.0-rc.30: %v", err)
		}
		f.src[relSkill] = &fstest.MapFile{Data: alphaV2}
		return f
	}

	t.Run("older binary refuses the refresh", func(t *testing.T) {
		f := setup(t)
		older := f.installer(t)
		older.MoaiVersion = "v3.2.0-rc.29"
		res, err := older.Install(nil)
		if err != nil {
			t.Fatalf("Install under v3.2.0-rc.29: %v", err)
		}
		if got := readBytes(t, filepath.Join(f.home, relSkill)); string(got) != string(f.alphaV1) {
			t.Errorf("older binary refreshed the file: bytes %q, want the recorded %q", got, f.alphaV1)
		}
		m, err := Load(ManifestPath(f.home))
		if err != nil {
			t.Fatalf("load manifest: %v", err)
		}
		if fe := m.Files[manifestKey]; fe.MoaiVersion != "v3.2.0-rc.30" || fe.SHA256 != sha256Hex(f.alphaV1) {
			t.Errorf("older binary rewrote the record: %+v, want the v3.2.0-rc.30 record of the recorded bytes", fe)
		}
		if res.Refreshed != 0 {
			t.Errorf("Refreshed = %d, want 0: a refused refresh is not counted", res.Refreshed)
		}
		reported := false
		for _, o := range res.Failures {
			if o.Path == manifestKey {
				reported = true
			}
		}
		if !reported {
			t.Errorf("refusal is not observable: no failure for %s in %+v", manifestKey, res.Failures)
		}
	})

	t.Run("newer binary still refreshes the same fixture", func(t *testing.T) {
		f := setup(t)
		newer := f.installer(t)
		newer.MoaiVersion = "v3.2.0-rc.31"
		if _, err := newer.Install(nil); err != nil {
			t.Fatalf("Install under v3.2.0-rc.31: %v", err)
		}
		if got := readBytes(t, filepath.Join(f.home, relSkill)); string(got) != string(alphaV2) {
			t.Errorf("newer binary did not refresh the file: bytes %q, want the shipped %q", got, alphaV2)
		}
		m, err := Load(ManifestPath(f.home))
		if err != nil {
			t.Fatalf("load manifest: %v", err)
		}
		if fe := m.Files[manifestKey]; fe.MoaiVersion != "v3.2.0-rc.31" || fe.SHA256 != sha256Hex(alphaV2) {
			t.Errorf("newer binary left the record stale: %+v, want the v3.2.0-rc.31 record of the shipped bytes", fe)
		}
	})
}
