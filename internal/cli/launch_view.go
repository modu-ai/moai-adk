package cli

// launch_view.go — SPEC-LAUNCHER-ARGV-001 M1: the one canonical launch-argv
// resolver.
//
// Before this file, the launcher interpreted the same argv with several
// hand-written readers — the consumption loop in runLaunchClaude, the -w value
// scan, the --branch split, the chain-target extractor — and each of the six
// audit rounds of card t1604 found one more spelling on which two readers
// disagreed. This file is the single place that turns an argv into facts:
//
//   - the launcher consumption model (what runLaunchClaude eats and what it
//     passes through),
//   - the `--` separator meaning (the first one is the launcher's, the tail is
//     verbatim),
//   - the backend (claude) option model for the worktree selection (the LAST
//     option wins; the four spellings; the short-option cluster).
//
// The resolver and its walk helpers (worktreeFlagValue, normalizeWorktreeFlag,
// splitWorktreeBranchFlag, extractWorktreeLaunchTarget, selectBackendWorktree)
// are defined here and nowhere else. No other file on the launch path derives a
// worktree value from raw arguments: every reader receives a value this file
// produced. TestLaunchViewSingleParse scans for that.
//
// The view keeps TWO families of worktree facts apart, on purpose:
//
//   - the BACKEND selection (Selection): the last worktree option of the final
//     argv, decomposed by the backend's option model. Only the leader anchor
//     guard reads it — it must judge the tree claude will actually enter.
//   - the CONSUMER-scoped stage reads (StageA, StageB, Consumed.ChainTarget):
//     what the five non-guard readers read TODAY. They do not read one value —
//     they read the pipeline at three stages, each with its own first-match
//     rule (measured on the unrefactored tree: launcher_worktree_stage_cases_test.go).
//     SPEC §2 requires the non-leader behaviour to stay literally the same, so
//     each stage keeps its own field, computed by the same reader function it
//     always used, on the same stage input.

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	worktreepkg "github.com/modu-ai/moai-adk/internal/cli/worktree"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

// worktreeRead is one first-match -w / --worktree read: the value and whether a
// value was supplied (the result shape of worktreeFlagValue).
type worktreeRead struct {
	Value string
	Has   bool
}

// branchAnnotation is the launcher's own --branch consumption (card t295): the
// value the materialize step would check the new tree out at. Err carries a
// --branch parse failure, which the launch surfaces at the same point it always
// did (after the L2 prefix check).
type branchAnnotation struct {
	Present bool
	Branch  string
	Err     error
}

// launchConsumption is what runLaunchClaude's consumption loop makes of an argv.
// PassThrough is the backend-facing argv (the launcher-consumed options removed,
// the first separator dropped); the *Set fields say which launcher-owned options
// were present, with last-wins values.
type launchConsumption struct {
	PassThrough       []string
	PermissionMode    string
	PermissionModeSet bool
	Continue          bool
	Model             string
	ModelSet          bool
	// ChainTarget is the stage-C read: the worktree the chain-node populator
	// resolves from PassThrough.
	ChainTarget string
}

// worktreeSelection is the worktree the BACKEND would select from an argv.
type worktreeSelection struct {
	// Present reports that at least one worktree option stands in the argv.
	Present bool
	// Spelling names the token form that carried the selected (last) option.
	Spelling string
	// Value is its value; HasValue is false for a bare option (claude names the
	// tree itself) and for an empty equals form.
	Value    string
	HasValue bool
	// Index is the argv position of the selected option token.
	Index int
}

// Spellings recorded by selectBackendWorktree.
const (
	spellingLong       = "--worktree"
	spellingLongEquals = "--worktree="
	spellingShort      = "-w"
	spellingShortEqual = "-w="
	spellingAttached   = "-w<value>"
	spellingCluster    = "-<cluster>w<value>"
)

// launchView is the canonical resolution of one launch argv.
type launchView struct {
	// Entry is the argv the resolver was given.
	Entry []string

	// StageA is the first-match worktree read of Entry (the raw argv).
	StageA worktreeRead
	// Branch is the --branch consumption; Stripped is Entry without it.
	Branch   branchAnnotation
	Stripped []string
	// StageB is the first-match worktree read of Stripped.
	StageB worktreeRead
	// LaunchArgs is Stripped after normalizeWorktreeFlag: the argv handed to the
	// launch seam.
	LaunchArgs []string

	// Consumed is the consumption loop applied to LaunchArgs; FinalArgv is its
	// PassThrough — the backend argv without the launcher-injected
	// --permission-mode / --model / --continue flags.
	Consumed  launchConsumption
	FinalArgv []string

	// Selection is the backend-model worktree selection over FinalArgv.
	Selection worktreeSelection
}

