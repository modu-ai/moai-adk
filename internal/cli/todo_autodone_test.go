// todo_autodone_test.go — SPEC-TODO-LAND-AUTO-DONE-001 M2 acceptance tests:
// the `moai todo auto-done` verb, its evidence forms, its three misfire
// guards, its execution log and reversibility, and the close-surface
// exclusivity contract.
//
// Every AC is asserted behaviourally over a fixture repository and a fixture
// queue; no test touches a real remote or the operator's queue (the runTodo
// isolation guard fails the test before Execute touches any file). The fetch
// boundary (AC-AD-015) is counted through the same subprocess seam the
// census tests use: the assertion is about the whole command surface, so a
// second, unrouted exec call would be invisible to a seam only the routed
// path knows about.
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// autoDoneFixture is a todoFixture whose landed ref is pinned at chain
// level 1 (the configured git_strategy.worktree_base_branch) to
// `origin/develop`, and whose config file materializes that ref — so the
// close lines name the ref the ACs spell.
func autoDoneFixture(t *testing.T) (root string, store *kanban.BacklogStore) {
	t.Helper()
	root, store = todoFixture(t)
	cfgDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatalf("config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "git-strategy.yaml"),
		[]byte("git_strategy:\n  worktree_base_branch: develop\n"), 0o600); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}
	return root, store
}

// commitOnRef adds one empty commit with the given subject and returns its
// full SHA. The commit sits on the fixture's working branch; the test
// materializes refs/remotes/origin/develop afterwards.
func commitOnRef(t *testing.T, root, subject string) string {
	t.Helper()
	runGitIn(t, root, "commit", "--allow-empty", "-q", "-m", subject)
	return gitOut(t, root, "rev-parse", "HEAD")
}

// materializeOriginDevelop points refs/remotes/origin/develop at HEAD — the
// fixture's stand-in for the lead's pushed develop.
func materializeOriginDevelop(t *testing.T, root string) string {
	t.Helper()
	sha := gitOut(t, root, "rev-parse", "HEAD")
	runGitIn(t, root, "update-ref", "refs/remotes/origin/develop", sha)
	return sha
}

// seedCard appends a card with an EXPLICIT id, bypassing Add's issuance —
// the collision and false-negative fixtures need ids the ACs spell (t901,
// t902, ...), which the store only issues after the id space reaches them.
func seedCard(t *testing.T, store *kanban.BacklogStore, id, text string, state kanban.BacklogState) {
	t.Helper()
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		rec.Items = append(rec.Items, kanban.BacklogItem{
			ID: id, Text: text,
			AddedAt: time.Now().UTC().Format(time.RFC3339),
			State:   state,
		})
		return nil
	}); err != nil {
		t.Fatalf("seed card %s: %v", id, err)
	}
}

// seedArchivedCard appends an archived entry with an explicit id — a
// NON-colliding archive entry (the store's identity layer refuses a record
// holding the same id live AND archived, so the collision fixture injects
// its predecessor through injectArchivedPredecessor below).
func seedArchivedCard(t *testing.T, store *kanban.BacklogStore, id, text string) {
	t.Helper()
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		rec.Archived = append(rec.Archived, kanban.BacklogArchiveEntry{
			Item:     kanban.BacklogItem{ID: id, Text: text, State: kanban.BacklogStateQueued},
			Position: 0,
			Findings: []kanban.BacklogArchivedFinding{},
		})
		return nil
	}); err != nil {
		t.Fatalf("seed archived card %s: %v", id, err)
	}
}

// injectArchivedPredecessor creates the reissue fixture's predecessor half
// by writing the archived row DIRECTLY, bypassing the store's identity
// invariant — which refuses, on every whole-record write, a record holding
// the same id live AND archived (todo_identity.go ensureRecordIdentities).
// The AC-AD-004 "given" describes exactly that state, so the fixture
// constructs it at the storage layer: the reads the scan performs see the
// collision, and no write through the store is attempted while it holds
// (the collision card skips, so the scan issues no mutation).
func injectArchivedPredecessor(t *testing.T, store *kanban.BacklogStore, id, text string) {
	t.Helper()
	_, err := openQueueDB(t, store).Exec(
		`INSERT INTO archived_items(seq, id, text, added_at, spec_id, state, position)
		 VALUES ((SELECT COALESCE(MAX(seq), 0) + 1 FROM archived_items), ?, ?, ?, NULL, 'queued', 0)`,
		id, text, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("inject archived predecessor %s: %v", id, err)
	}
}

// writeSpecFixture writes a fixture SPEC directory whose frontmatter status
// is status, and returns the SPEC id.
func writeSpecFixture(t *testing.T, root, id, status string) string {
	t.Helper()
	dir := filepath.Join(root, ".moai", "specs", id)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("spec dir: %v", err)
	}
	body := "---\nid: " + id + "\nstatus: " + status + "\n---\n\n# fixture spec\n"
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(body), 0o600); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}
	return id
}

