package cli

// lane-19 35th review of `moai af land`: a gate that exits 0 while a background child still holds its output pipes
// must not fail the card and the child must not outlive the command; a secret is masked before the output is cut to
// its tail; and the help says what environment the gate receives.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestAFLand_AGateThatLeavesABackgroundChildLandsAndTheChildIsKilled(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, _, _ := afQueuedCard(t)
	pidFile := filepath.Join(t.TempDir(), "server.pid")
	gate := "sleep 60 & echo $! > '" + pidFile + "'"
	outcome, reason := afLandJSON(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", gate)
	if outcome != "landed" {
		t.Fatalf("outcome %q reason %q: a gate that exited 0 failed the card", outcome, reason)
	}
	raw, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("the gate did not record its child: %v", err)
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(raw)))
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if exec.Command("kill", "-0", strconv.Itoa(pid)).Run() != nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	_ = exec.Command("kill", "-9", strconv.Itoa(pid)).Run()
	t.Errorf("the gate's background child %d outlived the command", pid)
}

// The secret line is 33 bytes; the filler makes the 2048-byte tail start inside the secret's value, so a cut before
// masking leaves the end of the value behind.
func TestAFLand_ASecretCutByTheTailLimitIsStillRedacted(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, _, _ := afQueuedCard(t)
	gate := "echo password=hunter2supersecretvalue; head -c 2035 /dev/zero | tr '\\0' y; echo; exit 1"
	outcome, reason := afLandJSON(t, "land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", gate)
	if outcome != "failed" {
		t.Fatalf("outcome %q reason %q, want failed", outcome, reason)
	}
	if strings.Contains(reason, "secretvalue") || strings.Contains(reason, "ersecret") {
		t.Errorf("the tail kept the end of the secret: %.400q", reason)
	}
}

func TestAFLand_HelpSaysTheGateReceivesTheCallersEnvironment(t *testing.T) {
	for _, c := range newAFCmd().Commands() {
		if c.Name() == "land" {
			if !strings.Contains(c.Long, "caller's full environment") {
				t.Errorf("the help of `af land` does not warn that the gate receives the caller's environment:\n%s", c.Long)
			}
			return
		}
	}
	t.Fatal("no `af land` verb")
}
