package cli

// factory_net_m1_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M1 (card t1399),
// AC-015 / AC-009: the factory safety net and the enterable-pair matrix.
//
// The net exercises the factory through the existing launch seam
// (unifiedLaunchFunc, the codex direct-launch seam) and reads what a launched
// session would see: the environment live at the launch call, the argv handed
// to the engine, and the records the launcher wrote. It is authored BEFORE any
// production change so that every later milestone (the entry grammar, the
// kanban removal, the renames) runs against it, and it is probed with eight
// scratch mutants (acceptance.md AC-015 (a)-(h)) whose reds are recorded in
// progress.md.
//
// The lane entry is `-l` on cc and glm (re-pinned from the removed `-f lane` by M2)
// and on codex (re-pinned by M3). The marker
// names asserted here go through the internal/config constants for the
// factory-owned names; AC-016's frozen-value test owns the literals.
//
// Every test sets or clears every env axis it reads (netScrubLaneEnv): this
// suite runs inside factory lane sessions whose own MOAI_FACTORY* and
// MOAI_KANBAN* variables would otherwise read as published markers.

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// netLaunch is what one substituted engine launch observed.
type netLaunch struct {
	launched bool
	err      error
	args     []string
	env      map[string]string // the full process environment at the launch call
	settings map[string]any    // the parsed --settings payload, nil when none
}

func netCC(args []string) error {
	ccCmd.SetOut(io.Discard)
	ccCmd.SetErr(io.Discard)
	return runCC(ccCmd, args)
}

func netGLM(args []string) error {
	glmCmd.SetOut(io.Discard)
	glmCmd.SetErr(io.Discard)
	return runGLM(glmCmd, args)
}

// netDriveLaunch drives one cc/glm launch with the engine launch substituted
// and returns what the launcher stamped. The optional during callbacks run
// inside the substitution, while the stamped environment is still live — the
// enter helpers restore the environment when the entry returns. The settings
// file the launcher injects is read inside the substitution too, because the
// launcher removes it on return.
func netDriveLaunch(t *testing.T, root string, entry func([]string) error, args []string, during ...func()) netLaunch {
	t.Helper()
	var got netLaunch
	prevLaunch := unifiedLaunchFunc
	unifiedLaunchFunc = func(_ string, _ string, launchArgs []string) error {
		got.launched = true
		got.args = append([]string(nil), launchArgs...)
		got.env = netEnvMap()
		for i := 0; i+1 < len(launchArgs); i++ {
			if launchArgs[i] != settingsFlagLong {
				continue
			}
			raw, err := os.ReadFile(launchArgs[i+1])
			if err != nil {
				t.Errorf("the injected --settings file %q is unreadable at launch: %v", launchArgs[i+1], err)
				break
			}
			var payload map[string]any
			if err := json.Unmarshal(raw, &payload); err != nil {
				t.Errorf("the injected --settings file is not JSON: %v", err)
				break
			}
			got.settings = payload
			break
		}
		for _, fn := range during {
			fn()
		}
		return nil
	}
	prevRoot := findProjectRootFn
	findProjectRootFn = func() (string, error) { return root, nil }
	prevDeps := deps
	deps = nil
	defer func() {
		unifiedLaunchFunc = prevLaunch
		findProjectRootFn = prevRoot
		deps = prevDeps
	}()
	got.err = entry(args)
	return got
}

// netLaneFixture builds a project with one active factory run recorded by a
// leader of the given backend, anchors the process in it, and scrubs the lane
// marker environment — the fixture a `-l` launch needs.
func netLaneFixture(t *testing.T, leaderBackend string) string {
	t.Helper()
	root, _ := fcFixture(t)
	sdRecordLeaderRun(t, root, fcRun, leaderBackend)
	t.Chdir(root)
	t.Setenv(config.EnvClaudeProjectDir, root)
	netScrubLaneEnv(t)
	return root
}

// netLeaderFixture builds a bare project for a leader launch, which records
// its own run. The working directory enters the project too: a leader launch
// runs from its project, and the anchor guard reads the working directory
// (SPEC-LAUNCHER-ARGV-001), so the test process must not stand in whichever
// checkout runs the suite — a lane runs it from a card worktree.
func netLeaderFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Setenv("MOAI_HOME", t.TempDir())
	netScrubLaneEnv(t)
	return root
}

func netNamedArg(args []string) string {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == nameFlagLong {
			return args[i+1]
		}
	}
	return ""
}

// netRequireSettingsInjected asserts the transient `crossSessionInbound`
// settings injection reached the launch: the --settings pair in the argv, the
// accept payload in the file, and the signal variable in the environment.
func netRequireSettingsInjected(t *testing.T, launch netLaunch) {
	t.Helper()
	if !containsFlag(launch.args, settingsFlagLong) {
		t.Errorf("the launch argv carries no %s pair: %v", settingsFlagLong, launch.args)
	}
	if launch.settings["crossSessionInbound"] != "accept" {
		t.Errorf("the injected settings payload = %v, want crossSessionInbound=accept", launch.settings)
	}
	if got := launch.env[config.EnvFactorySettingsInjected]; got != "1" {
		t.Errorf("%s at launch = %q, want 1 (the SessionStart hook reads it)", config.EnvFactorySettingsInjected, got)
	}
}

func containsFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func TestFactoryNetLaneLaunch(t *testing.T) {
	for _, tc := range []struct {
		name    string
		backend string
		entry   func([]string) error
	}{
		{"cc", factory.BackendClaude, netCC},
		{"glm", factory.BackendGLM, netGLM},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := netLaneFixture(t, tc.backend)

			first := netDriveLaunch(t, root, tc.entry, []string{"-l"})
			if first.err != nil || !first.launched {
				t.Fatalf("%s -l: launched=%v err=%v", tc.name, first.launched, first.err)
			}
			label := factory.FactoryLaneLabel(1)
			for key, want := range map[string]string{
				config.EnvFactoryRole:                      config.FactoryRoleLane,
				config.EnvMoaiFactoryWorker:                label,
				config.EnvMoaiFactoryWorkers:               "0", // the count is unknown on the incremental form
				config.EnvFactoryRunID:                     fcRun,
				config.EnvFactoryBackend:                   tc.backend,
				config.EnvFactoryAutoDispatch:              config.FactoryDispatchAuto,
				config.EnvClaudeCodeMaxConcurrentSubagents: "10",
			} {
				if got := first.env[key]; got != want {
					t.Errorf("%s at launch = %q, want %q", key, got, want)
				}
			}
			if got := netNamedArg(first.args); got != label {
				t.Errorf("the backend argv names the session %q, want the claimed label %q (argv %v)", got, label, first.args)
			}
			netRequireSettingsInjected(t, first)

			// The lane claim: the registry holds the label under this process.
			reg := loadFactoryRegistry(factoryRegistryPath(root))
			if entry, ok := reg[label]; !ok || entry.PID != os.Getpid() {
				t.Errorf("registry[%s] = (%+v, %v), want a live claim under pid %d", label, entry, ok, os.Getpid())
			}

			// A second lane while the first claim is live bumps to lane-2.
			second := netDriveLaunch(t, root, tc.entry, []string{"-l"})
			if second.err != nil || !second.launched {
				t.Fatalf("%s second -l: launched=%v err=%v", tc.name, second.launched, second.err)
			}
			if got, want := second.env[config.EnvMoaiFactoryWorker], factory.FactoryLaneLabel(2); got != want {
				t.Errorf("second lane %s = %q, want %q (the first claim is live)", config.EnvMoaiFactoryWorker, got, want)
			}
			reg = loadFactoryRegistry(factoryRegistryPath(root))
			if _, ok := reg[factory.FactoryLaneLabel(2)]; !ok {
				t.Errorf("registry lost the second claim: %v", reg)
			}
		})
	}
}

// TestFactoryNetBlockCap pins the Stop-hook block-cap raise reaching a factory
// session through the FACTORY clause: with no kanban signal at all (the
// factory entry publishes none), the launched leader and the launched lane
// carry MOAI_FACTORY_WORKERS, and that alone makes the inject raise the cap;
// with the factory signal removed the inject leaves the environment unchanged
// (the negative control that proves the factory variable is the trigger).
func TestFactoryNetBlockCap(t *testing.T) {
	wantCap := config.EnvClaudeCodeStopHookBlockCap + "=200"
	base := []string{"PATH=/usr/bin", "HOME=/tmp"}

	check := func(t *testing.T, role string) {
		t.Helper()
		for _, key := range []string{retiredLeaderMarker, retiredLaneLabelMarker} {
			if v := os.Getenv(key); v != "" {
				t.Fatalf("%s=%q at launch: the factory entry must publish no kanban signal for this net to isolate the factory clause", key, v)
			}
		}
		if os.Getenv(config.EnvMoaiFactoryWorkers) == "" {
			t.Fatalf("%s role: MOAI_FACTORY_WORKERS is empty at launch", role)
		}
		raised := injectStopHookBlockCapForGoal(context.Background(), base, t.TempDir(), "")
		if !containsEntry(raised, wantCap) {
			t.Errorf("%s: injectStopHookBlockCapForGoal = %v, want %q (the factory clause)", role, raised, wantCap)
		}
		// Negative control: the same call without the factory signal.
		prev := os.Getenv(config.EnvMoaiFactoryWorkers)
		_ = os.Unsetenv(config.EnvMoaiFactoryWorkers)
		unraised := injectStopHookBlockCapForGoal(context.Background(), base, t.TempDir(), "")
		_ = os.Setenv(config.EnvMoaiFactoryWorkers, prev)
		if containsEntry(unraised, wantCap) || len(unraised) != len(base) {
			t.Errorf("%s negative control: %v, want the base environment unchanged without %s", role, unraised, config.EnvMoaiFactoryWorkers)
		}
	}

	t.Run("leader", func(t *testing.T) {
		root := netLeaderFixture(t)
		launch := netDriveLaunch(t, root, netCC, []string{"-f"}, func() { check(t, "leader") })
		if launch.err != nil || !launch.launched {
			t.Fatalf("cc -f: launched=%v err=%v", launch.launched, launch.err)
		}
	})
	t.Run("lane", func(t *testing.T) {
		root := netLaneFixture(t, factory.BackendClaude)
		launch := netDriveLaunch(t, root, netCC, []string{"-l"}, func() { check(t, "lane") })
		if launch.err != nil || !launch.launched {
			t.Fatalf("cc -l: launched=%v err=%v", launch.launched, launch.err)
		}
	})
}

func containsEntry(env []string, entry string) bool {
	for _, e := range env {
		if e == entry {
			return true
		}
	}
	return false
}
