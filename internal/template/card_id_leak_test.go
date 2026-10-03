// Package template — card-id leak class guard.
//
// TestTemplateNoInternalContentLeak flagged an internal SPEC id plus an ISO
// date but let a bare card id (tNNNN) through, so a card citation could ship
// to every user project unseen. These tests pin the card-id class: it fires on
// the card-id shape in every scanned template surface, and stays silent on
// token shapes that only look like one.
package template

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

const cardIDClassName = "C9-card-id"

// cardIDBaselineEntry pins one (file, card id literal) pair that predates the
// C9 class. The baseline is removal-only by convention: a new literal, or a
// listed literal in an unlisted file, is flagged, and an entry whose literal has
// left its file fails TestCardIDBaselineHasNoStaleEntries. It is per (file,
// literal), not per occurrence: a repeated citation of a listed id in its listed
// file is not seen, and nothing caps the length of the list.
// Measured on base 7109e0900: 19 pairs in 9 files. The mirrored rule files are
// byte-parity-enforced with their .claude/ source, so stripping a template copy
// alone would break mirror parity; cleaning a pair means editing both trees.
type cardIDBaselineEntry struct {
	File string // relative path under internal/template/templates/
	ID   string // literal card id expected in this file
}

var cardIDBaseline = []cardIDBaselineEntry{
	{".claude/hooks/moai/sync-phase-quality-gate.sh", "t1388"},
	{".claude/hooks/moai/sync-phase-quality-gate.sh", "t1389"},
	{".claude/hooks/moai/sync-phase-quality-gate.sh", "t1392"},
	{".claude/hooks/moai/sync-phase-quality-gate.sh", "t604"},
	{".claude/hooks/moai/sync-phase-quality-gate.sh", "t663"},
	{".claude/hooks/moai/sync-phase-quality-gate.sh", "t664"},
	{".claude/rules/moai/core/askuser-protocol-reference.md", "t1303"},
	{".claude/rules/moai/workflow/auto-semantics.md", "t1339"},
	{".claude/rules/moai/workflow/auto-semantics.md", "t1393"},
	{".claude/rules/moai/workflow/factory-dispatch-detail.md", "t133"},
	{".claude/rules/moai/workflow/factory-dispatch-detail.md", "t224"},
	{".claude/rules/moai/workflow/factory-dispatch.md", "t1330"},
	{".claude/rules/moai/workflow/session-handoff-format.md", "t1303"},
	{".claude/rules/moai/workflow/worktree-integration-ops.md", "t529"},
	{".claude/rules/moai/workflow/worktree-integration-ops.md", "t741"},
	{".claude/rules/moai/workflow/worktree-integration-ops.md", "t852"},
	{".claude/rules/moai/workflow/worktree-integration-ops.md", "t880"},
	{".claude/rules/moai/workflow/worktree-integration.md", "t1398"},
	{".claude/skills/moai/workflows/gtd.md", "t696"},
}

// isCardIDBaselined reports whether the (relPath, matched) pair is a baseline
// entry. The check is by literal path plus literal id; no regex, no line number.
func isCardIDBaselined(relPath, matched string) bool {
	for _, e := range cardIDBaseline {
		if e.File == relPath && e.ID == matched {
			return true
		}
	}
	return false
}

// TestCardIDBaselineHasNoStaleEntries keeps the baseline removal-only in the
// useful direction: an entry whose card id has left its file must be deleted, so
// a re-introduced citation of that id is flagged instead of silently excused.
func TestCardIDBaselineHasNoStaleEntries(t *testing.T) {
	t.Parallel()

	if len(cardIDBaseline) == 0 {
		t.Fatal("card-id baseline is empty: nothing is being checked")
	}
	idRe := regexp.MustCompile(`\bt[0-9]{3,4}\b`)
	for _, e := range cardIDBaseline {
		data, err := os.ReadFile(filepath.Join(templatesRoot, filepath.FromSlash(e.File)))
		if err != nil {
			t.Errorf("baseline entry %s %s: file unreadable: %v", e.File, e.ID, err)
			continue
		}
		found := false
		for _, m := range idRe.FindAllString(string(data), -1) {
			if m == e.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("stale baseline entry: %s no longer carries %s — delete the entry", e.File, e.ID)
		}
	}
}

func TestCardIDBaselineIsPerFileAndPerLiteral(t *testing.T) {
	t.Parallel()

	const baselinedFile, baselinedID = ".claude/skills/moai/workflows/gtd.md", "t696"
	if !isCardIDBaselined(baselinedFile, baselinedID) {
		t.Fatalf("positive control failed: %s %s is not in the baseline", baselinedFile, baselinedID)
	}
	text := "see " + baselinedID
	if got := collectLeakViolations(baselinedFile, baselinedFile, text, leakClasses); anyViolationHasClass(got, cardIDClassName) {
		t.Errorf("the baselined pair fired: %v", got)
	}
	other := ".claude/skills/moai/workflows/other.md"
	if got := collectLeakViolations(other, other, text, leakClasses); !anyViolationHasClass(got, cardIDClassName) {
		t.Errorf("the same literal in an unlisted file did not fire: %v", got)
	}
	if got := collectLeakViolations(baselinedFile, baselinedFile, "see t9999", leakClasses); !anyViolationHasClass(got, cardIDClassName) {
		t.Errorf("a second, unlisted id in a baselined file did not fire: %v", got)
	}
}

func TestCardIDLeakClassDetectsShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		path string
		text string
	}{
		{"card citation in an agent body", ".claude/agents/moai/example.md", "observed RED on card t1392 in a lane"},
		{"bare id in a rule", ".claude/rules/moai/workflow/example.md", "the same shape t1388 had already worked around"},
		{"bare id in a skill body", ".claude/skills/moai/workflows/example.md", "(t696):"},
		{"bare id in a hook comment", ".claude/hooks/moai/example.sh", "# exits silently (card t604)."},
		{"three-digit id", ".claude/agents/moai/example.md", "reproduced in t852 first-hand"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := collectLeakViolations(tc.path, tc.path, tc.text, leakClasses)
			if !anyViolationHasClass(got, cardIDClassName) {
				t.Errorf("class %q did not fire for %q at %q; violations: %v", cardIDClassName, tc.text, tc.path, got)
			}
		})
	}
}

func TestCardIDLeakClassIgnoresLookalikes(t *testing.T) {
	t.Parallel()

	path := ".claude/agents/moai/example.md"
	for _, text := range []string{
		"the wait is 5t0 steps",
		"sort the t0 and t1 buffers",
		"a hash t12345 is longer than a card id",
		"prefix_t1392 is an identifier, not a card",
		"at1392 is a word fragment",
		"the format is tNNNN",
	} {
		got := collectLeakViolations(path, path, text, leakClasses)
		if anyViolationHasClass(got, cardIDClassName) {
			t.Errorf("class %q fired on a lookalike %q: %v", cardIDClassName, text, got)
		}
	}
}
