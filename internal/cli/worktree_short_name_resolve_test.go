package cli

// Tests for the launcher's -w short-name .moai collision resolution (card
// t1791, GH #1813): a short-name -w that names an existing registered
// .moai/worktrees/<name> tree must resolve to that tree's absolute path (the
// session re-enters it), a same-named path that is NOT a registered worktree
// must warn and abort (never silently create the .claude twin), and a name
// with no same-named .moai tree must pass through unchanged (the backend
// creates .claude/worktrees/<name> as before).

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// moaiStateRecorder drives the state seam and records the names it was asked
// about, so tests can assert the seam is consulted exactly when a short name
// stands.
type moaiStateRecorder struct {
	calls []string
	state moaiShortNameState
	err   error
}

func (r *moaiStateRecorder) resolve(string, string) (moaiShortNameState, error) {
	r.calls = append(r.calls, "called")
	return r.state, r.err
}

// overrideMoaiStateSeam swaps the state seam for rec and restores it on
// cleanup.
func overrideMoaiStateSeam(t *testing.T, rec *moaiStateRecorder) {
	t.Helper()
	original := resolveMoaiWorktreeShortNameState
	resolveMoaiWorktreeShortNameState = rec.resolve
	t.Cleanup(func() { resolveMoaiWorktreeShortNameState = original })
}

// overrideFindRoot points findProjectRootFn at a throwaway directory for the
// test's duration.
func overrideFindRoot(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	original := findProjectRootFn
	findProjectRootFn = func() (string, error) { return dir, nil }
	t.Cleanup(func() { findProjectRootFn = original })
	return dir
}

