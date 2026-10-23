package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// SPEC-CARD-METADATA-VERBS-001 M2: the `moai factory set-spec` verb — flag
// gate, validation gate, two-store write, the cross-store failure matrix
// (including the ABA seed DB-1 carries into run), reason validation, and the
// gate-refusal atomicity sweep.

var mdSpecNow = time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)

// mdEnableFlag writes the local config that turns the metadata verbs on.
func mdEnableFlag(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	body := "workflow:\n  card_metadata_verbs:\n    enabled: true\n"
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// mdQueueSpec pins the spec id a queue item carries (nil clears it).
func mdQueueSpec(t *testing.T, store *factory.BacklogStore, cardID string, spec *string) {
	t.Helper()
	if err := store.Mutate(func(r *factory.BacklogRecord) error {
		for i := range r.Items {
			if r.Items[i].ID == cardID {
				r.Items[i].SpecID = spec
				return nil
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("set queue spec: %v", err)
	}
}

// mdQueueDump renders the queue's live items for a row-level comparison.
func mdQueueDump(t *testing.T, store *factory.BacklogStore) string {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	var b strings.Builder
	for _, it := range rec.Items {
		spec := "<nil>"
		if it.SpecID != nil {
			spec = *it.SpecID
		}
		b.WriteString(it.ID + "|" + string(it.State) + "|" + spec + "\n")
	}
	return b.String()
}

// mdCardsDump renders every factory card row for a row-level comparison
// (acceptance.md §A: store-identity is judged at ROW level).
func mdCardsDump(t *testing.T, root string) string {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	rows, err := db.DB.Query(`SELECT run_id,card_id,owner_label,state,version,evidence_path,updated_at,stage,spec_id,worktree_path FROM cards ORDER BY run_id,card_id`)
	if err != nil {
		t.Fatalf("dump cards: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var b strings.Builder
	for rows.Next() {
		var v [10]string
		ptrs := make([]any, len(v))
		for i := range v {
			ptrs[i] = &v[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatal(err)
		}
		b.WriteString(strings.Join(v[:], "|") + "\n")
	}
	return b.String()
}

// mdAuditRows reads the card_metadata_events rows for one card.
func mdAuditRows(t *testing.T, root, cardID string) []homestate.CardMetadataEvent {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	rows, err := db.DB.Query(`SELECT card_id,kind,payload_json FROM card_metadata_events WHERE card_id=? ORDER BY seq`, cardID)
	if err != nil {
		t.Fatalf("read audit rows: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var out []homestate.CardMetadataEvent
	for rows.Next() {
		var ev homestate.CardMetadataEvent
		var payload string
		if err := rows.Scan(&ev.CardID, &ev.Kind, &payload); err != nil {
			t.Fatal(err)
		}
		// Decode the payload column back into the event's fields (the same
		// shape appendCardMetadataRow wrote).
		var decoded struct {
			Actor        string `json:"actor"`
			Session      string `json:"session"`
			Field        string `json:"field"`
			Old          string `json:"old"`
			New          string `json:"new"`
			Reason       string `json:"reason"`
			ClassifiedAt string `json:"classified_at"`
		}
		if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
			t.Fatalf("decode audit payload: %v", err)
		}
		ev.Actor, ev.Session, ev.Field = decoded.Actor, decoded.Session, decoded.Field
		ev.Old, ev.New, ev.Reason, ev.ClassifiedAt = decoded.Old, decoded.New, decoded.Reason, decoded.ClassifiedAt
		out = append(out, ev)
	}
	return out
}

// mdAuditCount counts every metadata audit row in the store.
func mdAuditCount(t *testing.T, root string) int {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	var n int
	if err := db.DB.QueryRow(`SELECT count(*) FROM card_metadata_events`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// mdSeedAudit appends one audit row directly (a fixture placement).
func mdSeedAudit(t *testing.T, root, cardID, oldValue, newValue string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.RecordCardMetadataEvent(context.Background(), homestate.CardMetadataEvent{
		CardID: cardID, Kind: "set-spec", Actor: "operator", Field: "spec_id",
		Old: oldValue, New: newValue, Reason: "seed",
	}, mdSpecNow); err != nil {
		t.Fatalf("seed audit row: %v", err)
	}
}

// mdSetRecordSpec updates the existing record row's spec and version in
// place — a fixture edit for the seeded crash states (the row already
// exists; re-placing it would collide on the (run_id, card_id) key).
func mdSetRecordSpec(t *testing.T, root, cardID, spec string, version int64) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`UPDATE cards SET spec_id=?, version=? WHERE card_id=?`, spec, version, cardID); err != nil {
		t.Fatalf("seed record spec %s: %v", cardID, err)
	}
}

// mdPickedFixture is a picked queue card with a factory record and no spec.
func mdPickedFixture(t *testing.T) (string, *factory.BacklogStore) {
	t.Helper()
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked})
	return root, store
}

func TestSetSpec_FlagOff_Refused(t *testing.T) {
	root, store := mdPickedFixture(t)
	// The distributed default: no config file names the flag at all.
	queueBefore := mdQueueDump(t, store)
	cardsBefore := mdCardsDump(t, root)
	_, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-FOO-001", "--reason", "link it", "--run", fcRun)
	if err == nil {
		t.Fatal("set-spec ran with the flag off; it must refuse naming the flag key")
	}
	if !strings.Contains(err.Error(), "workflow.card_metadata_verbs.enabled") {
		t.Fatalf("refusal does not name the flag key: %v", err)
	}
	if got := mdQueueDump(t, store); got != queueBefore {
		t.Fatalf("queue changed on a flag-off refusal:\n got %s\nwant %s", got, queueBefore)
	}
	if got := mdCardsDump(t, root); got != cardsBefore {
		t.Fatalf("cards changed on a flag-off refusal:\n got %s\nwant %s", got, cardsBefore)
	}
	if n := mdAuditCount(t, root); n != 0 {
		t.Fatalf("audit rows = %d, want 0 on a flag-off refusal", n)
	}
}

func TestSetSpec_ValidationGate(t *testing.T) {
	root, store := mdPickedFixture(t)
	mdEnableFlag(t, root)

	refuse := func(label string, args ...string) {
		t.Helper()
		queueBefore := mdQueueDump(t, store)
		cardsBefore := mdCardsDump(t, root)
		_, _, err := runFactory(t, args...)
		if err == nil {
			t.Fatalf("%s: set-spec was admitted, want refusal", label)
		}
		if got := mdQueueDump(t, store); got != queueBefore {
			t.Fatalf("%s: queue changed on refusal:\n got %s\nwant %s", label, got, queueBefore)
		}
		if got := mdCardsDump(t, root); got != cardsBefore {
			t.Fatalf("%s: cards changed on refusal:\n got %s\nwant %s", label, got, cardsBefore)
		}
		if n := mdAuditCount(t, root); n != 0 {
			t.Fatalf("%s: audit rows = %d, want 0 on refusal", label, n)
		}
	}

	// Admission states: queued / dropped / hold / missing (the not-picked
	// family; POSITIVE enumeration lives in the gate).
	for _, state := range []factory.BacklogState{factory.BacklogStateQueued, factory.BacklogStateDropped, factory.BacklogStateHold} {
		root2, store2 := fcFixture(t)
		mdEnableFlag(t, root2)
		fcQueue(t, store2, state)
		fcPlace(t, root2, homestate.Card{CardID: "t1", State: homestate.CardPicked})
		before := mdQueueDump(t, store2)
		_, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-FOO-001", "--reason", "r", "--run", fcRun)
		if err == nil {
			t.Fatalf("admission state %s was admitted", state)
		}
		if got := mdQueueDump(t, store2); got != before {
			t.Fatalf("admission state %s: queue changed on refusal", state)
		}
	}

	// Missing queue item.
	refuse("missing item", "set-spec", "tX", "--spec", "SPEC-FOO-001", "--reason", "r", "--run", fcRun)

	// Gate (b): queue-picked, never assigned — no factory record, no row
	// created, no run identity invented.
	{
		root3, store3 := fcFixture(t)
		mdEnableFlag(t, root3)
		fcQueue(t, store3, factory.BacklogStatePicked)
		_, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-FOO-001", "--reason", "r", "--run", fcRun)
		if err == nil {
			t.Fatal("a queue-picked never-assigned card was admitted")
		}
		if fcHasCard(t, root3, "t1") {
			t.Fatal("a factory record was created for a never-assigned card (REQ-CMV-003 gate b)")
		}
		if n := mdAuditCount(t, root3); n != 0 {
			t.Fatalf("audit rows = %d after the gate-(b) refusal", n)
		}
		_ = store3
	}

	// Grammar gate (c): each refusal quotes the rejected value.
	for _, bad := range []string{"spec-auth-001", "SPEC-2FA-001", "SPEC-AUTH-1", "AUTH-001", ""} {
		_, _, err := runFactory(t, "set-spec", "t1", "--spec", bad, "--reason", "r", "--run", fcRun)
		if err == nil {
			t.Fatalf("grammar case %q was admitted", bad)
		}
		if !strings.Contains(err.Error(), `"`+bad+`"`) {
			t.Fatalf("grammar refusal does not quote %q: %v", bad, err)
		}
	}

	// No-change gate (d): converged-AND-audited duplicate refuses; the
	// audit surface keeps exactly its one seeded row (this case's refusal
	// assertion is "unchanged at 1", not 0 — the seed row is deliberate).
	spec := "SPEC-FOO-001"
	mdQueueSpec(t, store, "t1", &spec)
	mdSetRecordSpec(t, root, "t1", "SPEC-FOO-001", 2)
	mdSeedAudit(t, root, "t1", "", "SPEC-FOO-001")
	queueBefore, cardsBefore := mdQueueDump(t, store), mdCardsDump(t, root)
	if _, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-FOO-001", "--reason", "again", "--run", fcRun); err == nil {
		t.Fatal("the converged-and-audited duplicate was admitted")
	}
	if got := mdQueueDump(t, store); got != queueBefore {
		t.Fatalf("duplicate refusal moved the queue:\n got %s\nwant %s", got, queueBefore)
	}
	if got := mdCardsDump(t, root); got != cardsBefore {
		t.Fatalf("duplicate refusal moved the cards:\n got %s\nwant %s", got, cardsBefore)
	}
	if n := mdAuditCount(t, root); n != 1 {
		t.Fatalf("audit rows = %d after the duplicate refusal, want exactly the seeded 1", n)
	}
}

func TestSetSpec_TwoStoresWritten(t *testing.T) {
	root, store := mdPickedFixture(t)
	mdEnableFlag(t, root)
	out, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-CMV-999", "--reason", "link the SPEC after pick", "--run", fcRun)
	if err != nil {
		t.Fatalf("set-spec: %v", err)
	}
	if !strings.Contains(out, "SPEC-CMV-999") {
		t.Fatalf("output does not name the spec: %q", out)
	}
	// The queue item carries the value (the store cardToSpecID reads).
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if rec.Items[0].SpecID == nil || *rec.Items[0].SpecID != "SPEC-CMV-999" {
		t.Fatalf("queue spec = %v, want SPEC-CMV-999", rec.Items[0].SpecID)
	}
	// The factory record carries the value with a bumped version.
	card := fcCard(t, root, "t1")
	if card.SpecID != "SPEC-CMV-999" || card.Version != 2 {
		t.Fatalf("record = (spec %q, v%d), want (SPEC-CMV-999, v2)", card.SpecID, card.Version)
	}
	// Exactly one audit row, carrying the mutation verbatim.
	rows := mdAuditRows(t, root, "t1")
	if len(rows) != 1 {
		t.Fatalf("audit rows = %d, want exactly 1", len(rows))
	}
	if rows[0].Kind != "set-spec" || rows[0].Field != "spec_id" || rows[0].Old != "" || rows[0].New != "SPEC-CMV-999" || rows[0].Reason != "link the SPEC after pick" {
		t.Fatalf("audit row = %+v, want the recorded mutation verbatim", rows[0])
	}
	if rows[0].Actor == "" {
		t.Fatal("audit row carries no actor")
	}
}

func TestSetSpec_FailureMatrix(t *testing.T) {
	// Direction 1 — factory-side failure: nothing commits, the queue write
	// is never attempted.
	t.Run("direction1_factory_side", func(t *testing.T) {
		root, store := mdPickedFixture(t)
		mdEnableFlag(t, root)
		queueBefore, cardsBefore := mdQueueDump(t, store), mdCardsDump(t, root)
		restore := metadataWriteFault
		metadataWriteFault = func(stage string) error {
			if stage == "record-set" {
				return context.DeadlineExceeded
			}
			return nil
		}
		t.Cleanup(func() { metadataWriteFault = restore })
		if _, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-D1-001", "--reason", "r", "--run", fcRun); err == nil {
			t.Fatal("direction 1 was admitted")
		}
		if got := mdQueueDump(t, store); got != queueBefore {
			t.Fatalf("queue moved on a direction-1 refusal: %s", got)
		}
		if got := mdCardsDump(t, root); got != cardsBefore {
			t.Fatalf("cards moved on a direction-1 refusal: %s", got)
		}
		if n := mdAuditCount(t, root); n != 0 {
			t.Fatalf("audit rows = %d on direction 1", n)
		}
	})

	// Direction 2 — queue-side failure after the record commit: the
	// compensating restore lands and the verb exits non-zero with the
	// divergence sentinel; no audit row.
	t.Run("direction2_queue_side_restored", func(t *testing.T) {
		root, store := mdPickedFixture(t)
		mdEnableFlag(t, root)
		queueBefore := mdQueueDump(t, store)
		restore := metadataWriteFault
		metadataWriteFault = func(stage string) error {
			switch stage {
			case "queue-write":
				return context.DeadlineExceeded
			default:
				return nil
			}
		}
		t.Cleanup(func() { metadataWriteFault = restore })
		_, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-D2-001", "--reason", "r", "--run", fcRun)
		if err == nil {
			t.Fatal("direction 2 exited zero")
		}
		if !strings.Contains(err.Error(), "card metadata divergence") {
			t.Fatalf("direction 2 error lacks the divergence sentinel: %v", err)
		}
		if got := mdQueueDump(t, store); got != queueBefore {
			t.Fatalf("queue moved on direction 2: %s", got)
		}
		card := fcCard(t, root, "t1")
		if card.SpecID != "" || card.Version != 3 {
			t.Fatalf("record after the compensating restore = (spec %q, v%d), want (\"\", v3)", card.SpecID, card.Version)
		}
		if n := mdAuditCount(t, root); n != 0 {
			t.Fatalf("audit rows = %d on direction 2", n)
		}
	})

	// Direction 3 — the compensating write itself fails: the divergence
	// stands, surfaced with card, prior value, and attempted value.
	t.Run("direction3_restore_fails", func(t *testing.T) {
		root, store := mdPickedFixture(t)
		mdEnableFlag(t, root)
		queueBefore := mdQueueDump(t, store)
		restore := metadataWriteFault
		metadataWriteFault = func(stage string) error {
			switch stage {
			case "queue-write", "record-restore":
				return context.DeadlineExceeded
			default:
				return nil
			}
		}
		t.Cleanup(func() { metadataWriteFault = restore })
		_, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-D3-001", "--reason", "r", "--run", fcRun)
		if err == nil {
			t.Fatal("direction 3 exited zero")
		}
		for _, want := range []string{"card metadata divergence", "t1", "SPEC-D3-001"} {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("direction 3 error lacks %q: %v", want, err)
			}
		}
		if got := mdQueueDump(t, store); got != queueBefore {
			t.Fatalf("queue moved on direction 3: %s", got)
		}
		card := fcCard(t, root, "t1")
		if card.SpecID != "SPEC-D3-001" {
			t.Fatalf("record after the failed restore = %q, want the attempted SPEC-D3-001 (divergence stands)", card.SpecID)
		}
		if n := mdAuditCount(t, root); n != 0 {
			t.Fatalf("audit rows = %d on direction 3", n)
		}
	})

	// Direction 4 — audit-append failure after both commits: one retry, then
	// the audit-loss sentinel with the store writes left in place.
	t.Run("direction4_audit_append", func(t *testing.T) {
		root, store := mdPickedFixture(t)
		mdEnableFlag(t, root)
		attempts := 0
		restore := metadataWriteFault
		metadataWriteFault = func(stage string) error {
			if stage == "audit-append" {
				attempts++
				if attempts <= 2 { // the first attempt and its single retry
					return context.DeadlineExceeded
				}
			}
			return nil
		}
		t.Cleanup(func() { metadataWriteFault = restore })
		_, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-D4-001", "--reason", "r", "--run", fcRun)
		if err == nil {
			t.Fatal("direction 4 exited zero")
		}
		if !strings.Contains(err.Error(), "card metadata audit row lost") {
			t.Fatalf("direction 4 error lacks the audit-loss sentinel: %v", err)
		}
		if attempts != 2 {
			t.Fatalf("append attempts = %d, want the first attempt plus exactly one retry", attempts)
		}
		// The completed store writes stay in place — a surfaced missing row,
		// never a silent rollback.
		rec, err2 := store.LoadPure()
		if err2 != nil {
			t.Fatal(err2)
		}
		if rec.Items[0].SpecID == nil || *rec.Items[0].SpecID != "SPEC-D4-001" {
			t.Fatalf("queue spec after direction 4 = %v, want SPEC-D4-001 left in place", rec.Items[0].SpecID)
		}
		if card := fcCard(t, root, "t1"); card.SpecID != "SPEC-D4-001" {
			t.Fatalf("record spec after direction 4 = %q, want left in place", card.SpecID)
		}
		if n := mdAuditCount(t, root); n != 0 {
			t.Fatalf("audit rows = %d after the audit-loss sentinel, want 0", n)
		}
	})

	// Direction 4 recovery — the retry lands: one failed attempt, the retry
	// succeeds, exactly one audit row.
	t.Run("direction4_retry_lands", func(t *testing.T) {
		root, store := mdPickedFixture(t)
		mdEnableFlag(t, root)
		attempts := 0
		restore := metadataWriteFault
		metadataWriteFault = func(stage string) error {
			if stage == "audit-append" {
				attempts++
				if attempts == 1 {
					return context.DeadlineExceeded
				}
			}
			return nil
		}
		t.Cleanup(func() { metadataWriteFault = restore })
		if _, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-D4-002", "--reason", "r", "--run", fcRun); err != nil {
			t.Fatalf("retrying append: %v", err)
		}
		rows := mdAuditRows(t, root, "t1")
		if len(rows) != 1 {
			t.Fatalf("audit rows = %d, want exactly 1 after the landed retry", len(rows))
		}
		_ = store
	})

	// Direction 5 — seeded record-new/queue-old: the re-run is admitted (the
	// no-change gate reads the QUEUE item) and converges both stores with
	// exactly one audit row whose old value is the queue's pre-heal value.
	t.Run("direction5_rerun_heals", func(t *testing.T) {
		root, store := mdPickedFixture(t)
		mdEnableFlag(t, root)
		spec := "SPEC-OLD-001"
		mdQueueSpec(t, store, "t1", &spec)
		mdSetRecordSpec(t, root, "t1", "SPEC-NEW-001", 2)
		if _, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-NEW-001", "--reason", "heal", "--run", fcRun); err != nil {
			t.Fatalf("direction-5 re-run refused: %v", err)
		}
		rec, err := store.LoadPure()
		if err != nil {
			t.Fatal(err)
		}
		if rec.Items[0].SpecID == nil || *rec.Items[0].SpecID != "SPEC-NEW-001" {
			t.Fatalf("queue spec after the heal = %v, want converged on SPEC-NEW-001", rec.Items[0].SpecID)
		}
		rows := mdAuditRows(t, root, "t1")
		if len(rows) != 1 {
			t.Fatalf("audit rows = %d, want exactly 1 after the heal", len(rows))
		}
		if rows[0].Old != "SPEC-OLD-001" || rows[0].New != "SPEC-NEW-001" {
			t.Fatalf("heal row = (%q → %q), want (SPEC-OLD-001 → SPEC-NEW-001)", rows[0].Old, rows[0].New)
		}
	})

	// Direction 6 — seeded both-converged/unaudited: the re-run is the
	// converged-state audit completion; it emits exactly the owed row with
	// the re-run's actor and reason, mutates neither store, exits zero. The
	// duplicate re-run after that completion is refused with no second row.
	t.Run("direction6_catch_up_then_duplicate", func(t *testing.T) {
		root, store := mdPickedFixture(t)
		mdEnableFlag(t, root)
		spec := "SPEC-D6-001"
		mdQueueSpec(t, store, "t1", &spec)
		// Direction 6's true crash state: BOTH stores committed the
		// commanded value, then the process died before the append.
		mdSetRecordSpec(t, root, "t1", "SPEC-D6-001", 2)
		cardsBefore := mdCardsDump(t, root)
		if _, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-D6-001", "--reason", "catch-up attestation", "--run", fcRun); err != nil {
			t.Fatalf("direction-6 re-run refused: %v", err)
		}
		rows := mdAuditRows(t, root, "t1")
		if len(rows) != 1 {
			t.Fatalf("audit rows = %d, want exactly the owed row", len(rows))
		}
		if rows[0].Old != "" || rows[0].New != "SPEC-D6-001" || rows[0].Reason != "catch-up attestation" {
			t.Fatalf("catch-up row = %+v, want old=\"\" new=SPEC-D6-001 with the re-run's reason", rows[0])
		}
		// Neither store re-mutated: the record row is unchanged (version
		// included) and the queue item keeps its value.
		if got := mdCardsDump(t, root); got != cardsBefore {
			t.Fatalf("record re-mutated by the catch-up:\n got %s\nwant %s", got, cardsBefore)
		}
		// The duplicate re-run (converged AND audited) refuses with no second row.
		queueBefore := mdQueueDump(t, store)
		_, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-D6-001", "--reason", "again", "--run", fcRun)
		if err == nil {
			t.Fatal("the converged-and-audited duplicate was admitted")
		}
		if n := len(mdAuditRows(t, root, "t1")); n != 1 {
			t.Fatalf("audit rows = %d after the duplicate refusal, want 1", n)
		}
		if got := mdQueueDump(t, store); got != queueBefore {
			t.Fatalf("queue moved on the duplicate refusal: %s", got)
		}
	})

	// The ABA seed (DB-1, plan-audit iter6): audited X→A then audited A→B,
	// then a B→A run that crashes after both commits and before the append.
	// The re-run commanding A is ADMITTED — the duplicate test reads the
	// LATEST row (whose new value is B), never a prior row — and lands
	// exactly one row with old=B and new=A, stores row-unchanged.
	t.Run("aba_seed_latest_row_decides", func(t *testing.T) {
		root, store := mdPickedFixture(t)
		mdEnableFlag(t, root)
		spec := "SPEC-ABA-001"
		mdQueueSpec(t, store, "t1", &spec)
		mdSetRecordSpec(t, root, "t1", "SPEC-ABA-001", 2)
		mdSeedAudit(t, root, "t1", "SPEC-X-001", "SPEC-ABA-001") // X→A
		mdSeedAudit(t, root, "t1", "SPEC-ABA-001", "SPEC-B-001") // A→B
		cardsBefore := mdCardsDump(t, root)
		if _, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-ABA-001", "--reason", "aba catch-up", "--run", fcRun); err != nil {
			t.Fatalf("ABA re-run refused: %v", err)
		}
		rows := mdAuditRows(t, root, "t1")
		if len(rows) != 3 {
			t.Fatalf("audit rows = %d, want 3 (two seeds plus exactly one ABA row)", len(rows))
		}
		last := rows[len(rows)-1]
		if last.Old != "SPEC-B-001" || last.New != "SPEC-ABA-001" {
			t.Fatalf("ABA row = (%q → %q), want (SPEC-B-001 → SPEC-ABA-001): the LATEST row decides, not the prior A row", last.Old, last.New)
		}
		if got := mdCardsDump(t, root); got != cardsBefore {
			t.Fatalf("record re-mutated by the ABA catch-up")
		}
	})
}

