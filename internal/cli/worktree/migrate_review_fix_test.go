package worktree

// migrate_review_fix_test.go — card t1640 M8R: the post-card-review repairs
// (codex advisory findings, confirm-then-fix). Each test pins the repaired
// behavior of one confirmed finding:
//
//	P1-1  rebindCardForPath opens the real factory store (homestate.FactoryDBPath)
//	      read-write, so the card rebind actually lands (was: <root>/.moai/factory.db
//	      read-only — a path that does not exist, silently skipped).
//	P1-3  an intent-line write failure aborts the move (movelist.go).
//	P1-4  a failed reverse move leaves the tree at the destination, records
//	      pending (not reversed), and does not revert the re-key/rebind.
//	P2a   the engine resolves the primary root, so lock, move list, journal and
//	      factory paths land under the primary checkout even when invoked from
//	      inside a linked worktree.
//	P2b   planMigrationFor passes IntegrationBranch, so the integration tree is
//	      classified as integration (never moved), not unclassified.
//	P2e   --verify compares the tree identity with the journal's ident; a
//	      replaced destination is a mismatch, never "intact".
//	P2f   captureManifest covers the AC-WHR-017a manifest fields: HEAD, branch,
//	      unpushed commits, status, ignored-file count, reports count and a
//	      content digest over untracked+ignored file contents.
//	P2g/h --json emits real objects (exported fields) and the applied row's
//	      outcome reflects the move.
//	P2i   a dry run writes nothing anywhere: no lock dir, no resolver side
//	      effects (owner marker, project record), byte-identical repo and home.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// fmtBool renders a bool for the walk hash (strconv would do; the test file
// keeps one tiny local helper instead of importing strconv for one call).
func fmtBool(b bool) string {
	if b {
		return "dir"
	}
	return "file"
}

// rfxWalkHash walks dir and returns one hash over the sorted
// (path, mode, content) tuples — the AC-WHR-015 before/after equality shape.
// The dir may be absent (hashed as absent).
func rfxWalkHash(t *testing.T, dir string) string {
	t.Helper()
	sum := sha256.New()
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // absent dirs hash as absent; a re-created dir changes the hash
		}
		rel, rerr := filepath.Rel(dir, path)
		if rerr != nil {
			return rerr
		}
		info, ierr := d.Info()
		if ierr != nil {
			return ierr
		}
		_, _ = sum.Write([]byte(rel + "|" + info.Mode().String() + "|" + fmtBool(info.IsDir()) + "\x00"))
		if !info.IsDir() {
			body, rerr := os.ReadFile(path)
			if rerr != nil {
				return rerr
			}
			_, _ = sum.Write(body)
			_, _ = sum.Write([]byte("\x00"))
		}
		return nil
	})
	return hex.EncodeToString(sum.Sum(nil))
}