// resolveLaunchView resolves entry once. It is pure: no filesystem, environment
// or process state is read.
//
// @MX:ANCHOR: [AUTO] the single launch-argv resolver; the cc and glm entries each call it twice (the anchor guard, then the launch readers)
// @MX:REASON: [AUTO] fan_in 4 (cc.go x2, glm.go x2); the guard and every worktree reader must judge the same view, so a change to the view shape or its walk rules moves all four sites together
// @MX:SPEC: SPEC-LAUNCHER-ARGV-001
func resolveLaunchView(entry []string) launchView {
	v, _ := resolveOwning(entry, "")
	return v
}

// resolveOwning is the one walk behind resolveLaunchView. With ownFlag set, the
// same walk also takes that launcher-owned flag where it stands (see
// walkLaunchArgs): the returned view is the view of the argv the backend
// receives — the flag already gone from FinalArgv when it was an override — and
// the returned indices say which tokens of entry were the overrides.
//
// The branch split and the worktree normalization do not leave the tokens alone:
// the split removes the --branch tokens, and the normalization rewrites
// `--worktree=<v>` / `-w=<v>` into two tokens, the second synthesized out of
// the value. Every token of LaunchArgs therefore carries its origin — the index
// of the typed token in entry it came from, -1 where the rewrite made it — so a
// token the user did not type as a standalone flag can never stand as the
// override, and a token is removed by its origin, not by counting equal ones.
func resolveOwning(entry []string, ownFlag string) (launchView, []int) {
	v := launchView{Entry: entry}
	v.StageA = readFirstWorktree(entry)

	rest, restOrigin, branch, has, err := splitWorktreeBranchFlagFrom(entry)
	v.Branch = branchAnnotation{Present: has, Branch: branch, Err: err}
	v.Stripped = entry
	strippedOrigin := []int(nil) // nil: v.Stripped is entry itself, origin is the identity
	if err == nil && has {
		v.Stripped, strippedOrigin = rest, restOrigin
	}
	v.StageB = readFirstWorktree(v.Stripped)
	var launchOrigin []int
	v.LaunchArgs, launchOrigin = normalizeWorktreeFlagFrom(v.Stripped)
	typed := make([]int, len(launchOrigin))
	for k, o := range launchOrigin {
		typed[k] = o
		if o >= 0 && strippedOrigin != nil {
			typed[k] = strippedOrigin[o]
		}
	}

	if err != nil {
		// A malformed --branch is refused when the launch materializes it, so there
		// is nothing to override: removing the flag must not turn an argv the launch
		// refuses into one it accepts.
		ownFlag = ""
	}
	var dropped []int
	v.Consumed, dropped = walkLaunchArgs(v.LaunchArgs, typed, ownFlag)
	v.FinalArgv = v.Consumed.PassThrough
	v.Selection = selectBackendWorktree(v.FinalArgv)
	return v, dropped
}

// materializeBranch runs the card-t295 --branch materialization the view
// annotates: nothing without --branch, the parse error when --branch was
// malformed, otherwise the worktree is created at the existing branch.
func (v launchView) materializeBranch(warn io.Writer) error {
	if v.Branch.Err != nil {
		return v.Branch.Err
	}
	if !v.Branch.Present {
		return nil
	}
	return materializeWorktreeBranch(v.StageB, v.Branch.Branch, warn)
}

// readFirstWorktree wraps worktreeFlagValue into the stage-read shape.
func readFirstWorktree(args []string) worktreeRead {
	value, ok := worktreeFlagValue(args)
	return worktreeRead{Value: value, Has: ok}
}

