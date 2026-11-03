//go:build !windows

package cli

// codex_close_descendant_repro_test.go — a minimal, network-free reproduction of
// the ordering behind "TempDir RemoveAll cleanup: ... directory not empty" in
// TestCodexLive_ReviewStartBaseBranchIsNotRejected (Issue #1948).
//
// The live test closes its codex session and returns; t.TempDir then removes
// CODEX_HOME. The real codex app-server writes .codex/.tmp/plugins-clone-* in
// the background. realCodexConn.close closes stdin, waits at most 3s, and kills
// only the DIRECT child; it never waits for descendants. This file asks the one
// question that ordering turns on, with a fake child in place of codex:
//
//	when close returns, is anything still writing into the session directory?
//
// The fake is a shell script. Its background writer detaches from the pipes (as
// a git clone run by codex would) and ends on its own: a stop file, or 400
// iterations, whichever comes first. No real codex, no network, no kill by name.
// It was RED (behind the afred tag) until close waited for the session's
// process group; observed before and after the repair in Issue #1948, it now runs
// in the default suite. The shell fake keeps it off windows.

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const reproChildScript = `#!/bin/sh
w="$2/.tmp/plugins-clone-X/plugins"
mkdir -p "$w"
if [ "$4" = "descendant" ]; then
  (
    i=0
    while [ ! -e "$3" ] && [ $i -lt 400 ]; do
      : > "$w/f$i"
      i=$((i+1))
      sleep 0.01
    done
  ) >/dev/null 2>&1 &
fi
if [ "$5" = "ignore-eof" ]; then
  while [ ! -e "$3" ]; do sleep 0.05; done
else
  read line
fi
exit 0
`

type fakeCloseRun struct {
	closeDur time.Duration
	alive    bool  // the session directory still grew after close returned
	rmErr    error // os.RemoveAll of the session directory right after that
}

func runFakeCodexClose(t *testing.T, descendant, ignoreEOF bool) fakeCloseRun {
	t.Helper()
	scriptDir := t.TempDir()
	bin := filepath.Join(scriptDir, "codex")
	if err := os.WriteFile(bin, []byte(reproChildScript), 0o755); err != nil {
		t.Fatal(err)
	}
	work, err := os.MkdirTemp("", "fakecodexhome-")
	if err != nil {
		t.Fatal(err)
	}
	stop := filepath.Join(scriptDir, "stop")
	t.Cleanup(func() {
		// Cleanup guarantee: the writer ends on the stop file, then the directory goes.
		_ = os.WriteFile(stop, nil, 0o600)
		time.Sleep(150 * time.Millisecond)
		_ = os.RemoveAll(work)
	})
	desc, eof := "none", "eof"
	if descendant {
		desc = "descendant"
	}
	if ignoreEOF {
		eof = "ignore-eof"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := (realCodexSessionRunner{}).start(ctx, bin, []string{"app-server", work, stop, desc, eof}, scriptDir)
	if err != nil {
		t.Fatalf("start the fake child: %v", err)
	}
	plugins := filepath.Join(work, ".tmp", "plugins-clone-X", "plugins")
	ready := filepath.Join(plugins, "f0")
	if !descendant {
		ready = plugins
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("the fake child never created %s", ready)
		}
		time.Sleep(5 * time.Millisecond)
	}
	count := func() int {
		es, _ := os.ReadDir(plugins)
		return len(es)
	}
	start := time.Now()
	_ = conn.close()
	r := fakeCloseRun{closeDur: time.Since(start)}
	n1 := count()
	time.Sleep(300 * time.Millisecond)
	r.alive = count() > n1
	r.rmErr = os.RemoveAll(work)
	return r
}

func TestRepro_CloseReturnsWhileADescendantStillWrites(t *testing.T) {
	cases := []struct {
		name       string
		descendant bool
		ignoreEOF  bool
		wantAlive  bool // false: close must not return while the session dir is still written
	}{
		{"control_no_descendant", false, false, false},
		{"child_exits_on_eof_descendant_left", true, false, false},
		{"child_needs_kill_descendant_left", true, true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := runFakeCodexClose(t, c.descendant, c.ignoreEOF)
			t.Logf("MEASURED close=%s stillWritingAfterClose=%v removeAllErr=%v", r.closeDur.Round(time.Millisecond), r.alive, r.rmErr)
			if r.alive != c.wantAlive {
				t.Errorf("RED: close returned while a descendant still wrote into the session directory")
			}
		})
	}
}

// The control must be clean every time, so a failure in the descendant case is
// the descendant and not the harness. Before the repair the descendant case lost
// the race in 1 of 10 rounds (probabilistic); after it, no round may fail.
func TestRepro_RemoveAllAfterCloseFailureRate(t *testing.T) {
	const rounds = 5
	fail := map[bool]int{}
	for _, descendant := range []bool{false, true} {
		for i := 0; i < rounds; i++ {
			if r := runFakeCodexClose(t, descendant, false); r.rmErr != nil {
				fail[descendant]++
				t.Logf("descendant=%v round %d: %v", descendant, i, r.rmErr)
			}
		}
	}
	t.Logf("MEASURED RemoveAll failures after close: control %d/%d, descendant %d/%d", fail[false], rounds, fail[true], rounds)
	if fail[false] != 0 {
		t.Fatalf("the control (no descendant) failed %d/%d: the harness is not clean", fail[false], rounds)
	}
	if fail[true] != 0 {
		t.Errorf("RemoveAll failed %d/%d rounds after close with a descendant left", fail[true], rounds)
	}
}
