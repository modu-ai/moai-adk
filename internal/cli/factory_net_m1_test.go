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
// Today's lane entry is `-f lane`; `-l` does not exist yet. M2 (cc, glm) and
// M3 (codex) re-pin the lane launches below from `-f lane` to `-l`. The marker
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
	"os/exec"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
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
// marker environment — the fixture a `-f lane` launch needs.
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
// its own run.
func netLeaderFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Setenv("MOAI_HOME", t.TempDir())
	netScrubLaneEnv(t)
	return root
}

// netCodexLane is what one substituted codex lane child observed.
type netCodexLane struct {
	root string
	env  map[string]string
}

// netCodexLaneChild runs `moai codex -f lane` over a one-card queue with the
// codex binary and child launch substituted, and returns the first child's
// full environment. The substituted session moves its own card to merge-ready
// and exits 0, so the supervising loop ends.
func netCodexLaneChild(t *testing.T) netCodexLane {
	t.Helper()
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStatePicked)
	sdRecordLeaderRun(t, root, fcRun, kanban.BackendClaude)
	t.Chdir(root)
	netScrubLaneEnv(t)

	var childEnv map[string]string
	prevLook, prevDirect := codexLookPath, codexDirectLaunchFn
	codexLookPath = func(string) (string, error) { return "/sentinel/codex", nil }
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		if childEnv == nil {
			childEnv = sdEnvOf(t, c.Env)
		}
		sdCodexSessionWork(t, root, sdEnvOf(t, c.Env)[config.EnvMoaiKanbanCard])
		return nil
	}
	t.Cleanup(func() { codexLookPath, codexDirectLaunchFn = prevLook, prevDirect })

	if _, _, err := runCodexCmd(t, "-f", "lane"); err != nil {
		t.Fatalf("codex lane launch: %v", err)
	}
	if childEnv == nil {
		t.Fatal("the substituted codex child was never launched")
	}
	return netCodexLane{root: root, env: childEnv}
}

// netNamedArg returns the value following the first --name in args.
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
	if got := launch.env[config.EnvMoaiKanbanSettingsInjected]; got != "1" {
		t.Errorf("%s at launch = %q, want 1 (the SessionStart hook reads it)", config.EnvMoaiKanbanSettingsInjected, got)
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

// netRunRecorded reports whether the project's run records name runID, with
// the given backend, in both stores the leader start writes: the queue's
// runtime section and the factory database (with the derived-capacity marker
// a count-less leader records).
func netRunRecorded(t *testing.T, root, runID, backend string) {
	t.Helper()
	record, err := kanban.NewBacklogStore(kanban.BacklogPathForRoot(root)).LoadPure()
	if err != nil {
		t.Fatalf("load the queue record: %v", err)
	}
	found := false
	for _, run := range record.Runtime.Runs {
		if run.RunID == runID {
			found = true
			if run.Backend != backend {
				t.Errorf("queue run %s backend = %q, want %q", runID, run.Backend, backend)
			}
		}
	}
	if !found {
		t.Errorf("the queue's runtime section has no run %q", runID)
	}
	capacity, ok := recordedFactoryLaneCapacity(root, runID)
	if !ok {
		t.Fatalf("the factory database has no run %q", runID)
	}
	if capacity != homestate.LaneCapacityDerived {
		t.Errorf("run %s lane capacity = %d, want the derived-capacity marker %d (a count-less leader never declares a bound)",
			runID, capacity, homestate.LaneCapacityDerived)
	}
}

// TestFactoryNetLeaderLaunch is the leader half of the net (AC-015): a bare
// `-f` on cc and on glm starts a factory leader that records its run, stamps
// the leader markers (and only those), names itself, and receives the
// transient settings injection.
func TestFactoryNetLeaderLaunch(t *testing.T) {
	for _, tc := range []struct {
		name    string
		backend string
		entry   func([]string) error
	}{
		{"cc", kanban.BackendClaude, netCC},
		{"glm", kanban.BackendGLM, netGLM},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := netLeaderFixture(t)
			launch := netDriveLaunch(t, root, tc.entry, []string{"-f"})
			if launch.err != nil || !launch.launched {
				t.Fatalf("%s -f: launched=%v err=%v", tc.name, launch.launched, launch.err)
			}

			runID := launch.env[config.EnvMoaiKanbanID]
			if runID == "" {
				t.Fatalf("%s at launch is empty: the leader carries no run id", config.EnvMoaiKanbanID)
			}
			if got := launch.env[config.EnvMoaiFactoryWorkers]; got != "1" {
				t.Errorf("%s at launch = %q, want 1 (the count-less leader's one-lane default)", config.EnvMoaiFactoryWorkers, got)
			}
			if got, want := launch.env[config.EnvMoaiKanbanLeadAddr], kanban.FactoryLeaderSocketPath(runID); got != want {
				t.Errorf("%s at launch = %q, want the run's leader socket %q", config.EnvMoaiKanbanLeadAddr, got, want)
			}
			if got := launch.env[config.EnvMoaiKanbanBackend]; got != tc.backend {
				t.Errorf("%s at launch = %q, want %q", config.EnvMoaiKanbanBackend, got, tc.backend)
			}
			// A factory leader seeds no chain and is no lane.
			for _, key := range []string{config.EnvMoaiKanban, config.EnvMoaiKanbanLabel, config.EnvMoaiFactoryWorker, config.EnvFactoryRole} {
				if got := launch.env[key]; got != "" {
					t.Errorf("%s at a leader launch = %q, want unset (the marker belongs to another role)", key, got)
				}
			}
			leaderName := launch.env[config.EnvMoaiKanbanLeadName]
			if _, ok := kanban.SplitLeaderLabel(leaderName); !ok {
				t.Errorf("%s at launch = %q, want a leader-shaped name", config.EnvMoaiKanbanLeadName, leaderName)
			}
			if got := netNamedArg(launch.args); got != leaderName {
				t.Errorf("the backend argv names the session %q, want the exported leader name %q (argv %v)", got, leaderName, launch.args)
			}
			if containsFlag(launch.args, "-f") || containsFlag(launch.args, "--factory") {
				t.Errorf("the entry token reached the engine argv: %v", launch.args)
			}
			netRequireSettingsInjected(t, launch)
			netRunRecorded(t, root, runID, tc.backend)

			// The enter helpers restore the process environment on return.
			for _, key := range []string{config.EnvMoaiFactoryWorkers, config.EnvMoaiKanbanID, config.EnvMoaiKanbanLeadAddr, config.EnvMoaiKanbanBackend} {
				if v, ok := os.LookupEnv(key); ok {
					t.Errorf("%s = %q survived the launch; the leader markers must be restored", key, v)
				}
			}
		})
	}
}