// TestSetSpec_CatchUpAdmission is the sync-audit round-1 B1 regression
// table: the direction-6 catch-up admission reads BOTH stores. The
// queue-side comparison alone admitted a queue-converged/record-stale state
// (the assign-without-spec blank this SPEC §A exists to close): it exited
// zero on a record it never wrote and attested a transition that never
// landed. The heal rows fall through to the write path — the record lands
// the commanded value with a bumped version, the queue keeps its value, and
// the audit row's old value is the record's true prior (the store that
// actually moves). The first row is the direction-6 control: both stores
// converged still takes the no-store-mutation catch-up.
func TestSetSpec_CatchUpAdmission(t *testing.T) {
	spec := "SPEC-FOO-001"
	for _, tc := range []struct {
		name       string
		recordSpec string // the seeded record spec_id; "" is the assign-without-spec blank
		recordVer  int64
		wantMoved  bool   // the record must land the commanded value with a bumped version
		wantOld    string // the audit row's old value
	}{
		{"direction6_both_converged_still_catches_up", spec, 2, false, ""},
		{"queue_converged_record_blank_heals", "", 1, true, ""},
		{"queue_converged_record_stale_value_heals", "SPEC-OLD-009", 1, true, "SPEC-OLD-009"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, store := mdPickedFixture(t)
			mdEnableFlag(t, root)
			mdQueueSpec(t, store, "t1", &spec)
			if tc.recordSpec != "" {
				mdSetRecordSpec(t, root, "t1", tc.recordSpec, tc.recordVer)
			}
			cardsBefore := mdCardsDump(t, root)
			out, _, err := runFactory(t, "set-spec", "t1", "--spec", spec, "--reason", "converge", "--run", fcRun)
			if err != nil {
				t.Fatalf("set-spec: %v", err)
			}
			if !strings.Contains(out, spec) {
				t.Fatalf("output does not name the spec: %q", out)
			}
			// The record post-state: the heal rows write the commanded value
			// and bump the version; the catch-up row re-mutates nothing.
			if tc.wantMoved {
				card := fcCard(t, root, "t1")
				if card.SpecID != spec {
					t.Fatalf("record spec = %q, want %q written to the record", card.SpecID, spec)
				}
				if card.Version != tc.recordVer+1 {
					t.Fatalf("record version = %d, want %d (the write bumps it)", card.Version, tc.recordVer+1)
				}
			} else if got := mdCardsDump(t, root); got != cardsBefore {
				t.Fatalf("catch-up re-mutated the record:\n got %s\nwant %s", got, cardsBefore)
			}
			// The queue post-state: converged on the commanded value, still picked.
			rec, err := store.LoadPure()
			if err != nil {
				t.Fatal(err)
			}
			if rec.Items[0].SpecID == nil || *rec.Items[0].SpecID != spec || rec.Items[0].State != factory.BacklogStatePicked {
				t.Fatalf("queue item = (%s, %v), want (picked, %s)", rec.Items[0].State, rec.Items[0].SpecID, spec)
			}
			// Exactly one audit row carrying the honest transition.
			rows := mdAuditRows(t, root, "t1")
			if len(rows) != 1 {
				t.Fatalf("audit rows = %d, want exactly 1", len(rows))
			}
			if rows[0].Old != tc.wantOld || rows[0].New != spec || rows[0].Reason != "converge" {
				t.Fatalf("audit row = (%q → %q, reason %q), want (%q → %q, reason converge)", rows[0].Old, rows[0].New, rows[0].Reason, tc.wantOld, spec)
			}
		})
	}
}

