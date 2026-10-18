package homestate

// factory_run_orphan_test.go — SPEC-FACTORY-LEADER-ADOPT-CORE-001 M1: the
// read-only orphan finder behind the leader-start refusal (AC-FLAC-001). An
// orphan run is an ACTIVE run whose owner is classified dead on the run row's
// OWN stamp (the proof basis the adoption's compare-and-set needs) and which
// holds at least one non-terminal card. Boot-proof and lead-peer
// classifications carry no row-level identity to adopt against and are
// excluded; workers-only and terminal-cards-only runs are not orphans.

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// orphanFixtureRunIDs are the run ids the AC-FLAC-001 fixture seeds, one per
// qualifying/excluded shape the acceptance criterion enumerates.
const (
	orphanRunA   = "orphan-a"  // dead stamp, 2 non-terminal + 1 terminal card → orphan
	orphanRunA2  = "orphan-a2" // dead stamp, 1 non-terminal card → orphan
	orphanRunB   = "orphan-b"  // dead stamp, terminal cards only → excluded
	orphanRunC   = "orphan-c"  // dead stamp, workers only, no cards → excluded
	orphanRunD   = "orphan-d"  // live owner, 1 non-terminal card → excluded
	orphanRunE   = "orphan-e"  // peer-classified dead, 1 non-terminal card → excluded
	orphanRunF   = "orphan-f"  // indeterminate owner, 1 non-terminal card → excluded
	orphanRunG   = "orphan-g"  // dead stamp, retired status, 1 non-terminal card → excluded
	orphanRunH   = "orphan-h"  // no identity, boot-proof dead, 1 non-terminal card → excluded
	orphanStampT = "2026-10-14T00:00:00Z"
)

// orphanFixtureClassifier answers the AC-FLAC-001 per-pid classification:
// 101 dead, 102 dead, 103 dead, 104 live, 106 indeterminate, 107 dead,
// 108 dead, 42 dead.
func orphanFixtureClassifier(pid int, _ string) OwnerClassification {
	switch pid {
	case 101, 102, 103, 107, 108, 42:
		return OwnerDead
	case 104:
		return OwnerLive
	case 106:
		return OwnerIndeterminate
	default:
		return OwnerIndeterminate
	}
}

// mustRecordRunAtSeed records a run and then pins its row and run.started
// event timestamps to a fixed value, so a fixture can place a run's whole
// recorded activity before an injected boot time without touching the host
// clock.
func mustRecordRunAtSeed(t *testing.T, db *FactoryDB, run FactoryRun, seededAt string) {
	t.Helper()
	mustRecordRun(t, db, run)
	if seededAt == "" {
		return
	}
	if _, err := db.DB.Exec(`UPDATE runs SET created_at=?, updated_at=? WHERE run_id=?`, seededAt, seededAt, run.RunID); err != nil {
		t.Fatalf("pin run %s timestamps: %v", run.RunID, err)
	}
	if _, err := db.DB.Exec(`UPDATE events SET created_at=? WHERE run_id=? AND kind='run.started'`, seededAt, run.RunID); err != nil {
		t.Fatalf("pin run %s event timestamps: %v", run.RunID, err)
	}
}

