// factory_m4_test.go — SPEC-FACTORY-SELF-DISPATCH-001 M4 AC tests (card
// t1240): the cc/glm lane launcher (marker + lane label + backend stamp, and
// the child's working directory), the git-working-tree requirement, the
// no-remote lane cycle, and the no-headless-engine walk. AC-SD-002, -005,
// -006, -007. The REQ-SD-017 cli arm lands in this file with the launcher
// implementation (it classifies the captured launch environment, which needs
// the launcher's marker stamp to exist first).
//
// Every fixture is built under t.TempDir() with an isolated git config and
// MOAI_HOME sandboxed away (§B of acceptance.md); ./internal/cli runs only
// through the anchored -run selectors naming one of these tests.
package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// sdScrubLauncherEnv neutralizes every MOAI_* variable the launcher stamps or
// the lane predicates read, so a lane session's inherited factory environment
// cannot leak into a capture (§B: the fixture environment is the test's own).
func sdScrubLauncherEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		config.EnvMoaiKanban, config.EnvMoaiKanbanID, config.EnvMoaiKanbanLabel,
		config.EnvMoaiKanbanLeadAddr, config.EnvMoaiKanbanSettingsInjected,
		config.EnvFactoryRole, config.EnvMoaiFactoryWorker, config.EnvMoaiFactoryWorkers,
		config.EnvMoaiKanbanBackend, config.EnvMoaiKanbanCard,
	} {
		t.Setenv(k, "")
	}
}

// sdRecordLeaderRun inserts the active factory run a `-f lane` join resolves
// (the leader's recording, as fixture).
func sdRecordLeaderRun(t *testing.T, root, runID, backend string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{
		RunID: runID, LeadSessionID: "leader", Backend: backend, ManifestJSON: "{}",
	}); err != nil {
		t.Fatalf("record leader run: %v", err)
	}
}

// sdLaunchCapture is one substituted engine launch: the binary and argv the
// launcher would hand the engine, the environment the launcher stamped (the
// exec'd child inherits the process environment), and the child's working
// directory.
type sdLaunchCapture struct {
	binary string
	argv   []string
	env    map[string]string
	cwd    string
}

// sdDriveLaneLaunch drives one cc/glm lane launch with the engine launch
// substituted (unifiedLaunchFunc), and returns what the launcher stamped. The
// optional during callbacks run inside the substitution, while the stamped
// environment is still live — a classification arm needs that window, because
// the enter helpers restore the environment on return.
func sdDriveLaneLaunch(t *testing.T, root string, entry func([]string) error, during ...func()) sdLaunchCapture {
	t.Helper()
	sdScrubLauncherEnv(t)
	captured := sdLaunchCapture{env: map[string]string{}}
	prevLaunch := unifiedLaunchFunc
	unifiedLaunchFunc = func(_ string, mode string, args []string) error {
		// Both cc and glm hand the claude engine its mode through the
		// environment; the mode is recorded verbatim on the binary field for
		// the headless walk.
		captured.binary = mode
		captured.argv = append([]string(nil), args...)
		for _, kv := range os.Environ() {
			name, value, found := strings.Cut(kv, "=")
			if found {
				captured.env[name] = value
			}
		}
		captured.cwd, _ = os.Getwd()
		for _, fn := range during {
			fn()
		}
		return nil
	}
	t.Cleanup(func() { unifiedLaunchFunc = prevLaunch })
	prevFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return root, nil }
	t.Cleanup(func() { findProjectRootFn = prevFn })
	prevDeps := deps
	deps = nil
	t.Cleanup(func() { deps = prevDeps })

	if err := entry([]string{"-f", "lane"}); err != nil {
		t.Fatalf("lane launch: %v", err)
	}
	return captured
}

// sdCCEntry and sdGLMEntry adapt the two launcher entries to one shape.
func sdCCEntry(args []string) error  { return runCC(ccCmd, args) }
func sdGLMEntry(args []string) error { return runGLM(glmCmd, args) }

// sdNextFreeLaneLabel is the lane label the launcher should claim on the
// fixture roster — computed through the same producer the launcher uses, so
// the assertion follows the roster rather than a guessed number.
func sdNextFreeLaneLabel(t *testing.T, root string) string {
	t.Helper()
	next := kanban.NextFactoryLaneNumber(loadFactoryRegistry(factoryRegistryPath(root)), factoryProcessAlive)
	return kanban.FactoryLaneLabel(next)
}

