package cli

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// deadOwnerIdentity returns the identity of a process that has exited.
// Whether the pid is later recycled or not the classification is dead: an
// unrecycled pid probes dead, a recycled one carries a different fingerprint.
func deadOwnerIdentity(t *testing.T) (int, string) {
	t.Helper()
	cmd := exec.Command("/bin/sh", "-c", "exit 0")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start short-lived process: %v", err)
	}
	pid := cmd.Process.Pid
	fingerprint, _ := homestate.ProbeProcessIdentity(pid)
	if fingerprint == "" {
		fingerprint = "fingerprint-of-a-process-that-has-exited"
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("wait for short-lived process: %v", err)
	}
	return pid, fingerprint
}

func runFactoryCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newFactoryCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

// AC-010 — `moai factory runs` reports every active run with its status and
// owner classification; `--retire` refuses a live owner AND an indeterminate
// owner, each refusal exiting non-zero, leaving the run active, and naming its
// classification as the reason. Only the dead-owner run is retired.
func TestFactoryRunsCommandReportsAndRefuses(t *testing.T) {
	root := runOwnerSandbox(t)
	t.Chdir(root)

	livePID := os.Getpid()
	liveStart := homestate.CurrentProcessFingerprint()
	if liveStart == "" {
		t.Skip("this host cannot report its own process-start fingerprint")
	}
	deadPID, deadStart := deadOwnerIdentity(t)

	seedRun(t, root, "run-live", livePID, liveStart)
	seedRun(t, root, "run-dead", deadPID, deadStart)
	seedRun(t, root, "run-unknown", 0, "")

	out, err := runFactoryCommand(t, "runs")
	if err != nil {
		t.Fatalf("factory runs: %v (output %q)", err, out)
	}
	for _, want := range []string{
		"run-live", "live",
		"run-dead", "dead",
		"run-unknown", "indeterminate",
		"active",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("listing %q does not contain %q", out, want)
		}
	}

	for _, run := range []struct{ id, classification string }{
		{"run-live", "live"},
		{"run-unknown", "indeterminate"},
	} {
		out, err := runFactoryCommand(t, "runs", "--retire", run.id)
		if err == nil {
			t.Fatalf("--retire %s succeeded; a run whose owner is %s must be refused", run.id, run.classification)
		}
		if !strings.Contains(err.Error(), run.classification) {
			t.Fatalf("--retire %s error %q does not name the classification %q (output %q)", run.id, err, run.classification, out)
		}
		if got := runStatusOf(t, root, run.id); got != "active" {
			t.Fatalf("--retire %s left status %q, want \"active\"", run.id, got)
		}
	}

	if _, err := runFactoryCommand(t, "runs", "--retire", "run-dead"); err != nil {
		t.Fatalf("--retire run-dead: %v", err)
	}
	if got := runStatusOf(t, root, "run-dead"); got != "retired" {
		t.Fatalf("run-dead status = %q, want \"retired\"", got)
	}
}

func runStatusOf(t *testing.T, root, runID string) string {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	var status string
	if err := db.DB.QueryRow(`SELECT status FROM runs WHERE run_id=?`, runID).Scan(&status); err != nil {
		t.Fatalf("read status %s: %v", runID, err)
	}
	return status
}
