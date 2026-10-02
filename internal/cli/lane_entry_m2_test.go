package cli

// lane_entry_m2_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M2 (card t1399): the
// cc/glm lane entry grammar. `-l` / `--lane` joins the running factory as the
// next free lane through the shared lane-slot claim; the removed spellings
// (`-f lane`, `-f lane-<n>`, the explicit lane `--name` under `-f`, the `-l`
// short of `--leader`) are refused at the parser, one line, nothing launched
// and nothing written. The codex rows of AC-003/004/005 joined these tests at
// M3 (lane_entry_m3_test.go carries the codex-only AC-002 and AC-007 tests).
//
// Every test sets or clears every env axis it reads (netScrubLaneEnv) and runs
// through the launch seam (netDriveLaunch), so no session starts and no real
// home state is written.

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// laneVerbs are the launcher verbs whose entry grammar M2 owns.
var laneVerbs = []struct {
	name    string
	backend string
	entry   func([]string) error
}{
	{"cc", kanban.BackendClaude, netCC},
	{"glm", kanban.BackendGLM, netGLM},
}

// removedFormPattern is AC-024's removed-form search over refusal text: a
// refusal names the replacement, never a removed spelling. The `-k` forms are
// left out because the `-l -k` refusal names the token it refuses.
var removedFormPattern = regexp.MustCompile(`(-f|--factory)[ =]lane|-l, --leader|(-f|--factory)[ =](<N>|N\b|[0-9])`)

// m2Refusal drives one launch that must be refused and asserts the shared
// refusal contract: one diagnostic line, exit 1, nothing launched, no lane
// claim, no new run row, and no transient settings file.
func m2Refusal(t *testing.T, backend string, entry func([]string) error, args []string) error {
	t.Helper()
	root := netLaneFixture(t, backend)
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	runsBefore := runRowCount(t, root)

	launch := netDriveLaunch(t, root, entry, args)
	if launch.launched {
		t.Errorf("%v launched a session; a refusal launches nothing", args)
	}
	if launch.err == nil {
		t.Fatalf("%v was accepted, want a refusal", args)
	}
	msg := launch.err.Error()
	if msg == "" || strings.Contains(msg, "\n") {
		t.Errorf("%v refusal = %q, want exactly one non-empty line", args, msg)
	}
	if code, ok := ResolveExitCode(launch.err); ok && code != 1 {
		t.Errorf("%v exit code = %d, want 1", args, code)
	}
	if loc := removedFormPattern.FindString(msg); loc != "" {
		t.Errorf("%v refusal %q names the removed form %q", args, msg, loc)
	}
	if reg := loadFactoryRegistry(factoryRegistryPath(root)); len(reg) != 0 {
		t.Errorf("%v wrote lane claims: %v", args, reg)
	}
	if got := runRowCount(t, root); got != runsBefore {
		t.Errorf("%v changed the run rows: %d -> %d", args, runsBefore, got)
	}
	if leftovers, _ := os.ReadDir(tmp); len(leftovers) != 0 {
		t.Errorf("%v wrote %d settings/temp file(s) under TMPDIR", args, len(leftovers))
	}
	return launch.err
}

func requireTokens(t *testing.T, args []string, msg string, tokens ...string) {
	t.Helper()
	for _, token := range tokens {
		if !strings.Contains(msg, token) {
			t.Errorf("%v refusal %q lacks %q", args, msg, token)
		}
	}
}

