package worktree

// migrate_hardening_a_t2053_test.go — card t2053 consolidated audit, group A:
// the journal and the verify entry.
//
//	C1   the journal is replaced atomically (temp file + fsync + rename)
//	C12  two runs never share a journal file name
//	C2   --verify counts unsettled, stranded, unregistered and unreadable
//	     journals as failures, and a rolled-back step as restored

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// t2053StateDir is the migration state directory of the repository.
func t2053StateDir(repo string) string {
	return filepath.Join(repo, ".moai", "state", "worktree-migrate")
}

// t2053PlaceJournal copies a journal file into the repository's state
// directory, where --verify reads them, and returns its new path.
func t2053PlaceJournal(t *testing.T, repo, src, name string) string {
	t.Helper()
	body, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read journal: %v", err)
	}
	dir := t2053StateDir(repo)
	if merr := os.MkdirAll(dir, 0o755); merr != nil {
		t.Fatalf("mkdir state dir: %v", merr)
	}
	dst := filepath.Join(dir, name)
	if werr := os.WriteFile(dst, body, 0o644); werr != nil {
		t.Fatalf("place journal: %v", werr)
	}
	return dst
}

// TestJournalWrite_ReplacesTheFileAtomically — C1: a journal flush replaces the
// file instead of truncating it in place (an in-place truncate leaves a
// partial journal if the process dies mid-write, and the partial journal is
// the recovery handle for every tree the run already moved).
func TestJournalWrite_ReplacesTheFileAtomically(t *testing.T) {
	t2085SkipOnWindows(t, t2085ReasonSameFile)
	path := filepath.Join(t.TempDir(), "run.json")
	if err := journalWrite(path, []byte(`{"steps":[]}`)); err != nil {
		t.Fatalf("first write: %v", err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if werr := journalWrite(path, []byte(`{"steps":[{"outcome":"pending"}]}`)); werr != nil {
		t.Fatalf("second write: %v", werr)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if os.SameFile(before, after) {
		t.Errorf("the journal was rewritten in place (same inode): a kill mid-write leaves a truncated journal")
	}
	body, _ := os.ReadFile(path)
	if !json.Valid(body) || !strings.Contains(string(body), "pending") {
		t.Errorf("journal content after the second write: %q", body)
	}
}

// TestJournalWrite_InterruptedReplaceKeepsThePreviousJournal — C1: when the
// final replace step fails (the process died before the rename), the previous
// journal is still complete and parseable and no temp file is left behind.
func TestJournalWrite_InterruptedReplaceKeepsThePreviousJournal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "run.json")
	if err := journalWrite(path, []byte(`{"steps":[{"outcome":"moved"}]}`)); err != nil {
		t.Fatalf("first write: %v", err)
	}
	orig := journalReplace
	journalReplace = func(oldpath, newpath string) error { return errors.New("killed before the rename") }
	t.Cleanup(func() { journalReplace = orig })

	if err := journalWrite(path, []byte(`{"steps":[{"outcome":"moved"},{"outcome":"pending"}]}`)); err == nil {
		t.Fatalf("an interrupted replace reported success")
	}
	body, _ := os.ReadFile(path)
	if !json.Valid(body) || !strings.Contains(string(body), `"moved"`) || strings.Contains(string(body), "pending") {
		t.Errorf("the previous journal was damaged: %q", body)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("a temp file was left behind: %v", entries)
	}
}

// TestMigrateApply_JournalNameCollision_NeverOverwrites — C12: a second run
// that lands on the same second-resolution stamp keeps the first journal.
func TestMigrateApply_JournalNameCollision_NeverOverwrites(t *testing.T) {
	w := t2053TwoTreeWorld(t)
	orig := journalStamp
	journalStamp = func() string { return "20990101T000000" }
	t.Cleanup(func() { journalStamp = orig })

	first := filepath.Join(t2053StateDir(w.repo), "20990101T000000.json")
	if err := os.MkdirAll(filepath.Dir(first), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	const sentinel = `{"steps":[{"old_path":"/first/run","outcome":"moved"}]}`
	if err := os.WriteFile(first, []byte(sentinel), 0o644); err != nil {
		t.Fatalf("write first journal: %v", err)
	}
	if _, err := t2053EngineApply(t, w.repo); err != nil {
		t.Fatalf("engine: %v", err)
	}
	if got, _ := os.ReadFile(first); string(got) != sentinel {
		t.Errorf("the first run's journal was overwritten:\n%s", got)
	}
	entries, _ := os.ReadDir(t2053StateDir(w.repo))
	journals := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			journals++
		}
	}
	if journals != 2 {
		t.Errorf("want the sentinel journal plus the new run's, got %d journal file(s)", journals)
	}
}

