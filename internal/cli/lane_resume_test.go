package cli

import (
	"context"
	"io"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/spf13/cobra"
)

// TestLaneJoinChildArgv (AC-SCV-008, REQ-SCV-008) — the pure assembler
// returns the child argv carrying the name pair, the settings pair, and the
// pass-through tokens in a deterministic order; it mutates no input, and two
// runs over the same inputs produce identical argv. Purity is by
// construction: the function takes no environment, filesystem, or process
// parameter and returns []string.
func TestLaneJoinChildArgv(t *testing.T) {
	launcherArgs := []string{"--name", "lane-3", "--", "--resume", "<session-id>"}
	settings := []string{"--settings", "<file>"}

	got := laneJoinChildArgv(launcherArgs, "lane-3", settings)
	want := []string{"--name", "lane-3", "--", "--resume", "<session-id>", "--settings", "<file>"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argv = %q, want %q", got, want)
	}
	again := laneJoinChildArgv(launcherArgs, "lane-3", settings)
	if !reflect.DeepEqual(got, again) {
		t.Fatalf("assembler is not deterministic: %q vs %q", got, again)
	}
	if !reflect.DeepEqual(launcherArgs, []string{"--name", "lane-3", "--", "--resume", "<session-id>"}) {
		t.Fatalf("the assembler mutated its input: %q", launcherArgs)
	}

	// Nameless launcher args get the injected pair before the pass-through
	// marker, so the desugared lane name can never land in the child's
	// pass-through region.
	got = laneJoinChildArgv([]string{"--", "--resume", "<session-id>"}, "lane-7", nil)
	want = []string{"--name", "lane-7", "--", "--resume", "<session-id>"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("nameless argv = %q, want %q", got, want)
	}
}

// TestResumeRequiresValue (AC-SCV-009, REQ-SCV-009) — a --resume token with
// no following value refuses before any launch, naming the required
// --resume <session-id> form, and the launch seam is never called. The
// refusing shapes run through runClaudeEntry with a counting launch stub; the
// accepting shapes assert the validation itself passes them.
func TestResumeRequiresValue(t *testing.T) {
	refusing := []struct {
		name string
		args []string
	}{
		{"space form at argv end", []string{"--", "--resume"}},
		{"space form before another flag", []string{"--", "--resume", "--settings", "x"}},
		{"equals form empty value", []string{"--", "--resume="}},
	}
	for _, tc := range refusing {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			launch := func(string, string, []string) error {
				calls++
				return nil
			}
			cmd := &cobra.Command{}
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			err := runClaudeEntry(cmd, tc.args, "cc", "claude", factory.BackendClaude, launch)
			if err == nil {
				t.Fatal("expected the valueless --resume to refuse")
			}
			if !strings.Contains(err.Error(), "--resume <session-id>") {
				t.Errorf("refusal must name --resume <session-id>; got: %v", err)
			}
			if calls != 0 {
				t.Errorf("launch seam called %d time(s), want 0", calls)
			}
		})
	}

	t.Run("well-formed values pass validation", func(t *testing.T) {
		if err := validateResumeArgs([]string{"--", "--resume", "<session-id>"}); err != nil {
			t.Errorf("space form refused: %v", err)
		}
		if err := validateResumeArgs([]string{"--", "--resume=<session-id>"}); err != nil {
			t.Errorf("equals form refused: %v", err)
		}
		if err := validateResumeArgs([]string{"--name", "lane-3", "--"}); err != nil {
			t.Errorf("a launch with no --resume token refused: %v", err)
		}
	})
}

// saveRelaunchSeams captures the relaunch loop's three seams and restores
// them at cleanup.
func saveRelaunchSeams(t *testing.T) (gateCalls, leaseCalls, launchCalls *int) {
	t.Helper()
	prevGate := factoryLaneRunGateFn
	prevLease := factoryLaneLeaseFn
	prevLaunch := factoryLaneCardLaunchFn
	gate, lease, launch := 0, 0, 0
	factoryLaneRunGateFn = func(string, string, string, *factoryLaunchTiming) (func(), error) {
		gate++
		return func() {}, nil
	}
	factoryLaneLeaseFn = func(context.Context, string, string, string) (homestate.Card, bool, error) {
		lease++
		return homestate.Card{}, false, nil
	}
	factoryLaneCardLaunchFn = func(*exec.Cmd) error {
		launch++
		return nil
	}
	t.Cleanup(func() {
		factoryLaneRunGateFn = prevGate
		factoryLaneLeaseFn = prevLease
		factoryLaneCardLaunchFn = prevLaunch
	})
	return &gate, &lease, &launch
}

// TestRelaunchRefusesResumeToken (AC-SCV-010, REQ-SCV-010) — under the
// relaunch policy, child arguments carrying a --resume token in either
// spelling refuse the loop before its first iteration: the error names the
// one-shot lane-join form verbatim, and zero card sessions are started (the
// token is neither propagated nor stripped).
func TestRelaunchRefusesResumeToken(t *testing.T) {
	refusing := []struct {
		name string
		args []string
	}{
		{"space form", []string{"--name", "lane-3", "--", "--resume", "<session-id>"}},
		{"equals form", []string{"--name", "lane-3", "--", "--resume=<session-id>"}},
	}
	for _, tc := range refusing {
		t.Run(tc.name, func(t *testing.T) {
			gateCalls, leaseCalls, launchCalls := saveRelaunchSeams(t)
			cmd := &cobra.Command{}
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			err := runFactoryLaneRelaunch(cmd, "lane-3", tc.args, "", "")
			if err == nil {
				t.Fatal("expected the relaunch loop to refuse a --resume token")
			}
			if !strings.Contains(err.Error(), "moai cc -l -- --resume <session-id>") {
				t.Errorf("refusal must name the one-shot lane-join form verbatim; got: %v", err)
			}
			if *gateCalls != 0 || *leaseCalls != 0 || *launchCalls != 0 {
				t.Errorf("refusal must precede every seam: gate=%d lease=%d launch=%d, want all 0",
					*gateCalls, *leaseCalls, *launchCalls)
			}
		})
	}

	t.Run("no resume token enters the loop", func(t *testing.T) {
		// relaunchLoopDrive (the rerun SPEC's harness) stubs the three seams
		// against a fixture primary checkout and drives the loop with
		// token-free args: the gate answers once, the lease answers "no
		// card", and the loop stops on its own stop condition — proving the
		// guard is not what ends this loop.
		obs := relaunchLoopDrive(t, []string{"X"}, 0, "", "")
		if obs.err != nil {
			t.Fatalf("loop without a resume token must run: %v", obs.err)
		}
		if len(obs.gateCalls) != 1 {
			t.Errorf("gate calls = %d, want 1", len(obs.gateCalls))
		}
		if !reflect.DeepEqual(obs.leaseRunIDs, []string{"X"}) {
			t.Errorf("leases = %v, want [X]", obs.leaseRunIDs)
		}
	})
}
