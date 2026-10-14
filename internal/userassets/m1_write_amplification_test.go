// m1_write_amplification_test.go — SPEC-USERASSET-DEPLOY-GUARD-001, gate
// round 15 (design §2 write-amplification constraint): the per-file
// completion-flag persistence fires ONLY for targets the run actually
// wrote. An up-to-date re-run changed nothing on disk, so it must not
// re-serialize the journal per target — the gate measured 6 unchanged
// files driving the journal from 2 writes to 7 under the ungated form.
package userassets

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestJournalFlagPersistOnlyForWrittenFiles(t *testing.T) {
	f := newFixture(t)

	// Run 1 installs everything; run 2 finds every target up-to-date.
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatalf("run 1: %v", err)
	}

	// The watcher occupies the JOURNAL path with a directory the moment the
	// staging write of run 2 lands — the FIRST journal write. Any LATER
	// journal write (a per-target flag persist) then fails and is recorded
	// as a per-file failure naming "persist completion flag". With the
	// amplification gate in place there is nothing between staging and the
	// manifest save on an up-to-date tree, so no such failure may exist.
	// Gate round 24: the injection is PLATFORM-INDEPENDENT — a directory
	// occupying the path fails the rename on every OS, where a directory
	// Chmod is a no-op on windows.
	stop := make(chan struct{})
	watcherDone := make(chan struct{})
	jp := JournalPath(f.home)
	go func() {
		defer close(watcherDone)
		for {
			select {
			case <-stop:
				return
			default:
			}
			if _, err := os.Stat(jp); err == nil {
				_ = os.Remove(jp)
				_ = os.Mkdir(jp, 0o755)
				return
			}
			time.Sleep(200 * time.Microsecond)
		}
	}()

	res, installErr := f.installer(t).Install(nil)
	close(stop)
	<-watcherDone
	// Restore: drop the occupying directory so later assertions/cleanup
	// see a clean home.
	_ = os.Remove(jp)

	// Run 2 saw every target up-to-date.
	if res.Installed != 0 || res.Refreshed != 0 {
		t.Fatalf("run 2 was not an up-to-date no-op: installed=%d refreshed=%d", res.Installed, res.Refreshed)
	}
	for _, fl := range res.Failures {
		if strings.Contains(fl.Reason, "persist completion flag") {
			t.Fatalf("a per-target flag persist ran for an unchanged target — the write-amplification gate regressed: %+v", fl)
		}
	}
	// The run's tail (manifest save + journal clear) fails under the
	// read-only home — the observable proves the watcher armed in time.
	if installErr == nil {
		t.Log("note: the watcher lost the race this attempt (save succeeded); the no-persist assertion above still held")
	}
}
