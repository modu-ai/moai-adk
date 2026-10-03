// Package cli — `moai hook codex-review-gate` Stop-hook logic
// (SPEC-MOAI-MCP-SERVER-001 M2, REQ-MCP-008 / AC-MCP-009 / AC-MCP-010).
//
// codex_review_gate.go holds the PURE gate logic. The CLI subcommand wiring
// (stdin read → project-root resolve → config-gate read → this handler →
// stdout) lives in hook.go runCodexReviewGate. Keeping the logic pure + in the
// cli package lets it reuse the codex backend (mcp_codex.go, same package) and
// the readCodexReviewGateEnabled config reader without an import cycle
// (internal/cli → internal/hook already exists for the HookInput/HookOutput
// types). It returns the standard ALLOW/BLOCK HookOutput and NEVER invokes
// AskUserQuestion (subagent boundary — the orchestrator translates a BLOCK).
//
// @MX:SPEC: SPEC-MOAI-MCP-SERVER-001
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"

	"github.com/spf13/cobra"
)

// reviewGateRuntimePrefixes are the working-tree paths the change detector
// ignores: they are hook/session-written state that drifts on every turn, so
// counting them as "reviewable" would make the self-gate fire on every Stop
// and defeat AC-MCP-009 (no false block on a non-editing turn).
var reviewGateRuntimePrefixes = []string{
	".moai/state/",
	".moai/cache/",
	".moai/reports/",
	".moai/logs/",
	".moai/harness/",
	".claude/agent-memory/",
}

// reviewGateTreeConfigExactPaths and reviewGateTreeConfigDirPrefixes are the
// runtime-managed CONFIGURATION surfaces the TREE-scope self-gate additionally
// ignores (SPEC-CODEX-GATE-SCOPING-001 REQ-CGSC-007): the session's local
// Claude settings file and the MoAI managed config tree are known local state,
// not reviewable work. Deliberately SEPARATE from reviewGateRuntimePrefixes
// AND from the shared parser below (card-review repair R1): the exclusion is
// consulted only by the tree-scope self-gate (treeConfigOnlyFromPorcelain) —
// the card scope's path filter shares only the runtime list, a card's own
// commits under .moai/config/ keep counting as card work (REQ-CGSC-005 /
// AC-CGSC-009), and the multi-review gates keep the shared baseline detector.
// The settings file matches EXACTLY (card-review repair R4): a sibling name
// that merely extends it (.claude/settings.json.template) is a source-shaped
// path and stays reviewable; only the directory matches by prefix.
var (
	reviewGateTreeConfigExactPaths  = []string{".claude/settings.json"}
	reviewGateTreeConfigDirPrefixes = []string{".moai/config/"}
)

// reviewGateChangeDetector is the injectable "is there reviewable uncommitted
// work?" seam. The production default runs `git status --porcelain` and filters
// runtime-managed paths; tests swap it to drive the self-gate deterministically
// (true = reviewable change present, false = clean / no-edit turn).
var reviewGateChangeDetector = hasReviewableChanges