// recordBytes is the whole-record comparison the dry-run criterion asserts
// on: the full decoded record marshalled — not a field sample.
func recordBytes(t *testing.T, store *kanban.BacklogStore) []byte {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load record: %v", err)
	}
	b, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal record: %v", err)
	}
	return b
}

// scanLogPath is the scan's execution log, resolved the same way the
// command resolves it.
func scanLogPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(kanban.RuntimeStateDirForRoot(resolveTodoQueueRoot()), "auto-done-log.jsonl")
}

// readLogRows decodes the execution log, tolerating an absent file (a scan
// that has not run yet).
func readLogRows(t *testing.T, path string) []map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read log: %v", err)
	}
	var rows []map[string]any
	for _, line := range strings.Split(strings.TrimRight(string(raw), "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatalf("log row is not valid JSON: %v\nline: %s", err, line)
		}
		rows = append(rows, row)
	}
	return rows
}

// lastLogRowFor returns the LAST row naming id with the given outcome.
func lastLogRowFor(t *testing.T, path, id, outcome string) map[string]any {
	t.Helper()
	var found map[string]any
	for _, row := range readLogRows(t, path) {
		if row["card_id"] == id && row["outcome"] == outcome {
			found = row
		}
	}
	if found == nil {
		t.Fatalf("no %s row for %s in the log at %s", outcome, id, path)
	}
	return found
}

// installFetchCounter replaces the subprocess seam with one that counts
// `git fetch` invocations and delegates to the real exec otherwise, and
// returns the counter. Restored via t.Cleanup.
func installFetchCounter(t *testing.T) *int {
	t.Helper()
	count := 0
	orig := todoRunCommand
	t.Cleanup(func() { todoRunCommand = orig })
	todoRunCommand = func(name string, args ...string) (string, error) {
		if name == "git" {
			for _, a := range args {
				if a == "fetch" {
					count++
					break
				}
			}
		}
		child := exec.Command(name, args...)
		child.Dir = resolveProjectDir()
		out, err := child.Output()
		return string(out), err
	}
	return &count
}

// liveItem returns the live item for id, failing when it is absent.
func liveItem(t *testing.T, store *kanban.BacklogStore, id string) kanban.BacklogItem {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, it := range rec.Items {
		if it.ID == id {
			return it
		}
	}
	t.Fatalf("card %s is not live in the queue", id)
	return kanban.BacklogItem{}
}

// AC-AD-001 — the scan evaluates ONLY live queued/picked items: the dropped
// card's record is byte-identical before and after, and the archive holds
// exactly the two closed ids.
func TestTodoAutoDone_LiveFilter(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t901", "queued work", kanban.BacklogStateQueued)
	seedCard(t, store, "t911", "picked work", kanban.BacklogStatePicked)
	seedCard(t, store, "t912", "dropped work", kanban.BacklogStateDropped)
	seedArchivedCard(t, store, "t913", "already archived work")
	commitOnRef(t, root, "Merge branch 'WT-x' into develop (card t901)")
	commitOnRef(t, root, "Merge branch 'WT-y' into develop (card t911)")
	commitOnRef(t, root, "Merge branch 'WT-z' into develop (card t912)")
	commitOnRef(t, root, "Merge branch 'WT-w' into develop (card t913)")
	materializeOriginDevelop(t, root)

	before, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load before: %v", err)
	}
	stdout, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("auto-done: %v", err)
	}

	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load after: %v", err)
	}
	if len(rec.Items) != 1 || rec.Items[0].ID != "t912" {
		t.Fatalf("live items = %v, want only the dropped card", rec.Items)
	}
	if rec.Items[0].State != kanban.BacklogStateDropped {
		t.Errorf("dropped card state = %q, want dropped", rec.Items[0].State)
	}
	// The scan added EXACTLY the two closeable cards to the archive; the
	// dropped card joined nothing and the pre-existing entry (t913) is
	// untouched: same position, same text.
	closed := map[string]bool{}
	for _, entry := range rec.Archived {
		closed[entry.Item.ID] = true
		if entry.Item.ID == "t913" {
			if entry.Item.Text != "already archived work" || entry.Position != 0 {
				t.Errorf("pre-existing archive entry mutated: %+v", entry)
			}
		}
	}
	if !closed["t901"] || !closed["t911"] {
		t.Errorf("archive = %v, want t901 and t911", closed)
	}
	if closed["t912"] {
		t.Error("the dropped card was archived")
	}
	if len(rec.Archived) != 3 {
		t.Errorf("archive holds %d entries, want 3 (the pre-existing t913 + the two closes)", len(rec.Archived))
	}
	for _, was := range before.Items {
		if was.ID != "t912" {
			continue
		}
		if was.Text != rec.Items[0].Text || was.AddedAt != rec.Items[0].AddedAt {
			t.Error("the dropped card's record changed")
		}
	}
	if !strings.Contains(stdout, "done t901 landing=landed") || !strings.Contains(stdout, "done t911 landing=landed") {
		t.Errorf("stdout %q lacks the close lines", stdout)
	}
}

