// todo_composed_upgrade_test.go — SPEC-QUEUE-UPGRADE-PROOF-001 (card t470):
// the composed v3.1.2-to-next queue upgrade, entered through the `moai todo`
// command path rather than through the BacklogStore API.
//
// v3.1.2 ships the queue as a JSON document in `.moai/state/kanban/`; local
// develop ships it as SQLite in `.moai/state/todo/`. No release tag carries
// the SQLite merge, so the next release asks every existing user to cross TWO
// transitions in one step: a state-directory relocation (state_dir.go) and a
// JSON-to-SQLite conversion (backlog_migrate.go).
//
// Each transition is separately well covered. Their COMPOSITION is not: every
// existing test plants its layout by hand and calls the API directly, and
// `grep -rn "LegacyStateDirForRoot" internal/cli/` finds no test seed. This
// file is that missing cell — one `moai todo` invocation against a genuine
// v3.1.2 layout, firing the relocation and then the conversion in sequence.
//
// This is a PROOF, not a repair: it changes no production behavior
// (REQ-QUP-009). A failing assertion here is a defect discovery to report,
// never something to fix by editing the mechanism under the proof.
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// f1LegacyBacklogJSON is fixture F1: the record shape a v3.1.2 binary
// actually writes. That release's BacklogRecord carries exactly three fields
// — Version, LastSeq, Items (git show v3.1.2:internal/kanban/backlog_store.go)
// — so this literal carries `version`, `last_seq`, `items` AND NOTHING ELSE
// (REQ-QUP-006 / AC-QUP-006). `findings` and `archived` exist only on
// develop; a fixture carrying them would describe a state no real upgrading
// user can be in, and the proof would prove a fiction.
//
// The items span the three BacklogState values that exist on both sides, and
// last_seq (7) is strictly above the highest item id (5), so the high-water
// mark crossing the composition is observable rather than coincidental.
const f1LegacyBacklogJSON = `{"version":1,"last_seq":7,"items":[` +
	`{"id":"t2","text":"legacy queued one","added_at":"2026-08-14T00:00:00Z","spec_id":null,"state":"queued"},` +
	`{"id":"t3","text":"legacy picked one","added_at":"2026-08-14T00:01:00Z","spec_id":"SPEC-LEGACY-001","state":"picked"},` +
	`{"id":"t5","text":"legacy dropped one","added_at":"2026-08-14T00:02:00Z","spec_id":null,"state":"dropped"}]}`

// f1SentinelJSON is the relocation sentinel: a per-session registry file
// seeded INSIDE the legacy directory. state_dir.go:13-16 records that the
// relocation moves the DIRECTORY, not a file list, so the registry files ride
// along by construction. Finding this file, byte-identical, under the current
// directory name is the POSITIVE evidence that the legacy directory's
// contents arrived rather than merely vanishing (AC-QUP-002).
const f1SentinelJSON = `{"companions":["lane-1"],"sentinel":"t470"}`

