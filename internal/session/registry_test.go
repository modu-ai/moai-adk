package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// newTestRegistry builds a registry bound to a fresh t.TempDir() with a
// FakeClock starting at a fixed reference time. Per AP-MSC-003 + AP-MSC-004.
func newTestRegistry(t *testing.T) (*Registry, *FakeClock) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "active-sessions.json")
	clock := &FakeClock{Current: time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)}
	return NewRegistry(path, clock), clock
}

func mustRead(t *testing.T, r *Registry) []Entry {
	t.Helper()
	entries, err := r.readAllUnlocked()
	if err != nil {
		t.Fatalf("readAllUnlocked: %v", err)
	}
	return entries
}

// TestRegisterSession verifies the happy path of Register: a single
// invocation creates one entry with the canonical schema fields populated.
func TestRegisterSession(t *testing.T) {
	r, clock := newTestRegistry(t)

	cases := []struct {
		name      string
		sessionID string
		specID    string
		phase     string
		wantErr   bool
	}{
		{"happy path with SPEC", "uuid-aaa", "SPEC-X-001", "plan", false},
		{"happy path no SPEC", "uuid-bbb", SpecIDNone, PhaseNone, false},
		{"empty sessionID rejected", "", "SPEC-X-001", "plan", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := r.Register(tc.sessionID, tc.specID, tc.phase)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Register(%q): want error, got nil", tc.sessionID)
				}
				return
			}
			if err != nil {
				t.Fatalf("Register: %v", err)
			}
			entries := mustRead(t, r)
			found := false
			for _, e := range entries {
				if e.SessionID == tc.sessionID {
					found = true
					if e.SpecID != tc.specID || e.Phase != tc.phase {
						t.Errorf("entry mismatch: got spec=%q phase=%q, want spec=%q phase=%q",
							e.SpecID, e.Phase, tc.specID, tc.phase)
					}
					if !e.StartedAt.Equal(clock.Current) {
						t.Errorf("StartedAt: got %v, want %v", e.StartedAt, clock.Current)
					}
					if !e.LastHeartbeat.Equal(clock.Current) {
						t.Errorf("LastHeartbeat: got %v, want %v", e.LastHeartbeat, clock.Current)
					}
				}
			}
			if !found {
				t.Errorf("entry for %q not found", tc.sessionID)
			}
		})
	}
}

// TestRegisterIdempotentOnDuplicate confirms that calling Register twice
// with the same sessionID updates in place (no duplicate entries).
func TestRegisterIdempotentOnDuplicate(t *testing.T) {
	r, clock := newTestRegistry(t)
	if err := r.Register("uuid-dup", "SPEC-A", "plan"); err != nil {
		t.Fatal(err)
	}
	clock.Current = clock.Current.Add(5 * time.Minute)
	if err := r.Register("uuid-dup", "SPEC-A", "run"); err != nil {
		t.Fatal(err)
	}
	entries := mustRead(t, r)
	if len(entries) != 1 {
		t.Fatalf("entry count: got %d, want 1", len(entries))
	}
	if entries[0].Phase != "run" {
		t.Errorf("phase: got %q, want %q", entries[0].Phase, "run")
	}
}

// TestHeartbeat verifies that Heartbeat updates only LastHeartbeat and
// preserves all other fields (REQ-COORD-004, AC-COORD-002).
func TestHeartbeat(t *testing.T) {
	r, clock := newTestRegistry(t)
	if err := r.Register("uuid-hb", "SPEC-A", "run"); err != nil {
		t.Fatal(err)
	}
	original := mustRead(t, r)[0]

	clock.Current = clock.Current.Add(15 * time.Minute)
	if err := r.Heartbeat("uuid-hb"); err != nil {
		t.Fatal(err)
	}

	after := mustRead(t, r)[0]
	if !after.StartedAt.Equal(original.StartedAt) {
		t.Errorf("StartedAt should be preserved: got %v, want %v", after.StartedAt, original.StartedAt)
	}
	if after.LastHeartbeat.Equal(original.LastHeartbeat) {
		t.Errorf("LastHeartbeat should be updated, but unchanged: %v", after.LastHeartbeat)
	}
	if !after.LastHeartbeat.Equal(clock.Current) {
		t.Errorf("LastHeartbeat: got %v, want %v", after.LastHeartbeat, clock.Current)
	}
	if after.SpecID != original.SpecID || after.Phase != original.Phase {
		t.Errorf("non-heartbeat fields mutated: spec=%q phase=%q", after.SpecID, after.Phase)
	}
}

// TestHeartbeatIdempotentOnMissing verifies Heartbeat on a missing
// session_id returns nil (idempotent) and does not create a phantom entry.
func TestHeartbeatIdempotentOnMissing(t *testing.T) {
	r, _ := newTestRegistry(t)
	if err := r.Heartbeat("ghost"); err != nil {
		t.Errorf("Heartbeat on missing: want nil, got %v", err)
	}
	entries := mustRead(t, r)
	if len(entries) != 0 {
		t.Errorf("phantom entry created: %d entries", len(entries))
	}
}

