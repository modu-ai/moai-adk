package cli

// lane-19 34th review of `moai af land` (c58e444a87): a card that left MERGING must not land; the gate command is
// run like any other child of the merge queue (killed with its process group on timeout, environment scrubbed of
// repository-scoping variables, output bounded, secrets in its output redacted); the landed note records which
// gate verified the landing; and the landing lock of a dead holder does not hold the queue for its whole ttl.
// Everything runs on a real ledger and a real git repository; a gate is a shell command.

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func afSkipWithoutSh(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the gate runs with sh, which native Windows does not have")
	}
}

func afLandJSON(t *testing.T, args ...string) (outcome, reason string) {
	t.Helper()
	var res struct {
		Outcome string `json:"outcome"`
		Reason  string `json:"reason"`
	}
	out := runAF(t, args...)
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		t.Fatalf("af output %q: %v", out, err)
	}
	return res.Outcome, res.Reason
}

func TestAFLand_ACardThatLeftMergingIsNotLanded(t *testing.T) {
	ctx := context.Background()
	db, repo, _, st, epoch := afQueuedCard(t)
	before := afGit(t, repo, "rev-parse", "refs/heads/af/trunk")
	// the holder withdraws the work: the card goes back to REVIEWING while its queue row stays at the same epoch
	runAF(t, "move", "t1", "--db", db, "--epoch", strconv.FormatInt(epoch, 10), "--stage", "REVIEWING", "--actor", "lane", "--reason", "withdrawn")

	cmd := newAFCmd()
	cmd.SetArgs([]string{"land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", "true"})
	cmd.SetOut(&strings.Builder{})
	cmd.SetErr(&strings.Builder{})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "MERGING") {
		t.Fatalf("landing a card that left MERGING: %v, want a refusal naming MERGING", err)
	}
	if got := afGit(t, repo, "rev-parse", "refs/heads/af/trunk"); got != before {
		t.Errorf("the withdrawn work reached the trunk: %s -> %s", before, got)
	}
	rows, _ := st.Queue(ctx)
	if len(rows) != 1 || rows[0].State == ledger.QueueLanded {
		t.Errorf("queue %+v after the refusal", rows)
	}
}

func TestAFLand_ATimedOutGateIsKilledWithItsChildren(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, _, _ := afQueuedCard(t)
	pidFile := filepath.Join(t.TempDir(), "child.pid")
	gate := "sleep 60 & echo $! > '" + pidFile + "'; wait"
	start := time.Now()
	outcome, reason := afLandJSON(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", gate, "--gate-timeout", "1s")
	if outcome != "failed" || !strings.Contains(reason, "gate") {
		t.Fatalf("outcome %q reason %q, want a failed card naming the gate", outcome, reason)
	}
	if took := time.Since(start); took > 20*time.Second {
		t.Errorf("the timed-out gate held the command for %v", took)
	}
	raw, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("the gate did not record its child: %v", err)
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(raw)))
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := exec.Command("kill", "-0", strconv.Itoa(pid)).Run(); err != nil {
			return // gone
		}
		time.Sleep(100 * time.Millisecond)
	}
	_ = exec.Command("kill", "-9", strconv.Itoa(pid)).Run()
	t.Errorf("the gate's child process %d outlived the timeout", pid)
}

func TestAFLand_TheGateRunsWithoutRepositoryScopingVariables(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, _, _ := afQueuedCard(t)
	prev, had := os.LookupEnv("GIT_DIR")
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("GIT_DIR", prev)
		} else {
			_ = os.Unsetenv("GIT_DIR")
		}
	})
	_ = os.Setenv("GIT_DIR", filepath.Join(t.TempDir(), "elsewhere"))
	outcome, reason := afLandJSON(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", `test -z "$GIT_DIR"`)
	if outcome != "landed" {
		t.Errorf("outcome %q reason %q: the gate saw GIT_DIR", outcome, reason)
	}
}

func TestAFLand_ALandedNoteSaysWhichGateVerifiedIt(t *testing.T) {
	afSkipWithoutSh(t)
	ctx := context.Background()
	db, repo, _, st, _ := afQueuedCard(t)
	gate := "test -f a.txt"
	if outcome, reason := afLandJSON(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", gate); outcome != "landed" {
		t.Fatalf("outcome %q reason %q", outcome, reason)
	}
	notes, _ := st.NotesSince(ctx, 0)
	for _, n := range notes {
		if n.Kind == ledger.NoteLanded && n.CardID == "t1" {
			if !strings.Contains(n.Text, afGateID(gate)) || strings.Contains(n.Text, gate) { // N29: named by an id of its text, never by the text
				t.Errorf("landed note %q does not name the gate %q by its id %q", n.Text, gate, afGateID(gate))
			}
			return
		}
	}
	t.Errorf("no landed note: %+v", notes)
}

func TestAFLand_ASecretInAFailedGatesOutputIsRedactedFromTheReason(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, st, _ := afQueuedCard(t)
	const secret = "hunter2supersecretvalue"
	outcome, reason := afLandJSON(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", "echo password="+secret+"; exit 1")
	if outcome != "failed" {
		t.Fatalf("outcome %q reason %q, want failed", outcome, reason)
	}
	if strings.Contains(reason, secret) {
		t.Errorf("the reason carries the secret: %q", reason)
	}
	rows, _ := st.Queue(context.Background())
	if len(rows) == 1 && strings.Contains(rows[0].FailReason, secret) {
		t.Errorf("the ledger's fail reason carries the secret: %q", rows[0].FailReason)
	}
}

func TestAFLand_TheLandingLockOfADeadHolderDoesNotHoldTheQueue(t *testing.T) {
	afSkipWithoutSh(t)
	ctx := context.Background()
	db, repo, _, st, _ := afQueuedCard(t)
	// a process that held the landing lock and is gone: start one, record its identity, kill it
	child := exec.Command("sleep", "60")
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	fp, state := homestate.ProbeProcessIdentity(child.Process.Pid)
	if state != homestate.ProcessIdentityLive || fp == "" {
		_ = child.Process.Kill()
		_ = child.Wait()
		t.Skipf("this platform cannot fingerprint a process (state %v)", state)
	}
	if _, err := st.AcquireLock(ctx, ledger.LockMerge, ledger.LockHolder{SessionID: "dead-af-land", PID: child.Process.Pid, PIDStart: fp}, 10*time.Minute); err != nil {
		t.Fatal(err)
	}
	_ = child.Process.Kill()
	_ = child.Wait()

	outcome, reason := afLandJSON(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", "true")
	if outcome != "landed" {
		t.Errorf("outcome %q reason %q: a dead holder's lock held the queue", outcome, reason)
	}
}