// AC-AD-002 — evidence form 1: the recorded delivering SHA closes, the close
// line carries the full contract, and the log row names form and full SHA.
func TestTodoAutoDone_FormRecordedSHA(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t921", "sha-recorded work", kanban.BacklogStateQueued)
	// The delivering commit sits BELOW the ref head, so the form-1 evidence
	// is genuinely a reachability answer, not a ref-position coincidence.
	sha := commitOnRef(t, root, "chore: the delivering commit, naming no card")
	commitOnRef(t, root, "chore: a later tip, naming no card")
	materializeOriginDevelop(t, root)

	if _, _, err := runTodo(t, "landed", "t921", "--sha", sha); err != nil {
		t.Fatalf("record sha: %v", err)
	}
	stdout, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("auto-done: %v", err)
	}

	want := "done t921 landing=landed source=auto-land ref=origin/develop form=sha-recorded"
	if !strings.Contains(stdout, want) {
		t.Errorf("stdout %q lacks close line %q", stdout, want)
	}
	row := lastLogRowFor(t, scanLogPath(t), "t921", "closed")
	if row["form"] != kanban.AutoDoneFormSHA {
		t.Errorf("log form = %v, want %q", row["form"], kanban.AutoDoneFormSHA)
	}
	if row["recorded_sha"] != sha {
		t.Errorf("log recorded_sha = %v, want the full resolved SHA %s", row["recorded_sha"], sha)
	}
	if _, ok := liveItemOK(t, store, "t921"); ok {
		t.Error("t921 still live after the scan")
	}
}

// liveItemOK reports whether id is live (the inverse convenience).
func liveItemOK(t *testing.T, store *kanban.BacklogStore, id string) (kanban.BacklogItem, bool) {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, it := range rec.Items {
		if it.ID == id {
			return it, true
		}
	}
	return kanban.BacklogItem{}, false
}

// AC-AD-003 — evidence form 2: an attributed subject closes, and the log row
// carries the attributed subject line and that commit's SHA.
func TestTodoAutoDone_FormSubjectAttribution(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t931", "subject-attributed work", kanban.BacklogStateQueued)
	subject := "Merge WT-fixture into develop (card t931)"
	sha := commitOnRef(t, root, subject)
	materializeOriginDevelop(t, root)

	stdout, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("auto-done: %v", err)
	}

	want := "done t931 landing=landed source=auto-land ref=origin/develop form=subject-attribution"
	if !strings.Contains(stdout, want) {
		t.Errorf("stdout %q lacks close line %q", stdout, want)
	}
	row := lastLogRowFor(t, scanLogPath(t), "t931", "closed")
	if row["form"] != kanban.AutoDoneFormSubject {
		t.Errorf("log form = %v, want %q", row["form"], kanban.AutoDoneFormSubject)
	}
	if row["subject"] != subject {
		t.Errorf("log subject = %v, want %q", row["subject"], subject)
	}
	if row["commit_sha"] != sha {
		t.Errorf("log commit_sha = %v, want %s", row["commit_sha"], sha)
	}
}

// AC-AD-004 — guard M1: a reissued id skips ambiguous-id on subject evidence
// alone and stays live.
func TestTodoAutoDone_CollisionSkipsAmbiguous(t *testing.T) {
	root, store := autoDoneFixture(t)
	// The predecessor carried t902 first; its `fix(t902)` commit IS on the
	// landed ref. The id was then REISSUED to a live card with different
	// text and no recorded SHA.
	seedCard(t, store, "t902", "the reissued card's different text", kanban.BacklogStateQueued)
	commitOnRef(t, root, "fix(t902): the predecessor's landed work")
	injectArchivedPredecessor(t, store, "t902", "the predecessor's text")
	materializeOriginDevelop(t, root)

	stdout, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("auto-done: %v", err)
	}
	if !strings.Contains(stdout, "skip t902 reason=ambiguous-id") {
		t.Errorf("stdout %q lacks skip t902 reason=ambiguous-id", stdout)
	}
	if _, ok := liveItemOK(t, store, "t902"); !ok {
		t.Error("t902 was archived — the collision gate failed closed")
	}
}

