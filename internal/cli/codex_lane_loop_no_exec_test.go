//go:build !windows

// codex_lane_loop_no_exec_test.go — card t1488: the `moai codex -l` lane loop
// must survive its first card child on the DEFAULT direct-door launch path.
// The POSIX default once replaced the launcher process (syscall.Exec), so the
// loop never reached the second lease. The test re-executes the test binary as
// a child process (the replacement would otherwise end the test runner itself)
// and lets that child run the loop with the production launch function and a
// fake codex executable that logs one line per invocation.
package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

const laneLoopNoExecChildEnv = "MOAI_T1488_LANE_LOOP_CHILD"

func TestCodexLaneLoopDefaultLaunchContinuesAfterFirstCard(t *testing.T) {
	if logPath := os.Getenv(laneLoopNoExecChildEnv); logPath != "" {
		laneLoopNoExecChild(t, logPath)
		return
	}
	logPath := filepath.Join(t.TempDir(), "codex.log")
	child := exec.Command(os.Args[0], "-test.run=^TestCodexLaneLoopDefaultLaunchContinuesAfterFirstCard$", "-test.v")
	// The child's HOME is a temp dir: the fixture clears MOAI_HOME, so the
	// lane's run-state writes would otherwise land in the real ~/.moai.
	child.Env = append(os.Environ(), laneLoopNoExecChildEnv+"="+logPath, "HOME="+t.TempDir())
	out, err := child.CombinedOutput()
	childPID := 0
	if child.Process != nil {
		childPID = child.Process.Pid
	}
	data, _ := os.ReadFile(logPath)
	lines := strings.Fields(strings.TrimSpace(string(data)))
	t.Logf("child pid=%d err=%v\nfake-codex log:\n%s", childPID, err, data)
	if err != nil {
		t.Fatalf("child test failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "lane loop returned") {
		t.Errorf("the lane loop did not return in the child (process replaced?):\n%s", out)
	}
	if len(lines) != 2 {
		t.Fatalf("fake codex invoked %d times, want 2 (one per queued card): %q", len(lines), lines)
	}
	for _, l := range lines {
		if strings.HasSuffix(l, "pid="+strconv.Itoa(childPID)) {
			t.Errorf("fake codex ran as the launcher process itself (%s): the launcher was replaced", l)
		}
	}
}

func laneLoopNoExecChild(t *testing.T, logPath string) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	sdRecordLeaderRun(t, root, fcRun, factory.BackendClaude)
	t.Chdir(root)
	sdScrubLauncherEnv(t)

	fake := filepath.Join(t.TempDir(), "codex")
	script := "#!/bin/sh\necho \"card=$" + config.EnvFactoryCard + ",pid=$$\" >> '" + logPath + "'\nexit 0\n"
	if err := os.WriteFile(fake, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	prevLook := codexLookPath
	codexLookPath = func(string) (string, error) { return fake, nil }
	prevDirect := codexDirectLaunchFn
	// The production launch function runs; the session's own card work is
	// simulated only AFTER it returns, which a replaced process never reaches.
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		err := defaultCodexDirectLaunch(c)
		sdCodexSessionWork(t, root, sdEnvOf(t, c.Env)[config.EnvFactoryCard])
		return err
	}
	t.Cleanup(func() { codexLookPath, codexDirectLaunchFn = prevLook, prevDirect })

	_, _, err := runCodexCmd(t, "-l")
	t.Logf("lane loop returned: %v", err)
	if err != nil {
		t.Fatal(err)
	}
}