func TestSetSpec_ReasonValidation(t *testing.T) {
	root, store := mdPickedFixture(t)
	mdEnableFlag(t, root)
	for _, bad := range []string{"", "   ", strings.Repeat("r", 501)} {
		_, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-R-001", "--reason", bad, "--run", fcRun)
		if err == nil {
			t.Fatalf("reason %q was admitted", bad)
		}
	}
	if n := mdAuditCount(t, root); n != 0 {
		t.Fatalf("audit rows = %d after the reason refusals, want 0", n)
	}
	// 500 runes is accepted.
	good := strings.Repeat("r", 500)
	if _, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-R-001", "--reason", good, "--run", fcRun); err != nil {
		t.Fatalf("a 500-rune reason was refused: %v", err)
	}
	rows := mdAuditRows(t, root, "t1")
	if len(rows) != 1 || rows[0].Reason != good {
		t.Fatalf("audit rows = %d, want 1 carrying the 500-rune reason verbatim", len(rows))
	}
	_ = store
}

func TestSetSpec_RefusalAtomicity(t *testing.T) {
	// The gate-refusal set only — the failure directions deliberately leave
	// residue and are asserted per-direction by TestSetSpec_FailureMatrix.
	for _, tc := range []struct {
		label string
		args  []string
	}{
		{"flag-off", nil}, // handled specially below
		{"admission-queued", []string{"set-spec", "t1", "--spec", "SPEC-RA-001", "--reason", "r", "--run", fcRun}},
		{"grammar", []string{"set-spec", "t1", "--spec", "nope", "--reason", "r", "--run", fcRun}},
		{"reason-empty", []string{"set-spec", "t1", "--spec", "SPEC-RA-001", "--reason", "", "--run", fcRun}},
	} {
		if tc.label == "flag-off" {
			continue
		}
		t.Run(tc.label, func(t *testing.T) {
			root, store := mdPickedFixture(t)
			mdEnableFlag(t, root)
			if tc.label == "admission-queued" {
				// Reset the item to queued for this row.
				if err := store.Mutate(func(r *factory.BacklogRecord) error {
					for i := range r.Items {
						if r.Items[i].ID == "t1" {
							r.Items[i].State = factory.BacklogStateQueued
						}
					}
					return nil
				}); err != nil {
					t.Fatal(err)
				}
			}
			queueBefore, cardsBefore := mdQueueDump(t, store), mdCardsDump(t, root)
			if _, _, err := runFactory(t, tc.args...); err == nil {
				t.Fatalf("%s was admitted", tc.label)
			}
			if got := mdQueueDump(t, store); got != queueBefore {
				t.Fatalf("%s: queue changed:\n got %s\nwant %s", tc.label, got, queueBefore)
			}
			if got := mdCardsDump(t, root); got != cardsBefore {
				t.Fatalf("%s: cards changed:\n got %s\nwant %s", tc.label, got, cardsBefore)
			}
			if n := mdAuditCount(t, root); n != 0 {
				t.Fatalf("%s: audit rows = %d, want 0", tc.label, n)
			}
		})
	}
}

