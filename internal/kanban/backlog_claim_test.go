// backlog_claim_test.go — SPEC-TODO-CLAIM-LEASE-001 M2/M5 acceptance tests:
// the store-level claim family (Claim / RenewLease / ReclaimExpired), all
// routed through the Mutate flock chokepoint (REQ-TCL-004).
//
// Expiry is driven by seeding lease_expires_at directly through Mutate (the
// engine-bypass seed pattern the AC-TST-012 golden fixture established) — no
// clock seam, no sleeping. C4 is pinned by an unparseable expiry reading as
// EXPIRED; AC-TCL-002's exactly-one-wins and partial-write detection live in
// the claim-contention phase of TestConcurrencyStress.
package kanban

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// claimFixture returns a store over a fresh temp queue.
func claimFixture(t *testing.T) *BacklogStore {
	t.Helper()
	return NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
}

// seedQueuedCards adds texts as queued cards, returning the issued ids in
// order (t1, t2, ...).
func seedQueuedCards(t *testing.T, store *BacklogStore, texts ...string) []string {
	t.Helper()
	ids := make([]string, 0, len(texts))
	for _, text := range texts {
		item, _, err := store.Add(text)
		if err != nil {
			t.Fatalf("Add(%q): %v", text, err)
		}
		ids = append(ids, item.ID)
	}
	return ids
}

// stampLease stamps a lease directly through Mutate — state picked with the
// given holder and RFC 3339 expiry, bypassing the claim family so a test
// controls the clock by construction.
func stampLease(t *testing.T, store *BacklogStore, id, holder, expiresAt string) {
	t.Helper()
	if err := store.Mutate(func(rec *BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID != id {
				continue
			}
			h, e, pa := holder, expiresAt, "2026-09-29T00:00:00Z"
			rec.Items[i].State = BacklogStatePicked
			rec.Items[i].PickedBy = &h
			rec.Items[i].LeaseExpiresAt = &e
			rec.Items[i].PickedAt = &pa
			return nil
		}
		return fmt.Errorf("no backlog item %s", id)
	}); err != nil {
		t.Fatalf("stamp lease on %s: %v", id, err)
	}
}

// leaseAt renders an expiry offset from now in the store's stamp format.
func leaseAt(offset time.Duration) string {
	return time.Now().Add(offset).UTC().Format(time.RFC3339)
}

// recordJSON renders the record through the pure reader for byte-identity
// assertions.
func recordJSON(t *testing.T, store *BacklogStore) string {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("LoadPure: %v", err)
	}
	raw, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal record: %v", err)
	}
	return string(raw)
}

// findItem returns a copy of the named live item.
func findItem(t *testing.T, store *BacklogStore, id string) BacklogItem {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("LoadPure: %v", err)
	}
	for _, it := range rec.Items {
		if it.ID == id {
			return it
		}
	}
	t.Fatalf("no live item %s", id)
	return BacklogItem{}
}

// assertLeaseCleared asserts reclamation's cleared-field set: the lease
// fields plus the pick-time spec_id (generalized unpick, REQ-TCL-009).
func assertLeaseCleared(t *testing.T, it BacklogItem) {
	t.Helper()
	if it.State != BacklogStateQueued {
		t.Errorf("state = %s, want queued after reclamation", it.State)
	}
	if it.PickedBy != nil || it.LeaseExpiresAt != nil || it.PickedAt != nil || it.SpecID != nil {
		t.Errorf("cleared-field set broken on %s: picked_by=%v lease_expires_at=%v picked_at=%v spec_id=%v — all four must be nil",
			it.ID, it.PickedBy, it.LeaseExpiresAt, it.PickedAt, it.SpecID)
	}
}

