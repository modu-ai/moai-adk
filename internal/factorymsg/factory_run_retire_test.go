package factorymsg

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// retireSandbox returns a project root OUTSIDE this repository's worktree set
// with HOME / MOAI_HOME / MOAI_CLAUDE_BIN scrubbed to it (REQ-012).
// CanonicalProjectRoot converges a linked worktree onto the primary checkout,
// so "use a temp dir" alone would mutate the developer's real factory state.
func retireSandbox(t *testing.T) string {
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

// deadIdentity returns the identity of a process that has exited. Whether the
// pid is later recycled or not, the classification is dead: an unrecycled pid
// probes dead, and a recycled one carries a different start fingerprint.
func deadIdentity(t *testing.T) (int, string) {
	t.Helper()
	cmd := exec.Command("/bin/sh", "-c", "exit 0")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start short-lived process: %v", err)
	}
	pid := cmd.Process.Pid
	fingerprint, _ := homestate.ProbeProcessIdentity(pid)
	if fingerprint == "" {
		fingerprint = "fingerprint-of-a-process-that-has-exited"
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("wait for short-lived process: %v", err)
	}
	return pid, fingerprint
}

func liveIdentity(t *testing.T) (int, string) {
	t.Helper()
	pid := os.Getpid()
	fingerprint := homestate.CurrentProcessFingerprint()
	if fingerprint == "" {
		t.Skip("this host cannot report its own process-start fingerprint")
	}
	return pid, fingerprint
}

func recordRun(t *testing.T, root, runID string, pid int, start string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{
		RunID: runID, Backend: "claude", ManifestJSON: "{}", LeadPID: pid, LeadProcessStart: start,
	}); err != nil {
		t.Fatalf("record run %s: %v", runID, err)
	}
}

func statusOf(t *testing.T, root, runID string) string {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	var status string
	if err := db.DB.QueryRow(`SELECT status FROM runs WHERE run_id=?`, runID).Scan(&status); err != nil {
		t.Fatalf("read status %s: %v", runID, err)
	}
	return status
}

// AC-004 — a stale run whose owner is dead is retired at resolution time and
// the join succeeds on the survivor.
//
// NOTE (SPEC defect, reported with the run): AC-004's Given says "two active
// runs whose owners are BOTH dead" while its Then says the join succeeds with
// the joined run as the sole remaining active row. Those cannot both hold:
// reconciliation retires every dead owner, so two dead owners leave zero
// active runs and resolution correctly fails closed. Both readings are
// exercised — this test covers the Then (one dead + one live, the shape the
// reproduction actually produced), and TestResolveActiveRunBothOwnersDead
// covers the Given.
func TestResolveActiveRunRetiresDeadOwnerAndJoins(t *testing.T) {
	root := retireSandbox(t)
	deadPID, deadStart := deadIdentity(t)
	livePID, liveStart := liveIdentity(t)
	recordRun(t, root, "run-stale", deadPID, deadStart)
	recordRun(t, root, "run-current", livePID, liveStart)

	got, err := ResolveActiveRun(context.Background(), root, "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != "run-current" {
		t.Fatalf("resolved %q, want run-current", got)
	}
	if st := statusOf(t, root, "run-stale"); st != "retired" {
		t.Fatalf("stale run status = %q, want \"retired\"", st)
	}
	if st := statusOf(t, root, "run-current"); st != "active" {
		t.Fatalf("current run status = %q, want \"active\"", st)
	}
}

// AC-004 (the Given, followed to its actual consequence) — when every active
// owner is dead, every run is retired and resolution fails closed with
// NO_ACTIVE_FACTORY. Reconciliation only reduces the active set.
func TestResolveActiveRunBothOwnersDead(t *testing.T) {
	root := retireSandbox(t)
	pidA, startA := deadIdentity(t)
	pidB, startB := deadIdentity(t)
	recordRun(t, root, "run-a", pidA, startA)
	recordRun(t, root, "run-b", pidB, startB)

	_, err := ResolveActiveRun(context.Background(), root, "")
	if err == nil || !strings.Contains(err.Error(), "NO_ACTIVE_FACTORY") {
		t.Fatalf("err = %v, want NO_ACTIVE_FACTORY", err)
	}
	if st := statusOf(t, root, "run-a"); st != "retired" {
		t.Fatalf("run-a status = %q, want \"retired\"", st)
	}
	if st := statusOf(t, root, "run-b"); st != "retired" {
		t.Fatalf("run-b status = %q, want \"retired\"", st)
	}
}

// AC-008 — two runs surviving reconciliation (one live, one indeterminate)
// still fail closed, and the failure names each remaining run id together with
// its owner classification.
func TestResolveActiveRunAmbiguityNamesClassifications(t *testing.T) {
	root := retireSandbox(t)
	livePID, liveStart := liveIdentity(t)
	recordRun(t, root, "run-live", livePID, liveStart)
	// No stamp and no lead peer: neither identity source yields anything.
	recordRun(t, root, "run-unknown", 0, "")

	_, err := ResolveActiveRun(context.Background(), root, "")
	if err == nil {
		t.Fatal("resolve succeeded, want AMBIGUOUS_FACTORY")
	}
	msg := err.Error()
	for _, want := range []string{"AMBIGUOUS_FACTORY", "run-live", "live", "run-unknown", "indeterminate"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q does not contain %q", msg, want)
		}
	}
	if st := statusOf(t, root, "run-live"); st != "active" {
		t.Fatalf("live run status = %q, want \"active\"", st)
	}
	if st := statusOf(t, root, "run-unknown"); st != "active" {
		t.Fatalf("indeterminate run status = %q, want \"active\"", st)
	}
}

