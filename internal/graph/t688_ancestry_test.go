package graph

import (
	"os"
	"path/filepath"
	"testing"
)

// AC-GSA-001 / AC-GSA-003 — a clean stamp whose commit object EXISTS in this
// checkout but is NOT an ancestor of HEAD is freshness-UNMEASURED, not stale
// and not fresh. Object presence and comparability are different conditions:
// `git cat-file -e` answers the first and says nothing about the second.
//
// The report carries the compatibility `VerdictAbsent` carrier plus a
// non-nil system error (exit 2 at the CLI boundary), and NONE of the
// measurement fields are populated — no value, no threshold comparison, no
// content anchor, no contribution, no driving paths (REQ-GSA-004).
//
// Mutant direction: an implementation that treats `git cat-file -e` success as
// comparability continues into the diff and returns a nil error here.
func TestCheckCodemaps_ExistingNonAncestorStampIsUnmeasured(t *testing.T) {
	root := newCheckFixture(t)
	baseBranch := gitFix(t, root, "symbolic-ref", "--short", "HEAD")
	gitFix(t, root, "switch", "-q", "-c", "stamp-side")
	if err := os.WriteFile(filepath.Join(root, "internal", "side.go"), []byte("package internal\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitFix(t, root, "add", "internal/side.go")
	gitFix(t, root, "commit", "-q", "-m", "side stamp")
	stamp := gitFix(t, root, "rev-parse", "HEAD")
	gitFix(t, root, "switch", "-q", baseBranch)
	writeCodemapsProvenance(t, root, stamp)

	if !objectExists(t, root, stamp) {
		t.Fatalf("fixture precondition failed: stamp %s must resolve as a commit object", stamp)
	}
	if isAncestor(t, root, stamp, "HEAD") {
		t.Fatalf("fixture precondition failed: stamp %s must NOT be an ancestor of HEAD", stamp)
	}

	rep, err := checkCodemaps(root, DefaultThresholds())
	if err == nil {
		t.Fatalf("existing non-ancestor stamp must be freshness-unmeasured with a system error; got verdict=%q value=%d threshold=%d content_anchor=%q",
			rep.Verdict, rep.Value, rep.Threshold, rep.ContentAnchor)
	}
	if rep.Verdict != VerdictAbsent {
		t.Fatalf("verdict = %q, want %q compatibility carrier", rep.Verdict, VerdictAbsent)
	}
	assertUnmeasured(t, rep)
	assertUnreachableReason(t, rep.Reason)
}

// AC-GSA-008 — the merge / squash / rebase-like lineage table. All three
// topologies retain the stamped commit as a reachable OBJECT; only the
// merge-commit lineage keeps it an ANCESTOR of HEAD. Squash and cherry-pick
// rewrite content onto a new commit and leave the original off HEAD's history,
// so both must land on the unmeasured path even though the object resolves.
//
// Mutant directions: substituting object existence for ancestry fails the two
// non-ancestor rows; rejecting every non-linear history fails the merge row.
func TestCheckCodemaps_MergeSquashRebaseLikeTopologies(t *testing.T) {
	t.Run("merge commit preserves ancestry", func(t *testing.T) {
		root := newCheckFixture(t)
		baseBranch := gitFix(t, root, "symbolic-ref", "--short", "HEAD")
		gitFix(t, root, "switch", "-q", "-c", "stamp-side")
		stamp := commitTopologyFile(t, root, "internal/stamp.go", "package internal\n", "stamp source")
		gitFix(t, root, "switch", "-q", baseBranch)
		commitTopologyFile(t, root, "internal/base.go", "package internal\n", "base advance")
		gitFix(t, root, "merge", "-q", "--no-ff", "stamp-side", "-m", "merge stamp side")
		assertTopologyDisposition(t, root, stamp, true)
	})

	t.Run("squash retains object but drops ancestry", func(t *testing.T) {
		root := newCheckFixture(t)
		baseBranch := gitFix(t, root, "symbolic-ref", "--short", "HEAD")
		gitFix(t, root, "switch", "-q", "-c", "stamp-side")
		stamp := commitTopologyFile(t, root, "internal/stamp.go", "package internal\n", "stamp source")
		gitFix(t, root, "switch", "-q", baseBranch)
		gitFix(t, root, "merge", "-q", "--squash", "stamp-side")
		gitFix(t, root, "commit", "-q", "-m", "squash stamp side")
		assertTopologyDisposition(t, root, stamp, false)
	})

	t.Run("rebase-like rewrite retains object but drops ancestry", func(t *testing.T) {
		root := newCheckFixture(t)
		baseBranch := gitFix(t, root, "symbolic-ref", "--short", "HEAD")
		gitFix(t, root, "switch", "-q", "-c", "stamp-side")
		stamp := commitTopologyFile(t, root, "internal/stamp.go", "package internal\n", "stamp source")
		gitFix(t, root, "switch", "-q", baseBranch)
		commitTopologyFile(t, root, "internal/base.go", "package internal\n", "base advance")
		gitFix(t, root, "cherry-pick", stamp)
		assertTopologyDisposition(t, root, stamp, false)
	})
}

// AC-GSA-005 boundary — the stamp being HEAD itself is ancestor-or-self, so the
// precheck must let it through to the normal freshness path. Mutant direction:
// a strict-ancestor test (excluding self) rejects this and turns every
// just-stamped tree into a system error.
func TestCheckCodemaps_StampAtHeadIsComparable(t *testing.T) {
	root := newCheckFixture(t)
	head := gitFix(t, root, "rev-parse", "HEAD")
	writeCodemapsProvenance(t, root, head)

	rep, err := checkCodemaps(root, DefaultThresholds())
	if err != nil {
		t.Fatalf("a stamp at HEAD is ancestor-or-self and must stay measurable, got system error: %v", err)
	}
	if rep.Verdict == VerdictAbsent {
		t.Fatalf("verdict = %q — a stamp at HEAD must not be treated as unreachable; reason=%s", rep.Verdict, rep.Reason)
	}
}

// C1 regression — a codemaps body that does not exist at all keeps its
// pre-existing determinate-observation disposition (absent + nil error, exit 1)
// even though the stamp is a non-ancestor. The ancestry precheck must not
// reclassify a body-absent tree as a failed measurement (spec.md §D row 1).
func TestCheckCodemaps_BodyAbsentWinsOverNonAncestorStamp(t *testing.T) {
	th := DefaultThresholds()
	root := newCheckFixture(t)

	baseBranch := gitFix(t, root, "symbolic-ref", "--short", "HEAD")
	gitFix(t, root, "switch", "-q", "-c", "stamp-side")
	head := commitDescribedChurn(t, root, th.CodemapsChangedFiles, "gen")
	gitFix(t, root, "switch", "-q", baseBranch)
	stampAt(t, root, head) // provenance.json only — no codemaps body

	if isAncestor(t, root, head, "HEAD") {
		t.Fatalf("fixture precondition failed: stamp %s must NOT be an ancestor of HEAD", head)
	}

	rep, err := checkCodemaps(root, th)
	if err != nil {
		t.Fatalf("body absence is a determinate observation and outranks the ancestry precheck; want nil error, got: %v", err)
	}
	if rep.Verdict != VerdictAbsent {
		t.Errorf("verdict = %q, want %q", rep.Verdict, VerdictAbsent)
	}
	assertUnmeasured(t, rep)
}

// AC-GSA-010 regression — an unresolvable stamp keeps its pre-existing
// not-comparable reason and system error. The ancestry precheck runs after
// object resolution, so a stamp git cannot resolve must NOT be relabelled as
// "unreachable" (spec.md §D rows 2 vs 3 are distinct dispositions).
func TestCheckCodemaps_UnresolvableStampKeepsNotComparableReason(t *testing.T) {
	root := newCheckFixture(t)
	writeCodemapsProvenance(t, root, "0000000000000000000000000000000000000000")

	rep, err := checkCodemaps(root, DefaultThresholds())
	if err == nil {
		t.Fatal("an unresolvable stamp must return a system error")
	}
	if rep.Verdict != VerdictAbsent {
		t.Errorf("verdict = %q, want %q", rep.Verdict, VerdictAbsent)
	}
	if rep.Reason != "stamped commit not comparable (unmeasured, system error follows)" {
		t.Errorf("reason = %q, want the pre-existing not-comparable reason preserved", rep.Reason)
	}
	assertUnmeasured(t, rep)
}

// --- helpers -------------------------------------------------------------

// objectExists reports whether the sha resolves as a commit object in root.
func objectExists(t *testing.T, root, sha string) bool {
	t.Helper()
	_, err := gitOutput(root, "cat-file", "-e", sha+"^{commit}")
	return err == nil
}

// assertUnmeasured pins REQ-GSA-004: an unmeasured report presents no
// measurement field as a result.
func assertUnmeasured(t *testing.T, rep LayerReport) {
	t.Helper()
	if rep.Value != 0 || rep.ContentAnchor != "" || rep.ContentAnchorSource != "" ||
		rep.Contribution != nil || len(rep.DrivingPaths) != 0 || rep.DrivingPathsOmitted != 0 {
		t.Fatalf("unmeasured path fabricated freshness fields: %+v", rep)
	}
}

// assertUnreachableReason pins REQ-GSA-003: the reason must distinguish an
// unreachable stamp from a missing body, and must say freshness went unmeasured.
func assertUnreachableReason(t *testing.T, reason string) {
	t.Helper()
	for _, token := range []string{"unreachable stamp", "freshness unmeasured"} {
		if !containsFold(reason, token) {
			t.Fatalf("reason = %q, want it to carry %q", reason, token)
		}
	}
}

func containsFold(haystack, needle string) bool {
	return len(needle) == 0 || indexFold(haystack, needle) >= 0
}

func indexFold(s, sub string) int {
	ls, lsub := len(s), len(sub)
	for i := 0; i+lsub <= ls; i++ {
		if equalFold(s[i:i+lsub], sub) {
			return i
		}
	}
	return -1
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// commitTopologyFile writes and commits one file, returning the new HEAD sha.
func commitTopologyFile(t *testing.T, root, path, content, message string) string {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	gitFix(t, root, "add", path)
	gitFix(t, root, "commit", "-q", "-m", message)
	return gitFix(t, root, "rev-parse", "HEAD")
}

// assertTopologyDisposition checks one row of the lineage table: the object is
// always present, ancestry is what varies, and the checker disposition follows
// ancestry rather than object presence.
func assertTopologyDisposition(t *testing.T, root, stamp string, wantAncestor bool) {
	t.Helper()
	if !objectExists(t, root, stamp) {
		t.Fatalf("every topology must retain the stamp object; %s does not resolve", stamp)
	}
	if got := isAncestor(t, root, stamp, "HEAD"); got != wantAncestor {
		t.Fatalf("stamp ancestor of HEAD = %v, want %v", got, wantAncestor)
	}
	writeCodemapsProvenance(t, root, stamp)
	rep, err := checkCodemaps(root, DefaultThresholds())
	if wantAncestor {
		if err != nil {
			t.Fatalf("merge topology must remain freshness-judgeable: %v", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("object-present non-ancestor stamp must be freshness-unmeasured; got verdict=%q value=%d anchor=%q",
			rep.Verdict, rep.Value, rep.ContentAnchor)
	}
	if rep.Verdict != VerdictAbsent {
		t.Fatalf("verdict = %q, want %q compatibility carrier", rep.Verdict, VerdictAbsent)
	}
	assertUnmeasured(t, rep)
	assertUnreachableReason(t, rep.Reason)
}