// AC-AD-005 — guard M1 override: a recorded SHA disambiguates a reissued id.
//
// The decision layer's proof is TestAutoDoneDecide/"recorded SHA closes
// through a collision" (internal/kanban). The CLI-level proof below runs the
// WHOLE scan over the injected reissue state: --dry-run exercises the
// decision end-to-end and prints the close the scan would produce (dry-run
// writes nothing, so the identity-invariant refusal cannot fire), and the
// non-dry run then shows the STORE's deeper guard — the identity layer
// refuses to write ANY record holding the id live and archived, so the close
// is refused loudly rather than applied. In the field the reissue shape's
// predecessor sits in another store (the id allocator's blind spot), where
// the live record holds no duplicate and the close applies.
func TestTodoAutoDone_CollisionRecordedSHAOverrides(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t902", "the reissued card's different text", kanban.BacklogStateQueued)
	// The CURRENT card's own commit: reachable, attributing nothing (the
	// recorded SHA is the disambiguation, not the subject stream).
	c2 := commitOnRef(t, root, "chore: the reissued card's own delivery")
	commitOnRef(t, root, "fix(t902): the predecessor's landed work")
	materializeOriginDevelop(t, root)
	if _, _, err := runTodo(t, "landed", "t902", "--sha", c2); err != nil {
		t.Fatalf("record sha (clean record): %v", err)
	}
	injectArchivedPredecessor(t, store, "t902", "the predecessor's text")
	materializeOriginDevelop(t, root)

	// The scan's DECISION: the recorded SHA closes through the collision.
	dryStdout, _, err := runTodo(t, "auto-done", "--dry-run")
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if !strings.Contains(dryStdout, "done t902 landing=landed source=auto-land") ||
		!strings.Contains(dryStdout, "form=sha-recorded") {
		t.Errorf("dry-run stdout %q lacks the sha-recorded close for t902", dryStdout)
	}

	// The write: the store's identity invariant refuses the whole mutation,
	// loudly — and the queue record is unchanged (Mutate's byte-identity
	// contract).
	if _, _, err := runTodo(t, "auto-done"); err == nil {
		t.Fatal("the scan wrote a record holding an id live and archived — the identity invariant did not fire")
	} else if !strings.Contains(err.Error(), "duplicate card identity") {
		t.Errorf("write error = %v, want the identity invariant's duplicate-id refusal", err)
	}
}

// AC-AD-006 — guard M2: a run-landed sync-pending card skips
// spec-not-completed while its completed control closes.
func TestTodoAutoDone_SpecNotCompletedSkips(t *testing.T) {
	root, store := autoDoneFixture(t)
	_ = writeSpecFixture(t, root, "SPEC-FIXTURE-IMP-001", "implemented")
	_ = writeSpecFixture(t, root, "SPEC-FIXTURE-DONE-001", "completed")
	seedCard(t, store, "t941", "run landed, sync pending", kanban.BacklogStatePicked)
	seedCard(t, store, "t942", "run landed, sync done", kanban.BacklogStatePicked)
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for i := range rec.Items {
			switch rec.Items[i].ID {
			case "t941":
				v := "SPEC-FIXTURE-IMP-001"
				rec.Items[i].SpecID = &v
			case "t942":
				v := "SPEC-FIXTURE-DONE-001"
				rec.Items[i].SpecID = &v
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("attach specs: %v", err)
	}
	commitOnRef(t, root, "fix(t941): the run commit")
	commitOnRef(t, root, "fix(t942): the run commit")
	materializeOriginDevelop(t, root)

	stdout, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("auto-done: %v", err)
	}
	if !strings.Contains(stdout, "skip t941 reason=spec-not-completed") {
		t.Errorf("stdout %q lacks skip t941 reason=spec-not-completed", stdout)
	}
	if _, ok := liveItemOK(t, store, "t941"); !ok {
		t.Error("t941 was archived while its SPEC is only implemented")
	}
	if !strings.Contains(stdout, "done t942 landing=landed") {
		t.Errorf("stdout %q lacks the control card's close line", stdout)
	}
	if _, ok := liveItemOK(t, store, "t942"); ok {
		t.Error("t942 (completed SPEC) stayed live — the control failed")
	}
}

// AC-AD-006 edge (§D.1) — a SPEC whose status cannot be read is NOT
// completed: unknown is not a pass.
func TestTodoAutoDone_SpecUnreadableIsNotCompleted(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t943", "spec directory missing", kanban.BacklogStateQueued)
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == "t943" {
				v := "SPEC-FIXTURE-ABSENT-001"
				rec.Items[i].SpecID = &v
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	commitOnRef(t, root, "fix(t943): landed work")
	materializeOriginDevelop(t, root)

	stdout, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("auto-done: %v", err)
	}
	if !strings.Contains(stdout, "skip t943 reason=spec-not-completed") {
		t.Errorf("stdout %q lacks the unreadable-SPEC skip", stdout)
	}
	if _, ok := liveItemOK(t, store, "t943"); !ok {
		t.Error("t943 archived on an unreadable SPEC — unknown was treated as a pass")
	}
}

