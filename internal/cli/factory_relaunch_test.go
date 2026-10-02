package cli

// factory_relaunch_test.go — SPEC-FACTORY-STALE-RUN-HEAL-001 M1: the
// `moai factory relaunch` verb (REQ-SRH-002, -012, -013, -014, -016).

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// runRelaunch executes `moai factory relaunch <args>` and returns stdout,
// stderr and the command error separately.
func runRelaunch(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := newFactoryCommand()
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	var out, errb bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errb)
	cmd.SetArgs(append([]string{"relaunch"}, args...))
	err = cmd.Execute()
	return out.String(), errb.String(), err
}

// captureRelaunchLaunch replaces the launcher-child seam and returns the argv
// (after the program name) of every launch it saw.
func captureRelaunchLaunch(t *testing.T) *[][]string {
	t.Helper()
	var calls [][]string
	prev := factoryRelaunchExecFn
	factoryRelaunchExecFn = func(c *exec.Cmd) error {
		calls = append(calls, append([]string(nil), c.Args[1:]...))
		return nil
	}
	t.Cleanup(func() { factoryRelaunchExecFn = prev })
	return &calls
}

// AC-SRH-001 — the dry-run matrix, and every printed line is accepted by the
// target launcher's own argument classifier.
func TestFactoryRelaunchDryRunMatrix(t *testing.T) {
	root := runOwnerSandbox(t)
	t.Chdir(root)

	cases := []struct {
		name string
		args []string
		want string
	}{
		{"cc pinned lane", []string{"--lane", "lane-3", "--provider", "cc"}, "moai cc -l"},
		{"glm pinned lane", []string{"--lane", "lane-3", "--provider", "glm"}, "moai glm -l"},
		{"cc pinned lane and run", []string{"--lane", "lane-3", "--provider", "cc", "--run", "runA"}, "moai cc -l --factory-run runA"},
		{"cc lane omitted", []string{"--provider", "cc"}, "moai cc -l"},
		{"codex", []string{"--provider", "codex"}, "moai codex -l"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := runRelaunch(t, append(tc.args, "--dry-run")...)
			if err != nil {
				t.Fatalf("dry-run: %v (stderr %q)", err, stderr)
			}
			if stdout != tc.want+"\n" {
				t.Fatalf("stdout = %q, want %q", stdout, tc.want+"\n")
			}
			if stderr != "" {
				t.Errorf("stderr = %q, want empty", stderr)
			}
			// The printed line must be accepted by the launcher it names.
			fields := strings.Fields(strings.TrimSpace(stdout))
			provider, rest := fields[1], fields[2:]
			switch provider {
			case "cc", "glm":
				entry, perr := parseLauncherEntry(rest)
				if perr != nil {
					t.Fatalf("launcher classifier refused %q: %v", stdout, perr)
				}
				if !entry.FactoryEnabled {
					t.Errorf("launcher classifier did not select factory mode for %q: %+v", stdout, entry)
				}
				wantRun := ""
				for i, f := range rest {
					if f == "--factory-run" {
						wantRun = rest[i+1]
					}
				}
				if entry.FactoryRun != wantRun {
					t.Errorf("classifier FactoryRun = %q, want %q", entry.FactoryRun, wantRun)
				}
			case "codex":
				if got, diag := codexFactoryEntryClassify(rest); got != codexFactoryEntryLane || diag != "" {
					t.Errorf("Codex classifier on %q = (%v, %q), want the lane entry", stdout, got, diag)
				}
			default:
				t.Fatalf("unexpected provider %q", provider)
			}
		})
	}
}

