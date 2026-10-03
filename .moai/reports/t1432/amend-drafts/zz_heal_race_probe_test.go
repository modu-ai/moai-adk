//go:build !windows

package harness

// DRAFT measurement instrument (t1432 amendment 0.4.0), not part of the change and not in the tree.
// A variant of the audit's removal-window probe that goes through the production heal entry point
// healStateEntry instead of the bare removal helper, and whose swapper behaves as a healer that HONOURS
// the heal lock: it takes an exclusive flock on <log>.prune-heal (blocking) around its rename. At the
// tree 7639c04c1 no heal lock exists in the production code, so the swapper's lock changes nothing and
// the window stays reachable; after the heal-lock implementation the same probe is expected to count 0.
//
// Per trial: a symbolic link owned by the current user sits at the state path and is the inspected
// entry; one swapper renames a fresh file F over the path (under the heal lock) while the production
// healStateEntry runs once. A violation is a trial after which the state path is absent: the heal
// removed the swapped-in F.

import (
	"math/rand"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestZZProbeHealWindow(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	statePath := logPath + stampSuffix
	healPath := logPath + ".prune-heal"
	victim := filepath.Join(dir, "victim")
	if err := os.WriteFile(victim, []byte("V"), 0o600); err != nil {
		t.Fatal(err)
	}
	tmpF := filepath.Join(dir, "tmpF")
	ret := NewRetention(logPath, filepath.Join(dir, "archive"), nil)
	rng := rand.New(rand.NewSource(1))
	// Amendment 0.4.1: the trial floor, not the clock, is the stop condition. The loop ends at the 5th
	// violation (the pre-fix control) or at 20000 trials (the post-fix measurement of Definition of Done
	// 10); the 240 s deadline is a safety stop only, and a run that ends on it with fewer than 20000
	// trials is a Gap (run with -timeout 300s).
	deadline := time.Now().Add(240 * time.Second)
	trials, violations, helperFirst, swapFirst, healErrors := 0, 0, 0, 0, 0
	for time.Now().Before(deadline) && violations < 5 && trials < 20000 {
		trials++
		_ = os.Remove(statePath)
		_ = os.Remove(tmpF)
		if err := os.Symlink(victim, statePath); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(tmpF, []byte("F"), 0o600); err != nil {
			t.Fatal(err)
		}
		inspected, err := os.Lstat(statePath)
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
			hf, herr := os.OpenFile(healPath, os.O_RDWR|os.O_CREATE, 0o600)
			if herr == nil {
				_ = syscall.Flock(int(hf.Fd()), syscall.LOCK_EX)
			}
			_ = os.Rename(tmpF, statePath)
			if hf != nil {
				_ = syscall.Flock(int(hf.Fd()), syscall.LOCK_UN)
				_ = hf.Close()
			}
		}()
		go1.Store(true)
		herr := ret.healStateEntry(statePath, inspected, "symbolic link")
		<-done
		if herr != nil {
			healErrors++
		}
		fi, perr := os.Lstat(statePath)
		switch {
		case perr != nil:
			violations++
		case fi.Mode().IsRegular():
			swapFirst++ // F is in place: the swap landed after the heal's removal, or the heal declined to remove it
		default:
			helperFirst++
		}
	}
	t.Logf("PROBE heal-window: trials=%d heal_removed_the_swapped_in_fresh_entry=%d path_holds_F=%d path_holds_other=%d heal_errors=%d", trials, violations, swapFirst, helperFirst, healErrors)
}