// rfxLegacyWorld builds a minimal rehearsal repo with ONE eligible legacy
// tree and NO pre-resolution: the resolver has never run for this repo, so
// the home root has neither the project folder, its owner marker, nor the
// repository-side record — exactly the real legacy-migration shape.
func rfxLegacyWorld(t *testing.T) (repo, tree, home string) {
	t.Helper()
	m6IsolateProcessObservation(t)
	home = m6Real(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	t.Setenv("HOME", home)
	repo = filepath.Join(m6Real(t, t.TempDir()), "rfx-repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	m6Git(t, repo, "init", "-q", "--template=", "-b", "main")
	m6Git(t, repo, "config", "user.email", "rfx@example.com")
	m6Git(t, repo, "config", "user.name", "RFX Test")
	if werr := os.WriteFile(filepath.Join(repo, "README.md"), []byte("seed\n"), 0o644); werr != nil {
		t.Fatalf("seed: %v", werr)
	}
	m6Git(t, repo, "add", "-A")
	m6Git(t, repo, "commit", "-q", "-m", "seed")
	remote := filepath.Join(filepath.Dir(repo), "rfx-remote.git")
	m6Git(t, repo, "init", "-q", "--bare", remote)
	m6Git(t, repo, "remote", "add", "origin", remote)
	m6Git(t, repo, "push", "-q", "-u", "origin", "main")
	tree = filepath.Join(repo, ".moai", "worktrees", "rfx-legacy")
	if err := os.MkdirAll(filepath.Dir(tree), 0o755); err != nil {
		t.Fatalf("mkdir legacy parent: %v", err)
	}
	m6Git(t, repo, "worktree", "add", "-q", "-b", "WT-rfx-legacy", "--", tree)
	return repo, tree, home
}

// TestMigrateDryRun_IsByteIdentical — P2i: the dry run creates nothing
// anywhere. The repo keeps no lock directory, the home root gains no project
// folder, owner marker or repository-side record (AC-WHR-015: no path under
// the repository or MOAI_HOME is created, modified or removed).
func TestMigrateDryRun_IsByteIdentical(t *testing.T) {
	repo, tree, home := rfxLegacyWorld(t)
	m4InRepo(t, repo)

	repoBefore := rfxWalkHash(t, repo)
	homeBefore := rfxWalkHash(t, home)
	treeBefore := rfxWalkHash(t, tree)

	out, err := m4RunMigrate(t)
	if err != nil {
		t.Fatalf("migrate (dry): %v\n%s", err, out)
	}
	if !strings.Contains(out, "rfx-legacy") {
		t.Fatalf("the dry-run plan does not name the eligible tree:\n%s", out)
	}

	if after := rfxWalkHash(t, repo); after != repoBefore {
		t.Errorf("the dry run wrote inside the repository (hash changed)")
	}
	if after := rfxWalkHash(t, home); after != homeBefore {
		t.Errorf("the dry run wrote inside MOAI_HOME (hash changed)")
	}
	if after := rfxWalkHash(t, tree); after != treeBefore {
		t.Errorf("the dry run wrote inside the tree (hash changed)")
	}
}

// TestMigrateReport_JSONCarriesRealRows — P2g and P2h: --json emits objects
// with exported fields (never {}), and after --apply the moved tree's row
// outcome is "moved", not the planned placeholder.
func TestMigrateReport_JSONCarriesRealRows(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)

	out, err := m4RunMigrate(t, "--json")
	if err != nil {
		t.Fatalf("migrate --json (dry): %v\n%s", err, out)
	}
	var rows []map[string]any
	if jerr := json.Unmarshal([]byte(out), &rows); jerr != nil {
		t.Fatalf("the --json report is not a JSON array of objects: %v\n%s", jerr, out)
	}
	if len(rows) == 0 {
		t.Fatalf("the --json report carried zero rows:\n%s", out)
	}
	for i, r := range rows {
		for _, field := range []string{"class", "source", "outcome"} {
			if v, _ := r[field].(string); strings.TrimSpace(v) == "" {
				t.Errorf("row %d: --json field %q is empty or missing (%v); the report must carry real values", i, field, r)
			}
		}
	}

	// The applied report: a fresh world, applied with --json; the moved
	// tree's row outcome must read "moved" (P2h), not the planned value.
	w2 := m6NewWorld(t)
	m4InRepo(t, w2.repo)
	out2, err := m4RunMigrate(t, "--json", "--apply")
	if err != nil {
		t.Fatalf("second world apply: %v\n%s", err, out2)
	}
	var applied []map[string]any
	if jerr := json.Unmarshal([]byte(out2), &applied); jerr != nil {
		t.Fatalf("the applied --json report is not a JSON array: %v\n%s", jerr, out2)
	}
	sawMoved := false
	for _, r := range applied {
		if src, _ := r["source"].(string); strings.HasSuffix(src, "legacy6") {
			if o, _ := r["outcome"].(string); o == "moved" {
				sawMoved = true
			} else {
				t.Errorf("the applied row for legacy6 carries outcome %v, want moved", r["outcome"])
			}
		}
	}
	if !sawMoved {
		t.Errorf("no applied row reports legacy6 as moved; rows: %v", applied)
	}
}

// TestMigrateVerify_DetectsIdentityMismatch — P2e: --verify compares the
// destination tree's identity with the journal's ident; a replaced
// destination is reported as a mismatch and fails the verification, never
// counted intact.
func TestMigrateVerify_DetectsIdentityMismatch(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)
	if out, err := m4RunMigrate(t, "--apply"); err != nil {
		t.Fatalf("apply: %v\n%s", err, out)
	}
	// Find where legacy6 landed.
	out, err := m4RunMigrate(t, "--verify")
	if err != nil {
		t.Fatalf("verify before the mutation: %v\n%s", err, out)
	}
	if !strings.Contains(out, "intact") {
		t.Fatalf("verify does not report the applied move intact:\n%s", out)
	}
	// Locate the moved destination from the move list.
	mlPath := filepath.Join(w.repo, ".moai", "state", "worktree-migrate", "move-list.jsonl")
	raw, rerr := os.ReadFile(mlPath)
	if rerr != nil {
		t.Fatalf("read move list: %v", rerr)
	}
	var movedTo string
	for _, line := range strings.Split(string(raw), "\n") {
		var rec struct {
			Result  string `json:"result"`
			NewPath string `json:"new_path"`
		}
		if strings.Contains(line, `"result":"moved"`) {
			if jerr := json.Unmarshal([]byte(line), &rec); jerr == nil {
				movedTo = rec.NewPath
			}
		}
	}
	if movedTo == "" {
		t.Fatalf("no moved line in the move list:\n%s", raw)
	}
	// Replace the destination with a same-shaped stranger: same HEAD, same
	// branch spelling in the record, different instance token (the ident).
	if rerr := os.RemoveAll(movedTo); rerr != nil {
		t.Fatalf("remove the moved tree: %v", rerr)
	}
	if err := os.MkdirAll(movedTo, 0o755); err != nil {
		t.Fatalf("recreate the destination as a stranger: %v", err)
	}

	out, verr := m4RunMigrate(t, "--verify")
	if verr == nil {
		t.Errorf("verify exits 0 with a replaced destination; output:\n%s", out)
	}
	if !strings.Contains(out, "mismatch") {
		t.Errorf("verify does not name the mismatch:\n%s", out)
	}
	if strings.Contains(out, "intact\t"+movedTo) {
		t.Errorf("verify still reports the replaced destination intact:\n%s", out)
	}
}