// AC-AD-007 — guard M3: a non-landing declaration attributes nothing.
func TestTodoAutoDone_NegationAttributesNothing(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t951", "attempt not merged", kanban.BacklogStateQueued)
	seedCard(t, store, "t952", "notes not landed", kanban.BacklogStateQueued)
	commitOnRef(t, root, "fix(t951): attempt (not merged)")
	commitOnRef(t, root, "docs(t952): notes (NOT landed)")
	materializeOriginDevelop(t, root)

	stdout, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("auto-done: %v", err)
	}
	for _, id := range []string{"t951", "t952"} {
		if !strings.Contains(stdout, "skip "+id+" reason=not-landed") {
			t.Errorf("stdout %q lacks skip %s reason=not-landed", stdout, id)
		}
		if _, ok := liveItemOK(t, store, id); !ok {
			t.Errorf("%s was archived on a negated subject", id)
		}
	}
	if strings.Contains(stdout, "done t951") || strings.Contains(stdout, "done t952") {
		t.Errorf("stdout %q closes a negated card", stdout)
	}
}

// AC-AD-008 — an inconclusive evaluation never closes: the ALWAYS-emitted
// skip line is pinned, the exit code is 0, and no close line appears.
func TestTodoAutoDone_InconclusiveNeverCloses(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t961", "eligible on every other axis", kanban.BacklogStateQueued)
	// The configured ref names origin/develop, but the fixture never
	// materializes it — the landed ref is unresolvable.
	commitOnRef(t, root, "fix(t961): work the unresolvable ref cannot see")

	stdout, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("auto-done exited %v — an inconclusive CARD is a skip outcome, not a command failure", err)
	}
	if !strings.Contains(stdout, "skip t961 reason=query-inconclusive") {
		t.Errorf("stdout %q lacks skip t961 reason=query-inconclusive", stdout)
	}
	if strings.Contains(stdout, "done t961") {
		t.Errorf("stdout %q closes a card the query could not answer", stdout)
	}
	if _, ok := liveItemOK(t, store, "t961"); !ok {
		t.Error("t961 archived on an inconclusive query")
	}
}

// AC-AD-009 — the close-line contract: canonical prefix, source marker,
// one skip line per skipped card, summary last, and factory state recorded
// for the closed card.
func TestTodoAutoDone_CloseLineContract(t *testing.T) {
	t.Setenv(config.EnvMoaiKanbanID, "run-1")
	t.Setenv(config.EnvMoaiFactoryWorkers, "2")
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t971", "closes", kanban.BacklogStateQueued)
	seedCard(t, store, "t972", "skips", kanban.BacklogStateQueued)
	commitOnRef(t, root, "fix(t971): landed work")
	materializeOriginDevelop(t, root)

	stdout, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("auto-done: %v", err)
	}
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) == 0 {
		t.Fatal("stdout is empty")
	}
	if !strings.HasPrefix(lines[len(lines)-1], "auto-done scanned=") {
		t.Errorf("last line %q is not the summary", lines[len(lines)-1])
	}
	if !strings.Contains(stdout, "done t971 landing=landed source=auto-land") {
		t.Errorf("stdout %q lacks the canonical close prefix + source marker", stdout)
	}
	if !strings.Contains(stdout, "skip t972 reason=") {
		t.Errorf("stdout %q lacks a skip line for t972", stdout)
	}
	// Factory state observable as for a manual done.
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	found := false
	for _, a := range rec.Runtime.Assignments {
		if a.CardID == "t971" && a.EventKind == "card.completed" && a.ReportedState == "completed" {
			found = true
		}
	}
	if !found {
		t.Errorf("no card.completed factory event for t971 in %+v", rec.Runtime.Assignments)
	}
}