// AC-SRH-002 — legacy lane spellings and Codex pins are refused with the
// reason named; a legal pin is accepted in the same test (positive control).
func TestFactoryRelaunchRefusesLegacyAndCodexPins(t *testing.T) {
	root := runOwnerSandbox(t)
	t.Chdir(root)

	if stdout, stderr, err := runRelaunch(t, "--dry-run", "--lane", "lane-3", "--provider", "cc"); err != nil || stdout != "moai cc -l\n" {
		t.Fatalf("positive control failed: stdout=%q stderr=%q err=%v", stdout, stderr, err)
	}
	refused := []struct {
		name string
		args []string
		want []string
	}{
		{"legacy lane", []string{"--lane", "worker-3", "--provider", "cc"}, []string{"lane-3", "legacy"}},
		{"legacy agent lane", []string{"--lane", "agent-2", "--provider", "cc"}, []string{"lane-2", "legacy"}},
		{"codex lane pin", []string{"--provider", "codex", "--lane", "lane-3"}, []string{"Codex", "--lane"}},
		{"codex run pin", []string{"--provider", "codex", "--run", "runA"}, []string{"Codex", "--run"}},
		{"bare role token", []string{"--lane", "lane"}, []string{"lane-<n>"}},
		{"unknown provider", []string{"--provider", "claude"}, []string{"cc", "glm", "codex"}},
		{"malformed run id", []string{"--run", "bad id"}, []string{"not a run id"}},
		{"malformed from-run id", []string{"--from-run", "../x"}, []string{"not a run id"}},
	}
	for _, tc := range refused {
		t.Run(tc.name, func(t *testing.T) {
			captured := captureRelaunchLaunch(t)
			stdout, _, err := runRelaunch(t, append(tc.args, "--dry-run")...)
			if err == nil {
				t.Fatalf("accepted %v, stdout %q", tc.args, stdout)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want empty on a refusal", stdout)
			}
			for _, want := range tc.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not name %q", err, want)
				}
			}
			if len(*captured) != 0 {
				t.Errorf("a refused invocation launched %v", *captured)
			}
		})
	}
}

// runRows reads every runs row and the run.retired event count for runID.
func runRows(t *testing.T, root string) (rows string, events int) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	r, err := db.DB.Query(`SELECT run_id||'|'||status||'|'||lead_pid FROM runs ORDER BY run_id`)
	if err != nil {
		t.Fatalf("read runs: %v", err)
	}
	defer func() { _ = r.Close() }()
	var parts []string
	for r.Next() {
		var s string
		if err := r.Scan(&s); err != nil {
			t.Fatalf("scan run: %v", err)
		}
		parts = append(parts, s)
	}
	if err := db.DB.QueryRow(`SELECT count(*) FROM events`).Scan(&events); err != nil {
		t.Fatalf("count events: %v", err)
	}
	return strings.Join(parts, ";"), events
}

// AC-SRH-003 — --from-run retires a run only when it measures active and its
// owner classifies dead; every other case leaves it untouched, launches
// anyway, and names the outcome.
func TestFactoryRelaunchFromRunRetiresDeadOwnerOnly(t *testing.T) {
	root := runOwnerSandbox(t)
	t.Chdir(root)
	liveStart := homestate.CurrentProcessFingerprint()
	if liveStart == "" {
		t.Skip("this host cannot report its own process-start fingerprint")
	}
	deadPID, deadStart := deadOwnerIdentity(t)
	seedRun(t, root, "runDead", deadPID, deadStart)
	seedRun(t, root, "runLive", os.Getpid(), liveStart)
	seedRun(t, root, "runUnknownOwner", 0, "")
	seedRun(t, root, "runRetired", deadPID, deadStart)
	if _, err := runFactoryCommand(t, "runs", "--retire", "runRetired"); err != nil {
		t.Fatalf("pre-retire runRetired: %v", err)
	}
	_, baseEvents := runRows(t, root)

	cases := []struct {
		fromRun     string
		wantStatus  string
		wantRetired bool
		wantOutcome string
	}{
		{"runDead", "retired", true, "retired"},
		{"runLive", "active", false, "live"},
		{"runUnknownOwner", "active", false, "indeterminate"},
		{"runRetired", "retired", false, "not active"},
		{"runNoSuchRun", "", false, "unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.fromRun, func(t *testing.T) {
			captured := captureRelaunchLaunch(t)
			_, eventsBefore := runRows(t, root)
			stdout, stderr, err := runRelaunch(t, "--provider", "cc", "--from-run", tc.fromRun)
			if err != nil {
				t.Fatalf("relaunch --from-run %s: %v (stderr %q)", tc.fromRun, err, stderr)
			}
			if want := [][]string{{"cc", "-l"}}; !reflect.DeepEqual(*captured, want) {
				t.Errorf("launch argv = %v, want %v", *captured, want)
			}
			if !strings.Contains(stdout+stderr, tc.wantOutcome) {
				t.Errorf("output %q does not name the outcome %q", stdout+stderr, tc.wantOutcome)
			}
			if tc.wantStatus != "" {
				if got := runStatusOf(t, root, tc.fromRun); got != tc.wantStatus {
					t.Errorf("%s status = %q, want %q", tc.fromRun, got, tc.wantStatus)
				}
			}
			_, eventsAfter := runRows(t, root)
			wantDelta := 0
			if tc.wantRetired {
				wantDelta = 1
			}
			if eventsAfter-eventsBefore != wantDelta {
				t.Errorf("event delta = %d, want %d (base %d)", eventsAfter-eventsBefore, wantDelta, baseEvents)
			}
		})
	}
	// The one retirement carries classification and basis.
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var payload string
	if err := db.DB.QueryRow(`SELECT payload_json FROM events WHERE run_id='runDead' AND kind='run.retired'`).Scan(&payload); err != nil {
		t.Fatalf("read run.retired payload: %v", err)
	}
	if !strings.Contains(payload, `"classification":"dead"`) || !strings.Contains(payload, `"basis"`) {
		t.Errorf("run.retired payload %q lacks classification and basis", payload)
	}
}