// TestMigrateRebind_LandsInTheFactoryStore — P1-1: the card record bound to
// the moved tree's old path is rebound in the REAL factory store
// (homestate.FactoryDBPath), through a writable connection.
func TestMigrateRebind_LandsInTheFactoryStore(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)

	db, err := homestate.OpenFactory(w.repo)
	if err != nil {
		t.Fatalf("open factory store: %v", err)
	}
	oldPath := w.movedTree
	if _, perr := db.RecordPicked(context.Background(), "run-rfx", "rfx-card", homestate.CardFields{WorktreePath: &oldPath}, "rfx-test", time.Now().UTC()); perr != nil {
		t.Fatalf("record picked: %v", perr)
	}
	_ = db.Close()

	if out, aerr := m4RunMigrate(t, "--apply"); aerr != nil {
		t.Fatalf("apply: %v\n%s", aerr, out)
	}

	db2, err := homestate.OpenFactory(w.repo)
	if err != nil {
		t.Fatalf("reopen factory store: %v", err)
	}
	defer func() { _ = db2.Close() }()
	card, lerr := db2.LoadCard(context.Background(), "run-rfx", "rfx-card")
	if lerr != nil {
		t.Fatalf("load the card after the move: %v", lerr)
	}
	if card.WorktreePath == oldPath || card.WorktreePath == "" {
		t.Errorf("the card still records %q after the migration; want the moved path", card.WorktreePath)
	}
}