// AC-AD-010 — the execution log: valid JSONL, one row per closed card plus
// skip rows, each carrying the required facts.
func TestTodoAutoDone_ExecutionLog(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t981", "closes on subject", kanban.BacklogStateQueued)
	seedCard(t, store, "t982", "skips", kanban.BacklogStateQueued)
	subject := "Merge branch 'WT-f' into develop (card t981)"
	sha := commitOnRef(t, root, subject)
	materializeOriginDevelop(t, root)

	if _, _, err := runTodo(t, "auto-done"); err != nil {
		t.Fatalf("auto-done: %v", err)
	}
	path := scanLogPath(t)
	rows := readLogRows(t, path)
	if len(rows) < 2 {
		t.Fatalf("log holds %d rows, want >= 2 (one closed + one skipped)", len(rows))
	}
	closed := lastLogRowFor(t, path, "t981", "closed")
	for _, key := range []string{"card_id", "at", "form", "subject", "commit_sha", "ref", "ref_head", "source"} {
		if v, ok := closed[key]; !ok || v == "" {
			t.Errorf("closed row missing %s: %v", key, closed)
		}
	}
	if closed["subject"] != subject || closed["commit_sha"] != sha {
		t.Errorf("closed row subject/sha = %v/%v, want %s/%s", closed["subject"], closed["commit_sha"], subject, sha)
	}
	at, _ := closed["at"].(string)
	if !strings.HasSuffix(at, "Z") {
		t.Errorf("log instant %q is not RFC 3339 UTC", at)
	}
	if _, err := time.Parse(time.RFC3339, at); err != nil {
		t.Errorf("log instant %q does not parse as RFC 3339: %v", at, err)
	}
	if closed["ref"] != "origin/develop" {
		t.Errorf("log ref = %v, want origin/develop", closed["ref"])
	}
	if closed["ref_head"] == "" {
		t.Error("log row carries no ref head — the scan-time ref position is a mandatory fact")
	}
	skipped := lastLogRowFor(t, path, "t982", "skipped")
	if skipped["reason"] == "" {
		t.Errorf("skipped row carries no reason: %v", skipped)
	}
}

// AC-AD-011 — reversibility: undone restores a scan-closed card, the log
// gains a reversal row naming the original closure, and a second undone is
// refused.
func TestTodoAutoDone_ReversalRow(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t991", "reversible work", kanban.BacklogStateQueued)
	commitOnRef(t, root, "fix(t991): landed work")
	materializeOriginDevelop(t, root)

	if _, _, err := runTodo(t, "auto-done"); err != nil {
		t.Fatalf("auto-done: %v", err)
	}
	path := scanLogPath(t)
	closed := lastLogRowFor(t, path, "t991", "closed")

	if _, _, err := runTodo(t, "undone", "t991"); err != nil {
		t.Fatalf("undone: %v", err)
	}
	if _, ok := liveItemOK(t, store, "t991"); !ok {
		t.Fatal("t991 did not return to the live queue")
	}
	reversal := lastLogRowFor(t, path, "t991", "reversed")
	if reversal["original_at"] != closed["at"] {
		t.Errorf("reversal row original_at = %v, want the closure row's at %v", reversal["original_at"], closed["at"])
	}
	// A second undone is refused — the archive entry is gone.
	if _, _, err := runTodo(t, "undone", "t991"); err == nil {
		t.Fatal("second undone succeeded — the refusal is gone")
	} else if !strings.Contains(err.Error(), "no archived backlog item") {
		t.Errorf("second undone error = %v, want the no-archived-item refusal", err)
	}
}

// AC-AD-012 — dry-run: whole-record byte identity while stdout carries
// exactly the same close and skip lines the non-dry run produces.
func TestTodoAutoDone_DryRunByteIdentity(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t995", "would close", kanban.BacklogStateQueued)
	seedCard(t, store, "t996", "would skip", kanban.BacklogStateQueued)
	commitOnRef(t, root, "fix(t995): landed work")
	materializeOriginDevelop(t, root)

	before := append([]byte(nil), recordBytes(t, store)...)
	dryStdout, _, err := runTodo(t, "auto-done", "--dry-run")
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	afterDry := recordBytes(t, store)
	if !bytes.Equal(before, afterDry) {
		t.Error("the queue record changed bytes under --dry-run")
	}

	liveStdout, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("live run: %v", err)
	}
	if dryStdout != liveStdout {
		t.Errorf("dry-run stdout differs from the live run:\n--dry-run:\n%s\nlive:\n%s", dryStdout, liveStdout)
	}
	if !strings.Contains(dryStdout, "done t995 landing=landed") || !strings.Contains(dryStdout, "skip t996 reason=") {
		t.Errorf("dry-run stdout %q lacks the decision lines", dryStdout)
	}
}

// AC-AD-013 — idempotence: the second scan closes nothing and reports zero.
func TestTodoAutoDone_Idempotence(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t997", "closes once", kanban.BacklogStateQueued)
	commitOnRef(t, root, "fix(t997): landed work")
	materializeOriginDevelop(t, root)

	if _, _, err := runTodo(t, "auto-done"); err != nil {
		t.Fatalf("first scan: %v", err)
	}
	stdout, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("second scan: %v", err)
	}
	if strings.Contains(stdout, "done t997") {
		t.Errorf("second scan re-closed t997: %q", stdout)
	}
	if !strings.Contains(stdout, "closed=0") {
		t.Errorf("second scan summary %q does not report zero closes", stdout)
	}
	if _, ok := liveItemOK(t, store, "t997"); ok {
		t.Error("t997 live after the first scan closed it")
	}
}

