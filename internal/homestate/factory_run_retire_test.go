package homestate

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// factorySandbox returns a project root OUTSIDE this repository's worktree set
// with HOME / MOAI_HOME / MOAI_CLAUDE_BIN scrubbed to it (REQ-012). A temp dir
// alone is not sufficient: CanonicalProjectRoot converges a linked worktree
// onto the primary checkout, so a fixture rooted inside a worktree would write
// to the developer's real factory.db.
func factorySandbox(t *testing.T) string {
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

func openSandboxFactory(t *testing.T) *FactoryDB {
	t.Helper()
	db, err := OpenFactory(factorySandbox(t))
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func mustRecordRun(t *testing.T, db *FactoryDB, run FactoryRun) {
	t.Helper()
	if err := db.RecordRun(context.Background(), run); err != nil {
		t.Fatalf("record run %s: %v", run.RunID, err)
	}
}

func runStatus(t *testing.T, db *FactoryDB, runID string) string {
	t.Helper()
	var status string
	if err := db.DB.QueryRow(`SELECT status FROM runs WHERE run_id=?`, runID).Scan(&status); err != nil {
		t.Fatalf("read status for %s: %v", runID, err)
	}
	return status
}

// fixedClassifier answers with one classification regardless of identity, so a
// test can drive a retirement path with a value the production code does not
// enumerate.
func fixedClassifier(c OwnerClassification) OwnerClassifier {
	return func(int, string) OwnerClassification { return c }
}

// AC-001 — a recorded run carries a non-zero owner pid and a non-empty
// process-start fingerprint, and the database reports schema version 3.
func TestRecordRunStampsSessionOwnerIdentity(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-a", Backend: "claude", ManifestJSON: "{}", LeadPID: 4242, LeadProcessStart: "1700000000.000001"})

	var pid int
	var start string
	if err := db.DB.QueryRow(`SELECT lead_pid, lead_process_start FROM runs WHERE run_id=?`, "run-a").Scan(&pid, &start); err != nil {
		t.Fatalf("read owner stamp: %v", err)
	}
	if pid != 4242 || start != "1700000000.000001" {
		t.Fatalf("owner stamp = (%d, %q), want (4242, \"1700000000.000001\")", pid, start)
	}

	var version string
	if err := db.DB.QueryRow(`SELECT value FROM meta WHERE key='schema_version'`).Scan(&version); err != nil {
		t.Fatalf("read schema version: %v", err)
	}
	if version != "3" {
		t.Fatalf("schema_version = %q, want \"3\"", version)
	}
}

// AC-002 — retirement is a status transition that preserves the row and
// appends a run.retired event; it never deletes.
func TestRetireRunPreservesRowAndAppendsEvent(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-dead", ManifestJSON: "{}", LeadPID: 1, LeadProcessStart: "x"})

	rec, err := db.ReconcileActiveRuns(context.Background(), ReconcileOptions{Classify: fixedClassifier(OwnerDead)})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if len(rec.Retired) != 1 || rec.Retired[0].RunID != "run-dead" {
		t.Fatalf("retired = %+v, want exactly run-dead", rec.Retired)
	}

	var count int
	if err := db.DB.QueryRow(`SELECT count(*) FROM runs WHERE run_id=?`, "run-dead").Scan(&count); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("row count = %d, want 1 (the row must survive retirement)", count)
	}
	if got := runStatus(t, db, "run-dead"); got != "retired" {
		t.Fatalf("status = %q, want \"retired\"", got)
	}

	var events int
	if err := db.DB.QueryRow(`SELECT count(*) FROM events WHERE run_id=? AND kind='run.retired'`, "run-dead").Scan(&events); err != nil {
		t.Fatalf("count events: %v", err)
	}
	if events != 1 {
		t.Fatalf("run.retired event count = %d, want 1", events)
	}
}

// AC-003 — classification consults PID *together with* the process-start
// fingerprint. The fingerprints are fixture values: on linux the real
// fingerprint has one-second resolution, so a genuine same-second reuse is
// indistinguishable there by construction.
func TestClassifyOwnerUsesFingerprintNotBarePID(t *testing.T) {
	live := func(int) (string, ProcessIdentityState) { return "start-A", ProcessIdentityLive }
	dead := func(int) (string, ProcessIdentityState) { return "", ProcessIdentityDead }
	unreadable := func(int) (string, ProcessIdentityState) { return "", ProcessIdentityIndeterminate }

	if got := ClassifyOwnerWith(live, 99, "start-A"); got != OwnerLive {
		t.Fatalf("matching fingerprint = %q, want live", got)
	}
	// PID-reuse shape: the pid is live, the fingerprint differs.
	if got := ClassifyOwnerWith(live, 99, "start-B"); got != OwnerDead {
		t.Fatalf("reused pid = %q, want dead (a classifier consulting only the PID fails here)", got)
	}
	if got := ClassifyOwnerWith(dead, 99, "start-A"); got != OwnerDead {
		t.Fatalf("dead pid = %q, want dead", got)
	}
	if got := ClassifyOwnerWith(unreadable, 99, "start-A"); got != OwnerIndeterminate {
		t.Fatalf("unreadable fingerprint = %q, want indeterminate", got)
	}
	// REQ-003b — an indistinguishable fingerprint pair resolves toward live.
	if got := ClassifyOwnerWith(live, 99, "start-A"); got == OwnerDead {
		t.Fatalf("indistinguishable pair classified dead; the residual error must fall on the live side")
	}
	// No identity from either source.
	if got := ClassifyOwnerWith(live, 0, ""); got != OwnerIndeterminate {
		t.Fatalf("absent identity = %q, want indeterminate", got)
	}
}

// AC-005 / AC-015b — a dead owner's run is retired, a live owner's run is not.
func TestReconcileRetiresDeadAndLeavesLive(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-live", ManifestJSON: "{}", LeadPID: 10, LeadProcessStart: "alive"})
	mustRecordRun(t, db, FactoryRun{RunID: "run-dead", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "gone"})

	classify := func(_ int, start string) OwnerClassification {
		if start == "alive" {
			return OwnerLive
		}
		return OwnerDead
	}
	rec, err := db.ReconcileActiveRuns(context.Background(), ReconcileOptions{Classify: classify})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if got := runStatus(t, db, "run-live"); got != "active" {
		t.Fatalf("live-owner run status = %q, want \"active\"", got)
	}
	if got := runStatus(t, db, "run-dead"); got != "retired" {
		t.Fatalf("dead-owner run status = %q, want \"retired\"", got)
	}
	if len(rec.Remaining) != 1 || rec.Remaining[0].RunID != "run-live" || rec.Remaining[0].Classification != OwnerLive {
		t.Fatalf("remaining = %+v, want exactly run-live classified live", rec.Remaining)
	}
}

// AC-006 — an owner that cannot be probed stays active and is reported
// indeterminate.
func TestReconcileLeavesIndeterminateActive(t *testing.T) {
	db := openSandboxFactory(t)
	// No stamp and no lead peer: neither identity source yields anything.
	mustRecordRun(t, db, FactoryRun{RunID: "run-unknown", ManifestJSON: "{}"})

	rec, err := db.ReconcileActiveRuns(context.Background(), ReconcileOptions{})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if got := runStatus(t, db, "run-unknown"); got != "active" {
		t.Fatalf("status = %q, want \"active\"", got)
	}
	if len(rec.Remaining) != 1 || rec.Remaining[0].Classification != OwnerIndeterminate {
		t.Fatalf("remaining = %+v, want run-unknown classified indeterminate", rec.Remaining)
	}
	if len(rec.Retired) != 0 {
		t.Fatalf("retired = %+v, want none", rec.Retired)
	}
}

// AC-007 (homestate half) — a run with no owner stamp routes to the
// role='lead' peer identity supplied by the fallback.
func TestReconcileUnstampedRunUsesLeadPeerFallback(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "legacy-1", ManifestJSON: "{}"})
	mustRecordRun(t, db, FactoryRun{RunID: "legacy-2", ManifestJSON: "{}"})

	fallback := func(runID string) (int, string, bool) {
		switch runID {
		case "legacy-1":
			return 31, "peer-start-1", true
		case "legacy-2":
			return 32, "peer-start-2", true
		}
		return 0, "", false
	}
	rec, err := db.ReconcileActiveRuns(context.Background(), ReconcileOptions{
		Fallback: fallback,
		Classify: fixedClassifier(OwnerDead),
	})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if len(rec.Retired) != 2 {
		t.Fatalf("retired = %+v, want both legacy rows", rec.Retired)
	}
	if len(rec.Remaining) != 0 {
		t.Fatalf("remaining = %+v, want none", rec.Remaining)
	}
}

