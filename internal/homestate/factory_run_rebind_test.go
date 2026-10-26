package homestate

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

// factory_run_rebind_test.go — SPEC-FACTORY-COMPLETION-RECOVERY-001 M2-B
// (card t1963): the run-consolidation operation (REQ-FCR-020), its guards
// (AC-FCR-021), its atomicity (AC-FCR-020), its concurrency refusals
// (AC-FCR-019 b/c/d), and the worker-liveness predicate (AC-FCR-024).
//
// Every fixture is an isolated sandbox factory (openSandboxFactory): the live
// factory database is never opened.

// Fixture pids. The classifier and the worker probe are keyed on them, so a
// test states ownership and worker liveness without a real process.
const (
	rbPIDDead          = 20 // owner classifies dead
	rbPIDLive          = 21 // owner classifies live
	rbPIDThirdLive     = 22
	rbPIDIndeterminate = 23
	rbPIDWorkerDead    = 31 // worker probes Dead
	rbPIDWorkerLive    = 32 // worker probes Live
	rbPIDWorkerUnknown = 33 // worker probe cannot decide
)

// rbOpts is the injected classification: owner pids map to a class (anything
// unlisted is dead), worker pids map to a probe answer (anything unlisted is
// Live, the safe direction).
func rbOpts() ReconcileOptions {
	return ReconcileOptions{
		Classify: func(pid int, _ string) OwnerClassification {
			switch pid {
			case rbPIDLive, rbPIDThirdLive:
				return OwnerLive
			case rbPIDIndeterminate:
				return OwnerIndeterminate
			}
			return OwnerDead
		},
		WorkerProbe: func(pid int) ProcessIdentityState {
			switch pid {
			case rbPIDWorkerDead:
				return ProcessIdentityDead
			case rbPIDWorkerUnknown:
				return ProcessIdentityIndeterminate
			}
			return ProcessIdentityLive
		},
	}
}

func rbRun(t *testing.T, db *FactoryDB, runID string, pid int) {
	t.Helper()
	mustRecordRun(t, db, FactoryRun{RunID: runID, ManifestJSON: "{}", LeadPID: pid, LeadProcessStart: "fp"})
}

func rbCard(t *testing.T, db *FactoryDB, runID, cardID, state string) {
	t.Helper()
	frPlace(t, db, Card{RunID: runID, CardID: cardID, State: state, Version: 3, Stage: "run", EvidenceSHA: "sha-" + cardID, UpdatedAt: "2026-10-14T00:00:00Z"})
}

func rbWorker(t *testing.T, db *FactoryDB, runID, label string, pid int) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO workers(label,pid,run_id,registered_at,heartbeat_at) VALUES(?,?,?,?,?)`,
		label, pid, runID, "2026-10-14T00:00:00Z", "2026-10-14T00:00:00Z"); err != nil {
		t.Fatalf("insert worker %s: %v", label, err)
	}
}

func rbBind(t *testing.T, db *FactoryDB, cardID, runID string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO card_dispatch(card_id,run_id,recorded_at) VALUES(?,?,?)`, cardID, runID, "2026-10-14T00:00:00Z"); err != nil {
		t.Fatalf("bind %s: %v", cardID, err)
	}
}

