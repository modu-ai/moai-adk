package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// This file implements the seventeen-scenario oracle for IsBranchMerged
// (SPEC-WORKTREE-SQUASH-MERGE-001 acceptance.md §C) plus the AC-WSM-007
// probe-exit table and the AC-WSM-008 synthetic-commit determinism test.
//
// Every scenario builds a REAL isolated repository (never a mock) using the
// package helpers initTestRepo / runGit / writeTestFile plus the construction
// helpers below. clean_stale_test.go stubs IsBranchMerged at the CLI layer
// (mockIsBranchMergedFunc), so the CLI layer never exercises the real
// predicate — these tests are the only place the predicate's own behavior is
// verified against the matrix.
//
// Construction dates are pinned to a monotonic counter (commitDate) so commits
// created within the same second do not collide on SHA. This is load-bearing
// rather than cosmetic: a SHA collision moves the merge-base and silently
// corrupts the matrix (acceptance.md §B, SC-9 note).

// commitStep is a monotonic per-builder counter feeding pinned commit dates so
// two commits created within the same second cannot collide on SHA.
type commitStep struct {
	t    *testing.T
	dir  string
	step int
}

func newCommitStep(t *testing.T, dir string) *commitStep {
	t.Helper()
	return &commitStep{t: t, dir: dir}
}

// dateEnv returns the GIT_AUTHOR_DATE / GIT_COMMITTER_DATE pair for the next
// commit, advancing the counter. Dates use the "@<epoch> +0000" form.
func (c *commitStep) dateEnv() []string {
	c.step++
	d := fmt.Sprintf("@%d +0000", 1700000000+c.step)
	return []string{
		"GIT_AUTHOR_DATE=" + d,
		"GIT_COMMITTER_DATE=" + d,
	}
}

func (c *commitStep) commit(msg string) {
	c.t.Helper()
	runGitEnv(c.t, c.dir, c.dateEnv(), "commit", "-m", msg)
}

func (c *commitStep) commitAllowEmpty(msg string) {
	c.t.Helper()
	runGitEnv(c.t, c.dir, c.dateEnv(), "commit", "--allow-empty", "-m", msg)
}

// runGitEnv runs git with an extra environment overlay and fails on a non-zero
// exit. Used for construction commits that pin identity dates.
func runGitEnv(t *testing.T, dir string, env []string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	e := append(append([]string{}, os.Environ()...), "GIT_TERMINAL_PROMPT=0", "LC_ALL=C")
	e = append(e, env...)
	cmd.Env = e
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %s: %v", args, dir, string(out), err)
	}
	return strings.TrimSpace(string(out))
}

// runGitLiteral runs git with --literal-pathspecs as a git-level option, so a
// path whose first byte is ':' is staged/removed as a literal filename rather
// than parsed as pathspec magic. Used to build SC-13 without subjecting the
// construction itself to the defect under test (acceptance.md §B SC-13 note).
func runGitLiteral(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return runGit(t, dir, append([]string{"--literal-pathspecs"}, args...)...)
}

// submoduleAllowFileProtocolEnv returns the env overlay that allows the file
// transport for `git submodule add` of a local path on modern git. Set as a
// per-call env (the documented env-equivalent of `git -c`) rather than via
// `git config`, because protocol.file.allow is a protected key git ignores
// from repo-local config (acceptance.md §B SC-15 note).
func submoduleAllowFileProtocolEnv() []string {
	return []string{
		"GIT_CONFIG_COUNT=1",
		"GIT_CONFIG_KEY_0=protocol.file.allow",
		"GIT_CONFIG_VALUE_0=always",
	}
}

// checkout switches the repo's branch (or detached ref) and fails on error.
func checkout(t *testing.T, dir, ref string) {
	t.Helper()
	runGit(t, dir, "checkout", "-q", ref)
}