// AC-007 (migration half) — a schema-v2 database carrying unstamped active
// rows migrates to version 3 with the rows intact and legible.
func TestMigrateFactoryV2ToV3PreservesRows(t *testing.T) {
	root := factorySandbox(t)
	path, err := FactoryDBPath(root)
	if err != nil {
		t.Fatalf("factory db path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create factory dir: %v", err)
	}
	seed, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
	if err != nil {
		t.Fatalf("open seed db: %v", err)
	}
	// The v2 `runs` shape: no owner columns.
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
INSERT INTO runs(run_id,status,created_at,updated_at) VALUES('old-1','active','t','t');
INSERT INTO runs(run_id,status,created_at,updated_at) VALUES('old-2','active','t','t');
`); err != nil {
		t.Fatalf("seed v2 database: %v", err)
	}
	if err := seed.Close(); err != nil {
		t.Fatalf("close seed db: %v", err)
	}

	db, err := OpenFactoryPath(path)
	if err != nil {
		t.Fatalf("open migrated factory: %v", err)
	}
	defer func() { _ = db.Close() }()

	var version string
	if err := db.DB.QueryRow(`SELECT value FROM meta WHERE key='schema_version'`).Scan(&version); err != nil {
		t.Fatalf("read schema version: %v", err)
	}
	if version != "3" {
		t.Fatalf("schema_version = %q, want \"3\"", version)
	}
	var pid int
	if err := db.DB.QueryRow(`SELECT lead_pid FROM runs WHERE run_id='old-1'`).Scan(&pid); err != nil {
		t.Fatalf("read migrated owner column: %v", err)
	}
	if pid != 0 {
		t.Fatalf("legacy lead_pid = %d, want 0 (the sentinel that routes to the peer fallback)", pid)
	}
}

// AC-015a — removing the liveness guard, so that a live or indeterminate owner
// becomes retirable, must make a test fail. This is that test.
func TestReconcileNeverRetiresLiveOrIndeterminate(t *testing.T) {
	for _, c := range []OwnerClassification{OwnerLive, OwnerIndeterminate} {
		db := openSandboxFactory(t)
		mustRecordRun(t, db, FactoryRun{RunID: "run-x", ManifestJSON: "{}", LeadPID: 7, LeadProcessStart: "s"})
		if _, err := db.ReconcileActiveRuns(context.Background(), ReconcileOptions{Classify: fixedClassifier(c)}); err != nil {
			t.Fatalf("reconcile (%s): %v", c, err)
		}
		if got := runStatus(t, db, "run-x"); got != "active" {
			t.Fatalf("classification %s retired the run (status %q); a live run must never be retired", c, got)
		}
	}
}

// AC-017 — a classification value the retirement code does not enumerate is
// declined by EVERY retirement path, without the code having been told about
// it. This is the criterion that separates a positive `dead` gate from a
// reject-list: a guard written `if c == live || c == indeterminate { refuse }`
// keeps AC-005, AC-006 and AC-010 green and turns this one red.
func TestEveryRetirementPathDeclinesUnenumeratedClassification(t *testing.T) {
	const fourth = OwnerClassification("unknown-host")

	t.Run("reconciler", func(t *testing.T) {
		db := openSandboxFactory(t)
		mustRecordRun(t, db, FactoryRun{RunID: "run-4th", ManifestJSON: "{}", LeadPID: 9, LeadProcessStart: "s"})
		rec, err := db.ReconcileActiveRuns(context.Background(), ReconcileOptions{Classify: fixedClassifier(fourth)})
		if err != nil {
			t.Fatalf("reconcile: %v", err)
		}
		if len(rec.Retired) != 0 {
			t.Fatalf("retired %+v on an unenumerated classification; retirement must be gated on a positive dead", rec.Retired)
		}
		if got := runStatus(t, db, "run-4th"); got != "active" {
			t.Fatalf("status = %q, want \"active\"", got)
		}
	})

	t.Run("migration pass", func(t *testing.T) {
		db := openSandboxFactory(t)
		// Unstamped legacy row reconciled through the peer fallback.
		mustRecordRun(t, db, FactoryRun{RunID: "legacy-4th", ManifestJSON: "{}"})
		rec, err := db.ReconcileActiveRuns(context.Background(), ReconcileOptions{
			Fallback: func(string) (int, string, bool) { return 11, "peer", true },
			Classify: fixedClassifier(fourth),
		})
		if err != nil {
			t.Fatalf("reconcile: %v", err)
		}
		if len(rec.Retired) != 0 {
			t.Fatalf("migration pass retired %+v on an unenumerated classification", rec.Retired)
		}
		if got := runStatus(t, db, "legacy-4th"); got != "active" {
			t.Fatalf("status = %q, want \"active\"", got)
		}
	})

	t.Run("operator retire", func(t *testing.T) {
		db := openSandboxFactory(t)
		mustRecordRun(t, db, FactoryRun{RunID: "run-op", ManifestJSON: "{}", LeadPID: 9, LeadProcessStart: "s"})
		got, err := db.RetireRunIfDead(context.Background(), "run-op", ReconcileOptions{Classify: fixedClassifier(fourth)})
		if !errors.Is(err, ErrRunOwnerNotDead) {
			t.Fatalf("err = %v, want ErrRunOwnerNotDead", err)
		}
		if got != fourth {
			t.Fatalf("classification = %q, want %q", got, fourth)
		}
		if st := runStatus(t, db, "run-op"); st != "active" {
			t.Fatalf("status = %q, want \"active\"", st)
		}
	})
}

// AC-010 (predicate half) — the operator retire path refuses a live owner AND
// an indeterminate owner, names the classification as the reason, and leaves
// the run active. Only a dead owner is retired.
func TestRetireRunIfDeadRefusesLiveAndIndeterminate(t *testing.T) {
	for _, c := range []OwnerClassification{OwnerLive, OwnerIndeterminate} {
		db := openSandboxFactory(t)
		mustRecordRun(t, db, FactoryRun{RunID: "run-r", ManifestJSON: "{}", LeadPID: 5, LeadProcessStart: "s"})
		got, err := db.RetireRunIfDead(context.Background(), "run-r", ReconcileOptions{Classify: fixedClassifier(c)})
		if !errors.Is(err, ErrRunOwnerNotDead) {
			t.Fatalf("%s: err = %v, want ErrRunOwnerNotDead", c, err)
		}
		if !strings.Contains(err.Error(), string(c)) {
			t.Fatalf("%s: error %q does not name the classification", c, err)
		}
		if got != c {
			t.Fatalf("classification = %q, want %q", got, c)
		}
		if st := runStatus(t, db, "run-r"); st != "active" {
			t.Fatalf("%s: status = %q, want \"active\"", c, st)
		}
	}

	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-d", ManifestJSON: "{}", LeadPID: 5, LeadProcessStart: "s"})
	if _, err := db.RetireRunIfDead(context.Background(), "run-d", ReconcileOptions{Classify: fixedClassifier(OwnerDead)}); err != nil {
		t.Fatalf("dead owner: %v", err)
	}
	if st := runStatus(t, db, "run-d"); st != "retired" {
		t.Fatalf("status = %q, want \"retired\"", st)
	}
}

// AC-010 (listing half) — every run is reported with its status and owner
// classification.
func TestClassifyRunsReportsStatusAndClassification(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-1", ManifestJSON: "{}", LeadPID: 1, LeadProcessStart: "s"})
	mustRecordRun(t, db, FactoryRun{RunID: "run-2", ManifestJSON: "{}"})

	owners, err := db.ClassifyRuns(context.Background(), ReconcileOptions{Classify: fixedClassifier(OwnerLive)})
	if err != nil {
		t.Fatalf("classify runs: %v", err)
	}
	if len(owners) != 2 {
		t.Fatalf("owners = %+v, want 2", owners)
	}
	byID := map[string]RunOwner{}
	for _, o := range owners {
		byID[o.RunID] = o
	}
	if byID["run-1"].Classification != OwnerLive || byID["run-1"].Status != "active" {
		t.Fatalf("run-1 = %+v", byID["run-1"])
	}
	// run-2 has no stamp and no fallback, so no identity reaches the classifier.
	if byID["run-2"].Classification != OwnerIndeterminate {
		t.Fatalf("run-2 = %+v, want indeterminate", byID["run-2"])
	}
}

// AC-012 (state half) — clearing the owner stamp leaves no row carrying the
// launching process's identity.
func TestStampAndClearRunOwner(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-s", ManifestJSON: "{}", LeadPID: 111, LeadProcessStart: "launcher"})

	if err := db.StampRunOwner(context.Background(), "run-s", 222, "session"); err != nil {
		t.Fatalf("stamp: %v", err)
	}
	var pid int
	var start string
	if err := db.DB.QueryRow(`SELECT lead_pid, lead_process_start FROM runs WHERE run_id='run-s'`).Scan(&pid, &start); err != nil {
		t.Fatalf("read stamp: %v", err)
	}
	if pid != 222 || start != "session" {
		t.Fatalf("stamp = (%d, %q), want (222, \"session\")", pid, start)
	}

	if err := db.ClearRunOwner(context.Background(), "run-s"); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if err := db.DB.QueryRow(`SELECT lead_pid, lead_process_start FROM runs WHERE run_id='run-s'`).Scan(&pid, &start); err != nil {
		t.Fatalf("read cleared stamp: %v", err)
	}
	if pid != 0 || start != "" {
		t.Fatalf("cleared stamp = (%d, %q), want (0, \"\")", pid, start)
	}
}