// AC-004 — --dry-run writes nothing even with a dead-owner --from-run, and the
// lane's clear-policy environment is untouched.
func TestFactoryRelaunchDoesNotMutateRunRecords(t *testing.T) {
	root := runOwnerSandbox(t)
	t.Chdir(root)
	deadPID, deadStart := deadOwnerIdentity(t)
	seedRun(t, root, "runX", deadPID, deadStart)
	t.Setenv(config.EnvFactoryClearPolicy, "relaunch")
	rowsBefore, eventsBefore := runRows(t, root)
	captured := captureRelaunchLaunch(t)

	stdout, stderr, err := runRelaunch(t, "--dry-run", "--provider", "cc", "--from-run", "runX")
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if stdout != "moai cc -l\n" || stderr != "" {
		t.Errorf("stdout=%q stderr=%q, want exactly the launch line", stdout, stderr)
	}
	rowsAfter, eventsAfter := runRows(t, root)
	if rowsBefore != rowsAfter || eventsBefore != eventsAfter {
		t.Errorf("--dry-run changed the run records: [%s %d] -> [%s %d]", rowsBefore, eventsBefore, rowsAfter, eventsAfter)
	}
	if len(*captured) != 0 {
		t.Errorf("--dry-run launched %v", *captured)
	}
	if got := os.Getenv(config.EnvFactoryClearPolicy); got != "relaunch" {
		t.Errorf("clear-policy environment = %q, want untouched", got)
	}
}

// AC-004 — the help text states the distinction from the clear-policy loop.
func TestFactoryRelaunchHelpDistinguishesClearPolicy(t *testing.T) {
	stdout, _, err := runRelaunch(t, "--help")
	if err != nil {
		t.Fatalf("--help: %v", err)
	}
	if !strings.Contains(stdout, "--clear-policy relaunch") {
		t.Errorf("help does not distinguish the verb from --clear-policy relaunch:\n%s", stdout)
	}
}

// AC-006 — every line the shared builder can print is accepted by the verb,
// and the verb answers with exactly the launch line the same arguments imply.
func TestRelaunchCommandRoundTrip(t *testing.T) {
	root := runOwnerSandbox(t)
	t.Chdir(root)
	n := 0
	for _, provider := range []string{"cc", "glm", "codex"} {
		for _, lane := range []string{"", "lane-3"} {
			for _, fromRun := range []string{"", "runX"} {
				for _, run := range []string{"", "runY"} {
					c := factory.RelaunchCommand{Provider: provider, Lane: lane, Run: run, FromRun: fromRun}
					line := c.Line()
					argv := strings.Fields(strings.TrimPrefix(line, "moai factory "))
					stdout, stderr, err := runRelaunch(t, append(argv[1:], "--dry-run")...)
					if err != nil {
						t.Errorf("verb rejected its own line %q: %v (stderr %q)", line, err, stderr)
						continue
					}
					if want := c.LaunchLine() + "\n"; stdout != want {
						t.Errorf("verb on %q printed %q, want %q", line, stdout, want)
					}
					n++
				}
			}
		}
	}
	if n != 24 {
		t.Fatalf("round-trip swept %d combinations, want 24", n)
	}
}

// The verb, without --dry-run, hands the provider's own lane-join argv to the
// launch seam — and surfaces the child's refusal rather than swallowing it.
func TestFactoryRelaunchLaunchesProviderEntry(t *testing.T) {
	root := runOwnerSandbox(t)
	t.Chdir(root)
	captured := captureRelaunchLaunch(t)
	if _, _, err := runRelaunch(t, "--provider", "glm", "--lane", "lane-3", "--run", "runA"); err != nil {
		t.Fatalf("relaunch: %v", err)
	}
	if want := [][]string{{"glm", "-l", "--factory-run", "runA"}}; !reflect.DeepEqual(*captured, want) {
		t.Errorf("launch argv = %v, want %v", *captured, want)
	}
	if _, _, err := runRelaunch(t, "--provider", "codex"); err != nil {
		t.Fatalf("relaunch codex: %v", err)
	}
	if len(*captured) != 2 || !reflect.DeepEqual((*captured)[1], []string{"codex", "-l"}) {
		t.Errorf("launches = %v, want the codex launch [codex -l] second", *captured)
	}
}