// seedLegacyV312Layout materializes the v3.1.2 on-disk layout under root: the
// legacy state directory holding backlogJSON plus the relocation sentinel. It
// returns the legacy and current state-directory paths.
//
// preCreateCurrentDir is the AC-QUP-010 mutation seam, and it is OFF on every
// committed call. Turning it on additionally creates an EMPTY current-name
// directory, which makes resolveStateDir take the stale-copy branch
// (state_dir.go:81-88) and return the current directory unconditionally: the
// relocation branch (state_dir.go:100-103) is never reached, the resolved
// queue path never names the seeded document, and the migration
// (backlog_store.go:603-607, state B) never sees it. That is what a severed
// composed path looks like, and it is how RED is established before GREEN is
// claimed.
func seedLegacyV312Layout(t *testing.T, root, backlogJSON string, preCreateCurrentDir bool) (legacyDir, currentDir string) {
	t.Helper()

	legacyDir = kanban.LegacyStateDirForRoot(root)
	currentDir = kanban.StateDirForRoot(root)

	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatalf("mkdir legacy state dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, "backlog.json"), []byte(backlogJSON), 0o600); err != nil {
		t.Fatalf("seed legacy backlog.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, "companions.json"), []byte(f1SentinelJSON), 0o600); err != nil {
		t.Fatalf("seed relocation sentinel: %v", err)
	}
	if preCreateCurrentDir {
		if err := os.MkdirAll(currentDir, 0o755); err != nil {
			t.Fatalf("mutation: mkdir current state dir: %v", err)
		}
	}
	return legacyDir, currentDir
}

// assertPreUpgradeState pins the preconditions AC-QUP-002 depends on. They are
// load-bearing, not ceremony: the criterion's final limb is a
// negative-existence assertion ("the legacy directory no longer exists"),
// which passes vacuously when the directory was never created. Asserting its
// presence — and the current name's absence — before any command runs is what
// makes that limb falsifiable.
//
// currentDirPreCreated is the mutation seam's other half, and is false on
// every committed call. Under the AC-QUP-010 mutation the current-name
// directory exists BY CONSTRUCTION, so asserting its absence would stop the
// run at the precondition and hide the symptom the criterion names — the
// legacy directory still standing AFTER the command. Relaxing exactly that
// one limb, and only when the mutation is active, is what lets RED land where
// the criterion says it should.
func assertPreUpgradeState(t *testing.T, legacyDir, currentDir string, currentDirPreCreated bool) {
	t.Helper()

	if _, err := os.Stat(filepath.Join(legacyDir, "backlog.json")); err != nil {
		t.Fatalf("precondition: legacy backlog.json must exist before the upgrade: %v", err)
	}
	if _, err := os.Stat(filepath.Join(legacyDir, "companions.json")); err != nil {
		t.Fatalf("precondition: relocation sentinel must exist before the upgrade: %v", err)
	}
	if currentDirPreCreated {
		return
	}
	if _, err := os.Stat(currentDir); !os.IsNotExist(err) {
		t.Fatalf("precondition: current state dir %q must NOT exist before the upgrade (stat err = %v)", currentDir, err)
	}
}

// composedUpgradeFixture prepares an isolated project root for the composed
// upgrade and returns it (REQ-QUP-008 / AC-QUP-008).
//
// todoFixture supplies the isolation the command path requires: a t.TempDir()
// root that is a COMMITTED git repository, reached through CLAUDE_PROJECT_DIR.
// The git repository is not optional — resolveTodoQueueRoot resolves through
// git, so a bare t.TempDir() would fall through to the home-based fallback and
// the fixture planted in the temp project would never be read.
//
// userHomeDirFn is additionally redirected to a second temp directory, so even
// that fallback branch cannot reach the operator's real home.
func composedUpgradeFixture(t *testing.T) string {
	t.Helper()

	home := t.TempDir()
	orig := userHomeDirFn
	userHomeDirFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDirFn = orig })

	root, _ := todoFixture(t)
	return root
}

// assertQueueRootIsolated asserts the resolved queue root is the temp fixture
// root and lies under the OS temp tree — checked BEFORE any todo command runs,
// so a resolution that escaped to the live repository fails the test instead of
// mutating the operator's real backlog.
func assertQueueRootIsolated(t *testing.T, root string) {
	t.Helper()

	resolved := resolveTodoQueueRoot()
	if !sameDir(resolved, root) {
		t.Fatalf("resolved queue root = %q, want the temp fixture root %q", resolved, root)
	}
	if !queueRootInsideTemp(resolved) {
		t.Fatalf("resolved queue root %q is not under the OS temp tree", resolved)
	}
}

