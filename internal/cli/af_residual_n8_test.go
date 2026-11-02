//go:build !windows

package cli

// N8 (audit 22a0afd247 onwards, residual classification): the help of `moai af land` says "a background process the gate left
// running is killed", but the gate is killed by process group and a child that left the group (setsid) survives the landing.
// The survival cannot be cured with a process group, and the test measures it; what the test asserts is that the help does not
// promise more than the verb does: it names the process that leaves the group.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestAFResidual_N8TheHelpOfLandNamesTheProcessThatLeavesTheGroup(t *testing.T) {
	afSkipWithoutSh(t)
	if _, err := exec.LookPath("perl"); err != nil {
		t.Skip("perl is needed to start a process in a session of its own")
	}
	db, _, _, _, _ := afQueuedCard(t)
	pidFile := filepath.Join(t.TempDir(), "escaped.pid")
	// the child starts its own session (POSIX::setsid, what setsid(1) does), closes the gate's pipes and sleeps
	gate := `perl -e 'use POSIX; if (fork() == 0) { POSIX::setsid(); open(STDIN, "</dev/null"); open(STDOUT, ">/dev/null"); open(STDERR, ">&STDOUT"); ` +
		`open(F, ">", $ARGV[0]); print F $$; close(F); exec("sleep", "60"); }' '` + pidFile + `'; sleep 0.5`
	out, errOut, err := afRunBoth("land", "t1", "--db", db, "--gate", gate)
	if err != nil || !strings.Contains(out, `"outcome":"landed"`) {
		t.Fatalf("setup: the landing did not land: %v %.200s %.200s", err, out, errOut)
	}
	raw, rerr := os.ReadFile(pidFile)
	if rerr != nil {
		t.Fatalf("setup: the child wrote no pid: %v", rerr)
	}
	pid, perr := strconv.Atoi(strings.TrimSpace(string(raw)))
	if perr != nil || pid <= 1 {
		t.Fatalf("setup: pid %q: %v", raw, perr)
	}
	t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) }) // that pid only, recorded above
	time.Sleep(200 * time.Millisecond)
	alive := syscall.Kill(pid, 0) == nil
	t.Logf("the process that left the group (pid %d) is alive after the landing returned: %v", pid, alive)
	help := afHelpOf(t, "land")
	if alive && !strings.Contains(help, "setsid") {
		t.Errorf("a process the gate started in a session of its own survived the landing (pid %d), and the help promises only that a background process the gate left running is killed", pid)
	}
}
