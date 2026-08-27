// backlog_sqlite_test.go — the statusline's read across BOTH queue layouts,
// and the per-render latency budget (SPEC-TODO-SQLITE-001 REQ-TOSQ-009,
// AC-TOSQ-012, C-2).
//
// The statusline runs as a fresh process on every render. Two properties
// therefore decide whether the storage swap is safe here, and neither is
// visible from the happy path alone: the read must stay PURE (a render must
// never perform the one-time storage cutover) and it must stay bounded (a
// render is on the operator's critical path).
package statusline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// seedQueue writes n cards through the store, cycling the three states, and
// returns the expected picked/queued counts.
func seedQueue(t *testing.T, root string, n int) (picked, queued int) {
	t.Helper()
	store := kanban.NewBacklogStore(kanban.BacklogPathForRootAdopting(root))
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for i := 0; i < n; i++ {
			rec.LastSeq++
			state := kanban.BacklogStateQueued
			switch i % 3 {
			case 1:
				state = kanban.BacklogStatePicked
			case 2:
				state = kanban.BacklogStateDropped
			}
			rec.Items = append(rec.Items, kanban.BacklogItem{
				ID:      "t" + itoa(rec.LastSeq),
				Text:    "synthesized card",
				AddedAt: "2026-01-02T03:04:05Z",
				State:   state,
			})
			switch state {
			case kanban.BacklogStatePicked:
				picked++
			case kanban.BacklogStateQueued:
				queued++
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("seed queue: %v", err)
	}
	return picked, queued
}

// itoa is strconv.Itoa without the import churn in this small file.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// REQ-TOSQ-009: the render reads a migrated (database) queue correctly, and
// leaves it exactly as it found it.
func TestResolveBacklogCounts_DatabaseLayoutIsReadPurely(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	wantPicked, wantQueued := seedQueue(t, root, 12)

	stateDir := kanban.StateDirForRoot(root)
	before := dirNames(t, stateDir)

	got := resolveBacklogCounts(root)
	if !got.Available {
		t.Fatal("Available = false, want true for a readable queue")
	}
	if got.Picked != wantPicked || got.Queued != wantQueued {
		t.Fatalf("counts = picked %d / queued %d, want %d / %d", got.Picked, got.Queued, wantPicked, wantQueued)
	}

	if after := dirNames(t, stateDir); !equalNames(before, after) {
		t.Fatalf("the render changed the state directory %v -> %v — it must move nothing", before, after)
	}
}

// REQ-TOSQ-015: a render against a queue still on the LEGACY directory reads
// it in place and does not relocate it. An operator who has not yet run a
// `moai todo` verb since upgrading still gets their counts, and their
// directory is still where they left it.
func TestResolveBacklogCounts_LegacyLayoutIsReadWithoutRelocating(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	legacy := kanban.LegacyStateDirForRoot(root)
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatalf("seed legacy dir: %v", err)
	}
	body := `{"version":1,"last_seq":3,"items":[
	  {"id":"t1","text":"a","added_at":"T","spec_id":null,"state":"queued"},
	  {"id":"t2","text":"b","added_at":"T","spec_id":"SPEC-EXAMPLE-001","state":"picked"},
	  {"id":"t3","text":"c","added_at":"T","spec_id":null,"state":"dropped"}
	],"findings":[]}`
	if err := os.WriteFile(filepath.Join(legacy, "backlog.json"), []byte(body), 0o644); err != nil {
		t.Fatalf("seed legacy queue: %v", err)
	}

	got := resolveBacklogCounts(root)
	if !got.Available || got.Picked != 1 || got.Queued != 1 {
		t.Fatalf("counts = %+v, want picked 1 / queued 1 / available", got)
	}
	if _, err := os.Stat(kanban.StateDirForRoot(root)); err == nil {
		t.Error("the render relocated the state directory")
	}
	if _, err := os.Stat(filepath.Join(legacy, "backlog.db")); err == nil {
		t.Error("the render migrated the queue into a database")
	}
}

// Fail-open: an unreadable queue renders nothing rather than an
// authoritative-looking zero. "No cards" and "could not read the cards" are
// different claims, and Available is what separates them.
func TestResolveBacklogCounts_UnreadableIsUnavailableNotZero(t *testing.T) {
	t.Parallel()

	t.Run("absent queue", func(t *testing.T) {
		if got := resolveBacklogCounts(t.TempDir()); got.Available {
			t.Fatalf("counts = %+v, want Available=false for an absent queue", got)
		}
	})

	t.Run("corrupt database", func(t *testing.T) {
		root := t.TempDir()
		seedQueue(t, root, 2)
		db := filepath.Join(kanban.StateDirForRoot(root), "backlog.db")
		if err := os.WriteFile(db, []byte("not a database"), 0o600); err != nil {
			t.Fatalf("corrupt the database: %v", err)
		}
		got := resolveBacklogCounts(root)
		if got.Available {
			t.Fatalf("counts = %+v, want Available=false for a corrupt queue", got)
		}
		if got.Picked != 0 || got.Queued != 0 {
			t.Fatalf("counts = %+v, want zeroes alongside Available=false", got)
		}
	})

	t.Run("empty board root", func(t *testing.T) {
		if got := resolveBacklogCounts(""); got.Available {
			t.Fatalf("counts = %+v, want Available=false for an empty root", got)
		}
	})
}