// HandleCodexReviewGate is the Stop-hook gate logic. ALLOW/BLOCK contract:
//   - empty HookOutput ({})                            = ALLOW (Claude may stop)
//   - {Decision: "block", Reason: "..."}               = BLOCK (keep working)
//
// Decision order (AC-MCP-009 self-gate + AC-MCP-010 opt-in + REQ-MCP-012
// fail-open; scope resolution per SPEC-CODEX-GATE-SCOPE-001):
//  1. gate disabled (config off)            → ALLOW (opt-in default-off, C6)
//  2. stop_hook_active (loop prevention)    → ALLOW (mandatory CC protocol)
//  3. scope resolution (REQ-CGS-001)        → card | tree, from the session tree
//     (REQ-CGS-005); class + basis logged (REQ-CGS-010), env context only
//     3a. tree class, no WT- branch, tree_scope skip → ALLOW before the self-gate
//     (SPEC-CODEX-REVIEW-OWNERSHIP-001 REQ-CRO-002; one policy row logged)
//  4. no reviewable change IN THE SCOPE     → ALLOW (self-gate; no false block)
//  5. codex missing                         → ALLOW (fail-open; can't trap the session)
//  6. codex review pass / inconclusive      → ALLOW
//  7. codex review FAIL, every finding on a runtime-managed config surface
//     → ALLOW + recorded reclassification (REQ-CGSC-008); otherwise BLOCK
//     (the gate's only block path)
//
// `enabled` is read by the caller (runCodexReviewGate via
// readCodexReviewGateEnabled) and passed in; it is re-checked here as
// defense-in-depth. The only config I/O in this function is step 3a's read of
// tree_scope, and only for a tree-class session: the root it reads is the one
// the caller read `enabled` from (reviewGateConfigRoot(projectDir)), through an
// injectable reader, so the logic stays testable without a real config tree.
func HandleCodexReviewGate(input *hook.HookInput, enabled bool, projectDir string) (*hook.HookOutput, error) {
	allow := &hook.HookOutput{}
	if !enabled {
		return allow, nil // (1) opt-in default-off
	}
	if input != nil && input.StopHookActive {
		return allow, nil // (2) loop prevention — never re-block an already-continuing turn
	}
	// (3) The scope is determined BEFORE any review or consult (REQ-CGS-001),
	// from the SESSION's working-directory tree — never from a spawn-frozen
	// CLAUDE_PROJECT_DIR naming a different tree (REQ-CGS-005). One resolver
	// serves both execution paths (REQ-CGS-009).
	scope := reviewScopeResolver(reviewScopeSessionDir(input, projectDir))
	reviewGateScopeLogger(scope, reviewGateEnvContext())
	// (3a) The tree_scope policy: a tree-class session with no WT- evidence has
	// no card to attribute its tree to. The read root is the one `enabled` came
	// from (reviewGateConfigRoot), resolved only when the class is tree.
	if treeScopeSkipApplies(scope, func() string { return reviewGateConfigRoot(projectDir) }) {
		return allow, nil
	}
	if !reviewGateScopedChangeDetector(scope) {
		return allow, nil // (4) scoped self-gate — nothing reviewable in the session's scope ⇒ no false block
	}

	binaryPath, err := codexLookPath(codexBinaryName)
	if err != nil {
		return allow, nil // (5) fail-open: a missing reviewer must not trap the session
	}

	// The 900s override (config.DefaultCodexReviewGateTimeout) is pinned in the
	// hook manifest (M4) for this hook only; here we enforce it on the codex
	// call itself so a hung review cannot stall the Stop beyond the manifest
	// budget. The moai-default 5s hook timeout does NOT apply (AC-MCP-010).
	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultCodexReviewGateTimeout)
	defer cancel()
	// The review request carries the scope: tree scope stays shape-identical
	// to its pre-SPEC form (REQ-CGS-003 / REQ-CRT-006), card scope names the
	// card diff (REQ-CGS-002). projectDir is the CONFIG root only (it feeds
	// reviewGateConfigRoot for the tree_scope read in step 3a) — never the
	// review target when the session tree differs (REQ-CGS-005).
	out, rpcErr := runCodexReviewRPC(ctx, binaryPath, codexMethodReviewStart, reviewRequestParams(scope))
	if rpcErr != nil {
		// (6) fail-open: an inconclusive or erroring reviewer ⇒ ALLOW. The error
		// rides back with the ALLOW so runCodexReviewGate can log WHY on stderr;
		// it does not change the decision. Swallowing it here made a gate that
		// was turned on but structurally unable to reach a verdict look exactly
		// like a gate that had reviewed the change and found nothing wrong.
		return allow, rpcErr
	}
	if isBlockVerdict(out.Verdict) {
		// (7-pre) REQ-CGSC-008: a TREE-scope review whose every finding targets
		// only the runtime-managed configuration surfaces is known local drift
		// (settings/config churn), not a review defect this session owns — the
		// turn is allowed and the reclassification recorded (REQ-CGSC-011).
		// Mixed findings keep the gate's only block path below.
		if scope.Class == reviewScopeTree {
			if targets, ok := runtimeConfigOnlyFindings(out.Findings, scope.Dir); ok {
				logRuntimeDriftReclassification(scope, targets)
				return allow, nil
			}
		}
		return &hook.HookOutput{
			Decision: hook.DecisionBlock, // (7) the gate's only BLOCK path
			Reason:   "codex review gate: " + out.Summary,
		}, nil
	}
	return allow, nil // pass / inconclusive ⇒ ALLOW
}