// TestLaneEntryJoinsNextFreeSlot — AC-001: `-l` and `--lane` take the next free
// lane label through the shared claim — one past the highest LIVE claim, dead
// claims pruned, never backfilling — and the claim row exists afterwards.
func TestLaneEntryJoinsNextFreeSlot(t *testing.T) {
	type fixture struct {
		name   string
		live   map[string]int // label -> pid held by a live process
		dead   map[string]int // label -> pid no process holds
		want   string
		spells []string // lane entry spellings exercised (F1 carries both)
	}
	fixtures := []fixture{
		{name: "F1", live: map[string]int{"lane-1": 41001}, want: "lane-2", spells: []string{"-l", "--lane"}},
		{name: "F2", live: map[string]int{"lane-1": 41001, "lane-3": 41003}, want: "lane-4", spells: []string{"-l"}},
		{name: "F3", dead: map[string]int{"lane-1": 41099}, want: "lane-1", spells: []string{"-l"}},
		{name: "F4", live: map[string]int{"lane-2": 41002}, dead: map[string]int{"lane-5": 41098}, want: "lane-3", spells: []string{"-l"}},
	}
	for _, fx := range fixtures {
		for _, verb := range laneVerbs {
			for _, spelling := range fx.spells {
				t.Run(fx.name+"_"+verb.name+"_"+spelling, func(t *testing.T) {
					root := netLaneFixture(t, verb.backend)
					reg := map[string]factoryLaneEntry{}
					livePIDs := map[int]bool{}
					for label, pid := range fx.live {
						reg[label] = factoryLaneEntry{PID: pid}
						livePIDs[pid] = true
					}
					for label, pid := range fx.dead {
						reg[label] = factoryLaneEntry{PID: pid}
					}
					if err := saveFactoryRegistry(factoryRegistryPath(root), reg); err != nil {
						t.Fatalf("seed registry: %v", err)
					}
					probe := factoryProcessAlive
					factoryProcessAlive = func(pid int) bool { return livePIDs[pid] }
					t.Cleanup(func() { factoryProcessAlive = probe })

					launch := netDriveLaunch(t, root, verb.entry, []string{spelling})
					if launch.err != nil || !launch.launched {
						t.Fatalf("%s %s: launched=%v err=%v", verb.name, spelling, launch.launched, launch.err)
					}
					for key, want := range map[string]string{
						config.EnvFactoryRole:       config.FactoryRoleLane,
						config.EnvMoaiFactoryWorker: fx.want,
						config.EnvFactoryBackend:    verb.backend,
					} {
						if got := launch.env[key]; got != want {
							t.Errorf("%s at launch = %q, want %q", key, got, want)
						}
					}
					if got := netNamedArg(launch.args); got != fx.want {
						t.Errorf("the backend argv names the session %q, want %q (argv %v)", got, fx.want, launch.args)
					}
					// The claim row, not only the env label.
					after := loadFactoryRegistry(factoryRegistryPath(root))
					if entry, ok := after[fx.want]; !ok || entry.PID != os.Getpid() {
						t.Errorf("registry[%s] = (%+v, %v), want a claim held by pid %d", fx.want, entry, ok, os.Getpid())
					}
					for label, pid := range fx.live {
						if entry, ok := after[label]; !ok || entry.PID != pid {
							t.Errorf("live row %s = (%+v, %v), want it unchanged under pid %d", label, entry, ok, pid)
						}
					}
					for label := range fx.dead {
						if _, ok := after[label]; ok && label != fx.want {
							t.Errorf("dead row %s survived the claim (a dead claim is pruned)", label)
						}
					}
				})
			}
		}
	}
}

// laneRefusalShapes are AC-003's fifteen shapes: eight with `-l`, seven with
// `--lane`. kind selects the contract a refusal line must meet.
var laneRefusalShapes = []struct {
	args []string
	kind string // "argument", "tokens", "name", "retired"
}{
	{[]string{"-l", "lane-2"}, "argument"},
	{[]string{"-l", "3"}, "argument"},
	{[]string{"-l=lane-2"}, "argument"},
	{[]string{"-l", "leader-2"}, "argument"},
	{[]string{"-l", "-f"}, "tokens"},
	{[]string{"-f", "-l"}, "tokens"},
	{[]string{"-l", "-k"}, "retired"},
	{[]string{"-l", "--name", "lane-2"}, "name"},
	{[]string{"--lane", "lane-2"}, "argument"},
	{[]string{"--lane=lane-2"}, "argument"},
	{[]string{"--lane", "3"}, "argument"},
	{[]string{"--lane", "leader-2"}, "argument"},
	{[]string{"--lane", "-f"}, "tokens"},
	{[]string{"--lane", "-k"}, "retired"},
	{[]string{"--lane", "--name", "lane-2"}, "name"},
}