// assertMerged builds the scenario repo and asserts IsBranchMerged("feat","main")
// returns the expected verdict with no error.
func assertMerged(t *testing.T, label string, build func(*testing.T) string, want bool) {
	t.Helper()
	dir := build(t)
	wm := NewWorktreeManager(dir)
	got, err := wm.IsBranchMerged("feat", "main")
	if err != nil {
		t.Fatalf("%s: IsBranchMerged(feat,main) returned unexpected error: %v", label, err)
	}
	if got != want {
		t.Errorf("%s: IsBranchMerged(feat,main) = %v, want %v", label, got, want)
	}
}

// ===== Scenario builders (acceptance.md §B recipes, one per matrix row) =====

// SC-1 squash-merged: feat adds x.txt then y.txt in two commits; main adds both
// in one squash commit. Required verdict: merged (S4 + S5).
func buildSC1Squash(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	writeTestFile(t, filepath.Join(dir, "x.txt"), "x\n")
	runGit(t, dir, "add", "x.txt")
	c.commit("add x")
	writeTestFile(t, filepath.Join(dir, "y.txt"), "y\n")
	runGit(t, dir, "add", "y.txt")
	c.commit("add y")
	checkout(t, dir, "main")
	runGit(t, dir, "merge", "--squash", "feat")
	c.commit("squash feat")
	return dir
}

// SC-2 rebase-merged: feat adds p.txt then q.txt; main drifts then cherry-picks
// both commits individually. Required verdict: merged (S3 + S5).
func buildSC2Rebase(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	writeTestFile(t, filepath.Join(dir, "p.txt"), "p\n")
	runGit(t, dir, "add", "p.txt")
	pCommit := commitHash(t, dir, func() { c.commit("add p") })
	writeTestFile(t, filepath.Join(dir, "q.txt"), "q\n")
	runGit(t, dir, "add", "q.txt")
	qCommit := commitHash(t, dir, func() { c.commit("add q") })
	checkout(t, dir, "main")
	// main drifts independently before replaying.
	writeTestFile(t, filepath.Join(dir, "drift.txt"), "d\n")
	runGit(t, dir, "add", "drift.txt")
	c.commit("main drift")
	runGitEnv(t, dir, c.dateEnv(), "cherry-pick", pCommit)
	runGitEnv(t, dir, c.dateEnv(), "cherry-pick", qCommit)
	return dir
}

// SC-3 true merge commit: feat adds m.txt; main runs git merge --no-ff feat.
// Required verdict: merged (S1 + S2).
func buildSC3MergeCommit(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	writeTestFile(t, filepath.Join(dir, "m.txt"), "m\n")
	runGit(t, dir, "add", "m.txt")
	c.commit("add m")
	checkout(t, dir, "main")
	runGitEnv(t, dir, c.dateEnv(), "merge", "--no-ff", "feat", "-m", "merge feat")
	return dir
}

// SC-4 strictly behind: feat created at base and left alone; main advances.
// Required verdict: merged (S1 + S2).
func buildSC4Behind(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	checkout(t, dir, "main")
	writeTestFile(t, filepath.Join(dir, "ahead.txt"), "a\n")
	runGit(t, dir, "add", "ahead.txt")
	c.commit("main advances")
	return dir
}

// SC-5 empty-diff branch: feat carries one --allow-empty commit; main advances.
// Required verdict: merged (S2).
func buildSC5EmptyDiff(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	c.commitAllowEmpty("empty commit on feat")
	checkout(t, dir, "main")
	writeTestFile(t, filepath.Join(dir, "ahead.txt"), "a\n")
	runGit(t, dir, "add", "ahead.txt")
	c.commit("main advances")
	return dir
}

// SC-6 partially applied: feat adds a.txt then b.txt; main adds only a.txt.
// Required verdict: not merged (S5 withholds).
func buildSC6Partial(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	writeTestFile(t, filepath.Join(dir, "a.txt"), "a\n")
	runGit(t, dir, "add", "a.txt")
	c.commit("add a")
	writeTestFile(t, filepath.Join(dir, "b.txt"), "b\n")
	runGit(t, dir, "add", "b.txt")
	c.commit("add b")
	checkout(t, dir, "main")
	writeTestFile(t, filepath.Join(dir, "a.txt"), "a\n")
	runGit(t, dir, "add", "a.txt")
	c.commit("main adds only a")
	return dir
}