// isBlockVerdict reports whether a codex verdict should BLOCK the session. A
// BLOCK fires ONLY on a clear failure; the inconclusive fail-open value and any
// non-fail verdict ALLOW. Matched case-insensitively on a "fail" prefix so
// "fail" / "FAIL" / "failed" all block while "inconclusive" / "pass" do not.
func isBlockVerdict(verdict string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(verdict)), "fail")
}

// hasReviewableChanges reports whether the working tree at projectDir carries
// any uncommitted change to a NON-runtime-managed path. It shells out to
// `git status --porcelain` (the SAME signal internal/hook/security turn-review
// uses). Fail-open: a non-git dir, missing git, or git error ⇒ false (nothing
// reviewable ⇒ ALLOW), so the gate never false-blocks outside a git repo.
func hasReviewableChanges(projectDir string) bool {
	if projectDir == "" {
		return false
	}
	out, err := exec.Command("git", "-C", projectDir, "status", "--porcelain").Output()
	if err != nil {
		return false
	}
	return reviewableFromPorcelain(string(out))
}

// reviewableFromPorcelain is the pure parser over a `git status --porcelain`
// payload: it returns true iff at least one changed path is NOT under a
// runtime-managed prefix. Split out so the filtering rule is unit-testable
// without a git repo. Porcelain line shape is fixed-width: columns 0-1 are the
// XY status (either may be a space), column 2 is a space, and the path begins
// at column 3 — so the path MUST be sliced from the raw line, NOT from a
// TrimSpace'd copy (TrimSpace would strip a leading-space status and shift the
// path off by one, dropping its leading "." and defeating the prefix filter).
// Renames use "XY <old> -> <new>"; the prefix check against <old> is sufficient.
//
// This is the SHARED baseline parser (card-review repair R1): it consults only
// the runtime-managed state prefixes, never the tree-only config surfaces —
// the multi-review gate (HandleMultiReviewGate) and Codex Stop-chain member 7
// consume it through the reviewGateChangeDetector seam, and a config-only
// exclusion here let them silently allow over a stored required FAIL. The
// tree-scope counterpart is treeConfigOnlyFromPorcelain below.
func reviewableFromPorcelain(porcelain string) bool {
	for _, raw := range strings.Split(porcelain, "\n") {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		if len(raw) <= 3 {
			continue // malformed — no path column
		}
		path := strings.TrimSpace(raw[3:]) // path column; trim trailing space / CR only
		if idx := strings.Index(path, " -> "); idx >= 0 {
			path = path[:idx] // rename source for the prefix check
		}
		if path == "" || isRuntimeManagedPath(path) {
			continue
		}
		return true
	}
	return false
}

// treeExcludedPath reports whether one porcelain path is excluded on the TREE
// path: the shared runtime-managed state prefixes plus the tree-only config
// surfaces.
func treeExcludedPath(path string) bool {
	return isRuntimeManagedPath(path) || isTreeRuntimeConfigPath(path)
}

// treeConfigOnlyFromPorcelain is the pure TREE-scope counterpart of
// reviewableFromPorcelain (card-review repair R1): it reports whether a
// `git status --porcelain` payload carries at least one change and EVERY one
// of them is excluded on the tree path — the config-only turn the tree
// self-gate must not review (REQ-CGSC-007). An empty payload and any
// non-excluded record read false, so the probe can only narrow the shared
// detector's answer, never widen it. A rename record is config-only only when
// BOTH sides are excluded (card-review repair R5):
// `.claude/settings.json -> main.go` is a real source change — the
// destination must not inherit the source's exclusion.
func treeConfigOnlyFromPorcelain(porcelain string) bool {
	any := false
	for _, raw := range strings.Split(porcelain, "\n") {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		if len(raw) <= 3 {
			continue // malformed — no path column
		}
		path := strings.TrimSpace(raw[3:]) // path column; trim trailing space / CR only
		excluded := false
		if idx := strings.Index(path, " -> "); idx >= 0 {
			excluded = treeExcludedPath(strings.TrimSpace(path[:idx])) &&
				treeExcludedPath(strings.TrimSpace(path[idx+4:]))
		} else {
			excluded = treeExcludedPath(path)
		}
		if !excluded {
			return false
		}
		any = true
	}
	return any
}

