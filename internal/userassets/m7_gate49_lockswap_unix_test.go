//go:build unix

package userassets

// m7_gate49_lockswap_unix_test.go — gate round 49 [12th raise]: the
// lock-family record read is HANDLE-BOUND (readLockRecord: non-blocking
// open, type check on the open handle, read from the same handle). A
// SUBPROCESS swapping the lock file with atomic renames while the parent
// classifies must surface only WHOLE records — never a torn pid line the
// path-based Lstat→ReadFile pair could read mid-swap. The subprocess is
// this test binary re-entered with the child env set (no shell in the
// middle).

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"
)

func TestLockSwapSubprocess(t *testing.T) {
	if os.Getenv("UA_LOCK_SWAP_CHILD") == "1" {
		lockSwapChild(t)
		return
	}

	dir := t.TempDir()
	lockPath := filepath.Join(dir, "user-assets.lock")
	deadPID := deadProcessPID(t)

	cmd := exec.Command(os.Args[0], "-test.run=TestLockSwapSubprocess$", "-test.v")
	cmd.Env = append(os.Environ(),
		"UA_LOCK_SWAP_CHILD=1",
		"UA_LOCK_SWAP_PATH="+lockPath,
		"UA_LOCK_SWAP_DEAD="+strconv.Itoa(deadPID),
		// the LIVE record names the PARENT — a pid the parent's allow-list
		// knows and one that is provably alive for the whole run
		"UA_LOCK_SWAP_LIVE="+strconv.Itoa(os.Getpid()),
	)
	if err := cmd.Start(); err != nil {
		t.Fatalf("spawn the swap subprocess: %v", err)
	}

	// Classify tightly across the swap churn: every verdict must name one
	// of the two written pids (a WHOLE record) — Ownerless from a
	// pid-bearing record is a torn read, the defect class the handle-bound
	// reader exists to close. The run also requires at least ONE whole
	// record observed — a sweep that only saw the rename window's
	// transient states would assert nothing.
	livePID := os.Getpid()
	sawWholeRecord := false
	for i := 0; i < 800 && !sawWholeRecord; i++ {
		state, pid := ClassifyLockFile(lockPath)
		switch state {
		case GuardMarkerOwnerAlive, GuardMarkerOwnerDead:
			sawWholeRecord = true
			if pid != deadPID && pid != livePID {
				_ = cmd.Process.Kill()
				_ = cmd.Wait()
				t.Fatalf("classification %d carried pid %d — a record the subprocess never wrote (torn read)", i, pid)
			}
		case GuardMarkerAbsent, GuardMarkerIrregular:
			// the rename window's legitimate states; a micro-yield keeps
			// the bound spanning the child's spawn + first writes
			time.Sleep(200 * time.Microsecond)
		default:
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			t.Fatalf("classification %d read Ownerless from a pid-bearing record — torn read (gate 49 readLockRecord)", i)
		}
	}
	if !sawWholeRecord {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatal("800 classifications never saw a whole record — the swap churn was not exercised (a vacuous pass)")
	}
	// The churn continues past the first whole record: keep judging until
	// the child exits or the bound trips, so the swap crossings are real.
	for i := 0; i < 800; i++ {
		state, pid := ClassifyLockFile(lockPath)
		switch state {
		case GuardMarkerOwnerAlive, GuardMarkerOwnerDead:
			if pid != deadPID && pid != livePID {
				_ = cmd.Process.Kill()
				_ = cmd.Wait()
				t.Fatalf("classification %d carried pid %d — a record the subprocess never wrote (torn read)", i, pid)
			}
		case GuardMarkerAbsent, GuardMarkerIrregular:
		default:
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			t.Fatalf("classification %d read Ownerless from a pid-bearing record — torn read (gate 49 readLockRecord)", i)
		}
	}

	_ = cmd.Process.Kill()
	_ = cmd.Wait()
}

// TestUnixPidlessMarkerFIFOOpenNonBlocking — gate round 51-4: the SECOND
// open (classifyPidlessMarker's flock probe) carries readLockRecord's
// discipline — O_NONBLOCK + the type check bound to THIS handle. A FIFO
// swap before the open surfaces fail-closed (held) instead of parking the
// classify on the writer wait. The old plain os.Open hung here.
func TestUnixPidlessMarkerFIFOOpenNonBlocking(t *testing.T) {
	dir := t.TempDir()
	fifo := filepath.Join(dir, "user-assets.acquire-guard.guard")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}

	done := make(chan GuardMarkerState, 1)
	go func() {
		state, _ := classifyPidlessMarker(fifo)
		done <- state
	}()
	select {
	case state := <-done:
		if state != GuardMarkerOwnerAlive {
			t.Fatalf("a FIFO at the marker path classified %s, want the fail-closed owner-alive (gate 51-4)", state)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("classifyPidlessMarker parked on the FIFO open — the second open lost the O_NONBLOCK discipline (gate 51-4)")
	}
}

// lockSwapChild runs inside the re-entered test binary: it rewrites the
// lock path with atomic renames, alternating a dead-pid and a live-pid
// record. The live pid is the PARENT's (passed through the env) — a pid
// the parent's verdict allow-list knows.
func lockSwapChild(t *testing.T) {
	lockPath := os.Getenv("UA_LOCK_SWAP_PATH")
	deadPID, err := strconv.Atoi(os.Getenv("UA_LOCK_SWAP_DEAD"))
	if err != nil {
		t.Fatal(err)
	}
	livePID, err := strconv.Atoi(os.Getenv("UA_LOCK_SWAP_LIVE"))
	if err != nil {
		t.Fatal(err)
	}
	record := func(pid int) []byte {
		return []byte("pid=" + itoaTest(pid) + " token=swap acquired=2026-10-09T00:00:00Z\n")
	}
	for i := 0; i < 400; i++ {
		if err := os.WriteFile(lockPath+".tmp", record(deadPID), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(lockPath+".tmp", lockPath); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(lockPath+".tmp", record(livePID), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(lockPath+".tmp", lockPath); err != nil {
			t.Fatal(err)
		}
	}
}
