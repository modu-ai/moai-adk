//go:build unix

package hook

// instructions_loaded_fifo_unix_test.go — the C1 named-pipe fixture (card
// t2117): syscall.Mkfifo exists on unix only. A writer-less FIFO sitting
// where the instruction-file anchor is expected blocks os.ReadFile forever;
// the set derivation must skip it as non-regular instead of hanging.

import (
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
	"time"
)

// TestInstructionFileSetSkipsFIFOAnchor (card t2117 C1 RED): the collection
// runs under a timeout guard so the current unguarded read FAILS the test by
// timing out (the defect observed) rather than hanging the suite; the fixed
// code must skip the FIFO and return.
func TestInstructionFileSetSkipsFIFOAnchor(t *testing.T) {
	root := t.TempDir()
	if err := syscall.Mkfifo(filepath.Join(root, "AGENTS.md"), 0o600); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}
	rules := filepath.Join(root, ".claude", "rules", "moai")
	if err := os.MkdirAll(rules, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rules, "always.md"), []byte("# rule\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	type result struct{ set []string }
	done := make(chan result, 1)
	go func() {
		done <- result{set: instructionFileSet(root)}
	}()
	select {
	case got := <-done:
		want := []string{filepath.Join(rules, "always.md")}
		if !reflect.DeepEqual(got.set, want) {
			t.Errorf("instructionFileSet() = %v, want %v (the FIFO anchor must be skipped as non-regular)", got.set, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("instructionFileSet hung on the FIFO anchor — the read has no regular-file guard (card t2117 C1)")
	}
}
