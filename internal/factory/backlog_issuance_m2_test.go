package factory

import (
	"context"
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// M2 storage-layer acceptance tests (SPEC-TODO-CARD-ISSUANCE-001,
// AC-TCI-008..011). The retrofit/round-trip/parity trio follows the
// classification column's recorded schema decision (card t1332) — the same
// guarantees, one column pair later.

func issuanceSeed() *BacklogIssuance {
	lines := 42
	return &BacklogIssuance{
		SpawnedBy:  "t1",
		Origin:     "follow-up",
		SizeLines:  &lines,
		Files:      []string{"internal/cli/todo.go"},
		DropReason: "operator decision",
	}
}

func ptrString(s string) *string { return &s }

func issuanceArchivedFixture(t *testing.T) (string, *BacklogStore) {
	t.Helper()
	root := t.TempDir()
	store := NewBacklogStore(BacklogPathForRoot(root))
	if _, _, err := store.Add("card with issuance attributes"); err != nil {
		t.Fatal(err)
	}
	if err := store.Mutate(func(rec *BacklogRecord) error {
		rec.Items[0].Issuance = issuanceSeed()
		rec.Findings = append(rec.Findings, BacklogFinding{
			SubjectID: "t1", RelatedID: "t1", Relation: "near-duplicate",
			Source: "jev", Score: 0.9, Note: "seed", At: "2026-10-05T12:00:00+09:00",
			Disposition: ptrString("accept"),
		})
		return rec.ArchiveCard("t1")
	}); err != nil {
		t.Fatal(err)
	}
	return root, store
}

// AC-TCI-008: a database opened after the change carries both new columns —
// the ensure path added them at open, before any read or write.
func TestBacklogIssuanceColumnRetrofit(t *testing.T) {
	root := t.TempDir()
	if _, _, err := NewBacklogStore(BacklogPathForRoot(root)).Add("pre-column card"); err != nil {
		t.Fatal(err)
	}
	eng, err := openBacklogEngine(BacklogPathForRoot(root))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = eng.close() }()
	ctx := context.Background()
	for _, tc := range []struct{ table, column string }{
		{"items", "issuance"}, {"archived_items", "issuance"},
		{"findings", "disposition"}, {"archived_findings", "disposition"},
	} {
		present, err := eng.hasColumn(ctx, tc.table, tc.column)
		if err != nil {
			t.Fatal(err)
		}
		if !present {
			t.Errorf("%s lacks the %s column after the retrofit", tc.table, tc.column)
		}
	}
}

// AC-TCI-008: the pure reader runs no schema change — LoadPure over an
// opened database leaves the queue storage byte-identical.
func TestBacklogIssuancePureReaderNoDDL(t *testing.T) {
	root := t.TempDir()
	store := NewBacklogStore(BacklogPathForRoot(root))
	if _, _, err := store.Add("reader card"); err != nil {
		t.Fatal(err)
	}
	before := moaiTreeDigest(t, root)
	if _, err := NewBacklogStore(BacklogPathForRoot(root)).LoadPure(); err != nil {
		t.Fatal(err)
	}
	after := moaiTreeDigest(t, root)
	if string(before) != string(after) {
		t.Errorf("LoadPure changed the queue storage")
	}
}

// moaiTreeDigest hashes every file under root/.moai — the byte-identity
// input for read-only contracts (the store layout may be db, json, or both).
func moaiTreeDigest(t *testing.T, root string) []byte {
	t.Helper()
	h := sha256.New()
	base := filepath.Join(root, ".moai")
	err := filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		h.Write([]byte(path))
		h.Write(data)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", base, err)
	}
	return h.Sum(nil)
}

// AC-TCI-009: archiving a card and restoring it carries the issuance
// attributes and the finding dispositions unchanged.
func TestBacklogIssuanceArchiveRestoreRoundTrip(t *testing.T) {
	_, store := issuanceArchivedFixture(t)
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Archived) != 1 {
		t.Fatalf("archived = %d, want 1", len(rec.Archived))
	}
	entry := rec.Archived[0]
	if entry.Item.Issuance == nil || entry.Item.Issuance.Origin != "follow-up" {
		t.Errorf("archived issuance = %+v, want the follow-up origin", entry.Item.Issuance)
	}
	if len(entry.Findings) != 1 || entry.Findings[0].Finding.Disposition == nil ||
		*entry.Findings[0].Finding.Disposition != "accept" {
		t.Fatalf("archived finding disposition missing: %+v", entry.Findings)
	}
	if err := store.Mutate(func(r *BacklogRecord) error { return r.RestoreCard("t1") }); err != nil {
		t.Fatal(err)
	}
	rec, err = store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if rec.Items[0].Issuance == nil || rec.Items[0].Issuance.Origin != "follow-up" {
		t.Errorf("restored issuance = %+v", rec.Items[0].Issuance)
	}
	if len(rec.Findings) != 1 || rec.Findings[0].Disposition == nil ||
		*rec.Findings[0].Disposition != "accept" {
		t.Errorf("restored finding disposition = %+v", rec.Findings)
	}
}

