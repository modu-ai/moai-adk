package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestCodexFactoryLeaderCanControlIntegrationPolicy(t *testing.T) {
	root, _ := fcFixture(t)
	t.Chdir(root)
	netScrubLaneEnv(t)
	withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	capture := codexDirectLaunchFn
	checked := false
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		if role := launchEnvValue(c.Env, config.EnvFactoryRole); role != factory.RoleLeader {
			t.Fatalf("leader child role=%q", role)
		}
		hold := newIntegrationPolicyHoldCmd()
		hold.SetArgs([]string{"--reason", "leader-release-cut"})
		if err := hold.Execute(); err != nil {
			t.Fatalf("Codex leader must control policy hold: %v", err)
		}
		policy, err := factory.ReadIntegrationWindowPolicy(root)
		if err != nil || policy.Policy != factory.PolicyHold || policy.Reason != "leader-release-cut" {
			t.Fatalf("leader hold readback=%+v err=%v", policy, err)
		}
		if err := newIntegrationPolicyOpenCmd().Execute(); err != nil {
			t.Fatalf("Codex leader must control policy open: %v", err)
		}
		policy, err = factory.ReadIntegrationWindowPolicy(root)
		if err != nil || policy.Policy != factory.PolicyOpen {
			t.Fatalf("leader open readback=%+v err=%v", policy, err)
		}
		checked = true
		return capture(c)
	}
	if _, _, err := runCodexCmd(t, "-f"); err != nil {
		t.Fatal(err)
	}
	if !checked {
		t.Fatal("leader launch boundary was not reached")
	}
}

func TestIntegrationPolicyRejectsLaneAndUnknownRoles(t *testing.T) {
	for _, role := range []string{config.FactoryRoleLane, "lane-2", "unknown-role", "LEADER"} {
		t.Run(role, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv(config.EnvClaudeProjectDir, root)
			t.Setenv(config.EnvFactoryRole, role)
			if err := factory.WriteIntegrationWindowPolicy(root, factory.IntegrationWindowPolicy{Policy: factory.PolicyHold, Reason: "existing-hold"}); err != nil {
				t.Fatal(err)
			}
			if err := newIntegrationPolicyOpenCmd().Execute(); err == nil {
				t.Fatal("lane or unknown role opened the policy")
			}
			policy, err := factory.ReadIntegrationWindowPolicy(root)
			if err != nil || policy.Policy != factory.PolicyHold || policy.Reason != "existing-hold" {
				t.Fatalf("refused policy write changed record=%+v err=%v", policy, err)
			}
		})
	}
}

func TestCodexFactoryLeaderStartedSessionFailureIsNotLaunchFailure(t *testing.T) {
	root, _ := fcFixture(t)
	t.Chdir(root)
	netScrubLaneEnv(t)
	withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	var runID string
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		runID = launchEnvValue(c.Env, config.EnvFactoryRunID)
		path, err := homestate.FactoryDBPath(root)
		if err != nil {
			t.Fatal(err)
		}
		c.Path = exe
		c.Args = []string{exe, "-test.run=^TestHelperProcess_CodexLeaderExitAfterHandoff$"}
		c.Env = append(c.Env, "MOAI_TEST_CODEX_LEADER_EXIT=1", "MOAI_TEST_CODEX_LEADER_DB="+path, "MOAI_TEST_CODEX_LEADER_RUN="+runID)
		// Exercise the real start-and-wait door used on Windows, even when
		// this portable regression runs on another host.
		err = codexStartAndWait(c)
		if c.ProcessState == nil || c.ProcessState.ExitCode() != 7 {
			t.Fatalf("child did not finish its registered session with exit 7: %v", err)
		}
		if pid, _ := runOwnerStamp(t, root, runID); pid != c.Process.Pid {
			t.Fatalf("child was not the registered run owner: %d", pid)
		}
		// A session's owner can already be cleared by cleanup when its exit
		// reaches the launcher. That must not turn a real session into a
		// never-started run merely because no cards were dispatched.
		if clearErr := clearFactoryRunOwner(root, runID); clearErr != nil {
			t.Fatal(clearErr)
		}
		return err
	}
	_, _, err = runCodexCmd(t, "-f")
	if code, ok := ResolveExitCode(err); !ok || code != 7 {
		t.Fatalf("session exit code=%d intentional=%v err=%v", code, ok, err)
	}
	db := fcOpen(t, root)
	var status string
	if err := db.DB.QueryRowContext(context.Background(), "SELECT status FROM runs WHERE run_id=?", runID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	var failures int
	if err := db.DB.QueryRowContext(context.Background(), "SELECT count(*) FROM events WHERE run_id=? AND kind='run.launch_failed'", runID).Scan(&failures); err != nil {
		t.Fatal(err)
	}
	if status != "active" || failures != 0 {
		t.Fatalf("started session was treated as a launch failure: status=%q launch_failed=%d", status, failures)
	}
}

func TestHelperProcess_CodexLeaderExitAfterHandoff(t *testing.T) {
	if os.Getenv("MOAI_TEST_CODEX_LEADER_EXIT") != "1" {
		return
	}
	db, err := homestate.OpenFactoryReadonly(os.Getenv("MOAI_TEST_CODEX_LEADER_DB"))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(91)
	}
	defer db.Close()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		var pid int
		if err := db.DB.QueryRow("SELECT lead_pid FROM runs WHERE run_id=?", os.Getenv("MOAI_TEST_CODEX_LEADER_RUN")).Scan(&pid); err == nil && pid == os.Getpid() {
			_ = db.Close()
			os.Exit(7)
		}
		time.Sleep(5 * time.Millisecond)
	}
	_, _ = fmt.Fprintln(os.Stderr, "child owner handoff was not observed")
	os.Exit(92)
}