// AC-SD-002 — cc/glm lane launch stamps the marker, the lane label, and the
// backend value; the child's working directory is the parent checkout.
func TestSD_AC002_LaneLaunchStampsMarkerAndLabel(t *testing.T) {
	for _, tc := range []struct {
		name    string
		backend string
		entry   func([]string) error
	}{
		{"cc", kanban.BackendClaude, sdCCEntry},
		{"glm", kanban.BackendGLM, sdGLMEntry},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, _ := fcFixture(t)
			sdRecordLeaderRun(t, root, fcRun, kanban.BackendClaude)
			t.Chdir(root)
			t.Setenv(config.EnvClaudeProjectDir, root)
			wantLabel := sdNextFreeLaneLabel(t, root)

			captured := sdDriveLaneLaunch(t, root, tc.entry)

			if got := captured.env[config.EnvFactoryRole]; got != config.FactoryRoleLane {
				t.Errorf("child env %s = %q, want the value constant %q (the marker name and value travel only through the internal/config constants)",
					config.EnvFactoryRole, got, config.FactoryRoleLane)
			}
			if got := captured.env[config.EnvMoaiFactoryWorker]; got != wantLabel {
				t.Errorf("child env %s = %q, want the next free lane label %q",
					config.EnvMoaiFactoryWorker, got, wantLabel)
			}
			if got := captured.env[config.EnvMoaiKanbanBackend]; got != tc.backend {
				t.Errorf("child env %s = %q, want %q (the backend value the launch already exports)",
					config.EnvMoaiKanbanBackend, got, tc.backend)
			}
			if captured.cwd != root {
				t.Errorf("child working directory = %q, want the parent checkout %q", captured.cwd, root)
			}
		})
	}
}

// AC-SD-005 — a lane launch outside a git working tree exits non-zero with
// one line naming the git requirement, and writes no record.
func TestSD_AC005_LaneLaunchRequiresGitRepo(t *testing.T) {
	for _, tc := range []struct {
		name  string
		entry func([]string) error
	}{
		{"cc", sdCCEntry},
		{"glm", sdGLMEntry},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir() // not a git working tree
			t.Setenv(config.EnvClaudeProjectDir, root)
			t.Chdir(root)

			prevLaunch := unifiedLaunchFunc
			unifiedLaunchFunc = func(string, string, []string) error {
				t.Error("the engine launch ran; the git requirement must refuse before any launch")
				return nil
			}
			t.Cleanup(func() { unifiedLaunchFunc = prevLaunch })
			prevFn := findProjectRootFn
			findProjectRootFn = func() (string, error) { return root, nil }
			t.Cleanup(func() { findProjectRootFn = prevFn })
			prevDeps := deps
			deps = nil
			t.Cleanup(func() { deps = prevDeps })

			err := tc.entry([]string{"-f", "lane"})
			if err == nil {
				t.Fatal("lane launch outside a git working tree succeeded; want the git-requirement refusal")
			}
			msg := err.Error()
			if strings.Contains(msg, "\n") {
				t.Errorf("refusal is more than one line:\n%s", msg)
			}
			if !strings.Contains(msg, "git") {
				t.Errorf("refusal does not name the git requirement: %s", msg)
			}
			// The lane registry, the factory database, and the queue file are
			// all absent: the refusal writes no state file (the parse step may
			// scaffold empty directories; no file under them may exist).
			_ = filepath.WalkDir(filepath.Join(root, ".moai"), func(path string, d os.DirEntry, err error) error {
				if err == nil && !d.IsDir() {
					t.Errorf("the refused launch wrote %s", path)
				}
				return nil
			})
		})
	}
}

// sdGitRecorder installs a recording git wrapper first on PATH: every git
// invocation the cycle makes is appended to a log file, then run for real
// (the TestFR_AC018 recorder). Returns the log path.
func sdGitRecorder(t *testing.T) string {
	t.Helper()
	real, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "git-calls.log")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '" + logPath + "'\nexec '" + real + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return logPath
}