// TestDeregisterSession verifies removal + idempotent re-invocation
// (REQ-COORD-005, AC-COORD-003).
func TestDeregisterSession(t *testing.T) {
	r, _ := newTestRegistry(t)
	if err := r.Register("uuid-dereg", "SPEC-A", "run"); err != nil {
		t.Fatal(err)
	}
	if err := r.Deregister("uuid-dereg"); err != nil {
		t.Fatalf("first Deregister: %v", err)
	}
	if got := mustRead(t, r); len(got) != 0 {
		t.Errorf("after Deregister: want 0 entries, got %d", len(got))
	}
}

// TestDeregisterSessionIdempotent ensures the second call on the same
// sessionID is a no-op (AC-COORD-003).
func TestDeregisterSessionIdempotent(t *testing.T) {
	r, _ := newTestRegistry(t)
	if err := r.Register("uuid-x", "SPEC-A", "run"); err != nil {
		t.Fatal(err)
	}
	if err := r.Deregister("uuid-x"); err != nil {
		t.Fatal(err)
	}
	if err := r.Deregister("uuid-x"); err != nil {
		t.Errorf("second Deregister on missing: want nil, got %v", err)
	}
}

// TestQueryActiveWorkFilter verifies the optSpecID filter behavior
// (REQ-COORD-006, AC-COORD-014).
func TestQueryActiveWorkFilter(t *testing.T) {
	r, _ := newTestRegistry(t)
	for _, spec := range []string{"SPEC-A", "SPEC-B", "SPEC-A"} {
		id := "uuid-" + spec + "-" + fmt.Sprintf("%p", &spec)
		if err := r.Register(id, spec, "run"); err != nil {
			t.Fatalf("Register %s: %v", id, err)
		}
	}

	cases := []struct {
		name   string
		filter string
		want   int
	}{
		{"empty filter returns all", "", 3},
		{"SPEC-A returns 2", "SPEC-A", 2},
		{"SPEC-B returns 1", "SPEC-B", 1},
		{"non-existent returns 0", "SPEC-Z", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := r.Query(tc.filter)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != tc.want {
				t.Errorf("Query(%q): got %d entries, want %d", tc.filter, len(got), tc.want)
			}
			for _, e := range got {
				if tc.filter != "" && e.SpecID != tc.filter {
					t.Errorf("filter leak: entry spec=%q does not match %q", e.SpecID, tc.filter)
				}
			}
		})
	}
}

// TestPurgeStale exercises the 30-minute cutoff. Per AC-COORD-004.
// Uses FakeClock to deterministically advance time.
func TestPurgeStale(t *testing.T) {
	r, clock := newTestRegistry(t)

	// Register 3 entries with distinct started_at + heartbeat times.
	if err := r.Register("uuid-fresh", "SPEC-A", "plan"); err != nil {
		t.Fatal(err)
	}

	// Entry 2: register 25 minutes earlier, heartbeat 25 min old.
	clock.Current = clock.Current.Add(-25 * time.Minute)
	if err := r.Register("uuid-25min", "SPEC-B", "plan"); err != nil {
		t.Fatal(err)
	}

	// Entry 3: register 35 minutes earlier.
	clock.Current = clock.Current.Add(-10 * time.Minute) // now -35 from origin
	if err := r.Register("uuid-35min", "SPEC-C", "plan"); err != nil {
		t.Fatal(err)
	}

	// Reset clock to "now" (origin) for PurgeStale evaluation.
	clock.Current = time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)

	purged, err := r.Purge(30)
	if err != nil {
		t.Fatal(err)
	}
	if purged != 1 {
		t.Errorf("purged count: got %d, want 1", purged)
	}

	entries := mustRead(t, r)
	if len(entries) != 2 {
		t.Errorf("post-purge entry count: got %d, want 2", len(entries))
	}
	for _, e := range entries {
		if e.SessionID == "uuid-35min" {
			t.Errorf("uuid-35min should have been purged")
		}
	}
}

// TestPurgeStaleDefaultThreshold verifies that thresholdMinutes<=0 falls
// back to DefaultStaleMinutes.
func TestPurgeStaleDefaultThreshold(t *testing.T) {
	r, clock := newTestRegistry(t)
	if err := r.Register("uuid-old", "SPEC-A", "plan"); err != nil {
		t.Fatal(err)
	}
	// Advance clock by 31 min; entry should now be stale per default 30.
	clock.Current = clock.Current.Add(31 * time.Minute)
	purged, err := r.Purge(0)
	if err != nil {
		t.Fatal(err)
	}
	if purged != 1 {
		t.Errorf("with threshold=0 (default 30): got %d, want 1", purged)
	}
}