// SC-7 fully unmerged (control): feat adds z.txt; main drifts independently.
// Required verdict: not merged (no signal fires).
func buildSC7Unmerged(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	writeTestFile(t, filepath.Join(dir, "z.txt"), "z\n")
	runGit(t, dir, "add", "z.txt")
	c.commit("add z")
	checkout(t, dir, "main")
	writeTestFile(t, filepath.Join(dir, "drift.txt"), "d\n")
	runGit(t, dir, "add", "drift.txt")
	c.commit("main drift")
	return dir
}

// SC-8 partially reverted: feat adds a.txt then b.txt; main cherry-picks BOTH,
// then reverts the b commit. Required verdict: not merged (S5 withholds even
// though S3 fires on history alone).
func buildSC8Revert(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	writeTestFile(t, filepath.Join(dir, "a.txt"), "a\n")
	runGit(t, dir, "add", "a.txt")
	aCommit := commitHash(t, dir, func() { c.commit("add a") })
	writeTestFile(t, filepath.Join(dir, "b.txt"), "b\n")
	runGit(t, dir, "add", "b.txt")
	bCommit := commitHash(t, dir, func() { c.commit("add b") })
	checkout(t, dir, "main")
	runGitEnv(t, dir, c.dateEnv(), "cherry-pick", aCommit)
	runGitEnv(t, dir, c.dateEnv(), "cherry-pick", bCommit)
	// Revert the b commit (HEAD). git revert takes no -q (acceptance.md §B note).
	runGitEnv(t, dir, c.dateEnv(), "revert", "--no-edit", "HEAD")
	return dir
}

// SC-9 base is a strict superset: feat adds s.txt; main INDEPENDENTLY adds
// s.txt as its own commit (NOT a cherry-pick — that would collide on SHA and
// degenerate into SC-4), then additionally adds extra.txt.
// Required verdict: merged (S3 + S4 + S5).
func buildSC9Superset(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	writeTestFile(t, filepath.Join(dir, "s.txt"), "s\n")
	runGit(t, dir, "add", "s.txt")
	c.commit("feat adds s")
	checkout(t, dir, "main")
	// main re-creates the same change as its own commit (different SHA).
	writeTestFile(t, filepath.Join(dir, "s.txt"), "s\n")
	runGit(t, dir, "add", "s.txt")
	c.commit("main independently adds s")
	writeTestFile(t, filepath.Join(dir, "extra.txt"), "e\n")
	runGit(t, dir, "add", "extra.txt")
	c.commit("main adds extra")
	return dir
}

// SC-10 rename + re-add (squash): base holds a 10-line old.txt; feat runs
// `git mv old.txt new.txt` and commits, then appends a line to new.txt and
// commits — TWO commits; main squash-merges feat, then re-adds old.txt.
// Required verdict: not merged (S5 withholds; --no-renames keeps old.txt in
// the enumerated path list).
func buildSC10RenameReAddSquash(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	// Seed a 10-line old.txt at the merge-base (>= 50% similarity for rename
	// detection; acceptance.md §B note).
	var sb strings.Builder
	for i := 1; i <= 10; i++ {
		fmt.Fprintf(&sb, "line %d\n", i)
	}
	writeTestFile(t, filepath.Join(dir, "old.txt"), sb.String())
	runGit(t, dir, "add", "old.txt")
	c.commit("seed old.txt")
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	runGit(t, dir, "mv", "old.txt", "new.txt")
	c.commit("rename old -> new")
	// Append a line to new.txt in a SECOND commit (two-commit form is
	// load-bearing; acceptance.md §B SC-10 note).
	writeTestFile(t, filepath.Join(dir, "new.txt"), sb.String()+"appended line\n")
	runGit(t, dir, "add", "new.txt")
	c.commit("edit new.txt")
	checkout(t, dir, "main")
	runGit(t, dir, "merge", "--squash", "feat")
	c.commit("squash feat")
	// main re-adds old.txt with its original content.
	writeTestFile(t, filepath.Join(dir, "old.txt"), sb.String())
	runGit(t, dir, "add", "old.txt")
	c.commit("main re-adds old.txt")
	return dir
}

