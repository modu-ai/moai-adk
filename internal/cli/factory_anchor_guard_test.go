package cli

// factory_anchor_guard_test.go — SPEC-LAUNCHER-ARGV-001 M2: the leader-launch
// anchor guard.
//
//   - TestLauncherAnchorFinalArgv          the three iter8 families (a)(b)(c) plus the
//                                          full §D.4 table through the real leader entry
//   - TestLauncherAnchorGuardSemantics     placement, exemptions, override, fail-open, env independence
//   - TestLauncherAnchorBranchAnnotation   the --branch effective-branch rules
//   - TestLauncherAnchorRunCountInstrument family (d): the run-record counter is not vacuous
//
// Every cell drives the real runCC / runGLM leader entry (`-f`) in a real git
// repository whose card tree sits on a WT- branch; only the engine launch, the
// claude binary and the --branch materializer are substituted.

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// leaderFixture is a primary checkout with three linked trees under
// .claude/worktrees/: WT-card (branch WT-card — a card tree), plain (branch
// plain-branch) and plain-tree (branch plain-tree-branch).
type leaderFixture struct {
	root string
	base string
	card string
}

func newLeaderFixture(t *testing.T) *leaderFixture {
	t.Helper()
	repo := newAnchorRepo(t)
	t.Setenv("MOAI_HOME", t.TempDir())
	netScrubLaneEnv(t)
	fx := &leaderFixture{root: repo.root, base: repo.base}
	add := func(name, branch string) string {
		path := filepath.Join(repo.root, ".claude", "worktrees", name)
		anchorGit(t, repo.root, "worktree", "add", "-q", "-b", branch, path, repo.base)
		return path
	}
	fx.card = add("WT-card", "WT-card")
	add("plain", "plain-branch")
	add("plain-tree", "plain-tree-branch")
	t.Chdir(repo.root)
	return fx
}

// leaderObs is what one leader launch observably did.
type leaderObs struct {
	err          error
	launched     bool
	execArgs     []string
	materialized []string
	stderr       string

	// The state the guard saw as it started (leaderAnchorGuardProbe).
	probed          bool
	probeFactoryEnv map[string]string
	probeRuns       int
	probeSettings   int

	runsBefore, runsAfter int
}

// launcherRunCount is the instrument that counts the runs a project recorded:
// the canonical queue reader (LoadPure), which serves whichever storage layout
// the project carries. A reader of one layout only reads zero on the other —
// the vacuous counter of the iter8 family (d).
func launcherRunCount(t *testing.T, root string) int {
	t.Helper()
	record, err := factory.NewBacklogStore(factory.BacklogPathForRoot(root)).LoadPure()
	if err != nil {
		t.Fatalf("load the queue record: %v", err)
	}
	return len(record.Runtime.Runs)
}

// settingsArtifacts counts the transient factory settings files under the
// directory the launcher writes them to.
func settingsArtifacts(dir string) int {
	matches, _ := filepath.Glob(filepath.Join(dir, "moai-factory-*.json"))
	return len(matches)
}