// AC-TCL-004 — a lapsed lease returns the card to queued with the full
// cleared-field set, the output names the id and the previous holder, and a
// following claim takes that card (the reclaimed card is the oldest, so the
// same claim that ran the expiry-first pass takes it).
func TestBacklogClaim_ReclaimsExpired(t *testing.T) {
	t.Parallel()
	store := claimFixture(t)
	ids := seedQueuedCards(t, store, "expired card", "fresh card")
	stampLease(t, store, ids[0], "lane-1", leaseAt(-15*time.Minute))

	result, err := store.Claim("lane-2")
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}

	// The reclamation is surfaced with id + previous holder (C5 audit line).
	if len(result.Reclaimed) != 1 {
		t.Fatalf("reclaimed = %+v, want exactly the lapsed t1", result.Reclaimed)
	}
	if result.Reclaimed[0].ItemID != ids[0] || result.Reclaimed[0].PrevHolder != "lane-1" {
		t.Errorf("reclaimed entry = %+v, want id %s prev holder lane-1", result.Reclaimed[0], ids[0])
	}
	// The reclaimed card was reclaimed and then claimed by this same claim
	// (it is the oldest queued card after the expiry-first pass).
	if result.Item.ID != ids[0] {
		t.Errorf("claimed = %s, want the reclaimed oldest card %s", result.Item.ID, ids[0])
	}
	if result.Item.PickedBy == nil || *result.Item.PickedBy != "lane-2" {
		t.Errorf("claimed picked_by = %v, want lane-2", result.Item.PickedBy)
	}
	if result.Item.LeaseExpiresAt == nil {
		t.Fatal("claimed lease_expires_at = nil, want ≈ now + DefaultFactoryLeaseDuration")
	}
	expiresAt, perr := time.Parse(time.RFC3339, *result.Item.LeaseExpiresAt)
	if perr != nil {
		t.Fatalf("claimed expiry %q is not RFC 3339: %v", *result.Item.LeaseExpiresAt, perr)
	}
	if d := time.Until(expiresAt); d < 14*time.Minute || d > 16*time.Minute {
		t.Errorf("claimed lease duration = %v, want ≈ 15m (DefaultFactoryLeaseDuration)", d)
	}
	if result.Item.PickedAt == nil {
		t.Error("claimed picked_at = nil, want stamped")
	}

	// The untouched card stays queued, untouched.
	fresh := findItem(t, store, ids[1])
	if fresh.State != BacklogStateQueued || fresh.PickedBy != nil {
		t.Errorf("fresh card = %+v, want queued and untouched", fresh)
	}
}

// AC-TCL-011 — the live-lease guard. Arm 1: a claim against a live-leased
// queue refuses with the raced indication and the record stays byte-
// identical. Arm 2: reclaiming card A's LAPSED lease leaves live-leased
// card B's lease fields untouched — no cross-holder mutation. Arm 3 (the
// operator unpick carve-out) is pinned by the existing
// TestTodoUnpick_RevertsPickedToQueued / TestTodoUnpick_RefusalsLeaveFileUntouched
// guards in internal/cli, cited by acceptance and left name-invariant.
func TestBacklogClaim_LiveLeaseGuard(t *testing.T) {
	t.Parallel()
	// Arm 1 — raced refusal, byte-identical record.
	store := claimFixture(t)
	ids := seedQueuedCards(t, store, "live leased card")
	// leaseAt() re-reads the wall clock on every call, so the stamped value
	// and the expected value must come from ONE call: two calls a second
	// apart disagree once -race slows the run, flaking the comparison below.
	wantLease := leaseAt(15 * time.Minute)
	stampLease(t, store, ids[0], "lane-1", wantLease)
	before := recordJSON(t, store)

	if _, err := store.Claim("lane-2"); !errors.Is(err, ErrClaimRaced) {
		t.Fatalf("Claim over a live lease: err = %v, want ErrClaimRaced", err)
	}
	if after := recordJSON(t, store); after != before {
		t.Errorf("raced claim mutated the record\nbefore: %s\n after: %s", before, after)
	}
	guarded := findItem(t, store, ids[0])
	if guarded.PickedBy == nil || *guarded.PickedBy != "lane-1" ||
		guarded.LeaseExpiresAt == nil || *guarded.LeaseExpiresAt != wantLease {
		t.Errorf("live lease fields moved under a refused claim: %+v", guarded)
	}

	// Arm 2 — reclaiming A's lapsed lease leaves B's live lease untouched.
	store2 := claimFixture(t)
	ids2 := seedQueuedCards(t, store2, "lapsed card", "live card")
	stampLease(t, store2, ids2[0], "lane-1", leaseAt(-1*time.Minute))
	stampLease(t, store2, ids2[1], "lane-2", leaseAt(15*time.Minute))
	bLeaseBefore := findItem(t, store2, ids2[1])

	reclaimed, err := store2.ReclaimExpired()
	if err != nil {
		t.Fatalf("ReclaimExpired: %v", err)
	}
	if len(reclaimed) != 1 || reclaimed[0].ItemID != ids2[0] || reclaimed[0].PrevHolder != "lane-1" {
		t.Fatalf("reclaimed = %+v, want exactly %s (prev holder lane-1)", reclaimed, ids2[0])
	}
	assertLeaseCleared(t, findItem(t, store2, ids2[0]))
	bAfter := findItem(t, store2, ids2[1])
	if bAfter.PickedBy == nil || *bAfter.PickedBy != *bLeaseBefore.PickedBy ||
		bAfter.LeaseExpiresAt == nil || *bAfter.LeaseExpiresAt != *bLeaseBefore.LeaseExpiresAt {
		t.Errorf("live-leased card B mutated by A's reclamation: before %+v after %+v", bLeaseBefore, bAfter)
	}
	if bAfter.State != BacklogStatePicked {
		t.Errorf("live-leased card B state = %s, want still picked", bAfter.State)
	}
}

