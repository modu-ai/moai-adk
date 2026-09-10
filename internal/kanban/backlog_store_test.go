// backlog_store_test.go — the lock-guarded backlog queue store
// (SPEC-KANBAN-TODO-CLI-001 REQ-TODO-006..013, M1).
//
// Store-level slices of AC-TODO-007..013. The CLI-process-level ACs
// (AC-TODO-001, AC-TODO-003..006) land with M2's command wiring; here the
// same behaviors are proven against the store the verbs will call.
package kanban

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// newTestBacklogStore builds a store over a temp-dir backlog file.
func newTestBacklogStore(t *testing.T) *BacklogStore {
	t.Helper()
	return NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
}

// seedBacklogFile writes raw bytes as the backlog file and returns its
// checksum for byte-identity assertions.
func seedBacklogFile(t *testing.T, store *BacklogStore, content string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(store.Path()), 0o755); err != nil {
		t.Fatalf("seed backlog dir: %v", err)
	}
	if err := os.WriteFile(store.Path(), []byte(content), 0o644); err != nil {
		t.Fatalf("seed backlog file: %v", err)
	}
	return checksumBacklogFile(t, store)
}

// checksumBacklogFile digests the backlog file bytes.
func checksumBacklogFile(t *testing.T, store *BacklogStore) string {
	t.Helper()
	raw, err := os.ReadFile(store.Path())
	if err != nil {
		t.Fatalf("read backlog file for checksum: %v", err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// readBacklogDirNames lists the backlog directory's entry names.
func readBacklogDirNames(t *testing.T, store *BacklogStore) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Dir(store.Path()))
	if err != nil {
		t.Fatalf("read backlog dir: %v", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// productionShapedBacklog is a version-1 record WITHOUT last_seq: three
// items, all three states, spec_id null and non-null — the load-compat
// fixture for REQ-TODO-009/013.
const productionShapedBacklog = `{
  "version": 1,
  "items": [
    {
      "id": "t2",
      "text": "Rework the auth middleware error paths",
      "added_at": "2026-08-14T01:02:03Z",
      "spec_id": null,
      "state": "queued"
    },
    {
      "id": "t5",
      "text": "Audit the session registry locking",
      "added_at": "2026-08-14T01:03:04Z",
      "spec_id": "SPEC-KANBAN-TODO-CLI-001",
      "state": "picked"
    },
    {
      "id": "t19",
      "text": "Retire the legacy flag",
      "added_at": "2026-08-14T01:04:05Z",
      "spec_id": null,
      "state": "dropped"
    }
  ]
}
`

// TestBacklogLoad_MissingFileIsEmptyQueue — REQ-TODO-012 missing half: an
// absent backlog file is an empty queue, never an error.
func TestBacklogLoad_MissingFileIsEmptyQueue(t *testing.T) {
	t.Parallel()
	store := newTestBacklogStore(t)

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load(missing file) error = %v, want empty queue", err)
	}
	if rec.Version != 1 {
		t.Fatalf("Version = %d, want 1", rec.Version)
	}
	if len(rec.Items) != 0 {
		t.Fatalf("Items = %d, want 0", len(rec.Items))
	}
}

// TestBacklogAdd_CreatesVersion1File — REQ-TODO-012: add against a missing
// file creates a valid version-1 file containing exactly one item.
func TestBacklogAdd_CreatesVersion1File(t *testing.T) {
	t.Parallel()
	store := newTestBacklogStore(t)

	item, pos, err := store.Add("first card")
	if err != nil {
		t.Fatalf("Add(missing file) error = %v, want file created", err)
	}
	if item.ID != "t1" {
		t.Fatalf("first issued id = %q, want t1", item.ID)
	}
	if item.Text != "first card" {
		t.Fatalf("item.Text = %q, want %q", item.Text, "first card")
	}
	if item.State != BacklogStateQueued {
		t.Fatalf("item.State = %q, want %q", item.State, BacklogStateQueued)
	}
	if item.SpecID != nil {
		t.Fatalf("item.SpecID = %v, want nil", item.SpecID)
	}
	if item.AddedAt == "" {
		t.Fatal("item.AddedAt empty, want RFC3339 timestamp")
	}
	if _, err := time.Parse(time.RFC3339, item.AddedAt); err != nil {
		t.Fatalf("item.AddedAt %q is not RFC3339: %v", item.AddedAt, err)
	}
	if pos != 1 {
		t.Fatalf("queue position = %d, want 1", pos)
	}

	// The physical carrier changed with the storage swap; the record it
	// carries did not. Read it back through the store's own surface — that
	// surface is what every caller contracts on (REQ-TOSQ-010), and asserting
	// against the file bytes would only re-assert which engine is underneath.
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load after Add: %v", err)
	}
	if rec.Version != 1 {
		t.Fatalf("stored version = %d, want 1", rec.Version)
	}
	if rec.LastSeq != 1 {
		t.Fatalf("stored last_seq = %d, want 1", rec.LastSeq)
	}
	if len(rec.Items) != 1 || rec.Items[0].ID != "t1" {
		t.Fatalf("stored items = %+v, want exactly [t1]", rec.Items)
	}
	if rec.Items[0] != *item {
		t.Fatalf("stored item %+v != issued item %+v", rec.Items[0], *item)
	}

	// The physical artifact is the sibling database, and nothing else claims
	// to be the queue.
	dbPath := filepath.Join(filepath.Dir(store.Path()), "backlog.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("engine artifact absent at %s: %v", dbPath, err)
	}
	if _, err := os.Stat(store.Path()); err == nil {
		t.Fatal("a legacy backlog.json was created — the engine is the only writer now")
	}
}

// TestBacklogLock_TimeoutNamesLockPath — REQ-TODO-006/007: a foreign holder
// beyond the bounded retry window surfaces as an explicit error naming the
// lock artifact path.
func TestBacklogLock_TimeoutNamesLockPath(t *testing.T) {
	t.Parallel()
	store := newTestBacklogStore(t)

	impl, err := acquireBoardLockImpl(store.LockPath())
	if err != nil {
		t.Fatalf("hold foreign lock: %v", err)
	}
	foreign := &BoardLock{path: store.LockPath(), impl: impl}
	defer func() { _ = foreign.Release() }()

	start := time.Now()
	_, _, err = store.Add("blocked card")
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("Add under a held lock err = nil, want bounded-retry timeout")
	}
	if !IsBoardLockHeld(err) {
		t.Fatalf("err = %v, want the lock-held sentinel", err)
	}
	if !strings.Contains(err.Error(), store.LockPath()) {
		t.Fatalf("err %q does not name the lock artifact path %q", err, store.LockPath())
	}
	if elapsed < 500*time.Millisecond {
		t.Fatalf("Add returned after %v, before the ~1s bounded window elapsed", elapsed)
	}
	// The failed mutation must not have created or altered the backlog file.
	if _, statErr := os.Stat(store.Path()); statErr == nil {
		t.Fatal("backlog file appeared despite lock timeout")
	}
}

// TestBacklogMutate_LoserSerializedNotFailed — AC-TODO-007's composition
// half at store level: a mutation racing a short-lived holder retries within
// the window and completes once the holder releases.
func TestBacklogMutate_LoserSerializedNotFailed(t *testing.T) {
	t.Parallel()
	store := newTestBacklogStore(t)

	impl, err := acquireBoardLockImpl(store.LockPath())
	if err != nil {
		t.Fatalf("hold foreign lock: %v", err)
	}
	foreign := &BoardLock{path: store.LockPath(), impl: impl}
	go func() {
		time.Sleep(80 * time.Millisecond)
		_ = foreign.Release()
	}()

	item, _, err := store.Add("waiting card")
	if err != nil {
		t.Fatalf("Add racing a short holder failed: %v, want serialized success", err)
	}
	if item.ID != "t1" {
		t.Fatalf("issued id = %q, want t1", item.ID)
	}
}

// TestBacklogIDIssuedUnderLock_SequentialMutations — REQ-TODO-008: ids come
// from last_seq read inside the locked mutation; two sequential adds yield
// t1 then t2 with no reuse.
func TestBacklogIDIssuedUnderLock_SequentialMutations(t *testing.T) {
	t.Parallel()
	store := newTestBacklogStore(t)

	first, _, err := store.Add("one")
	if err != nil {
		t.Fatalf("first Add: %v", err)
	}
	second, _, err := store.Add("two")
	if err != nil {
		t.Fatalf("second Add: %v", err)
	}
	if first.ID != "t1" || second.ID != "t2" {
		t.Fatalf("issued ids = %q, %q; want t1, t2", first.ID, second.ID)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if rec.LastSeq != 2 {
		t.Fatalf("persisted last_seq = %d, want 2", rec.LastSeq)
	}
}

// TestBacklogHighWater_SurvivesRemoval — REQ-TODO-008: `done` removes rows,
// so the persisted high-water mark — not max-present-id — decides the next
// id. Remove t20, the next add issues t21, never t20 again.
func TestBacklogHighWater_SurvivesRemoval(t *testing.T) {
	t.Parallel()
	store := newTestBacklogStore(t)
	seedBacklogFile(t, store, `{
  "version": 1,
  "last_seq": 19,
  "items": [
    {"id": "t20", "text": "to be removed", "added_at": "2026-08-14T02:00:00Z", "spec_id": null, "state": "queued"}
  ]
}
`)

	if err := store.Mutate(func(rec *BacklogRecord) error {
		kept := rec.Items[:0]
		for _, it := range rec.Items {
			if it.ID != "t20" {
				kept = append(kept, it)
			}
		}
		rec.Items = kept
		return nil
	}); err != nil {
		t.Fatalf("Mutate(remove t20): %v", err)
	}

	item, _, err := store.Add("after removal")
	if err != nil {
		t.Fatalf("Add after removal: %v", err)
	}
	if item.ID != "t21" {
		t.Fatalf("id after removing t20 = %q, want t21 — removal never enables reuse", item.ID)
	}
}

// TestBacklogHighWater_DeriveOnAbsent — REQ-TODO-009: a version-1 file
// predating last_seq derives the initial mark from the maximum existing item
// id (t19 here) and persists it on the first write; the next issued id is t20.
func TestBacklogHighWater_DeriveOnAbsent(t *testing.T) {
	t.Parallel()
	store := newTestBacklogStore(t)
	seedBacklogFile(t, store, productionShapedBacklog)

	item, _, err := store.Add("derived mark")
	if err != nil {
		t.Fatalf("Add over derive-on-absent file: %v", err)
	}
	if item.ID != "t20" {
		t.Fatalf("id after deriving from max t19 = %q, want t20", item.ID)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if rec.LastSeq != 20 {
		t.Fatalf("persisted last_seq = %d, want 20 (derived then advanced)", rec.LastSeq)
	}
}

// TestBacklogHighWater_HandEditedLowSeq — acceptance.md §C bullet 4: a
// hand-edited last_seq BELOW a present id must not mint a duplicate; the
// effective mark is max(persisted, max-present).
func TestBacklogHighWater_HandEditedLowSeq(t *testing.T) {
	t.Parallel()
	store := newTestBacklogStore(t)
	seedBacklogFile(t, store, `{
  "version": 1,
  "last_seq": 3,
  "items": [
    {"id": "t10", "text": "hand edited high", "added_at": "2026-08-14T03:00:00Z", "spec_id": null, "state": "queued"}
  ]
}
`)

	item, _, err := store.Add("must clear t10")
	if err != nil {
		t.Fatalf("Add over hand-edited file: %v", err)
	}
	if item.ID != "t11" {
		t.Fatalf("id = %q, want t11 (greater than every present id)", item.ID)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if rec.LastSeq != 11 {
		t.Fatalf("persisted last_seq = %d, want 11", rec.LastSeq)
	}
	seen := map[string]bool{}
	for _, it := range rec.Items {
		if seen[it.ID] {
			t.Fatalf("duplicate id %q after hand-edited regression", it.ID)
		}
		seen[it.ID] = true
	}
}

// TestBacklogMalformed_ReportedEverywhereFileUntouched — REQ-TODO-012
// malformed half: every path reports the parse error and leaves the file
// byte-identical; no verb resets, no silent repair exists.
func TestBacklogMalformed_ReportedEverywhereFileUntouched(t *testing.T) {
	t.Parallel()
	store := newTestBacklogStore(t)
	before := seedBacklogFile(t, store, `{"version":1,"items":[ TRUNCATED`)

	if _, err := store.Load(); err == nil {
		t.Fatal("Load(malformed) err = nil, want parse error")
	}
	if _, _, err := store.Add("into malformed"); err == nil {
		t.Fatal("Add(malformed) err = nil, want parse error")
	}
	if err := store.Mutate(func(rec *BacklogRecord) error { return nil }); err == nil {
		t.Fatal("Mutate(malformed) err = nil, want parse error")
	}

	if after := checksumBacklogFile(t, store); after != before {
		t.Fatal("malformed backlog file changed after failed verbs — never silently reset")
	}
}

// TestBacklogRoundTrip_PreservesFieldsAndVersion — REQ-TODO-013: a
// production-shaped version-1 file survives load → mutate → write with every
// pre-existing item's five fields unchanged, version still 1, no new per-item
// field, and last_seq as the only top-level addition.
func TestBacklogRoundTrip_PreservesFieldsAndVersion(t *testing.T) {
	t.Parallel()
	store := newTestBacklogStore(t)
	seedBacklogFile(t, store, productionShapedBacklog)

	if _, _, err := store.Add("new row"); err != nil {
		t.Fatalf("Add: %v", err)
	}

	// Read back through the store surface: the seeded JSON migrated into the
	// engine and the mutation landed on top of it, so this asserts the whole
	// seed -> migrate -> mutate chain preserved every field.
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load after round trip: %v", err)
	}
	if rec.Version != 1 {
		t.Fatalf("version = %d, want 1 (no version bump)", rec.Version)
	}
	if rec.LastSeq != 20 {
		t.Fatalf("last_seq = %d, want 20", rec.LastSeq)
	}
	if len(rec.Items) != 4 {
		t.Fatalf("items = %d, want 4 (3 preserved + 1 added)", len(rec.Items))
	}
	if rec.Items[0].ID != "t2" || rec.Items[0].State != BacklogStateQueued || rec.Items[0].SpecID != nil {
		t.Fatalf("item t2 not preserved verbatim: %+v", rec.Items[0])
	}
	if rec.Items[1].ID != "t5" || rec.Items[1].State != BacklogStatePicked ||
		rec.Items[1].SpecID == nil || *rec.Items[1].SpecID != "SPEC-KANBAN-TODO-CLI-001" {
		t.Fatalf("item t5 not preserved verbatim: %+v", rec.Items[1])
	}
	if rec.Items[2].ID != "t19" || rec.Items[2].State != BacklogStateDropped {
		t.Fatalf("item t19 not preserved verbatim: %+v", rec.Items[2])
	}
	// The seed carries no last_seq, so the mark derives from max-present (t19)
	// and the new card takes t20 — the derive-on-absent rule surviving the
	// migration intact.
	if rec.Items[3].ID != "t20" {
		t.Fatalf("added item id = %q, want t20 (derived past max-present t19)", rec.Items[3].ID)
	}
}

// TestBacklogWrite_NoTmpResidue — REQ-TODO-010: the atomic write leaves the
// backlog file and its lock sibling — nothing else; no .tmp residue.
func TestBacklogWrite_NoTmpResidue(t *testing.T) {
	t.Parallel()
	store := newTestBacklogStore(t)

	if _, _, err := store.Add("residue probe"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	// The engine's own artifacts are the database and its WAL siblings; the
	// lock is unchanged. Anything else in the directory is residue.
	allowed := map[string]bool{
		"backlog.db":     true,
		"backlog.db-wal": true,
		"backlog.db-shm": true,
		"backlog.lock":   true,
	}
	for _, name := range readBacklogDirNames(t, store) {
		if !allowed[name] {
			t.Fatalf("backlog dir holds leaked file %q — no write path may leave residue", name)
		}
	}
}

// TestBacklogConcurrentAdd_UniqueIDs — REQ-TODO-008 store-level slice of
// AC-TODO-001/008: concurrent adds through the store all land, all ids are
// distinct, and every text survives.
func TestBacklogConcurrentAdd_UniqueIDs(t *testing.T) {
	t.Parallel()
	store := newTestBacklogStore(t)

	const adds = 8
	ids := make([]string, adds)
	errs := make([]error, adds)
	var wg sync.WaitGroup
	for i := 0; i < adds; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			item, _, err := store.Add(fmt.Sprintf("card %d", i))
			ids[i], errs[i] = item.ID, err
		}(i)
	}
	wg.Wait()

	seen := map[string]bool{}
	for i := 0; i < adds; i++ {
		if errs[i] != nil {
			t.Fatalf("concurrent Add %d failed: %v", i, errs[i])
		}
		if seen[ids[i]] {
			t.Fatalf("duplicate id %q across concurrent adds", ids[i])
		}
		seen[ids[i]] = true
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load after concurrent adds: %v", err)
	}
	if len(rec.Items) != adds {
		t.Fatalf("items = %d, want %d — concurrent adds lost rows", len(rec.Items), adds)
	}
	if rec.LastSeq != adds {
		t.Fatalf("last_seq = %d, want %d", rec.LastSeq, adds)
	}
	texts := map[string]bool{}
	for _, it := range rec.Items {
		texts[it.Text] = true
	}
	for i := 0; i < adds; i++ {
		if !texts[fmt.Sprintf("card %d", i)] {
			t.Fatalf("text %q lost across concurrent adds", fmt.Sprintf("card %d", i))
		}
	}
}

// TestBacklogStore_NoLeadRoleGuard — REQ-TODO-011: the backlog applies no
// requireLeadRole-equivalent gate. This test process declares no role and
// holds no kanban identity at all; the write must succeed anyway (explicit
// contrast with the board's sole-writer guard — board files untouched).
func TestBacklogStore_NoLeadRoleGuard(t *testing.T) {
	t.Parallel()
	store := newTestBacklogStore(t)

	item, _, err := store.Add("anonymous writer")
	if err != nil {
		t.Fatalf("Add without any role declaration: %v — the backlog is the operator's queue, any session may write", err)
	}
	if item.ID != "t1" {
		t.Fatalf("issued id = %q, want t1", item.ID)
	}
	if IsNotSoleWriter(err) {
		t.Fatal("backlog write hit the board's sole-writer refusal")
	}
}

// TestNewBacklogStore_SweepsLegacyLockArtifact (t106): constructing a store
// removes a superseded legacy lock artifact (backlog.json.lock) sitting
// beside the queue file, while the live lock name (backlog.lock) is left to
// the lock machinery. Installs that lived through the lock rename carry the
// stale zero-byte twin; the sweep settles the directory on one name.
func TestNewBacklogStore_SweepsLegacyLockArtifact(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, legacyBacklogLockFileName)
	live := filepath.Join(dir, backlogLockFileName)
	if err := os.WriteFile(legacy, nil, 0o600); err != nil {
		t.Fatalf("seed legacy lock: %v", err)
	}
	if err := os.WriteFile(live, []byte("holder"), 0o600); err != nil {
		t.Fatalf("seed live lock: %v", err)
	}

	store := NewBacklogStore(filepath.Join(dir, "backlog.json"))

	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatalf("legacy lock artifact survived store construction (stat err = %v)", err)
	}
	if _, err := os.Stat(live); err != nil {
		t.Fatalf("live lock artifact disturbed by the sweep: %v", err)
	}
	if store == nil {
		t.Fatal("store is nil")
	}
}
