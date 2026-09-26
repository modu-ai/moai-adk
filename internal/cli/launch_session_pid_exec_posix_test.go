//go:build !windows

package cli

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// execHelperEnv gates the helper below. It is deliberately checked FIRST in the
// helper and answered with a skip, so a plain `go test` run — which has no such
// variable — never execs anything. The parent runs the helper with an explicit
// -test.run filter as well, so a stray invocation cannot fan out into the rest
// of the suite.
const execHelperEnv = "MOAI_TEST_LAUNCH_EXEC_HELPER"

// TestSessionPIDStampExecHelper is the child half of the round-trip below. It
// is not a test in its own right: under a normal run it skips immediately.
//
// When the parent enables it, it calls the production launch path with /bin/sh
// standing in for the claude binary, and the shell prints two numbers: the
// stamped MOAI_SESSION_PID, and its own PID. execve(2) replaces this process
// rather than forking, so the shell IS this process — the two numbers must
// agree.
func TestSessionPIDStampExecHelper(t *testing.T) {
	if os.Getenv(execHelperEnv) != "1" {
		t.Skip("helper process for TestExecOrSpawnClaude_StampsLiveSessionPID")
	}
	// From here on the process is replaced; nothing after this call runs.
	err := execOrSpawnClaude("/bin/sh",
		[]string{"sh", "-c", `printf '%s %s' "$MOAI_SESSION_PID" "$$"`},
		os.Environ())
	t.Fatalf("execOrSpawnClaude returned instead of replacing the process: %v", err)
}

// TestExecOrSpawnClaude_StampsLiveSessionPID is the behavioral proof that the
// launch path hands its successor a PID that names the successor itself: the
// value the launcher stamps and the PID the exec'd process actually runs under
// are the same number, with no ancestry walk anywhere in the chain.
//
// It spawns exactly one bounded child, which immediately execs a shell — there
// is no recursion into the suite and no background load.
func TestExecOrSpawnClaude_StampsLiveSessionPID(t *testing.T) {
	checkExecStampsLiveSessionPID(t)
}

// TestExecOrSpawnClaude_StampsLiveSessionPIDUnderLaneEnv runs the same check
// with a factory lane's launch variables exported. The child inherits the
// parent's environment, so unless the check pins those variables the child
// takes the factory launch-pending path — against the real broker in a real
// lane (card t1222). MOAI_HOME is redirected first so this test itself can
// never reach a live broker.
func TestExecOrSpawnClaude_StampsLiveSessionPIDUnderLaneEnv(t *testing.T) {
	t.Setenv(config.EnvHome, t.TempDir())
	t.Setenv(config.EnvMoaiKanbanID, "run-t1222-probe")
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-7")
	t.Setenv(config.EnvMoaiFactoryWorkers, "3")
	checkExecStampsLiveSessionPID(t)
}

func checkExecStampsLiveSessionPID(t *testing.T) {
	t.Helper()
	// The child inherits os.Environ(); clearing the factory launch variables
	// here keeps it off the launch-pending path that would write a peer into
	// whatever broker MOAI_HOME names.
	clearFactoryTestEnv(t)
	requireFactoryLaunchDisabled(t)
	if _, err := os.Stat("/bin/sh"); err != nil {
		t.Skipf("/bin/sh unavailable: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[0],
		"-test.run=^TestSessionPIDStampExecHelper$", "-test.timeout=20s")
	cmd.Env = append(os.Environ(), execHelperEnv+"=1")

	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("helper process: %v (output %q)", err, out)
	}

	fields := strings.Fields(string(out))
	if len(fields) != 2 {
		t.Fatalf("helper printed %q, want two space-separated PIDs", out)
	}
	stamped, actual := fields[0], fields[1]
	if stamped == "" {
		t.Fatalf("MOAI_SESSION_PID was not stamped into the launch environment (helper printed %q)", out)
	}
	if stamped != actual {
		t.Errorf("stamped MOAI_SESSION_PID = %s, but the exec'd process runs as PID %s; "+
			"the stamp must name the process that becomes the session", stamped, actual)
	}
}