func rbApproval(t *testing.T, db *FactoryDB, cardUUID, runID, cardID string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO leader_approvals(card_uuid,run_id,card_id,factory_version,evidence_hash,issuer,issuer_role,issued_at) VALUES(?,?,?,?,?,?,?,?)`,
		cardUUID, runID, cardID, 3, "hash", "leader", "leader", "2026-10-14T00:00:00Z"); err != nil {
		t.Fatalf("insert approval: %v", err)
	}
}

// rbOwner classifies the runs and returns one run's RunOwner as the caller of
// the operation would hold it.
func rbOwner(t *testing.T, db *FactoryDB, opts ReconcileOptions, runID string) RunOwner {
	t.Helper()
	owners, err := db.ClassifyRuns(context.Background(), opts)
	if err != nil {
		t.Fatalf("classify: %v", err)
	}
	for _, o := range owners {
		if o.RunID == runID {
			return o
		}
	}
	t.Fatalf("run %s not found", runID)
	return RunOwner{}
}

// rbDump renders every table the operation may touch, row by row in rowid
// order, so "the dump is unchanged" is a byte comparison and a refused
// operation that committed anything — an event included — shows up.
func rbDump(t *testing.T, db *FactoryDB) string {
	t.Helper()
	var b strings.Builder
	for _, table := range []string{"cards", "card_dispatch", "runs", "events", "leader_approvals", "workers"} {
		b.WriteString(rbDumpOne(t, db, table))
	}
	return b.String()
}

// rbDumpOne renders one table, header included.
func rbDumpOne(t *testing.T, db *FactoryDB, table string) string {
	t.Helper()
	var b strings.Builder
	b.WriteString("## " + table + "\n")
	// SQL: the table name comes from a fixed list in the callers.
	rows, err := db.DB.Query(`SELECT * FROM ` + table + ` ORDER BY rowid`)
	if err != nil {
		t.Fatalf("dump %s: %v", table, err)
	}
	cols, err := rows.Columns()
	if err != nil {
		t.Fatalf("dump %s columns: %v", table, err)
	}
	for rows.Next() {
		vals := make([]any, len(cols))
		for i := range vals {
			vals[i] = new(sql.NullString)
		}
		if err := rows.Scan(vals...); err != nil {
			t.Fatalf("dump %s scan: %v", table, err)
		}
		parts := make([]string, len(vals))
		for i, v := range vals {
			parts[i] = v.(*sql.NullString).String
		}
		b.WriteString(strings.Join(parts, "|") + "\n")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("dump %s rows: %v", table, err)
	}
	_ = rows.Close()
	return b.String()
}

func rbCount(t *testing.T, db *FactoryDB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := db.DB.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("count %q: %v", query, err)
	}
	return n
}

func rbBound(t *testing.T, db *FactoryDB, cardID string) string {
	t.Helper()
	var run string
	if err := db.DB.QueryRow(`SELECT run_id FROM card_dispatch WHERE card_id=?`, cardID).Scan(&run); err != nil {
		t.Fatalf("binding of %s: %v", cardID, err)
	}
	return run
}

// rbBase seeds the AC-FCR-017 shape: a dead source run holding c1(run),
// c2(picked), c4(needs-decision) plus the terminal c5(done), a live target
// holding c3(assigned) and a duplicate c4(picked), three dispatch bindings
// (the abandoned c4 included), and one approval row under the source.
func rbBase(t *testing.T) (*FactoryDB, ReconcileOptions) {
	t.Helper()
	db := openSandboxFactory(t)
	rbRun(t, db, "run-old", rbPIDDead)
	rbRun(t, db, "run-new", rbPIDLive)
	rbCard(t, db, "run-old", "c1", CardRun)
	rbCard(t, db, "run-old", "c2", CardPicked)
	rbCard(t, db, "run-old", "c4", CardNeedsDecision)
	rbCard(t, db, "run-old", "c5", CardDone)
	rbCard(t, db, "run-new", "c3", CardAssigned)
	rbCard(t, db, "run-new", "c4", CardPicked)
	rbBind(t, db, "c1", "run-old")
	rbBind(t, db, "c2", "run-old")
	rbBind(t, db, "c4", "run-old")
	// The terminal c5 stays in the retired source run, so its binding must keep
	// naming that run (REQ-FCR-020 (5), SPEC v0.3.3): repointing it would leave a
	// binding whose target run holds no row for the card.
	rbBind(t, db, "c5", "run-old")
	rbApproval(t, db, "uuid-c1", "run-old", "c1")
	return db, rbOpts()
}

func rbRequest(t *testing.T, db *FactoryDB, opts ReconcileOptions, stamp bool) RebindRequest {
	t.Helper()
	return RebindRequest{Source: rbOwner(t, db, opts, "run-old"), Target: rbOwner(t, db, opts, "run-new"), RequireStampBasis: stamp}
}

func rbRefusedWith(t *testing.T, err error, want ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("operation succeeded, want a refusal naming %v", want)
	}
	if !errors.Is(err, ErrRebindRefused) {
		t.Fatalf("err = %v, want ErrRebindRefused", err)
	}
	for _, w := range want {
		if !strings.Contains(err.Error(), w) {
			t.Fatalf("refusal %q does not name %q", err.Error(), w)
		}
	}
}

// AC-FCR-017 at the operation level (REQ-FCR-020, REQ-FCR-023): the source's
// non-terminal cards are re-keyed (or abandoned with a pointer when the
// target holds the card), every binding that named the source moves — the
// abandoned card's included — and the source retires in the same commit.
func TestRebindMovesAbandonsAndRetires(t *testing.T) {
	db, opts := rbBase(t)
	before := map[string]Card{}
	for _, id := range []string{"c1", "c2"} {
		c, err := db.LoadCard(context.Background(), "run-old", id)
		if err != nil {
			t.Fatal(err)
		}
		before[id] = c
	}
	c3Before, err := db.LoadCard(context.Background(), "run-new", "c3")
	if err != nil {
		t.Fatal(err)
	}
	c4TargetBefore, err := db.LoadCard(context.Background(), "run-new", "c4")
	if err != nil {
		t.Fatal(err)
	}
	cardsBefore := rbCount(t, db, `SELECT count(*) FROM cards`)
	eventsBefore := rbCount(t, db, `SELECT count(*) FROM events`)
	approvalsBefore := rbDumpTable(t, db, "leader_approvals")

	res, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts)
	if err != nil {
		t.Fatalf("rebind: %v", err)
	}
	if !slices.Equal(res.Moved, []string{"c1", "c2"}) || !slices.Equal(res.Abandoned, []string{"c4"}) {
		t.Fatalf("result = %+v, want moved [c1 c2] abandoned [c4]", res)
	}
	for id, was := range before {
		now, err := db.LoadCard(context.Background(), "run-new", id)
		if err != nil {
			t.Fatalf("re-keyed %s not under run-new: %v", id, err)
		}
		was.RunID = "run-new"
		if now != was {
			t.Errorf("re-keyed %s changed beyond run_id:\n got %+v\nwant %+v", id, now, was)
		}
		if _, err := db.LoadCard(context.Background(), "run-old", id); !errors.Is(err, ErrCardNotFound) {
			t.Errorf("%s still under run-old after re-key (err=%v)", id, err)
		}
	}
	old4, err := db.LoadCard(context.Background(), "run-old", "c4")
	if err != nil {
		t.Fatal(err)
	}
	if old4.State != CardAbandoned {
		t.Errorf("run-old c4 state = %s, want abandoned", old4.State)
	}
	if got, _ := db.LoadCard(context.Background(), "run-new", "c4"); got != c4TargetBefore {
		t.Errorf("target c4 changed: %+v want %+v", got, c4TargetBefore)
	}
	if got, _ := db.LoadCard(context.Background(), "run-new", "c3"); got != c3Before {
		t.Errorf("target c3 changed: %+v want %+v", got, c3Before)
	}
	if n := rbCount(t, db, `SELECT count(*) FROM cards`); n != cardsBefore {
		t.Errorf("cards rows = %d, want %d (no delete, no copy)", n, cardsBefore)
	}
	if n := rbCount(t, db, `SELECT count(*) FROM events`); n < eventsBefore {
		t.Errorf("events rows shrank: %d < %d", n, eventsBefore)
	}
	// Every binding of a card the operation moved — re-keyed or abandoned with a
	// pointer (debt R10) — follows it to the target run.
	for _, id := range []string{"c1", "c2", "c4"} {
		if got := rbBound(t, db, id); got != "run-new" {
			t.Errorf("card_dispatch[%s] = %s, want run-new", id, got)
		}
	}
	// The terminal c5 stayed in the retired source run: its binding is
	// UNCHANGED and still resolves to that run's row. Repointing every binding
	// that named the source would send it to a run holding no row for c5, and
	// the completion gate's row resolution would fail closed (SPEC v0.3.3).
	if got := rbBound(t, db, "c5"); got != "run-old" {
		t.Errorf("card_dispatch[c5] = %s, want it unchanged at run-old", got)
	}
	if row, linked, err := recordedCardRow(context.Background(), db.DB, "c5"); err != nil || !linked || row.RunID != "run-old" || row.State != CardDone {
		t.Errorf("recordedCardRow(c5) = (%+v, linked=%v, err=%v), want the run-old done row", row, linked, err)
	}
	if got := runStatus(t, db, "run-old"); got != "retired" {
		t.Errorf("source status = %s, want retired", got)
	}
	if got := runStatus(t, db, "run-new"); got != "active" {
		t.Errorf("target status = %s, want active", got)
	}
	if got := rbDumpTable(t, db, "leader_approvals"); got != approvalsBefore {
		t.Errorf("leader_approvals changed:\n got %s\nwant %s", got, approvalsBefore)
	}

	rebound := frEvents(t, db, "card.rebound")
	if len(rebound) != 1 {
		t.Fatalf("card.rebound events = %d, want 1: %v", len(rebound), rebound)
	}
	var rp map[string]any
	if err := json.Unmarshal([]byte(rebound[0]), &rp); err != nil {
		t.Fatal(err)
	}
	for k, want := range map[string]any{"card_id": "c4", "from_run": "run-old", "to_run": "run-new", "state": CardNeedsDecision, "version": float64(3), "stage": "run", "evidence_sha": "sha-c4"} {
		if rp[k] != want {
			t.Errorf("card.rebound[%s] = %v, want %v (payload %v)", k, rp[k], want, rp)
		}
	}
	retired := frEvents(t, db, "run.retired")
	if len(retired) != 1 {
		t.Fatalf("run.retired events = %d, want 1: %v", len(retired), retired)
	}
	var tp struct {
		Classification string   `json:"classification"`
		Basis          string   `json:"basis"`
		ReboundTo      string   `json:"rebound_to"`
		Moved          int      `json:"moved"`
		Abandoned      int      `json:"abandoned"`
		MovedCards     []string `json:"moved_cards"`
		AbandonedCards []string `json:"abandoned_cards"`
	}
	if err := json.Unmarshal([]byte(retired[0]), &tp); err != nil {
		t.Fatal(err)
	}
	if tp.Classification != "dead" || tp.Basis != "stamp" || tp.ReboundTo != "run-new" || tp.Moved != 2 || tp.Abandoned != 1 ||
		!slices.Equal(tp.MovedCards, []string{"c1", "c2"}) || !slices.Equal(tp.AbandonedCards, []string{"c4"}) {
		t.Errorf("run.retired payload = %+v", tp)
	}
}

func rbDumpTable(t *testing.T, db *FactoryDB, table string) string {
	t.Helper()
	return rbDumpOne(t, db, table)
}

// AC-FCR-020 — an injected failure at any of the three points rolls the whole
// operation back: the dump is byte-identical to the pre-operation dump. The
// points are placed so every split of the writes into two transactions leaves
// the earlier half committed at one of them.
func TestRebindRetireIsAtomic(t *testing.T) {
	for _, point := range []string{rebindPointSecondCard, rebindPointAfterCards, rebindPointBeforeCommit} {
		t.Run(point, func(t *testing.T) {
			db := openSandboxFactory(t)
			opts := rbOpts()
			rbRun(t, db, "run-old", rbPIDDead)
			rbRun(t, db, "run-new", rbPIDLive)
			for _, id := range []string{"c1", "c2", "c3"} {
				rbCard(t, db, "run-old", id, CardRun)
				rbBind(t, db, id, "run-old")
			}
			before := rbDump(t, db)
			req := rbRequest(t, db, opts, true)

			boom := errors.New("injected failure at " + point)
			fired := 0
			rebindFault = func(p string) error {
				if p == point {
					fired++
					return boom
				}
				return nil
			}
			t.Cleanup(func() { rebindFault = nil })

			_, err := db.RebindRun(context.Background(), req, opts)
			if !errors.Is(err, boom) {
				t.Fatalf("err = %v, want the injected failure", err)
			}
			if fired != 1 {
				t.Fatalf("seam %s fired %d times, want 1", point, fired)
			}
			if after := rbDump(t, db); after != before {
				t.Fatalf("dump changed after the failed operation at %s:\n--- before\n%s\n--- after\n%s", point, before, after)
			}

			// Control: the same request, no fault, succeeds — the refusal above
			// was the injected failure and not a fixture defect.
			rebindFault = nil
			if _, err := db.RebindRun(context.Background(), req, opts); err != nil {
				t.Fatalf("control run without the fault: %v", err)
			}
		})
	}
}

// AC-FCR-021 — a card held by another run whose owner is not dead refuses the
// whole operation. The conflicting card is probed in both sort positions: first
// and not first (debt R12 / D2), so a check that skips the first card or only
// checks the first card both fail.
func TestRebindRefusesCardHeldByAnotherLiveRun(t *testing.T) {
	cases := []struct {
		name     string
		conflict string // the card the third run also holds
		third    int    // the third run's owner pid
	}{
		{"live third run, conflict not first", "c3", rbPIDThirdLive},
		{"live third run, conflict first", "c1", rbPIDThirdLive},
		{"indeterminate third run, conflict not first", "c3", rbPIDIndeterminate},
		{"indeterminate third run, conflict first", "c1", rbPIDIndeterminate},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := openSandboxFactory(t)
			opts := rbOpts()
			rbRun(t, db, "run-old", rbPIDDead)
			rbRun(t, db, "run-new", rbPIDLive)
			rbRun(t, db, "run-third", tc.third)
			for _, id := range []string{"c1", "c2", "c3"} {
				rbCard(t, db, "run-old", id, CardRun)
				rbBind(t, db, id, "run-old")
			}
			rbCard(t, db, "run-third", tc.conflict, CardAssigned)
			before := rbDump(t, db)
			_, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts)
			rbRefusedWith(t, err, tc.conflict, "run-third")
			if after := rbDump(t, db); after != before {
				t.Fatalf("dump changed after the refusal:\n--- before\n%s\n--- after\n%s", before, after)
			}
		})
	}
}

// A third run whose owner is dead does not block: its row is not a live claim.
func TestRebindIgnoresCardHeldByADeadThirdRun(t *testing.T) {
	db := openSandboxFactory(t)
	opts := rbOpts()
	rbRun(t, db, "run-old", rbPIDDead)
	rbRun(t, db, "run-new", rbPIDLive)
	rbRun(t, db, "run-third", rbPIDDead)
	rbCard(t, db, "run-old", "c1", CardRun)
	rbCard(t, db, "run-third", "c1", CardAssigned)
	if _, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, true), opts); err != nil {
		t.Fatalf("rebind beside a dead third run: %v", err)
	}
}

// AC-FCR-021 (c) — a live worker on the source refuses the operation. The
// worker rows are probed in both orders (debt R11): a check that only looks at
// the first or only at the last row fails one of them.
func TestRebindRefusesSourceWithLiveWorker(t *testing.T) {
	for _, tc := range []struct {
		name string
		pids [2]int
	}{
		{"dead worker first", [2]int{rbPIDWorkerDead, rbPIDWorkerLive}},
		{"live worker first", [2]int{rbPIDWorkerLive, rbPIDWorkerDead}},
		{"undecided worker", [2]int{rbPIDWorkerDead, rbPIDWorkerUnknown}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := openSandboxFactory(t)
			opts := rbOpts()
			rbRun(t, db, "run-old", rbPIDDead)
			rbRun(t, db, "run-new", rbPIDLive)
			rbCard(t, db, "run-old", "c1", CardRun)
			rbWorker(t, db, "run-old", "lane-a", tc.pids[0])
			rbWorker(t, db, "run-old", "lane-b", tc.pids[1])
			before := rbDump(t, db)
			_, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, false), opts)
			wantLabel := "lane-b"
			if tc.pids[0] != rbPIDWorkerDead {
				wantLabel = "lane-a"
			}
			rbRefusedWith(t, err, wantLabel)
			if after := rbDump(t, db); after != before {
				t.Fatalf("dump changed after the refusal")
			}
		})
	}
}

// AC-FCR-021 (d)(e) — step (0): a source whose owner is not dead is refused on
// both entry shapes (the operator path and the leader-entry path), naming the
// run and its classification.
func TestRebindRefusesSourceWhoseOwnerIsNotDead(t *testing.T) {
	for _, tc := range []struct {
		pid  int
		want string
	}{{rbPIDThirdLive, "live"}, {rbPIDIndeterminate, "indeterminate"}} {
		for _, stamp := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stamp=%v", tc.want, stamp), func(t *testing.T) {
				db := openSandboxFactory(t)
				opts := rbOpts()
				rbRun(t, db, "run-old", tc.pid)
				rbRun(t, db, "run-new", rbPIDLive)
				rbCard(t, db, "run-old", "c1", CardRun)
				before := rbDump(t, db)
				_, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, stamp), opts)
				rbRefusedWith(t, err, "run-old", tc.want)
				if !errors.Is(err, ErrRunOwnerNotDead) {
					t.Errorf("err = %v, want it to also match ErrRunOwnerNotDead", err)
				}
				if after := rbDump(t, db); after != before {
					t.Fatalf("dump changed after the refusal")
				}
			})
		}
	}
}

// The leader-entry path additionally requires the source to be dead on its OWN
// stamp: a run classified dead through the peer or boot proof is refused there
// and accepted on the operator path (REQ-FCR-020 step 0).
func TestRebindStampBasisRequirement(t *testing.T) {
	for _, basis := range []ProofBasis{BasisPeer, BasisBoot} {
		t.Run(string(basis), func(t *testing.T) {
			db, opts := rbBase(t)
			req := rbRequest(t, db, opts, true)
			req.Source.Basis = basis
			before := rbDump(t, db)
			_, err := db.RebindRun(context.Background(), req, opts)
			rbRefusedWith(t, err, "run-old", string(basis))
			if after := rbDump(t, db); after != before {
				t.Fatalf("dump changed after the refusal")
			}
			req.RequireStampBasis = false
			if _, err := db.RebindRun(context.Background(), req, opts); err != nil {
				t.Fatalf("operator path with basis %s: %v", basis, err)
			}
		})
	}
}

// The target must classify live; a dead or unknown target is refused.
func TestRebindRefusesTargetThatIsNotLive(t *testing.T) {
	db := openSandboxFactory(t)
	opts := rbOpts()
	rbRun(t, db, "run-old", rbPIDDead)
	rbRun(t, db, "run-new", rbPIDDead)
	rbCard(t, db, "run-old", "c1", CardRun)
	before := rbDump(t, db)
	_, err := db.RebindRun(context.Background(), rbRequest(t, db, opts, false), opts)
	rbRefusedWith(t, err, "run-new", "dead")
	if after := rbDump(t, db); after != before {
		t.Fatalf("dump changed after the refusal")
	}
}

// AC-FCR-019 (b)(c)(d) — concurrent leaders. The classified stamp of the
// target (b) or the source (d) changed after classification: the operation is
// refused and the dump (events included) is unchanged. A second attempt on an
// already consolidated source is refused and leaves one run.retired event (c).
func TestRunConsolidationConcurrentLeaders(t *testing.T) {
	t.Run("b target restamped after classification", func(t *testing.T) {
		db, opts := rbBase(t)
		req := rbRequest(t, db, opts, true)
		if _, err := db.DB.Exec(`UPDATE runs SET lead_pid=?, lead_process_start='other-proc' WHERE run_id='run-new'`, rbPIDLive); err != nil {
			t.Fatal(err)
		}
		before := rbDump(t, db)
		_, err := db.RebindRun(context.Background(), req, opts)
		rbRefusedWith(t, err, "run-new")
		if after := rbDump(t, db); after != before {
			t.Fatalf("dump changed after the refusal")
		}
	})
	t.Run("c second attempt on a consolidated source", func(t *testing.T) {
		db, opts := rbBase(t)
		req := rbRequest(t, db, opts, true)
		if _, err := db.RebindRun(context.Background(), req, opts); err != nil {
			t.Fatalf("first attempt: %v", err)
		}
		_, err := db.RebindRun(context.Background(), req, opts)
		rbRefusedWith(t, err, "run-old")
		if n := len(frEvents(t, db, "run.retired")); n != 1 {
			t.Fatalf("run.retired events = %d, want 1", n)
		}
		if n := len(frEvents(t, db, "card.rebound")); n != 1 {
			t.Fatalf("card.rebound events = %d, want 1", n)
		}
	})
	t.Run("d source adopted by another leader after classification", func(t *testing.T) {
		db, opts := rbBase(t)
		req := rbRequest(t, db, opts, true)
		if _, err := db.DB.Exec(`UPDATE runs SET lead_pid=?, lead_process_start='adopter' WHERE run_id='run-old'`, rbPIDThirdLive); err != nil {
			t.Fatal(err)
		}
		before := rbDump(t, db)
		_, err := db.RebindRun(context.Background(), req, opts)
		rbRefusedWith(t, err, "run-old")
		if after := rbDump(t, db); after != before {
			t.Fatalf("dump changed after the refusal")
		}
		if n := rbCount(t, db, `SELECT count(*) FROM cards WHERE run_id='run-new'`); n != 2 {
			t.Fatalf("target cards = %d, want its own 2 untouched", n)
		}
	})
}