// SC-11 rename + re-add (cherry-pick): base holds 10-line old.txt; feat runs
// `git mv old.txt new.txt`; main cherry-picks that commit, then re-adds old.txt.
// Required verdict: not merged (S3 + S4 fire; S5 withholds via --no-renames).
func buildSC11RenameReAddCherryPick(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	var sb strings.Builder
	for i := 1; i <= 10; i++ {
		fmt.Fprintf(&sb, "line %d\n", i)
	}
	writeTestFile(t, filepath.Join(dir, "old.txt"), sb.String())
	runGit(t, dir, "add", "old.txt")
	c.commit("seed old.txt")
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	runGit(t, dir, "mv", "old.txt", "new.txt")
	renameCommit := commitHash(t, dir, func() { c.commit("rename old -> new") })
	checkout(t, dir, "main")
	runGitEnv(t, dir, c.dateEnv(), "cherry-pick", renameCommit)
	writeTestFile(t, filepath.Join(dir, "old.txt"), sb.String())
	runGit(t, dir, "add", "old.txt")
	c.commit("main re-adds old.txt")
	return dir
}

// SC-12 non-ASCII path removed from base: feat adds plain.txt and 문서.txt in
// one commit; main squash-merges, then removes 문서.txt. Required verdict: not
// merged (S5 withholds; -z keeps the non-ASCII path round-tripping verbatim).
func buildSC12NonASCII(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	// Set core.quotePath true explicitly (this machine's ~/.gitconfig overrides
	// it to false; acceptance.md §B SC-12 note).
	runGit(t, dir, "config", "core.quotePath", "true")
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	writeTestFile(t, filepath.Join(dir, "plain.txt"), "plain\n")
	writeTestFile(t, filepath.Join(dir, "문서.txt"), "doc\n")
	runGit(t, dir, "add", "plain.txt", "문서.txt")
	c.commit("feat adds plain + 문서")
	checkout(t, dir, "main")
	runGit(t, dir, "merge", "--squash", "feat")
	c.commit("squash feat")
	runGit(t, dir, "rm", "문서.txt")
	c.commit("main removes 문서")
	return dir
}

// SC-13 leading-colon path removed from base: feat adds plain.txt and :note.txt
// in one commit; main squash-merges, then removes :note.txt. Every add/rm of
// the hazard path runs under --literal-pathspecs so the construction is not
// subject to the defect under test (acceptance.md §B SC-13 note).
// Required verdict: not merged (S5 withholds; --literal-pathspecs matches the
// leading-colon path as a literal filename).
func buildSC13ColonPath(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	writeTestFile(t, filepath.Join(dir, "plain.txt"), "plain\n")
	writeTestFile(t, filepath.Join(dir, ":note.txt"), "note\n")
	runGitLiteral(t, dir, "add", "plain.txt", ":note.txt")
	c.commit("feat adds plain + :note")
	checkout(t, dir, "main")
	runGit(t, dir, "merge", "--squash", "feat")
	c.commit("squash feat")
	runGitLiteral(t, dir, "rm", ":note.txt")
	c.commit("main removes :note")
	return dir
}