// AC-TOSQ-012 / C-2: the ADDED per-render latency stays inside budget on a
// >=500-item fixture — median added <= 10ms, hard ceiling 25ms.
//
// C-2 words the budget as "added per-render latency of the backlog-count read
// ... VERSUS the current whole-file JSON read". That is a differential, and
// measuring it as one is not a technicality — it is what makes the number mean
// anything. An absolute wall-clock threshold measures the machine as much as
// the code: run inside a full concurrent suite, this same read showed a p95 of
// 29ms against 0.6ms when run alone. A test asserting the absolute figure
// would pass on a quiet laptop and fail in CI, which is a flaky gate rather
// than a budget.
//
// Both arms are measured INTERLEAVED against equivalent fixtures, so whatever
// load the machine is under lands on both and largely cancels. The absolute
// distributions are logged either way, because the differential is the gate
// but the absolute numbers are what a reader wants to see.
func TestResolveBacklogCounts_LatencyBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("latency measurement skipped under -short")
	}

	// Arm A: the database layout, as shipped.
	dbRoot := t.TempDir()
	seedQueue(t, dbRoot, 500)

	// Arm B: the same queue in the legacy whole-file JSON layout — the
	// baseline C-2 names. Built by exporting arm A's content, so the two arms
	// hold identical cards rather than merely similar ones.
	jsonRoot := t.TempDir()
	legacyDir := kanban.LegacyStateDirForRoot(jsonRoot)
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatalf("seed legacy dir: %v", err)
	}
	src, err := kanban.NewBacklogStore(kanban.BacklogPathForRoot(dbRoot)).LoadPure()
	if err != nil {
		t.Fatalf("read arm A for the baseline fixture: %v", err)
	}
	encoded, err := json.Marshal(src)
	if err != nil {
		t.Fatalf("encode baseline fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, "backlog.json"), encoded, 0o644); err != nil {
		t.Fatalf("write baseline fixture: %v", err)
	}
	if base := kanban.BacklogCountsForRoot(jsonRoot); !base.Available {
		t.Fatal("the JSON baseline fixture is unreadable; the comparison would be meaningless")
	}

	const runs = 41
	dbSamples := make([]time.Duration, 0, runs)
	jsonSamples := make([]time.Duration, 0, runs)
	for i := 0; i < runs; i++ {
		start := time.Now()
		if got := resolveBacklogCounts(dbRoot); !got.Available {
			t.Fatalf("database render %d read nothing", i)
		}
		dbSamples = append(dbSamples, time.Since(start))

		start = time.Now()
		if got := resolveBacklogCounts(jsonRoot); !got.Available {
			t.Fatalf("json render %d read nothing", i)
		}
		jsonSamples = append(jsonSamples, time.Since(start))
	}

	sort.Slice(dbSamples, func(a, b int) bool { return dbSamples[a] < dbSamples[b] })
	sort.Slice(jsonSamples, func(a, b int) bool { return jsonSamples[a] < jsonSamples[b] })
	mid, p95i := len(dbSamples)/2, (len(dbSamples)*95)/100

	addedMedian := dbSamples[mid] - jsonSamples[mid]
	addedP95 := dbSamples[p95i] - jsonSamples[p95i]

	t.Logf("AC-TOSQ-012: 500-item fixture, %d interleaved render pairs\n"+
		"  database : median %v  p95 %v  max %v\n"+
		"  json     : median %v  p95 %v  max %v\n"+
		"  ADDED    : median %v  p95 %v   (budget: median <=10ms, ceiling 25ms)",
		runs,
		dbSamples[mid], dbSamples[p95i], dbSamples[len(dbSamples)-1],
		jsonSamples[mid], jsonSamples[p95i], jsonSamples[len(jsonSamples)-1],
		addedMedian, addedP95)

	if addedMedian > 10*time.Millisecond {
		t.Errorf("added median %v exceeds the 10ms budget (C-2)", addedMedian)
	}
	if addedP95 > 25*time.Millisecond {
		t.Errorf("added p95 %v exceeds the 25ms hard ceiling (C-2)", addedP95)
	}
}

// dirNames lists a directory's entries, sorted.
func dirNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}

// equalNames compares two sorted name lists.
func equalNames(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
