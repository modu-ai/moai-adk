package worktree

// migrate_hardening_c_t2053_test.go — card t2053 consolidated audit, group C:
// the rerun and the report.
//
//	C7   a rerun after an interrupted migration names the journal; a failed run prints it
//	C8   report rows are printed before the error
//	C9   any non-zero git exit while reading a manifest is an error
//	C10  a state-file entry or card binding spelled differently still follows the move
//	C16  a tree already at its destination is not a row

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

// t2053InterruptedWorld leaves the state a killed forward move leaves: the
// tree renamed to its destination, git still registering the source (so the
// source reads prunable), a pending move-list line and a journal naming it.
// It returns the journal path.
func t2053InterruptedWorld(t *testing.T) (*m6World, migratePlan, string) {
	t.Helper()
	w, p, id, mlPath := t2053HalfMovedWorld(t)
	if err := appendMoveListLine(mlPath, m6MoveLine("pending", p.source, p.destination, id.Ident())); err != nil {
		t.Fatalf("pending line: %v", err)
	}
	body, _ := json.Marshal(map[string]any{"steps": []journalStep{{
		OldPath: p.source, NewPath: p.destination, Outcome: "pending", Ident: id.Ident(),
	}}})
	src := filepath.Join(t.TempDir(), "run.json")
	if err := os.WriteFile(src, body, 0o644); err != nil {
		t.Fatalf("write journal: %v", err)
	}
	return w, p, t2053PlaceJournal(t, w.repo, src, "20260101T000000.json")
}

// firstJSONValue decodes the first JSON value of out (cobra appends an
// "Error:" line after it).
func firstJSONValue(t *testing.T, out string, v any) {
	t.Helper()
	if err := json.NewDecoder(strings.NewReader(out)).Decode(v); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
}

// TestPlan_InterruptedMigration_NamesTheJournal — C7: a rerun finds the
// unfinished migration, says so, names the journal and the rollback — and does
// not send the operator to `git worktree prune`.
func TestPlan_InterruptedMigration_NamesTheJournal(t *testing.T) {
	_, p, journal := t2053InterruptedWorld(t)

	out, err := m4RunMigrate(t)
	if err == nil {
		t.Fatalf("a rerun over an interrupted migration exited clean:\n%s", out)
	}
	for _, want := range []string{"interrupted", journal, "--rollback"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "prune") {
		t.Errorf("the output invites `git worktree prune`:\n%s", out)
	}

	jsonOut, _ := m4RunMigrate(t, "--json")
	var rows []map[string]any
	firstJSONValue(t, jsonOut, &rows)
	found := false
	for _, r := range rows {
		if r["source"] == p.source {
			found = true
			if r["outcome"] != "interrupted" || r["journal"] != journal {
				t.Errorf("row = %v, want outcome interrupted and journal %s", r, journal)
			}
		}
	}
	if !found {
		t.Errorf("no row for %s: %v", p.source, rows)
	}
}

// TestApply_InterruptedMigration_IsNotMovedAgain — C7: --apply over the same
// state moves nothing, writes no line, and still exits non-zero.
func TestApply_InterruptedMigration_IsNotMovedAgain(t *testing.T) {
	w, p, journal := t2053InterruptedWorld(t)
	mlPath := moveListPath(w.repo)
	before, _ := os.ReadFile(mlPath)

	out, err := m4RunMigrate(t, "--apply")
	if err == nil {
		t.Fatalf("apply over an interrupted migration exited clean:\n%s", out)
	}
	if !strings.Contains(out, journal) {
		t.Errorf("output does not name the journal %s:\n%s", journal, out)
	}
	// Other eligible trees of the world may move; the interrupted tree gets no
	// new line.
	after, _ := os.ReadFile(mlPath)
	if got, was := strings.Count(string(after), p.source), strings.Count(string(before), p.source); got != was {
		t.Errorf("the rerun wrote %d new move-list line(s) for the interrupted tree:\n%s", got-was, after)
	}
	if _, serr := os.Stat(p.destination); serr != nil {
		t.Errorf("the half-moved tree is gone from %s: %v", p.destination, serr)
	}
}