// SC-14 textconv driver collapses a changed file: the seed adds .gitattributes
// binding *.pdf to a textconv driver whose script emits a constant; feat adds
// plain.txt + report.pdf; main squash-merges, then overwrites report.pdf with
// different content. .gitattributes stays in the worktree (the attribute is
// read from the checked-out file, not the tree under judgement; acceptance.md
// §B SC-14 note). Required verdict: not merged (S5 withholds; --no-textconv
// compares blobs, not the driver's rendering).
func buildSC14Textconv(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	// Executable textconv script at an absolute path (git requires an absolute
	// path or it reports "cannot run").
	textconvPath := filepath.Join(dir, "textconv.sh")
	writeTestFile(t, textconvPath, "#!/bin/sh\ncat >/dev/null\necho constant\n")
	if err := os.Chmod(textconvPath, 0o755); err != nil {
		t.Fatalf("chmod textconv: %v", err)
	}
	// Seed: .gitattributes + the diff driver. The attribute must remain in the
	// worktree at judgement time.
	writeTestFile(t, filepath.Join(dir, ".gitattributes"), "*.pdf diff=pdf\n")
	runGit(t, dir, "add", ".gitattributes")
	runGit(t, dir, "config", "diff.pdf.textconv", textconvPath)
	c.commit("seed .gitattributes + textconv driver")
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	writeTestFile(t, filepath.Join(dir, "plain.txt"), "plain\n")
	writeTestFile(t, filepath.Join(dir, "report.pdf"), "PDF-ORIGINAL\n")
	runGit(t, dir, "add", "plain.txt", "report.pdf")
	c.commit("feat adds plain + report.pdf")
	checkout(t, dir, "main")
	runGit(t, dir, "merge", "--squash", "feat")
	c.commit("squash feat")
	// Overwrite report.pdf with genuinely different content. Both blobs render
	// to "constant" under the driver, so an unpatched comparison reports equal.
	writeTestFile(t, filepath.Join(dir, "report.pdf"), "PDF-OVERWRITTEN-DIFFERENT\n")
	runGit(t, dir, "add", "report.pdf")
	c.commit("main overwrites report.pdf")
	return dir
}

// SC-15 submodule pointer moved under an ignore directive: a second throwaway
// repo is built with commits s0/s1/s2; the parent adds it as a submodule at
// `sub` pinned to s0; feat moves the pointer to s1 and adds plain.txt WITHOUT
// touching .gitmodules; main squash-merges, then moves the pointer to s2. The
// ignore directive (submodule.sub.ignore=all) is committed INSIDE .gitmodules,
// modeling an exposure a repository ships to every clone (acceptance.md §B
// SC-15 note). Required verdict: not merged (S5 withholds; --ignore-submodules=none
// on BOTH enumeration and comparison is required to see the moved gitlink).
func buildSC15Submodule(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)

	// Build the submodule repo fully (s0, s1, s2) BEFORE the parent adds it, so
	// `submodule add` fetches all three commits into the parent's
	// `.git/modules/sub` store and subsequent checkouts inside dir/sub resolve.
	subDir := resolveSymlinks(t, t.TempDir())
	runGit(t, subDir, "init", "-b", "main")
	runGit(t, subDir, "config", "user.email", "sub@e.x")
	runGit(t, subDir, "config", "user.name", "Sub")
	writeTestFile(t, filepath.Join(subDir, "sub_file.txt"), "s0\n")
	runGit(t, subDir, "add", "sub_file.txt")
	runGitEnv(t, subDir, nil, "commit", "-m", "s0")
	writeTestFile(t, filepath.Join(subDir, "sub_file.txt"), "s1\n")
	runGit(t, subDir, "add", "sub_file.txt")
	runGitEnv(t, subDir, nil, "commit", "-m", "s1")
	s1 := runGit(t, subDir, "rev-parse", "main")
	writeTestFile(t, filepath.Join(subDir, "sub_file.txt"), "s2\n")
	runGit(t, subDir, "add", "sub_file.txt")
	runGitEnv(t, subDir, nil, "commit", "-m", "s2")
	s2 := runGit(t, subDir, "rev-parse", "main")
	s0 := runGit(t, subDir, "rev-parse", "main~2")

	// Parent adds the submodule. protocol.file.allow=always is needed for a
	// local path on modern git. This pins HEAD (s2) and fetches all objects.
	runGitEnv(t, dir, submoduleAllowFileProtocolEnv(), "submodule", "add", subDir, "sub")
	// Move the parent's pointer from s2 back to s0 so feat can advance it to s1
	// and main to s2, modelling the scenario's divergence.
	runGit(t, filepath.Join(dir, "sub"), "checkout", s0)
	// Commit the ignore directive inside the TRACKED .gitmodules so it ships to
	// every clone with no operator config.
	gitmodulesPath := filepath.Join(dir, ".gitmodules")
	orig, err := os.ReadFile(gitmodulesPath)
	if err != nil {
		t.Fatalf("read .gitmodules: %v", err)
	}
	if !strings.Contains(string(orig), "ignore = all") {
		appendLine := "\tignore = all\n"
		if err := os.WriteFile(gitmodulesPath, append(orig, []byte(appendLine)...), 0o644); err != nil {
			t.Fatalf("write .gitmodules: %v", err)
		}
	}
	runGit(t, dir, "add", ".gitmodules", "sub")
	c.commit("add submodule at s0 with ignore=all")

	// feat moves the pointer to s1 and adds plain.txt (NO .gitmodules change).
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	runGit(t, filepath.Join(dir, "sub"), "checkout", s1)
	writeTestFile(t, filepath.Join(dir, "plain.txt"), "plain\n")
	runGit(t, dir, "add", "sub", "plain.txt")
	c.commit("feat moves submodule to s1 + adds plain")
	checkout(t, dir, "main")
	runGit(t, dir, "merge", "--squash", "feat")
	c.commit("squash feat")
	// main moves the pointer to s2.
	runGit(t, filepath.Join(dir, "sub"), "checkout", s2)
	runGit(t, dir, "add", "sub")
	c.commit("main moves submodule to s2")
	return dir
}

