package cli

// SPEC-CODEX-GATE-SCOPE-001 — the codex review gate's review-target scoping.
//
// The gate must not review changes it cannot attribute (spec.md §0). The
// scope discriminator (REQ-CGS-004) decides from ONE input only — the session
// tree's current branch carrying the WT- prefix, the committed lane-protocol
// invariant (kanban-dispatch § Isolation, gitflow-lane-protocol §1). Session
// env labels are observation context, never decision inputs (REQ-CGS-004,
// decision-index Q2 CONFIRMED): stale labels survive /clear (card t1373) and
// the label namespace is being reworked (t1378), so consuming a value here
// would couple this code to both. No lane-label format matching exists — the
// discriminator reads the branch name's prefix and nothing else.
//
// Fail-open direction (decision-index Q5 CONFIRMED): an unidentified session
// — detached HEAD, a plain branch, a non-git dir, or a WT- branch whose merge
// base cannot be computed — falls to TREE scope and keeps today's behavior.
// The narrower card-scope behavior activates on card evidence only.
//
// [HARD] (plan §C) BOTH execution paths — HandleCodexReviewGate and the
// receipt producer produceCodexReviewReceipt (plus the Codex Stop chain's
// member 6) — call the ONE reviewScopeResolver. Two discriminators could
// disagree for the same session state, and REQ-CGS-009's parity would break.
//
// @MX:SPEC: SPEC-CODEX-GATE-SCOPE-001

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/execerr"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// Scope classes (REQ-CGS-001): card-scope (the session tree is a WT- card
// worktree) and tree-scope (today's whole-uncommitted-tree behavior).
const (
	reviewScopeCard = "card"
	reviewScopeTree = "tree"
)

// cardBaseBranch is the gitflow integration branch the card diff is measured
// against; cardBranchPrefix is the committed lane-protocol invariant the
// discriminator's primary signal rides (REQ-CGS-004).
const (
	cardBaseBranch   = "develop"
	cardBranchPrefix = "WT-"
)

// cardDigestHexLen mirrors internal/verify's binding-digest width so the card
// receipt key reads like every other receipt key ("<head>:<digest[:16]>").
const cardDigestHexLen = 16

// reviewScope is the resolved scope of one session state: the class, the log
// basis (branch match / none — REQ-CGS-010), the tree the scope resolves from,
// the branch, and — card-scope only — the merge base recomputed at resolution
// time (never a pinned SHA, gitflow-lane-protocol §8). Primary carries the
// primary-checkout determination (SPEC-CODEX-GATE-SCOPING-001 REQ-CGSC-001):
// the session tree's git dir is the repository's common git dir. It is
// resolved only on the non-card tree arms — the card class and a WT- branch
// with an unavailable base carry card evidence, so the primary policy never
// evaluates them (REQ-CGSC-005, REQ-CRO-004).
type reviewScope struct {
	Class     string
	Basis     string
	Dir       string
	Branch    string
	MergeBase string
	Primary   bool
}

// reviewScopeResolver is the injectable scope seam (REQ-CGS-001, the
// reviewGateChangeDetector precedent): tests drive the discriminator with no
// live codex, and both execution paths share this one variable.
var reviewScopeResolver = resolveReviewScope

// primaryScopeDetector is the injectable primary-checkout seam (the
// reviewScopeResolver precedent, SPEC-CODEX-GATE-SCOPING-001 REQ-CGSC-001):
// tests drive the git-dir/git-common-dir comparison with no git probe.
var primaryScopeDetector = isPrimaryCheckoutGit