// TestSetSpec_MissingRecordRefusalSparesTheStore pins the D-B alignment
// (sync-audit round 1): the gate-(b) refusal precedes the store open. A
// project whose factory database was never created must still have none
// after the refusal (OpenFactory migrates on open — refusing after it would
// materialize an empty store as a side effect of a read that found nothing,
// the accessor convention the lease gate set: a refused open leaves no
// database behind); with the store present but the row absent, the same
// refusal stands.
func TestSetSpec_MissingRecordRefusalSparesTheStore(t *testing.T) {
	root, store := fcFixture(t)
	mdEnableFlag(t, root)
	fcQueue(t, store, factory.BacklogStatePicked)
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, serr := os.Stat(path); serr == nil {
		t.Fatal("precondition: the factory database already exists")
	}
	if _, _, rerr := runFactory(t, "set-spec", "t1", "--spec", "SPEC-DB-001", "--reason", "r", "--run", fcRun); rerr == nil {
		t.Fatal("a never-assigned card was admitted")
	} else if !strings.Contains(rerr.Error(), "has no factory record") {
		t.Fatalf("refusal lacks the missing-record shape: %v", rerr)
	}
	if _, serr := os.Stat(path); serr == nil {
		t.Fatal("the refusal materialized a factory database (sync-audit-1 D-B)")
	}
	// With the store present and the row absent, the same refusal stands.
	fcPlace(t, root, homestate.Card{CardID: "t2", State: homestate.CardPicked})
	if _, _, rerr := runFactory(t, "set-spec", "t1", "--spec", "SPEC-DB-001", "--reason", "r", "--run", fcRun); rerr == nil {
		t.Fatal("a record-less card was admitted on an existing store")
	} else if !strings.Contains(rerr.Error(), "has no factory record") {
		t.Fatalf("refusal lacks the missing-record shape on an existing store: %v", rerr)
	}
	if n := mdAuditCount(t, root); n != 0 {
		t.Fatalf("audit rows = %d after the refusal pair, want 0", n)
	}
}