// P3c mode-only divergence: the seed adds x.txt; feat chmods it to 755 and adds
// plain.txt; main squash-merges, then chmods x.txt back to 644. Content is
// byte-identical on both sides throughout; only the mode differs. core.fileMode
// is left at its default (true). Required verdict: not merged (S5 withholds; git
// diff compares modes, so a mode-only change is detected).
func buildP3cModeOnly(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	writeTestFile(t, filepath.Join(dir, "x.txt"), "x\n")
	runGit(t, dir, "add", "x.txt")
	c.commit("seed x.txt")
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	if err := os.Chmod(filepath.Join(dir, "x.txt"), 0o755); err != nil {
		t.Fatalf("chmod 755: %v", err)
	}
	writeTestFile(t, filepath.Join(dir, "plain.txt"), "plain\n")
	runGit(t, dir, "add", "x.txt", "plain.txt")
	c.commit("feat chmods x + adds plain")
	checkout(t, dir, "main")
	runGit(t, dir, "merge", "--squash", "feat")
	c.commit("squash feat")
	if err := os.Chmod(filepath.Join(dir, "x.txt"), 0o644); err != nil {
		t.Fatalf("chmod 644: %v", err)
	}
	runGit(t, dir, "add", "x.txt")
	c.commit("main chmods x back")
	return dir
}