// resolveReviewScope classifies a session tree. The decision chain is branch
// detection first (REQ-CGS-004), then the merge base (REQ-CGS-002); every
// failure arm falls to tree scope with the reason carried in Basis — the
// reason is observability, never a block path (spec.md §F).
func resolveReviewScope(sessionDir string) reviewScope {
	if sessionDir == "" {
		return reviewScope{Class: reviewScopeTree, Basis: "no session tree", Dir: ""}
	}
	branch, err := reviewScopeGit(sessionDir, "branch", "--show-current")
	if err != nil || branch == "" {
		// An unreadable tree or a detached HEAD (empty branch output): the
		// documented tree-scope fail-open (decision-index Q5 CONFIRMED). The
		// primary axis is still decidable on a readable-but-detached tree —
		// the determination is branch-independent (git-dir identity, REQ-CGSC-001).
		s := reviewScope{Class: reviewScopeTree, Basis: "no card branch (unreadable or detached)", Dir: sessionDir}
		s.Primary = primaryScopeDetector(sessionDir)
		return s
	}
	if !cardScopeFromBranch(branch) {
		s := reviewScope{Class: reviewScopeTree, Basis: "no card branch: " + branch, Dir: sessionDir, Branch: branch}
		s.Primary = primaryScopeDetector(sessionDir)
		return s
	}
	base, err := cardMergeBase(sessionDir)
	if err != nil {
		// A WT- branch whose base cannot be computed keeps today's behavior;
		// the merge base failure reason rides the log basis (REQ-CGS-010).
		// Primary stays false: card evidence, never primary-skipped (REQ-CRO-004).
		return reviewScope{Class: reviewScopeTree, Basis: "card branch detected but merge base unavailable: " + err.Error(), Dir: sessionDir, Branch: branch}
	}
	return reviewScope{Class: reviewScopeCard, Basis: "branch match: " + branch, Dir: sessionDir, Branch: branch, MergeBase: base}
}

// isPrimaryCheckoutGit reports whether dir is the repository's primary working
// tree: the checkout whose git dir is the repository's common git dir
// (REQ-CGSC-001). Fail-open in the PRESERVE direction: an empty dir, a non-git
// directory, or a git error reads false — not primary — so the
// primary-checkout skip never applies to an undecidable tree and today's
// behavior for that state is kept (REQ-CGSC-006). Both git outputs may be
// relative to dir, so each is resolved against dir before the comparison.
func isPrimaryCheckoutGit(dir string) bool {
	if dir == "" {
		return false
	}
	gitDir, err := reviewScopeGit(dir, "rev-parse", "--git-dir")
	if err != nil || gitDir == "" {
		return false
	}
	commonDir, err := reviewScopeGit(dir, "rev-parse", "--git-common-dir")
	if err != nil || commonDir == "" {
		return false
	}
	return reviewScopeGitPath(dir, gitDir) == reviewScopeGitPath(dir, commonDir)
}

// reviewScopeEvalPath resolves symlinks in one resolved git-path candidate,
// returning it unchanged when the resolution fails (a missing final
// component, a broken link) — both sides of the comparison degrade the same
// way, so the equality stays meaningful (card-review repair round 2, N4).
func reviewScopeEvalPath(p string) string {
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	return p
}

// reviewScopeGitPath resolves one git subcommand's possibly-relative path
// output against the tree it was produced in, so the equality check compares
// LOCATIONS rather than spellings. The tree itself is symlink-resolved before
// the join (card-review repair round 2, N4): through a symlinked subdirectory
// git reports --git-dir as the REAL absolute path while --git-common-dir
// comes back relative to the link, and joining the relative output against
// the unresolved link spells a location that differs only by the link.
func reviewScopeGitPath(dir, out string) string {
	if filepath.IsAbs(out) {
		return reviewScopeEvalPath(filepath.Clean(out))
	}
	return reviewScopeEvalPath(filepath.Clean(filepath.Join(reviewScopeEvalPath(dir), out)))
}

// cardScopeFromBranch is the pure decision core of the discriminator: a card
// branch is a branch carrying the WT- prefix. Env labels are absent from this
// signature BY CONSTRUCTION — the decision cannot read them (REQ-CGS-004).
func cardScopeFromBranch(branch string) bool {
	return strings.HasPrefix(branch, cardBranchPrefix)
}

// reviewScopeGit runs one read-only git subcommand in dir. StatusDetail, not
// %w: a raw *exec.ExitError chain would read as an intentional ExitCoder at
// the cmd seam (the internal/verify precedent).
func reviewScopeGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), execerr.StatusDetail(err))
	}
	return strings.TrimSpace(string(out)), nil
}

// cardMergeBase computes the card diff base at evaluation time — never a
// pinned SHA (REQ-CGS-002, gitflow-lane-protocol §8: an absorption must move
// the base, and only a recompute follows it).
func cardMergeBase(dir string) (string, error) {
	// The ONE ancestry read in this file — the card-diff base measurement, not
	// a binary-lag comparison (REQ-ABI-006: binlag.Evaluate stays the only
	// binary-lag judge; the sweep allowlist pins this coordinate).
	base, err := reviewScopeGit(dir, "merge-base", cardBaseBranch, "HEAD")
	if err != nil {
		return "", fmt.Errorf("card merge base: %s", execerr.StatusDetail(err))
	}
	return base, nil
}