// TestUpdateConsolidation_Interrupted_WarnsAndNamesJournal — C7: the update
// path surfaces the interrupted migration as a warning naming the journal.
func TestUpdateConsolidation_Interrupted_WarnsAndNamesJournal(t *testing.T) {
	w, _, journal := t2053InterruptedWorld(t)
	buf := &strings.Builder{}
	if err := runUpdateConsolidation(w.repo, false, buf); err != nil {
		t.Fatalf("the update must degrade, not fail: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "interrupted") || !strings.Contains(buf.String(), journal) {
		t.Errorf("the update output does not warn about the interrupted migration:\n%s", buf.String())
	}
}

// TestApply_FailedTrees_PrintTheJournalPath — C7: a run that failed trees
// prints the journal path, where --rollback reads from. Since t2094 (F3) the
// rollback hint is gated on the journal holding moves to undo, so the failure
// here is one tree of two: the other moves, the journal carries that move, and
// the hint prints (a run where every tree failed before moving no longer
// advertises rollback — pinned in migrate_busy_tree_t2094_test.go).
func TestApply_FailedTrees_PrintTheJournalPath(t *testing.T) {
	w := t2053TwoTreeWorld(t)
	orig := forwardGitMove
	forwardGitMove = func(root, from, to string) error {
		if from == w.movedTree {
			return fmt.Errorf("fatal: refused before any rename")
		}
		return orig(root, from, to)
	}
	t.Cleanup(func() { forwardGitMove = orig })

	out, err := m4RunMigrate(t, "--apply")
	if err == nil {
		t.Fatalf("failed trees exited clean:\n%s", out)
	}
	journals, _ := filepath.Glob(filepath.Join(t2053StateDir(w.repo), "*.json"))
	if len(journals) != 1 {
		t.Fatalf("journals = %v, want exactly one", journals)
	}
	if !strings.Contains(out, journals[0]) {
		t.Errorf("the output does not name the journal %s:\n%s", journals[0], out)
	}

	buf := &strings.Builder{}
	if uerr := runUpdateConsolidation(w.repo, false, buf); uerr != nil {
		t.Fatalf("update: %v", uerr)
	}
	updateJournals, _ := filepath.Glob(filepath.Join(t2053StateDir(w.repo), "*.json"))
	if len(updateJournals) != 2 {
		t.Fatalf("journals after the update = %v, want two", updateJournals)
	}
	named := false
	for _, j := range updateJournals {
		named = named || strings.Contains(buf.String(), j)
	}
	if !named {
		t.Errorf("the update output names no journal:\n%s", buf.String())
	}
}

// t2053ReadOnlyMoveList makes the move list unwritable once the first line is
// in, by wrapping the forward move seam.
func t2053ReadOnlyMoveList(t *testing.T, root string) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("root ignores file modes")
	}
	orig := forwardGitMove
	forwardGitMove = func(r, from, to string) error {
		_ = os.Chmod(moveListPath(r), 0o444)
		return orig(r, from, to)
	}
	t.Cleanup(func() {
		forwardGitMove = orig
		_ = os.Chmod(moveListPath(root), 0o644)
	})
}

// TestApply_ReportRowsPrecedeTheError — C8: an engine error after trees were
// handled still prints the rows (verb text and JSON), then fails.
func TestApply_ReportRowsPrecedeTheError(t *testing.T) {
	w := t2053TwoTreeWorld(t)
	t2053ReadOnlyMoveList(t, w.repo)

	out, err := m4RunMigrate(t, "--apply", "--json")
	if err == nil {
		t.Fatalf("a move-list write failure exited clean:\n%s", out)
	}
	var rows []map[string]any
	firstJSONValue(t, out, &rows)
	if len(rows) == 0 {
		t.Errorf("no rows were printed before the error:\n%s", out)
	}
}

// TestUpdateConsolidation_ReportRowsPrecedeTheError — C8 for the update path.
func TestUpdateConsolidation_ReportRowsPrecedeTheError(t *testing.T) {
	w := t2053TwoTreeWorld(t)
	t2053ReadOnlyMoveList(t, w.repo)

	buf := &strings.Builder{}
	err := runUpdateConsolidation(w.repo, false, buf)
	if err == nil {
		t.Fatalf("a move-list write failure exited clean:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "Worktree migration moved") && !strings.Contains(buf.String(), "Worktree migration failed") {
		t.Errorf("no report lines were printed before the error:\n%s", buf.String())
	}
}

// t2053FailingGit puts a `git` at the front of PATH that exits 1 when any
// argument equals failArg, and otherwise runs the real git.
func t2053FailingGit(t *testing.T, failArg string) {
	t.Helper()
	t2053FakeBin(t, "git", fmt.Sprintf(`#!/bin/sh
for a in "$@"; do
  if [ "$a" = %q ]; then echo "fatal: injected failure" >&2; exit 1; fi
done
exec %q "$@"
`, failArg, mustLookGit(t)))
}

func mustLookGit(t *testing.T) string {
	t.Helper()
	p, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git unavailable: %v", err)
	}
	return p
}

// TestApply_ManifestGitFailure_MovesNothing — C9: a manifest read in which a
// git call exited non-zero proves nothing, so the tree stays where it is.
func TestApply_ManifestGitFailure_MovesNothing(t *testing.T) {
	w := t2053TwoTreeWorld(t)
	t2053FailingGit(t, "status")

	report, err := t2053EngineApply(t, w.repo)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	if report.moved != 0 {
		t.Errorf("moved = %d, want 0: %+v", report.moved, report.rows)
	}
	for _, r := range report.rows {
		if r.Source == w.movedTree && !strings.Contains(r.SkipReason, "manifest unavailable") {
			t.Errorf("row for %s: %+v, want a manifest-unavailable skip", r.Source, r)
		}
	}
	if _, serr := os.Stat(w.movedTree); serr != nil {
		t.Errorf("the tree left its source: %v", serr)
	}
}

