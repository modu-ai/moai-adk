package kanban

// settings_drift_resolution_test.go — card t765.
//
// The gate records detections and never records resolutions, so a reader of
// the ledger sees every detection as still open, forever. The measured control
// run: clean tree -> no ledger; drift -> one row; restore the file by hand ->
// re-check reports "clean (0 matches)" and the ledger STILL holds one row and
// nothing else.
//
// The assertions below are about what a later READER can conclude. They read
// the ledger as untyped JSON rather than through a struct, because the
// load-bearing property of a legacy row is the ABSENCE of a field: unmarshalled
// into a struct, an absent `status` and a present `"status":""` are the same
// value, and the distinction this card turns on would be invisible.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// rawLedger reads the ledger as one map per line, preserving field presence.
func rawLedger(t *testing.T, root string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(SettingsDriftDir(root), SettingsDriftLedgerName))
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	var rows []map[string]any
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatalf("ledger line %q is not JSON: %v", line, err)
		}
		rows = append(rows, row)
	}
	return rows
}

// restoreFixtureSettings puts the committed content back, which is what a human
// does with `git checkout -- .claude/settings.json` after reporting a hit.
func restoreFixtureSettings(t *testing.T, worktree string) {
	t.Helper()
	gitFixture(t, worktree, "checkout", "--", SettingsDriftWatchedPath)
	if out := gitFixture(t, worktree, "--no-optional-locks", "status", "--porcelain", "--", SettingsDriftWatchedPath); strings.TrimSpace(out) != "" {
		t.Fatalf("fixture: the tree is still dirty after restore: %q", out)
	}
}

// TestLedgerRecordsAResolutionAfterADetection is the card's whole point: a
// clean assessment of a tree whose most recent row is an unresolved detection
// appends ONE resolution row, and that row carries enough to join it to the
// detection it closes.
func TestLedgerRecordsAResolutionAfterADetection(t *testing.T) {
	t.Parallel()
	worktree, root, _ := driftFixture(t)

	hit := AssessSettingsDrift(SettingsDriftParams{
		Dir: worktree, Root: root, Card: "t765", Branch: "WT-drift-ledger", Runner: NewExecRunner(),
	})
	if hit.Status != SettingsDriftDetected {
		t.Fatalf("fixture: first assessment is %q, want %q", hit.Status, SettingsDriftDetected)
	}
	preservedAfterHit := preservedFiles(t, root)
	if len(preservedAfterHit) != 1 {
		t.Fatalf("fixture: %d preserved copies after the hit, want 1", len(preservedAfterHit))
	}

	restoreFixtureSettings(t, worktree)

	clean := AssessSettingsDrift(SettingsDriftParams{
		Dir: worktree, Root: root, Card: "t765", Branch: "WT-drift-ledger", Runner: NewExecRunner(),
	})
	if clean.Status != SettingsDriftClean {
		t.Fatalf("second assessment: got %q, want %q (err=%v)", clean.Status, SettingsDriftClean, clean.Err)
	}
	if clean.PreserveErr != nil {
		t.Errorf("second assessment: preserve error %v, want none", clean.PreserveErr)
	}

	rows := rawLedger(t, root)
	if len(rows) != 2 {
		t.Fatalf("ledger holds %d rows, want 2 (one detection, one resolution)", len(rows))
	}
	detection, resolution := rows[0], rows[1]

	if got := detection["status"]; got != "detected" {
		t.Errorf("detection row status: got %v, want %q", got, "detected")
	}
	if got := resolution["status"]; got != "resolved" {
		t.Errorf("resolution row status: got %v, want %q", got, "resolved")
	}
	if got := resolution["match_count"]; got != float64(0) {
		t.Errorf("resolution row match_count: got %v, want 0", got)
	}
	if got := resolution["worktree"]; got != detection["worktree"] {
		t.Errorf("resolution row worktree: got %v, want %v", got, detection["worktree"])
	}

	// The join: the resolution names the detection it closes, by the two
	// values that identify it in this file — its digest and its timestamp.
	if got := resolution["resolves_sha256"]; got != detection["sha256"] {
		t.Errorf("resolves_sha256: got %v, want %v (the detection's sha256)", got, detection["sha256"])
	}
	if got := resolution["resolves_measured_at"]; got != detection["measured_at"] {
		t.Errorf("resolves_measured_at: got %v, want %v (the detection's measured_at)", got, detection["measured_at"])
	}

	// A resolution measures nothing about the file, so it reports nothing
	// about it. A `sha256` on this row would read as a digest of the restored
	// content, which was never hashed.
	if _, present := resolution["sha256"]; present {
		t.Errorf("resolution row carries sha256=%v; it hashed nothing", resolution["sha256"])
	}
	if _, present := resolution["preserved_path"]; present {
		t.Errorf("resolution row carries preserved_path=%v; it preserved nothing", resolution["preserved_path"])
	}

	// Nothing was preserved, and — REQ-PSD-006 — nothing was restored by the
	// gate: the detection's preserved copy is untouched and still alone.
	if after := preservedFiles(t, root); len(after) != len(preservedAfterHit) {
		t.Errorf("preserved copies: %d after the resolution, want %d unchanged", len(after), len(preservedAfterHit))
	}
}