// AC-TCL-005 — an unparseable lease expiry is judged EXPIRED (C4, the
// factory default; not the slot-lease conservative inverse).
func TestBacklogClaim_UnparseableExpiryExpired(t *testing.T) {
	t.Parallel()
	store := claimFixture(t)
	ids := seedQueuedCards(t, store, "corrupt lease card")
	stampLease(t, store, ids[0], "lane-1", "not-a-timestamp")

	reclaimed, err := store.ReclaimExpired()
	if err != nil {
		t.Fatalf("ReclaimExpired: %v", err)
	}
	if len(reclaimed) != 1 || reclaimed[0].ItemID != ids[0] {
		t.Fatalf("reclaimed = %+v, want %s judged EXPIRED and reclaimed", reclaimed, ids[0])
	}
	assertLeaseCleared(t, findItem(t, store, ids[0]))

	// The claim family agrees with the standalone reclaim: a claim over the
	// re-seeded corrupt lease succeeds by reclaiming first.
	stampLease(t, store, ids[0], "lane-3", "also-not-a-time")
	result, err := store.Claim("lane-4")
	if err != nil {
		t.Fatalf("Claim over an unparseable expiry: %v", err)
	}
	if len(result.Reclaimed) != 1 || result.Reclaimed[0].ItemID != ids[0] {
		t.Errorf("claim-time reclaimed = %+v, want the unparseable lease judged EXPIRED", result.Reclaimed)
	}
}

// AC-TCL-006 — renew. A live lease extends by the lease duration and no
// other field moves; a foreign holder is refused; and the renew-after-expiry
// arm does NOT extend — the expiry-first return commits first, then the
// refusal lands.
func TestBacklogClaim_RenewAfterExpiry(t *testing.T) {
	t.Parallel()
	store := claimFixture(t)
	ids := seedQueuedCards(t, store, "renewable card")

	// Renew-after-expiry arm: the lapsed lease is returned (committed) and
	// the renew refuses.
	stampLease(t, store, ids[0], "lane-1", leaseAt(-2*time.Minute))
	before := recordJSON(t, store)
	if _, err := store.RenewLease(ids[0], "lane-1"); !errors.Is(err, ErrLeaseExpired) {
		t.Fatalf("RenewLease on a lapsed lease: err = %v, want ErrLeaseExpired", err)
	}
	if after := recordJSON(t, store); after == before {
		t.Error("renew-after-expiry refused without committing the expiry-first return")
	}
	assertLeaseCleared(t, findItem(t, store, ids[0]))

	// Live renew arm: extend, and no other field moves. Renewal resets the
	// expiry to now+DefaultFactoryLeaseDuration — the absolute-reset shape
	// the RenewLease pattern this SPEC ports (card_transition.go) uses — so
	// a lease stamped 3 minutes out reads ≈15 minutes after the renewal.
	stampLease(t, store, ids[0], "lane-1", leaseAt(3*time.Minute))
	live := findItem(t, store, ids[0])
	renewedAt := time.Now()
	renewed, err := store.RenewLease(ids[0], "lane-1")
	if err != nil {
		t.Fatalf("RenewLease on a live lease: %v", err)
	}
	if renewed.Item.ID != ids[0] {
		t.Errorf("renewed = %s, want %s", renewed.Item.ID, ids[0])
	}
	newExpiry, perr := time.Parse(time.RFC3339, *renewed.Item.LeaseExpiresAt)
	if perr != nil {
		t.Fatalf("renewed expiry %q is not RFC 3339: %v", *renewed.Item.LeaseExpiresAt, perr)
	}
	if d := newExpiry.Sub(renewedAt); d < 14*time.Minute || d > 16*time.Minute {
		t.Errorf("renewed lease = %v after renewal, want ≈ 15m (DefaultFactoryLeaseDuration)", d)
	}
	oldExpiry, _ := time.Parse(time.RFC3339, *live.LeaseExpiresAt)
	if !newExpiry.After(oldExpiry) {
		t.Errorf("renewal did not move the expiry forward: %s → %s", *live.LeaseExpiresAt, *renewed.Item.LeaseExpiresAt)
	}
	after := findItem(t, store, ids[0])
	if after.PickedBy == nil || *after.PickedBy != "lane-1" {
		t.Errorf("renew moved picked_by: %v", after.PickedBy)
	}
	if after.PickedAt == nil || *after.PickedAt != *live.PickedAt {
		t.Errorf("renew moved picked_at: %v → %v", live.PickedAt, after.PickedAt)
	}
	if after.State != BacklogStatePicked {
		t.Errorf("renew moved state: %s", after.State)
	}

	// Holder mismatch arm: refused with no change.
	beforeMismatch := recordJSON(t, store)
	if _, err := store.RenewLease(ids[0], "lane-9"); !errors.Is(err, ErrLeaseHolder) {
		t.Fatalf("RenewLease by a foreign holder: err = %v, want ErrLeaseHolder", err)
	}
	if afterMismatch := recordJSON(t, store); afterMismatch != beforeMismatch {
		t.Error("a refused renewal mutated the record")
	}
}

