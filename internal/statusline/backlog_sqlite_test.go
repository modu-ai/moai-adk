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

// AC-TOSQ-012 / C-2: the per-render read stays inside the latency budget on a
// >=500-item fixture — median <= 10ms, hard ceiling 25ms. Measured on this
// machine, in this run; the number is logged so the evidence is the observed
// distribution rather than a pass/fail with nothing behind it.
func TestResolveBacklogCounts_LatencyBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("latency measurement skipped under -short")
	}
	root := t.TempDir()
	seedQueue(t, root, 500)

	const runs = 41
	samples := make([]time.Duration, 0, runs)
	for i := 0; i < runs; i++ {
		start := time.Now()
		got := resolveBacklogCounts(root)
		samples = append(samples, time.Since(start))
		if !got.Available {
			t.Fatalf("render %d read nothing", i)
		}
	}
	sort.Slice(samples, func(a, b int) bool { return samples[a] < samples[b] })
	median := samples[len(samples)/2]
	p95 := samples[(len(samples)*95)/100]

	t.Logf("AC-TOSQ-012: 500-item fixture, %d renders — median %v, p95 %v, max %v (budget: median <=10ms, ceiling 25ms)",
		runs, median, p95, samples[len(samples)-1])

	if median > 10*time.Millisecond {
		t.Errorf("median render read %v exceeds the 10ms budget (C-2)", median)
	}
	if p95 > 25*time.Millisecond {
		t.Errorf("p95 render read %v exceeds the 25ms hard ceiling (C-2)", p95)
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
