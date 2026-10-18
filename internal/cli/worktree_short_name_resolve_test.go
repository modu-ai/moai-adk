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
// stands — and about the name the backend grammar selects.
type moaiStateRecorder struct {
	calls []string
	state moaiShortNameState
	err   error
}

func (r *moaiStateRecorder) resolve(_ string, name string) (moaiShortNameState, error) {
	r.calls = append(r.calls, name)
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

// TestResolveMoaiWorktreeShortName_AfterLauncherSeparatorHonored pins the
// last-honored grammar across the launcher's separator: the first "--" is the
// launcher's own (consumed; the tail flows to claude verbatim), so a tail -w
// IS the worktree the backend selects — selectBackendWorktree over the
// pass-through argv honors it, and the resolution judges it like any other
// honored occurrence. The pre-correction first-match read stopped at the
// separator and passed the collision through (sync-audit F1's grammar split).
func TestResolveMoaiWorktreeShortName_AfterLauncherSeparatorHonored(t *testing.T) {
	root := overrideFindRoot(t)
	rec := &moaiStateRecorder{state: moaiShortNameRegistered}
	overrideMoaiStateSeam(t, rec)

	got, err := resolveMoaiWorktreeShortName([]string{"--", "-w", "feat-login"}, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(root, ".moai", "worktrees", "feat-login")
	wantArgs := []string{"--", "-w", want}
	if strings.Join(got, "\x00") != strings.Join(wantArgs, "\x00") {
		t.Fatalf("args = %v, want %v", got, wantArgs)
	}
	if len(rec.calls) != 1 || rec.calls[0] != "feat-login" {
		t.Fatalf("state seam consulted about %v, want exactly [feat-login]", rec.calls)
	}
	view := resolveLaunchView(got)
	if !view.Selection.HasValue || view.Selection.Value != want {
		t.Fatalf("backend selection = %+v, want the .moai tree %q", view.Selection, want)
	}
}

// TestResolveMoaiWorktreeShortName_LastHonoredValueRewritten is the sync-audit
// F1 scenario: the backend selection is LAST-wins (selectBackendWorktree
// returns a fresh selection per -w occurrence), so the resolver must judge and
// rewrite the LAST honored occurrence — `-w fresh -w card` re-enters the
// registered .moai/worktrees/card instead of letting the backend twin it.
func TestResolveMoaiWorktreeShortName_LastHonoredValueRewritten(t *testing.T) {
	root := overrideFindRoot(t)
	rec := &moaiStateRecorder{state: moaiShortNameRegistered}
	overrideMoaiStateSeam(t, rec)

	got, err := resolveMoaiWorktreeShortName([]string{"-w", "fresh", "-w", "card"}, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(root, ".moai", "worktrees", "card")
	wantArgs := []string{"-w", "fresh", "-w", want}
	if strings.Join(got, "\x00") != strings.Join(wantArgs, "\x00") {
		t.Fatalf("args = %v, want %v (the LAST honored -w value rewritten)", got, wantArgs)
	}
	if len(rec.calls) != 1 || rec.calls[0] != "card" {
		t.Fatalf("state seam consulted about %v, want exactly [card] — the backend-selected value", rec.calls)
	}
	// The rewritten argv's backend selection is the .moai tree: what the
	// resolver judged is what selectBackendWorktree will pick.
	view := resolveLaunchView(got)
	if !view.Selection.HasValue || view.Selection.Value != want {
		t.Fatalf("backend selection = %+v, want the .moai tree %q", view.Selection, want)
	}
}

// TestResolveMoaiWorktreeShortName_LastWinsSkipsEarlierRegistered pins the
// same grammar from the other side: with `-w card -w fresh` the backend
// enters fresh, so a registered card tree earlier in the argv is never the
// collision — the resolution must not touch it.
func TestResolveMoaiWorktreeShortName_LastWinsSkipsEarlierRegistered(t *testing.T) {
	overrideFindRoot(t)
	rec := &moaiStateRecorder{state: moaiShortNameAbsent}
	overrideMoaiStateSeam(t, rec)

	args := []string{"-w", "card", "-w", "fresh"}
	got, err := resolveMoaiWorktreeShortName(args, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Join(got, "\x00") != strings.Join(args, "\x00") {
		t.Fatalf("args = %v, want unchanged %v — the last-honored name is fresh, not card", got, args)
	}
	if len(rec.calls) != 1 || rec.calls[0] != "fresh" {
		t.Fatalf("state seam consulted about %v, want exactly [fresh]", rec.calls)
	}
}

// TestResolveMoaiWorktreeShortName_AttachedSpellingRefused covers the attached
// and cluster spellings scanShortCluster honors and worktreeFlagValue never
// scans (sync-audit F1): with a registered same-named .moai tree the launch is
// refused as ambiguous launcher input — the first-match stage readers cannot
// judge these forms, so rewriting one would resolve a value the downstream
// readers never see.
func TestResolveMoaiWorktreeShortName_AttachedSpellingRefused(t *testing.T) {
	overrideFindRoot(t)

	tests := []struct {
		name string
		args []string
	}{
		{name: "attached", args: []string{"-wcard"}},
		{name: "cluster attached", args: []string{"-cwcard"}},
		{name: "cluster two-token", args: []string{"-cw", "card"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &moaiStateRecorder{state: moaiShortNameRegistered}
			overrideMoaiStateSeam(t, rec)
			var warn strings.Builder
			got, err := resolveMoaiWorktreeShortName(tt.args, &warn)
			if err == nil {
				t.Fatalf("args %v: expected a refusal for the attached spelling over a registered .moai tree", tt.args)
			}
			if got != nil {
				t.Fatalf("args = %v on refusal, want nil", got)
			}
			if !strings.Contains(warn.String(), "card") {
				t.Fatalf("warning %q does not name the attached value", warn.String())
			}
			if len(rec.calls) != 1 || rec.calls[0] != "card" {
				t.Fatalf("state seam consulted about %v, want exactly [card]", rec.calls)
			}
		})
	}
}

// TestResolveMoaiWorktreeShortName_AttachedSpellingAbsentPassesThrough keeps
// today's boundary for the no-collision case: an attached spelling with no
// same-named .moai tree passes through — the backend creates
// .claude/worktrees/card as before.
func TestResolveMoaiWorktreeShortName_AttachedSpellingAbsentPassesThrough(t *testing.T) {
	overrideFindRoot(t)
	rec := &moaiStateRecorder{state: moaiShortNameAbsent}
	overrideMoaiStateSeam(t, rec)

	args := []string{"-wcard"}
	var warn strings.Builder
	got, err := resolveMoaiWorktreeShortName(args, &warn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Join(got, "\x00") != strings.Join(args, "\x00") {
		t.Fatalf("args = %v, want unchanged %v", got, args)
	}
	if warn.Len() != 0 {
		t.Fatalf("warning written on the pass-through path: %q", warn.String())
	}
	if len(rec.calls) != 1 || rec.calls[0] != "card" {
		t.Fatalf("state seam consulted about %v, want exactly [card]", rec.calls)
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

// TestResolveMoaiWorktreeShortName_ValueConsumerAdjacentHonoredLast is the
// sync-audit delta D1 shape A: the launcher strips -c, so the backend's
// required-value `--append-system-prompt` consumes the SECOND -w as its value
// and the selection stays `card`. The value the resolver judges must be the
// one the launcher-consumed argv selects — walking the raw argv honors the
// last `-w fresh` instead and never judges the registered card tree.
func TestResolveMoaiWorktreeShortName_ValueConsumerAdjacentHonoredLast(t *testing.T) {
	root := overrideFindRoot(t)
	rec := &moaiStateRecorder{state: moaiShortNameRegistered}
	overrideMoaiStateSeam(t, rec)

	got, err := resolveMoaiWorktreeShortName([]string{"-w", "card", "--append-system-prompt", "-c", "-w", "fresh"}, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(root, ".moai", "worktrees", "card")
	wantArgs := []string{"-w", want, "--append-system-prompt", "-c", "-w", "fresh"}
	if strings.Join(got, "\x00") != strings.Join(wantArgs, "\x00") {
		t.Fatalf("args = %v, want %v (the backend-selected card value rewritten)", got, wantArgs)
	}
	if len(rec.calls) != 1 || rec.calls[0] != "card" {
		t.Fatalf("state seam consulted about %v, want exactly [card]", rec.calls)
	}
	// The rewritten argv's backend selection is the .moai tree: what the
	// resolver judged is what selectBackendWorktree will pick.
	view := resolveLaunchView(got)
	if !view.Selection.HasValue || view.Selection.Value != want {
		t.Fatalf("backend selection = %+v, want the .moai tree %q", view.Selection, want)
	}
}

// TestResolveMoaiWorktreeShortName_BareWorktreeBeforeLauncherBoolean is the
// sync-audit delta D1 shape B: the raw walk reads `-w` valueless (the next
// token -c is flag-shaped), but the launcher REMOVES -c, so the backend sees
// `-w card` and honors card. The judgment must run over the stripped argv —
// otherwise the registered card tree is never consulted and the backend twins
// it.
func TestResolveMoaiWorktreeShortName_BareWorktreeBeforeLauncherBoolean(t *testing.T) {
	root := overrideFindRoot(t)
	rec := &moaiStateRecorder{state: moaiShortNameRegistered}
	overrideMoaiStateSeam(t, rec)

	got, err := resolveMoaiWorktreeShortName([]string{"-w", "-c", "card"}, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(root, ".moai", "worktrees", "card")
	// The rewrite replaces the raw position of the value token the backend
	// honors — card, AFTER the launcher-owned -c, which stays for the launcher
	// to strip; the normalized argv then reads --worktree -c <tree>, and the
	// consumed backend argv reads --worktree <tree>.
	wantArgs := []string{"-w", "-c", want}
	if strings.Join(got, "\x00") != strings.Join(wantArgs, "\x00") {
		t.Fatalf("args = %v, want %v (the backend-honored card value rewritten)", got, wantArgs)
	}
	if len(rec.calls) != 1 || rec.calls[0] != "card" {
		t.Fatalf("state seam consulted about %v, want exactly [card]", rec.calls)
	}
	view := resolveLaunchView(got)
	if !view.Selection.HasValue || view.Selection.Value != want {
		t.Fatalf("backend selection = %+v, want the .moai tree %q", view.Selection, want)
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

// TestResolveMoaiWorktreeShortNameStateReal_DanglingSymlinkIsConflict is the
// sync-audit F2 case: a dangling symlink at .moai/worktrees/<name> exists as a
// directory entry but resolves to nothing. os.Stat follows the link and reads
// ENOENT, which must NOT classify absent — that would let the backend
// materialize the .claude twin beside a same-named .moai path entry.
func TestResolveMoaiWorktreeShortNameStateReal_DanglingSymlinkIsConflict(t *testing.T) {
	repo, _, _ := setupRegisteredTreeRepo(t, "feat-login")
	dangling := filepath.Join(repo, ".moai", "worktrees", "dangling")
	if err := os.Symlink(filepath.Join(repo, ".moai", "worktrees", "no-such-target"), dangling); err != nil {
		t.Fatal(err)
	}
	got, err := resolveMoaiWorktreeShortNameStateReal(repo, "dangling")
	if err != nil || got != moaiShortNameConflict {
		t.Fatalf("state(dangling symlink) = %d, %v; want conflict, nil", got, err)
	}
}

// TestResolveMoaiWorktreeShortNameStateReal_InheritedGitDirScrubbed is the
// sync-audit F3 case: an exported GIT_DIR outranks -C, so the registration
// check's git subprocess must run with the scrubbed environment — otherwise it
// reads ANOTHER repository's worktree list and a legitimately registered tree
// spuriously aborts.
func TestResolveMoaiWorktreeShortNameStateReal_InheritedGitDirScrubbed(t *testing.T) {
	repo, treePath, _ := setupRegisteredTreeRepo(t, "feat-login")
	t.Setenv("GIT_DIR", "/dev/null")
	got, err := resolveMoaiWorktreeShortNameStateReal(repo, "feat-login")
	if err != nil || got != moaiShortNameRegistered {
		t.Fatalf("state(%s) under inherited GIT_DIR = %d, %v; want registered, nil", treePath, got, err)
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

// lastBackendWorktreeValue mirrors the backend's last-wins selection over the
// normalized launch argv the seam receives: the value of the LAST --worktree
// option, not the first.
func lastBackendWorktreeValue(args []string) (string, bool) {
	value, has := "", false
	for i, tok := range args {
		if tok == "--worktree" && i+1 < len(args) {
			value, has = args[i+1], true
		}
	}
	return value, has
}

// TestCCEntry_LastHonoredShortNameEntersMoaiTree is the sync-audit F1
// end-to-end shape through the cc entry: `-w fresh -w feat-login` with a
// registered .moai/worktrees/feat-login must launch with the .moai tree as
// the LAST --worktree value — the occurrence the backend actually honors.
func TestCCEntry_LastHonoredShortNameEntersMoaiTree(t *testing.T) {
	_, treePath, _ := setupRegisteredTreeRepo(t, "feat-login")
	rec := &wiringLaunchRecorder{}
	overrideUnifiedLaunch(t, rec)

	buf := new(strings.Builder)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	if err := ccCmd.RunE(ccCmd, []string{"-w", "fresh", "-w", "feat-login"}); err != nil {
		t.Fatalf("cc entry: %v", err)
	}
	if rec.args == nil {
		t.Fatal("launch seam never ran")
	}
	got, ok := lastBackendWorktreeValue(rec.args)
	if !ok {
		t.Fatalf("no --worktree option in the launched argv %v", rec.args)
	}
	if got != treePath {
		t.Fatalf("backend worktree value = %q, want the .moai tree %q — the last-honored -w must re-enter the registered tree, not twin it", got, treePath)
	}
}

// TestGLMEntry_LastHonoredShortNameEntersMoaiTree is the same last-honored
// resolution through the glm entry, which mirrors the cc flow.
func TestGLMEntry_LastHonoredShortNameEntersMoaiTree(t *testing.T) {
	_, treePath, _ := setupRegisteredTreeRepo(t, "feat-login")
	rec := &wiringLaunchRecorder{}
	overrideUnifiedLaunch(t, rec)

	buf := new(strings.Builder)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	if err := runGLM(glmCmd, []string{"-w", "fresh", "-w", "feat-login"}); err != nil {
		t.Fatalf("glm entry: %v", err)
	}
	if rec.args == nil {
		t.Fatal("launch seam never ran")
	}
	got, ok := lastBackendWorktreeValue(rec.args)
	if !ok {
		t.Fatalf("no --worktree option in the launched argv %v", rec.args)
	}
	if got != treePath {
		t.Fatalf("backend worktree value = %q, want the .moai tree %q", got, treePath)
	}
}

// TestCCEntry_LauncherStrippedSelectionEntersMoaiTree is the sync-audit delta
// D1 shape A end to end: `-w card --append-system-prompt -c -w fresh` with a
// registered .moai/worktrees/card — the launcher strips -c, the backend's
// required-value option consumes the second -w, and the session must enter
// the .moai tree the backend actually selects, never twin card.
func TestCCEntry_LauncherStrippedSelectionEntersMoaiTree(t *testing.T) {
	_, treePath, _ := setupRegisteredTreeRepo(t, "card")
	rec := &wiringLaunchRecorder{}
	overrideUnifiedLaunch(t, rec)

	buf := new(strings.Builder)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	if err := ccCmd.RunE(ccCmd, []string{"-w", "card", "--append-system-prompt", "-c", "-w", "fresh"}); err != nil {
		t.Fatalf("cc entry: %v", err)
	}
	if rec.args == nil {
		t.Fatal("launch seam never ran")
	}
	sel := selectBackendWorktree(consumeLaunchArgs(rec.args).PassThrough)
	if !sel.HasValue || sel.Value != treePath {
		t.Fatalf("backend selection over the launched argv = %+v, want the .moai tree %q", sel, treePath)
	}
}

// TestGLMEntry_BareWorktreeBeforeLauncherBooleanEntersMoaiTree is the
// sync-audit delta D1 shape B end to end through the glm entry: the launcher
// removes -c from between the bare -w and its would-be positional, the
// backend honors card, and the launched argv must select the registered
// .moai/worktrees/card.
func TestGLMEntry_BareWorktreeBeforeLauncherBooleanEntersMoaiTree(t *testing.T) {
	_, treePath, _ := setupRegisteredTreeRepo(t, "card")
	rec := &wiringLaunchRecorder{}
	overrideUnifiedLaunch(t, rec)

	buf := new(strings.Builder)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	if err := runGLM(glmCmd, []string{"-w", "-c", "card"}); err != nil {
		t.Fatalf("glm entry: %v", err)
	}
	if rec.args == nil {
		t.Fatal("launch seam never ran")
	}
	sel := selectBackendWorktree(consumeLaunchArgs(rec.args).PassThrough)
	if !sel.HasValue || sel.Value != treePath {
		t.Fatalf("backend selection over the launched argv = %+v, want the .moai tree %q", sel, treePath)
	}
}

// TestCCEntry_AttachedSpellingRefusedWhenRegistered is the sync-audit F1
// attached-spelling shape through the cc entry: `-wfeat-login` over a
// registered same-named .moai tree refuses the launch instead of letting the
// backend twin it.
func TestCCEntry_AttachedSpellingRefusedWhenRegistered(t *testing.T) {
	_, treePath, _ := setupRegisteredTreeRepo(t, "feat-login")
	rec := &wiringLaunchRecorder{}
	overrideUnifiedLaunch(t, rec)

	buf := new(strings.Builder)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	if err := ccCmd.RunE(ccCmd, []string{"-wfeat-login"}); err == nil {
		t.Fatalf("attached -w spelling over a registered .moai tree must refuse; launched %v", rec.args)
	}
	if rec.args != nil {
		t.Fatalf("launch seam ran despite the refusal: %v", rec.args)
	}
	if !strings.Contains(buf.String(), treePath) {
		t.Fatalf("refusal output %q does not name the registered tree", buf.String())
	}
}

// TestGLMEntry_AttachedSpellingRefusedWhenRegistered is the same refusal
// through the glm entry.
func TestGLMEntry_AttachedSpellingRefusedWhenRegistered(t *testing.T) {
	_, treePath, _ := setupRegisteredTreeRepo(t, "feat-login")
	rec := &wiringLaunchRecorder{}
	overrideUnifiedLaunch(t, rec)

	buf := new(strings.Builder)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	if err := runGLM(glmCmd, []string{"-wfeat-login"}); err == nil {
		t.Fatalf("attached -w spelling over a registered .moai tree must refuse; launched %v", rec.args)
	}
	if rec.args != nil {
		t.Fatalf("launch seam ran despite the refusal: %v", rec.args)
	}
	if !strings.Contains(buf.String(), treePath) {
		t.Fatalf("refusal output %q does not name the registered tree", buf.String())
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