// TestResolveMoaiWorktreeShortName_RegisteredMoaiTreeRewritten is scenario (a):
// a short name with an existing registered .moai/worktrees/<name> tree
// resolves to that tree's absolute path — the session re-enters the MoAI tree,
// and no .claude/worktrees/<name> twin is materialized.
func TestResolveMoaiWorktreeShortName_RegisteredMoaiTreeRewritten(t *testing.T) {
	root := overrideFindRoot(t)
	rec := &moaiStateRecorder{state: moaiShortNameRegistered}
	overrideMoaiStateSeam(t, rec)

	got, err := resolveMoaiWorktreeShortName([]string{"-w", "feat-login", "-c"}, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(root, ".moai", "worktrees", "feat-login")
	wantArgs := []string{"-w", want, "-c"}
	if strings.Join(got, "\x00") != strings.Join(wantArgs, "\x00") {
		t.Fatalf("args = %v, want %v", got, wantArgs)
	}
	if len(rec.calls) != 1 {
		t.Fatalf("state seam calls = %d, want 1", len(rec.calls))
	}

	// The rewritten argv resolves end to end: every stage read and the backend
	// selection see the .moai tree, not a short name for claude to resolve
	// against .claude/worktrees/.
	view := resolveLaunchView(got)
	if view.StageA.Value != want || view.StageB.Value != want {
		t.Fatalf("stage reads = %q / %q, want both %q", view.StageA.Value, view.StageB.Value, want)
	}
	if !view.Selection.HasValue || view.Selection.Value != want {
		t.Fatalf("backend selection = %+v, want the .moai tree %q", view.Selection, want)
	}
	if strings.Contains(view.Selection.Value, filepath.Join(".claude", "worktrees")) {
		t.Fatalf("selection resolved to a .claude twin path: %q", view.Selection.Value)
	}
}

// TestResolveMoaiWorktreeShortName_ConflictWarnsAndAborts is scenario (b): a
// same-named .moai/worktrees/<name> path that is NOT a registered worktree
// aborts the launch with a loud warning instead of silently creating the
// .claude twin.
func TestResolveMoaiWorktreeShortName_ConflictWarnsAndAborts(t *testing.T) {
	root := overrideFindRoot(t)
	rec := &moaiStateRecorder{state: moaiShortNameConflict}
	overrideMoaiStateSeam(t, rec)

	var warn strings.Builder
	got, err := resolveMoaiWorktreeShortName([]string{"-w", "stale-tree"}, &warn)
	if err == nil {
		t.Fatal("expected an abort error for a non-registered same-named path")
	}
	if got != nil {
		t.Fatalf("args = %v on abort, want nil", got)
	}
	if !strings.Contains(warn.String(), filepath.Join(root, ".moai", "worktrees", "stale-tree")) {
		t.Fatalf("warning %q does not name the colliding path", warn.String())
	}
	if !strings.Contains(warn.String(), ".claude") {
		t.Fatalf("warning %q does not name the .claude twin it refuses to create", warn.String())
	}
}

// TestResolveMoaiWorktreeShortName_AbsentPassesThrough is scenario (c): a
// short name with no same-named .moai tree passes through unchanged — the
// backend creates .claude/worktrees/<name> exactly as before.
func TestResolveMoaiWorktreeShortName_AbsentPassesThrough(t *testing.T) {
	overrideFindRoot(t)
	rec := &moaiStateRecorder{state: moaiShortNameAbsent}
	overrideMoaiStateSeam(t, rec)

	args := []string{"-w", "feat-login", "--permission-mode", "auto"}
	var warn strings.Builder
	got, err := resolveMoaiWorktreeShortName(args, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Join(got, "\x00") != strings.Join(args, "\x00") {
		t.Fatalf("args mutated without a same-named .moai tree: %v", got)
	}
	if warn.Len() != 0 {
		t.Fatalf("warning written on the pass-through path: %q", warn.String())
	}
}

// TestResolveMoaiWorktreeShortName_NonShortNameShapesNoop pins the no-op
// shapes: no -w at all, a bare -w, an absolute -w, and a value that is not a
// single path segment are all passed through untouched — the state seam is
// never consulted for them.
func TestResolveMoaiWorktreeShortName_NonShortNameShapesNoop(t *testing.T) {
	overrideFindRoot(t)
	rec := &moaiStateRecorder{state: moaiShortNameRegistered}
	overrideMoaiStateSeam(t, rec)

	tests := []struct {
		name string
		args []string
	}{
		{name: "no worktree flag", args: []string{"-c", "--model", "opus"}},
		{name: "bare -w", args: []string{"-w"}},
		{name: "bare -w before flag", args: []string{"-w", "-c"}},
		{name: "absolute -w", args: []string{"-w", "/tmp/somewhere"}},
		{name: "separator in name", args: []string{"-w", "a/b"}},
		{name: "dot name", args: []string{"-w", "."}},
		{name: "dotdot name", args: []string{"-w", ".."}},
		{name: "after dashdash", args: []string{"--", "-w", "feat-login"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveMoaiWorktreeShortName(tt.args, io.Discard)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.Join(got, "\x00") != strings.Join(tt.args, "\x00") {
				t.Fatalf("args = %v, want unchanged %v", got, tt.args)
			}
		})
	}
	if len(rec.calls) != 0 {
		t.Fatalf("state seam consulted %d times on no-op shapes, want 0", len(rec.calls))
	}
}

// TestResolveMoaiWorktreeShortName_Spellings covers every two-token and =
// spelling worktreeFlagValue understands: the value is replaced in place and
// the token count never changes.
func TestResolveMoaiWorktreeShortName_Spellings(t *testing.T) {
	root := overrideFindRoot(t)
	rec := &moaiStateRecorder{state: moaiShortNameRegistered}
	overrideMoaiStateSeam(t, rec)

	want := filepath.Join(root, ".moai", "worktrees", "feat-login")
	tests := []struct {
		name     string
		args     []string
		wantArgs []string
	}{
		{
			name:     "short two-token",
			args:     []string{"-w", "feat-login", "-c"},
			wantArgs: []string{"-w", want, "-c"},
		},
		{
			name:     "long two-token",
			args:     []string{"--worktree", "feat-login"},
			wantArgs: []string{"--worktree", want},
		},
		{
			name:     "long equals",
			args:     []string{"--worktree=feat-login", "-c"},
			wantArgs: []string{"--worktree=" + want, "-c"},
		},
		{
			name:     "short equals",
			args:     []string{"-w=feat-login"},
			wantArgs: []string{"-w=" + want},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveMoaiWorktreeShortName(tt.args, io.Discard)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.Join(got, "\x00") != strings.Join(tt.wantArgs, "\x00") {
				t.Fatalf("args = %v, want %v", got, tt.wantArgs)
			}
		})
	}
}

// TestResolveMoaiWorktreeShortName_StateErrorPropagates: a failure reading the
// worktree state aborts the launch rather than degrading to twin creation.
func TestResolveMoaiWorktreeShortName_StateErrorPropagates(t *testing.T) {
	overrideFindRoot(t)
	sentinel := errors.New("git failed (sentinel)")
	rec := &moaiStateRecorder{err: sentinel}
	overrideMoaiStateSeam(t, rec)

	if _, err := resolveMoaiWorktreeShortName([]string{"-w", "feat-login"}, io.Discard); !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want the state-read failure %v", err, sentinel)
	}
}