// requireLaneRefusalTokens asserts the line a refused lane shape carries: an
// argument names `-l` as taking none and `--leader <name>` as the selector, a
// combination names the one-entry-token rule, a --name names the role clash. A
// `-k` beside the lane entry earns the retired-entry line instead (M5a: the
// retired spelling is refused before the one-entry-token rule runs).
func requireLaneRefusalTokens(t *testing.T, args []string, kind, msg string) {
	t.Helper()
	switch kind {
	case "retired":
		requireTokens(t, args, msg, "retired", "-f", "-l")
	case "argument":
		requireTokens(t, args, msg, "-l", "no argument", "--leader <name>")
	case "tokens":
		requireTokens(t, args, msg, "entry token")
	case "name":
		requireTokens(t, args, msg, "-l", "already names the role")
	}
}

// TestLaneEntryRefusals — AC-003: `-l` / `--lane` with any argument, or with
// another entry token, or with an operator --name, is refused with one line —
// on cc and glm (M2) and on codex (M3).
func TestLaneEntryRefusals(t *testing.T) {
	for _, verb := range laneVerbs {
		for _, shape := range laneRefusalShapes {
			t.Run(verb.name+"_"+strings.Join(shape.args, "_"), func(t *testing.T) {
				err := m2Refusal(t, verb.backend, verb.entry, shape.args)
				if err == nil {
					return
				}
				requireLaneRefusalTokens(t, shape.args, shape.kind, err.Error())
			})
		}
	}
	for _, shape := range laneRefusalShapes {
		t.Run("codex_"+strings.Join(shape.args, "_"), func(t *testing.T) {
			requireLaneRefusalTokens(t, shape.args, shape.kind, m3CodexRefusal(t, shape.args))
		})
	}
}

// TestLaneEntryEnvParity — AC-004 (cc, glm, codex): a `-l` / `--lane` launch publishes
// the marker set the golden captured from today's `-f lane`, label aside, and no
// name outside the constants of envkeys.go. The codex row joined at M3.
func TestLaneEntryEnvParity(t *testing.T) {
	raw, err := os.ReadFile("../config/envkeys.go")
	if err != nil {
		t.Fatalf("read envkeys.go: %v", err)
	}
	known := map[string]bool{}
	for _, m := range regexp.MustCompile(`"(MOAI_[A-Z0-9_]+)"`).FindAllStringSubmatch(string(raw), -1) {
		known[m[1]] = true
	}
	golden := readLaneGolden(t)

	for _, verb := range laneVerbs {
		for _, spelling := range []string{"-l", "--lane"} {
			root := netLaneFixture(t, verb.backend)
			launch := netDriveLaunch(t, root, verb.entry, []string{spelling})
			if launch.err != nil || !launch.launched {
				t.Errorf("%s %s: launched=%v err=%v", verb.name, spelling, launch.launched, launch.err)
				continue
			}
			got := laneMarkerEnv(launch.env)
			delete(got, retiredLaneLabelMarker) // REQ-003's one exception (the Codex lane child stamp)
			want := map[string]string{}
			for k, v := range golden[verb.name] {
				if k != retiredLaneLabelMarker {
					want[k] = v
				}
			}
			if formatMarkerRow(got) != formatMarkerRow(want) {
				t.Errorf("%s %s markers differ from the `-f lane` golden\n got: %s\nwant: %s", verb.name, spelling, formatMarkerRow(got), formatMarkerRow(want))
			}
			for key := range got {
				if !known[key] {
					t.Errorf("%s %s published %s, which is no constant of internal/config/envkeys.go", verb.name, spelling, key)
				}
			}
		}
	}

	// The codex row (M3): the per-card child environment of the relaunch loop
	// `-l` / `--lane` starts. The golden row carries MOAI_KANBAN_LABEL (today's
	// child stamps it); the comparison excludes exactly that key.
	for _, spelling := range []string{"-l", "--lane"} {
		got := laneMarkerEnv(netCodexLaneChildFor(t, spelling).env)
		delete(got, retiredLaneLabelMarker)
		want := map[string]string{}
		for k, v := range golden["codex"] {
			if k != retiredLaneLabelMarker {
				want[k] = v
			}
		}
		if formatMarkerRow(got) != formatMarkerRow(want) {
			t.Errorf("codex %s markers differ from the `-f lane` golden\n got: %s\nwant: %s", spelling, formatMarkerRow(got), formatMarkerRow(want))
		}
		for key := range got {
			if !known[key] {
				t.Errorf("codex %s published %s, which is no constant of internal/config/envkeys.go", spelling, key)
			}
		}
	}
}