// TestSetSpec_OpaqueStatErrorDoesNotClaimMissingRecord pins the D-G scoping
// (sync-audit round 2): the pre-open existence probe refuses with the
// missing-record message only on a truly absent store. A stat error with
// another cause — here the store path made a self-referential symlink, so
// stat fails with ELOOP — is not evidence about the record: the verb falls
// through to the open and the accessor's own error surfaces the real cause.
// The loop is unhealable by the accessor (openFactoryPathBusy normalizes
// the store DIRECTORY's mode, which cannot touch a looping file path), and
// ELOOP is uid-independent, so no root skip is needed.
func TestSetSpec_OpaqueStatErrorDoesNotClaimMissingRecord(t *testing.T) {
	root, store := fcFixture(t)
	mdEnableFlag(t, root)
	fcQueue(t, store, factory.BacklogStatePicked)
	// The store was materialized (fcPlace); its path is then swapped for a
	// symlink that resolves to itself, so the stat probe fails for a reason
	// that is not absence while every earlier gate reads a live store.
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked})
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(path, path); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(path) })
	_, _, rerr := runFactory(t, "set-spec", "t1", "--spec", "SPEC-DG-001", "--reason", "r", "--run", fcRun)
	if rerr == nil {
		t.Fatal("an unopenable store was admitted")
	}
	if strings.Contains(rerr.Error(), "has no factory record") {
		t.Fatalf("a non-not-exist stat error was misnamed as a missing record: %v", rerr)
	}
}

// TestReclassify_RefusalSparesTheStore extends the no-materialization probe
// family to reclassify (card-review round 2, CR2-2): on a project with no
// factory database, reclassify's refusable inputs — an unknown card, an
// invalid priority or mode — must refuse without opening the store (an open
// that creates it). The queue-state and classification validation runs
// inside the queue lock before the open; pre/post os.Stat pins the absence,
// with --run omitted and explicit.
func TestReclassify_RefusalSparesTheStore(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"unknown_card", []string{"reclassify", "tX", "--priority", "high", "--reason", "r"}},
		{"invalid_priority", []string{"reclassify", "t1", "--priority", "urgent", "--reason", "r"}},
		{"invalid_mode", []string{"reclassify", "t1", "--mode", "fast", "--reason", "r"}},
	} {
		for _, extra := range [][]string{nil, {"--run", fcRun}} {
			name := tc.name
			if extra != nil {
				name += "/explicit_run"
			}
			t.Run(name, func(t *testing.T) {
				root, store := fcFixture(t)
				mdEnableFlag(t, root)
				fcQueue(t, store, factory.BacklogStateQueued)
				path, err := homestate.FactoryDBPath(root)
				if err != nil {
					t.Fatal(err)
				}
				if _, serr := os.Stat(path); serr == nil {
					t.Fatal("precondition: the factory database already exists")
				}
				args := append(append([]string{}, tc.args...), extra...)
				if _, _, rerr := runFactory(t, args...); rerr == nil {
					t.Fatalf("%v was admitted", args)
				}
				if _, serr := os.Stat(path); serr == nil {
					t.Fatalf("the refused reclassify materialized a factory database (CR2-2): %v", args)
				}
			})
		}
	}
}

// TestSetSpec_OmittedRunRefusalSparesTheStore pins the CR1 ordering
// (card-review): the store-existence probe precedes the run resolution.
// With --run omitted the shared resolver opens the store to list active
// runs — an open that CREATES it — so a verb that resolves before probing
// materializes a store on a refused path. The probe runs first; the
// refusal below leaves the store as it found it: absent.
func TestSetSpec_OmittedRunRefusalSparesTheStore(t *testing.T) {
	root, store := fcFixture(t)
	mdEnableFlag(t, root)
	fcQueue(t, store, factory.BacklogStatePicked)
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, serr := os.Stat(path); serr == nil {
		t.Fatal("precondition: the factory database already exists")
	}
	if _, _, rerr := runFactory(t, "set-spec", "t1", "--spec", "SPEC-CR1-001", "--reason", "r"); rerr == nil {
		t.Fatal("set-spec without --run was admitted on an absent store")
	}
	if _, serr := os.Stat(path); serr == nil {
		t.Fatal("the refused run materialized a factory database (card-review CR1)")
	}
	// The probe, not the run resolution, answered the refusal: a resolver
	// first would have named the missing active run instead.
	if _, _, rerr := runFactory(t, "set-spec", "t1", "--spec", "SPEC-CR1-001", "--reason", "r"); rerr == nil || !strings.Contains(rerr.Error(), "has no factory record") {
		t.Fatalf("refusal did not come from the existence probe: %v", rerr)
	}
}

