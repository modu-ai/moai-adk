// prlink_landed_corpus_test.go — SPEC-TODO-LANDING-ATTRIBUTION-001 corpus
// reproduction contract (constraint D8): the predicate must reproduce card
// t482's exhaustive classification over the PINNED corpus commit 7835148d3 —
// of the ids appearing anywhere in a subject, exactly 310 attributed / 37
// unattributed (347 total). The reference transcription is
// .moai/reports/t482/forms.py; this test drives the GO predicate over the
// same subject stream and pins the partition.
//
// RE-PIN (SPEC-TODO-LAND-AUTO-DONE-001, 2026-09-13): form 2c (the
// comma-form trailing parenthetical, AC-AD-017) reclassified exactly one
// corpus id. t68's subject `merge: Factory Mode -f N worker fan-out (t68,
// SPEC-FACTORY-WORKER-FANOUT-001)` is the comma-form shape the SPEC makes
// first-class — the group opens with the card id and continues with a
// non-card qualifier. The pre-form-2c partition measured 309/38 with t68 in
// the residual; the negation guard (REQ-AD-008) moved no corpus id. A reader
// who needs the pre-form-2c figures re-runs at the pre-amendment predicate
// rather than re-citing this note.
//
// The pin is to a COMMIT, not a branch name (VCI §2.1 remedy R1). The test
// skips when the pinned commit is absent (a shallow CI clone), and the
// pinned numbers are dated references measured at branch WT-landed-drift-detect
// HEAD c43c07c3d — a reader who needs them re-runs rather than re-cites.
package kanban

import (
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// pinnedCorpusCommit is the corpus pin every SPEC figure is measured against.
const pinnedCorpusCommit = "7835148d3"

// pinnedResidual is the 38-id residual (ids present in some subject but
// attributed by no form), transcribed from .moai/reports/t482/forms.py's
// output over the pinned corpus. Pinning the SET, not just the counts, is
// what makes the partition checkable rather than merely plausible.
var pinnedResidual = []string{
	"t2", "t21", "t40", "t46", "t73", "t74", "t80", "t94",
	"t121", "t123", "t124", "t128", "t129", "t131", "t132", "t133", "t134",
	"t135", "t137", "t139", "t141", "t142", "t143", "t144", "t147", "t148",
	"t149", "t155", "t157", "t158", "t216", "t225", "t250", "t311", "t409",
	"t443", "t460",
}

func TestLandedPredicate_PinnedCorpusReproduction(t *testing.T) {
	// Locate the repository and check the pin is reachable; skip honestly on
	// a shallow clone rather than passing on an empty sweep.
	top, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Skipf("not inside a git work tree: %v", err)
	}
	root := strings.TrimSpace(string(top))
	if err := exec.Command("git", "-C", root, "cat-file", "-e", pinnedCorpusCommit+"^{commit}").Run(); err != nil {
		t.Skipf("pinned corpus commit %s not reachable (shallow clone?): %v", pinnedCorpusCommit, err)
	}
	out, err := exec.Command("git", "-C", root, "log", pinnedCorpusCommit, "--format=%s").Output()
	if err != nil {
		t.Fatalf("git log %s: %v", pinnedCorpusCommit, err)
	}
	subjects := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if len(subjects) < 5000 {
		t.Fatalf("subject stream unexpectedly small: %d subjects — the pin resolved to the wrong history", len(subjects))
	}

	// The resolved landed ref is origin/develop; the branch is DERIVED from
	// it (REQ-TLA-013), exercising the same derivation the querier performs.
	branch := landedBranchFromRef("origin/develop")
	if branch != "develop" {
		t.Fatalf("derived branch = %q, want develop", branch)
	}

	mentioned := map[string]bool{}
	attributed := map[string]bool{}
	for _, s := range subjects {
		for _, tok := range subjectCardToken.FindAllString(s, -1) {
			mentioned[tok] = true
		}
		if card := subjectAttribution(s, branch); card != "" {
			attributed[card] = true
		}
	}

	residual := make([]string, 0, len(mentioned))
	for id := range mentioned {
		if !attributed[id] {
			residual = append(residual, id)
		}
	}
	sort.Slice(residual, func(i, j int) bool {
		a, _ := strconv.Atoi(residual[i][1:])
		b, _ := strconv.Atoi(residual[j][1:])
		return a < b
	})

	if len(mentioned) != 347 {
		t.Errorf("ids appearing anywhere in a subject = %d, want 347", len(mentioned))
	}
	if len(attributed) != 310 {
		t.Errorf("ids attributed = %d, want 310", len(attributed))
	}
	if len(residual) != 37 {
		t.Errorf("subject-present but unattributed = %d, want 37", len(residual))
	}
	if strings.Join(residual, " ") != strings.Join(pinnedResidual, " ") {
		t.Errorf("residual set diverged:\n got: %s\nwant: %s", strings.Join(residual, " "), strings.Join(pinnedResidual, " "))
	}
}
