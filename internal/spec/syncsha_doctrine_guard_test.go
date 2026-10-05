package spec

// syncsha_doctrine_guard_test.go — the answer to "what catches it when the
// convention documents and the landed slot rule drift apart?" (card t398).
//
// THE DRIFT THIS EXISTS FOR. `syncsha.go` fixes the value grammar for
// `sync_commit_sha`: a commit SHA, or a member of the `pending-backfill`
// placeholder family. The documents that TELL an author what to write into that
// slot live outside Go — under `.claude/` — so nothing connected the two, and
// the two drifted: a convention document described the sync commit as
// "populating `sync_commit_sha`" without saying what a commit that cannot know
// its own hash is supposed to put there, and an actor reading it wrote an empty
// slot. An empty value is neither a SHA nor a placeholder, so it is exactly the
// class the slot-format rule reports on.
//
// WHAT THIS GUARD ASSERTS, in both directions:
//
//   - Every placeholder spelling the doctrine PRESCRIBES is accepted by the
//     predicate that implements the grammar. Narrow the grammar and the
//     documents stop being true; this reddens.
//   - The empty value is NOT accepted. The doctrine's "never leave it empty"
//     clause rests on that; if the predicate ever admitted it, the stated
//     rationale would be false while reading as if it still held.
//
// ANTI-VACUITY. A guard that scans a file set and finds nothing passes for the
// same reason a healthy tree passes. Both floors below are measured, not
// guessed: each of the four documents is asserted to be READ and to carry at
// least one prescribed spelling, so deleting the guidance — or losing a mirror
// subtree — reddens rather than quietly emptying the scan.
//
// The four documents are the root copies and their `internal/template/`
// mirrors. The mirror is what `go:embed` ships, so a root-only repair would
// leave every user project reading the stale convention.

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// syncSHADoctrineFiles are the documents that prescribe what goes into the
// slot. Explicit rather than globbed: adding one is a deliberate change a
// reviewer can see.
var syncSHADoctrineFiles = []string{
	".claude/rules/moai/workflow/spec-workflow.md",
	"internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md",
	".claude/agents/moai/manager-docs.md",
	"internal/template/templates/.claude/agents/moai/manager-docs.md",
}

// prescribedPlaceholderPattern captures a backtick-quoted placeholder spelling
// as the doctrine writes it. Backticks are required so prose ABOUT the family
// ("the pending-backfill family") is not mistaken for a prescription.
var prescribedPlaceholderPattern = regexp.MustCompile("`(pending-backfill[A-Za-z0-9-]*)`")

// TestSyncSHADoctrineMatchesPredicate is the doc-to-code coupling.
//
// Sentinel on failure: SYNCSHA_DOCTRINE_DRIFT
func TestSyncSHADoctrineMatchesPredicate(t *testing.T) {
	root := repoRoot(t)

	totalPrescribed := 0
	for _, rel := range syncSHADoctrineFiles {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("SYNCSHA_DOCTRINE_DRIFT: cannot read doctrine file %s: %v "+
				"(a document that prescribes the slot value must exist and be readable; "+
				"if it moved, update syncSHADoctrineFiles deliberately)", rel, err)
		}

		matches := prescribedPlaceholderPattern.FindAllStringSubmatch(string(data), -1)
		if len(matches) == 0 {
			t.Errorf("SYNCSHA_DOCTRINE_DRIFT: %s prescribes no `pending-backfill` spelling. "+
				"Either the slot-value guidance was removed — leaving authors to guess, which is how "+
				"the empty-slot drift happened — or this guard is scanning a file that no longer carries it.", rel)
			continue
		}

		for _, m := range matches {
			token := m[1]
			if !isSyncSHAPlaceholder(token) {
				t.Errorf("SYNCSHA_DOCTRINE_DRIFT: %s prescribes %q, which syncSHAPlaceholderPattern rejects. "+
					"The grammar and the document that teaches it have diverged: either the grammar was narrowed "+
					"without updating the doctrine, or the doctrine names a spelling that was never admitted.", rel, token)
			}
			totalPrescribed++
		}
	}

	// Measured floor: 2 spellings in each spec-workflow copy + 1 in each
	// manager-docs copy. Stated as a floor so ADDING guidance never reddens the
	// guard, while losing it does.
	const minPrescribed = 6
	if totalPrescribed < minPrescribed {
		t.Errorf("SYNCSHA_DOCTRINE_DRIFT: scanned %d prescribed spellings across %d documents, expected at least %d. "+
			"A shortfall means guidance was deleted or a mirror subtree went missing — not that the corpus got tidier.",
			totalPrescribed, len(syncSHADoctrineFiles), minPrescribed)
	}
}

// TestSyncSHAEmptySlotIsNotAPlaceholder pins the premise the doctrine's
// "never leave it empty" clause rests on.
//
// This is deliberately separate from the predicate's own unit tests: those
// assert the grammar for the grammar's sake, this one asserts the grammar still
// supports a sentence written in a document that cannot import it.
//
// Sentinel on failure: SYNCSHA_DOCTRINE_DRIFT
func TestSyncSHAEmptySlotIsNotAPlaceholder(t *testing.T) {
	for _, empty := range []string{"", " ", `""`, "''"} {
		token := syncSHAValueToken(empty)
		if isCommitSHAToken(token) || isSyncSHAPlaceholder(token) {
			t.Errorf("SYNCSHA_DOCTRINE_DRIFT: the empty slot %q normalizes to %q, which the predicates now accept. "+
				"The convention documents state that an empty slot is neither a SHA nor a recognized placeholder and is "+
				"therefore reported; that sentence is now false and must be rewritten with the grammar change, not left standing.",
				empty, token)
		}
	}
}