// TestSetSpec_CatchUpAppendRetry pins the D-A alignment (sync-audit round
// 1): the direction-6 catch-up append carries the write path's append
// discipline — one retry, then the audit-loss sentinel, exit non-zero, with
// both stores row-unchanged through the failure (the catch-up mutates no
// store either way) — and a retry that lands leaves exactly the owed row.
func TestSetSpec_CatchUpAppendRetry(t *testing.T) {
	spec := "SPEC-D6-001"
	seed := func(t *testing.T) (string, *factory.BacklogStore, string) {
		t.Helper()
		root, store := mdPickedFixture(t)
		mdEnableFlag(t, root)
		mdQueueSpec(t, store, "t1", &spec)
		mdSetRecordSpec(t, root, "t1", spec, 2)
		return root, store, mdCardsDump(t, root)
	}
	// Both attempts fail: the sentinel fires and nothing was mutated.
	root, store, cardsBefore := seed(t)
	// Captured BEFORE the run, so the no-mutation comparison below is real
	// (sync-audit-2 D-E: a post-run capture compares post-run against
	// post-run and is vacuous).
	queueBefore := mdQueueDump(t, store)
	attempts := 0
	restore := metadataWriteFault
	metadataWriteFault = func(stage string) error {
		if stage == "audit-append" {
			attempts++
			if attempts <= 2 {
				return context.DeadlineExceeded
			}
		}
		return nil
	}
	t.Cleanup(func() { metadataWriteFault = restore })
	_, _, err := runFactory(t, "set-spec", "t1", "--spec", spec, "--reason", "r", "--run", fcRun)
	if err == nil {
		t.Fatal("the catch-up audit-loss sentinel did not fire")
	}
	if !strings.Contains(err.Error(), "card metadata audit row lost") {
		t.Fatalf("error lacks the audit-loss sentinel: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("append attempts = %d, want the first attempt plus exactly one retry", attempts)
	}
	if got := mdCardsDump(t, root); got != cardsBefore {
		t.Fatalf("catch-up failure re-mutated the record:\n got %s\nwant %s", got, cardsBefore)
	}
	if got := mdQueueDump(t, store); got != queueBefore {
		t.Fatalf("catch-up failure moved the queue:\n got %s\nwant %s", got, queueBefore)
	}
	if n := mdAuditCount(t, root); n != 0 {
		t.Fatalf("audit rows = %d after the catch-up sentinel, want 0", n)
	}

	// The retry lands: one failed attempt, the retry succeeds, one row.
	root2, _, _ := seed(t)
	attempts2 := 0
	restore2 := metadataWriteFault
	metadataWriteFault = func(stage string) error {
		if stage == "audit-append" {
			attempts2++
			if attempts2 == 1 {
				return context.DeadlineExceeded
			}
		}
		return nil
	}
	t.Cleanup(func() { metadataWriteFault = restore2 })
	if _, _, err := runFactory(t, "set-spec", "t1", "--spec", spec, "--reason", "r", "--run", fcRun); err != nil {
		t.Fatalf("retrying catch-up append: %v", err)
	}
	if rows := mdAuditRows(t, root2, "t1"); len(rows) != 1 {
		t.Fatalf("audit rows = %d, want exactly 1 after the landed catch-up retry", len(rows))
	}
}

// TestReclassify_CatchUpAppendRetry is the reclassify mirror of
// TestSetSpec_CatchUpAppendRetry (sync-audit round 1, D-A): the converged
// catch-up append retries once, then surfaces the audit-loss sentinel with
// the queue write left exactly as the crashed run committed it.
func TestReclassify_CatchUpAppendRetry(t *testing.T) {
	// mdClassifiedFixture already carries the axes a
	// `--mode serial --priority high` run commands: the re-run is the
	// converged-state catch-up.
	root, store := mdClassifiedFixture(t)
	queueBefore := mdQueueDump(t, store)
	// The no-mutation assertion pins the CLASSIFICATION record itself — the
	// field the verb mutates — stamp and reason included; the dump above
	// renders id/state/spec only and cannot see it (sync-audit-2 D-F).
	classBefore := *mdClass(t, store, "t1")
	attempts := 0
	restore := metadataWriteFault
	metadataWriteFault = func(stage string) error {
		if stage == "audit-append" {
			attempts++
			if attempts <= 2 {
				return context.DeadlineExceeded
			}
		}
		return nil
	}
	t.Cleanup(func() { metadataWriteFault = restore })
	_, _, err := runFactory(t, "reclassify", "t1", "--mode", "serial", "--priority", "high", "--reason", "r")
	if err == nil {
		t.Fatal("the catch-up audit-loss sentinel did not fire")
	}
	if !strings.Contains(err.Error(), "card metadata audit row lost") {
		t.Fatalf("error lacks the audit-loss sentinel: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("append attempts = %d, want the first attempt plus exactly one retry", attempts)
	}
	if got := mdQueueDump(t, store); got != queueBefore {
		t.Fatalf("catch-up failure moved the queue:\n got %s\nwant %s", got, queueBefore)
	}
	if got := mdClass(t, store, "t1"); got == nil || *got != classBefore {
		t.Fatalf("classification mutated by the failed catch-up:\n got %+v\nwant %+v", got, classBefore)
	}
	if n := mdAuditCount(t, root); n != 0 {
		t.Fatalf("audit rows = %d after the catch-up sentinel, want 0", n)
	}

	// The retry lands: one failed attempt, the retry succeeds, one row.
	root2, _ := mdClassifiedFixture(t)
	attempts2 := 0
	restore2 := metadataWriteFault
	metadataWriteFault = func(stage string) error {
		if stage == "audit-append" {
			attempts2++
			if attempts2 == 1 {
				return context.DeadlineExceeded
			}
		}
		return nil
	}
	t.Cleanup(func() { metadataWriteFault = restore2 })
	if _, _, err := runFactory(t, "reclassify", "t1", "--mode", "serial", "--priority", "high", "--reason", "r"); err != nil {
		t.Fatalf("retrying catch-up append: %v", err)
	}
	if rows := mdAuditRows(t, root2, "t1"); len(rows) != 1 {
		t.Fatalf("audit rows = %d, want exactly 1 after the landed catch-up retry", len(rows))
	}
}

// TestReclassify_PriorityResortsQueue pins the CR2-1 guarantee (card-review
// round 2): a reclassify is a classification WRITE, so the same locked write
// re-establishes the queue's classification order (REQ-TCD-005) exactly as
// every add does. Without the re-sort the stored order stays insertion
// order — and selection reads stored order as priority order, so a later
// card promoted to high still loses the next lease to the earlier card.
func TestReclassify_PriorityResortsQueue(t *testing.T) {
	root, store := fcFixture(t)
	mdEnableFlag(t, root)
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStateQueued)
	// The reclassify runs BEFORE the lane env is set: the metadata verbs are
	// leader/operator acts (REQ-CMV-015) and refuse in a lane session, while
	// the `next` lease below needs one.
	// t2 ranks later in insertion order; the reclassify promotes it to high.
	if _, _, err := runFactory(t, "reclassify", "t2", "--priority", "high", "--reason", "escalate"); err != nil {
		t.Fatalf("reclassify: %v", err)
	}
	sdRegisterLane(t, root, "lane-1")
	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)
	// The stored order moved: the promoted card precedes the earlier one.
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if rec.Items[0].ID != "t2" || rec.Items[1].ID != "t1" {
		t.Fatalf("queue order = [%s, %s], want [t2, t1] after promoting t2", rec.Items[0].ID, rec.Items[1].ID)
	}
	// And the next lease takes the promoted card, not the earlier one.
	if _, _, err := runFactory(t, "next", "--run", fcRun); err != nil {
		t.Fatalf("next: %v", err)
	}
	if c := fcCard(t, root, "t2"); c.State != homestate.CardLeased {
		t.Fatalf("t2 = %s, want the promoted card leased first", c.State)
	}
	// An unleased card has no factory row (the lease creates it) — absence is
	// the not-leased proof here.
	if fcHasCard(t, root, "t1") {
		if c := fcCard(t, root, "t1"); c.State == homestate.CardLeased {
			t.Fatal("t1 leased ahead of the promoted t2 — the stored order was not re-sorted")
		}
	}
}

// SPEC-CARD-METADATA-VERBS-001 M3: the `moai factory reclassify` verb —
// admission gate (positive enumeration), closed sets through the real
// validator, axis preservation, write semantics, and the reclassify
// audit-window reconciliation of REQ-CMV-010.

