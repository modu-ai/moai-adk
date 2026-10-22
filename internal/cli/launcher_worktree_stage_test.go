package cli

// launcher_worktree_stage_test.go — SPEC-LAUNCHER-ARGV-001 AC-LA-014 (card t1748):
// the pre-refactor characterization of how the NON-LEADER launch path reads the
// worktree selection today.
//
// The five non-guard readers do not read one argv. They read the launcher
// pipeline at three stages:
//
//	stage A  the raw entry argv           resolveWorktreeL2Path, splitWorktreeBranchFlag
//	stage B  A after the --branch strip   requireShortWorktreeName, ccWorktreeWriterPrecheck,
//	                                      seedAdmittedWorktreeHooks
//	stage C  the consumption loop's own   extractWorktreeLaunchTarget (chain node)
//	         passThrough output
//
// Every cell here drives the real runCC / runGLM entry down to the exec seam
// (runLaunchClaude with execOrSpawnClaudeFunc captured) and records what each
// reader stage observably did. The cells were observed GREEN on the unrefactored
// tree before the resolver landed; they are the SPEC §2 literal-preservation
// pin the view's consumer-scoped fields must reproduce.

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// stageObservation is what one launch of one spelling observably did.
type stageObservation struct {
	// Err is the launch error ("" when the launch was allowed), reduced to a
	// stable token: "l2-prefix" for the out-of-prefix refusal, "branch-name" for
	// the --branch name refusal, "writer-anchored" for the precheck refusal.
	Err string
	// Materialize lists the --branch materializations as "name@branch".
	Materialize []string
	// LaunchArgs is the argv handed to the launch seam (post normalizeWorktreeFlag).
	LaunchArgs []string
	// ExecTail is the exec-seam argv after the launcher-injected
	// --permission-mode/--model/--continue pairs and the leading "claude".
	ExecTail []string
	// ChainTrees are the chain-ledger node worktree paths, relative to the root.
	ChainTrees []string
	// EnvChainNode reports whether MOAI_CHAIN_NODE_ID reached the child env.
	EnvChainNode bool
	// PrecheckTree is the tree the writer precheck read ("" when none), observed
	// through its refusal text (run 2). cc arm only: the glm flow has no precheck.
	PrecheckTree string
	// SeedTree is the tree the hook seeder resolved ("" when none), observed at
	// the sessionWorktreeSeedCodexHooks seam (run 3).
	SeedTree string
}

// stageCandidateNames are the worktree names the cells use; run 2 and run 3
// create each as a directory outside Git, or as a registered Git worktree, so
// the precheck and the seeder have an authenticated tree to read.
var stageCandidateNames = []string{"plain", "other"}