// reviewScopeSessionDir picks the tree the scope resolves from: the SESSION's
// working directory first (the CWD arm — REQ-CGS-005; a spawn-frozen
// CLAUDE_PROJECT_DIR naming a different tree must not win), then the legacy
// payload field, then the caller-resolved projectDir (today's behavior for
// payloads carrying no directory at all).
func reviewScopeSessionDir(input *hook.HookInput, projectDir string) string {
	if input != nil {
		if input.CWD != "" {
			return input.CWD
		}
		if input.ProjectDir != "" {
			return input.ProjectDir
		}
	}
	return projectDir
}

// reviewRequestParams assembles the review request for a resolved scope.
//
// Tree scope keeps the pre-SPEC shape byte-identical (REQ-CGS-003, the
// REQ-CRT-006 line): target "uncommittedChanges" + the resolved tree as cwd.
//
// Card scope names the card diff: a baseBranch target pinned to the
// RECOMPUTED merge base — the card tree's `git diff` against that base is
// exactly the confirmed union measurement (committed leg ∪ uncommitted leg,
// decision-index Q4; plan §B.1 form (가)) — with cwd = the card worktree, so
// codex reviews that tree (REQ-CGS-005). buildCodexReviewTarget passes an
// already-tagged target map through untouched, and if codex rejects a SHA in
// the branch field the review call errors and the gate fails OPEN
// (REQ-CGS-008) — the failure direction is the documented one.
func reviewRequestParams(scope reviewScope) map[string]any {
	if scope.Class == reviewScopeCard {
		return map[string]any{
			"target": map[string]any{"type": codexTargetBaseBranch, "branch": scope.MergeBase},
			"cwd":    scope.Dir,
		}
	}
	return map[string]any{
		"target": codexTargetUncommitted,
		"cwd":    scope.Dir,
	}
}

// reviewGateScopedChangeDetector is the self-gate over a RESOLVED scope
// (REQ-CGS-006): the card class measures the card diff; every other class
// keeps the pre-SPEC detector over the scope's tree, so the existing seam and
// its tests are unchanged on the tree path.
//
// The tree class composes one more exclusion (SPEC-CODEX-GATE-SCOPING-001
// REQ-CGSC-007, card-review repair R1): a turn whose ONLY changes are the
// runtime-config surfaces is not reviewable. The consult lives HERE — the
// tree-scope-only caller — never inside the shared reviewableFromPorcelain the
// multi-review gates also consume: a card or a multi-review turn over
// config-only changes keeps full reviewability (REQ-CGSC-005 / AC-CGSC-009).
// The probe can only narrow the detector's answer, never widen it.
func reviewGateScopedChangeDetector(scope reviewScope) bool {
	if scope.Class == reviewScopeCard {
		return hasReviewableCardChanges(scope)
	}
	if !reviewGateChangeDetector(scope.Dir) {
		return false
	}
	return !treeConfigOnlyChanges(scope.Dir)
}