// AC-SD-006 — one card carried by a Claude lane from `next` to `complete`
// reaches `merged-local`, and the recorded git command log contains no fetch
// and no push (the §B fixture: parent on main, develop provisioned at
// .moai/worktrees/develop, no remote anywhere).
func TestSD_AC006_LaneCycleWithoutRemote(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the recording wrapper is a POSIX shell script (the TestFR_AC018 family)")
	}
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStateQueued)
	sdRegisterLane(t, root, "lane-1")
	// The §B fixture layout: the integration branch is develop, configured as
	// the project's integration branch and checked out in the provisioned
	// integration worktree; the repository has no remote.
	fcGit(t, root, "branch", "develop")
	integWT := filepath.Join(root, sessionWorktreeSubdir, "develop")
	fcGit(t, root, "worktree", "add", "-q", integWT, "develop")
	sdGitFlowDevelop(t, root)
	if remotes := fcGit(t, root, "remote"); remotes != "" {
		t.Fatalf("fixture carries a remote: %s", remotes)
	}

	logPath := sdGitRecorder(t)

	sdLaneEnv(t, "lane-1", "")
	t.Chdir(root)

	// next: the lane leases the card and gains its card worktree.
	if _, _, err := runFactory(t, "next", "--run", fcRun); err != nil {
		t.Fatalf("next: %v", err)
	}
	card := fcCard(t, root, "t1")
	if card.WorktreePath == "" {
		t.Fatal("next recorded no card worktree")
	}
	// The lane's own work: one commit on the card branch.
	if err := os.WriteFile(filepath.Join(card.WorktreePath, "done.txt"), []byte("work\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, card.WorktreePath, "add", "-A")
	fcGit(t, card.WorktreePath, "commit", "-q", "-m", "card work")
	runSHA := fcGit(t, card.WorktreePath, "rev-parse", "HEAD")

	// A fresh queue-promoted lease carries no stage: the T4b resume admits it
	// into the run phase (Class B — this fixture runs no plan phase).
	if _, _, err := runFactory(t, "stage", "t1", "run", "--run", fcRun); err != nil {
		t.Fatalf("stage run: %v", err)
	}

	// sync: the run-phase commit is the T10 evidence.
	if _, _, err := runFactory(t, "stage", "t1", "sync", runSHA, "--run", fcRun); err != nil {
		t.Fatalf("stage sync: %v", err)
	}
	// sync-audit: the entry artifact is the committed work; the PASS verdict
	// is written into the worktree's report directory uncommitted and names
	// the recorded evidence commit (the E-VERDICT prefix rule).
	if _, _, err := runFactory(t, "stage", "t1", "sync-audit", runSHA+":done.txt", "--run", fcRun); err != nil {
		t.Fatalf("stage sync-audit: %v", err)
	}
	verdictAbs := filepath.Join(card.WorktreePath, ".moai", "reports", "t1", "sync-audit.md")
	if err := os.MkdirAll(filepath.Dir(verdictAbs), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(verdictAbs, []byte("verdict: PASS\naudited_sha: "+runSHA+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// merge-ready: the T13 PASS verdict reads the committed file.
	if _, _, err := runFactory(t, "stage", "t1", "merge-ready", "--run", fcRun); err != nil {
		t.Fatalf("stage merge-ready: %v", err)
	}

	// The lane holds the integration window on develop, then completes — the
	// merge itself runs inside complete, inside the integration worktree.
	sdHoldWindow(t, root, "sess-lane-1", "lane-1", "develop", kanban.BranchSourceConfig, integWT, "t1")
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
		t.Fatalf("complete: %v", err)
	}

	if c := fcCard(t, root, "t1"); c.State != homestate.CardMergedLocal {
		t.Fatalf("t1 = %s, want merged-local", c.State)
	}

	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("recorder saw no git call at all (positive control): %v", err)
	}
	calls := strings.Split(strings.TrimSpace(string(raw)), "\n")
	sawMerge, sawWorktree := false, false
	for _, call := range calls {
		fields := strings.Fields(call)
		if slices.Contains(fields, "fetch") {
			t.Errorf("the lane cycle ran git fetch: %q", call)
		}
		if slices.Contains(fields, "push") {
			t.Errorf("the lane cycle ran git push: %q", call)
		}
		sawMerge = sawMerge || slices.Contains(fields, "merge")
		sawWorktree = sawWorktree || slices.Contains(fields, "worktree")
	}
	// Positive control: the recorder saw the cycle's own git work (the merge
	// and the card worktree creation), so an empty or near-empty log cannot
	// pass as "no fetch and no push".
	if !sawMerge || !sawWorktree {
		t.Errorf("positive control: recorder did not see the cycle's git work (merge=%v worktree=%v):\n%s",
			sawMerge, sawWorktree, raw)
	}
}

// sdLaunchPathArgv is one entry of AC-SD-007's enumeration: a launch path
// this SPEC adds, with the engine binary and the captured argv it hands it.
type sdLaunchPathArgv struct {
	name   string
	binary string
	argv   []string
}

// AC-SD-007 — no launch path spawns a headless engine. The walk reads the
// enumeration below; M5 (the codex relaunch) appends its entry to the slice
// without restructuring the walk.
func TestSD_AC007_NoHeadlessEngineArgv(t *testing.T) {
	// The capture fixtures: one factory run with a leader, one lane launch
	// per entry.
	capture := func(t *testing.T, name string, entry func([]string) error) sdLaunchPathArgv {
		t.Helper()
		root, _ := fcFixture(t)
		sdRecordLeaderRun(t, root, fcRun, kanban.BackendClaude)
		t.Chdir(root)
		captured := sdDriveLaneLaunch(t, root, entry)
		return sdLaunchPathArgv{name: name, binary: captured.binary, argv: captured.argv}
	}

	// M4 scope: the launch paths that exist are the cc and glm lane launches
	// (both hand the claude engine its mode). M5 appends the codex relaunch
	// entry here — the walk below already knows the codex arm.
	paths := []sdLaunchPathArgv{
		capture(t, "cc -f lane", sdCCEntry),
		capture(t, "glm -f lane", sdGLMEntry),
	}
	if len(paths) < 2 {
		t.Fatalf("the enumeration lost the M4 lane launches: %d entries", len(paths))
	}

	for _, p := range paths {
		switch p.binary {
		case "claude", "glm":
			// The claude engine's headless form is -p/--print.
			if slices.Contains(p.argv, "-p") || slices.Contains(p.argv, "--print") {
				t.Errorf("%s: claude argv carries a headless flag: %v", p.name, p.argv)
			}
		case "codex":
			// The codex engine's headless form is the `exec` subcommand; a
			// lane launch hands the interactive session (M5's argv lands in
			// this arm).
			if slices.Contains(p.argv, "exec") {
				t.Errorf("%s: codex argv carries the exec subcommand: %v", p.name, p.argv)
			}
		default:
			t.Errorf("%s: engine binary %q is not a known interactive engine", p.name, p.binary)
		}
	}
}

// sdReadSource reads one production file for the AC-SD-017 source scan.
func sdReadSource(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

// AC-SD-017 (cli arm) — the environment captured from a lane launch arms the
// contract guard, and the production files that stamp or compare the marker
// do it only through the internal/config constants.
func TestSD_AC017_StampedMarkerArmsContractGuard(t *testing.T) {
	root, _ := fcFixture(t)
	sdRecordLeaderRun(t, root, fcRun, kanban.BackendClaude)
	t.Chdir(root)
	t.Setenv(config.EnvClaudeProjectDir, root)

	// The classification runs inside the launch substitution: that is the one
	// window where the captured environment is the live process environment,
	// exactly what the guard reads at the tool boundary.
	var decision, reason string
	captured := sdDriveLaneLaunch(t, root, sdCCEntry, func() {
		decision, reason = hook.CheckContractSignClassify("moai contract sign SPEC-X --signer llm")
	})
	t.Logf("captured %s=%q %s=%q %s=%q",
		config.EnvFactoryRole, captured.env[config.EnvFactoryRole],
		config.EnvMoaiFactoryWorker, captured.env[config.EnvMoaiFactoryWorker],
		config.EnvMoaiKanbanBackend, captured.env[config.EnvMoaiKanbanBackend])
	if decision != hook.DecisionDeny {
		t.Fatalf("contract sign under the lane launch environment: decision = %q (reason %q), want deny", decision, reason)
	}
	if !strings.Contains(reason, "CONTRACT_SIGN_AGENT_VIOLATION:") {
		t.Errorf("deny reason = %q, want the %s sentinel", reason, "CONTRACT_SIGN_AGENT_VIOLATION:")
	}

	// The source scan: the production files that stamp or compare the marker
	// name it and its value only through the internal/config constants.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller could not name this file")
	}
	cliDir := filepath.Dir(thisFile)
	stampSite := filepath.Join(cliDir, "factory.go")
	stampBody := sdReadSource(t, stampSite)
	const wantStamp = "os.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)"
	if !strings.Contains(stampBody, wantStamp) {
		t.Errorf("%s does not carry the constant stamp %s — the marker is stamped through a literal or not at all", stampSite, wantStamp)
	}
	scanned := map[string]string{
		"factory.go":             stampBody,
		"factory_card.go":        sdReadSource(t, filepath.Join(cliDir, "factory_card.go")),
		"contract_sign_guard.go": sdReadSource(t, filepath.Join(cliDir, "..", "hook", "contract_sign_guard.go")),
	}
	for name, body := range scanned {
		for _, literal := range []string{
			`"MOAI_FACTORY_ROLE"`, `"MOAI_KANBAN_BACKEND"`, `"MOAI_FACTORY_WORKER"`, `"MOAI_KANBAN_CARD"`,
		} {
			if strings.Contains(body, literal) {
				t.Errorf("%s carries the raw literal %s — the env name exists only as the internal/config constant", name, literal)
			}
		}
		// The role VALUE literal is the pinned -f token's declaration alone
		// (the REQ-AP-013 pin); the stamp and compare sites carry no copy.
		for _, line := range strings.Split(body, "\n") {
			if !strings.Contains(line, `"lane"`) {
				continue
			}
			if name == "factory.go" && strings.Contains(line, "factoryLaneRoleToken = ") {
				continue
			}
			t.Errorf("%s carries the value literal \"lane\" outside the pinned token declaration: %s", name, strings.TrimSpace(line))
		}
	}
}
