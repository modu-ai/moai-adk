package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/spf13/cobra"
)

// factory_lane_relaunch_rerun_test.go pins SPEC-FACTORY-STALE-RUN-HEAL-001
// REQ-SRH-011 / AC-SRH-015: the cc/glm relaunch loop re-enters the lane-join
// gate before EVERY lease, so a run switch between two iterations moves the
// next lease and the next child onto the new run. The launch, the join gate
// and the lease are the three substituted seams; nothing here depends on
// wall-clock time or on a live factory run.

// relaunchLoopRun is one loop invocation's observations.
type relaunchLoopRun struct {
	err         error
	gateCalls   []relaunchGateCall
	leaseRunIDs []string
	childRunIDs []string
	childDirs   []string
}

type relaunchGateCall struct{ explicit, leadTarget string }

// relaunchLoopDrive runs runFactoryLaneRelaunch with the three seams
// substituted. gateAnswers[i] is the join gate's answer on its (i+1)-th call:
// a run id (stamped into the environment exactly as the real gate does) or,
// when it is an "!"-prefixed string, a refusal carrying the remainder.
// leaseCards is how many cards the lease seam hands out before it answers
// "no card".
func relaunchLoopDrive(t *testing.T, gateAnswers []string, leaseCards int, explicit, leadTarget string) *relaunchLoopRun {
	t.Helper()
	root, _ := fcFixture(t)
	t.Chdir(root)
	t.Setenv(config.EnvClaudeProjectDir, root)
	sdScrubLauncherEnv(t)
	// The launcher's own gate stamped the first run before the divert.
	t.Setenv(config.EnvFactoryRunID, "X")

	obs := &relaunchLoopRun{}

	prevLook := claudeLookPath
	claudeLookPath = func(string) (string, error) { return "/sentinel/claude", nil }
	prevLaunch := factoryLaneCardLaunchFn
	factoryLaneCardLaunchFn = func(c *exec.Cmd) error {
		env := sdEnvOf(t, c.Env)
		obs.childRunIDs = append(obs.childRunIDs, env[config.EnvFactoryRunID])
		obs.childDirs = append(obs.childDirs, c.Dir)
		return nil
	}
	prevGate := factoryLaneRunGateFn
	factoryLaneRunGateFn = func(_, gotExplicit, gotLead string, _ *factoryLaunchTiming) (func(), error) {
		call := len(obs.gateCalls)
		obs.gateCalls = append(obs.gateCalls, relaunchGateCall{explicit: gotExplicit, leadTarget: gotLead})
		if call >= len(gateAnswers) {
			t.Errorf("join gate called %d times, the scenario scripts only %d answers", call+1, len(gateAnswers))
			return func() {}, errors.New("unscripted gate call")
		}
		answer := gateAnswers[call]
		if refusal, refused := strings.CutPrefix(answer, "!"); refused {
			return func() {}, errors.New(refusal)
		}
		prev, had := os.LookupEnv(config.EnvFactoryRunID)
		_ = os.Setenv(config.EnvFactoryRunID, answer)
		return func() {
			if had {
				_ = os.Setenv(config.EnvFactoryRunID, prev)
			} else {
				_ = os.Unsetenv(config.EnvFactoryRunID)
			}
		}, nil
	}
	prevLease := factoryLaneLeaseFn
	leased := 0
	factoryLaneLeaseFn = func(_ context.Context, _, runID, _ string) (homestate.Card, bool, error) {
		obs.leaseRunIDs = append(obs.leaseRunIDs, runID)
		if leased >= leaseCards {
			return homestate.Card{}, false, nil
		}
		leased++
		return homestate.Card{CardID: "t" + string(rune('0'+leased)), WorktreePath: t.TempDir()}, true, nil
	}
	prevRoot := findProjectRootFn
	findProjectRootFn = func() (string, error) { return root, nil }
	prevDeps := deps
	deps = nil
	t.Cleanup(func() {
		claudeLookPath, factoryLaneCardLaunchFn = prevLook, prevLaunch
		factoryLaneRunGateFn, factoryLaneLeaseFn = prevGate, prevLease
		findProjectRootFn, deps = prevRoot, prevDeps
	})

	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	obs.err = runFactoryLaneRelaunch(cmd, "lane-3", []string{"--name", "lane-3"}, explicit, leadTarget)
	return obs
}

func TestRelaunchLoopReResolvesRun(t *testing.T) {
	t.Run("a run switch between iterations moves the next lease and child onto the new run", func(t *testing.T) {
		// Iteration 1 joins X; X is retired and Y activated while its child
		// runs; iteration 2 must join Y; iteration 3 finds no card.
		obs := relaunchLoopDrive(t, []string{"X", "Y", "Y"}, 2, "", "")
		if obs.err != nil {
			t.Fatalf("loop: %v", obs.err)
		}
		if got := len(obs.gateCalls); got != 3 {
			t.Errorf("the join gate was called %d times, want once per iteration (3: two cards plus the no-card iteration)", got)
		}
		if want := []string{"X", "Y", "Y"}; !reflect.DeepEqual(obs.leaseRunIDs, want) {
			t.Errorf("leases were taken from runs %v, want %v (iteration 2 must lease from the NEW run)", obs.leaseRunIDs, want)
		}
		if want := []string{"X", "Y"}; !reflect.DeepEqual(obs.childRunIDs, want) {
			t.Errorf("children saw %s=%v, want %v (iteration 2's child must carry the new run)", config.EnvFactoryRunID, obs.childRunIDs, want)
		}
		if got := os.Getenv(config.EnvFactoryRunID); got != "X" {
			t.Errorf("after the loop %s = %q, want the launcher's own stamp %q restored", config.EnvFactoryRunID, got, "X")
		}
	})

	t.Run("no active run at iteration 2 stops with the gate's refusal and leases nothing", func(t *testing.T) {
		obs := relaunchLoopDrive(t, []string{"X", "!" + noActiveFactorySentinel}, 5, "", "")
		if obs.err == nil || !strings.Contains(obs.err.Error(), noActiveFactorySentinel) {
			t.Fatalf("loop error = %v, want one carrying the gate's %s text", obs.err, noActiveFactorySentinel)
		}
		if want := []string{"X"}; !reflect.DeepEqual(obs.leaseRunIDs, want) {
			t.Errorf("leases = %v, want only iteration 1's %v (a refusal leases nothing)", obs.leaseRunIDs, want)
		}
		if got := len(obs.childRunIDs); got != 1 {
			t.Errorf("%d children started, want 1", got)
		}
	})

	t.Run("an explicit selection that retires stops the loop and is passed to every gate call", func(t *testing.T) {
		obs := relaunchLoopDrive(t, []string{"X", "!" + noActiveFactorySentinel}, 5, "X", "lead-9")
		if obs.err == nil || !strings.Contains(obs.err.Error(), noActiveFactorySentinel) {
			t.Fatalf("loop error = %v, want the explicit selection's refusal", obs.err)
		}
		if got := len(obs.leaseRunIDs); got != 1 {
			t.Errorf("%d leases, want 1 (the retired explicit run is never leased from again)", got)
		}
		want := []relaunchGateCall{{explicit: "X", leadTarget: "lead-9"}, {explicit: "X", leadTarget: "lead-9"}}
		if !reflect.DeepEqual(obs.gateCalls, want) {
			t.Errorf("gate calls = %+v, want %+v (the launcher's explicit selector and leader target on every iteration)", obs.gateCalls, want)
		}
	})
}