// cardChangedPaths returns the changed paths of the card diff: the union
// measurement (decision-index Q4) — tracked changes from a name-only diff
// against the merge base (committed leg + uncommitted tracked
// leg, plan §B.1 form (가)) plus the untracked leg from `git status
// --porcelain`, runtime-managed prefixes excluded (REQ-CGS-002). The prefix
// filter applied to the tracked list is deliberately the same one the tree
// detector uses: a card commit under a hook-written state prefix is session
// churn, not card work.
func cardChangedPaths(scope reviewScope) ([]string, error) {
	nameOnly, err := reviewScopeGit(scope.Dir, "diff", "--name-only", scope.MergeBase)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var paths []string
	add := func(path string) {
		if path == "" || seen[path] || isRuntimeManagedPath(path) {
			return
		}
		seen[path] = true
		paths = append(paths, path)
	}
	for _, p := range strings.Split(nameOnly, "\n") {
		add(strings.TrimSpace(p))
	}
	// The untracked leg is read at FILE level (ls-files, the same source
	// cardScopeKeyParts digests), not from `status --porcelain`: porcelain
	// collapses a fully-untracked directory to `?? .moai/`, and a collapsed
	// top-level path defeats the runtime-prefix filter (".moai/" does not
	// carry the ".moai/state/" prefix) — hook state churn would count as
	// reviewable card work. File-level listing keeps the exclusion honest.
	untracked, err := reviewScopeGit(scope.Dir, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	for _, p := range strings.Split(untracked, "\n") {
		add(strings.TrimSpace(p))
	}
	return paths, nil
}

// hasReviewableCardChanges reports whether the card diff carries any
// reviewable path. Fail-open: a measurement failure ⇒ false (nothing
// reviewable ⇒ ALLOW), the same direction the tree detector takes outside a
// git repo.
func hasReviewableCardChanges(scope reviewScope) bool {
	if scope.Class != reviewScopeCard || scope.Dir == "" {
		return false
	}
	paths, err := cardChangedPaths(scope)
	if err != nil {
		return false
	}
	return len(paths) > 0
}

// cardScopeKeyParts computes the card-scope receipt binding (head, digest).
// The digest covers the merge base, the card diff content (a full diff
// against that base — the union measurement), and the non-runtime untracked
// files' paths and contents (the verify.Key freshness discipline: a re-edit
// of an untracked card file must move the binding). The runtime-prefix filter
// keeps hook/session state churn from invalidating a card receipt.
func cardScopeKeyParts(ctx context.Context, scope reviewScope) (string, string, error) {
	head, err := reviewScopeGit(scope.Dir, "rev-parse", "HEAD")
	if err != nil {
		return "", "", fmt.Errorf("card scope key: rev-parse HEAD: %s", execerr.StatusDetail(err))
	}
	diff, err := reviewScopeGit(scope.Dir, "diff", scope.MergeBase)
	if err != nil {
		return "", "", fmt.Errorf("card scope key: diff %s: %s", scope.MergeBase, execerr.StatusDetail(err))
	}
	untracked, err := reviewScopeGit(scope.Dir, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return "", "", fmt.Errorf("card scope key: untracked files: %s", execerr.StatusDetail(err))
	}
	h := sha256.New()
	// The merge base is a digest INPUT, not a pin: an absorption recreates it,
	// and the new base must not satisfy a receipt recorded under the old one
	// (REQ-CGS-002 / AC-CGS-012).
	h.Write([]byte(scope.MergeBase))
	h.Write([]byte{0})
	h.Write([]byte(diff))
	for _, name := range strings.Split(untracked, "\x00") {
		if name == "" || isRuntimeManagedPath(name) {
			continue // REQ-CGS-002: runtime-managed paths are outside the scope
		}
		root := filepath.Join(scope.Dir, filepath.FromSlash(name))
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			rel, err := filepath.Rel(scope.Dir, path)
			if err != nil {
				return err
			}
			h.Write([]byte(filepath.ToSlash(rel)))
			h.Write([]byte{0})
			switch {
			case entry.IsDir():
				h.Write([]byte{'d'})
			case entry.Type()&os.ModeSymlink != 0:
				target, err := os.Readlink(path)
				if err != nil {
					return err
				}
				h.Write([]byte{'l'})
				h.Write([]byte(target))
			case entry.Type().IsRegular():
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				h.Write([]byte{'f'})
				h.Write(content)
			default:
				return fmt.Errorf("unsupported file type at %q", path)
			}
			h.Write([]byte{0})
			return nil
		})
		if err != nil {
			return "", "", fmt.Errorf("card scope key: read untracked %q: %w", name, err)
		}
	}
	return head, hex.EncodeToString(h.Sum(nil))[:cardDigestHexLen], nil
}

// reviewGateEnvContext reads the session's factory env labels as OBSERVATION
// CONTEXT ONLY (REQ-CGS-010): the value rides the structured log row and is
// consumed by no decision (REQ-CGS-004 — the resolver's signature cannot read
// it). Referenced through the config constant, never a string literal
// (spec.md §G).
func reviewGateEnvContext() map[string]string {
	env := make(map[string]string)
	if v := os.Getenv(config.EnvMoaiFactoryWorker); v != "" {
		env[config.EnvMoaiFactoryWorker] = v
	}
	return env
}

// reviewGateScopeLogger is the scope-observation sink (REQ-CGS-010); tests
// swap it to capture the row.
var reviewGateScopeLogger = logReviewScope

// logReviewScope writes one structured JSON line to stderr — the gate's
// diagnostic channel; stdout stays the pure HookOutput contract. The class
// and its basis (branch match / none) are distinguishable per class, with env
// labels as context only.
func logReviewScope(scope reviewScope, env map[string]string) {
	row := map[string]any{"gate": "codex-review-gate", "scope": scope.Class, "basis": scope.Basis}
	if scope.Branch != "" {
		row["branch"] = scope.Branch
	}
	if len(env) > 0 {
		row["env"] = env
	}
	b, err := json.Marshal(row)
	if err != nil {
		return
	}
	fmt.Fprintln(os.Stderr, string(b))
}