// AC-AD-014 — no landing-column writes: no item's Landing field changed,
// and the provenance value set stays closed at operator.
func TestTodoAutoDone_NoLandingColumnWrites(t *testing.T) {
	root, store := autoDoneFixture(t)
	seedCard(t, store, "t998", "closes", kanban.BacklogStateQueued)
	seedCard(t, store, "t999", "skips", kanban.BacklogStateQueued)
	sha := commitOnRef(t, root, "fix(t998): landed work")
	commitOnRef(t, root, "fix(t999): landed work too")
	materializeOriginDevelop(t, root)
	// One card carries recorded evidence — the scan must preserve it, not
	// rewrite or clear it.
	if _, _, err := runTodo(t, "landed", "t999", "--sha", sha); err != nil {
		t.Fatalf("record sha: %v", err)
	}
	before, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := runTodo(t, "auto-done"); err != nil {
		t.Fatalf("auto-done: %v", err)
	}
	after, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, was := range before.Items {
		for _, is := range after.Items {
			if is.ID != was.ID {
				continue
			}
			if (was.Landing == nil) != (is.Landing == nil) {
				t.Errorf("card %s Landing presence changed: before=%v after=%v", is.ID, was.Landing, is.Landing)
			} else if was.Landing != nil && *was.Landing != *is.Landing {
				t.Errorf("card %s Landing changed: before=%+v after=%+v", is.ID, *was.Landing, *is.Landing)
			}
		}
	}
}

// AC-AD-015 — the fetch boundary: WITH --fetch exactly one git fetch runs
// and the scan evaluates the post-fetch ref; WITHOUT it, zero fetches.
func TestTodoAutoDone_FetchBoundary(t *testing.T) {
	t.Run("with --fetch runs exactly one fetch", func(t *testing.T) {
		root, store := autoDoneFixture(t)
		seedCard(t, store, "t9981", "fetch-scoped work", kanban.BacklogStateQueued)
		commitOnRef(t, root, "fix(t9981): landed work")
		materializeOriginDevelop(t, root)
		count := installFetchCounter(t)

		if _, _, err := runTodo(t, "auto-done", "--fetch"); err != nil {
			t.Fatalf("auto-done --fetch: %v", err)
		}
		if *count != 1 {
			t.Errorf("fetch invocations = %d, want exactly 1 for one scan", *count)
		}
		if _, ok := liveItemOK(t, store, "t9981"); ok {
			t.Error("the --fetch scan did not close the landed card")
		}
	})

	t.Run("without --fetch runs zero fetches", func(t *testing.T) {
		root, store := autoDoneFixture(t)
		seedCard(t, store, "t9982", "offline-scoped work", kanban.BacklogStateQueued)
		commitOnRef(t, root, "fix(t9982): landed work")
		materializeOriginDevelop(t, root)
		count := installFetchCounter(t)

		if _, _, err := runTodo(t, "auto-done"); err != nil {
			t.Fatalf("auto-done: %v", err)
		}
		if *count != 0 {
			t.Errorf("fetch invocations = %d, want 0 — the scan performs no network I/O without --fetch", *count)
		}
		if _, ok := liveItemOK(t, store, "t9982"); ok {
			t.Error("the offline scan did not close the landed card — the local remote-tracking ref must answer")
		}
	})
}

// AC-AD-016 — close-surface exclusivity: only `done` and `auto-done` reach
// the archive transition, structurally asserted with the allowlist named;
// and a `todo landed` control run transitions nothing.
func TestTodoAutoDone_CloseSurfaceExclusivity(t *testing.T) {
	t.Run("structural allowlist", func(t *testing.T) {
		allowed := map[string]bool{"done": true, "auto-done": true}
		files, err := filepath.Glob("todo*.go")
		if err != nil {
			t.Fatalf("glob: %v", err)
		}
		hits := map[string]int{}
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			raw, err := os.ReadFile(f)
			if err != nil {
				t.Fatalf("read %s: %v", f, err)
			}
			if n := strings.Count(string(raw), ".ArchiveCard("); n > 0 {
				hits[f] = n
			}
		}
		// The store's own definition site is internal/kanban, outside this
		// glob; every CALL SITE in the CLI surface must sit in a file the
		// allowlist's verb owns.
		verbOfFile := map[string]string{
			"todo.go":          "done",
			"todo_autodone.go": "auto-done",
		}
		for f := range hits {
			if _, ok := verbOfFile[f]; !ok {
				t.Errorf("file %s reaches .ArchiveCard( and is owned by no allowlisted verb (%v) — a third close surface", f, allowed)
			}
		}
		if len(hits) != 2 {
			t.Errorf("ArchiveCard call sites = %v, want exactly the two close surfaces", hits)
		}
	})

	t.Run("todo landed transitions nothing", func(t *testing.T) {
		root, store := autoDoneFixture(t)
		seedCard(t, store, "t9983", "control card", kanban.BacklogStateQueued)
		sha := commitOnRef(t, root, "fix(t9983): landing-eligible work")
		materializeOriginDevelop(t, root)
		beforeRec, err := store.LoadPure()
		if err != nil {
			t.Fatal(err)
		}

		if _, _, err := runTodo(t, "landed", "t9983", "--sha", sha); err != nil {
			t.Fatalf("landed: %v", err)
		}
		after, err := store.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		if len(after.Items) != len(beforeRec.Items) {
			t.Fatalf("live count changed: %d -> %d — recording evidence transitioned a card", len(beforeRec.Items), len(after.Items))
		}
		item := liveItem(t, store, "t9983")
		if item.State != kanban.BacklogStateQueued || item.Text != "control card" {
			t.Errorf("control card changed: %+v", item)
		}
		if len(after.Archived) != 0 {
			t.Errorf("archive gained %d entries — recording evidence added one", len(after.Archived))
		}
		if item.Landing == nil || item.Landing.SHA != sha {
			t.Errorf("the landing field was not written: %+v", item.Landing)
		}
	})
}

