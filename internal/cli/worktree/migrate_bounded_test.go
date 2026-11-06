package worktree

// migrate_bounded_test.go — card t2053: the worktree migration must neither
// hash ignored bulk content, nor wait unboundedly on a subprocess, nor stay
// silent while it runs.
//
//	manifest   untracked non-ignored files keep full content hashing; ignored
//	           files are recorded by path + size + type through Lstat only and
//	           never read (the same-size content blind spot is documented in
//	           captureTreeManifest and pinned here as behavior).
//	timeouts   lsof and every git call are bounded; a timeout fails SAFE — the
//	           tree is skipped or the move fails, never "moved".
//	progress   the apply announces the tree count and prints one live line per
//	           tree before the final rows.
//
// Every test builds its repositories under t.TempDir() and the blocking
// subprocesses are throwaway `exec sleep` stand-ins bounded by the timeout
// under test (killed by it), so nothing keeps running after a test returns.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// t2053FakeBin installs an executable shell script named name at the front of
// PATH for the duration of the test.
func t2053FakeBin(t *testing.T, name, script string) {
	t.Helper()
	t2085SkipOnWindows(t, t2085ReasonShim)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
		t.Fatalf("write fake %s: %v", name, err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// t2053BlockingGit puts a `git` at the front of PATH that blocks (until the
// caller's timeout kills it) when any argument equals blockArg, and otherwise
// runs the real git.
func t2053BlockingGit(t *testing.T, blockArg string) {
	t.Helper()
	t2085SkipOnWindows(t, t2085ReasonShim)
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git unavailable: %v", err)
	}
	t2053FakeBin(t, "git", fmt.Sprintf(`#!/bin/sh
for a in "$@"; do
  if [ "$a" = %q ]; then exec sleep 30; fi
done
exec %q "$@"
`, blockArg, realGit))
}

// t2053IgnoredTree returns a legacy tree whose .gitignore (committed) hides
// *.log, plus the path of an ignored file inside it.
func t2053IgnoredTree(t *testing.T) (tree, ignored string) {
	t.Helper()
	_, tree, _ = rfxLegacyWorld(t)
	if err := os.WriteFile(filepath.Join(tree, ".gitignore"), []byte("*.log\n"), 0o644); err != nil {
		t.Fatalf("write gitignore: %v", err)
	}
	m6Git(t, tree, "add", ".gitignore")
	m6Git(t, tree, "commit", "-q", "-m", "ignore logs")
	ignored = filepath.Join(tree, "bulk.log")
	if err := os.WriteFile(ignored, []byte("aaaa"), 0o644); err != nil {
		t.Fatalf("write ignored file: %v", err)
	}
	return tree, ignored
}

// TestCaptureManifest_IgnoredContentIsNotRead pins the documented blind spot:
// a same-size content change inside an IGNORED file leaves the manifest
// unchanged, because the file is recorded by path + size + type and its
// content is never read. (Before card t2053 every ignored file was read and
// SHA-256 hashed — the node_modules cost.)
func TestCaptureManifest_IgnoredContentIsNotRead(t *testing.T) {
	tree, ignored := t2053IgnoredTree(t)
	before := captureManifest(tree)
	if err := os.WriteFile(ignored, []byte("bbbb"), 0o644); err != nil {
		t.Fatalf("rewrite ignored file: %v", err)
	}
	if after := captureManifest(tree); before != after {
		t.Errorf("a same-size content change in an ignored file changed the manifest: the content was read")
	}
}

// TestCaptureManifest_IgnoredUnreadableFile_DeletionAndSizeChangeDetected — an
// ignored file that cannot even be opened (mode 000) is still tracked by
// metadata: its deletion and a size change are detected, and an untouched file
// reads the same twice.
func TestCaptureManifest_IgnoredUnreadableFile_DeletionAndSizeChangeDetected(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("mode 000 does not stop root from reading")
	}
	tree, ignored := t2053IgnoredTree(t)
	if err := os.Chmod(ignored, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(ignored, 0o644) })

	base := captureManifest(tree)
	if again := captureManifest(tree); again != base {
		t.Fatalf("the manifest is not stable across two captures of an untouched tree")
	}

	// Size change: widen the file, then lock it again.
	if err := os.Chmod(ignored, 0o644); err != nil {
		t.Fatalf("chmod 644: %v", err)
	}
	if err := os.WriteFile(ignored, []byte("aaaa-longer"), 0o644); err != nil {
		t.Fatalf("grow: %v", err)
	}
	if err := os.Chmod(ignored, 0o000); err != nil {
		t.Fatalf("chmod 000: %v", err)
	}
	grown := captureManifest(tree)
	if grown == base {
		t.Errorf("a size change in an unreadable ignored file went undetected")
	}

	// Deletion.
	if err := os.Remove(ignored); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if gone := captureManifest(tree); gone == grown || gone == base {
		t.Errorf("deleting an ignored file went undetected")
	}
}

