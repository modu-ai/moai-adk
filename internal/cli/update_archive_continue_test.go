// SPEC-UPDATE-LEGACY-SKILL-LIST-001 / M4
// update_archive_continue_test.go — archiveLegacySkills must accumulate
// per-entry failures and keep going, rather than returning on the first one.
//
// Pre-fix behaviour: all three in-loop `return archived, fmt.Errorf(...)` sites
// returned before the post-loop `total:` emission, so a single failing entry
// both skipped every remaining entry and suppressed the summary line.

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestArchiveLegacySkills_ContinuesAfterFailure seeds two archivable skills and
// makes the FIRST one fail, so that "the loop continued" is distinguishable
// from "the failure happened to be the last entry".
//
// Failure injection is portable and needs no OS skip: the failing ID's archive
// destination is seeded as a regular FILE. archiveSkill's os.Stat on that path
// therefore succeeds, the drift check runs against a non-directory, and the
// per-entry call returns ARCHIVE_DRIFT. (plan.md §E M4 step 1 suggests reaching
// MkdirAll's ENOTDIR instead; that branch is unreachable this way, because Stat
// succeeds on a regular file and MkdirAll is only called when Stat fails. The
// property the fixture needs — a deterministic, cross-platform, per-entry
// failure with no chmod and no root caveat — holds either way.)
func TestArchiveLegacySkills_ContinuesAfterFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// Both IDs must be real legacySkillIDs entries: archiveLegacySkills only
	// walks that list. Indices 0 and 1 are in range at 13 entries.
	failID := legacySkillIDs[0]
	okID := legacySkillIDs[1]

	makeSkillDir(t, root, failID, "# "+failID+"\nfixture source.")
	makeSkillDir(t, root, okID, "# "+okID+"\nfixture source.")

	archiveRoot := filepath.Join(root, ".moai", "archive", "skills", archiveVersion)
	if err := os.MkdirAll(archiveRoot, 0o755); err != nil {
		t.Fatalf("seed archive root: %v", err)
	}
	// The failure injection: a regular file where a directory is expected.
	if err := os.WriteFile(filepath.Join(archiveRoot, failID), []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("seed failing destination: %v", err)
	}

	var out bytes.Buffer
	archived, err := archiveLegacySkills(root, &out, false)

	if err == nil {
		t.Fatalf("expected an aggregate error naming the failed entry, got nil")
	}
	if !strings.Contains(err.Error(), failID) {
		t.Errorf("aggregate error must name the failed entry %q, got: %v", failID, err)
	}

	// The load-bearing assertion: an entry AFTER the failure was still archived.
	laterArchive := filepath.Join(archiveRoot, okID, "SKILL.md")
	if _, statErr := os.Stat(laterArchive); statErr != nil {
		t.Errorf("the entry after the failing one must still be archived, but %s is absent: %v", laterArchive, statErr)
	}

	// The summary must survive a failing entry — pre-fix, the early return
	// skipped it entirely.
	if got := out.String(); !strings.Contains(got, "total:") {
		t.Errorf("the total: summary must be emitted even when an entry failed; output was:\n%s", got)
	}

	t.Run("success_count_excludes_failures", func(t *testing.T) {
		const wantArchived = 1 // okID succeeded; failID did not
		if archived != wantArchived {
			t.Errorf("archived = %d, want %d — the returned count reports successes only and must exclude the failed entry", archived, wantArchived)
		}
	})
}