// assertComposedUpgradeSideEffects measures the composition's own evidence
// after the first `moai todo` command: AC-QUP-002 (the directory relocated and
// its contents came with it), AC-QUP-003 (the legacy document quarantined
// rather than destroyed), and AC-QUP-004 (the SQLite artifact present).
//
// Every limb is t.Errorf rather than t.Fatalf, so one severed step reports
// every symptom it caused instead of hiding the rest behind the first.
func assertComposedUpgradeSideEffects(t *testing.T, legacyDir, currentDir string) {
	t.Helper()

	// AC-QUP-002 — the directory relocated, and its contents came with it.
	if _, err := os.Stat(currentDir); err != nil {
		t.Errorf("current state dir %q must exist after the upgrade: %v", currentDir, err)
	}
	dbPath := filepath.Join(currentDir, "backlog.db")
	if info, err := os.Stat(dbPath); err != nil {
		t.Errorf("current state dir must hold the queue at %q: %v", dbPath, err)
	} else if info.Size() == 0 {
		t.Errorf("queue database %q is empty", dbPath)
	}
	sentinel, err := os.ReadFile(filepath.Join(currentDir, "companions.json"))
	if err != nil {
		t.Errorf("relocation sentinel must be readable under the current name: %v", err)
	} else if string(sentinel) != f1SentinelJSON {
		t.Errorf("relocation sentinel bytes = %q, want the seeded %q", sentinel, f1SentinelJSON)
	}
	if _, err := os.Stat(legacyDir); !os.IsNotExist(err) {
		t.Errorf("legacy state dir %q must no longer exist after the upgrade (stat err = %v)", legacyDir, err)
	}

	// AC-QUP-003 — the legacy document is quarantined, not destroyed.
	quarantine := filepath.Join(currentDir, "backlog.json.migrated")
	if got, err := os.ReadFile(quarantine); err != nil {
		t.Errorf("quarantined legacy document %q must exist: %v", quarantine, err)
	} else if string(got) != f1LegacyBacklogJSON {
		t.Errorf("quarantined document bytes diverge from the seeded fixture:\n got %q\nwant %q", got, f1LegacyBacklogJSON)
	}
	if _, err := os.Stat(filepath.Join(currentDir, "backlog.json")); !os.IsNotExist(err) {
		t.Errorf("no backlog.json may remain beside the quarantine (stat err = %v)", err)
	}

	// AC-QUP-004 — the SQLite artifact exists and is non-empty.
	if info, err := os.Stat(dbPath); err != nil {
		t.Errorf("SQLite artifact %q must exist after the upgrade: %v", dbPath, err)
	} else if info.Size() == 0 {
		t.Errorf("SQLite artifact %q is empty", dbPath)
	}
}

// TestTodoComposedUpgrade_FromLegacyV312Layout is the composed-path proof
// (M1 / G1). One `moai todo` invocation against the F1 layout must fire the
// directory relocation and then the JSON-to-SQLite conversion, and the queue
// must come through both intact.
//
// Criteria measured here: AC-QUP-001a (cards, states, order), AC-QUP-001b
// (last_seq high-water mark), AC-QUP-002 (relocation + sentinel), AC-QUP-003
// (legacy document quarantined, not destroyed), AC-QUP-004 (SQLite artifact).
func TestTodoComposedUpgrade_FromLegacyV312Layout(t *testing.T) {
	root := composedUpgradeFixture(t)
	legacyDir, currentDir := seedLegacyV312Layout(t, root, f1LegacyBacklogJSON, false)

	assertPreUpgradeState(t, legacyDir, currentDir, false)
	assertQueueRootIsolated(t, root)

	// The composed path's single entrance: the first `moai todo` command.
	out, stderr, err := runTodo(t, "list", "--json")
	if err != nil {
		t.Fatalf("first todo command against the legacy layout: %v\nstderr: %s", err, stderr)
	}

	// The side effects come first, because they are the composition's own
	// evidence: the relocation and the conversion either happened or they did
	// not, and a queue-content assertion failing ahead of them would report a
	// symptom while leaving the named criterion unevaluated.
	assertComposedUpgradeSideEffects(t, legacyDir, currentDir)

	// AC-QUP-001a — every card, its state, and the seeded order survive.
	var rec kanban.BacklogRecord
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &rec); err != nil {
		t.Fatalf("decode `todo list --json` output %q: %v", out, err)
	}
	wantIDs := []string{"t2", "t3", "t5"}
	wantStates := []kanban.BacklogState{
		kanban.BacklogStateQueued, kanban.BacklogStatePicked, kanban.BacklogStateDropped,
	}
	wantTexts := []string{"legacy queued one", "legacy picked one", "legacy dropped one"}
	if len(rec.Items) != len(wantIDs) {
		t.Fatalf("composed upgrade yielded %d items, want %d: %+v", len(rec.Items), len(wantIDs), rec.Items)
	}
	for i, item := range rec.Items {
		if item.ID != wantIDs[i] || item.State != wantStates[i] || item.Text != wantTexts[i] {
			t.Errorf("item[%d] = {id %q, state %q, text %q}, want {id %q, state %q, text %q}",
				i, item.ID, item.State, item.Text, wantIDs[i], wantStates[i], wantTexts[i])
		}
	}
	if rec.Items[1].SpecID == nil || *rec.Items[1].SpecID != "SPEC-LEGACY-001" {
		t.Errorf("picked card's spec_id did not survive: %v", rec.Items[1].SpecID)
	}
	if rec.LastSeq != 7 {
		t.Errorf("last_seq after the composed upgrade = %d, want the seeded 7", rec.LastSeq)
	}

	// AC-QUP-001b — the high-water mark crossed the composition rather than
	// being re-derived from the items: the next id is last_seq+1, not max+1.
	addOut, addErr, err := runTodo(t, "add", "post-upgrade card")
	if err != nil {
		t.Fatalf("add after the composed upgrade: %v\nstderr: %s", err, addErr)
	}
	if got := strings.Fields(strings.TrimSpace(addOut)); len(got) == 0 || got[0] != "t8" {
		t.Errorf("post-upgrade add issued %q, want id t8 (seeded last_seq 7 + 1); "+
			"t6 would mean the mark was re-derived from the items", strings.TrimSpace(addOut))
	}

}