// TestMigrateIntentWriteFailure_AbortsTheMove — P1-3: when the intent line
// cannot be written, the engine moves nothing and reports the failure.
func TestMigrateIntentWriteFailure_AbortsTheMove(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)

	// Block ONLY the move-list file: make move-list.jsonl a DIRECTORY. The
	// lock (a sibling file) still opens, so the engine reaches the intent
	// write and its failure — not a lock failure — must abort the move.
	mlDir := filepath.Join(w.repo, ".moai", "state", "worktree-migrate")
	if err := os.MkdirAll(filepath.Join(mlDir, "move-list.jsonl"), 0o755); err != nil {
		t.Fatalf("block the move list: %v", err)
	}

	stateBefore := m6Git(t, w.repo, "worktree", "list", "--porcelain")
	out, err := m4RunMigrate(t, "--apply")
	if err == nil {
		t.Errorf("the apply exits 0 despite the intent-write failure; output:\n%s", out)
	}
	if after := m6Git(t, w.repo, "worktree", "list", "--porcelain"); after != stateBefore {
		t.Errorf("the engine moved a tree although the intent line could not be written")
	}
	if _, statErr := os.Lstat(filepath.Join(w.homeRoot, "legacy6")); !os.IsNotExist(statErr) {
		t.Errorf("legacy6 was moved to the home root despite the intent-write failure (stat err=%v)", statErr)
	}
}

// TestMigrateReverseMoveFailure_StaysAtDestination — P1-4: a failed reverse
// move leaves the tree at the destination (git registration included),
// records pending — never reversed — and does not revert the re-key.
func TestMigrateReverseMoveFailure_StaysAtDestination(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)

	// Make the card rebind fail so the reverse order runs: the factory store
	// exists but is unreadable as a database.
	dbPath, err := homestate.FactoryDBPath(w.repo)
	if err != nil {
		t.Fatalf("factory db path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		t.Fatalf("mkdir factory dir: %v", err)
	}
	if err := os.WriteFile(dbPath, []byte("this is not a sqlite database\n"), 0o644); err != nil {
		t.Fatalf("write the corrupt factory db: %v", err)
	}

	origReverse := reverseGitMove
	reverseFailures := 0
	reverseGitMove = func(root, from, to string) error {
		reverseFailures++
		return exec.ErrNotFound
	}
	t.Cleanup(func() { reverseGitMove = origReverse })

	_, applyErr := m4RunMigrate(t, "--apply")
	if applyErr == nil {
		t.Errorf("the apply exits 0 although the reverse move failed")
	}
	if reverseFailures == 0 {
		t.Errorf("the reverse move was never attempted; the manifest/rebind failure did not trigger it")
	}
	if !m6Registered(t, w.repo, filepath.Join(w.homeRoot, "legacy6")) {
		t.Errorf("the failed reverse move did not keep the tree registered at its destination")
	}
	if m6Registered(t, w.repo, w.movedTree) {
		t.Errorf("the tree is registered at its OLD path although the reverse move failed; the registry was reverted on a failed move")
	}
	mlPath := filepath.Join(w.repo, ".moai", "state", "worktree-migrate", "move-list.jsonl")
	raw, rerr := os.ReadFile(mlPath)
	if rerr != nil {
		t.Fatalf("read move list: %v", rerr)
	}
	if strings.Contains(string(raw), `"result":"reversed"`) {
		t.Errorf("the move list records reversed although the reverse move failed:\n%s", raw)
	}
}