// launchWithStageSeams drives one entry (arm "cc" or "glm") for args and returns
// the observation. rootKind selects the project root shape: "bare" (no tree
// directories), "dirs-nongit" (candidate trees exist, root is not a git repo, so
// the writer precheck refuses on the first tree it reads), "dirs-git" (candidate
// trees exist inside a git repo, so the precheck passes and the seeder runs).
//
// Callers must not use t.Parallel(): the helper chdirs, sets process
// environment and overrides package-level seams.
func launchWithStageSeams(t *testing.T, arm string, args []string, rootKind string) stageObservation {
	t.Helper()
	root := overrideProjectRoot(t)
	// Characterize this project's metadata, independently of repositories
	// (or sandbox protection mounts) above the temporary fixture.
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(root))
	if rootKind == "dirs-git" {
		m3Gitf(t, root, "init", "-q", "--template=", "-b", "main")
		m3Gitf(t, root, "-c", "commit.gpgsign=false", "commit", "-q", "--allow-empty", "-m", "stage fixture")
	}
	if rootKind != "bare" {
		for _, n := range stageCandidateNames {
			tree := filepath.Join(root, ".claude", "worktrees", n)
			if rootKind == "dirs-git" {
				m3Gitf(t, root, "worktree", "add", "-q", "--detach", tree, "HEAD")
			} else if err := os.MkdirAll(tree, 0o755); err != nil {
				t.Fatal(err)
			}
		}
	}
	// A bare .claude directory keeps syncPermissionModeToSettingsLocal inside the tmp root.
	if err := os.MkdirAll(filepath.Join(root, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	// The recorded root is the symlink-resolved one: findProjectRoot (cwd based)
	// resolves /var -> /private/var, and observations are compared relative to it.
	origDir, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(origDir) })
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvChainNodeID, "")
	t.Setenv(config.EnvClaudeBin, writeExecutable(t, filepath.Join(t.TempDir(), "claude-stub")))

	base := t.TempDir()
	origBase := profile.BaseDirOverride
	t.Cleanup(func() { profile.BaseDirOverride = origBase })
	profile.BaseDirOverride = base
	if err := os.WriteFile(filepath.Join(base, "preferences.yaml"), []byte("user_name: tester\neffort_level: medium\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var obs stageObservation
	var execArgs, execEnv []string

	origExec := execOrSpawnClaudeFunc
	t.Cleanup(func() { execOrSpawnClaudeFunc = origExec })
	execOrSpawnClaudeFunc = func(_ string, a []string, env []string) error {
		execArgs, execEnv = a, env
		return nil
	}
	origMat := launcherWorktreeMaterialize
	t.Cleanup(func() { launcherWorktreeMaterialize = origMat })
	launcherWorktreeMaterialize = func(_ string, name, branch string, _ io.Writer) error {
		obs.Materialize = append(obs.Materialize, name+"@"+branch)
		return nil
	}
	origSeed := sessionWorktreeSeedCodexHooks
	t.Cleanup(func() { sessionWorktreeSeedCodexHooks = origSeed })
	sessionWorktreeSeedCodexHooks = func(tree string, _ io.Writer) {
		obs.SeedTree = relToStageRoot(root, tree)
	}
	origLaunch := unifiedLaunchFunc
	t.Cleanup(func() { unifiedLaunchFunc = origLaunch })
	unifiedLaunchFunc = func(profileName, _ string, a []string) error {
		obs.LaunchArgs = append([]string(nil), a...)
		return runLaunchClaude(profileName, a)
	}
	origDeps := deps
	t.Cleanup(func() { deps = origDeps })
	deps = nil

	buf := new(bytes.Buffer)
	var err error
	switch arm {
	case "cc":
		ccCmd.SetOut(buf)
		ccCmd.SetErr(buf)
		err = runCC(ccCmd, args)
	case "glm":
		glmCmd.SetOut(buf)
		glmCmd.SetErr(buf)
		err = runGLM(glmCmd, args)
	default:
		t.Fatalf("unknown arm %q", arm)
	}

	obs.Err = stageErrToken(err)
	if err != nil && strings.Contains(err.Error(), worktreeWriterAnchoredSentinel) {
		obs.PrecheckTree = relToStageRoot(root, stageRefusedTree(err.Error()))
	}
	obs.ExecTail = stageExecTail(execArgs)
	obs.EnvChainNode = envValue(execEnv, config.EnvChainNodeID) != ""
	if _, statErr := os.Stat(filepath.Join(root, ".moai", "state", "chain", ChainEventsFile)); statErr == nil {
		for _, n := range reopenLedger(t, root) {
			obs.ChainTrees = append(obs.ChainTrees, relToStageRoot(root, n.WorktreePath))
		}
	}
	return obs
}

// stageErrToken reduces a launch error to a stable token.
func stageErrToken(err error) string {
	switch {
	case err == nil:
		return ""
	case strings.Contains(err.Error(), "is not under an accepted worktree prefix"):
		return "l2-prefix"
	case strings.Contains(err.Error(), worktreeWriterAnchoredSentinel):
		return "writer-anchored"
	case strings.Contains(err.Error(), "--branch"):
		return "branch-name"
	}
	return "other: " + err.Error()
}

// stageRefusedTree extracts the tree path from a writer-anchored refusal text.
func stageRefusedTree(msg string) string {
	const marker = "cannot read the worktree lock state of "
	if i := strings.Index(msg, marker); i >= 0 {
		rest := msg[i+len(marker):]
		if j := strings.Index(rest, " "); j >= 0 {
			return strings.TrimSuffix(rest[:j], ":")
		}
		return rest
	}
	return ""
}

// stageExecTail strips the leading "claude" and the launcher-injected
// --permission-mode/--model/--continue flags from the exec argv.
func stageExecTail(a []string) []string {
	var out []string
	for i := 0; i < len(a); i++ {
		switch {
		case i == 0 && a[i] == "claude":
		case a[i] == "--permission-mode" || a[i] == "--model":
			i++
		case a[i] == "--continue":
		default:
			out = append(out, a[i])
		}
	}
	return out
}

// relToStageRoot renders p relative to root, resolving the macOS /var symlink
// on both sides; paths outside root stay absolute.
func relToStageRoot(root, p string) string {
	if p == "" {
		return ""
	}
	// The raw spelling first (the chain ledger records the unresolved root), then
	// the symlink-resolved one (the writer precheck and the seeder report
	// whatever EvalSymlinks produced).
	if rel, err := filepath.Rel(root, p); err == nil && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(rel)
	}
	rr, pr := root, p
	if r, err := filepath.EvalSymlinks(root); err == nil {
		rr = r
	}
	if r, err := filepath.EvalSymlinks(p); err == nil {
		pr = r
	}
	if rel, err := filepath.Rel(rr, pr); err == nil && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(p)
}