// driveLeaderLaunch runs one leader (or other) entry for arm ("cc"/"glm") with
// the launch substituted below the entry: unifiedLaunchFunc hands the argv to
// the real runLaunchClaude, whose exec seam is captured.
func driveLeaderLaunch(t *testing.T, fx *leaderFixture, arm string, args []string) leaderObs {
	t.Helper()
	var obs leaderObs

	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	t.Setenv("TEMP", tmp)
	t.Setenv("TMP", tmp)

	base := t.TempDir()
	origBase := profile.BaseDirOverride
	t.Cleanup(func() { profile.BaseDirOverride = origBase })
	profile.BaseDirOverride = base
	if err := os.WriteFile(filepath.Join(base, "preferences.yaml"), []byte("user_name: tester\neffort_level: medium\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvClaudeBin, writeExecutable(t, filepath.Join(t.TempDir(), "claude-stub")))

	origExec := execOrSpawnClaudeFunc
	t.Cleanup(func() { execOrSpawnClaudeFunc = origExec })
	execOrSpawnClaudeFunc = func(_ string, a []string, _ []string) error {
		obs.execArgs = append([]string(nil), a...)
		return nil
	}
	origMat := launcherWorktreeMaterialize
	t.Cleanup(func() { launcherWorktreeMaterialize = origMat })
	launcherWorktreeMaterialize = func(_ string, name, branch string, _ io.Writer) error {
		obs.materialized = append(obs.materialized, name+"@"+branch)
		return nil
	}
	origLaunch := unifiedLaunchFunc
	t.Cleanup(func() { unifiedLaunchFunc = origLaunch })
	unifiedLaunchFunc = func(profileName, _ string, a []string) error {
		obs.launched = true
		return runLaunchClaude(profileName, a)
	}
	origProbe := leaderAnchorGuardProbe
	t.Cleanup(func() { leaderAnchorGuardProbe = origProbe })
	leaderAnchorGuardProbe = func() {
		obs.probed = true
		obs.probeFactoryEnv = factoryMarkerEnv()
		obs.probeRuns = launcherRunCount(t, fx.root)
		obs.probeSettings = settingsArtifacts(tmp)
	}
	origRoot := findProjectRootFn
	t.Cleanup(func() { findProjectRootFn = origRoot })
	findProjectRootFn = func() (string, error) { return fx.root, nil }
	origDeps := deps
	t.Cleanup(func() { deps = origDeps })
	deps = nil

	obs.runsBefore = launcherRunCount(t, fx.root)
	errBuf := new(bytes.Buffer)
	switch arm {
	case "cc":
		ccCmd.SetOut(io.Discard)
		ccCmd.SetErr(errBuf)
		obs.err = runCC(ccCmd, args)
	case "glm":
		glmCmd.SetOut(io.Discard)
		glmCmd.SetErr(errBuf)
		obs.err = runGLM(glmCmd, args)
	default:
		t.Fatalf("unknown arm %q", arm)
	}
	obs.stderr = errBuf.String()
	obs.runsAfter = launcherRunCount(t, fx.root)
	return obs
}

// factoryMarkerEnv returns the factory / kanban marker variables of the process.
func factoryMarkerEnv() map[string]string {
	out := map[string]string{}
	for k, v := range netEnvMap() {
		if strings.HasPrefix(k, "MOAI_FACTORY") || strings.HasPrefix(k, "MOAI_KANBAN") {
			out[k] = v
		}
	}
	return out
}

// stripLauncherAppended removes the launcher-appended tail flags of a leader
// launch (--name, --settings and the --effort the factory preparation appends)
// from the end of an exec argv tail. They are appended after the guard has
// judged the view and carry no worktree option.
func stripLauncherAppended(tail []string) []string {
	out := append([]string(nil), tail...)
	for len(out) >= 2 {
		switch out[len(out)-2] {
		case "--name", "--settings", "--effort":
			out = out[:len(out)-2]
		default:
			return out
		}
	}
	return out
}

// assertRefusedBeforeMutation checks the three placement indicators of AC-LA-006
// for a refused leader launch: no factory marker in the environment, no run
// record and no settings artifact when the guard started, no launch.
func assertRefusedBeforeMutation(t *testing.T, label string, obs leaderObs) {
	t.Helper()
	if obs.err == nil {
		t.Errorf("%s: the anchored leader launch was not refused (launched=%v)", label, obs.launched)
		return
	}
	if obs.launched || obs.execArgs != nil {
		t.Errorf("%s: a refused launch reached the engine (launched=%v)", label, obs.launched)
	}
	if !obs.probed {
		t.Errorf("%s: the guard never ran", label)
		return
	}
	if len(obs.probeFactoryEnv) != 0 {
		t.Errorf("%s: factory markers stamped before the guard ran: %v", label, obs.probeFactoryEnv)
	}
	if obs.probeRuns != obs.runsBefore || obs.runsAfter != obs.runsBefore {
		t.Errorf("%s: run records before=%d at guard=%d after=%d, want all equal", label, obs.runsBefore, obs.probeRuns, obs.runsAfter)
	}
	if obs.probeSettings != 0 {
		t.Errorf("%s: %d settings artifacts existed when the guard ran, want 0", label, obs.probeSettings)
	}
	if len(obs.materialized) != 0 {
		t.Errorf("%s: the worktree was materialized before the refusal: %v", label, obs.materialized)
	}
}

// leaderArms names the two launcher arms with their backend constants.
var leaderArms = []string{"cc", "glm"}