// AC-TCI-009: one queue's in-memory record, its write-back and the re-read
// agree on the new storage — the parity assertion over the written and
// re-read records covers issuance and disposition by construction.
func TestBacklogParityCoversIssuanceAndDisposition(t *testing.T) {
	_, store := issuanceArchivedFixture(t)
	written, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	// Force a whole-record write-back of what was read, then read again:
	// the parity assertion refuses a record whose JSON projection and rows
	// disagree.
	if err := store.Mutate(func(rec *BacklogRecord) error {
		*rec = *written
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	reloaded, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if err := assertBacklogParity(written, reloaded); err != nil {
		t.Errorf("parity over the new storage: %v", err)
	}
	if len(written.Archived) == 1 && len(reloaded.Archived) == 1 {
		wIss, rIss := written.Archived[0].Item.Issuance, reloaded.Archived[0].Item.Issuance
		if (wIss == nil) != (rIss == nil) {
			t.Fatalf("archived issuance presence diverged: %+v vs %+v", wIss, rIss)
		}
		if wIss != nil && rIss != nil &&
			(wIss.Origin != rIss.Origin || wIss.SpawnedBy != rIss.SpawnedBy ||
				wIss.DropReason != rIss.DropReason || strings.Join(wIss.Files, ",") != strings.Join(rIss.Files, ",")) {
			t.Errorf("archived issuance diverged: %+v vs %+v", wIss, rIss)
		}
	}
}

// AC-TCI-007: absence is NULL, never {} and never "".
func TestBacklogIssuanceStoredAsNullWhenAbsent(t *testing.T) {
	_, store := issuanceArchivedFixture(t)
	if _, _, err := store.Add("card without issuance"); err != nil {
		t.Fatal(err)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range rec.Items {
		if it.Text == "card without issuance" && it.Issuance != nil {
			t.Errorf("absent issuance = %+v, want nil", it.Issuance)
		}
	}
}

// AC-TCI-011: a finding with no disposition stays absent.
func TestBacklogDispositionStoredAsNullWhenAbsent(t *testing.T) {
	_, store := issuanceArchivedFixture(t)
	if _, _, err := store.Add("fresh card"); err != nil {
		t.Fatal(err)
	}
	if err := store.Mutate(func(rec *BacklogRecord) error {
		rec.Findings = append(rec.Findings, BacklogFinding{
			SubjectID: "t2", RelatedID: "t2", Relation: "relates-to",
			Source: "agent", At: "2026-10-05T12:00:00+09:00",
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for i := range rec.Findings {
		if rec.Findings[i].Disposition != nil {
			t.Errorf("finding %d carries a disposition it never recorded: %q",
				i, *rec.Findings[i].Disposition)
		}
	}
}

// AC-TCI-010 (closed-at): the accessor derives the closing time from the
// archive stamp, else the drop stamp, else "unknown".
func TestCardClosedAtAccessor(t *testing.T) {
	archived := "2026-10-05T12:00:00+09:00"
	dropped := "2026-10-04T09:00:00+09:00"
	if got := CardClosedAt(BacklogItem{DroppedAt: &dropped}, &archived); got != archived {
		t.Errorf("archived stamp wins: %q", got)
	}
	if got := CardClosedAt(BacklogItem{DroppedAt: &dropped}, nil); got != dropped {
		t.Errorf("drop stamp: %q", got)
	}
	if got := CardClosedAt(BacklogItem{}, nil); got != "unknown" {
		t.Errorf("neither stamp: %q, want unknown", got)
	}
}

// AC-TCI-011: recording a disposition changes no card, no finding relation
// and no queue order.
func TestFindingDispositionRecordOnly(t *testing.T) {
	_, store := todoLikeFixture(t)
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	orderBefore := itemOrder(rec)
	findingsBefore := len(rec.Findings)
	if err := RecordFindingDisposition(store, "t2", "t1", "accept"); err != nil {
		t.Fatal(err)
	}
	rec, err = store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(itemOrder(rec), ",") != strings.Join(orderBefore, ",") {
		t.Errorf("queue order moved: %v → %v", orderBefore, itemOrder(rec))
	}
	if len(rec.Findings) != findingsBefore {
		t.Errorf("findings count moved: %d → %d", findingsBefore, len(rec.Findings))
	}
	matched := 0
	for i := range rec.Findings {
		f := &rec.Findings[i]
		if f.Names("t1") && f.Names("t2") {
			matched++
			if f.Disposition == nil || *f.Disposition != "accept" {
				t.Errorf("disposition = %+v, want accept", f.Disposition)
			}
		}
	}
	if matched != 1 {
		t.Errorf("matched findings = %d, want 1", matched)
	}
}

func todoLikeFixture(t *testing.T) (string, *BacklogStore) {
	t.Helper()
	root := t.TempDir()
	store := NewBacklogStore(BacklogPathForRoot(root))
	if _, _, err := store.Add("first card for the disposition pair"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Add("second card for the disposition pair"); err != nil {
		t.Fatal(err)
	}
	if err := store.Mutate(func(rec *BacklogRecord) error {
		rec.Findings = append(rec.Findings, BacklogFinding{
			SubjectID: "t2", RelatedID: "t1", Relation: "near-duplicate",
			Source: "jev", Score: 0.85, Note: "pair", At: "2026-10-05T12:00:00+09:00",
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return root, store
}

func itemOrder(rec *BacklogRecord) []string {
	out := make([]string, 0, len(rec.Items))
	for _, it := range rec.Items {
		out = append(out, it.ID)
	}
	return out
}

// AC-TCI-011: a legacy finding carries no disposition until one is recorded.
func TestFindingDispositionAbsentForLegacy(t *testing.T) {
	_, store := todoLikeFixture(t)
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for i := range rec.Findings {
		if rec.Findings[i].Disposition != nil {
			t.Errorf("legacy finding %d grew a disposition", i)
		}
	}
}