// AC-AD-017 — the observed false-negative shapes attribute and close, while
// the reissued-id collision gates hold in the same fixture pair.
func TestTodoAutoDone_FalseNegativeShapes(t *testing.T) {
	root, store := autoDoneFixture(t)

	// Shape A (t603, landing commit d8b7836aa) — the comma-form trailing
	// parenthetical: the card id opens the group, a non-card token follows
	// the comma.
	seedCard(t, store, "t6030", "comma-form card", kanban.BacklogStateQueued)
	commitOnRef(t, root, "fix(hooks): sync-phase 게이트의 C++ 검사 복구 (t6030, H08)")

	// Shape B (t681, merge ae980ef2d / repair 4fb28a5c6) — the merge subject
	// mentions the id in a NON-attributing position (mid-subject, no
	// trailing group), the actual repair commit's subject carries NO id, and
	// the card closes only on its recorded Landing.SHA.
	seedCard(t, store, "t6031", "sha-only card", kanban.BacklogStateQueued)
	commitOnRef(t, root, "Merge branch 'WT-shape-b' (card t6031 work) into develop")
	shapeBSha := commitOnRef(t, root, "chore: the actual repair, no id in sight")
	materializeOriginDevelop(t, root)
	if _, _, err := runTodo(t, "landed", "t6031", "--sha", shapeBSha); err != nil {
		t.Fatalf("record shape-B sha: %v", err)
	}

	stdout, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("auto-done: %v", err)
	}
	if !strings.Contains(stdout, "done t6030 landing=landed") || !strings.Contains(stdout, "form=subject-attribution") {
		t.Errorf("Shape A: stdout %q lacks the comma-form attribution close", stdout)
	}
	if !strings.Contains(stdout, "done t6031 landing=landed") || !strings.Contains(stdout, "form=sha-recorded") {
		t.Errorf("Shape B: stdout %q lacks the recorded-SHA close", stdout)
	}

	// The contrast, judged by the SAME gates: attributing the comma form did
	// not loosen the collision gate — a reissued id on subject evidence
	// alone still skips.
	seedCard(t, store, "t6032", "reissued text B", kanban.BacklogStateQueued)
	commitOnRef(t, root, "fix(t6032): some landed-looking work")
	injectArchivedPredecessor(t, store, "t6032", "original text A")
	materializeOriginDevelop(t, root)
	stdout2, _, err := runTodo(t, "auto-done")
	if err != nil {
		t.Fatalf("auto-done contrast: %v", err)
	}
	if !strings.Contains(stdout2, "skip t6032 reason=ambiguous-id") {
		t.Errorf("contrast: stdout %q lacks skip t6032 reason=ambiguous-id — the comma form loosened the gate", stdout2)
	}
}

// DoD — --help documents the two evidence forms, the canonical four-token
// skip set, the exit-code policy, and the dry-run contract.
func TestTodoAutoDone_HelpDocumentsContract(t *testing.T) {
	autoDoneFixture(t)
	stdout, _, err := runTodo(t, "auto-done", "--help")
	if err != nil {
		t.Fatalf("help: %v", err)
	}
	for _, token := range kanban.AutoDoneSkipReasons() {
		if !strings.Contains(stdout, token) {
			t.Errorf("--help does not document skip token %q", token)
		}
	}
	for _, want := range []string{"sha-recorded", "subject-attribution", "dry-run", "exit"} {
		if !strings.Contains(strings.ToLower(stdout), want) {
			t.Errorf("--help does not document %q", want)
		}
	}
}
