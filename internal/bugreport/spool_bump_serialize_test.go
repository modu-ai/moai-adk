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