// mdQueueClass pins a queue item's classification (nil clears it).
func mdQueueClass(t *testing.T, store *factory.BacklogStore, cardID string, class *factory.CardClassification) {
	t.Helper()
	if err := store.Mutate(func(r *factory.BacklogRecord) error {
		for i := range r.Items {
			if r.Items[i].ID == cardID {
				r.Items[i].Classification = class
				return nil
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("set queue classification: %v", err)
	}
}

// mdClass reads one queue item's classification pointer back.
func mdClass(t *testing.T, store *factory.BacklogStore, cardID string) *factory.CardClassification {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range rec.Items {
		if it.ID == cardID {
			return it.Classification
		}
	}
	t.Fatalf("no queue item %s", cardID)
	return nil
}

// mdClassifiedFixture is a queued card carrying the AC-CMV-010 classification.
func mdClassifiedFixture(t *testing.T) (string, *factory.BacklogStore) {
	t.Helper()
	root, store := fcFixture(t)
	mdEnableFlag(t, root)
	fcQueue(t, store, factory.BacklogStateQueued)
	mdQueueClass(t, store, "t1", &factory.CardClassification{
		Priority: factory.ClassPriorityHigh, Blocked: true, Mode: factory.ClassModeSerial,
		Decider: factory.DeciderIdentityDefault, ClassifiedAt: "2026-10-20T00:00:00Z", Reason: "add-time judgment",
	})
	return root, store
}

func TestReclassify_ValidationGate(t *testing.T) {
	// Admission refusals, positively enumerated: dropped, hold, and a
	// missing item refuse; queued and picked are the admitted states.
	for _, tc := range []struct {
		state    factory.BacklogState
		admitted bool
	}{
		{factory.BacklogStateQueued, true},
		{factory.BacklogStatePicked, true},
		{factory.BacklogStateDropped, false},
		{factory.BacklogStateHold, false},
	} {
		root, store := fcFixture(t)
		mdEnableFlag(t, root)
		fcQueue(t, store, tc.state)
		before := mdQueueDump(t, store)
		_, _, err := runFactory(t, "reclassify", "t1", "--priority", "high", "--reason", "r")
		if tc.admitted {
			if err != nil {
				t.Fatalf("admission state %s was refused: %v", tc.state, err)
			}
			continue
		}
		if err == nil {
			t.Fatalf("admission state %s was admitted", tc.state)
		}
		if got := mdQueueDump(t, store); got != before {
			t.Fatalf("state %s: queue changed on refusal", tc.state)
		}
	}
	// A missing item refuses.
	root, store := mdClassifiedFixture(t)
	if _, _, err := runFactory(t, "reclassify", "tX", "--priority", "high", "--reason", "r"); err == nil {
		t.Fatal("a missing item was admitted")
	}
	_ = store

	// Closed sets go through the REAL validator — every out-of-set value is
	// refused and the refusal quotes it.
	for _, bad := range []string{"fast", "", "SERIAL", "parallel"} {
		_, _, err := runFactory(t, "reclassify", "t1", "--mode", bad, "--reason", "r")
		if err == nil {
			t.Fatalf("mode %q was admitted", bad)
		}
		if !strings.Contains(err.Error(), `"`+bad+`"`) {
			t.Fatalf("mode refusal does not quote %q: %v", bad, err)
		}
	}
	for _, bad := range []string{"urgent", "", "HIGH", "normal-"} {
		_, _, err := runFactory(t, "reclassify", "t1", "--priority", bad, "--reason", "r")
		if err == nil {
			t.Fatalf("priority %q was admitted", bad)
		}
		if !strings.Contains(err.Error(), `"`+bad+`"`) {
			t.Fatalf("priority refusal does not quote %q: %v", bad, err)
		}
	}

	// No axis: a usage refusal — neither --mode nor --priority supplied.
	if _, _, err := runFactory(t, "reclassify", "t1", "--reason", "r"); err == nil {
		t.Fatal("an axis-less reclassify was admitted")
	}

	// Every refusal above wrote no audit row.
	if n := mdAuditCount(t, root); n != 0 {
		t.Fatalf("audit rows = %d after the refusal table, want 0", n)
	}
}

func TestReclassify_UnspecifiedAxisPreserved(t *testing.T) {
	_, store := mdClassifiedFixture(t)
	if _, _, err := runFactory(t, "reclassify", "t1", "--priority", "low", "--reason", "lower it"); err != nil {
		t.Fatalf("reclassify: %v", err)
	}
	got := mdClass(t, store, "t1")
	if got == nil {
		t.Fatal("classification vanished")
	}
	if got.Priority != factory.ClassPriorityLow {
		t.Fatalf("priority = %q, want low", got.Priority)
	}
	if got.Mode != factory.ClassModeSerial {
		t.Fatalf("mode = %q, want serial preserved (REQ-CMV-008)", got.Mode)
	}
	if !got.Blocked {
		t.Fatal("blocked = false, want true preserved")
	}
}

func TestReclassify_BlockedUntouched(t *testing.T) {
	_, store := mdClassifiedFixture(t)
	if _, _, err := runFactory(t, "reclassify", "t1", "--mode", "parallelizable", "--reason", "parallel work"); err != nil {
		t.Fatalf("reclassify: %v", err)
	}
	got := mdClass(t, store, "t1")
	if !got.Blocked {
		t.Fatal("the reclassify write altered the blocked axis (REQ-CMV-008 forbids it)")
	}
	if got.Mode != factory.ClassModeParallelizable || got.Priority != factory.ClassPriorityHigh {
		t.Fatalf("axes = (mode %q, priority %q), want (parallelizable, high)", got.Mode, got.Priority)
	}
}

func TestReclassify_WriteSemantics(t *testing.T) {
	// First-time classification acquires the nullable field additively: the
	// derived priority, decider human, a fresh stamp, and the caller reason.
	root, store := fcFixture(t)
	mdEnableFlag(t, root)
	fcQueue(t, store, factory.BacklogStateQueued)
	// SPEC-FACTORY-DEFAULT-MODE-001: the add path now records a
	// decider-default classification at creation (REQ-FDM-002), so without
	// this seed a same-axes reclassify takes the converged path and never
	// re-stamps. Seed the absence back: this test names the FIRST-TIME
	// acquire, where reclassify itself stamps decider human.
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == "t1" {
				rec.Items[i].Classification = nil
				return nil
			}
		}
		return fmt.Errorf("no t1 in the fixture queue")
	}); err != nil {
		t.Fatalf("seed the absent classification: %v", err)
	}
	if _, _, err := runFactory(t, "reclassify", "t1", "--mode", "parallelizable", "--reason", "first judgment"); err != nil {
		t.Fatalf("first-time reclassify: %v", err)
	}
	got := mdClass(t, store, "t1")
	if got == nil {
		t.Fatal("the classification field was not acquired")
	}
	if got.Priority != factory.ClassPriorityNormal {
		t.Fatalf("derived priority = %q, want normal", got.Priority)
	}
	if got.Mode != factory.ClassModeParallelizable {
		t.Fatalf("mode = %q, want parallelizable", got.Mode)
	}
	if got.Decider != factory.DeciderIdentityHuman {
		t.Fatalf("decider = %q, want human (REQ-CMV-009)", got.Decider)
	}
	if got.ClassifiedAt == "" || got.ClassifiedAt == "2026-10-20T00:00:00Z" {
		t.Fatalf("classified_at = %q, want a fresh stamp", got.ClassifiedAt)
	}
	if got.Reason != "first judgment" {
		t.Fatalf("reason = %q, want the caller reason verbatim", got.Reason)
	}
	// Exactly one audit row carrying the value transition.
	rows := mdAuditRows(t, root, "t1")
	if len(rows) != 1 {
		t.Fatalf("audit rows = %d, want 1", len(rows))
	}
	if rows[0].Kind != metadataKindReclassify || rows[0].Field != metadataFieldClassific {
		t.Fatalf("row = (%s/%s), want (reclassify/classification)", rows[0].Kind, rows[0].Field)
	}
	// After the flip the absent read derives parallelizable (REQ-FDM-001),
	// so the old axes on a first acquire are the absent-derived axes — the
	// same axes the command commanded, because the command re-asserted the
	// derived mode. The old=new pair on this row is the flip's honest
	// record, not a lost transition.
	if rows[0].Old != `{"blocked":false,"mode":"parallelizable","priority":"normal"}` {
		t.Fatalf("old = %q, want the absent-derived axes (REQ-FDM-001)", rows[0].Old)
	}
	if rows[0].New != `{"blocked":false,"mode":"parallelizable","priority":"normal"}` {
		t.Fatalf("new = %q, want the commanded axes", rows[0].New)
	}

	// A reclassify on a card that HAD a classification re-stamps it.
	_, store2 := mdClassifiedFixture(t)
	if _, _, err := runFactory(t, "reclassify", "t1", "--priority", "low", "--reason", "re-judged"); err != nil {
		t.Fatalf("reclassify on classified card: %v", err)
	}
	got2 := mdClass(t, store2, "t1")
	if got2.Decider != factory.DeciderIdentityHuman || got2.Reason != "re-judged" {
		t.Fatalf("re-stamped = (decider %q, reason %q), want (human, re-judged)", got2.Decider, got2.Reason)
	}
	if got2.ClassifiedAt == "2026-10-20T00:00:00Z" {
		t.Fatalf("classified_at = %q, want a fresh stamp", got2.ClassifiedAt)
	}
}

func TestReclassify_SerialSlotConflictRefused(t *testing.T) {
	// REQ-FDM-004 on the reclassify surface (card t1737 delta F1): the
	// reclassify write commanding serial runs the same serial-conflict gate
	// the todo classify verb runs — an in-flight serial holder refuses the
	// second serial card instead of letting reclassify create it by going
	// around the todo classify gate. The refused write leaves the queue
	// item's classification untouched (nil stays nil).
	root, store := fcFixture(t)
	mdEnableFlag(t, root)
	fcQueue(t, store, factory.BacklogStateQueued)
	// t-slot holds the serial slot: in flight, driven, live lease — the
	// TestMCPRefusal_SerialSlotStructured holder mold. The lease anchor is
	// the real clock (factoryCardNow is the real time on this verb), not the
	// fixture's frozen fcNow — an fcNow-anchored lease reads as expired and
	// the t1828 liveness clause frees the slot. The holder's serial judgment
	// is recorded on its queue item too (the classification-flip world: an
	// absent read derives parallelizable, so an unrecorded holder would not
	// read serial), and the fcRun seed row gives the homestate card its run.
	fcPlace(t, root, homestate.Card{
		CardID: "t-slot", State: homestate.CardRun, Stage: homestate.CardRun,
		OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		LeaseExpiresAt: time.Now().Add(10 * time.Minute).Format(time.RFC3339Nano),
		WorktreePath:   root,
	})
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		rec.Items = append(rec.Items, factory.BacklogItem{
			ID:      "t-slot",
			Text:    "serial slot holder",
			AddedAt: "2026-09-26T00:00:00Z",
			State:   factory.BacklogStatePicked,
		})
		return nil
	}); err != nil {
		t.Fatalf("seed t-slot queue item: %v", err)
	}
	fcClassify(t, store, "t-slot", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	// The add path records a decider-default parallelizable classification at
	// creation (REQ-FDM-002), so the refusal scenario is a recorded
	// parallelizable card commanding serial — a mode transition, not a
	// convergence (different axes), so the gate arms.
	start := mdClass(t, store, "t1")
	if start == nil || start.Mode != factory.ClassModeParallelizable {
		t.Fatalf("fixture precondition: t1 starts recorded-parallelizable, got %v", start)
	}
	if _, _, err := runFactory(t, "reclassify", "t1", "--mode", "serial", "--reason", "wants serial"); err == nil {
		t.Fatal("reclassify to serial over a held slot: want a refusal, got success")
	}
	after := mdClass(t, store, "t1")
	if after == nil || after.Mode != factory.ClassModeParallelizable {
		t.Fatalf("the refused reclassify must leave t1 parallelizable, got %v", after)
	}
}

