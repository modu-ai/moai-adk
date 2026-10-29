package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// adoptSandbox mirrors the homestate factorySandbox fixture: a project root
// OUTSIDE this repository's worktree set with the home variables scrubbed to a
// temp dir, so CanonicalProjectRoot never converges onto the developer's real
// factory.db.
func adoptSandbox(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MOAI_HOME", filepath.Join(home, ".moai"))
	t.Setenv("MOAI_CLAUDE_BIN", filepath.Join(home, "bin", "claude"))
	root := filepath.Join(t.TempDir(), "sandbox-project")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatalf("create sandbox project root: %v", err)
	}
	return root
}

func openAdoptFactory(t *testing.T, root string) *homestate.FactoryDB {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	return db
}

// seedDeadRunWithLiveRecords records one active run whose stamp can never
// classify live — the pid exceeds every platform's pid_max and the recorded
// fingerprint names a process that never ran — plus one non-terminal card, so
// the default production options (the batch identity probe) classify it dead.
func seedDeadRunWithLiveRecords(t *testing.T, db *homestate.FactoryDB, runID string) {
	t.Helper()
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{
		RunID: runID, ManifestJSON: "{}", LeadPID: 99999999, LeadProcessStart: "gone-before-resume",
	}); err != nil {
		t.Fatalf("record run %s: %v", runID, err)
	}
	if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,updated_at) VALUES(?,?,?,?)`,
		runID, "card-"+runID, "run", "2026-10-14T00:00:00Z"); err != nil {
		t.Fatalf("insert card for %s: %v", runID, err)
	}
}

// The incident shape: a resume-shaped, freshly-minted leader start adopts the
// single dead-owner live-records run — the published id becomes the prior
// run's id, the adopted row is restamped to this process, and nothing is
// warned.
func TestAdoptResumedLeaderRunAdoptsSingleCandidate(t *testing.T) {
	root := adoptSandbox(t)
	db := openAdoptFactory(t, root)
	seedDeadRunWithLiveRecords(t, db, "run-old")
	if err := db.Close(); err != nil {
		t.Fatalf("close seeding handle: %v", err)
	}

	t.Setenv(config.EnvFactoryRunID, "minted01")
	var warn bytes.Buffer
	adopted, err := adoptResumedLeaderRun(root, true, &warn)
	if err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if adopted != "run-old" {
		t.Fatalf("adopted = %q, want run-old (the run id must survive the leader resume)", adopted)
	}
	if got := os.Getenv(config.EnvFactoryRunID); got != "run-old" {
		t.Fatalf("env %s = %q, want run-old (messaging must rebind to the adopted run)", config.EnvFactoryRunID, got)
	}
	// Sync-audit t1707 F4 — the conventional address is republished from the
	// adopted id too: the minted id's address must not survive the adoption.
	if got := os.Getenv(config.EnvFactoryLeadAddr); got != factory.FactoryLeaderSocketPath("run-old") {
		t.Fatalf("env %s = %q, want the adopted run's address %q", config.EnvFactoryLeadAddr, got, factory.FactoryLeaderSocketPath("run-old"))
	}
	if warn.Len() != 0 {
		t.Fatalf("warn = %q, want empty", warn.String())
	}

	db = openAdoptFactory(t, root)
	defer func() { _ = db.Close() }()
	var pid int
	var start string
	if err := db.DB.QueryRow(`SELECT lead_pid, lead_process_start FROM runs WHERE run_id='run-old'`).Scan(&pid, &start); err != nil {
		t.Fatalf("read adopted row: %v", err)
	}
	if pid != os.Getpid() || start != homestate.CurrentProcessFingerprint() {
		t.Fatalf("adopted stamp = (%d, %q), want this process (%d, %q)", pid, start, os.Getpid(), homestate.CurrentProcessFingerprint())
	}
}

// A start that is not resume-shaped (a cold start, or an explicit
// --factory-run) adopts nothing and leaves the published id alone.
func TestAdoptResumedLeaderRunSkipsWhenNotResumeShaped(t *testing.T) {
	root := adoptSandbox(t)
	db := openAdoptFactory(t, root)
	seedDeadRunWithLiveRecords(t, db, "run-old")
	if err := db.Close(); err != nil {
		t.Fatalf("close seeding handle: %v", err)
	}

	t.Setenv(config.EnvFactoryRunID, "minted01")
	var warn bytes.Buffer
	adopted, err := adoptResumedLeaderRun(root, false, &warn)
	if err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if adopted != "" {
		t.Fatalf("adopted = %q, want none", adopted)
	}
	if got := os.Getenv(config.EnvFactoryRunID); got != "minted01" {
		t.Fatalf("env %s = %q, want minted01 (untouched)", config.EnvFactoryRunID, got)
	}
	if warn.Len() != 0 {
		t.Fatalf("warn = %q, want empty", warn.String())
	}
}

// Several live-records runs with dead owners no longer stop the adoption
// (SPEC-FACTORY-COMPLETION-RECOVERY-001 REQ-FCR-017/018, decision Q2 — the one
// expectation this card reverses): the NEWEST is adopted, the other's cards
// move under it, and the other run retires. The warn output names the moved
// card; the ambiguity-refusal this test used to pin is replaced by the guards
// in TestLeaderStartAutoAdoptGuards.
func TestAdoptResumedLeaderRunWarnsOnMultipleCandidates(t *testing.T) {
	root := adoptSandbox(t)
	db := openAdoptFactory(t, root)
	seedDeadRunWithLiveRecords(t, db, "run-a")
	seedDeadRunWithLiveRecords(t, db, "run-b")
	// run-b is the newer run: pin the creation times so the order is stated,
	// not left to the clock.
	for id, at := range map[string]string{"run-a": "2026-10-01T00:00:00Z", "run-b": "2026-10-02T00:00:00Z"} {
		if _, err := db.DB.Exec(`UPDATE runs SET created_at=? WHERE run_id=?`, at, id); err != nil {
			t.Fatalf("pin created_at of %s: %v", id, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close seeding handle: %v", err)
	}

	t.Setenv(config.EnvFactoryRunID, "minted01")
	var warn bytes.Buffer
	adopted, err := adoptResumedLeaderRun(root, true, &warn)
	if err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if adopted != "run-b" {
		t.Fatalf("adopted = %q, want run-b (the newest dead-owner run)", adopted)
	}
	if got := os.Getenv(config.EnvFactoryRunID); got != "run-b" {
		t.Fatalf("env %s = %q, want run-b", config.EnvFactoryRunID, got)
	}
	out := warn.String()
	if !strings.Contains(out, "run-a") || !strings.Contains(out, "card-run-a") {
		t.Fatalf("warn %q does not name the consolidated run and its moved card", out)
	}

	db = openAdoptFactory(t, root)
	defer func() { _ = db.Close() }()
	var status string
	if err := db.DB.QueryRow(`SELECT status FROM runs WHERE run_id='run-a'`).Scan(&status); err != nil || status != "retired" {
		t.Fatalf("run-a status = %q (%v), want retired", status, err)
	}
	var owner string
	if err := db.DB.QueryRow(`SELECT run_id FROM cards WHERE card_id='card-run-a'`).Scan(&owner); err != nil || owner != "run-b" {
		t.Fatalf("card-run-a lives under %q (%v), want run-b", owner, err)
	}
}

// No candidate at all (a genuinely fresh machine state) mints as before and
// stays silent — the cold-start semantics are unchanged.
func TestAdoptResumedLeaderRunMintsWhenNoCandidates(t *testing.T) {
	root := adoptSandbox(t)

	t.Setenv(config.EnvFactoryRunID, "minted01")
	var warn bytes.Buffer
	adopted, err := adoptResumedLeaderRun(root, true, &warn)
	if err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if adopted != "" {
		t.Fatalf("adopted = %q, want none", adopted)
	}
	if got := os.Getenv(config.EnvFactoryRunID); got != "minted01" {
		t.Fatalf("env %s = %q, want minted01 (untouched)", config.EnvFactoryRunID, got)
	}
	if warn.Len() != 0 {
		t.Fatalf("warn = %q, want empty", warn.String())
	}
}

// Card-review round 2 — the automatic path shares the explicit path's
// cross-backend guard: a codex-led run is warned, never adopted by a claude
// leader.
func TestAdoptResumedLeaderRunWarnsCodexLedCandidate(t *testing.T) {
	root := adoptSandbox(t)
	db := openAdoptFactory(t, root)
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{
		RunID: "run-cx", Backend: "codex", ManifestJSON: "{}", LeadPID: 99999999, LeadProcessStart: "gone-before-resume",
	}); err != nil {
		t.Fatalf("record run: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,updated_at) VALUES(?,?,?,?)`,
		"run-cx", "card-run-cx", "run", "2026-10-14T00:00:00Z"); err != nil {
		t.Fatalf("insert card: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close seeding handle: %v", err)
	}

	t.Setenv(config.EnvFactoryRunID, "minted01")
	var warn bytes.Buffer
	adopted, err := adoptResumedLeaderRun(root, true, &warn)
	if err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if adopted != "" {
		t.Fatalf("adopted = %q, want none (a codex-led run is not adoptable by a claude leader)", adopted)
	}
	if got := os.Getenv(config.EnvFactoryRunID); got != "minted01" {
		t.Fatalf("env %s = %q, want minted01 (untouched)", config.EnvFactoryRunID, got)
	}
	if !strings.Contains(warn.String(), "codex") {
		t.Fatalf("warn %q does not name the codex-led refusal", warn.String())
	}
}

// The minted-ness predicate mirrors leaderRunID's three steps: only the
// NewRunID() fallback counts as minted.
func TestLeaderRunIDWasMinted(t *testing.T) {
	cases := []struct {
		label   string
		priorID string
		want    bool
	}{
		{"leader", "", true},         // bare role, no env, no pasted id → minted
		{"leader", "priorid", false}, // the environment carried the prior id
		{"leader-abc123", "", false}, // an operator-pasted run id
		{"leader-1", "", true},       // a collision bump number is not a run id
		{"leader-2x", "", false},     // a non-numeric suffix reads as a pasted id
	}
	for _, tc := range cases {
		if got := leaderRunIDWasMinted(tc.label, tc.priorID); got != tc.want {
			t.Fatalf("leaderRunIDWasMinted(%q, %q) = %v, want %v", tc.label, tc.priorID, got, tc.want)
		}
	}
}