// TestRegisterSessionConcurrent stresses Register under -race with 10
// goroutines doing 100 registrations each. After all goroutines complete,
// the registry must contain exactly 1000 unique entries.
//
// AC-COORD-001 pass criterion: 0 race detector warnings + final count = 1000.
func TestRegisterSessionConcurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent stress test in -short mode")
	}
	r, _ := newTestRegistry(t)
	// Stress test holds the lock 1000 times in rapid succession; the
	// production 2s timeout is insufficient for the 10-goroutine pile-up
	// under -race. Override with 60s for this test only.
	r = r.WithLockTimeout(60 * time.Second)

	const (
		workers   = 10
		perWorker = 100
	)
	var wg sync.WaitGroup
	errCh := make(chan error, workers*perWorker)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				id := fmt.Sprintf("uuid-%d-%d", workerID, i)
				if err := r.Register(id, "SPEC-STRESS", "run"); err != nil {
					errCh <- err
					return
				}
			}
		}(w)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("concurrent Register: %v", err)
	}

	entries := mustRead(t, r)
	if len(entries) != workers*perWorker {
		t.Errorf("final entry count: got %d, want %d", len(entries), workers*perWorker)
	}
	// Verify all sessionIDs are unique (no collision-induced merging).
	seen := make(map[string]struct{})
	for _, e := range entries {
		if _, dup := seen[e.SessionID]; dup {
			t.Errorf("duplicate sessionID in final state: %s", e.SessionID)
		}
		seen[e.SessionID] = struct{}{}
	}
}

// TestQueryEmptyRegistry checks that Query on a missing/empty registry
// returns an empty slice (not nil, not error). REQ-COORD-018 backward
// compatibility depends on this behavior.
func TestQueryEmptyRegistry(t *testing.T) {
	r, _ := newTestRegistry(t)
	got, err := r.Query("")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Error("Query on empty registry: got nil slice, want empty slice")
	}
	if len(got) != 0 {
		t.Errorf("Query on empty registry: got %d entries, want 0", len(got))
	}
}

// TestRegistryAtomicWrite verifies that the registry file is never
// corrupt mid-write: read interleavings always observe valid JSON.
func TestRegistryAtomicWrite(t *testing.T) {
	r, _ := newTestRegistry(t)
	// Pre-populate with some entries.
	for i := 0; i < 5; i++ {
		if err := r.Register(fmt.Sprintf("uuid-pre-%d", i), "SPEC-A", "plan"); err != nil {
			t.Fatal(err)
		}
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 50; i++ {
			data, err := os.ReadFile(r.path)
			if err != nil {
				continue // file may transiently not exist; not an error here
			}
			if len(data) == 0 {
				continue
			}
			var entries []Entry
			if err := json.Unmarshal(data, &entries); err != nil {
				t.Errorf("registry file became invalid JSON during concurrent write: %v", err)
				return
			}
		}
	}()

	for i := 5; i < 30; i++ {
		if err := r.Register(fmt.Sprintf("uuid-stress-%d", i), "SPEC-A", "plan"); err != nil {
			t.Fatal(err)
		}
	}
	<-done
}

// TestFormatStderrReminder verifies the system-reminder format used by
// SessionStart hook Step 3 (REQ-COORD-015).
func TestFormatStderrReminder(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	entries := []Entry{
		{
			SessionID:     "current-uuid-1234",
			SpecID:        "SPEC-CURRENT",
			Phase:         "plan",
			LastHeartbeat: now,
			PID:           111,
			CWD:           "/tmp/cur",
		},
		{
			SessionID:     "other-uuid-5678",
			SpecID:        "SPEC-OTHER",
			Phase:         "run",
			LastHeartbeat: now.Add(-15 * time.Minute),
			PID:           222,
			CWD:           "/tmp/other",
		},
	}

	got := FormatStderrReminder("current-uuid-1234", entries, now)
	if got == "" {
		t.Fatal("FormatStderrReminder: want non-empty output, got empty")
	}
	for _, want := range []string{"<system-reminder>", "other-uu", "SPEC-OTHER", "phase=run", "age=15m"} {
		if !registryContains(got, want) {
			t.Errorf("FormatStderrReminder missing %q. Got: %s", want, got)
		}
	}
	// When no other sessions exist, output should be empty.
	emptyOut := FormatStderrReminder("current-uuid-1234", entries[:1], now)
	if emptyOut != "" {
		t.Errorf("FormatStderrReminder with no others: want empty, got %q", emptyOut)
	}
}