// TestLaneEntryComposesWithLaneOptions — AC-005: the lane options compose with
// `-l` / `--lane` as they composed with `-f lane`, and the leader selector keeps
// its long form.
func TestLaneEntryComposesWithLaneOptions(t *testing.T) {
	// Parse-level rows: the selections reach the entry parse.
	for _, tc := range []struct {
		args []string
		pass func(entry launcherEntryParse) string // "" = ok, otherwise the complaint
	}{
		{[]string{"-l", "--leader", "leader-2"}, func(e launcherEntryParse) string { return wantString("FactoryLead", e.FactoryLead, "leader-2") }},
		{[]string{"--lane", "--leader=leader-2"}, func(e launcherEntryParse) string { return wantString("FactoryLead", e.FactoryLead, "leader-2") }},
		{[]string{"-l", "--clear-policy", config.FactoryClearPolicyEach}, func(e launcherEntryParse) string {
			return wantString("ClearPolicy", e.ClearPolicy, config.FactoryClearPolicyEach)
		}},
		{[]string{"-l", "--no-auto-dispatch"}, func(e launcherEntryParse) string {
			if !e.AutoDispatchManual {
				return "AutoDispatchManual = false, want true"
			}
			return ""
		}},
		{[]string{"-l", "--factory-run", "runx0001"}, func(e launcherEntryParse) string { return wantString("FactoryRun", e.FactoryRun, "runx0001") }},
		{[]string{"-l", "-w", "feat-x"}, func(e launcherEntryParse) string {
			if !containsFlag(e.Rest, "-w") || !containsFlag(e.Rest, "feat-x") {
				return "Rest = " + strings.Join(e.Rest, " ") + ", want the -w pair preserved"
			}
			return ""
		}},
		{[]string{"-l", "--", "--print"}, func(e launcherEntryParse) string {
			if !containsFlag(e.Rest, "--print") {
				return "Rest = " + strings.Join(e.Rest, " ") + ", want the passthrough token kept"
			}
			return ""
		}},
	} {
		entry, err := parseLauncherEntry(tc.args)
		if err != nil {
			t.Errorf("parseLauncherEntry(%v): %v", tc.args, err)
			continue
		}
		if !entry.FactoryEnabled || !entry.FactoryAutoNumber {
			t.Errorf("parseLauncherEntry(%v): FactoryEnabled=%v FactoryAutoNumber=%v, want a lane join with an automatic number", tc.args, entry.FactoryEnabled, entry.FactoryAutoNumber)
		}
		if msg := tc.pass(entry); msg != "" {
			t.Errorf("parseLauncherEntry(%v): %s", tc.args, msg)
		}
	}

	// Launch-level rows: the selections reach the launched lane.
	for _, verb := range laneVerbs {
		root := netLaneFixture(t, verb.backend)
		launch := netDriveLaunch(t, root, verb.entry, []string{"-l", "--clear-policy", config.FactoryClearPolicyWhenFull, "--no-auto-dispatch", "-p", "work"})
		if launch.err != nil || !launch.launched {
			t.Errorf("%s -l with options: launched=%v err=%v", verb.name, launch.launched, launch.err)
			continue
		}
		if got := launch.env[config.EnvFactoryClearPolicy]; got != config.FactoryClearPolicyWhenFull {
			t.Errorf("%s %s = %q, want %q", verb.name, config.EnvFactoryClearPolicy, got, config.FactoryClearPolicyWhenFull)
		}
		if got := launch.env[config.EnvFactoryAutoDispatch]; got != config.FactoryDispatchManual {
			t.Errorf("%s %s = %q, want %q", verb.name, config.EnvFactoryAutoDispatch, got, config.FactoryDispatchManual)
		}

		// A pass-through region is forwarded verbatim and never inspected: the
		// launch is still the lane the entry token named (the desugared --name
		// sits ahead of the marker), and the child still receives the tokens.
		root = netLaneFixture(t, verb.backend)
		launch = netDriveLaunch(t, root, verb.entry, []string{"-l", "--", "--print"})
		if launch.err != nil || !launch.launched {
			t.Errorf("%s -l -- --print: launched=%v err=%v", verb.name, launch.launched, launch.err)
			continue
		}
		if launch.env[config.EnvFactoryRole] != config.FactoryRoleLane || launch.env[config.EnvMoaiFactoryWorker] == "" {
			t.Errorf("%s -l -- --print did not launch as a lane: role=%q worker=%q", verb.name, launch.env[config.EnvFactoryRole], launch.env[config.EnvMoaiFactoryWorker])
		}
		if !containsFlag(launch.args, "--print") {
			t.Errorf("%s -l -- --print: the pass-through token was not forwarded: %v", verb.name, launch.args)
		}
	}
}