// backlogItemTuples reads the items table as id → column → TEXT value (NULL
// rendered as ""), with the column set taken from pragma_table_info — the
// full-table pre/post tuple surface the partial-write detector compares
// (AC-TCL-002).
func backlogItemTuples(t *testing.T, store *BacklogStore) map[string]map[string]string {
	t.Helper()
	eng, err := openBacklogReader(backlogSQLitePath(store.Path()))
	if err != nil {
		t.Fatalf("open tuple reader: %v", err)
	}
	defer func() { _ = eng.close() }()
	ctx := context.Background()
	var columns []string
	colRows, err := eng.db.QueryContext(ctx, `SELECT name FROM pragma_table_info('items') ORDER BY cid`)
	if err != nil {
		t.Fatalf("read pragma_table_info: %v", err)
	}
	for colRows.Next() {
		var name string
		if err := colRows.Scan(&name); err != nil {
			t.Fatalf("scan column name: %v", err)
		}
		columns = append(columns, name)
	}
	if err := colRows.Err(); err != nil {
		t.Fatalf("iterate columns: %v", err)
	}
	_ = colRows.Close()
	if len(columns) == 0 {
		t.Fatal("pragma_table_info(items) returned no columns")
	}
	quoted := make([]string, len(columns))
	for i, c := range columns {
		quoted[i] = `"` + c + `"`
	}
	rows, err := eng.db.QueryContext(ctx,
		`SELECT `+strings.Join(quoted, ", ")+` FROM items ORDER BY seq`)
	if err != nil {
		t.Fatalf("read items tuples: %v", err)
	}
	defer func() { _ = rows.Close() }()
	out := map[string]map[string]string{}
	for rows.Next() {
		values := make([]sql.NullString, len(columns))
		dests := make([]any, len(columns))
		for i := range values {
			dests[i] = &values[i]
		}
		if err := rows.Scan(dests...); err != nil {
			t.Fatalf("scan tuple: %v", err)
		}
		row := map[string]string{}
		for i, c := range columns {
			row[c] = ""
			if values[i].Valid {
				row[c] = values[i].String
			}
		}
		out[row["id"]] = row
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate tuples: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("tuple reader saw no rows")
	}
	return out
}

// assertNoPartialClaimWrite is the AC-TCL-002 partial-write detector: every
// post-contention row is a valid TERMINAL tuple — queued with an empty lease
// field set, or picked with holder, expiry, and pick stamp all present. A
// row in any other shape is a partial write the flock failed to make atomic.
func assertNoPartialClaimWrite(t *testing.T, tuples map[string]map[string]string, wantPicked int) {
	t.Helper()
	picked := 0
	for id, row := range tuples {
		switch row["state"] {
		case "queued":
			if row["picked_by"] != "" || row["lease_expires_at"] != "" {
				t.Errorf("partial write: queued row %s carries lease fields picked_by=%q lease_expires_at=%q",
					id, row["picked_by"], row["lease_expires_at"])
			}
		case "picked":
			picked++
			if row["picked_by"] == "" || row["lease_expires_at"] == "" || row["picked_at"] == "" {
				t.Errorf("partial write: picked row %s is not fully stamped (picked_by=%q lease_expires_at=%q picked_at=%q)",
					id, row["picked_by"], row["lease_expires_at"], row["picked_at"])
			}
		default:
			t.Errorf("unexpected state %q on row %s", row["state"], id)
		}
	}
	if picked != wantPicked {
		t.Errorf("picked rows = %d, want %d (exactly-one-wins per card)", picked, wantPicked)
	}
}