// TestCaptureManifest_IgnoredTypeChangeDetected — a regular file replaced by a
// symlink of the same byte length is a type change.
func TestCaptureManifest_IgnoredTypeChangeDetected(t *testing.T) {
	tree, ignored := t2053IgnoredTree(t)
	if err := os.WriteFile(ignored, []byte("ab"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	before := captureManifest(tree)
	if err := os.Remove(ignored); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := os.Symlink("xy", ignored); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	if after := captureManifest(tree); after == before {
		t.Errorf("a regular file replaced by a same-length symlink went undetected")
	}
}

// TestCaptureManifest_UntrackedContentChangeStillDetected — an untracked,
// NON-ignored file keeps full content hashing: a same-size change is seen.
func TestCaptureManifest_UntrackedContentChangeStillDetected(t *testing.T) {
	tree, _ := t2053IgnoredTree(t)
	notes := filepath.Join(tree, "notes.txt")
	if err := os.WriteFile(notes, []byte("one"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	before := captureManifest(tree)
	if err := os.WriteFile(notes, []byte("two"), 0o644); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	if after := captureManifest(tree); after == before {
		t.Errorf("a same-size content change in an untracked, non-ignored file went undetected")
	}
}

// TestDefaultActiveProcessCWDs_TimesOutOnBlockedLsof — a hung lsof returns an
// error wrapping ErrCommandTimeout instead of blocking the update.
func TestDefaultActiveProcessCWDs_TimesOutOnBlockedLsof(t *testing.T) {
	t2053FakeBin(t, "lsof", "#!/bin/sh\nexec sleep 30\n")
	orig := lsofCommandTimeout
	lsofCommandTimeout = 300 * time.Millisecond
	t.Cleanup(func() { lsofCommandTimeout = orig })

	start := time.Now()
	_, err := defaultActiveProcessCWDs()
	if err == nil {
		t.Fatalf("a blocked lsof returned success")
	}
	if !errors.Is(err, ErrCommandTimeout) {
		t.Errorf("error %v does not wrap ErrCommandTimeout", err)
	}
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Errorf("the lsof scan took %s; the bound did not hold", elapsed)
	}
}

// TestPlanMigration_SkipsEveryTreeWhenLsofTimesOut — the process-cwd scan is a
// skip cause: when it cannot answer, no tree is planned for a move.
func TestPlanMigration_SkipsEveryTreeWhenLsofTimesOut(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)
	activeProcessCWDProbe = defaultActiveProcessCWDs // this test pins the real scan; m6NewWorld stubs it
	t2053FakeBin(t, "lsof", "#!/bin/sh\nexec sleep 30\n")
	orig := lsofCommandTimeout
	lsofCommandTimeout = 300 * time.Millisecond
	t.Cleanup(func() { lsofCommandTimeout = orig })

	var found bool
	plans, planErr := planMigrationFor(w.repo, time.Now(), false)
	if planErr != nil {
		t.Fatalf("plan: %v", planErr)
	}
	for _, p := range plans {
		if p.source != w.movedTree {
			continue
		}
		found = true
		if !strings.Contains(p.skip, "cannot inspect active process working directories") ||
			!strings.Contains(p.skip, "timed out") {
			t.Errorf("legacy tree skip = %q, want the unanswerable lsof scan with a timeout reason", p.skip)
		}
	}
	if !found {
		t.Fatalf("the plan does not list %s", w.movedTree)
	}
}

// TestMigrateApply_ManifestGitTimeout_TreeNotMoved — a git call that hangs
// while the pre-move manifest is captured leaves the tree where it is, with a
// reported reason; the engine never reports it moved.
func TestMigrateApply_ManifestGitTimeout_TreeNotMoved(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)
	t2053BlockingGit(t, "status")
	orig := gitCommandTimeout
	// Long enough that a loaded machine does not time out the real git calls
	// (the listing, rev-parse) — only the blocked `status` may hit it.
	gitCommandTimeout = 2 * time.Second
	t.Cleanup(func() { gitCommandTimeout = orig })

	report, err := t2053EngineApply(t, w.repo)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	if report.moved != 0 {
		t.Errorf("moved = %d, want 0 when the manifest cannot be captured", report.moved)
	}
	var sawTimeout bool
	for _, r := range report.rows {
		if r.Outcome == "moved" {
			t.Errorf("row %s reported moved", r.Source)
		}
		if r.Source == w.movedTree {
			sawTimeout = strings.Contains(r.SkipReason, "timed out")
			if r.Outcome != "skipped" {
				t.Errorf("legacy tree outcome = %q, want skipped", r.Outcome)
			}
		}
	}
	if !sawTimeout {
		t.Errorf("the legacy tree row carries no timed-out reason: %+v", report.rows)
	}
	if fi, statErr := os.Lstat(w.movedTree); statErr != nil || !fi.IsDir() {
		t.Errorf("the tree is no longer at its source (stat err=%v)", statErr)
	}
}

// TestMigrateApply_MoveTimeout_TreeStaysAndFails — a hung `git worktree move`
// fails the tree with a timed-out reason; nothing reaches the destination.
func TestMigrateApply_MoveTimeout_TreeStaysAndFails(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)
	t2053BlockingGit(t, "move")
	orig := gitMoveTimeout
	gitMoveTimeout = 300 * time.Millisecond
	t.Cleanup(func() { gitMoveTimeout = orig })

	report, err := t2053EngineApply(t, w.repo)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	if report.moved != 0 {
		t.Errorf("moved = %d, want 0", report.moved)
	}
	var legacy *migrationRow
	for i := range report.rows {
		if report.rows[i].Source == w.movedTree {
			legacy = &report.rows[i]
		}
	}
	if legacy == nil {
		t.Fatalf("no row for %s", w.movedTree)
	}
	if legacy.Outcome != "failed" || !strings.Contains(legacy.SkipReason, "timed out") {
		t.Errorf("legacy row = %+v, want failed with a timed-out reason", *legacy)
	}
	if _, statErr := os.Lstat(filepath.Join(w.homeRoot, "legacy6")); !os.IsNotExist(statErr) {
		t.Errorf("the destination exists after a timed-out move (stat err=%v)", statErr)
	}
	if !m6Registered(t, w.repo, w.movedTree) {
		t.Errorf("the tree is no longer registered at its source")
	}
}

