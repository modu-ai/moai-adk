// status_read_test.go — the branch-side status read
// (SPEC-KANBAN-BOARD-001 REQ-KB-020/024, M2; AC-KB-022/025).
//
// The fixture is the measured three-branch shape of SPEC-NAVIGATOR-SYNC-003
// (spec.md §A.4b): three branches matching the card by REQ-KW-003's
// exact-token rule, carrying draft / in-progress / completed, with the
// primary checkout carrying completed. All four observations run over that
// ONE branch set — the first two differ only in worktree liveness, which is
// what makes them a pair rather than two tests.
package kanban

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// specFixtureRepo builds the AC-KB-022 fixture: a primary repo whose card
// directory carries `completed` on the checked-out branch, plus three
// branches matching SPEC-NAV-X by the exact-token rule carrying draft /
// in-progress / completed. Returns the primary root.
func specFixtureRepo(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}
	mustGit(t, root, "init", "-b", "main")
	mustGit(t, root, "config", "user.email", "m2-test@example.com")
	mustGit(t, root, "config", "user.name", "M2 Test")
	// Every `git commit` otherwise spawns `git maintenance run --auto
	// --no-quiet --detach` — a DETACHED process that outlives the commit the
	// test waited on and keeps touching .git while t.TempDir's RemoveAll is
	// deleting it (measured 2026-08-24: 4 spawns per fixture repo, git 2.50.1;
	// 0 with these two settings). CI observed the resulting cleanup failure as
	// `unlinkat .../001/.git/objects: directory not empty`. This is not the
	// gc.auto loose-object threshold: `maintenance run --auto` detaches first
	// and decides afterwards, so the object count never gates the spawn.
	mustGit(t, root, "config", "gc.auto", "0")
	mustGit(t, root, "config", "maintenance.auto", "false")

	// Primary copy (on main): completed.
	writeSpecMD(t, root, "SPEC-NAV-X", StatusCompleted)
	mustGit(t, root, "add", ".")
	mustGit(t, root, "commit", "-m", "primary copy completed")

	// Three phase branches, each rewriting the card's spec.md status.
	// The branch names match SPEC-NAV-X by REQ-KW-003's rule: the segment
	// after the type prefix begins with the identifier, followed by a
	// hyphen or end-of-segment.
	for _, tc := range []struct{ branch, status string }{
		{"plan/SPEC-NAV-X", StatusDraft},
		{"feat/SPEC-NAV-X-run", StatusInProgress},
		{"sync/SPEC-NAV-X", StatusCompleted},
	} {
		mustGit(t, root, "branch", tc.branch)
		mustGit(t, root, "checkout", tc.branch)
		writeSpecMD(t, root, "SPEC-NAV-X", tc.status)
		mustGit(t, root, "add", ".")
		mustGit(t, root, "commit", "-m", tc.branch+" "+tc.status)
	}
	mustGit(t, root, "checkout", "main")
	return root
}

// fixtureBases returns a temp worktree base pair (Claude L1-style base under
// the fixture tree, plus an L2-style base) for worktree placement.
func fixtureBases(t *testing.T) WorktreeBases {
	t.Helper()
	base := t.TempDir()
	return WorktreeBases{
		Claude: filepath.Join(base, "claude-worktrees"),
		MoAI:   filepath.Join(base, "moai-worktrees"),
	}
}

// writeSpecMD writes a minimal spec.md with the given frontmatter status.
func writeSpecMD(t *testing.T, root, specID, status string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "specs", specID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir spec dir: %v", err)
	}
	body := "---\nid: " + specID + "\nstatus: " + status + "\n---\n\n# " + specID + "\n"
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}
}

// mustGit runs git in dir, failing the test on error.
func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %s: %v", args, dir, string(out), err)
	}
}