// seedOrphanFixture builds the full AC-FLAC-001 database: nine runs covering
// the two qualifying orphans and every excluded shape. Run E resolves its
// owner through the Fallback (basis peer); run H has no identity at all and
// its recorded activity predates the injected boot time (basis boot).
func seedOrphanFixture(t *testing.T, db *FactoryDB) {
	t.Helper()
	seededAt := orphanStampT
	mustRecordRunAtSeed(t, db, FactoryRun{RunID: orphanRunA, ManifestJSON: "{}", LeadPID: 101, LeadProcessStart: "gone-a"}, seededAt)
	mustInsertCard(t, db, orphanRunA, "a-1", "run")
	mustInsertCard(t, db, orphanRunA, "a-2", "blocked")
	mustInsertCard(t, db, orphanRunA, "a-3", "done")

	mustRecordRunAtSeed(t, db, FactoryRun{RunID: orphanRunA2, ManifestJSON: "{}", LeadPID: 108, LeadProcessStart: "gone-a2"}, "2026-10-17T09:15:00Z")
	mustInsertCard(t, db, orphanRunA2, "a2-1", "blocked")

	mustRecordRunAtSeed(t, db, FactoryRun{RunID: orphanRunB, ManifestJSON: "{}", LeadPID: 102, LeadProcessStart: "gone-b"}, seededAt)
	mustInsertCard(t, db, orphanRunB, "b-1", "done")
	mustInsertCard(t, db, orphanRunB, "b-2", "failed")
	mustInsertCard(t, db, orphanRunB, "b-3", "abandoned")

	mustRecordRunAtSeed(t, db, FactoryRun{RunID: orphanRunC, ManifestJSON: "{}", LeadPID: 103, LeadProcessStart: "gone-c"}, seededAt)
	mustInsertWorker(t, db, orphanRunC, "lane-2")

	mustRecordRunAtSeed(t, db, FactoryRun{RunID: orphanRunD, ManifestJSON: "{}", LeadPID: 104, LeadProcessStart: "alive-d"}, seededAt)
	mustInsertCard(t, db, orphanRunD, "d-1", "run")

	mustRecordRunAtSeed(t, db, FactoryRun{RunID: orphanRunE, ManifestJSON: "{}", LeadPID: 0, LeadProcessStart: ""}, seededAt)
	mustInsertCard(t, db, orphanRunE, "e-1", "run")

	mustRecordRunAtSeed(t, db, FactoryRun{RunID: orphanRunF, ManifestJSON: "{}", LeadPID: 106, LeadProcessStart: "unknown-f"}, seededAt)
	mustInsertCard(t, db, orphanRunF, "f-1", "run")

	mustRecordRunAtSeed(t, db, FactoryRun{RunID: orphanRunG, ManifestJSON: "{}", LeadPID: 107, LeadProcessStart: "gone-g"}, seededAt)
	mustInsertCard(t, db, orphanRunG, "g-1", "run")
	if _, err := db.DB.Exec(`UPDATE runs SET status='retired' WHERE run_id=?`, orphanRunG); err != nil {
		t.Fatalf("retire run g: %v", err)
	}

	mustRecordRunAtSeed(t, db, FactoryRun{RunID: orphanRunH, ManifestJSON: "{}", LeadPID: 0, LeadProcessStart: ""}, seededAt)
	mustInsertCard(t, db, orphanRunH, "h-1", "run")
}

// orphanFixtureOptions returns the ReconcileOptions the AC-FLAC-001 fixture
// runs under: the per-pid classifier, a fallback that resolves run E only,
// and the two boot-proof premises with a boot time after every seeded
// timestamp (so run H classifies dead on the boot basis without touching the
// host clock).
func orphanFixtureOptions() ReconcileOptions {
	return ReconcileOptions{
		Classify: orphanFixtureClassifier,
		Fallback: func(runID string) (int, string, bool) {
			if runID == orphanRunE {
				return 42, "peer-start", true
			}
			return 0, "", false
		},
		BootTime:               func() (time.Time, bool) { return time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), true },
		LegacyPeerRecordAbsent: func(string) bool { return true },
	}
}

// orphanContentSnapshot renders every row of the four tables the finder must
// not write, ordered, as one string: equal snapshots before and after a
// FindOrphanRuns call prove the finder wrote nothing.
func orphanContentSnapshot(t *testing.T, db *FactoryDB) string {
	t.Helper()
	var b strings.Builder
	for _, table := range []struct {
		name  string
		cols  int
		query string
	}{
		{"runs", 10, `SELECT run_id,lead_session_id,lead_backend,status,manifest_json,lead_pid,lead_process_start,lane_capacity,created_at,updated_at FROM runs ORDER BY run_id`},
		{"cards", 8, `SELECT run_id,card_id,owner_label,state,version,evidence_path,updated_at,stage FROM cards ORDER BY run_id,card_id`},
		{"events", 5, `SELECT seq,run_id,kind,payload_json,created_at FROM events ORDER BY seq`},
		{"workers", 7, `SELECT label,pid,backend,session_id,run_id,registered_at,heartbeat_at FROM workers ORDER BY label`},
	} {
		b.WriteString("== " + table.name + "\n")
		rows, err := db.DB.Query(table.query)
		if err != nil {
			t.Fatalf("snapshot %s: %v", table.name, err)
		}
		for rows.Next() {
			vals := make([]any, table.cols)
			ptrs := make([]any, table.cols)
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				t.Fatalf("snapshot scan %s: %v", table.name, err)
			}
			for _, v := range vals {
				if raw, ok := v.([]byte); ok {
					b.WriteString(strings.ReplaceAll(string(raw), "\n", " "))
				} else {
					b.WriteString(strings.ReplaceAll(fmt.Sprintf("%v", v), "\n", " "))
				}
				b.WriteString("|")
			}
			b.WriteString("\n")
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("snapshot %s rows: %v", table.name, err)
		}
		_ = rows.Close()
	}
	return b.String()
}