// treeConfigOnlyChanges reports whether the working tree at dir carries ONLY
// tree-excluded changes (treeConfigOnlyFromPorcelain over `git status
// --porcelain`). Fail-open in the PRESERVE direction: a measurement failure
// reads false — the exclusion never widens on an unreadable tree, so the
// scoped self-gate keeps the shared detector's answer there.
func treeConfigOnlyChanges(dir string) bool {
	if dir == "" {
		return false
	}
	out, err := exec.Command("git", "-C", dir, "status", "--porcelain").Output()
	if err != nil {
		return false
	}
	return treeConfigOnlyFromPorcelain(string(out))
}

// isRuntimeManagedPath reports whether path falls under a hook/session-written
// prefix the gate must ignore (otherwise every turn's .moai/state drift would
// trip the self-gate).
func isRuntimeManagedPath(path string) bool {
	for _, p := range reviewGateRuntimePrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// isTreeRuntimeConfigPath reports whether path falls under a runtime-managed
// configuration surface the TREE path ignores (REQ-CGSC-007). The card path's
// filter never consults this: a card commit under .moai/config/ stays card
// work (REQ-CGSC-005). The settings FILE matches exactly (card-review repair
// R4); the managed config DIRECTORY matches by prefix.
func isTreeRuntimeConfigPath(path string) bool {
	for _, p := range reviewGateTreeConfigExactPaths {
		if path == p {
			return true
		}
	}
	for _, p := range reviewGateTreeConfigDirPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// normalizeFindingPath brings a review finding's file anchor into the
// repo-relative slash form the exclusion sets are written in (card-review
// repair R3): a "./"-prefixed relative is stripped, and an absolute path is
// relativized against the reviewed scope's tree. A path OUTSIDE the scope
// relativizes to a "../" form that matches no prefix — the fail-closed block.
func normalizeFindingPath(file, dir string) string {
	p := strings.TrimSpace(file)
	if p == "" {
		return ""
	}
	p = strings.TrimPrefix(p, "./")
	if filepath.IsAbs(p) && dir != "" {
		if rel, err := filepath.Rel(dir, filepath.FromSlash(p)); err == nil {
			p = filepath.ToSlash(rel)
		}
	}
	return p
}

// runtimeConfigOnlyFindings reports whether EVERY finding of a review targets
// only the runtime-managed configuration surfaces, and returns the distinct
// targets when so (REQ-CGSC-008). A review with no findings at all is NOT
// config-only: a fail verdict behind an unparseable findings list is the
// contradiction state, never a licence to allow. Findings without a file
// anchor, and anchors outside the surfaces, keep the review's block. The
// anchor is normalized against the reviewed scope's tree dir before the
// comparison (card-review repair R3), so an absolute anchor reclassifies
// exactly as its relative twin.
func runtimeConfigOnlyFindings(findings []Finding, dir string) ([]string, bool) {
	if len(findings) == 0 {
		return nil, false
	}
	var targets []string
	seen := make(map[string]bool)
	for _, f := range findings {
		file := normalizeFindingPath(f.File, dir)
		if file == "" || !isTreeRuntimeConfigPath(file) {
			return nil, false
		}
		if !seen[file] {
			seen[file] = true
			targets = append(targets, file)
		}
	}
	return targets, true
}

// logRuntimeDriftReclassification writes the reclassification row to stderr,
// the gate's diagnostic channel (REQ-CGSC-011): the reason in the SPEC's own
// words and the targeted paths, distinguishable from a skip row. stdout stays
// the pure HookOutput contract.
func logRuntimeDriftReclassification(scope reviewScope, targets []string) {
	row := map[string]any{
		"gate":         "codex-review-gate",
		"scope":        scope.Class,
		"reclassified": "runtime-managed drift",
		"targets":      targets,
	}
	b, err := json.Marshal(row)
	if err != nil {
		return
	}
	fmt.Fprintln(os.Stderr, string(b))
}

// runCodexReviewGate is the cobra RunE for `moai hook codex-review-gate`
// (SPEC-MOAI-MCP-SERVER-001 M2 REQ-MCP-008). It reads the Stop-hook stdin JSON,
// resolves the project root, reads the workflow.codex.review_gate.enabled
// opt-in flag, and forwards to HandleCodexReviewGate. The output is emitted as
// JSON on stdout (the Stop-hook contract). Fail-open: any error logs to stderr
// and exits 0 so the Stop pipeline is never broken (REQ-MCP-012).
func runCodexReviewGate(cmd *cobra.Command, _ []string) error {
	input, err := readHookInput(cmd.InOrStdin())
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "codex-review-gate: invalid stdin JSON (%v); emitting default output\n", err)
		// Fail-open: emit an empty ALLOW.
		return emitHookOutput(cmd.OutOrStdout(), &hook.HookOutput{})
	}
	projectDir := resolveProjectDirFromInput(input)
	enabled := readCodexReviewGateEnabled(reviewGateConfigRoot(projectDir))
	out, gateErr := HandleCodexReviewGate(input, enabled, projectDir)
	if gateErr != nil {
		// Fail-open: a handler error MUST NOT trap the Stop pipeline.
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "codex-review-gate: error:", gateErr)
		return emitHookOutput(cmd.OutOrStdout(), &hook.HookOutput{})
	}
	if out == nil {
		out = &hook.HookOutput{}
	}
	return emitHookOutput(cmd.OutOrStdout(), out)
}