// TestCaptureLegacyManifest_GitFailure_NeverMatches — C9: the legacy capture
// fails the same way (an empty digest never equals a recorded one).
func TestCaptureLegacyManifest_GitFailure_NeverMatches(t *testing.T) {
	w := t2053TwoTreeWorld(t)
	t2053FailingGit(t, "status")
	if got := captureLegacyManifest(w.movedTree); got != "" {
		t.Errorf("legacy digest = %q, want empty after a failed git call", got)
	}
}

// TestCaptureTreeManifest_UnbornHEAD_IsNotAnError — C9's benign case: a tree
// with no commit yet fails rev-parse HEAD and rev-list HEAD by nature.
func TestCaptureTreeManifest_UnbornHEAD_IsNotAnError(t *testing.T) {
	dir := m6Real(t, t.TempDir())
	m6Git(t, dir, "init", "-q", "-b", "main")
	if _, err := captureTreeManifest(dir); err != nil {
		t.Errorf("an unborn HEAD is not a failure: %v", err)
	}
}

// TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove — C10: the state
// file keys the tree by another spelling of the same directory.
func TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove(t *testing.T) {
	w := t2053TwoTreeWorld(t)
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(filepath.Dir(w.movedTree), link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	spelled := filepath.Join(link, filepath.Base(w.movedTree))
	stateFile := filepath.Join(w.repo, ".moai", "state", "worktrees.json")
	if err := os.MkdirAll(filepath.Dir(stateFile), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body, _ := json.Marshal([]map[string]any{{"path": spelled, "branch": "WT-legacy6", "agent_name": "a", "created_at": time.Now().UTC()}})
	if err := os.WriteFile(stateFile, body, 0o644); err != nil {
		t.Fatalf("seed the state file: %v", err)
	}

	report, err := t2053EngineApply(t, w.repo)
	if err != nil || report.moved != 2 {
		t.Fatalf("apply: moved=%d err=%v rows=%+v", report.moved, err, report.rows)
	}
	dest := filepath.Join(w.homeRoot, "legacy6")
	raw, _ := os.ReadFile(stateFile)
	// Decode before comparing: the file is JSON, so a Windows path is stored
	// with every backslash escaped (card t2085); a substring search of the raw
	// bytes for the unescaped path can never match there.
	var entries []struct {
		Path string `json:"path"`
	}
	if jerr := json.Unmarshal(raw, &entries); jerr != nil {
		t.Fatalf("the state file is not valid JSON: %v\n%s", jerr, raw)
	}
	var rekeyed bool
	for _, e := range entries {
		rekeyed = rekeyed || e.Path == dest
	}
	if !rekeyed {
		t.Errorf("the state entry was not re-keyed to %s:\n%s", dest, raw)
	}
}

// TestApply_CardBoundViaSymlinkSpelling_StillFollowsTheMove — C10 for the card
// binding.
func TestApply_CardBoundViaSymlinkSpelling_StillFollowsTheMove(t *testing.T) {
	w := t2053TwoTreeWorld(t)
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(filepath.Dir(w.movedTree), link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	spelled := filepath.Join(link, filepath.Base(w.movedTree))
	db, err := homestate.OpenFactory(w.repo)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	if _, perr := db.RecordPicked(context.Background(), "run-c10", "card-c10", homestate.CardFields{WorktreePath: &spelled}, "t2053-test", time.Now().UTC()); perr != nil {
		t.Fatalf("record picked: %v", perr)
	}
	_ = db.Close()

	if report, aerr := t2053EngineApply(t, w.repo); aerr != nil || report.moved != 2 {
		t.Fatalf("apply: moved=%d err=%v", report.moved, aerr)
	}
	db, err = homestate.OpenFactory(w.repo)
	if err != nil {
		t.Fatalf("reopen factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	card, lerr := db.LoadCard(context.Background(), "run-c10", "card-c10")
	if lerr != nil {
		t.Fatalf("load card: %v", lerr)
	}
	if dest := filepath.Join(w.homeRoot, "legacy6"); card.WorktreePath != dest {
		t.Errorf("card worktree = %q, want %q", card.WorktreePath, dest)
	}
}

// TestPlan_TreeAlreadyAtItsDestination_IsNotARow — C16: a tree that already
// sits at the home-root destination is not a candidate and earns no row.
func TestPlan_TreeAlreadyAtItsDestination_IsNotARow(t *testing.T) {
	w := m6NewWorld(t)
	// A home that carries the `.moai/worktrees` marker, as a real MOAI_HOME does.
	t.Setenv(paths.EnvHome, filepath.Join(w.home, ".moai"))
	root, err := worktreeroot.Root(w.repo)
	if err != nil {
		t.Fatalf("home root: %v", err)
	}
	atDest := filepath.Join(root, "atdest6")
	if merr := os.MkdirAll(filepath.Dir(atDest), 0o755); merr != nil {
		t.Fatalf("mkdir: %v", merr)
	}
	m6Git(t, w.repo, "worktree", "add", "-q", "-b", "WT-atdest6", "--", atDest)

	plans, perr := planMigrationFor(w.repo, time.Now(), false)
	if perr != nil {
		t.Fatalf("plan: %v", perr)
	}
	for _, p := range plans {
		if worktreeroot.Same(p.source, atDest) {
			t.Errorf("a tree already at its destination got a plan row: %+v", p)
		}
	}
}
