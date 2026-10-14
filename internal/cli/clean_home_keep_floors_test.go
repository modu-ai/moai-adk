package cli

// SPEC-CAPACITY-RECLAIM-001 AC-CAP-010: the home-cleanup keep floors.
// Backups keep the three newest removed-* directories and select the rest
// only when they are also older than the retention window (the A12 floor:
// removal requires BOTH "beyond the newest three" AND "older than the
// window"). Releases keep the current release plus the releaseKeep newest
// non-current ones (A10 — unchanged by this SPEC; pinned here as the same
// observation's control arm).

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// seedHomeBackup writes one backups/removed-* directory aged ageDays behind
// now, with one ordinary payload file inside (no carved-out names).
func seedHomeBackup(t *testing.T, root string, now time.Time, name string, ageDays int) {
	t.Helper()
	dir := filepath.Join(root, "backups", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir backup %s: %v", name, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("backup payload"), 0o644); err != nil {
		t.Fatalf("seed backup %s: %v", name, err)
	}
	stamp := now.AddDate(0, 0, -ageDays)
	if err := os.Chtimes(dir, stamp, stamp); err != nil {
		t.Fatalf("backdate backup %s: %v", name, err)
	}
}

// backupCandidates filters a scan result to the backups category and returns
// the base names, sorted.
func backupCandidates(t *testing.T, candidates []homeCleanCandidate) []string {
	t.Helper()
	var names []string
	for _, c := range candidates {
		if c.Category != "backups" {
			continue
		}
		names = append(names, filepath.Base(c.RelPath))
	}
	if len(names) == 0 {
		return nil
	}
	for i := 1; i < len(names); i++ {
		for j := i; j > 0 && names[j] < names[j-1]; j-- {
			names[j], names[j-1] = names[j-1], names[j]
		}
	}
	return names
}

func releaseCandidateCount(candidates []homeCleanCandidate) int {
	n := 0
	for _, c := range candidates {
		if c.Category == "releases" {
			n++
		}
	}
	return n
}

// TestHomeCleanKeepFloors covers AC-CAP-010's two fixtures.
func TestHomeCleanKeepFloors(t *testing.T) {
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)

	t.Run("all-aged backups keep the three newest; releases keep current plus keep-newest", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		// Five backups, all older than the 30-day window, distinct ages so
		// "the three newest" is deterministic.
		seedHomeBackup(t, root, now, "removed-old-a", 41)
		seedHomeBackup(t, root, now, "removed-old-b", 40)
		seedHomeBackup(t, root, now, "removed-old-c", 39)
		seedHomeBackup(t, root, now, "removed-old-d", 38)
		seedHomeBackup(t, root, now, "removed-old-e", 37)
		// Six release binaries: the current version plus five older
		// non-current ones.
		releases := filepath.Join(root, defs.ReleasesSubdir)
		if err := os.MkdirAll(releases, 0o755); err != nil {
			t.Fatalf("mkdir releases: %v", err)
		}
		for _, name := range []string{"moai-v9.9.9", "moai-v1.0.0", "moai-v1.1.0", "moai-v1.2.0", "moai-v1.3.0", "moai-v1.4.0"} {
			if err := os.WriteFile(filepath.Join(releases, name), []byte("binary"), 0o755); err != nil {
				t.Fatalf("seed release %s: %v", name, err)
			}
		}

		candidates := scanHomeCleanable(root, config.DefaultHomeCleanRetentionDays, config.DefaultReleaseKeep, "v9.9.9", now)

		got := backupCandidates(t, candidates)
		want := []string{"removed-old-a", "removed-old-b"} // the two oldest; the three newest stay
		if len(got) != len(want) {
			t.Fatalf("backup candidates = %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("backup candidates = %v, want %v", got, want)
			}
		}

		// Releases: exactly the two oldest non-current binaries (five minus
		// the keep-newest three); the current release is never a candidate.
		if gotReleases := releaseCandidateCount(candidates); gotReleases != 2 {
			var names []string
			for _, c := range candidates {
				if c.Category == "releases" {
					names = append(names, filepath.Base(c.RelPath))
				}
			}
			t.Fatalf("release candidates = %d (%v), want 2 (the two oldest non-current)", gotReleases, names)
		}
		for _, c := range candidates {
			if c.Category == "releases" && filepath.Base(c.RelPath) == "moai-v9.9.9" {
				t.Fatalf("the current release became a deletion candidate: %+v", c)
			}
		}
	})

	t.Run("mixed-age backups keep the newest three regardless of window", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		// Five backups: two younger than the window, three older. Sorted by
		// recency the newest three are young-1, young-2, old-1 — old-1 is
		// older than the window yet inside the newest-three keep, so exactly
		// the remaining two older-than-window dirs are candidates.
		seedHomeBackup(t, root, now, "removed-young-1", 5)
		seedHomeBackup(t, root, now, "removed-young-2", 10)
		seedHomeBackup(t, root, now, "removed-old-1", 31)
		seedHomeBackup(t, root, now, "removed-old-2", 45)
		seedHomeBackup(t, root, now, "removed-old-3", 60)

		candidates := scanHomeCleanable(root, config.DefaultHomeCleanRetentionDays, config.DefaultReleaseKeep, "v9.9.9", now)

		got := backupCandidates(t, candidates)
		want := []string{"removed-old-2", "removed-old-3"}
		if len(got) != len(want) {
			t.Fatalf("backup candidates = %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("backup candidates = %v, want %v", got, want)
			}
		}
	})
}
