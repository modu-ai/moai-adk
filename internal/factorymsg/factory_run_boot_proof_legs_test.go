package factorymsg

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// preBootAt is dated long before any plausible boot of the test host.
const preBootAt = "2001-09-11T10:25:37.923062Z"

// seedPreBootRun writes one active run row with the given owner stamp and a
// run.started event, every timestamp dated preBootAt.
func seedPreBootRun(t *testing.T, root, runID string, pid int, start string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`INSERT INTO runs(run_id,lead_session_id,lead_backend,status,manifest_json,created_at,updated_at,lead_pid,lead_process_start) VALUES(?,'','claude','active','{}',?,?,?,?)`, runID, preBootAt, preBootAt, pid, start); err != nil {
		t.Fatalf("seed run %s: %v", runID, err)
	}
	if _, err := db.DB.Exec(`INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,'run.started','{}',?)`, runID, preBootAt); err != nil {
		t.Fatalf("seed event %s: %v", runID, err)
	}
}

// reconcileProduction runs the reconciler with the option set every production
// retirement path passes, against the host's real boot time.
func reconcileProduction(t *testing.T, root string) homestate.Reconciliation {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	rec, err := db.ReconcileActiveRuns(context.Background(), ReconcileOptionsFor(root))
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	return rec
}

func remainingClassificationOf(rec homestate.Reconciliation, runID string) homestate.OwnerClassification {
	for _, o := range rec.Remaining {
		if o.RunID == runID {
			return o.Classification
		}
	}
	return ""
}

// assertDeclined checks both halves of a declining leg (AC-018): the run is
// still active AND it is reported indeterminate.
func assertDeclined(t *testing.T, root string, rec homestate.Reconciliation, runID string) {
	t.Helper()
	if st := statusOf(t, root, runID); st != "active" {
		t.Fatalf("%s status = %q, want \"active\" (the proof must decline)", runID, st)
	}
	if c := remainingClassificationOf(rec, runID); c != homestate.OwnerIndeterminate {
		t.Fatalf("%s classification = %q, want %q", runID, c, homestate.OwnerIndeterminate)
	}
}

// assertBootRetired is the positive control seeded beside each declining leg:
// with the same options and the same pre-boot timestamps, a run whose broker
// file is plainly absent IS retired, so the only premise the declining run
// fails is the one its leg names.
func assertBootRetired(t *testing.T, root string, rec homestate.Reconciliation, runID string) {
	t.Helper()
	if st := statusOf(t, root, runID); st != "retired" {
		t.Fatalf("control %s status = %q, want \"retired\" (the proof should hold for it)", runID, st)
	}
	for _, o := range rec.Retired {
		if o.RunID == runID && o.Classification == homestate.OwnerDead {
			return
		}
	}
	t.Fatalf("control %s is not among the retired-dead runs: %+v", runID, rec.Retired)
}

// AC-018 — a partial stamp (lead_pid >= 1, empty lead_process_start) is not
// routed to the peer lookup; with no broker file and every timestamp before
// the boot, the boot proof retires it dead.
func TestPartialStampWithoutBrokerIsBootProven(t *testing.T) {
	root := bootProofSandbox(t)
	seedPreBootRun(t, root, "run-partial", 4242, "")

	rec := reconcileProduction(t, root)

	if st := statusOf(t, root, "run-partial"); st != "retired" {
		t.Fatalf("run-partial status = %q, want \"retired\"", st)
	}
	if len(rec.Retired) != 1 || rec.Retired[0].RunID != "run-partial" || rec.Retired[0].Classification != homestate.OwnerDead {
		t.Fatalf("retired = %+v, want run-partial classified dead", rec.Retired)
	}
}

// AC-018 — the other side of premise 2 for a partial stamp. The broker file
// exists and carries a role='lead' peer with a COMPLETE identity (dead pid,
// non-empty process start), and every recorded timestamp predates the boot, so
// premise 2 is the only premise that declines. Routing the partial stamp to
// the peer lookup would classify it dead through the peer and retire it.
func TestPartialStampWithBrokerStaysIndeterminate(t *testing.T) {
	root := bootProofSandbox(t)
	seedPreBootRun(t, root, "run-partial", 4242, "")
	deadPID, deadStart := deadIdentity(t)
	if deadStart == "" {
		t.Fatal("dead peer identity carries an empty process start; the seed would not isolate premise 2")
	}
	registerLeadPeer(t, root, "run-partial", deadPID, deadStart)
	seedPreBootRun(t, root, "run-control", 0, "")

	rec := reconcileProduction(t, root)

	assertDeclined(t, root, rec, "run-partial")
	assertBootRetired(t, root, rec, "run-control")
}

// AC-018 — a broker check that fails for any reason other than "does not
// exist" answers "a record may exist". A regular file where the run's broker
// directory belongs makes the stat fail with ENOTDIR.
func TestLeadRecordAbsentForTreatsStatErrorAsPossibleRecord(t *testing.T) {
	root := bootProofSandbox(t)
	seedPreBootRun(t, root, "run-blocked", 0, "")
	seedPreBootRun(t, root, "run-control", 0, "")
	dir, err := homestate.FactoryDir(root)
	if err != nil {
		t.Fatalf("factory dir: %v", err)
	}
	messages := filepath.Join(dir, "messages")
	if err := os.MkdirAll(messages, 0o700); err != nil {
		t.Fatalf("create messages dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(messages, "run-blocked"), []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("plant blocking file: %v", err)
	}
	path, err := BrokerPath(root, "run-blocked")
	if err != nil {
		t.Fatalf("broker path: %v", err)
	}
	if _, err := os.Stat(path); err == nil || os.IsNotExist(err) {
		t.Fatalf("stat %s = %v, want an error other than not-exist (the leg's premise)", path, err)
	}

	if LeadRecordAbsentFor(root)("run-blocked") {
		t.Fatal("LeadRecordAbsentFor reported absence on a failed broker check")
	}
	rec := reconcileProduction(t, root)
	assertDeclined(t, root, rec, "run-blocked")
	assertBootRetired(t, root, rec, "run-control")
}

// AC-018 — a run id from which no broker path can be derived answers "a record
// may exist", never "absent".
func TestLeadRecordAbsentForRejectsUnderivableBrokerPath(t *testing.T) {
	root := bootProofSandbox(t)
	const unsafeID = "../x"
	if _, err := BrokerPath(root, unsafeID); err == nil {
		t.Fatalf("BrokerPath(%q) derived a path; the leg needs an underivable one", unsafeID)
	}
	seedPreBootRun(t, root, unsafeID, 0, "")
	seedPreBootRun(t, root, "run-control", 0, "")

	if LeadRecordAbsentFor(root)(unsafeID) {
		t.Fatal("LeadRecordAbsentFor reported absence for an underivable broker path")
	}
	rec := reconcileProduction(t, root)
	assertDeclined(t, root, rec, unsafeID)
	assertBootRetired(t, root, rec, "run-control")
}