// TestLedgerRecordsNoResolutionWithoutADetection — the noise control. Every
// acquire on a healthy tree runs this gate, so a resolution row written for a
// tree that was already clean would bury the detections it sits among.
func TestLedgerRecordsNoResolutionWithoutADetection(t *testing.T) {
	t.Parallel()
	worktree := newFixtureRepo(t)
	root := newFixtureRepoNamed(t, "primary")

	for i := 0; i < 2; i++ {
		result := AssessSettingsDrift(SettingsDriftParams{
			Dir: worktree, Root: root, Card: "t765", Branch: "main", Runner: NewExecRunner(),
		})
		if result.Status != SettingsDriftClean {
			t.Fatalf("assessment %d: got %q, want %q (err=%v)", i, result.Status, SettingsDriftClean, result.Err)
		}
	}
	if _, err := os.Stat(filepath.Join(SettingsDriftDir(root), SettingsDriftLedgerName)); !os.IsNotExist(err) {
		t.Errorf("a ledger exists for a tree that was never dirty (stat err=%v)", err)
	}
}

// TestLedgerRecordsTheResolutionOnce — a resolution closes the detection, so
// the assessments that follow it add nothing. Without this the noise control
// above is defeated one call later: the first clean run would resolve, and
// every subsequent one would resolve the resolution.
func TestLedgerRecordsTheResolutionOnce(t *testing.T) {
	t.Parallel()
	worktree, root, _ := driftFixture(t)

	AssessSettingsDrift(SettingsDriftParams{Dir: worktree, Root: root, Card: "t765", Runner: NewExecRunner()})
	restoreFixtureSettings(t, worktree)
	for i := 0; i < 3; i++ {
		AssessSettingsDrift(SettingsDriftParams{Dir: worktree, Root: root, Card: "t765", Runner: NewExecRunner()})
	}

	rows := rawLedger(t, root)
	if len(rows) != 2 {
		t.Fatalf("ledger holds %d rows after three clean re-checks, want 2", len(rows))
	}
}

// TestResolutionIsScopedToItsOwnWorktree — two lanes share the primary
// checkout's ledger, so a clean lane must not close a dirty lane's detection.
func TestResolutionIsScopedToItsOwnWorktree(t *testing.T) {
	t.Parallel()
	dirty, root, _ := driftFixture(t)
	clean := newFixtureRepoNamed(t, "second-lane")

	hit := AssessSettingsDrift(SettingsDriftParams{Dir: dirty, Root: root, Card: "t765-a", Runner: NewExecRunner()})
	if hit.Status != SettingsDriftDetected {
		t.Fatalf("fixture: got %q, want %q", hit.Status, SettingsDriftDetected)
	}

	other := AssessSettingsDrift(SettingsDriftParams{Dir: clean, Root: root, Card: "t765-b", Runner: NewExecRunner()})
	if other.Status != SettingsDriftClean {
		t.Fatalf("second lane: got %q, want %q (err=%v)", other.Status, SettingsDriftClean, other.Err)
	}

	rows := rawLedger(t, root)
	if len(rows) != 1 {
		t.Fatalf("ledger holds %d rows, want 1 — a clean lane closed another lane's detection", len(rows))
	}
}

// TestALegacyRowWithNoStatusReadsAsADetection — every row written before this
// change carries no `status` field, and all five in the real ledger are
// detections. A reader that treats an absent field as anything else leaves
// them unresolvable forever.
func TestALegacyRowWithNoStatusReadsAsADetection(t *testing.T) {
	t.Parallel()
	worktree := newFixtureRepo(t)
	root := newFixtureRepoNamed(t, "primary")

	dir := SettingsDriftDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("fixture: mkdir %s: %v", dir, err)
	}
	legacy := map[string]any{
		"measured_at":    "2026-09-08T01:02:03.000000Z",
		"card":           "t765-legacy",
		"branch":         "WT-legacy",
		"worktree":       worktree,
		"source_path":    filepath.Join(worktree, filepath.FromSlash(SettingsDriftWatchedPath)),
		"preserved_path": filepath.Join(dir, "settings.json.t765-legacy.20260908T010203.000Z.deadbeef"),
		"sha256":         "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		"size_bytes":     float64(42),
		"match_count":    float64(1),
		"bypassed":       false,
	}
	line, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("fixture: encode legacy row: %v", err)
	}
	if _, present := legacy["status"]; present {
		t.Fatalf("fixture: the legacy row must carry no status field")
	}
	if err := os.WriteFile(filepath.Join(dir, SettingsDriftLedgerName), append(line, '\n'), 0o600); err != nil {
		t.Fatalf("fixture: write legacy ledger: %v", err)
	}

	result := AssessSettingsDrift(SettingsDriftParams{
		Dir: worktree, Root: root, Card: "t765", Runner: NewExecRunner(),
	})
	if result.Status != SettingsDriftClean {
		t.Fatalf("assessment: got %q, want %q (err=%v)", result.Status, SettingsDriftClean, result.Err)
	}

	rows := rawLedger(t, root)
	if len(rows) != 2 {
		t.Fatalf("ledger holds %d rows, want 2 — the legacy detection was not read as open", len(rows))
	}
	if got := rows[1]["status"]; got != "resolved" {
		t.Errorf("appended row status: got %v, want %q", got, "resolved")
	}
	if got := rows[1]["resolves_sha256"]; got != legacy["sha256"] {
		t.Errorf("resolves_sha256: got %v, want %v", got, legacy["sha256"])
	}
	if got := rows[1]["resolves_measured_at"]; got != legacy["measured_at"] {
		t.Errorf("resolves_measured_at: got %v, want %v", got, legacy["measured_at"])
	}
}
