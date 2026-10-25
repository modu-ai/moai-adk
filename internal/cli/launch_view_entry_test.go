package cli

// Regression tests for card t1791 (GH #1813) and its audit rounds: the -w
// NAME form must enter an existing same-name .moai/worktrees/<name> MoAI
// tree instead of passing the name through to claude, whose short-name
// resolution covers only .claude/worktrees/<name> — the pass-through used to
// create a duplicate branch tree silently (the 2026-10-08 lane-12 incident).
// The entry judges the VIEW's backend-model selection (FinalArgv, option-
// model aware) and the backend argv loses every worktree selection on entry.
// Fixtures are scratch repositories under t.TempDir() with their own
// MOAI_HOME; the project-root seam is stubbed so nothing escapes the fixture.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// entryFixture builds a scratch primary repo with its own MOAI_HOME and
// points the project-root seam at it, returning the primary root. The repo
// must be a REAL git repository with a seed commit: worktreeroot.resolve
// derives the project key from the git common dir, and the entry validation
// requires candidates to be linked worktrees of this repository.
func entryFixture(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	primary := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(primary, 0o755); err != nil {
		t.Fatalf("mkdir primary: %v", err)
	}
	if out, err := exec.Command("git", "init", "-q", primary).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	for _, args := range [][]string{
		{"config", "user.email", "t1791@test.local"},
		{"config", "user.name", "t1791"},
		{"commit", "--allow-empty", "-q", "--allow-empty-message", "-m", "seed"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = primary
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	orig := findProjectRootFn
	findProjectRootFn = func() (string, error) { return primary, nil }
	t.Cleanup(func() { findProjectRootFn = orig })
	return primary
}

// mustLinkedTree creates the candidate as a LINKED worktree of the primary
// (card t1791 audit round 9 finding 2: the entry validates that the
// candidate belongs to the calling project's repository — a standalone repo
// at the candidate path is a different repository and is refused).
func mustLinkedTree(t *testing.T, primary, path string) {
	t.Helper()
	mustMkdirTree(t, filepath.Dir(path))
	cmd := exec.Command("git", "-C", primary, "worktree", "add", "-q", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("worktree add %s: %v\n%s", path, err, out)
	}
}

// mustMkdirTree creates a directory subtree, failing the test on error.
func mustMkdirTree(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

// entryDirFor resolves the entry directory for argv through the full view —
// the same shape the cc/glm flows judge.
func entryDirFor(t *testing.T, argv []string) string {
	t.Helper()
	return existingWorktreeEntryDir(resolveLaunchView(argv))
}

// canonical resolves symlinks (the macOS /var alias) so paths derived from
// the resolver's canonicalization agree with the paths asserted here.
func canonical(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("evalsymlinks %s: %v", p, err)
	}
	return r
}

// TestEntryDir_LegacyMoaiTreeIsEntered — the #1813 fix: a short name whose
// only existing candidate is .moai/worktrees/<name> answers that directory
// (the launcher enters it by directory and strips the selection), never a
// pass-through that would let claude create .claude/worktrees/<name>.
func TestEntryDir_LegacyMoaiTreeIsEntered(t *testing.T) {
	primary := entryFixture(t)
	name := "t1791"
	moaiTree := filepath.Join(primary, ".moai", "worktrees", name)
	mustLinkedTree(t, primary, moaiTree)
	if got, want := entryDirFor(t, []string{"-w", name}), canonical(t, moaiTree); got != want {
		t.Fatalf("entry dir = %q, want the existing .moai tree %q (a pass-through here creates a duplicate .claude branch tree — GH #1813)", got, want)
	}
}

// TestEntryDir_LegacyMoaiAndClaudeIsAmbiguous — both legacy roots carrying
// the name stay ambiguous: the entry answers "" and the ambiguity report
// belongs to the callers (REQ-WHR-022) — unchanged by the #1813 fix.
func TestEntryDir_LegacyMoaiAndClaudeIsAmbiguous(t *testing.T) {
	primary := entryFixture(t)
	name := "both"
	for _, rel := range []string{
		filepath.Join(".moai", "worktrees", name),
		filepath.Join(".claude", "worktrees", name),
	} {
		mustMkdirTree(t, filepath.Join(primary, rel))
	}
	if got := entryDirFor(t, []string{"-w", name}); got != "" {
		t.Fatalf("entry dir = %q, want \"\" for the two-root ambiguity", got)
	}
}

// TestEntryDir_LegacyClaudeStaysNative — a name whose only candidate is
// .claude/worktrees/<name> keeps answering "": claude re-enters that root
// natively, and the entry must not shadow it.
func TestEntryDir_LegacyClaudeStaysNative(t *testing.T) {
	primary := entryFixture(t)
	name := "native"
	mustMkdirTree(t, filepath.Join(primary, ".claude", "worktrees", name))
	if got := entryDirFor(t, []string{"-w", name}); got != "" {
		t.Fatalf("entry dir = %q, want \"\" for the legacy-claude native re-entry", got)
	}
}

// TestEntrySelection_StrippedOnEntry — the launcher-side half of the
// no-creation guarantee: when the entry resolves to a directory, the BACKEND
// argv handed to the launch carries no worktree selection at all, so claude
// can never resolve a short name against its own .claude/worktrees and
// create a duplicate tree.
func TestEntrySelection_StrippedOnEntry(t *testing.T) {
	primary := entryFixture(t)
	name := "t1791"
	mustLinkedTree(t, primary, filepath.Join(primary, ".moai", "worktrees", name))
	view := resolveLaunchView([]string{"-w", name})
	if existingWorktreeEntryDir(view) == "" {
		t.Fatal("precondition: the entry must resolve for the strip to be reachable")
	}
	rest, _ := stripWorktreeOptionsForBackend(view)
	for _, tok := range rest {
		if strings.HasPrefix(tok, "-w") || strings.HasPrefix(tok, "--worktree") || tok == name {
			t.Fatalf("worktree selection survived the strip: %v", rest)
		}
	}
}

// TestEntryDir_DuplicateSelectionEntersByFinal — a duplicate selection pair
// enters by the FINAL selection (the backend's own last-wins convention the
// view's Selection encodes; card t1791 audit round 2 finding 1): the first
// strip used to leave the second selection alive in the backend argv.
func TestEntryDir_DuplicateSelectionEntersByFinal(t *testing.T) {
	primary := entryFixture(t)
	second := filepath.Join(primary, ".moai", "worktrees", "second")
	mustLinkedTree(t, primary, second)
	got := entryDirFor(t, []string{"-w", "first", "-w", "second"})
	if got != canonical(t, second) {
		t.Fatalf("entry dir = %q, want the final selection's tree %q", got, canonical(t, second))
	}
}

// TestEntryRead_TargetsTheEnteredTree — the writer-precheck read of the
// entry flow names the ENTERED tree, never the view's first-match StageB
// read (card t1791 audit round 2 finding 1): under a duplicate pair the
// lock check must probe the tree that is about to be entered, or the
// concurrent-writer protection is bypassed.
func TestEntryRead_TargetsTheEnteredTree(t *testing.T) {
	primary := entryFixture(t)
	for _, name := range []string{"safe", "locked"} {
		mustLinkedTree(t, primary, filepath.Join(primary, ".moai", "worktrees", name))
	}
	view := resolveLaunchView([]string{"-w", "safe", "-w", "locked"})
	entryDir := existingWorktreeEntryDir(view)
	if entryDir == "" {
		t.Fatal("precondition: the entry must resolve to the final selection's tree")
	}
	read := entryWorktreeRead(entryDir)
	if !read.Has || !strings.Contains(read.Value, "locked") {
		t.Fatalf("precheck read = %+v, want the entered tree (locked), not StageB's %q", read, view.StageB.Value)
	}
}

// TestEntryDir_OptionValueIsNotASelection — a --worktree token that is the
// VALUE of another backend option is never a selection (card t1791 audit
// round 2 finding 2): the backend option model consumes it, the entry does
// not fire on it, and the backend argv passes through untouched.
func TestEntryDir_OptionValueIsNotASelection(t *testing.T) {
	primary := entryFixture(t)
	second := filepath.Join(primary, ".moai", "worktrees", "second")
	mustLinkedTree(t, primary, second)
	argv := []string{"-w", "first", "--append-system-prompt", "--worktree", "second"}
	view := resolveLaunchView(argv)
	if got := existingWorktreeEntryDir(view); got != "" {
		t.Fatalf("entry dir = %q, want \"\" — --worktree here is the prompt option's value, not a selection", got)
	}
	// And the non-entry flow hands the backend the untouched LaunchArgs (the
	// strip runs only on the entry path — cc.go/glm.go wire it there), so
	// the prompt option and its value survive the launch intact.
	joined := strings.Join(view.LaunchArgs, " ")
	if !strings.Contains(joined, "--append-system-prompt --worktree second") {
		t.Fatalf("the backend argv lost the prompt option or its value: %v", view.LaunchArgs)
	}
}

// TestStrip_FinalArgvLosesEverySelection — the backend argv after the
// launcher consumption loses every worktree selection on the entry path,
// including one the launcher's own dropped separator had sheltered (card
// t1791 audit round 2 finding 3): a surviving selection re-opens the
// nested-creation path after the directory entry.
func TestStrip_FinalArgvLosesEverySelection(t *testing.T) {
	primary := entryFixture(t)
	first := filepath.Join(primary, ".moai", "worktrees", "first")
	mustLinkedTree(t, primary, first)
	view := resolveLaunchView([]string{"-w", "first", "--", "--worktree=first"})
	if existingWorktreeEntryDir(view) == "" {
		t.Fatal("precondition: the entry must resolve")
	}
	rest, _ := stripWorktreeOptionsForBackend(view)
	for _, tok := range rest {
		if strings.Contains(tok, "worktree") || tok == "first" {
			t.Fatalf("a selection survived the backend-argv strip: %v", rest)
		}
	}
	// A non-selection tail token the backend owns survives the strip.
	view2 := resolveLaunchView([]string{"-w", "first", "--", "--keep", "prompt"})
	rest2, _ := stripWorktreeOptionsForBackend(view2)
	joined := strings.Join(rest2, " ")
	if !strings.Contains(joined, "--keep") || !strings.Contains(joined, "prompt") {
		t.Fatalf("backend-owned tail tokens lost: %v", rest2)
	}
}

// TestStrip_PreservesUserPermissionModeAndModel — the entry strip operates on
// the pre-consumption launch argv (card t1791 audit round 3 finding 1): the
// user's explicit --permission-mode and --model survive, so the launcher's
// consumption sees them and never injects the saved settings over them.
func TestStrip_PreservesUserPermissionModeAndModel(t *testing.T) {
	primary := entryFixture(t)
	mustLinkedTree(t, primary, filepath.Join(primary, ".moai", "worktrees", "plain"))
	view := resolveLaunchView([]string{"-w", "plain", "--permission-mode", "plan", "--model", "opus"})
	if existingWorktreeEntryDir(view) == "" {
		t.Fatal("precondition: the entry must resolve")
	}
	rest, _ := stripWorktreeOptionsForBackend(view)
	joined := strings.Join(rest, " ")
	if !strings.Contains(joined, "--permission-mode plan") {
		t.Fatalf("the user's explicit permission mode was lost: %v", rest)
	}
	if !strings.Contains(joined, "--model opus") {
		t.Fatalf("the user's explicit model was lost: %v", rest)
	}
}

// TestStrip_SecondSeparatorIsTheBackendTerminator — the launcher consumes
// only the FIRST separator (card t1791 audit round 3 finding 2): the second
// separator is the backend's own terminator and survives, so tokens past it
// are the backend's positionals — never re-parsed as a worktree selection.
func TestStrip_SecondSeparatorIsTheBackendTerminator(t *testing.T) {
	primary := entryFixture(t)
	mustLinkedTree(t, primary, filepath.Join(primary, ".moai", "worktrees", "plain"))
	view := resolveLaunchView([]string{"-w", "plain", "--", "--", "--worktree", "other"})
	if existingWorktreeEntryDir(view) == "" {
		t.Fatal("precondition: the entry must resolve")
	}
	// Assert the FINAL backend argv — the launcher's consumption eats the
	// first --, so the stripped LaunchArgs hold one more separator than the
	// backend will see.
	stripped, _ := stripWorktreeOptionsForBackend(view)
	backend := consumeLaunchArgs(stripped).PassThrough
	want := []string{"--", "--", "--", "--worktree", "other"}
	if len(backend) != len(want) {
		t.Fatalf("backend argv = %v, want %v (terminator + the exact original positionals)", backend, want)
	}
	for k := range want {
		if backend[k] != want[k] {
			t.Fatalf("backend argv = %v, want %v (token %d changed)", backend, want, k)
		}
	}
}

// TestStrip_ClusterCutKeepsTheLeadingLetters — a cluster token loses only its
// `w` and the attached value; the preceding letters are other options and
// survive the cut (card t1791 audit round 3 finding 3).

// TestEntryDir_NonGitDirectoryIsNotEntered — a candidate that is a plain
// directory (no git worktree behind it) is NOT isolation: entering it would
// run the backend in a subdirectory of the default checkout (card t1791
// audit round 5 finding 1). The entry answers "" for it.
func TestEntryDir_NonGitDirectoryIsNotEntered(t *testing.T) {
	primary := entryFixture(t)
	name := "plain"
	mustMkdirTree(t, filepath.Join(primary, ".moai", "worktrees", name))
	if got := entryDirFor(t, []string{"-w", name}); got != "" {
		t.Fatalf("entry dir = %q, want \"\" — the candidate is not a git worktree", got)
	}
}

// TestEntryDir_GitWorktreeIsEntered — the validation keeps a REAL linked
// git worktree of this repository enterable: a candidate linked to the
// primary (top level = the directory, same repository) answers its path.
func TestEntryDir_GitWorktreeIsEntered(t *testing.T) {
	primary := entryFixture(t)
	name := "real"
	tree := filepath.Join(primary, ".moai", "worktrees", name)
	mustLinkedTree(t, primary, tree)
	if got, want := entryDirFor(t, []string{"-w", name}), canonical(t, tree); got != want {
		t.Fatalf("entry dir = %q, want the real git worktree %q", got, want)
	}
}

// TestEntryStrip_TerminatorWallsThePrompt — the terminator construction
// preserves the prompt and neutralizes the tail (card t1791 audit round 5
// finding 2 and round 9 finding 1): removing [--worktree, plain] from
// [-w plain -- --debug -w plain PROMPT] leaves the backend
// [--, --, --debug, -w, plain, PROMPT] — the first -- is eaten by the
// launcher's consumption, the second ends option parsing before --debug, so
// the user's whole typed tail (--debug -w plain PROMPT) stays positional and
// the tail's selection syntax (-w plain, a re-selection the entry exists to
// prevent) is inert.
func TestEntryStrip_TerminatorWallsThePrompt(t *testing.T) {
	primary := entryFixture(t)
	mustLinkedTree(t, primary, filepath.Join(primary, ".moai", "worktrees", "plain"))
	view := resolveLaunchView([]string{"-w", "plain", "--", "--debug", "-w", "plain", "PROMPT"})
	if !entryStripIsSemanticallyClean(view) {
		t.Fatal("the terminator-walled backend argv must read clean — no worktree option survives the consumption")
	}
	stripped, _ := stripWorktreeOptionsForBackend(view)
	pos := positionalSubsequence(consumeLaunchArgs(stripped).PassThrough)
	want := []string{"--", "--debug", "PROMPT"}
	if len(pos) != len(want) {
		t.Fatalf("positionals = %v, want %v", pos, want)
	}
	for k := range want {
		if pos[k] != want[k] {
			t.Fatalf("positionals = %v, want %v (token %d changed)", pos, want, k)
		}
	}
}

// TestEntryStrip_CleanRemovalPasses — the guard keeps the clean shape
// flowing: a selection followed by a positional prompt removes without any
// role shift.
func TestEntryStrip_CleanRemovalPasses(t *testing.T) {
	primary := entryFixture(t)
	mustLinkedTree(t, primary, filepath.Join(primary, ".moai", "worktrees", "plain"))
	view := resolveLaunchView([]string{"-w", "plain", "--", "PROMPT"})
	if !entryStripIsSemanticallyClean(view) {
		stripped, kept := stripWorktreeOptionsForBackend(view)
		t.Fatalf("the clean selection-then-prompt shape must read clean: FinalArgv=%v Origin=%v stripped=%v kept=%v", view.FinalArgv, view.Consumed.Origin, stripped, kept)
	}
}

// TestEntryDir_SymlinkAliasOfPrimaryIsNotEntered — a candidate that is a
// symlink alias OF THE PRIMARY checkout resolves to the shared checkout
// (card t1791 audit round 6 finding 2): entering it would drop the isolation
// the -w asked for, so the entry answers "".
func TestEntryDir_SymlinkAliasOfPrimaryIsNotEntered(t *testing.T) {
	primary := entryFixture(t)
	alias := filepath.Join(primary, ".moai", "worktrees", "alias")
	if err := os.Symlink(primary, alias); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if got := entryDirFor(t, []string{"-w", "alias"}); got != "" {
		t.Fatalf("entry dir = %q, want \"\" — the alias resolves to the primary checkout itself", got)
	}
}

// TestEntryStrip_ClusterContinueSurvives — a -cw<name> cluster cut leaves
// the -c, which the consumption re-injects as Continue: the positional
// subsequences still match, so the guard reads it clean (card t1791 audit
// round 6 finding 3).
func TestEntryStrip_ClusterContinueSurvives(t *testing.T) {
	primary := entryFixture(t)
	mustLinkedTree(t, primary, filepath.Join(primary, ".moai", "worktrees", "plain"))
	view := resolveLaunchView([]string{"-cwplain"})
	if existingWorktreeEntryDir(view) == "" {
		t.Fatal("precondition: the entry must resolve")
	}
	if !entryStripIsSemanticallyClean(view) {
		t.Fatal("the -cw cluster cut (leaving -c for the consumption to re-inject) must read clean")
	}
}

// TestEntryDir_AliasRejectedWhenCalledFromALinkedWorktree — the alias check
// anchors on the REAL primary (resolved from the git common directory), not
// on findProjectRootFn's answer (card t1791 audit round 7 finding 1): when
// the launch runs inside a LINKED worktree, findProjectRootFn names that
// worktree, and an alias of the true primary must still be refused.
func TestEntryDir_AliasRejectedWhenCalledFromALinkedWorktree(t *testing.T) {
	primary := entryFixture(t)
	if out, err := exec.Command("git", "-C", primary, "commit", "--allow-empty", "-q", "--allow-empty-message", "-m", "seed").CombinedOutput(); err != nil {
		t.Fatalf("seed commit: %v\n%s", err, out)
	}
	linked := filepath.Join(t.TempDir(), "linked")
	if out, err := exec.Command("git", "-C", primary, "worktree", "add", "-q", linked).CombinedOutput(); err != nil {
		t.Fatalf("worktree add: %v\n%s", err, out)
	}
	mustMkdirTree(t, filepath.Join(primary, ".moai", "worktrees"))
	alias := filepath.Join(primary, ".moai", "worktrees", "alias")
	if err := os.Symlink(primary, alias); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	// The launch runs inside the LINKED tree: the project-root seam names it.
	orig := findProjectRootFn
	findProjectRootFn = func() (string, error) { return linked, nil }
	defer func() { findProjectRootFn = orig }()
	if got := entryDirFor(t, []string{"-w", "alias"}); got != "" {
		t.Fatalf("entry dir = %q, want \"\" — the alias is the primary checkout even when the launch runs from a linked worktree", got)
	}
}

// TestEntryDir_InheritedGitWorkTreeEnvCannotSpoofTheProbe — the top-level
// probe scrubs the inherited GIT_* environment (card t1791 audit round 7
// finding 2): GIT_WORK_TREE pointed at a plain non-git directory cannot make
// the probe report that directory as a worktree top level, so the entry
// stays "".
func TestEntryDir_InheritedGitWorkTreeEnvCannotSpoofTheProbe(t *testing.T) {
	primary := entryFixture(t)
	name := "plain"
	candidate := filepath.Join(primary, ".moai", "worktrees", name)
	mustMkdirTree(t, candidate)
	t.Setenv("GIT_WORK_TREE", candidate)
	if got := entryDirFor(t, []string{"-w", name}); got != "" {
		t.Fatalf("entry dir = %q, want \"\" — the inherited GIT_WORK_TREE must not pass the validation", got)
	}
}

// TestEntryDir_ForeignGitCommonDirCannotSpoofThePrimary — the primary
// resolution scrubs the inherited GIT_* environment (card t1791 audit round
// 8 finding 1): an inherited GIT_COMMON_DIR pointed at a foreign store
// cannot redirect the primary-root computation, so an alias of the primary
// stays refused even with the variable set.
func TestEntryDir_ForeignGitCommonDirCannotSpoofThePrimary(t *testing.T) {
	primary := entryFixture(t)
	alias := filepath.Join(primary, ".moai", "worktrees", "alias")
	if err := os.Symlink(primary, alias); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	foreign := t.TempDir() // a foreign store with no relation to the project
	t.Setenv("GIT_COMMON_DIR", filepath.Join(foreign, ".git"))
	if got := entryDirFor(t, []string{"-w", "alias"}); got != "" {
		t.Fatalf("entry dir = %q, want \"\" — the inherited GIT_COMMON_DIR must not redirect the primary resolution", got)
	}
}

// TestGitCommonDir_ScrubbedEnvIgnoresForeignOverride — the primary
// resolution's git probe scrubs the inherited GIT_* environment (card t1791
// audit round 8 finding 1): an inherited GIT_COMMON_DIR pointing at a
// foreign store cannot redirect gitCommonDirFn; the honest common dir of the
// queried repository comes back.
func TestGitCommonDir_ScrubbedEnvIgnoresForeignOverride(t *testing.T) {
	repo := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	foreign := filepath.Join(t.TempDir(), "foreign", ".git")
	t.Setenv("GIT_COMMON_DIR", foreign)
	got, err := gitCommonDirFn(repo)
	if err != nil {
		t.Fatalf("gitCommonDirFn: %v", err)
	}
	if !strings.Contains(got, filepath.Base(repo)) || strings.Contains(got, "foreign") {
		t.Fatalf("gitCommonDirFn = %q, want the queried repo's own common dir (the foreign GIT_COMMON_DIR override must be scrubbed)", got)
	}
}

func TestStrip_ClusterCutKeepsTheLeadingLetters(t *testing.T) {
	view := resolveLaunchView([]string{"--", "-pwplain", "hello"})
	// Assert the FINAL backend argv — the launcher's consumption eats the
	// first --, so the cluster cut's LaunchArgs carry one more separator
	// than the backend will see; the -p output mode must survive it.
	stripped, _ := stripWorktreeOptionsForBackend(view)
	backend := consumeLaunchArgs(stripped).PassThrough
	want := []string{"-p", "hello"}
	if len(backend) != len(want) {
		t.Fatalf("backend argv = %v, want %v (the -p output mode survives the w cut)", backend, want)
	}
	for k := range want {
		if backend[k] != want[k] {
			t.Fatalf("backend argv = %v, want %v (token %d changed)", backend, want, k)
		}
	}
}