// P4c symlink/file blob-OID collision: the seed adds a symlink link.txt->target
// and a regular file target; feat replaces link.txt with a REGULAR file whose
// content is exactly "target" (NO trailing newline) and adds plain.txt; main
// squash-merges, then restores link.txt as a symlink. Both sides share one blob
// OID and differ only in mode. Required verdict: not merged (S5 withholds; git
// diff is mode-sensitive).
func buildP4cSymlinkFile(t *testing.T) string {
	dir := initTestRepo(t)
	c := newCommitStep(t, dir)
	// Seed: symlink + its target.
	targetPath := filepath.Join(dir, "target")
	writeTestFile(t, targetPath, "target-content\n")
	linkPath := filepath.Join(dir, "link.txt")
	if err := os.Symlink("target", linkPath); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	runGit(t, dir, "add", "target", "link.txt")
	c.commit("seed symlink + target")
	runGit(t, dir, "checkout", "-q", "-b", "feat")
	// Replace the symlink with a regular file whose content is EXACTLY the
	// symlink's target with no trailing newline (so both share one blob OID).
	if err := os.Remove(linkPath); err != nil {
		t.Fatalf("remove symlink: %v", err)
	}
	if err := os.WriteFile(linkPath, []byte("target"), 0o644); err != nil {
		t.Fatalf("write regular file: %v", err)
	}
	writeTestFile(t, filepath.Join(dir, "plain.txt"), "plain\n")
	runGit(t, dir, "add", "link.txt", "plain.txt")
	c.commit("feat replaces symlink + adds plain")
	checkout(t, dir, "main")
	runGit(t, dir, "merge", "--squash", "feat")
	c.commit("squash feat")
	// Restore link.txt as a symlink.
	if err := os.Remove(linkPath); err != nil {
		t.Fatalf("remove regular file: %v", err)
	}
	if err := os.Symlink("target", linkPath); err != nil {
		t.Fatalf("restore symlink: %v", err)
	}
	runGit(t, dir, "add", "link.txt")
	c.commit("main restores symlink")
	return dir
}

// commitHash performs a commit via the supplied do-commit callback and returns
// the resulting commit SHA. Used to capture cherry-pick targets before they are
// replayed.
func commitHash(t *testing.T, dir string, doCommit func()) string {
	t.Helper()
	doCommit()
	return runGit(t, dir, "rev-parse", "HEAD")
}

// ===== Top-level tests (names match the AC-WSM-001..017 -run selectors) =====
// The selector regex is matched as an unanchored substring of the fully
// qualified test name, so each top-level test is named to contain its
// criterion's keyword (Squash, Rebase, EmptyDiff, Reachab, Partial|Unmerged,
// Rename|NonASCII|ColonPath|Textconv|Submodule, ModeOnly|SymlinkFile, ProbeExit,
// Revert). acceptance.md §A non-vacuity discipline is judged by reading the
// `--- PASS:` lines, not the exit code.

func TestIsBranchMerged_SquashMerged(t *testing.T) {
	assertMerged(t, "SC-1 squash", buildSC1Squash, true)
}

func TestIsBranchMerged_RebaseMerged(t *testing.T) {
	assertMerged(t, "SC-2 rebase", buildSC2Rebase, true)
}

func TestIsBranchMerged_Reachability_TrueMergeCommit(t *testing.T) {
	assertMerged(t, "SC-3 true merge commit", buildSC3MergeCommit, true)
}

func TestIsBranchMerged_Reachability_StrictlyBehind(t *testing.T) {
	assertMerged(t, "SC-4 strictly behind", buildSC4Behind, true)
}

func TestIsBranchMerged_EmptyDiffBranch(t *testing.T) {
	assertMerged(t, "SC-5 empty-diff branch", buildSC5EmptyDiff, true)
}

func TestIsBranchMerged_PartiallyApplied(t *testing.T) {
	assertMerged(t, "SC-6 partially applied", buildSC6Partial, false)
}

func TestIsBranchMerged_UnmergedControl(t *testing.T) {
	assertMerged(t, "SC-7 fully unmerged", buildSC7Unmerged, false)
}

func TestIsBranchMerged_RevertRemovedFromBase(t *testing.T) {
	assertMerged(t, "SC-8 partially reverted", buildSC8Revert, false)
}

func TestIsBranchMerged_Superset(t *testing.T) {
	assertMerged(t, "SC-9 base is strict superset", buildSC9Superset, true)
}

func TestIsBranchMerged_RenameReAddSquash(t *testing.T) {
	assertMerged(t, "SC-10 rename+re-add (squash)", buildSC10RenameReAddSquash, false)
}