// setupRegisteredTreeRepo builds a throwaway repository holding a real
// registered worktree at .moai/worktrees/<name>, points findProjectRootFn at
// it, and returns the repo root, the tree's absolute path, and — for the
// conflict case — a plain directory under .moai/worktrees/.
func setupRegisteredTreeRepo(t *testing.T, name string) (string, string, string) {
	t.Helper()
	if testing.Short() {
		t.Skip("real-git integration test")
	}
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve symlinks: %v", err)
	}
	runTestGit(t, repo, "init", "-b", "main")
	runTestGit(t, repo, "config", "user.email", "test@example.com")
	runTestGit(t, repo, "config", "user.name", "Test User")
	if werr := os.WriteFile(filepath.Join(repo, "README.md"), []byte("# Test\n"), 0o644); werr != nil {
		t.Fatal(werr)
	}
	runTestGit(t, repo, "add", "README.md")
	runTestGit(t, repo, "commit", "-m", "init")
	runTestGit(t, repo, "worktree", "add", "-b", "WT-"+name, filepath.Join(repo, ".moai", "worktrees", name))
	if mkErr := os.MkdirAll(filepath.Join(repo, ".moai", "worktrees", "plain-dir"), 0o755); mkErr != nil {
		t.Fatal(mkErr)
	}

	original := findProjectRootFn
	findProjectRootFn = func() (string, error) { return repo, nil }
	t.Cleanup(func() { findProjectRootFn = original })
	return repo, filepath.Join(repo, ".moai", "worktrees", name), filepath.Join(repo, ".moai", "worktrees", "plain-dir")
}

// TestResolveMoaiWorktreeShortNameStateReal_Integration runs the real state
// classifier against a throwaway repository: a real git worktree reads as
// registered, a plain directory as a conflict, a missing name as absent.
func TestResolveMoaiWorktreeShortNameStateReal_Integration(t *testing.T) {
	repo, treePath, plainDir := setupRegisteredTreeRepo(t, "feat-login")
	_ = repo

	if got, err := resolveMoaiWorktreeShortNameStateReal(repo, "feat-login"); err != nil || got != moaiShortNameRegistered {
		t.Fatalf("state(%s) = %d, %v; want registered, nil", treePath, got, err)
	}
	if got, err := resolveMoaiWorktreeShortNameStateReal(repo, "plain-dir"); err != nil || got != moaiShortNameConflict {
		t.Fatalf("state(%s) = %d, %v; want conflict, nil", plainDir, got, err)
	}
	if got, err := resolveMoaiWorktreeShortNameStateReal(repo, "no-such-tree"); err != nil || got != moaiShortNameAbsent {
		t.Fatalf("state(missing) = %d, %v; want absent, nil", got, err)
	}
}

// wiringLaunchRecorder captures the argv the entry hands to the launch seam.
type wiringLaunchRecorder struct {
	profile string
	mode    string
	args    []string
}

func (r *wiringLaunchRecorder) launch(profile, mode string, args []string) error {
	r.profile, r.mode, r.args = profile, mode, args
	return nil
}

// overrideUnifiedLaunch swaps the launch seam for rec and restores it.
func overrideUnifiedLaunch(t *testing.T, rec *wiringLaunchRecorder) {
	t.Helper()
	original := unifiedLaunchFunc
	unifiedLaunchFunc = rec.launch
	t.Cleanup(func() { unifiedLaunchFunc = original })
}