// The lane entry `-l` takes no argument, so a supplied --lane cannot pin the
// join: the verb says so once on stderr before launching, leaves stdout and the
// launched argv alone, and stays silent when --lane is not supplied.
func TestFactoryRelaunchLaneIsIgnoredWithNote(t *testing.T) {
	root := runOwnerSandbox(t)
	t.Chdir(root)
	const note = "--lane lane-3 is ignored"
	for _, provider := range []string{"cc", "glm"} {
		captured := captureRelaunchLaunch(t)
		stdout, stderr, err := runRelaunch(t, "--provider", provider, "--lane", "lane-3", "--run", "runA")
		if err != nil {
			t.Fatalf("%s relaunch with --lane: %v", provider, err)
		}
		if got := strings.Count(stderr, note); got != 1 {
			t.Errorf("%s: stderr carries the --lane note %d times, want exactly once:\n%s", provider, got, stderr)
		}
		if !strings.Contains(stderr, "next free lane") || !strings.Contains(stderr, "-l takes no argument") {
			t.Errorf("%s: note does not state the next-free-lane join and the reason:\n%s", provider, stderr)
		}
		if stdout != "" {
			t.Errorf("%s: stdout = %q, want empty", provider, stdout)
		}
		if want := [][]string{{provider, "-l", "--factory-run", "runA"}}; !reflect.DeepEqual(*captured, want) {
			t.Errorf("%s: launch argv = %v, want %v (no lane in the argv)", provider, *captured, want)
		}

		captured = captureRelaunchLaunch(t)
		_, stderr, err = runRelaunch(t, "--provider", provider, "--run", "runA")
		if err != nil {
			t.Fatalf("%s relaunch without --lane: %v", provider, err)
		}
		if strings.Contains(stderr, "--lane") {
			t.Errorf("%s: stderr mentions --lane although none was supplied:\n%s", provider, stderr)
		}
		if want := [][]string{{provider, "-l", "--factory-run", "runA"}}; !reflect.DeepEqual(*captured, want) {
			t.Errorf("%s: launch argv without --lane = %v, want %v", provider, *captured, want)
		}
	}
}

// A failing launcher child is surfaced: its exit status is propagated, and any
// other launch failure is wrapped with the line that was being launched.
func TestFactoryRelaunchSurfacesLaunchFailure(t *testing.T) {
	root := runOwnerSandbox(t)
	t.Chdir(root)
	prev := factoryRelaunchExecFn
	t.Cleanup(func() { factoryRelaunchExecFn = prev })

	factoryRelaunchExecFn = func(*exec.Cmd) error { return exec.Command("/bin/sh", "-c", "exit 3").Run() }
	_, _, err := runRelaunch(t, "--provider", "cc")
	var coded *exitCodeError
	if !errors.As(err, &coded) || coded.code != 3 {
		t.Errorf("child exit 3 surfaced as %v, want an exitCodeError with code 3", err)
	}

	factoryRelaunchExecFn = func(*exec.Cmd) error { return errors.New("cannot start") }
	_, _, err = runRelaunch(t, "--provider", "glm", "--lane", "lane-2")
	if err == nil || !strings.Contains(err.Error(), "moai glm -l") || !strings.Contains(err.Error(), "cannot start") {
		t.Errorf("launch failure surfaced as %v, want the launch line and the cause", err)
	}
}

// An unmeasurable factory state never retires anything: the outcome is named
// and the launch still proceeds.
func TestFactoryRelaunchFromRunUnavailableStateLeavesRunUntouched(t *testing.T) {
	root := runOwnerSandbox(t)
	t.Chdir(root)
	dbPath, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dbPath, []byte("definitely not a sqlite database"), 0o600); err != nil {
		t.Fatal(err)
	}
	captured := captureRelaunchLaunch(t)
	_, stderr, err := runRelaunch(t, "--provider", "cc", "--from-run", "runX")
	if err != nil {
		t.Fatalf("relaunch: %v", err)
	}
	if !strings.Contains(stderr, "unavailable") || !strings.Contains(stderr, "left untouched") {
		t.Errorf("outcome %q does not name the unavailable state", stderr)
	}
	if len(*captured) != 1 {
		t.Errorf("launches = %v, want the launch to proceed", *captured)
	}
}