// TestReadCardStatus_LiveWorktreeSuppliesStatus — AC-KB-022 observation 1:
// with a LIVE worktree reporting the in-progress branch and the card in the
// run column, the board reads exactly in-progress — not completed from the
// more advanced branch, not draft from the less advanced one, not completed
// from the primary checkout.
func TestReadCardStatus_LiveWorktreeSuppliesStatus(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("fixture uses posix git plumbing; windows covered by GOOS=windows build")
	}
	root := specFixtureRepo(t)
	bases := fixtureBases(t)

	// The live worktree on the in-progress branch.
	wt := filepath.Join(bases.MoAI, "SPEC-NAV-X")
	mustGit(t, root, "worktree", "add", wt, "feat/SPEC-NAV-X-run")
	t.Cleanup(func() { mustGit(t, root, "worktree", "remove", "--force", wt) })

	got, err := ReadCardStatus(root, "SPEC-NAV-X", bases)
	if err != nil {
		t.Fatalf("ReadCardStatus: %v", err)
	}
	if got.Status != StatusInProgress {
		t.Fatalf("status = %q, want exactly %q — not the more advanced branch (%q), not the less advanced (%q), not the primary (%q)",
			got.Status, StatusInProgress, StatusCompleted, StatusDraft, StatusCompleted)
	}
	if got.Source != StatusSourceBranch {
		t.Fatalf("source = %q, want %q", got.Source, StatusSourceBranch)
	}
}

// TestReadCardStatus_NoLiveWorktreeReadsPrimary — AC-KB-022 observation 2:
// the IDENTICAL branch set with NO live worktree, card in done: the primary
// checkout supplies completed and none of the three branches does — a
// disposed card retains every branch, so keying on branch existence would
// read a stale phase branch and pair (done, draft), which the table refuses.
func TestReadCardStatus_NoLiveWorktreeReadsPrimary(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("fixture uses posix git plumbing; windows covered by GOOS=windows build")
	}
	root := specFixtureRepo(t)
	bases := fixtureBases(t)

	// No worktree anywhere: the branches all exist, none is observed.
	got, err := ReadCardStatus(root, "SPEC-NAV-X", bases)
	if err != nil {
		t.Fatalf("ReadCardStatus: %v", err)
	}
	if got.Status != StatusCompleted {
		t.Fatalf("status = %q, want %q from the primary checkout — none of the three retained branches supplies it", got.Status, StatusCompleted)
	}
	if got.Source != StatusSourcePrimary {
		t.Fatalf("source = %q, want %q", got.Source, StatusSourcePrimary)
	}
}

// TestReadCardStatus_NoBranchAtAllReadsPrimary — AC-KB-022 observation 3: a
// backlog/plan-shaped card with no branch of any kind reads the primary
// checkout and is judged by the same table.
func TestReadCardStatus_NoBranchAtAllReadsPrimary(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("fixture uses posix git plumbing; windows covered by GOOS=windows build")
	}
	root := specFixtureRepo(t)
	bases := fixtureBases(t)

	// A card that exists only as a primary-side directory: no branch ever
	// named it.
	writeSpecMD(t, root, "SPEC-NAV-BACKLOG", StatusDraft)

	got, err := ReadCardStatus(root, "SPEC-NAV-BACKLOG", bases)
	if err != nil {
		t.Fatalf("ReadCardStatus: %v", err)
	}
	if got.Status != StatusDraft || got.Source != StatusSourcePrimary {
		t.Fatalf("got (%q, %q), want (%q, %q)", got.Status, got.Source, StatusDraft, StatusSourcePrimary)
	}

	// And a card with no spec.md at all reports the no-file condition —
	// the backlog admission case, distinguishable from unknown.
	got, err = ReadCardStatus(root, "SPEC-NEVER-PLANNED", bases)
	if err != nil {
		t.Fatalf("ReadCardStatus(no spec.md): %v", err)
	}
	if got.SpecFilePresent {
		t.Fatal("SpecFilePresent = true for a card with no spec.md")
	}
}