// existingWorktreeEntryDir reports the single existing home-root tree the
// launch argv's first -w short name names, or "" (REQ-WHR-006's entry
// clause, card t1640 review P1-2). The entry is BY DIRECTORY: the launcher
// starts the backend inside the tree and hands it no worktree selection —
// PM-3 measured claude 2.1.303 rejecting absolute --worktree values at
// argument validation (operator ruling bd72), so the absolute-path argv
// rewrite this used to be handed the backend a value it exits 1 on. A
// legacy-root match answers "" — claude re-enters legacy short names
// natively — and so does no match at all: under branch Y of D-5 the
// launcher creates nothing, and the short-name-without-existing-tree
// pass-through is the single documented exception to "no creation path
// names a legacy root" (REQ-WHR-006 branch Y, REQ-WHR-020). An ambiguous
// name (two or more roots) also answers ""; the ambiguity report belongs
// to the callers that enumerate candidates (REQ-WHR-022).
//
// Defined here next to the resolver walk helpers it reuses: worktree
// selection is derived in ONE file (launch_view.go) and every reader on the
// launch path receives values (TestLaunchViewSingleParse).
func existingWorktreeEntryDir(args []string) string {
	read := readFirstWorktree(args)
	if !read.Has || read.Value == "" || filepath.IsAbs(read.Value) {
		return ""
	}
	if read.Value == "." || read.Value == ".." || strings.ContainsAny(read.Value, `/\`) {
		return ""
	}
	root, err := findProjectRootFn()
	if err != nil {
		return ""
	}
	cands, err := worktreeroot.Candidates(root, read.Value)
	if err != nil || len(cands) != 1 || cands[0].Kind != worktreeroot.KindHome {
		return ""
	}
	return cands[0].Path
}

// refuseShadowedMoaiTreeName refuses the launch when the backend's worktree
// selection — the LAST -w value the view resolved, not the first flag —
// names an existing .moai/worktrees/<name> MoAI tree while the native
// .claude/worktrees/<name> tree does not exist: claude would resolve the
// name natively and silently create a duplicate branch tree beside the
// MoAI tree (the 2026-10-08 lane-12 incident, GH #1813). The refusal is
// pre-launch and argv-preserving — nothing is stripped, entered, or
// rewritten; the error guides into the existing tree by directory
// (`cd '<path>' && moai <verb>`, the path shellSingleQuoted — audit r1 P2-1
// measured claude 2.1.309 rejecting an absolute path as a --worktree value
// with "Invalid worktree name", so `-w <path>` is never guided). Both paths
// are spelled from the primary checkout resolved through the main worktree
// record — independent of the MoAI home metadata (audit r1 P2-2: a broken
// MOAI_HOME must not bypass the check) and of the git-dir layout (audit r2
// P2: a --separate-git-dir checkout is not the common-dir parent's child),
// and the probe failing refuses (fail closed) rather than proceeding into
// the duplicate-creation path. The native side is
// judged by its own existence (a native path aliasing the MoAI tree keeps
// the existing behavior). Every non-judgment answers nil: no selection or a
// bare one, an absolute or path-shaped value.
func refuseShadowedMoaiTreeName(view launchView, verb string) error {
	if !view.Selection.Present || !view.Selection.HasValue {
		return nil
	}
	value := view.Selection.Value
	if filepath.IsAbs(value) || value == "." || value == ".." || strings.ContainsAny(value, `/\`) {
		return nil
	}
	root, err := findProjectRootFn()
	if err != nil {
		return nil
	}
	primary, err := moaiPrimaryRoot(root)
	if err != nil {
		return fmt.Errorf("refusing -w %s: the repository root cannot be resolved for the same-name safety check (%v); run from inside the repository and retry", value, err)
	}
	moaiDir := filepath.Join(primary, ".moai", "worktrees", value)
	if fi, err := os.Stat(moaiDir); err != nil || !fi.IsDir() {
		return nil
	}
	nativeDir := filepath.Join(primary, ".claude", "worktrees", value)
	if _, err := os.Stat(nativeDir); err == nil {
		return nil // the native tree exists (aliases included); claude re-enters it natively
	}
	return fmt.Errorf("refusing -w %s: a MoAI worktree already exists at %s and no native .claude/worktrees/%s tree does; launching with -w %s would silently create a duplicate tree beside it (GH #1813). Enter the existing tree instead:\n  cd %s && moai %s",
		value, moaiDir, value, value, shellSingleQuoted(moaiDir), verb)
}

// moaiPrimaryRoot resolves the primary checkout's root from dir, by the
// git-dir identity first and the layout second (every step git-native,
// GIT_* scrubbed, fail closed):
//
//   - dir's git dir EQUALS the common dir — dir IS the main checkout,
//     whatever the git dir is called or where it lives (a normal .git, an
//     external --separate-git-dir metadata, one named ".git"; the basename
//     misjudged that last one, audit r3 P2). show-toplevel answers the
//     checkout.
//   - dir's git dir differs — dir is a LINKED worktree. The main checkout
//     is the checkout `git worktree list` names first, verified to own the
//     common dir as its .git entry (os.SameFile). An external git dir —
//     whatever its name, ".git" included — answers itself as the first
//     record (measured) and is refused: which checkout owns it is not
//     git-decidable, and guessing the metadata's parent launched into the
//     wrong tree (audit r2 P2, r4 P2-1).
func moaiPrimaryRoot(dir string) (string, error) {
	gitDir, err := moaiRevParse(dir, "--path-format=absolute", "--git-dir")
	if err != nil {
		return "", err
	}
	common, err := moaiRevParse(dir, "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	if gitDir == common {
		toplevel, err := moaiRevParse(dir, "--path-format=absolute", "--show-toplevel")
		if err != nil {
			return "", err
		}
		return moaiCanonical(toplevel), nil
	}
	list, err := moaiWorktreeList(dir)
	if err != nil {
		return "", err
	}
	if len(list) == 0 || list[0].Path == "" {
		return "", fmt.Errorf("no main worktree record for %s", dir)
	}
	candidate := moaiCanonical(filepath.FromSlash(list[0].Path))
	if filepath.Clean(candidate) == filepath.Clean(common) {
		return "", fmt.Errorf("the main worktree record is the git directory %s itself; the primary checkout cannot be resolved", common)
	}
	fiCommon, err := os.Stat(common)
	if err != nil {
		return "", err
	}
	fiDotGit, err := os.Stat(filepath.Join(candidate, ".git"))
	if err != nil || !fiDotGit.IsDir() || !os.SameFile(fiCommon, fiDotGit) {
		return "", fmt.Errorf("the git directory %s is not the .git of %s; the primary checkout cannot be resolved", common, candidate)
	}
	// An external git dir NAMED ".git" is git-indistinguishable from this
	// shape: git names the metadata's parent the main worktree (measured,
	// audit r4 P2-1) and SameFile agrees with itself. The checkout it
	// actually serves announces itself with a .git FILE linking directly to
	// the common dir — searched one level around the metadata's parent; a
	// linked worktree's .git file never links that directly (it carries the
	// worktrees/<name> path).
	if linked, ok := fileLinkedMainCheckout(common); ok {
		return linked, nil
	}
	return candidate, nil
}

// fileLinkedMainCheckout searches the directories one level around the
// common dir's parent for a .git FILE whose gitdir links DIRECTLY to
// common — the signature of a main checkout served by an external git dir
// (audit r4 P2-1). One such directory resolves it; none or several answer
// not-ok, which leaves the caller's structural answer standing.
func fileLinkedMainCheckout(common string) (string, bool) {
	grand := filepath.Dir(filepath.Dir(common))
	entries, err := os.ReadDir(grand)
	if err != nil {
		return "", false
	}
	found := ""
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dotGit := filepath.Join(grand, e.Name(), ".git")
		fi, err := os.Stat(dotGit)
		if err != nil || fi.IsDir() {
			continue
		}
		body, err := os.ReadFile(dotGit)
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(body)) != "gitdir: "+common {
			continue
		}
		if found != "" {
			return "", false // two checkouts claim it: undecidable
		}
		found = filepath.Join(grand, e.Name())
	}
	if found == "" {
		return "", false
	}
	return moaiCanonical(found), true
}

// moaiWorktreeList parses one `git -C dir worktree list --porcelain` probe,
// GIT_* scrubbed like the rev-parse probes.
func moaiWorktreeList(dir string) ([]worktreepkg.WorktreeRecord, error) {
	probe := exec.Command("git", "-C", dir, "worktree", "list", "--porcelain")
	probe.Env = scrubbedGitEnv()
	out, err := probe.Output()
	if err != nil {
		return nil, err
	}
	return worktreepkg.ParseWorktreeList(string(out))
}

// moaiRevParse runs one `git -C dir rev-parse` probe with the inherited
// GIT_* environment scrubbed — a planted GIT_WORK_TREE or GIT_DIR would
// redirect it — and returns the output with its single trailing newline
// removed. Only the newline: a checkout path may itself end in a space,
// and trimming it checked the wrong tree (audit r4 P2).
func moaiRevParse(dir string, args ...string) (string, error) {
	argv := append([]string{"-C", dir, "rev-parse"}, args...)
	probe := exec.Command("git", argv...)
	probe.Env = scrubbedGitEnv()
	out, err := probe.Output()
	if err != nil {
		return "", err
	}
	line := strings.TrimSuffix(strings.TrimSuffix(string(out), "\n"), "\r")
	if line == "" {
		return "", fmt.Errorf("git rev-parse %v answered empty in %s", args, dir)
	}
	return line, nil
}

// moaiCanonical resolves symlinks on a path that exists, passing the input
// through when it does not or cannot be resolved.
func moaiCanonical(p string) string {
	if real, err := filepath.EvalSymlinks(p); err == nil {
		return real
	}
	return p
}

// scrubbedGitEnv returns the process environment without any GIT_* variable.
func scrubbedGitEnv() []string {
	env := os.Environ()
	out := env[:0]
	for _, kv := range env {
		if strings.HasPrefix(kv, "GIT_") {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// stripWorktreeSelection removes the argv's first -w selection (both
// spellings, its value included): the entry-by-directory form starts the
// backend inside the tree, so the backend must receive no worktree
// selection at all — a surviving short name would resolve against the
// tree's own .claude/worktrees and create a duplicate inside it.
func stripWorktreeSelection(args []string) []string {
	read := readFirstWorktree(args)
	if !read.Has || read.Value == "" {
		return args
	}
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "--worktree="+read.Value {
			continue // the equals form carries its value
		}
		if (args[i] == "--worktree" || args[i] == "-w") && i+1 < len(args) && args[i+1] == read.Value {
			i++ // the two-token form: skip the flag and its value
			continue
		}
		out = append(out, args[i])
	}
	return out
}

// consumeLaunchArgs is runLaunchClaude's consumption loop: the launcher eats
// -b/--bypass, --permission-mode, -c/--continue, --model/-m (and the = forms)
// and drops the first `--` it does not eat as a value; everything else, and the
// whole tail after that separator, passes through verbatim. A second separator in
// the tail belongs to the backend and stays byte-for-byte intact.
func consumeLaunchArgs(extraArgs []string) launchConsumption {
	c, _ := walkLaunchArgs(extraArgs, nil, "")
	return c
}

// walkLaunchArgs is the consumption walk. consumeLaunchArgs runs it as is; the
// leader entry runs it with ownFlag, the launcher's own override flag
// (REQ-LA-005), so the override decision and the argv the backend receives are
// two outputs of ONE walk.
//
// The flag counts only where it stands: typed as a standalone token (typed[i] >= 0
// — a token the -w normalization synthesized out of an `=` form has -1 and never
// stands; nil typed reads every token as typed), in the launcher segment (before
// the first separator the walk did not eat as an option value) and not as the
// value of a backend option. The last two are read on what is left once the walk
// has dropped the launcher's own options — the sequence the backend reads — never
// on the typed argv, where a token's neighbour may be one the backend never sees.
// The flag is then dropped from PassThrough, and the returned indices name the
// dropped tokens by their typed index (typed[i], or i when typed is nil).
func walkLaunchArgs(extraArgs []string, typed []int, ownFlag string) (launchConsumption, []int) {
	var c launchConsumption
	var origin []int // origin[k]: the index in extraArgs of the launcher-segment token PassThrough[k]
	segment := -1    // len(PassThrough) at the first separator the walk did not eat
	for i := 0; i < len(extraArgs); i++ {
		arg := extraArgs[i]
		if arg == "--" {
			segment = len(c.PassThrough)
			c.PassThrough = append(c.PassThrough, extraArgs[i+1:]...)
			break
		}
		switch arg {
		case "-b", "--bypass":
			c.PermissionMode, c.PermissionModeSet = "bypassPermissions", true
		case "--permission-mode":
			if i+1 < len(extraArgs) {
				c.PermissionMode, c.PermissionModeSet = extraArgs[i+1], true
				i++
			}
		case "-c", "--continue":
			c.Continue = true
		case "--model", "-m":
			if i+1 < len(extraArgs) {
				c.Model, c.ModelSet = extraArgs[i+1], true
				i++
			}
		default:
			switch {
			case strings.HasPrefix(arg, "--model="):
				c.Model, c.ModelSet = strings.TrimPrefix(arg, "--model="), true
			case strings.HasPrefix(arg, "--permission-mode="):
				c.PermissionMode, c.PermissionModeSet = strings.TrimPrefix(arg, "--permission-mode="), true
			default:
				c.PassThrough = append(c.PassThrough, arg)
				origin = append(origin, i)
			}
		}
	}
	if segment < 0 {
		segment = len(c.PassThrough)
	}

	var dropped []int
	if ownFlag != "" {
		at := make([]int, segment) // at[k]: typed index of the launcher-segment token PassThrough[k]
		for k := range at {
			at[k] = origin[k]
			if typed != nil {
				at[k] = typed[origin[k]]
			}
		}
		if standing := standingOwnFlags(c.PassThrough[:segment], at, ownFlag); len(standing) > 0 {
			drop := make(map[int]bool, len(standing))
			for _, k := range standing {
				drop[k] = true
				dropped = append(dropped, at[k])
			}
			kept := make([]string, 0, len(c.PassThrough)-len(standing))
			for k, tok := range c.PassThrough {
				if !drop[k] {
					kept = append(kept, tok)
				}
			}
			c.PassThrough = kept
		}
	}
	c.ChainTarget = extractWorktreeLaunchTarget(c.PassThrough)
	return c, dropped
}

// standingOwnFlags returns the positions in segment — the launcher segment as the
// backend reads it — where flag stands as an option of its own: typed as a
// standalone token (at[i] >= 0, the typed index) and not taken by the backend
// option model as another option's value.
func standingOwnFlags(segment []string, at []int, flag string) []int {
	var standing []int
	for i := 0; i < len(segment); i++ {
		tok := segment[i]
		if tok == flag && at[i] >= 0 {
			standing = append(standing, i)
			continue
		}
		if len(tok) < 2 || tok[0] != '-' {
			continue // a positional: no option to read
		}
		_, span := backendOptionStep(segment, i, worktreeSelection{})
		i += span
	}
	return standing
}

// consumeAllowCardAnchor is the launcher's consumption of its own override flag
// (REQ-LA-005): it removes --allow-card-anchor from args where it stands and
// reports whether it stood anywhere. The decision is made by the walk that
// produces the exec argv (walkLaunchArgs inside resolveOwning), so it follows
// what the backend will actually read: a flag after the launcher's separator is
// the backend's, and a flag that is a backend option's value — after the
// launcher has dropped its own options — is that option's value. Both stay in
// the argv.
func consumeAllowCardAnchor(args []string) (rest []string, allowed bool) {
	_, dropped := resolveOwning(args, allowCardAnchorFlag)
	drop := make(map[int]bool, len(dropped))
	for _, i := range dropped {
		drop[i] = true
	}
	rest = make([]string, 0, len(args))
	for i, a := range args {
		if !drop[i] {
			rest = append(rest, a)
		}
	}
	return rest, len(dropped) > 0
}

// --- backend option model: the worktree selection ----------------------------

// selectBackendWorktree returns the worktree the backend selects from argv: the
// LAST worktree option wins. It reads the four spellings (`--worktree v`,
// `--worktree=v`, `-w v`, attached `-wv`) and a short-option cluster in which a
// boolean precedes the value option (`-pwv`: `p` is a boolean, `w` takes the
// rest as its value). An optional-value option's value is the next token unless
// that token starts with `-` or is `--`; a `--` ends option parsing (everything
// after it is positional) unless it is the VALUE of a required-value option —
// a required-value option takes the next token even when it is flag-shaped or
// `--` (measured on claude 2.1.302: `claude --append-system-prompt -- --version`
// prints the version), and parsing continues after it. A token that is the
// value of another option is not an option. The knowledge of which options take
// values is activeClaudeOptionModel, the repo's one measured claude option
// surface. What follows an optional-value worktree option's `--` is unmeasured
// and unchanged.
func selectBackendWorktree(argv []string) worktreeSelection {
	var sel worktreeSelection
	for i := 0; i < len(argv); i++ {
		tok := argv[i]
		if tok == argSeparator {
			break // the backend's option terminator
		}
		if len(tok) < 2 || tok[0] != '-' {
			continue // a positional (a prompt, or a value already consumed)
		}
		var span int
		sel, span = backendOptionStep(argv, i, sel)
		i += span
	}
	return sel
}

// backendOptionStep reads the single option the flag-shaped token argv[i] starts,
// by the backend's option model: the worktree selection after it (cur unless the
// token is a worktree option) and how many following tokens it takes as its value.
// It is the one step both the worktree selection and the override-flag position
// walk, so a token is a value to one exactly when it is to the other.
func backendOptionStep(argv []string, i int, cur worktreeSelection) (worktreeSelection, int) {
	tok := argv[i]
	switch {
	case tok == spellingLong || tok == spellingShort:
		sel := worktreeSelection{Present: true, Spelling: tok, Index: i}
		if v, ok := optionalValueAfter(argv, i); ok {
			sel.Value, sel.HasValue = v, true
			return sel, 1
		}
		return sel, 0
	case strings.HasPrefix(tok, spellingLongEquals):
		v := strings.TrimPrefix(tok, spellingLongEquals)
		return worktreeSelection{Present: true, Spelling: spellingLongEquals, Value: v, HasValue: v != "", Index: i}, 0
	case strings.HasPrefix(tok, spellingShortEqual):
		v := strings.TrimPrefix(tok, spellingShortEqual)
		return worktreeSelection{Present: true, Spelling: spellingShortEqual, Value: v, HasValue: v != "", Index: i}, 0
	case strings.HasPrefix(tok, "--"):
		return cur, longOptionValueSpan(argv, i)
	default:
		return scanShortCluster(argv, i, cur)
	}
}

// optionalValueAfter returns the token after argv[i] when it can be an
// optional-value option's value: present, not `--`, not flag-shaped.
func optionalValueAfter(argv []string, i int) (string, bool) {
	if i+1 < len(argv) && argv[i+1] != argSeparator && !strings.HasPrefix(argv[i+1], "-") {
		return argv[i+1], true
	}
	return "", false
}

// longOptionValueSpan reports how many following tokens a long option
// (no `=`) consumes as its value, by the claude option model.
func longOptionValueSpan(argv []string, i int) int {
	if strings.Contains(argv[i], "=") {
		return 0
	}
	switch activeClaudeOptionModel[argv[i]] {
	case claudeOptionRequiredValue:
		// A required-value option takes the next token as its value even when it
		// is `--` (measured on claude 2.1.302), and parsing continues after it.
		if i+1 < len(argv) {
			return 1
		}
	case claudeOptionOptionalValue:
		if _, ok := optionalValueAfter(argv, i); ok {
			return 1
		}
	case claudeOptionBoolean, claudeOptionUnknown:
	}
	return 0
}

// scanShortCluster decomposes the single-dash token argv[i] the way the
// backend's parser does: letters are read left to right; a boolean is skipped;
// the worktree letter `w` takes the rest of the token as its value (or the next
// token when nothing is attached); any other value-taking or unknown letter
// ends the walk, because the rest of the token is that option's value. It
// returns the (possibly updated) selection and how many following tokens the
// token consumed as a value.
func scanShortCluster(argv []string, i int, cur worktreeSelection) (worktreeSelection, int) {
	tok := argv[i]
	for j := 1; j < len(tok); j++ {
		letter := tok[j]
		attached := tok[j+1:]
		if letter == 'w' {
			spelling := spellingAttached
			if j > 1 {
				spelling = spellingCluster
			}
			if attached != "" {
				attached = strings.TrimPrefix(attached, "=")
				return worktreeSelection{Present: true, Spelling: spelling, Value: attached, HasValue: attached != "", Index: i}, 0
			}
			sel := worktreeSelection{Present: true, Spelling: spellingShort, Index: i}
			if j > 1 {
				sel.Spelling = spellingCluster
			}
			if v, ok := optionalValueAfter(argv, i); ok {
				sel.Value, sel.HasValue = v, true
				return sel, 1
			}
			return sel, 0
		}
		switch activeClaudeOptionModel["-"+string(letter)] {
		case claudeOptionBoolean:
			continue
		case claudeOptionRequiredValue:
			if attached == "" && i+1 < len(argv) {
				return cur, 1 // the next token, `--` included, is the value
			}
			return cur, 0
		case claudeOptionOptionalValue:
			if attached == "" {
				if _, ok := optionalValueAfter(argv, i); ok {
					return cur, 1
				}
			}
			return cur, 0
		case claudeOptionUnknown:
			return cur, 0
		}
	}
	return cur, 0
}

// --- first-match readers and token rewriters (moved here unchanged) -----------

// normalizeWorktreeFlag rewrites -w/--worktree[=name] forms (before any "--"
// pass-through marker) into the canonical two-token "--worktree [name]" form
// that Claude Code accepts, preserving argument order. Tokens after "--" are
// left untouched — they are already verbatim pass-through to claude.
// The name is optional: a bare -w lets claude auto-generate a worktree name.
func normalizeWorktreeFlag(args []string) []string {
	normalized, _ := normalizeWorktreeFlagFrom(args)
	return normalized
}

// normalizeWorktreeFlagFrom is normalizeWorktreeFlag that also returns, for each
// token of the result, the index in args of the token it came from — -1 for the
// name an `=` form was split into, which nobody typed as a token of its own.
func normalizeWorktreeFlagFrom(args []string) (normalized []string, origin []int) {
	normalized = make([]string, 0, len(args)+1)
	origin = make([]int, 0, len(args)+1)
	emit := func(tok string, from int) {
		normalized = append(normalized, tok)
		origin = append(origin, from)
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			// Everything after -- is verbatim pass-through.
			for j := i; j < len(args); j++ {
				emit(args[j], j)
			}
			break
		}
		switch {
		case arg == "-w" || arg == "--worktree":
			emit("--worktree", i)
			// Optional value: consume the next token unless it is another flag
			// or the pass-through marker.
			if i+1 < len(args) && args[i+1] != "--" && !strings.HasPrefix(args[i+1], "-") {
				emit(args[i+1], i+1)
				i++
			}
		case strings.HasPrefix(arg, "--worktree="):
			name := strings.TrimPrefix(arg, "--worktree=")
			emit("--worktree", i)
			if name != "" {
				emit(name, -1)
			}
		case strings.HasPrefix(arg, "-w="):
			name := strings.TrimPrefix(arg, "-w=")
			emit("--worktree", i)
			if name != "" {
				emit(name, -1)
			}
		default:
			emit(arg, i)
		}
	}

	return normalized, origin
}

// worktreeFlagValue scans args (stopping at the "--" pass-through marker) for
// the value of the -w / --worktree flag. Returns the value and whether a
// value was supplied. Mirrors the token-shape handling of normalizeWorktreeFlag
// without mutating args. The FIRST match decides.
func worktreeFlagValue(args []string) (string, bool) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			// Everything after -- is verbatim pass-through — do not scan.
			return "", false
		}
		switch {
		case arg == "-w" || arg == "--worktree":
			// Optional value: the next token is the value unless it is another
			// flag or the pass-through marker.
			if i+1 < len(args) && args[i+1] != "--" && !strings.HasPrefix(args[i+1], "-") {
				return args[i+1], true
			}
			// Bare -w (claude auto-generates a worktree name): no value.
			return "", false
		case strings.HasPrefix(arg, "--worktree="):
			return strings.TrimPrefix(arg, "--worktree="), true
		case strings.HasPrefix(arg, "-w="):
			return strings.TrimPrefix(arg, "-w="), true
		}
	}
	return "", false
}

// splitWorktreeBranchFlag finds and removes the --branch flag from args,
// scanning only tokens before the "--" pass-through marker (tokens after it
// are verbatim backend input and are never touched). Accepts both the
// two-token `--branch <value>` and the joined `--branch=<value>` forms. A
// missing or option-shaped value is a parse error at this boundary — an
// unconsumed option-shaped token must never leak into the backend argv.
func splitWorktreeBranchFlag(args []string) (rest []string, branch string, has bool, err error) {
	rest, _, branch, has, err = splitWorktreeBranchFlagFrom(args)
	return rest, branch, has, err
}

// splitWorktreeBranchFlagFrom is splitWorktreeBranchFlag that also returns, for
// each token of rest, its index in args.
func splitWorktreeBranchFlagFrom(args []string) (rest []string, origin []int, branch string, has bool, err error) {
	rest = make([]string, 0, len(args))
	origin = make([]int, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			for j := i; j < len(args); j++ {
				rest = append(rest, args[j])
				origin = append(origin, j)
			}
			return rest, origin, branch, has, nil
		}
		switch {
		case arg == "--branch":
			has = true
			if i+1 >= len(args) || args[i+1] == "--" || strings.HasPrefix(args[i+1], "-") {
				return nil, nil, "", true, fmt.Errorf("--branch requires a branch name")
			}
			branch = args[i+1]
			i++
		case strings.HasPrefix(arg, "--branch="):
			has = true
			branch = strings.TrimPrefix(arg, "--branch=")
			if branch == "" {
				return nil, nil, "", true, fmt.Errorf("--branch requires a branch name")
			}
		default:
			rest = append(rest, arg)
			origin = append(origin, i)
		}
	}
	return rest, origin, branch, has, nil
}

// extractWorktreeLaunchTarget returns the -w/--worktree value from normalized
// launch args. It understands the canonical two-token "--worktree <name>" form
// (the output of normalizeWorktreeFlag) and the "--worktree=<name>" spelling.
// A bare --worktree (claude auto-names the worktree) and flags after the "--"
// pass-through marker return "".
func extractWorktreeLaunchTarget(args []string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			return ""
		}
		switch {
		case args[i] == "--worktree":
			if i+1 < len(args) && args[i+1] != "--" && !strings.HasPrefix(args[i+1], "-") {
				return args[i+1]
			}
			return ""
		case strings.HasPrefix(args[i], "--worktree="):
			return strings.TrimPrefix(args[i], "--worktree=")
		}
	}
	return ""
}

// --- entry points for callers outside the cc/glm launch path ------------------
//
// The codex launcher and the existing unit tests hand an argv straight to a
// reader. These wrappers derive the stage value HERE, then call the reader's
// value-taking core, so the derivation stays in this file.

// resolveWorktreeL2Path validates the first -w/--worktree value of args against
// the accepted worktree prefixes (validateWorktreeL2Value).
func resolveWorktreeL2Path(args []string, warn io.Writer) error {
	return validateWorktreeL2Value(readFirstWorktree(args), warn)
}

// resolveWorktreeExistingBranch scans args for a --branch flag accompanying a
// short-name -w, materializes the worktree at the existing branch, and returns
// the args with the --branch tokens removed. When no --branch flag is present
// the args are returned unchanged and nothing runs.
func resolveWorktreeExistingBranch(args []string, warn io.Writer) ([]string, error) {
	rest, branch, has, err := splitWorktreeBranchFlag(args)
	if err != nil {
		return nil, err
	}
	if !has {
		return args, nil
	}
	if err := materializeWorktreeBranch(readFirstWorktree(rest), branch, warn); err != nil {
		return nil, err
	}
	return rest, nil
}

// seedAdmittedWorktreeHooks seeds the tree the first -w value of args names.
func seedAdmittedWorktreeHooks(args []string, warn io.Writer) {
	seedAdmittedWorktreeHooksFor(readFirstWorktree(args), warn)
}

// injectChainNodeForLaunch records the chain node for the worktree the
// pass-through args name (injectChainNodeForTarget).
func injectChainNodeForLaunch(extraArgs []string, env []string, warn io.Writer) []string {
	return injectChainNodeForTarget(extractWorktreeLaunchTarget(extraArgs), env, warn)
}