// AC-FLAC-001 — the finder returns exactly the two qualifying orphans with
// their non-terminal card counts, and writes nothing.
func TestFindOrphanRunsRequiresDeadStampAndCards(t *testing.T) {
	db := openSandboxFactory(t)
	seedOrphanFixture(t, db)
	before := orphanContentSnapshot(t, db)

	orphans, err := db.FindOrphanRuns(context.Background(), orphanFixtureOptions())
	if err != nil {
		t.Fatalf("find orphan runs: %v", err)
	}
	if len(orphans) != 2 {
		t.Fatalf("orphans = %+v, want exactly %s and %s", orphans, orphanRunA, orphanRunA2)
	}
	if orphans[0].RunID != orphanRunA || orphans[0].NonTerminalCards != 2 {
		t.Fatalf("orphans[0] = (%s, %d), want (%s, 2)", orphans[0].RunID, orphans[0].NonTerminalCards, orphanRunA)
	}
	if orphans[1].RunID != orphanRunA2 || orphans[1].NonTerminalCards != 1 {
		t.Fatalf("orphans[1] = (%s, %d), want (%s, 1)", orphans[1].RunID, orphans[1].NonTerminalCards, orphanRunA2)
	}

	after := orphanContentSnapshot(t, db)
	if before != after {
		t.Fatalf("the finder wrote to the database:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// AC-FLAC-001 — every excluded shape stays excluded: terminal-cards-only,
// workers-only, live owner, peer-classified, indeterminate owner, retired
// status, and boot-proof classified runs are never orphans.
func TestFindOrphanRunsExcludesNonQualifyingRuns(t *testing.T) {
	db := openSandboxFactory(t)
	seedOrphanFixture(t, db)
	before := orphanContentSnapshot(t, db)

	orphans, err := db.FindOrphanRuns(context.Background(), orphanFixtureOptions())
	if err != nil {
		t.Fatalf("find orphan runs: %v", err)
	}
	seen := map[string]int{}
	for _, o := range orphans {
		seen[o.RunID] = o.NonTerminalCards
	}
	for _, excluded := range []string{orphanRunB, orphanRunC, orphanRunD, orphanRunE, orphanRunF, orphanRunG, orphanRunH} {
		if _, ok := seen[excluded]; ok {
			t.Fatalf("orphan list contains excluded run %s (all: %v)", excluded, seen)
		}
	}

	after := orphanContentSnapshot(t, db)
	if before != after {
		t.Fatalf("the finder wrote to the database:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// AC-FLAC-001 — the count is the non-terminal card count only: a terminal
// card does not count (the done card on run A would make the count 3), a
// registered worker is not a card, and terminal-only or workers-only runs
// never qualify.
func TestFindOrphanRunsCountsNonTerminalCardsOnly(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "orphan-a", ManifestJSON: "{}", LeadPID: 101, LeadProcessStart: "gone-a"})
	mustInsertCard(t, db, "orphan-a", "a-1", "run")
	mustInsertCard(t, db, "orphan-a", "a-2", "blocked")
	mustInsertCard(t, db, "orphan-a", "a-3", "done")
	mustRecordRun(t, db, FactoryRun{RunID: "orphan-b", ManifestJSON: "{}", LeadPID: 102, LeadProcessStart: "gone-b"})
	mustInsertCard(t, db, "orphan-b", "b-1", "done")
	mustRecordRun(t, db, FactoryRun{RunID: "orphan-c", ManifestJSON: "{}", LeadPID: 103, LeadProcessStart: "gone-c"})
	mustInsertWorker(t, db, "orphan-c", "lane-2")
	before := orphanContentSnapshot(t, db)

	opts := ReconcileOptions{Classify: orphanFixtureClassifier}
	orphans, err := db.FindOrphanRuns(context.Background(), opts)
	if err != nil {
		t.Fatalf("find orphan runs: %v", err)
	}
	if len(orphans) != 1 {
		t.Fatalf("orphans = %+v, want exactly orphan-a (terminal-only and workers-only runs are not orphans)", orphans)
	}
	if orphans[0].RunID != "orphan-a" || orphans[0].NonTerminalCards != 2 {
		t.Fatalf("orphan = (%s, %d), want (orphan-a, 2) — the done card must not count", orphans[0].RunID, orphans[0].NonTerminalCards)
	}

	after := orphanContentSnapshot(t, db)
	if before != after {
		t.Fatalf("the finder wrote to the database:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}
