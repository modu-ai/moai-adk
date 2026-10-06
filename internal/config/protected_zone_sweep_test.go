package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestProtectedZoneSweep covers the dead-entry sweep helper: a paths entry that
// matches nothing is reported, a runtime_paths entry is exempt per entry, and the
// resolved and skipped counts are returned so an empty sweep cannot pass silently.
func TestProtectedZoneSweep(t *testing.T) {
	root := t.TempDir()
	writeZoneFile(t, root, "a/keep.txt", "x")
	writeZoneFile(t, root, "b/c_test.go", "x")
	if err := os.MkdirAll(filepath.Join(root, "emptydir"), 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := ParseProtectedZone([]byte(zoneShipped(
		"  mix:\n    paths: [\"a/\", \"**/*_test.go\", \"emptydir/\", \"gone/\", \"a/missing.txt\"]\n    runtime_paths: [\"never/shipped.json\"]\n")), true, ProtectedZoneShippedRel)
	if err != nil {
		t.Fatal(err)
	}
	got, err := SweepZoneEntries(ProtectedZone{Entries: entries}, ProtectedZoneShippedRel, os.DirFS(root))
	if err != nil {
		t.Fatal(err)
	}
	if got.Resolved != 3 || got.Skipped != 1 || len(got.Dead) != 2 {
		t.Fatalf("resolved=%d skipped=%d dead=%v, want 3/1/2 dead (gone/, a/missing.txt)", got.Resolved, got.Skipped, got.Dead)
	}
	// another source's entries are not swept
	got, err = SweepZoneEntries(ProtectedZone{Entries: entries}, ProtectedZoneOverlayRel, os.DirFS(root))
	if err != nil || got.Resolved != 0 || got.Skipped != 0 || len(got.Dead) != 0 {
		t.Fatalf("foreign source swept: %+v err=%v", got, err)
	}
}