// TestRegisterPersistsAtomically verifies that after Register, a second
// fresh Registry instance bound to the same path sees the entries.
// This indirectly verifies the temp + rename atomic-write path.
func TestRegisterPersistsAtomically(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	r1 := NewRegistry(path, realClock{})
	if err := r1.Register("uuid-persist", "SPEC-X", "plan"); err != nil {
		t.Fatal(err)
	}

	r2 := NewRegistry(path, realClock{})
	entries, err := r2.Query("")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].SessionID != "uuid-persist" {
		t.Errorf("r2 did not see r1's write: %+v", entries)
	}
}

// registryContains is a tiny strings.Contains shim local to the registry
// test file (avoid colliding with the existing contains helper in
// store_test.go which serves the SPEC-V3R2-RT-004 checkpoint suite).
func registryContains(haystack, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	if len(needle) > len(haystack) {
		return false
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// TestDetectHostNonEmpty ensures detectHost falls back gracefully and
// never returns an empty string (the registry uses this for the Host field).
func TestDetectHostNonEmpty(t *testing.T) {
	if got := detectHost(); got == "" {
		t.Error("detectHost: want non-empty, got empty")
	}
}

// TestPlatformTagFormat is a compile-only smoke test ensuring platformTag
// returns a non-empty "GOOS/GOARCH" string under the test build.
func TestPlatformTagFormat(t *testing.T) {
	tag := platformTag()
	if tag == "" || !registryContains(tag, "/") {
		t.Errorf("platformTag: want non-empty GOOS/GOARCH, got %q", tag)
	}
}

// withPackageRegistry temporarily chdir's into a fresh temp directory so
// that the package-level RegisterSession / Heartbeat / Query / Purge /
// Deregister helpers (which bind to DefaultRegistryPath relative to CWD)
// resolve to an isolated path. Restores CWD via t.Cleanup.
func withPackageRegistry(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
	return dir
}

// TestPackageLevelHelpers exercises the package-level entry points
// (RegisterSession, Heartbeat, DeregisterSession, QueryActiveWork,
// PurgeStale) which delegate to defaultRegistry() bound to
// DefaultRegistryPath. Covers the package-level wrappers explicitly.
func TestPackageLevelHelpers(t *testing.T) {
	withPackageRegistry(t)

	if err := RegisterSession("uuid-pkg-1", "SPEC-PKG", "plan"); err != nil {
		t.Fatalf("RegisterSession: %v", err)
	}
	if err := Heartbeat("uuid-pkg-1"); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	entries, err := QueryActiveWork("SPEC-PKG")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("QueryActiveWork(SPEC-PKG): got %d, want 1", len(entries))
	}
	purged, err := PurgeStale(0) // default threshold
	if err != nil {
		t.Fatal(err)
	}
	if purged != 0 {
		t.Errorf("PurgeStale on fresh entry: got %d, want 0", purged)
	}
	if err := DeregisterSession("uuid-pkg-1"); err != nil {
		t.Fatalf("DeregisterSession: %v", err)
	}
	entries, err = QueryActiveWork("")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("after Deregister: got %d entries, want 0", len(entries))
	}
}

// TestPackageLevelEmptySessionID verifies the empty-sessionID validation
// path exposed via the package-level helpers.
func TestPackageLevelEmptySessionID(t *testing.T) {
	withPackageRegistry(t)
	if err := RegisterSession("", "SPEC-X", "plan"); err == nil {
		t.Error("RegisterSession(\"\"): want error, got nil")
	}
	if err := Heartbeat(""); err == nil {
		t.Error("Heartbeat(\"\"): want error, got nil")
	}
	if err := DeregisterSession(""); err == nil {
		t.Error("DeregisterSession(\"\"): want error, got nil")
	}
}

// TestNewRegistryNilClock confirms passing a nil Clock falls back to realClock.
func TestNewRegistryNilClock(t *testing.T) {
	r := NewRegistry("/tmp/never-used.json", nil)
	if r.clock == nil {
		t.Error("NewRegistry with nil clock should default to realClock")
	}
	now := r.clock.Now()
	if now.IsZero() {
		t.Error("default clock returned zero time")
	}
}

// TestReadAllUnlockedCorrupt simulates a corrupt registry file (invalid
// JSON) and verifies readAllUnlocked surfaces a parse error.
func TestReadAllUnlockedCorrupt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "corrupt.json")
	if err := os.WriteFile(path, []byte("{ this is not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := NewRegistry(path, realClock{})
	if _, err := r.readAllUnlocked(); err == nil {
		t.Error("readAllUnlocked on corrupt JSON: want error, got nil")
	}
}

// TestShortIDEdgeCases covers the truncation logic for short and long IDs.
func TestShortIDEdgeCases(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"abc", "abc"},
		{"12345678", "12345678"},
		{"123456789", "12345678"},
		{"deadbeef-cafe-1234-5678-9012", "deadbeef"},
	}
	for _, tc := range cases {
		if got := shortID(tc.in); got != tc.want {
			t.Errorf("shortID(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
