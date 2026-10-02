package kanban

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestFactoryFreeSlots is the shared-cluster AC for the t85 lead loop's
// picker: an empty or missing registry reads as all-free (fail-open), a live
// claim makes its slot busy, a dead claim is pruned so its slot reads free,
// and claims outside 1..workers do not widen the result.
func TestFactoryFreeSlots(t *testing.T) {
	t.Parallel()

	alwaysAlive := func(int) bool { return true }
	neverAlive := func(int) bool { return false }

	t.Run("missing registry means all free", func(t *testing.T) {
		t.Parallel()
		got := FactoryFreeSlots(t.TempDir(), 3, alwaysAlive)
		if !slices.Equal(got, []int{1, 2, 3}) {
			t.Errorf("FactoryFreeSlots(missing, 3) = %v, want [1 2 3]", got)
		}
	})

	t.Run("live claim makes its slot busy", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := SaveFactoryRegistry(FactoryRegistryPath(root), map[string]FactoryLaneEntry{
			"lane-1": {PID: 11100},
			"lane-3": {PID: 11101},
		}); err != nil {
			t.Fatalf("seed registry: %v", err)
		}
		got := FactoryFreeSlots(root, 4, alwaysAlive)
		if !slices.Equal(got, []int{2, 4}) {
			t.Errorf("FactoryFreeSlots with live 1,3 = %v, want [2 4]", got)
		}
	})

	t.Run("dead claim is pruned and reads free", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := SaveFactoryRegistry(FactoryRegistryPath(root), map[string]FactoryLaneEntry{
			"lane-2": {PID: 11100},
		}); err != nil {
			t.Fatalf("seed registry: %v", err)
		}
		got := FactoryFreeSlots(root, 3, neverAlive)
		if !slices.Equal(got, []int{1, 2, 3}) {
			t.Errorf("FactoryFreeSlots with dead 2 = %v, want [1 2 3]", got)
		}
	})

	t.Run("claims beyond workers do not widen the result", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := SaveFactoryRegistry(FactoryRegistryPath(root), map[string]FactoryLaneEntry{
			"lane-9": {PID: 11100},
		}); err != nil {
			t.Fatalf("seed registry: %v", err)
		}
		got := FactoryFreeSlots(root, 2, alwaysAlive)
		if !slices.Equal(got, []int{1, 2}) {
			t.Errorf("FactoryFreeSlots with out-of-range claim = %v, want [1 2]", got)
		}
	})
}

func TestClaimFactoryLaneWithinBounds(t *testing.T) {
	root := t.TempDir()
	alive := func(int) bool { return true }
	// An older claim outside the current run's two slots must not make the
	// automatic join skip the first available in-range number.
	if _, err := ClaimFactoryLane(root, "lane-9", false, 9009, "old-run", alive); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"lane-1", "lane-2"} {
		claim, err := ClaimFactoryLaneWithin(root, "", true, os.Getpid(), "run", 2, alive)
		if err != nil || claim.Label != want {
			t.Fatalf("automatic claim = (%q, %v), want %s", claim.Label, err, want)
		}
	}
	before := LoadFactoryRegistry(FactoryRegistryPath(root))
	for _, tc := range []struct {
		label string
		auto  bool
		want  string
	}{
		{"", true, "no free lane slots"},
		{"lane-1", false, "already occupied"},
		{"lane-3", false, "outside the allowed slots"},
	} {
		if _, err := ClaimFactoryLaneWithin(root, tc.label, tc.auto, os.Getpid(), "run", 2, alive); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("claim %q auto=%v: err=%v, want %q", tc.label, tc.auto, err, tc.want)
		}
		after := LoadFactoryRegistry(FactoryRegistryPath(root))
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("failed claim changed registry: before=%v after=%v", before, after)
		}
	}
}

func TestClaimFactoryLaneWithinConcurrentOneSlot(t *testing.T) {
	root := t.TempDir()
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := ClaimFactoryLaneWithin(root, "", true, os.Getpid(), "run", 1, func(int) bool { return true })
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	succeeded, full := 0, 0
	for err := range results {
		switch {
		case err == nil:
			succeeded++
		case strings.Contains(err.Error(), "no free lane slots"):
			full++
		default:
			t.Fatalf("unexpected concurrent claim error: %v", err)
		}
	}
	if succeeded != 1 || full != 1 {
		t.Fatalf("concurrent claims: success=%d full=%d, want 1 each", succeeded, full)
	}
}