// TestFactoryNetLaneLaunch is the lane half of the net (AC-015): a lane
// launched through today's `-f lane` claims a registry slot under its own pid,
// stamps the lane markers, names itself with the claimed label, and receives
// the settings injection; a second lane from the same process takes the next
// label because the first claim is live.
func TestFactoryNetLaneLaunch(t *testing.T) {
	for _, tc := range []struct {
		name    string
		backend string
		entry   func([]string) error
	}{
		{"cc", kanban.BackendClaude, netCC},
		{"glm", kanban.BackendGLM, netGLM},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := netLaneFixture(t, tc.backend)

			first := netDriveLaunch(t, root, tc.entry, []string{"-f", "lane"})
			if first.err != nil || !first.launched {
				t.Fatalf("%s -f lane: launched=%v err=%v", tc.name, first.launched, first.err)
			}
			label := kanban.FactoryLaneLabel(1)
			for key, want := range map[string]string{
				config.EnvFactoryRole:                      config.FactoryRoleLane,
				config.EnvMoaiFactoryWorker:                label,
				config.EnvMoaiFactoryWorkers:               "0", // the count is unknown on the incremental form
				config.EnvMoaiKanbanID:                     fcRun,
				config.EnvMoaiKanbanBackend:                tc.backend,
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
			second := netDriveLaunch(t, root, tc.entry, []string{"-f", "lane"})
			if second.err != nil || !second.launched {
				t.Fatalf("%s second -f lane: launched=%v err=%v", tc.name, second.launched, second.err)
			}
			if got, want := second.env[config.EnvMoaiFactoryWorker], kanban.FactoryLaneLabel(2); got != want {
				t.Errorf("second lane %s = %q, want %q (the first claim is live)", config.EnvMoaiFactoryWorker, got, want)
			}
			reg = loadFactoryRegistry(factoryRegistryPath(root))
			if _, ok := reg[kanban.FactoryLaneLabel(2)]; !ok {
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
		for _, key := range []string{config.EnvMoaiKanban, config.EnvMoaiKanbanLabel} {
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
		root := netLaneFixture(t, kanban.BackendClaude)
		launch := netDriveLaunch(t, root, netCC, []string{"-f", "lane"}, func() { check(t, "lane") })
		if launch.err != nil || !launch.launched {
			t.Fatalf("cc -f lane: launched=%v err=%v", launch.launched, launch.err)
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

// TestFactoryEntryMatrix is the enterable-pair matrix (AC-015): every
// backend/role pair that is enterable today still launches — Claude and GLM
// as leader (`-f`), Claude, GLM and Codex as lane (today `-f lane`; re-pinned
// to `-l` by M2 and M3) — and the Codex leader stays refused with one line,
// exit 1, and no child, no run record.
func TestFactoryEntryMatrix(t *testing.T) {
	t.Run("claude leader", func(t *testing.T) {
		root := netLeaderFixture(t)
		launch := netDriveLaunch(t, root, netCC, []string{"-f"})
		if launch.err != nil || !launch.launched {
			t.Fatalf("cc -f: launched=%v err=%v", launch.launched, launch.err)
		}
		if launch.env[config.EnvMoaiFactoryWorkers] == "" || launch.env[config.EnvFactoryRole] != "" {
			t.Errorf("cc -f did not launch as a leader: workers=%q role=%q", launch.env[config.EnvMoaiFactoryWorkers], launch.env[config.EnvFactoryRole])
		}
	})
	t.Run("glm leader", func(t *testing.T) {
		root := netLeaderFixture(t)
		launch := netDriveLaunch(t, root, netGLM, []string{"-f"})
		if launch.err != nil || !launch.launched {
			t.Fatalf("glm -f: launched=%v err=%v", launch.launched, launch.err)
		}
		if launch.env[config.EnvMoaiFactoryWorkers] == "" || launch.env[config.EnvFactoryRole] != "" {
			t.Errorf("glm -f did not launch as a leader: workers=%q role=%q", launch.env[config.EnvMoaiFactoryWorkers], launch.env[config.EnvFactoryRole])
		}
	})
	for _, tc := range []struct {
		name    string
		backend string
		entry   func([]string) error
	}{
		{"claude lane", kanban.BackendClaude, netCC},
		{"glm lane", kanban.BackendGLM, netGLM},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := netLaneFixture(t, tc.backend)
			launch := netDriveLaunch(t, root, tc.entry, []string{"-f", "lane"})
			if launch.err != nil || !launch.launched {
				t.Fatalf("%s: launched=%v err=%v", tc.name, launch.launched, launch.err)
			}
			if launch.env[config.EnvFactoryRole] != config.FactoryRoleLane || launch.env[config.EnvMoaiFactoryWorker] == "" {
				t.Errorf("%s did not launch as a lane: role=%q worker=%q", tc.name, launch.env[config.EnvFactoryRole], launch.env[config.EnvMoaiFactoryWorker])
			}
		})
	}
	t.Run("codex lane", func(t *testing.T) {
		lane := netCodexLaneChild(t)
		if lane.env[config.EnvFactoryRole] != config.FactoryRoleLane || lane.env[config.EnvMoaiFactoryWorker] == "" {
			t.Errorf("codex lane child is not a lane: role=%q worker=%q", lane.env[config.EnvFactoryRole], lane.env[config.EnvMoaiFactoryWorker])
		}
		if got := lane.env[config.EnvMoaiKanbanBackend]; got != kanban.BackendGPT {
			t.Errorf("codex lane child %s = %q, want %q", config.EnvMoaiKanbanBackend, got, kanban.BackendGPT)
		}
	})
	t.Run("codex leader refused", func(t *testing.T) {
		root := pinCodexRefusalRoot(t)
		netScrubLaneEnv(t)
		prevDirect, prevSpawn := codexDirectLaunchFn, codexSpawnLaunchFn
		launches := 0
		codexDirectLaunchFn = func(*exec.Cmd) error { launches++; return nil }
		codexSpawnLaunchFn = func(string, string, []string, []string) error { launches++; return nil }
		prevLook := codexLookPath
		codexLookPath = func(string) (string, error) { return "/sentinel/codex", nil }
		t.Cleanup(func() { codexDirectLaunchFn, codexSpawnLaunchFn, codexLookPath = prevDirect, prevSpawn, prevLook })

		_, stderr, err := runCodexCmd(t, "-f")
		code, ok := ResolveExitCode(err)
		if !ok || code != 1 {
			t.Errorf("codex -f: exit code = (%d, %v), want (1, true); err=%v", code, ok, err)
		}
		if lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n"); len(lines) != 1 || lines[0] == "" {
			t.Errorf("codex -f: stderr = %q, want exactly one refusal line", stderr)
		}
		if launches != 0 {
			t.Errorf("codex -f launched %d child(ren); the refusal must start none", launches)
		}
		if record, loadErr := kanban.NewBacklogStore(kanban.BacklogPathForRoot(root)).LoadPure(); loadErr == nil && len(record.Runtime.Runs) != 0 {
			t.Errorf("codex -f recorded %d run(s): %+v", len(record.Runtime.Runs), record.Runtime.Runs)
		}
		if reg := loadFactoryRegistry(factoryRegistryPath(root)); len(reg) != 0 {
			t.Errorf("codex -f wrote lane claims: %v", reg)
		}
	})
}
