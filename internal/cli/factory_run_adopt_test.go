package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
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

// Several live-records runs with dead owners are ambiguous: nothing is
// adopted, the minted id stands, and every candidate is named in the warn
// output with its re-adoption command.
func TestAdoptResumedLeaderRunWarnsOnMultipleCandidates(t *testing.T) {
	root := adoptSandbox(t)
	db := openAdoptFactory(t, root)
	seedDeadRunWithLiveRecords(t, db, "run-a")
	seedDeadRunWithLiveRecords(t, db, "run-b")
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
		t.Fatalf("adopted = %q, want none (ambiguous)", adopted)
	}
	if got := os.Getenv(config.EnvFactoryRunID); got != "minted01" {
		t.Fatalf("env %s = %q, want minted01 (untouched)", config.EnvFactoryRunID, got)
	}
	out := warn.String()
	for _, id := range []string{"run-a", "run-b"} {
		if !strings.Contains(out, id) {
			t.Fatalf("warn %q does not name candidate %s", out, id)
		}
	}
	if !strings.Contains(out, "--factory-run") {
		t.Fatalf("warn %q does not offer the re-adoption command", out)
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