// TestPruneFactoryDeadClaims pins the shared prune rule on its own: only
// live, positively-numbered claims survive.
func TestPruneFactoryDeadClaims(t *testing.T) {
	t.Parallel()

	reg := map[string]FactoryLaneEntry{
		"lane-1": {PID: 11100}, // live
		"lane-2": {PID: 11101}, // dead
		"lane-3": {PID: 0},     // non-positive
		"lane-4": {PID: -5},    // negative
	}
	got := PruneFactoryDeadClaims(reg, func(pid int) bool { return pid == 11100 })
	if len(got) != 1 {
		t.Fatalf("pruned registry = %v, want only lane-1", got)
	}
	if _, ok := got["lane-1"]; !ok {
		t.Errorf("live claim lane-1 must survive, got %v", got)
	}
}

// TestBacklogQueuedCountSharedShape covers the one-call queued-count both the
// kanban notice and the factory lead loop render from: missing file reads 0,
// only state queued counts, and the path helper lands where the store reads.
func TestBacklogQueuedCountSharedShape(t *testing.T) {
	t.Parallel()

	t.Run("missing file and empty root read as zero", func(t *testing.T) {
		t.Parallel()
		if got := QueuedBacklogCountForRoot(t.TempDir()); got != 0 {
			t.Errorf("QueuedBacklogCountForRoot(missing) = %d, want 0", got)
		}
		if got := QueuedBacklogCountForRoot(""); got != 0 {
			t.Errorf("QueuedBacklogCountForRoot(\"\") = %d, want 0", got)
		}
	})

	t.Run("only queued items count", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := NewBacklogStore(BacklogPathForRoot(root)).Mutate(func(rec *BacklogRecord) error {
			rec.Items = []BacklogItem{
				{ID: "t1", State: BacklogStateQueued},
				{ID: "t2", State: BacklogStatePicked},
				{ID: "t3", State: BacklogStateQueued},
				{ID: "t4", State: BacklogStateDropped},
				{ID: "t5", State: BacklogStateQueued},
			}
			return nil
		}); err != nil {
			t.Fatalf("seed backlog: %v", err)
		}
		if got := QueuedBacklogCountForRoot(root); got != 3 {
			t.Errorf("QueuedBacklogCountForRoot = %d, want 3 (queued only)", got)
		}
	})

	t.Run("path helper lands under .moai/state/todo", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		want := filepath.Join(root, ".moai", "state", "todo", "backlog.json")
		if got := BacklogPathForRoot(root); got != want {
			t.Errorf("BacklogPathForRoot = %q, want %q", got, want)
		}
	})
}

// TestFactoryRegistryRoundTrip pins the moved cluster's load/save shape: the
// file the v1 cli code wrote is the file the shared cluster reads.
func TestFactoryRegistryRoundTrip(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	seed := map[string]FactoryLaneEntry{
		"lane-1": {PID: os.Getpid(), RegisteredAt: "2026-08-17T00:00:00Z"},
	}
	if err := SaveFactoryRegistry(FactoryRegistryPath(root), seed); err != nil {
		t.Fatalf("save: %v", err)
	}
	got := LoadFactoryRegistry(FactoryRegistryPath(root))
	if len(got) != 1 || got["lane-1"].PID != os.Getpid() {
		t.Errorf("round trip = %v, want lane-1 at this pid", got)
	}

	// A malformed file fails open to an empty registry (never an error).
	if err := os.WriteFile(FactoryRegistryPath(root), []byte("not json"), 0o600); err != nil {
		t.Fatalf("plant malformed file: %v", err)
	}
	if got := LoadFactoryRegistry(FactoryRegistryPath(root)); len(got) != 0 {
		t.Errorf("malformed registry = %v, want empty (fail-open)", got)
	}
}

func TestClaimFactoryWorkerNameConcurrentClaimsAreUnique(t *testing.T) {
	root := t.TempDir()
	const workers = 10
	results := make(chan string, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(pid int) {
			defer wg.Done()
			label, err := ClaimFactoryLaneName(root, "lane-1", pid, "testrun", func(int) bool { return true })
			results <- label
			errs <- err
		}(10000 + i)
	}
	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	seen := map[string]bool{}
	for label := range results {
		if seen[label] {
			t.Fatalf("duplicate concurrent claim %q", label)
		}
		seen[label] = true
	}
	if len(seen) != workers {
		t.Fatalf("unique claims=%d, want %d: %v", len(seen), workers, seen)
	}
}

// qasWorkerBackend reads the backend column the registry holds for label.
func qasWorkerBackend(t *testing.T, root, label string) string {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	var backend string
	if err := db.DB.QueryRow(`SELECT backend FROM workers WHERE label=?`, label).Scan(&backend); err != nil {
		t.Fatalf("read backend of %s: %v", label, err)
	}
	return backend
}