// t2053EngineApply runs the apply engine the way the entry points do, with the
// migration state directory their lock step creates.
func t2053EngineApply(t *testing.T, root string) (*migrationReport, error) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(migrationLockPath(root)), 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	return runMigrationEngine(root, true, "migrate")
}

// t2053TwoTreeWorld is the m6 world cut down to exactly two eligible legacy
// trees (legacy6 and rc30-preview6).
func t2053TwoTreeWorld(t *testing.T) *m6World {
	t.Helper()
	w := m6NewWorld(t)
	m6Git(t, w.repo, "worktree", "remove", "--force", w.workerTree)
	m6Git(t, w.repo, "worktree", "remove", "--force", w.occupiedTree)
	m4InRepo(t, w.repo)
	return w
}

var t2053ProgressLine = regexp.MustCompile(`Migrating worktrees \((\d+)/(\d+)\) (\S+)`)

// TestUpdateConsolidation_NonTTYProgress — on a non-TTY writer the apply
// announces the count, then writes one plain "(i/N) <tree>" line per tree,
// all BEFORE the final rows.
func TestUpdateConsolidation_NonTTYProgress(t *testing.T) {
	w := t2053TwoTreeWorld(t)

	buf := &strings.Builder{}
	if err := runUpdateConsolidation(w.repo, false, buf); err != nil {
		t.Fatalf("consolidation: %v\n%s", err, buf.String())
	}
	out := buf.String()
	if strings.Contains(out, "\x1b[") || strings.Contains(out, "\r") {
		t.Errorf("non-TTY progress carries escape sequences:\n%q", out)
	}
	one, two := strings.Index(out, "Migrating worktrees (1/2)"), strings.Index(out, "Migrating worktrees (2/2)")
	firstRow := strings.Index(out, "Worktree migration moved")
	if one < 0 || two < 0 {
		t.Fatalf("missing (1/2) or (2/2) progress line:\n%s", out)
	}
	if one >= two || two >= firstRow {
		t.Errorf("progress order wrong: (1/2)@%d (2/2)@%d first row@%d\n%s", one, two, firstRow, out)
	}
	if matches := t2053ProgressLine.FindAllStringSubmatch(out, -1); len(matches) != 2 || matches[0][3] == matches[1][3] {
		t.Errorf("want two progress lines naming two different trees, got %v", matches)
	}
	if !strings.Contains(out, "2 tree(s)") {
		t.Errorf("the count is not announced before the live lines:\n%s", out)
	}
	if !strings.Contains(out, "Worktree migration: 2 moved, 2 skipped, 0 failed") {
		t.Errorf("the final summary row changed:\n%s", out)
	}
}

