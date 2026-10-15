package userassets

// suspension_cleanup_test.go — review finding 3 (card t1591): the installer
// goroutine parked by the suspension seam must end before the test's temporary
// directory is removed, on every exit path — including a Fatal that fires
// before the test releases the run (AGENTS.md §4: a cleanup-guaranteed
// process, not a trailing kill).
//
// The failure path is driven through recordingTB: the seam registers its
// cleanup there instead of on the real *testing.T, so this test can run the
// registered cleanups itself, in the framework's last-in-first-out order,
// without failing the real test.

import (
	"fmt"
	"testing"
)

// recordingTB captures the cleanups and error reports a suspension makes, so a
// test can exercise the cleanup path without failing the real test.
type recordingTB struct {
	testing.TB
	cleanups []func()
	errs     []string
}

func (r *recordingTB) Helper()           {}
func (r *recordingTB) Cleanup(fn func()) { r.cleanups = append(r.cleanups, fn) }
func (r *recordingTB) Errorf(format string, args ...any) {
	r.errs = append(r.errs, fmt.Sprintf(format, args...))
}

// TestSuspensionEndsInstallerOnFailurePath — the run is parked and the test
// fails before releasing it. The registered cleanups, run last-in-first-out as
// the framework runs them before it removes the temporary directory, must end
// the installer goroutine.
func TestSuspensionEndsInstallerOnFailurePath(t *testing.T) {
	f := newFixture(t)
	rec := &recordingTB{TB: t}
	susp := newSuspension(rec, "claude-agents/manager-x.md")
	inst := f.installer(t)
	susp.wire(inst)
	susp.startInstall(inst)
	waitChannel(t, susp.parked, "mid-loop park")

	// The failure path: the test ends without releasing the parked run.
	for i := len(rec.cleanups) - 1; i >= 0; i-- {
		rec.cleanups[i]()
	}
	if len(rec.errs) != 0 {
		t.Fatalf("the suspension cleanup reported: %v", rec.errs)
	}
	select {
	case <-susp.installerDone:
	default:
		t.Fatal("the installer goroutine was still parked after the cleanups ran — it would outlive the temporary directory")
	}
}

// TestSuspensionReleaseIsIdempotent — the test's own release and the cleanup
// both release the run; the second release must not panic on a closed channel.
func TestSuspensionReleaseIsIdempotent(t *testing.T) {
	susp := newSuspension(t, "claude-agents/manager-x.md")
	susp.releaseRun()
	susp.releaseRun()
	select {
	case <-susp.release:
	default:
		t.Fatal("the release channel is not closed after releaseRun")
	}
}