// TestReadCardStatus_CommittedStateOnly — AC-KB-022 observation 4: a
// transition written into a worktree's working tree but NOT committed is
// invisible — the board observes what the repository records, so the branch
// still supplies its committed value.
func TestReadCardStatus_CommittedStateOnly(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("fixture uses posix git plumbing; windows covered by GOOS=windows build")
	}
	root := specFixtureRepo(t)
	bases := fixtureBases(t)

	wt := filepath.Join(bases.MoAI, "SPEC-NAV-X")
	mustGit(t, root, "worktree", "add", wt, "feat/SPEC-NAV-X-run")
	t.Cleanup(func() { mustGit(t, root, "worktree", "remove", "--force", wt) })

	// An uncommitted status edit in the worktree's working tree.
	writeSpecMD(t, wt, "SPEC-NAV-X", StatusImplemented)

	got, err := ReadCardStatus(root, "SPEC-NAV-X", bases)
	if err != nil {
		t.Fatalf("ReadCardStatus: %v", err)
	}
	if got.Status != StatusInProgress {
		t.Fatalf("status = %q, want the committed %q — a transition written but not committed is not yet a transition", got.Status, StatusInProgress)
	}
}

// TestReadCardStatus_DetachedWorktreeIsUnresolved — AC-KB-025 (REQ-KB-024):
// a card whose worktree exists but reports NO branch (detached HEAD) reads
// unresolved — not a member of the 8-value enum, no member substituted for
// the absence (the zero-value default reports draft and dispatches, which is
// the silent failure this exists to exclude).
func TestReadCardStatus_DetachedWorktreeIsUnresolved(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("fixture uses posix git plumbing; windows covered by GOOS=windows build")
	}
	root := specFixtureRepo(t)
	bases := fixtureBases(t)

	wt := filepath.Join(bases.Claude, "SPEC-NAV-DET")
	mustGit(t, root, "worktree", "add", "--detach", wt, "main")
	t.Cleanup(func() { mustGit(t, root, "worktree", "remove", "--force", wt) })

	got, err := ReadCardStatus(root, "SPEC-NAV-DET", bases)
	if err != nil {
		t.Fatalf("ReadCardStatus(detached): %v", err)
	}
	if got.Status != StatusUnresolved {
		t.Fatalf("status = %q, want %q — a worktree existing but reporting no branch leaves the source unresolved", got.Status, StatusUnresolved)
	}
	if IsCanonicalStatus(got.Status) {
		t.Fatalf("unresolved must not be a canonical status; got %q", got.Status)
	}
	if got.Source != StatusSourceUnresolved {
		t.Fatalf("source = %q, want %q", got.Source, StatusSourceUnresolved)
	}
}

// TestBranchNamesSpec_ExactTokenRule — REQ-KW-003's recognition rule, driven
// table-style: the segment after the type prefix begins with the identifier
// and the next character is either absent or a hyphen. A phase-suffixed
// branch is recognized; a branch merely embedding the identifier as a longer
// token is not.
func TestBranchNamesSpec_ExactTokenRule(t *testing.T) {
	t.Parallel()
	cases := []struct {
		branch string
		want   bool
	}{
		{"feat/SPEC-NAV-X", true},
		{"feat/SPEC-NAV-X-run", true},
		{"plan/SPEC-NAV-X", true},
		{"sync/SPEC-NAV-X-close", true},
		{"docs/SPEC-NAV-X-fork-resolution", true},
		{"SPEC-NAV-X", true}, // no type prefix: the whole name is the segment
		{"feat/SPEC-NAV-XY", false},
		{"feat/SPEC-NAV-XY-run", false},
		{"feat/aspec-nav-x-run", false},
		{"feat/prefix-SPEC-NAV-X", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := branchNamesSpec(tc.branch, "SPEC-NAV-X"); got != tc.want {
			t.Errorf("branchNamesSpec(%q, SPEC-NAV-X) = %v, want %v", tc.branch, got, tc.want)
		}
	}
	if branchNamesSpec("feat/SPEC-NAV-X", "") {
		t.Error("branchNamesSpec with empty specID must be false")
	}
}

// TestReadCardStatus_DoesNotSearchBranchSet — the branch set is never
// searched: a card whose worktree is absent does NOT consult the three
// retained branches (observed via the primary answer), and the observation
// route reads only what a live worktree reports. The companion guard: the
// implementation exposes no branch-search entry point.
func TestReadCardStatus_DoesNotSearchBranchSet(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("fixture uses posix git plumbing; windows covered by GOOS=windows build")
	}
	root := specFixtureRepo(t)
	bases := fixtureBases(t)

	// No worktree: the answer is the primary's, although feat/SPEC-NAV-X-run
	// carrying in-progress exists and is visible.
	got, err := ReadCardStatus(root, "SPEC-NAV-X", bases)
	if err != nil {
		t.Fatalf("ReadCardStatus: %v", err)
	}
	if got.Status != StatusCompleted {
		t.Fatalf("status = %q, want the primary %q — a search would have surfaced a branch instead", got.Status, StatusCompleted)
	}
	if strings.Contains(got.Source, "search") {
		t.Fatalf("source %q suggests a branch search", got.Source)
	}
}