// backendWorktreeValue extracts the worktree value the backend argv carries,
// or reports that none stands.
func backendWorktreeValue(args []string) (string, bool) {
	for i, tok := range args {
		if tok == "--worktree" && i+1 < len(args) {
			return args[i+1], true
		}
	}
	return "", false
}

// TestCCEntry_EntersExistingMoaiTreeOnShortName is the end-to-end scenario (a)
// through the cc entry: `moai cc -w feat-login` with a registered
// .moai/worktrees/feat-login launches the backend with the .moai tree's
// absolute path, not the short name for claude to twin.
func TestCCEntry_EntersExistingMoaiTreeOnShortName(t *testing.T) {
	_, treePath, _ := setupRegisteredTreeRepo(t, "feat-login")
	rec := &wiringLaunchRecorder{}
	overrideUnifiedLaunch(t, rec)

	buf := new(strings.Builder)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	if err := ccCmd.RunE(ccCmd, []string{"-w", "feat-login"}); err != nil {
		t.Fatalf("cc entry: %v", err)
	}
	if rec.args == nil {
		t.Fatal("launch seam never ran")
	}
	got, ok := backendWorktreeValue(rec.args)
	if !ok {
		t.Fatalf("no --worktree option in the launched argv %v", rec.args)
	}
	if got != treePath {
		t.Fatalf("backend worktree value = %q, want the .moai tree %q", got, treePath)
	}
}

// TestGLMEntry_EntersExistingMoaiTreeOnShortName is the same end-to-end
// resolution through the glm entry, which mirrors the cc flow.
func TestGLMEntry_EntersExistingMoaiTreeOnShortName(t *testing.T) {
	_, treePath, _ := setupRegisteredTreeRepo(t, "feat-login")
	rec := &wiringLaunchRecorder{}
	overrideUnifiedLaunch(t, rec)

	buf := new(strings.Builder)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	if err := runGLM(glmCmd, []string{"-w", "feat-login"}); err != nil {
		t.Fatalf("glm entry: %v", err)
	}
	if rec.args == nil {
		t.Fatal("launch seam never ran")
	}
	got, ok := backendWorktreeValue(rec.args)
	if !ok {
		t.Fatalf("no --worktree option in the launched argv %v", rec.args)
	}
	if got != treePath {
		t.Fatalf("backend worktree value = %q, want the .moai tree %q", got, treePath)
	}
}

// TestCCEntry_AbsentNameKeepsShortName pins scenario (c) end to end: with no
// same-named .moai tree the backend still receives the short name and creates
// .claude/worktrees/<name> itself.
func TestCCEntry_AbsentNameKeepsShortName(t *testing.T) {
	if testing.Short() {
		t.Skip("real-git integration test")
	}
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve symlinks: %v", err)
	}
	runTestGit(t, repo, "init", "-b", "main")
	runTestGit(t, repo, "config", "user.email", "test@example.com")
	runTestGit(t, repo, "config", "user.name", "Test User")
	if werr := os.WriteFile(filepath.Join(repo, "README.md"), []byte("# Test\n"), 0o644); werr != nil {
		t.Fatal(werr)
	}
	runTestGit(t, repo, "add", "README.md")
	runTestGit(t, repo, "commit", "-m", "init")
	original := findProjectRootFn
	findProjectRootFn = func() (string, error) { return repo, nil }
	t.Cleanup(func() { findProjectRootFn = original })

	rec := &wiringLaunchRecorder{}
	overrideUnifiedLaunch(t, rec)

	buf := new(strings.Builder)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	if err := ccCmd.RunE(ccCmd, []string{"-w", "fresh-name"}); err != nil {
		t.Fatalf("cc entry: %v", err)
	}
	got, ok := backendWorktreeValue(rec.args)
	if !ok {
		t.Fatalf("no --worktree option in the launched argv %v (repo %s)", rec.args, repo)
	}
	if got != "fresh-name" {
		t.Fatalf("backend worktree value = %q, want the unchanged short name", got)
	}
}
