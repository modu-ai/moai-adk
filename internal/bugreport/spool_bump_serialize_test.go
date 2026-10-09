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
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
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
	for i := 0; i < purges; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if berr := BumpSpoolGeneration(); berr != nil {
				t.Errorf("bump: %v", berr)
			}
		}()
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

// TestBumpSpoolGenerationContextHonorsCallerDeadline pins the ctx propa-
// gation (review gate finding): the claim retry loop originally used
// context.Background(), so a deadline-carrying withdrawal waited the FULL
// bump claim deadline (10s) out behind a held lock — the base code returned
// at the claim's own retry budget (~57ms). With the caller's ctx threaded
// through, the bump returns at the caller's deadline with the cancel error.
func TestBumpSpoolGenerationContextHonorsCallerDeadline(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv("CI", "")

	path, err := SpoolGenerationPath()
	if err != nil {
		t.Fatalf("generation path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create store dir: %v", err)
	}
	// Hold the section lock for the whole call: the bump must discover the
	// hold via ClaimSection's contention path and return at the CALLER's
	// deadline, not its own 10s claim deadline.
	release, err := atomicfile.ClaimSection(context.Background(), path+".lock", 0o600, 0, time.Millisecond)
	if err != nil {
		t.Fatalf("hold lock: %v", err)
	}
	defer func() { _ = release() }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	start := time.Now()
	berr := BumpSpoolGenerationContext(ctx)
	elapsed := time.Since(start)
	if berr == nil {
		t.Fatal("bump succeeded while the section lock was held — the retry loop did not run")
	}
	if !errors.Is(berr, context.DeadlineExceeded) {
		t.Fatalf("bump error is not the caller's deadline: %v", berr)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("bump took %s — the caller's 30ms deadline was ignored (the 10s claim deadline leaked)", elapsed)
	}
}