// qasGuardedRegistry returns a project root whose registry carries the two
// SPEC-QUOTA-AWARE-SCHEDULING-001 AC-QAS-023 test triggers: a BEFORE UPDATE on
// workers that aborts (a claim that inserts the row and then updates the
// backend fails) and an AFTER INSERT on workers that aborts when the inserted
// backend is empty (the insert statement itself must carry the backend).
func qasGuardedRegistry(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	for _, ddl := range []string{
		`CREATE TRIGGER qas_no_update BEFORE UPDATE ON workers BEGIN SELECT RAISE(ABORT, 'qas: update on workers'); END`,
		`CREATE TRIGGER qas_insert_needs_backend AFTER INSERT ON workers WHEN NEW.backend = '' BEGIN SELECT RAISE(ABORT, 'qas: insert without backend'); END`,
	} {
		if _, err := db.DB.Exec(ddl); err != nil {
			t.Fatalf("install trigger: %v", err)
		}
	}
	return root
}

// AC-QAS-023 — the lane claim records the backend in its own insert and
// nothing else changes (REQ-QAS-023).
func TestQAS_AC023_ClaimRecordsBackend(t *testing.T) {
	alive := func(int) bool { return true }

	t.Run("claim_with_backend_in_the_insert", func(t *testing.T) {
		root := qasGuardedRegistry(t)
		// Positive control: the insert trigger bites a claim that carries no
		// backend, so the passes below prove the insert statement itself wrote it.
		if _, err := ClaimFactoryLane(root, "", true, 9001, "run-c", alive); err == nil || !strings.Contains(err.Error(), "insert without backend") {
			t.Fatalf("empty-backend claim on the guarded registry: err = %v, want the insert trigger to abort it", err)
		}
		for i, backend := range []string{BackendClaude, BackendGLM, BackendGPT} {
			claim, err := ClaimFactoryLaneWithBackend(root, "", true, 9100+i, "run-c", backend, alive)
			if err != nil {
				t.Fatalf("claim as %s: %v", backend, err)
			}
			if got := qasWorkerBackend(t, root, claim.Label); got != backend {
				t.Errorf("%s recorded backend %q, want %q", claim.Label, got, backend)
			}
		}
		claim, err := ClaimFactoryLaneWithinWithBackend(root, "", true, 9200, "run-c", 8, BackendGLM, alive)
		if err != nil {
			t.Fatalf("bounded claim as glm: %v", err)
		}
		if got := qasWorkerBackend(t, root, claim.Label); got != BackendGLM {
			t.Errorf("bounded %s recorded backend %q, want %q", claim.Label, got, BackendGLM)
		}
		// Positive control: the update trigger bites a later backend update.
		db, err := homestate.OpenFactory(root)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = db.Close() }()
		if _, err := db.DB.Exec(`UPDATE workers SET backend='mutant' WHERE label=?`, claim.Label); err == nil || !strings.Contains(err.Error(), "update on workers") {
			t.Fatalf("update on the guarded registry: err = %v, want the update trigger to abort it", err)
		}
	})

	t.Run("empty_without_backend", func(t *testing.T) {
		root := t.TempDir()
		a, err := ClaimFactoryLane(root, "", true, 9301, "run-e", alive)
		if err != nil {
			t.Fatal(err)
		}
		b, err := ClaimFactoryLaneWithin(root, "", true, 9302, "run-e", 4, alive)
		if err != nil {
			t.Fatal(err)
		}
		cLabel, err := ClaimFactoryLaneName(root, "lane-9", 9303, "run-e", alive)
		if err != nil {
			t.Fatal(err)
		}
		// Positive control: a backend-carrying claim in the same registry reads
		// back non-empty, so the empty readings below are not a failed read.
		d, err := ClaimFactoryLaneWithBackend(root, "", true, 9304, "run-e", BackendGLM, alive)
		if err != nil {
			t.Fatal(err)
		}
		if got := qasWorkerBackend(t, root, d.Label); got != BackendGLM {
			t.Fatalf("control %s recorded backend %q, want %q", d.Label, got, BackendGLM)
		}
		for _, label := range []string{a.Label, b.Label, cLabel} {
			if got := qasWorkerBackend(t, root, label); got != "" {
				t.Errorf("%s claimed without a backend recorded %q, want empty", label, got)
			}
		}
	})

	t.Run("codex_token_normalized", func(t *testing.T) {
		root := qasGuardedRegistry(t)
		claim, err := ClaimFactoryLaneWithinWithBackend(root, "", true, 9401, "run-x", 4, "codex", alive)
		if err != nil {
			t.Fatalf("claim as the codex token: %v", err)
		}
		if got := qasWorkerBackend(t, root, claim.Label); got != BackendGPT {
			t.Errorf("codex token recorded %q, want %q", got, BackendGPT)
		}
		other, err := ClaimFactoryLaneWithBackend(root, "", true, 9402, "run-x", "codex", alive)
		if err != nil {
			t.Fatalf("unbounded claim as the codex token: %v", err)
		}
		if got := qasWorkerBackend(t, root, other.Label); got != BackendGPT {
			t.Errorf("unbounded codex token recorded %q, want %q", got, BackendGPT)
		}
	})

	t.Run("concurrent_claims_each_carry_backend", func(t *testing.T) {
		root := qasGuardedRegistry(t)
		const claimants = 9
		backends := []string{BackendClaude, BackendGLM, BackendGPT}
		var mu sync.Mutex
		want := map[string]string{}
		errs := make(chan error, claimants)
		var wg sync.WaitGroup
		for i := 0; i < claimants; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				backend := backends[i%len(backends)]
				claim, err := ClaimFactoryLaneWithBackend(root, "", true, 9500+i, "run-k", backend, alive)
				if err == nil {
					mu.Lock()
					want[claim.Label] = backend
					mu.Unlock()
				}
				errs <- err
			}(i)
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			if err != nil {
				t.Fatal(err)
			}
		}
		if len(want) != claimants {
			t.Fatalf("unique claims = %d, want %d: %v", len(want), claimants, want)
		}
		for label, backend := range want {
			if got := qasWorkerBackend(t, root, label); got != backend {
				t.Errorf("%s recorded backend %q, want %q", label, got, backend)
			}
		}
	})

	t.Run("registry_round_trips_backend", func(t *testing.T) {
		// Plan debt N9: SaveFactoryRegistry carries the backend, and a row
		// written by the pre-REQ-QAS-023 insert reads back as unknown (empty).
		root := t.TempDir()
		seed := map[string]FactoryLaneEntry{
			"lane-1": {PID: os.Getpid(), RegisteredAt: "2026-10-02T00:00:00Z", Backend: BackendGLM},
			"lane-2": {PID: os.Getpid(), RegisteredAt: "2026-10-02T00:00:00Z"},
		}
		if err := SaveFactoryRegistry(FactoryRegistryPath(root), seed); err != nil {
			t.Fatalf("save: %v", err)
		}
		got := LoadFactoryRegistry(FactoryRegistryPath(root))
		if got["lane-1"].Backend != BackendGLM {
			t.Errorf("lane-1 backend after round trip = %q, want %q", got["lane-1"].Backend, BackendGLM)
		}
		if got["lane-2"].Backend != "" {
			t.Errorf("lane-2 backend after round trip = %q, want empty", got["lane-2"].Backend)
		}
		db, err := homestate.OpenFactory(root)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = db.Close() }()
		if _, err := db.DB.Exec(`INSERT INTO workers(label,pid,registered_at,heartbeat_at,run_id) VALUES('lane-3',?,?,?,?)`,
			os.Getpid(), "2026-10-02T00:00:00Z", "2026-10-02T00:00:00Z", "old-run"); err != nil {
			t.Fatalf("insert pre-change row: %v", err)
		}
		if got := LoadFactoryRegistry(FactoryRegistryPath(root)); got["lane-3"].Backend != "" {
			t.Errorf("pre-change row backend = %q, want empty (unknown, never backfilled)", got["lane-3"].Backend)
		}
	})

	t.Run("schema_unchanged", func(t *testing.T) {
		root := qasGuardedRegistry(t)
		if _, err := ClaimFactoryLaneWithBackend(root, "", true, 9601, "run-s", BackendClaude, alive); err != nil {
			t.Fatal(err)
		}
		db, err := homestate.OpenFactory(root)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = db.Close() }()
		var version string
		if err := db.DB.QueryRow(`SELECT value FROM meta WHERE key='schema_version'`).Scan(&version); err != nil {
			t.Fatal(err)
		}
		if version != "5" {
			t.Errorf("meta.schema_version = %q after a backend claim, want 5", version)
		}
		raw, err := os.ReadFile(filepath.Join("..", "homestate", "factory.go"))
		if err != nil {
			t.Fatalf("read the factory DDL source: %v", err)
		}
		ddl := string(raw)
		// Positive control: the search sees the ALTER TABLE statements that do exist.
		if !strings.Contains(ddl, "ALTER TABLE runs") {
			t.Fatal("control: the factory DDL source carries no ALTER TABLE runs statement; the search is blind")
		}
		if strings.Contains(ddl, "ALTER TABLE workers") {
			t.Error("the factory DDL carries an ALTER TABLE workers statement; REQ-QAS-023 needs no schema change")
		}
	})
}