// reviewGateConfigRoot returns the root whose workflow config carries the Stop
// review gates' opt-in flags for a resolved projectDir: the primary checkout
// for a config-orphaned linked worktree, projectDir itself otherwise, and ""
// (gate disabled, the fail-open direction of these gates) when that primary
// cannot be identified (SPEC-WORKTREE-STATE-ROOT-001 REQ-WSR-008). The order
// in which projectDir is picked, and the tree the codex review targets, do
// not change.
func reviewGateConfigRoot(projectDir string) string {
	root, err := auditreceipt.StoreRoot(projectDir)
	if err != nil {
		return ""
	}
	return root
}

// readHookInput reads and parses the hook stdin JSON into a HookInput. It is
// broken out so the fail-open path can distinguish "could not read stdin" from
// "handler errored".
func readHookInput(r io.Reader) (*hook.HookInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}
	var input hook.HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("parse stdin: %w", err)
	}
	return &input, nil
}

// emitHookOutput serializes out as JSON to w. A nil-safe helper used by the
// fail-open path and the success path alike.
func emitHookOutput(w io.Writer, out *hook.HookOutput) error {
	enc := json.NewEncoder(w)
	return enc.Encode(out)
}

// resolveProjectDirFromInput picks the best available project dir from the hook
// input: ProjectDir first, then CWD, then the CLAUDE_PROJECT_DIR environment
// variable, then empty (the gate self-gates on empty).
//
// The CWD arm is load-bearing: Claude Code sends `cwd` in the Stop payload and
// treats `project_dir` as a legacy field (internal/hook/types.go), so a
// ProjectDir-only resolution returned "" on every real invocation and both gate
// readers fail-CLOSED to false no matter how the project was configured. The
// env arm mirrors how the shell wrappers resolve the project root, so the two
// layers agree on which workflow.yaml they are reading.
func resolveProjectDirFromInput(input *hook.HookInput) string {
	if input != nil {
		if input.ProjectDir != "" {
			return input.ProjectDir
		}
		if input.CWD != "" {
			return input.CWD
		}
	}
	return os.Getenv("CLAUDE_PROJECT_DIR")
}