// AC-009 — the fail-closed sentinels survive the new path. Zero active runs
// stay NO_ACTIVE_FACTORY; two live-owner runs stay AMBIGUOUS_FACTORY. Neither
// case returns a run id.
func TestResolveActiveRunPreservesFailClosedSentinels(t *testing.T) {
	root := retireSandbox(t)
	id, err := ResolveActiveRun(context.Background(), root, "")
	if err == nil || !strings.Contains(err.Error(), "NO_ACTIVE_FACTORY") {
		t.Fatalf("zero active: err = %v, want NO_ACTIVE_FACTORY", err)
	}
	if id != "" {
		t.Fatalf("zero active returned run id %q", id)
	}

	livePID, liveStart := liveIdentity(t)
	recordRun(t, root, "run-1", livePID, liveStart)
	recordRun(t, root, "run-2", livePID, liveStart)
	id, err = ResolveActiveRun(context.Background(), root, "")
	if err == nil || !strings.Contains(err.Error(), "AMBIGUOUS_FACTORY") {
		t.Fatalf("two live: err = %v, want AMBIGUOUS_FACTORY", err)
	}
	if id != "" {
		t.Fatalf("two live returned run id %q", id)
	}
	if st := statusOf(t, root, "run-1"); st != "active" {
		t.Fatalf("run-1 status = %q, want \"active\"", st)
	}
	if st := statusOf(t, root, "run-2"); st != "active" {
		t.Fatalf("run-2 status = %q, want \"active\"", st)
	}
}

// AC-007 — a schema-v2 database holding two unstamped active rows, each with a
// registered role='lead' peer whose pid is dead, migrates to v3 and both
// legacy rows are retired through the peer fallback, so the join no longer
// fails with AMBIGUOUS_FACTORY.
func TestResolveActiveRunMigratesAndReapsLegacyRowsViaPeerFallback(t *testing.T) {
	root := retireSandbox(t)
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatalf("factory db path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create factory dir: %v", err)
	}
	seedFactoryV2(t, path, "legacy-a", "legacy-b")

	deadA, startA := deadIdentity(t)
	deadB, startB := deadIdentity(t)
	registerLeadPeer(t, root, "legacy-a", deadA, startA)
	registerLeadPeer(t, root, "legacy-b", deadB, startB)

	_, err = ResolveActiveRun(context.Background(), root, "")
	if err == nil || strings.Contains(err.Error(), "AMBIGUOUS_FACTORY") {
		t.Fatalf("err = %v, want the ambiguity to be gone (NO_ACTIVE_FACTORY after both legacy rows retire)", err)
	}
	if st := statusOf(t, root, "legacy-a"); st != "retired" {
		t.Fatalf("legacy-a status = %q, want \"retired\"", st)
	}
	if st := statusOf(t, root, "legacy-b"); st != "retired" {
		t.Fatalf("legacy-b status = %q, want \"retired\"", st)
	}
}

// AC-016 (peer-agreement half) — LeadPeerIdentity reports the identity the
// run's role='lead' peer carries, including a launch-pending peer, which is
// the state a lead sits in until its SessionStart hook binds a session UUID.
func TestLeadPeerIdentityReadsLaunchPendingLead(t *testing.T) {
	root := retireSandbox(t)
	recordRun(t, root, "run-p", 0, "")
	registerLeadPeer(t, root, "run-p", 4321, "peer-start")

	pid, start, ok := LeadPeerIdentity(root, "run-p")
	if !ok || pid != 4321 || start != "peer-start" {
		t.Fatalf("LeadPeerIdentity = (%d, %q, %v), want (4321, \"peer-start\", true)", pid, start, ok)
	}

	if _, _, ok := LeadPeerIdentity(root, "run-absent"); ok {
		t.Fatal("LeadPeerIdentity reported an identity for a run with no broker")
	}
}

// seedFactoryV2 writes a factory database at schema version 2 — the shape of
// every database written before this SPEC lands — holding the named runs as
// unstamped active rows.
func seedFactoryV2(t *testing.T, path string, runIDs ...string) {
	t.Helper()
	seed, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
	if err != nil {
		t.Fatalf("open seed db: %v", err)
	}
	if _, err := seed.Exec(`
CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);
CREATE TABLE runs (
  run_id TEXT PRIMARY KEY,
  lead_session_id TEXT NOT NULL DEFAULT '',
  lead_backend TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  manifest_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
INSERT INTO meta(key,value) VALUES('schema_version','2');
`); err != nil {
		t.Fatalf("seed v2 schema: %v", err)
	}
	for _, id := range runIDs {
		if _, err := seed.Exec(`INSERT INTO runs(run_id,status,created_at,updated_at) VALUES(?,'active','t','t')`, id); err != nil {
			t.Fatalf("seed run %s: %v", id, err)
		}
	}
	if err := seed.Close(); err != nil {
		t.Fatalf("close seed db: %v", err)
	}
}

// registerLeadPeer registers the run's role='lead' peer in launch-pending
// state — the state a lead sits in from launch until its SessionStart hook
// binds a session UUID.
func registerLeadPeer(t *testing.T, root, runID string, pid int, processStart string) {
	t.Helper()
	s, err := Open(root, runID)
	if err != nil {
		t.Fatalf("open broker for %s: %v", runID, err)
	}
	defer func() { _ = s.Close() }()
	if _, err := s.RegisterLaunchPending(context.Background(), Peer{
		ProjectKey: "project", RunID: runID, Backend: "claude",
		Role: "lead", Slot: "lead", Generation: 1, PID: pid, ProcessStart: processStart,
	}); err != nil {
		t.Fatalf("register lead peer for %s: %v", runID, err)
	}
}
