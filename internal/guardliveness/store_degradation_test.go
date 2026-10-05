package guardliveness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/paths"
)

// REQ-GDL-008 — the default persistence lives under the user's ~/.moai state
// tree, which is outside every evaluated working tree. Asserted against a
// throwaway root rather than the operator's real one.
func TestDefaultStoreResolvesUnderTheMoaiStateTree(t *testing.T) {
	home := t.TempDir()
	t.Setenv(paths.EnvHome, home)

	store, err := DefaultStore()
	if err != nil {
		t.Fatalf("DefaultStore: %v", err)
	}
	root := t.TempDir()
	if err := store.Save(root, resultA(), time.Now()); err != nil {
		t.Fatalf("Save: %v", err)
	}

	written, err := filepath.Glob(filepath.Join(home, "state", storeSubdir, "*.json"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("found %d persisted file(s) under the state tree, want 1", len(written))
	}
	if _, err := store.Load(root); err != nil {
		t.Fatalf("Load from the default store: %v", err)
	}
}

// A persistence directory that cannot be created is a degradation, not a
// crash: the refresh still ran and only its carriage to the next activation
// was lost.
func TestSaveReportsAnUnusableDirectory(t *testing.T) {
	blocked := filepath.Join(t.TempDir(), "occupied")
	if err := os.WriteFile(blocked, []byte("not a directory\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	err := NewStore(blocked).Save(t.TempDir(), resultA(), time.Now())
	if err == nil {
		t.Fatal("Save reported success against a path that is not a directory")
	}
	if !strings.Contains(err.Error(), "guard liveness") {
		t.Errorf("error does not identify its origin: %v", err)
	}
}

// A persisted file that cannot be decoded must be reported as a read failure,
// never as an absent verdict and never as an empty result: an empty result
// partitions to nothing non-clean, which is an all-clear about a set nothing
// read.
func TestLoadReportsAMalformedPersistedResult(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	root := t.TempDir()
	if err := store.Save(root, resultA(), time.Now()); err != nil {
		t.Fatalf("Save: %v", err)
	}

	written, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil || len(written) != 1 {
		t.Fatalf("locate the persisted file: %v (%d found)", err, len(written))
	}
	if err := os.WriteFile(written[0], []byte("{ this is not json"), 0o644); err != nil {
		t.Fatalf("corrupt the persisted file: %v", err)
	}

	snap, err := store.Load(root)
	if err == nil {
		t.Fatal("Load reported success on a malformed file")
	}
	if !snap.TakenAt.IsZero() || snap.Result.Entries != nil {
		t.Errorf("Load returned a partial snapshot alongside its error: %+v", snap)
	}
}

// The age is what makes a stale advisory legible, so its rendering is asserted
// across every resolution it switches on — including the backwards-clock case,
// which must read as zero rather than as a measurement from the future.
func TestFormatAgeResolution(t *testing.T) {
	for _, tc := range []struct {
		in   time.Duration
		want string
	}{
		{-time.Minute, "0ms"},
		{0, "0ms"},
		{250 * time.Millisecond, "250ms"},
		{45 * time.Second, "45s"},
		{5 * time.Minute, "5m"},
		{59 * time.Minute, "59m"},
		{90 * time.Minute, "1h30m"},
		{26 * time.Hour, "26h00m"},
	} {
		if got := formatAge(tc.in); got != tc.want {
			t.Errorf("formatAge(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// An evaluator built without a producer degrades to the unwired one rather
// than panicking at session start, and still reaches the query layer.
func TestNewWithoutAProducerDegradesToUnwired(t *testing.T) {
	if _, err := New(nil).Refresh(t.Context(), Activation{Root: t.TempDir()}); err == nil {
		t.Fatal("an evaluator with no producer reported a successful refresh")
	}
}

// A refresh with no sink still runs. Persistence is how the verdict reaches the
// next activation; its absence must not stop the verdict being produced.
func TestRunWithoutASinkStillRefreshes(t *testing.T) {
	p := &countingProducer{}
	if _, err := New(p).Run(t.Context(), Activation{Root: t.TempDir()}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := p.count(); got != 1 {
		t.Fatalf("query-layer arrivals = %d, want 1", got)
	}
}