func TestReclassify_AuditWindowCatchUp(t *testing.T) {
	// The queue-commit→append crash window (REQ-CMV-010): seeded
	// converged-but-unaudited — the queue write committed, the append never
	// ran. The re-run commanding the converged value is admitted as the
	// converged-state audit completion: exactly the owed row, no store
	// mutation, exit zero.
	root, store := mdClassifiedFixture(t)
	seed := *mdClass(t, store, "t1")
	if _, _, err := runFactory(t, "reclassify", "t1", "--mode", "serial", "--priority", "high", "--reason", "catch-up attestation"); err != nil {
		t.Fatalf("converged-but-unaudited re-run refused: %v", err)
	}
	rows := mdAuditRows(t, root, "t1")
	if len(rows) != 1 {
		t.Fatalf("audit rows = %d, want exactly the owed row", len(rows))
	}
	if rows[0].Old != "" || rows[0].New != `{"blocked":true,"mode":"serial","priority":"high"}` || rows[0].Reason != "catch-up attestation" {
		t.Fatalf("catch-up row = %+v, want old=\"\" with the commanded axes and the re-run's reason", rows[0])
	}
	// No store mutation: the classification is exactly as the crashed run
	// left it — stamp and reason included.
	got := mdClass(t, store, "t1")
	if got == nil || *got != seed {
		t.Fatalf("classification re-mutated by the catch-up:\n got %+v\nwant %+v", got, seed)
	}

	// The converged-and-audited duplicate re-run refuses with no second row.
	queueBefore := mdQueueDump(t, store)
	if _, _, err := runFactory(t, "reclassify", "t1", "--mode", "serial", "--priority", "high", "--reason", "again"); err == nil {
		t.Fatal("the converged-and-audited duplicate was admitted")
	}
	if n := len(mdAuditRows(t, root, "t1")); n != 1 {
		t.Fatalf("audit rows = %d after the duplicate refusal, want 1", n)
	}
	if got := mdQueueDump(t, store); got != queueBefore {
		t.Fatalf("duplicate refusal moved the queue")
	}

	// An append failure after the queue commit retries once, then exits
	// non-zero with the audit-loss sentinel, the queue write left in place.
	root2, store2 := mdClassifiedFixture(t)
	attempts := 0
	restore := metadataWriteFault
	metadataWriteFault = func(stage string) error {
		if stage == "audit-append" {
			attempts++
			if attempts <= 2 {
				return context.DeadlineExceeded
			}
		}
		return nil
	}
	t.Cleanup(func() { metadataWriteFault = restore })
	if _, _, err := runFactory(t, "reclassify", "t1", "--priority", "low", "--reason", "r"); err == nil {
		t.Fatal("the audit-loss sentinel did not fire")
	} else if !strings.Contains(err.Error(), "card metadata audit row lost") {
		t.Fatalf("error lacks the audit-loss sentinel: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("append attempts = %d, want the first attempt plus exactly one retry", attempts)
	}
	got2 := mdClass(t, store2, "t1")
	if got2 == nil || got2.Priority != factory.ClassPriorityLow {
		t.Fatalf("queue write not left in place: %+v", got2)
	}
	if n := mdAuditCount(t, root2); n != 0 {
		t.Fatalf("audit rows = %d after the audit-loss sentinel, want 0", n)
	}

	// The retry landing: one failed attempt, the retry succeeds, one row.
	root3, store3 := mdClassifiedFixture(t)
	attempts3 := 0
	restore3 := metadataWriteFault
	metadataWriteFault = func(stage string) error {
		if stage == "audit-append" {
			attempts3++
			if attempts3 == 1 {
				return context.DeadlineExceeded
			}
		}
		return nil
	}
	t.Cleanup(func() { metadataWriteFault = restore3 })
	if _, _, err := runFactory(t, "reclassify", "t1", "--priority", "low", "--reason", "r"); err != nil {
		t.Fatalf("retrying reclassify append: %v", err)
	}
	if n := len(mdAuditRows(t, root3, "t1")); n != 1 {
		t.Fatalf("audit rows = %d after the landed retry, want 1", n)
	}
	_ = store3
}

// SPEC-CARD-METADATA-VERBS-001 M4: the actor capture layer, the lane
// boundary, and the consumer integration — pinning tests for the wiring the
// verbs' audit rows already ride (the capture-layer function shipped with
// the M2 verb seam; these tests observe it GREEN at arrival and the
// lane-boundary red is observed through the mutant probe recorded in
// progress.md §E.2).

func TestMetadataActor_Resolution(t *testing.T) {
	// The lane marker set → the lane label (REQ-CMV-011, the
	// factory-decide capture pattern).
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-9")
	actor, _ := metadataActor()
	if actor != "lane-9" {
		t.Fatalf("actor with the lane marker set = %q, want lane-9", actor)
	}
	// The marker unset → operator.
	t.Setenv(config.EnvMoaiFactoryWorker, "")
	actor, _ = metadataActor()
	if actor != "operator" {
		t.Fatalf("actor with the marker unset = %q, want operator", actor)
	}
	// The session UUID rides the existing session-identity capture when one
	// is resolvable.
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-uuid-1")
	_, session := metadataActor()
	if session != "sess-uuid-1" {
		t.Fatalf("session = %q, want the resolvable UUID sess-uuid-1", session)
	}
}

func TestMetadataVerbs_LaneBoundary_Refused(t *testing.T) {
	// REQ-CMV-015 (decision-index Q2, Default applied): in a lane session
	// both verbs refuse with the lane-boundary sentinel — and because the
	// boundary check runs ahead of the flag gate, the refusal is the lane
	// sentinel even though no local config enables the verbs.
	root, store := mdPickedFixture(t)
	queueBefore, cardsBefore := mdQueueDump(t, store), mdCardsDump(t, root)
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-9")
	for _, args := range [][]string{
		{"set-spec", "t1", "--spec", "SPEC-L-001", "--reason", "r", "--run", fcRun},
		{"reclassify", "t1", "--priority", "high", "--reason", "r"},
	} {
		_, _, err := runFactory(t, args...)
		if err == nil {
			t.Fatalf("%s was admitted in a lane session", args[0])
		}
		if !strings.Contains(err.Error(), "lane boundary") {
			t.Fatalf("%s refusal lacks the lane-boundary sentinel (the boundary check must run ahead of the flag gate): %v", args[0], err)
		}
	}
	// Nothing written: both stores row-unchanged and no audit row.
	if got := mdQueueDump(t, store); got != queueBefore {
		t.Fatalf("queue moved in a lane session:\n got %s\nwant %s", got, queueBefore)
	}
	if got := mdCardsDump(t, root); got != cardsBefore {
		t.Fatalf("record moved in a lane session:\n got %s\nwant %s", got, cardsBefore)
	}
	if n := mdAuditCount(t, root); n != 0 {
		t.Fatalf("audit rows = %d in a lane session, want 0", n)
	}
}

func TestSetSpec_ConsumerResolves(t *testing.T) {
	// REQ-CMV-006: the contract-report resolution path (cardToSpecID, the
	// store cardToSpecID reads) resolves the new spec id after set-spec.
	root, _ := mdPickedFixture(t)
	mdEnableFlag(t, root)
	if _, _, err := runFactory(t, "set-spec", "t1", "--spec", "SPEC-CMV-777", "--reason", "r", "--run", fcRun); err != nil {
		t.Fatalf("set-spec: %v", err)
	}
	specID, cause := cardToSpecID("t1")
	if specID != "SPEC-CMV-777" {
		t.Fatalf("cardToSpecID = (%q, %q), want (SPEC-CMV-777, \"\")", specID, cause)
	}
}