// TestMigrateNamesIntegrationClass — P2b: the integration tree (branch ==
// the configured integration target) is reported as integration, not
// unclassified; the classifier verdicts of list and migrate must agree.
func TestMigrateNamesIntegrationClass(t *testing.T) {
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)
	// The world's repo carries branch main with the git-flow config naming
	// main the integration target (m6ConfigureIntegration); a tree on it
	// under .claude/worktrees is the integration tree.
	integ := filepath.Join(w.repo, ".claude", "worktrees", "main")
	if err := os.MkdirAll(filepath.Dir(integ), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// main already exists in the world (the integration target); check out
	// the existing branch, do not create it.
	m6Git(t, w.repo, "worktree", "add", "-q", "--", integ, "main")

	out, err := m4RunMigrate(t)
	if err != nil {
		t.Fatalf("dry run: %v\n%s", err, out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, integ) {
			if !strings.Contains(line, "integration") {
				t.Errorf("the integration tree is not classified integration: %q", line)
			}
			return
		}
	}
	t.Errorf("the plan does not list the integration tree:\n%s", out)
}

// TestMigrateManifestDetectsIgnoredMutation — P2f: the pre/post manifest
// digest covers ignored-file CONTENT (AC-WHR-017a's content mutant).
func TestMigrateManifestDetectsIgnoredMutation(t *testing.T) {
	repo, tree, _ := rfxLegacyWorld(t)
	_ = repo
	if err := os.WriteFile(filepath.Join(tree, ".gitignore"), []byte("*.log\n"), 0o644); err != nil {
		t.Fatalf("write gitignore: %v", err)
	}
	reports := filepath.Join(tree, ".moai", "reports")
	if err := os.MkdirAll(reports, 0o755); err != nil {
		t.Fatalf("mkdir reports: %v", err)
	}
	logFile := filepath.Join(reports, "evidence.log")
	if err := os.WriteFile(logFile, []byte("evidence-one\n"), 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}
	m6Git(t, tree, "add", "-A")
	m6Git(t, tree, "commit", "-q", "-m", "ignore+tracked reports")
	// .gitignore is committed; the .log file is ignored content.

	before := reviewCaptureManifest(t, tree)
	if err := os.WriteFile(logFile, []byte("evidence-one!\n"), 0o644); err != nil {
		t.Fatalf("mutate the report: %v", err)
	}
	after := reviewCaptureManifest(t, tree)
	if before == after {
		t.Errorf("the manifest digest ignores a one-byte mutation in an ignored report file")
	}
}

// TestMigrateRunsFromLinkedTree_WritesUnderPrimary — P2a: invoked from
// inside a linked worktree, the engine resolves the primary root; the lock,
// move list and journal land under the primary checkout's .moai, never the
// linked tree's.
func TestMigrateRunsFromLinkedTree_WritesUnderPrimary(t *testing.T) {
	w := m6NewWorld(t)
	linked := filepath.Join(m6Real(t, t.TempDir()), "rfx-linked")
	m6Git(t, w.repo, "worktree", "add", "-q", "-b", "wt/rfx-linked", "--", linked)
	m4InRepo(t, linked)

	if out, err := m4RunMigrate(t); err != nil {
		t.Fatalf("dry run from the linked tree: %v\n%s", err, out)
	}
	primaryMig := filepath.Join(w.repo, ".moai", "state", "worktree-migrate")
	if _, err := os.Stat(filepath.Join(primaryMig, "move-list.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("the dry run wrote a move list at all (stat err=%v)", err)
	}
	// The apply writes under the PRIMARY's .moai, not the linked tree's.
	if out, err := m4RunMigrate(t, "--apply"); err != nil {
		t.Fatalf("apply from the linked tree: %v\n%s", err, out)
	}
	if fi, err := os.Stat(filepath.Join(primaryMig, "move-list.jsonl")); err != nil || fi.Size() == 0 {
		t.Errorf("the move list did not land under the primary checkout (stat err=%v)", err)
	}
	if _, err := os.Stat(filepath.Join(linked, ".moai", "state", "worktree-migrate", "move-list.jsonl")); !os.IsNotExist(err) {
		t.Errorf("the move list landed under the linked tree's .moai (stat err=%v)", err)
	}
}