func wantString(field, got, want string) string {
	if got != want {
		return field + " = " + got + ", want " + want
	}
	return ""
}

// TestLeaderSelectorRefusedWithoutLaneEntry — AC-005: `--leader` composes with
// the lane entry only; without it the launch is refused, naming `-l` and
// `--lane`, on cc and glm (M2) and on codex (M3).
func TestLeaderSelectorRefusedWithoutLaneEntry(t *testing.T) {
	for _, verb := range laneVerbs {
		for _, args := range [][]string{
			{"--leader", "leader-2"},
			{"--leader=leader-2"},
			{"-f", "--leader", "leader-2"},
			{"-f", "--leader=leader-2"},
		} {
			t.Run(verb.name+"_"+strings.Join(args, "_"), func(t *testing.T) {
				if err := m2Refusal(t, verb.backend, verb.entry, args); err != nil {
					requireTokens(t, args, err.Error(), "-l", "--lane")
				}
			})
		}
	}
	for _, args := range [][]string{
		{"--leader", "leader-2"},
		{"-f", "--leader", "leader-2"},
	} {
		t.Run("codex_"+strings.Join(args, "_"), func(t *testing.T) {
			requireTokens(t, args, m3CodexRefusal(t, args), "-l", "--lane")
		})
	}
}

// TestFactoryLaneSpellingsRefused — AC-006: on cc and glm the removed lane
// spellings are refused naming `-l`, and a non-lane --name keeps launching a
// named leader.
func TestFactoryLaneSpellingsRefused(t *testing.T) {
	for _, verb := range laneVerbs {
		for _, args := range [][]string{
			{"-f", "lane"},
			{"-f", "lane-2"},
			{"--factory", "lane"},
			{"-f=lane"},
			{"--factory=lane-2"},
			{"-f", "--name", "lane-2"},
			{"-f", "-n", "lane-2"},
			{"--factory", "--name=lane-2"},
		} {
			t.Run(verb.name+"_"+strings.Join(args, "_"), func(t *testing.T) {
				if err := m2Refusal(t, verb.backend, verb.entry, args); err != nil {
					requireTokens(t, args, err.Error(), "-l")
				}
			})
		}
		t.Run(verb.name+"_named_leader_launches", func(t *testing.T) {
			root := netLeaderFixture(t)
			launch := netDriveLaunch(t, root, verb.entry, []string{"-f", "--name", "leader-r7"})
			if launch.err != nil || !launch.launched {
				t.Fatalf("%s -f --name leader-r7: launched=%v err=%v", verb.name, launch.launched, launch.err)
			}
			if launch.env[config.EnvFactoryRole] != "" || launch.env[config.EnvMoaiFactoryWorkers] == "" {
				t.Errorf("%s -f --name leader-r7 did not launch as a leader: role=%q workers=%q", verb.name, launch.env[config.EnvFactoryRole], launch.env[config.EnvMoaiFactoryWorkers])
			}
		})
	}
}

// TestFactoryCountShapeStillRefused — AC-008: `-f <N>` stays refused; the line
// names the bare leader form and the lane entry, and no removed form.
func TestFactoryCountShapeStillRefused(t *testing.T) {
	for _, verb := range laneVerbs {
		t.Run(verb.name, func(t *testing.T) {
			args := []string{"-f", "3"}
			if err := m2Refusal(t, verb.backend, verb.entry, args); err != nil {
				requireTokens(t, args, err.Error(), "-f", "-l")
			}
		})
	}
}

// readLaneGolden loads the committed `-f lane` marker golden (AC-004).
func readLaneGolden(t *testing.T) map[string]map[string]string {
	t.Helper()
	raw, err := os.ReadFile(laneGoldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var golden map[string]map[string]string
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("parse golden: %v", err)
	}
	return golden
}