// f2ForwardCompatibleJSON is fixture F2. REACHABILITY CAVEAT: this record
// shape is NOT what a v3.1.2 user's project can hold — that release's
// BacklogRecord has no `findings` and no `archived` fields. Only someone who
// ran a DEVELOPMENT build before the release, and whose state directory was
// still on the legacy name, can be in this state. It is a forward-compatibility
// check, never a user-facing upgrade scenario.
const f2ForwardCompatibleJSON = `{"version":1,"last_seq":4,"items":[` +
	`{"id":"t1","text":"live card","added_at":"2026-08-14T00:00:00Z","spec_id":null,"state":"queued"}],` +
	`"findings":[{"subject_id":"t1","related_id":"t2","relation":"near-duplicate","source":"mechanical","score":0.91,"note":"","at":"2026-08-14T00:03:00Z"}],` +
	`"archived":[{"item":{"id":"t2","text":"archived card","added_at":"2026-08-14T00:01:00Z","spec_id":null,"state":"dropped"},"position":2,"findings":[]}]}`

// TestTodoComposedUpgrade_ForwardCompatibleFieldsSurvive is the OPTIONAL
// AC-QUP-007 criterion (M2): the two additive top-level fields survive the
// composed upgrade from the LEGACY directory. See f2ForwardCompatibleJSON for
// why this state is not reachable for a real v3.1.2 user.
func TestTodoComposedUpgrade_ForwardCompatibleFieldsSurvive(t *testing.T) {
	root := composedUpgradeFixture(t)
	legacyDir, currentDir := seedLegacyV312Layout(t, root, f2ForwardCompatibleJSON, false)

	assertPreUpgradeState(t, legacyDir, currentDir, false)
	assertQueueRootIsolated(t, root)

	out, stderr, err := runTodo(t, "list", "--json")
	if err != nil {
		t.Fatalf("first todo command against the legacy layout: %v\nstderr: %s", err, stderr)
	}

	var rec kanban.BacklogRecord
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &rec); err != nil {
		t.Fatalf("decode `todo list --json` output %q: %v", out, err)
	}
	if len(rec.Findings) != 1 {
		t.Fatalf("findings after the composed upgrade = %d, want 1: %+v", len(rec.Findings), rec.Findings)
	}
	if rec.Findings[0].SubjectID != "t1" || rec.Findings[0].RelatedID != "t2" ||
		rec.Findings[0].Relation != kanban.BacklogRelationNearDuplicate {
		t.Errorf("finding did not survive intact: %+v", rec.Findings[0])
	}
	if len(rec.Archived) != 1 {
		t.Fatalf("archived entries after the composed upgrade = %d, want 1: %+v", len(rec.Archived), rec.Archived)
	}
	if rec.Archived[0].Item.ID != "t2" || rec.Archived[0].Item.Text != "archived card" {
		t.Errorf("archived entry did not survive intact: %+v", rec.Archived[0])
	}
}