func TestIsBranchMerged_RenameReAddCherryPick(t *testing.T) {
	assertMerged(t, "SC-11 rename+re-add (cherry-pick)", buildSC11RenameReAddCherryPick, false)
}

func TestIsBranchMerged_NonASCIIPathRemoved(t *testing.T) {
	assertMerged(t, "SC-12 non-ASCII path removed", buildSC12NonASCII, false)
}

func TestIsBranchMerged_ColonPathRemovedFromBase(t *testing.T) {
	assertMerged(t, "SC-13 leading-colon path removed", buildSC13ColonPath, false)
}

func TestIsBranchMerged_TextconvDriverCollapsesChange(t *testing.T) {
	assertMerged(t, "SC-14 textconv driver", buildSC14Textconv, false)
}

func TestIsBranchMerged_SubmoduleIgnoreDirective(t *testing.T) {
	assertMerged(t, "SC-15 submodule ignore directive", buildSC15Submodule, false)
}

func TestIsBranchMerged_ModeOnlyDivergence(t *testing.T) {
	assertMerged(t, "P3c mode-only divergence", buildP3cModeOnly, false)
}

func TestIsBranchMerged_SymlinkFileOIDCollision(t *testing.T) {
	assertMerged(t, "P4c symlink/file OID collision", buildP4cSymlinkFile, false)
}

// TestIsBranchMerged_ProbeExit_Table binds AC-WSM-007: the per-probe exit-code
// table. A verdict-carrying exit (git diff --quiet rc 1, git merge-base rc 1 on
// unrelated histories) MUST NOT surface as an error; a true failure (unknown
// base ref) MUST surface as an error. Four subtests, one per row.
func TestIsBranchMerged_ProbeExit_Table(t *testing.T) {
	t.Run("UnknownBase", func(t *testing.T) {
		dir := buildSC1Squash(t)
		wm := NewWorktreeManager(dir)
		_, err := wm.IsBranchMerged("feat", "no-such-base")
		if err == nil {
			t.Fatal("expected error for unknown base ref; got nil")
		}
	})

	t.Run("UnrelatedHistories", func(t *testing.T) {
		dir := initTestRepo(t)
		// Create an orphan branch sharing no ancestor with main.
		runGit(t, dir, "checkout", "--orphan", "orphan")
		runGit(t, dir, "rm", "-rf", ".")
		writeTestFile(t, filepath.Join(dir, "orphan.txt"), "o\n")
		runGit(t, dir, "add", "orphan.txt")
		runGit(t, dir, "config", "user.email", "o@e.x")
		runGit(t, dir, "config", "user.name", "Orphan")
		runGit(t, dir, "commit", "-m", "orphan seed")
		runGit(t, dir, "checkout", "-q", "main")
		wm := NewWorktreeManager(dir)
		got, err := wm.IsBranchMerged("orphan", "main")
		if err != nil {
			t.Fatalf("unrelated histories must return (false, nil), got err: %v", err)
		}
		if got {
			t.Errorf("unrelated histories must return false, got true")
		}
	})

	t.Run("NonEmptyDiffNoError", func(t *testing.T) {
		// A branch with a non-empty diff vs merge-base exercises the
		// git diff --quiet rc=1 verdict path; it must not surface as an error.
		dir := buildSC7Unmerged(t)
		wm := NewWorktreeManager(dir)
		_, err := wm.IsBranchMerged("feat", "main")
		if err != nil {
			t.Fatalf("non-empty diff must not surface rc=1 as error: %v", err)
		}
	})

	t.Run("EmptyDiffMerged", func(t *testing.T) {
		dir := buildSC5EmptyDiff(t)
		wm := NewWorktreeManager(dir)
		got, err := wm.IsBranchMerged("feat", "main")
		if err != nil {
			t.Fatalf("empty-diff branch: %v", err)
		}
		if !got {
			t.Errorf("empty-diff branch must report merged, got false")
		}
	})
}