// TestUnresolvedCard_OutcomeDistinctAndByteUnchanged — AC-KB-025's
// board-level assertions over the detached fixture: the reconciled view is
// UNRESOLVED (not inconsistent — both outcomes refuse dispatch, so the test
// observes WHICH outcome was reported), the card keeps its recorded column,
// no enum member is substituted, and both the board state file and the
// card's spec.md are byte-unchanged by the reconciliation.
func TestUnresolvedCard_OutcomeDistinctAndByteUnchanged(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("fixture uses posix git plumbing; windows covered by GOOS=windows build")
	}
	root := specFixtureRepo(t)
	bases := fixtureBases(t)
	seedLead(t, root, "lead-sess")

	// Board record holding the card in run; the card's primary-side spec.md
	// exists (completed on main).
	writeBoardRaw(t, BoardPath(root), &BoardState{Cards: []Card{
		{SpecID: "SPEC-NAV-DET2", Column: ColumnRun, Holder: "s1", LastMovedAt: "t0"},
	}})
	beforeBoard := readBoardBytes(t, root)
	specPath := filepath.Join(root, ".moai", "specs", "SPEC-NAV-DET2", "spec.md")
	writeSpecMD(t, root, "SPEC-NAV-DET2", StatusInProgress)
	beforeSpec := readFileBytes(t, specPath)

	// The card's worktree exists but reports no branch.
	wt := filepath.Join(bases.Claude, "SPEC-NAV-DET2")
	mustGit(t, root, "worktree", "add", "--detach", wt, "main")
	t.Cleanup(func() { mustGit(t, root, "worktree", "remove", "--force", wt) })

	cs, err := ReadCardStatus(root, "SPEC-NAV-DET2", bases)
	if err != nil {
		t.Fatalf("ReadCardStatus: %v", err)
	}
	card := Card{SpecID: "SPEC-NAV-DET2", Column: ColumnRun, Holder: "s1", LastMovedAt: "t0"}
	view := ReconcileCard(card, cs)

	if !view.Unresolved {
		t.Fatalf("view.Unresolved = false — the detached-worktree card must report unresolved")
	}
	if view.Inconsistent {
		t.Fatalf("view.Inconsistent = true — unresolved is NOT the illegal-pairing outcome; the test observes which outcome was reported")
	}
	if view.Dispatchable {
		t.Fatal("unresolved card dispatchable")
	}
	if view.Card.Column != ColumnRun {
		t.Fatalf("card column = %q, want the recorded run — the recorded column stands", view.Card.Column)
	}
	if IsCanonicalStatus(view.Status) || view.Status == StatusDraft {
		t.Fatalf("status = %q — an enum member was substituted for the absence (draft would pair legally and dispatch)", view.Status)
	}

	// Nothing was written by the read or the reconciliation.
	if string(beforeBoard) != string(readBoardBytes(t, root)) {
		t.Fatal("board state file changed")
	}
	if string(beforeSpec) != string(readFileBytes(t, specPath)) {
		t.Fatal("spec.md changed")
	}
}