// TestUpdateConsolidation_DryRunPrintsNoProgress — the preview moves nothing
// and prints no live lines.
func TestUpdateConsolidation_DryRunPrintsNoProgress(t *testing.T) {
	w := t2053TwoTreeWorld(t)
	buf := &strings.Builder{}
	if err := runUpdateConsolidation(w.repo, true, buf); err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if strings.Contains(buf.String(), "Migrating worktrees") {
		t.Errorf("the dry run printed progress:\n%s", buf.String())
	}
}

// TestMigrateVerb_ApplyShowsLiveLines — the verb's apply shows the live lines.
func TestMigrateVerb_ApplyShowsLiveLines(t *testing.T) {
	_ = t2053TwoTreeWorld(t)
	out, err := m4RunMigrate(t, "--apply")
	if err != nil {
		t.Fatalf("apply: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Migrating worktrees (1/2)") || !strings.Contains(out, "Migrating worktrees (2/2)") {
		t.Errorf("the verb's apply shows no live lines:\n%s", out)
	}
}

// TestMigrateVerb_ApplyJSONStaysClean — --json output stays a parseable array:
// no live line leaks into it.
func TestMigrateVerb_ApplyJSONStaysClean(t *testing.T) {
	_ = t2053TwoTreeWorld(t)
	out, err := m4RunMigrate(t, "--apply", "--json")
	if err != nil {
		t.Fatalf("apply --json: %v\n%s", err, out)
	}
	var rows []map[string]any
	if uerr := json.Unmarshal([]byte(out), &rows); uerr != nil {
		t.Errorf("--json output is not a JSON array (progress leaked in?): %v\n%s", uerr, out)
	}
}

// TestPlanMigration_SkipsATreeAnActiveProcessIsInside — the process-cwd scan is
// a skip cause in both directions: a probe that names a process inside a tree
// skips that tree, while a tree it does not name stays eligible. m6NewWorld
// reports no active process by default, so this pins the occupied cause that
// the shared fixture no longer reaches through a live lsof (card t2064).
func TestPlanMigration_SkipsATreeAnActiveProcessIsInside(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)
	activeProcessCWDProbe = func() ([]string, error) { return []string{w.movedTree}, nil }

	plans, planErr := planMigrationFor(w.repo, time.Now(), false)
	if planErr != nil {
		t.Fatalf("plan: %v", planErr)
	}
	var sawOccupied, sawFree bool
	for _, p := range plans {
		switch p.source {
		case w.movedTree:
			sawOccupied = true
			if p.skip != "an active process is inside this worktree" {
				t.Errorf("occupied tree skip = %q, want the active-process cause", p.skip)
			}
		case w.workerTree:
			sawFree = true
			if p.skip != "" {
				t.Errorf("unoccupied eligible tree skip = %q, want it planned for a move", p.skip)
			}
		}
	}
	if !sawOccupied || !sawFree {
		t.Fatalf("the plan lists occupied=%v free=%v, want both", sawOccupied, sawFree)
	}
}