// TestVerify_NoJournalDirectory_NothingToVerify — C2: a repository that never
// migrated has nothing to verify; the verb is a success, and creates nothing.
func TestVerify_NoJournalDirectory_NothingToVerify(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)
	out, err := m4RunMigrate(t, "--verify")
	if err != nil {
		t.Fatalf("verify of a repository without journals failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "nothing to verify") {
		t.Errorf("output: %q", out)
	}
	if _, serr := os.Stat(t2053StateDir(w.repo)); !os.IsNotExist(serr) {
		t.Errorf("verify created the state directory (%v)", serr)
	}
}

// TestVerify_IntactMove_IsSuccess — C2 baseline: a settled move whose tree sits
// at the destination under its identity and is registered there is intact.
func TestVerify_IntactMove_IsSuccess(t *testing.T) {
	w, p, id, _ := t2053ReverseWorld(t)
	t2053PlaceJournal(t, w.repo, t2053RollbackJournal(t, p, id), "20990101T000001.json")
	out, err := m4RunMigrate(t, "--verify")
	if err != nil {
		t.Fatalf("verify of an intact move failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "intact") {
		t.Errorf("output: %q", out)
	}
}

// TestVerify_PendingStep_IsAFailure — C2: a step left pending by a killed run
// (here: the tree was renamed back and git still registers the destination)
// is reported as `pending` and the verb exits non-zero.
func TestVerify_PendingStep_IsAFailure(t *testing.T) {
	w, _, journal, _ := t2053RenamedBackWorld(t)
	t2053PlaceJournal(t, w.repo, journal, "20990101T000002.json")
	out, err := m4RunMigrate(t, "--verify")
	if err == nil {
		t.Fatalf("verify reported success with a pending step:\n%s", out)
	}
	if !strings.Contains(out, "pending") {
		t.Errorf("no `pending` row in the output: %q", out)
	}
}

// TestVerify_UnreadableJournal_IsAFailure — C2: an unparseable journal is a
// row and a failure, never a silent skip.
func TestVerify_UnreadableJournal_IsAFailure(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)
	dir := t2053StateDir(w.repo)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "20990101T000003.json"), []byte(`{"steps":[{"old_pa`), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	out, err := m4RunMigrate(t, "--verify")
	if err == nil {
		t.Fatalf("verify reported success with a truncated journal:\n%s", out)
	}
	if !strings.Contains(out, "unreadable") {
		t.Errorf("no `unreadable` row in the output: %q", out)
	}
}

// TestVerify_MovedButUnregistered_IsAFailure — C2: the directory exists under
// the journal identity but git registers the OLD path (the half-moved state);
// existence plus identity alone is not "intact".
func TestVerify_MovedButUnregistered_IsAFailure(t *testing.T) {
	w, p, id, _ := t2053HalfMovedWorld(t)
	t2053PlaceJournal(t, w.repo, t2053RollbackJournal(t, p, id), "20990101T000004.json")
	out, err := m4RunMigrate(t, "--verify")
	if err == nil {
		t.Fatalf("verify reported a half-moved tree as intact:\n%s", out)
	}
	if !strings.Contains(out, "unregistered") {
		t.Errorf("no `unregistered` row in the output: %q", out)
	}
}

// TestVerify_RolledBackStep_IsNotAFailure — C2: a step the rollback restored
// reads as restored, not as an absent destination.
func TestVerify_RolledBackStep_IsNotAFailure(t *testing.T) {
	w, p, id, _ := t2053ReverseWorld(t)
	journal := t2053PlaceJournal(t, w.repo, t2053RollbackJournal(t, p, id), "20990101T000005.json")
	if out, err := m4RunMigrate(t, "--rollback", journal); err != nil {
		t.Fatalf("rollback: %v\n%s", err, out)
	}
	out, err := m4RunMigrate(t, "--verify")
	if err != nil {
		t.Fatalf("verify after a rollback failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "restored") {
		t.Errorf("the rolled-back step is not named restored: %q", out)
	}
}

// TestVerify_JSON — C16: --verify --json emits one JSON array of rows.
func TestVerify_JSON(t *testing.T) {
	w, _, journal, _ := t2053RenamedBackWorld(t)
	t2053PlaceJournal(t, w.repo, journal, "20990101T000006.json")
	out, _ := m4RunMigrate(t, "--verify", "--json")
	var rows []map[string]any
	// The verb also prints cobra's "Error:" line after the array; decode the
	// first JSON value only.
	if err := json.NewDecoder(strings.NewReader(out)).Decode(&rows); err != nil {
		t.Fatalf("--verify --json is not a JSON array: %v\n%s", err, out)
	}
	if len(rows) == 0 || rows[0]["kind"] != "pending" {
		t.Errorf("rows: %v", rows)
	}
}
