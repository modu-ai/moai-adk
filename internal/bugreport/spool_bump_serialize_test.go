package bugreport

// The concurrent-bump serialization test (review gate finding, P1):
// BumpSpoolGeneration's read-increment-write was not atomic — two
// concurrent purges read the SAME counter and wrote the SAME next number,
// so N purges advanced the generation by less than N. A sender that read
// its baseline between the first purge and the last one then compared
// equal across the second purge's withdrawal and published over it. The
// bumps are serialized under the marker's own section lock: N concurrent
// bumps leave the counter at exactly N.

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestConcurrentBumpsSerializeTheGeneration(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv("CI", "")

	const purges = 32
	path, err := SpoolGenerationPath()
	if err != nil {
		t.Fatalf("generation path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create store dir: %v", err)
	}

	start := make(chan struct{})
	var wg sync.WaitGroup
	for range purges {
		wg.Go(func() {
			<-start
			if berr := BumpSpoolGeneration(); berr != nil {
				t.Errorf("bump: %v", berr)
			}
		})
	}
	close(start)
	wg.Wait()

	gen, gerr := SpoolGeneration()
	if gerr != nil {
		t.Fatalf("read generation: %v", gerr)
	}
	if gen != purges {
		t.Fatalf("generation = %d after %d concurrent purges, want exactly %d — two purges read the same counter and one withdrawal became invisible", gen, purges, purges)
	}
}

// TestBumpPropagatesAMarkerReadFailure pins the residual lost-update path
// (card t1606): a marker read that fails must propagate, never rewind the
// counter to 1. A silent reset makes every generation a prior withdrawal
// had retired publishable again — the exact hazard the bump exists to
// close — while a failed bump is retried by the purge.
func TestBumpPropagatesAMarkerReadFailure(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv("CI", "")

	path, err := SpoolGenerationPath()
	if err != nil {
		t.Fatalf("generation path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create store dir: %v", err)
	}
	if werr := os.WriteFile(path, []byte("7\n"), 0o600); werr != nil {
		t.Fatalf("seed marker: %v", werr)
	}

	prev := spoolGenerationFn
	t.Cleanup(func() { spoolGenerationFn = prev })
	readErr := errors.New("time box exceeded")
	spoolGenerationFn = func() (uint64, error) { return 0, readErr }

	if berr := BumpSpoolGeneration(); !errors.Is(berr, readErr) {
		t.Fatalf("bump with a failed marker read: error = %v, want the read error propagated — a silent reset would rewind the counter", berr)
	}
	// The marker is untouched: the withdrawal did not half-happen.
	gen, gerr := SpoolGeneration()
	if gerr != nil {
		t.Fatalf("read back marker: %v", gerr)
	}
	if gen != 7 {
		t.Fatalf("marker = %d after a failed bump, want 7 — the bump must not write past a failed read", gen)
	}
}
