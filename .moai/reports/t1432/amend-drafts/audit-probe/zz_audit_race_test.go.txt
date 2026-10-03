//go:build !windows

package harness

import (
	"math/rand"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// Per trial: A is at the state path and is the inspected entry; one swap renames a fresh file F over
// the path while the production helper removeStateEntryIfUnchanged runs once. Outcomes:
//   swap first      -> the helper's re-check sees F, removed=false, F present
//   helper first    -> the helper unlinks A, the swap then creates F: removed=true, F present
//   swap between the helper's re-check and its unlink -> the helper unlinks F: removed=true, path absent
func TestZZProbeRemoveWindow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state")
	tmpF := filepath.Join(dir, "tmpF")
	rng := rand.New(rand.NewSource(1))
	deadline := time.Now().Add(90 * time.Second)
	trials, violations, helperFirst, swapFirst := 0, 0, 0, 0
	for time.Now().Before(deadline) && violations < 5 && trials < 200000 {
		trials++
		_ = os.Remove(path)
		_ = os.Remove(tmpF)
		if err := os.WriteFile(path, []byte("A"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(tmpF, []byte("F"), 0o600); err != nil {
			t.Fatal(err)
		}
		inspected, err := os.Lstat(path)
		if err != nil {
			t.Fatal(err)
		}
		spin := rng.Intn(4000)
		var go1 atomic.Bool
		done := make(chan struct{})
		go func() {
			defer close(done)
			for !go1.Load() {
			}
			x := 0
			for i := 0; i < spin; i++ {
				x += i
			}
			_ = x
			_ = os.Rename(tmpF, path)
		}()
		go1.Store(true)
		removed, _ := removeStateEntryIfUnchanged(path, inspected)
		<-done
		_, perr := os.Lstat(path)
		switch {
		case removed && perr != nil:
			violations++
		case removed && perr == nil:
			helperFirst++
		case !removed:
			swapFirst++
		}
	}
	t.Logf("PROBE remove-window: trials=%d helper_removed_the_swapped_in_fresh_entry=%d helper_first=%d swap_first=%d", trials, violations, helperFirst, swapFirst)
}
