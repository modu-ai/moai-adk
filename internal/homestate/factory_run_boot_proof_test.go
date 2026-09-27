package homestate

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

// operatorBoot is the host boot instant measured on the machine that
// reproduced card t1168 (`sysctl -n kern.boottime` → sec = 1789526916,
// usec = 954947). The fixture rows below are that machine's factory.db rows,
// copied verbatim.
var operatorBoot = time.Unix(1789526916, 954947000).UTC()

func openBareFactory(t *testing.T) *FactoryDB {
	t.Helper()
	db, err := OpenFactoryPath(filepath.Join(t.TempDir(), "factory.db"))
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func mustExec(t *testing.T, db *FactoryDB, query string, args ...any) {
	t.Helper()
	if _, err := db.DB.Exec(query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

// seedOperatorLegacyRows inserts the five rows that blocked `moai glm -f agent`
// with AMBIGUOUS_FACTORY: no owner stamp (lead_pid 0, empty start), no broker
// peer, every recorded timestamp days before the host's last boot.
func seedOperatorLegacyRows(t *testing.T, db *FactoryDB) {
	t.Helper()
	const manifest = `{"spec_id":"","spec_path":"","spec_sha256":"","git_commit":"2213871afb7d655f411c46a34fd38bb8153278fc","captured_at":"2026-09-11T10:25:37.588176Z"}`
	rows := []struct{ id, backend, manifest, at string }{
		{"glm", "glm", manifest, "2026-09-11T10:25:37.923062Z"},
		{"tl4rkl", "", "{}", "2026-09-10T18:28:19.725674Z"},
		{"tl8stp", "claude", manifest, "2026-09-12T08:27:26.329341Z"},
		{"tl8v4m", "glm", manifest, "2026-09-12T09:17:10.953367Z"},
		{"tl8xsp", "gpt", manifest, "2026-09-12T10:14:50.001912Z"},
	}
	for _, r := range rows {
		mustExec(t, db, `INSERT INTO runs(run_id,lead_session_id,lead_backend,status,manifest_json,created_at,updated_at,lead_pid,lead_process_start) VALUES(?,'',?,'active',?,?,?,0,'')`, r.id, r.backend, r.manifest, r.at, r.at)
		mustExec(t, db, `INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,'run.started','{}',?)`, r.id, r.at)
	}
	// The latest recorded activity of the set: tl8v4m's card completion.
	mustExec(t, db, `INSERT INTO events(run_id,kind,payload_json,created_at) VALUES('tl8v4m','card.completed','{}','2026-09-13T02:28:13.362475Z')`)
}

func operatorBootOpts() ReconcileOptions {
	return ReconcileOptions{
		BootTime:         func() (time.Time, bool) { return operatorBoot, true },
		LeadRecordAbsent: func(string) bool { return true },
	}
}

// card t1168 RED/GREEN — a run with no owner identity from any source whose
// every recorded activity predates the host's last boot has a provably dead
// owner: no process alive now existed before the boot. The five operator rows
// are retired, so they no longer make resolution ambiguous.
func TestReconcileRetiresIdentitylessRunsThatPredateBoot(t *testing.T) {
	db := openBareFactory(t)
	seedOperatorLegacyRows(t, db)

	rec, err := db.ReconcileActiveRuns(context.Background(), operatorBootOpts())
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if len(rec.Remaining) != 0 {
		t.Fatalf("remaining = %+v, want none: every row predates boot", rec.Remaining)
	}
	if len(rec.Retired) != 5 {
		t.Fatalf("retired %d runs, want 5", len(rec.Retired))
	}
	for _, id := range []string{"glm", "tl4rkl", "tl8stp", "tl8v4m", "tl8xsp"} {
		if st := runStatus(t, db, id); st != "retired" {
			t.Fatalf("%s status = %q, want \"retired\"", id, st)
		}
	}
}

// Fail-closed half — each missing piece of the proof keeps the row
// indeterminate and active.
func TestBootProofDeclinesWithoutEveryPremise(t *testing.T) {
	cases := []struct {
		name  string
		opts  ReconcileOptions
		extra func(t *testing.T, db *FactoryDB)
	}{
		{"boot time unknown", ReconcileOptions{
			BootTime:         func() (time.Time, bool) { return time.Time{}, false },
			LeadRecordAbsent: func(string) bool { return true },
		}, nil},
		{"boot probe not wired", ReconcileOptions{LeadRecordAbsent: func(string) bool { return true }}, nil},
		{"lead record may exist", ReconcileOptions{
			BootTime:         func() (time.Time, bool) { return operatorBoot, true },
			LeadRecordAbsent: func(string) bool { return false },
		}, nil},
		{"lead-record check not wired", ReconcileOptions{
			BootTime: func() (time.Time, bool) { return operatorBoot, true },
		}, nil},
		{"event after boot", operatorBootOpts(), func(t *testing.T, db *FactoryDB) {
			mustExec(t, db, `INSERT INTO events(run_id,kind,payload_json,created_at) VALUES('glm','card.completed','{}','2026-09-20T00:00:00Z')`)
		}},
		{"worker heartbeat after boot", operatorBootOpts(), func(t *testing.T, db *FactoryDB) {
			mustExec(t, db, `INSERT INTO workers(label,pid,backend,session_id,run_id,registered_at,heartbeat_at) VALUES('worker-1',1,'glm','','glm','2026-09-11T10:30:00Z','2026-09-20T00:00:00Z')`)
		}},
		{"card updated after boot", operatorBootOpts(), func(t *testing.T, db *FactoryDB) {
			mustExec(t, db, `INSERT INTO cards(run_id,card_id,state,updated_at) VALUES('glm','c1','done','2026-09-20T00:00:00Z')`)
		}},
		{"run row touched after boot", operatorBootOpts(), func(t *testing.T, db *FactoryDB) {
			mustExec(t, db, `UPDATE runs SET updated_at='2026-09-20T00:00:00Z' WHERE run_id='glm'`)
		}},
		{"unparsable timestamp", operatorBootOpts(), func(t *testing.T, db *FactoryDB) {
			mustExec(t, db, `INSERT INTO events(run_id,kind,payload_json,created_at) VALUES('glm','x','{}','not-a-time')`)
		}},
		// Premise 4 demands "strictly earlier": an instant equal to the boot
		// is not proven to predate it.
		{"timestamp equal to boot", operatorBootOpts(), func(t *testing.T, db *FactoryDB) {
			mustExec(t, db, `INSERT INTO events(run_id,kind,payload_json,created_at) VALUES('glm','card.completed','{}',?)`, operatorBoot.Format(time.RFC3339Nano))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := openBareFactory(t)
			seedOperatorLegacyRows(t, db)
			if tc.extra != nil {
				tc.extra(t, db)
			}
			rec, err := db.ReconcileActiveRuns(context.Background(), tc.opts)
			if err != nil {
				t.Fatalf("reconcile: %v", err)
			}
			if st := runStatus(t, db, "glm"); st != "active" {
				t.Fatalf("glm status = %q, want \"active\" (proof premise missing)", st)
			}
			if c := remainingClassification(rec, "glm"); c != OwnerIndeterminate {
				t.Fatalf("glm classification = %q, want %q (proof premise missing)", c, OwnerIndeterminate)
			}
		})
	}
}

// A run carrying an owner identity is judged by that identity alone: the boot
// proof never overrides a live classification, even on a row whose recorded
// timestamps predate boot.
func TestBootProofNeverOverridesAnIdentity(t *testing.T) {
	db := openBareFactory(t)
	mustExec(t, db, `INSERT INTO runs(run_id,lead_session_id,lead_backend,status,manifest_json,created_at,updated_at,lead_pid,lead_process_start) VALUES('run-live','','claude','active','{}','2026-09-01T00:00:00Z','2026-09-01T00:00:00Z',4242,'start-live')`)
	opts := operatorBootOpts()
	opts.Classify = fixedClassifier(OwnerLive)

	rec, err := db.ReconcileActiveRuns(context.Background(), opts)
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if len(rec.Retired) != 0 || runStatus(t, db, "run-live") != "active" {
		t.Fatalf("live-identity run retired by boot proof: %+v", rec.Retired)
	}
}

// The operator surface shares the predicate: `moai factory runs --retire`
// retires a boot-proven row it previously refused as indeterminate.
func TestRetireRunIfDeadAcceptsBootProof(t *testing.T) {
	db := openBareFactory(t)
	seedOperatorLegacyRows(t, db)

	c, err := db.RetireRunIfDead(context.Background(), "tl4rkl", operatorBootOpts())
	if err != nil {
		t.Fatalf("retire: %v (classification %s)", err, c)
	}
	if c != OwnerDead || runStatus(t, db, "tl4rkl") != "retired" {
		t.Fatalf("classification %s status %s, want dead/retired", c, runStatus(t, db, "tl4rkl"))
	}
}

func TestSystemBootTimeIsInThePast(t *testing.T) {
	boot, ok := SystemBootTime()
	if !ok {
		t.Skip("host does not report a boot time; the proof declines here by design")
	}
	if !boot.Before(time.Now()) || boot.Before(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("SystemBootTime = %v, not a plausible past instant", boot)
	}
}

// remainingClassification reads one run's classification from a
// reconciliation's surviving set; a run absent from it reports "".
func remainingClassification(rec Reconciliation, runID string) OwnerClassification {
	for _, o := range rec.Remaining {
		if o.RunID == runID {
			return o.Classification
		}
	}
	return ""
}

// retiredPayload decodes the single run.retired event recorded for a run.
func retiredPayload(t *testing.T, db *FactoryDB, runID string) map[string]string {
	t.Helper()
	rows, err := db.DB.Query(`SELECT payload_json FROM events WHERE run_id=? AND kind='run.retired'`, runID)
	if err != nil {
		t.Fatalf("read run.retired for %s: %v", runID, err)
	}
	defer func() { _ = rows.Close() }()
	var payloads []string
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			t.Fatalf("scan run.retired for %s: %v", runID, err)
		}
		payloads = append(payloads, raw)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate run.retired for %s: %v", runID, err)
	}
	if len(payloads) != 1 {
		t.Fatalf("%s carries %d run.retired events, want 1", runID, len(payloads))
	}
	var got map[string]string
	if err := json.Unmarshal([]byte(payloads[0]), &got); err != nil {
		t.Fatalf("run.retired payload %q for %s is not a JSON object of strings: %v", payloads[0], runID, err)
	}
	return got
}

// AC-020 (REQ-010b) — every run.retired event names the proof that
// established dead. Three runs each die by a different proof; the operator leg
// retires a boot-proven run, because a stamp-dead run there would carry
// "stamp" under a correct implementation and under a broken one alike.
func TestRetiredEventRecordsProofBasis(t *testing.T) {
	const preBoot = "2026-09-01T00:00:00Z"
	t.Run("reconciler", func(t *testing.T) {
		db := openBareFactory(t)
		// Own owner stamp, probed dead.
		mustExec(t, db, `INSERT INTO runs(run_id,lead_session_id,lead_backend,status,manifest_json,created_at,updated_at,lead_pid,lead_process_start) VALUES('run-stamp','','claude','active','{}',?,?,4242,'start-stamp')`, preBoot, preBoot)
		// Unstamped: the role='lead' peer identity, probed dead.
		mustExec(t, db, `INSERT INTO runs(run_id,lead_session_id,lead_backend,status,manifest_json,created_at,updated_at,lead_pid,lead_process_start) VALUES('run-peer','','claude','active','{}',?,?,0,'')`, preBoot, preBoot)
		// No identity from any source: the boot proof.
		mustExec(t, db, `INSERT INTO runs(run_id,lead_session_id,lead_backend,status,manifest_json,created_at,updated_at,lead_pid,lead_process_start) VALUES('run-boot','','claude','active','{}',?,?,0,'')`, preBoot, preBoot)
		opts := operatorBootOpts()
		opts.Classify = fixedClassifier(OwnerDead)
		opts.Fallback = func(runID string) (int, string, bool) {
			if runID == "run-peer" {
				return 5151, "start-peer", true
			}
			return 0, "", false
		}

		rec, err := db.ReconcileActiveRuns(context.Background(), opts)
		if err != nil {
			t.Fatalf("reconcile: %v", err)
		}
		if len(rec.Retired) != 3 {
			t.Fatalf("retired = %+v, want all three runs", rec.Retired)
		}
		for runID, basis := range map[string]string{"run-stamp": "stamp", "run-peer": "peer", "run-boot": "boot"} {
			got := retiredPayload(t, db, runID)
			if got["classification"] != "dead" {
				t.Errorf("%s classification = %q, want \"dead\" (payload %v)", runID, got["classification"], got)
			}
			if got["basis"] != basis {
				t.Errorf("%s basis = %q, want %q (payload %v)", runID, got["basis"], basis, got)
			}
		}
	})
	t.Run("operator retire of a boot-proven run", func(t *testing.T) {
		db := openBareFactory(t)
		seedOperatorLegacyRows(t, db)

		c, err := db.RetireRunIfDead(context.Background(), "tl4rkl", operatorBootOpts())
		if err != nil || c != OwnerDead {
			t.Fatalf("retire = (%s, %v), want dead", c, err)
		}
		got := retiredPayload(t, db, "tl4rkl")
		if got["classification"] != "dead" || got["basis"] != "boot" {
			t.Fatalf("payload = %v, want classification dead and basis boot", got)
		}
	})
}
