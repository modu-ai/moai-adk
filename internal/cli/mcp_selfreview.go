// Package cli — the on-demand self-review tools (`codex_review`, `glm_review`).
//
// A session that owns a change can ask codex or GLM to review it without being
// an auditor: the result is ADVISORY, it files no audit receipt, and the
// `workflow.audit.gates.*` required-gate conversion never applies to it. That is
// the difference from `codex_audit` / `glm_audit`, whose receipts the receipt
// guard treats as evidence for a PASS.
//
// Scope is the caller's choice and is REQUIRED — a default would let a lane that
// forgot it review only the uncommitted slice of its card and call that a card
// review:
//
//   - card — the card diff, resolved by the SAME scope resolver the turn-end
//     review gate uses (reviewScopeResolver), so there is no second card
//     discriminator and no second merge base computation. The base is
//     recomputed on every call, never pinned. A tree that is not a card worktree
//     returns `inconclusive` without reviewing anything.
//   - uncommitted — the named tree's uncommitted changes, with the request shape
//     the gate sends for a tree-scope session.
//
// codex runs through the gate's own pin-free driver (runCodexReviewRPC); GLM has
// no filesystem, so its material is the diff, collected here with the runtime-
// managed prefixes excluded by pathspec (the same list the gate's self-gate
// ignores). Neither tool narrates progress: a call is bounded by the review
// budget, shorter than the host's idle window.
//
// @MX:SPEC: SPEC-CODEX-REVIEW-OWNERSHIP-001
package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Self-review domain identifiers.
const (
	selfReviewScopeCard        = "card"
	selfReviewScopeUncommitted = "uncommitted"
	selfReviewBackendCodex     = "codex"
	selfReviewBackendGLM       = "glm"
	selfReviewScopeArg         = "scope"
	selfReviewModelArg         = "model"
)

// SelfReviewOutput is a review result plus the facts that make it attributable.
// Every return path of both tools passes through runSelfReview's single
// constructor, so no path can answer without naming its scope, base, backend and
// tree. The embedded ReviewOutput is flattened in the JSON, which keeps the
// verdict vocabulary identical to the audit tools'.
type SelfReviewOutput struct {
	ReviewOutput

	// Advisory is always true: the result is non-binding and can never stand in
	// for an audit verdict.
	Advisory bool `json:"advisory"`
	// Scope echoes the requested scope (`card` | `uncommitted`).
	Scope string `json:"scope"`
	// Base is the merge base SHA a card-scope review was measured from. Empty on
	// every other path (uncommitted, non-card early return, empty material).
	Base string `json:"base"`
	// Backend is `codex` or `glm`.
	Backend string `json:"backend"`
	// Tree is the canonical (symlink-resolved) root that was reviewed — not the
	// spelling the caller passed.
	Tree string `json:"tree"`
	// Truncated reports that the GLM material was cut at the size cap. codex reads
	// the tree itself, so it is always false there.
	Truncated bool `json:"truncated"`
	// ExcludedUntracked names the untracked non-runtime paths GLM could not be
	// shown (a diff carries no untracked files), for both scopes. Empty for codex.
	ExcludedUntracked []string `json:"excluded_untracked"`
}

// handleCodexReview is the `codex_review` handler.
func handleCodexReview(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return runSelfReview(ctx, req, selfReviewBackendCodex), nil
}

// handleGLMReview is the `glm_review` handler.
func handleGLMReview(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return runSelfReview(ctx, req, selfReviewBackendGLM), nil
}

// runSelfReview is the shared core of both tools. A malformed call (bad scope,
// unusable project_root) is a tool error the caller can correct; everything
// after that fails open to `inconclusive` (REQ-MCP-012).
//
// @MX:NOTE: [AUTO] the card/uncommitted split lives here once; the card branch must keep going through reviewScopeResolver (REQ-CRO-008) — a local WT- check or merge base computation would fork the gate's discriminator
// @MX:SPEC: SPEC-CODEX-REVIEW-OWNERSHIP-001
func runSelfReview(ctx context.Context, req mcp.CallToolRequest, backend string) *mcp.CallToolResult {
	tool := backend + "_review"
	scope := req.GetString(selfReviewScopeArg, "")
	if scope != selfReviewScopeCard && scope != selfReviewScopeUncommitted {
		return toolErr(tool, fmt.Errorf("scope is required and must be %q or %q, got %q", selfReviewScopeCard, selfReviewScopeUncommitted, scope))
	}
	root, err := resolveToolProjectRoot(req)
	if err != nil {
		return toolErr(tool, err)
	}

	out := SelfReviewOutput{Advisory: true, Scope: scope, Backend: backend, Tree: root, ExcludedUntracked: []string{}}
	finish := func(r ReviewOutput) *mcp.CallToolResult {
		out.ReviewOutput = r
		res, err := mcp.NewToolResultJSON(out)
		if err != nil {
			return mcp.NewToolResultText(tool + ": " + out.Verdict + " (" + out.Summary + ")")
		}
		return res
	}

	gitBase := "HEAD"
	rs := reviewScope{Class: reviewScopeTree, Dir: root}
	if scope == selfReviewScopeCard {
		rs = reviewScopeResolver(root)
		if rs.Class != reviewScopeCard {
			return finish(selfReviewInconclusive("scope card needs a card worktree (a WT- branch whose merge base can be computed); this tree is not one: " + rs.Basis))
		}
		out.Base, gitBase = rs.MergeBase, rs.MergeBase
	}

	diff, untracked, err := selfReviewMaterial(root, gitBase)
	if err != nil {
		return finish(selfReviewInconclusive("cannot read the change to review: " + err.Error()))
	}
	if backend == selfReviewBackendGLM {
		out.ExcludedUntracked = untracked
		out.Truncated = len(diff) > reviewDiffMaxBytes
	}
	// Empty material calls neither backend. GLM sees only the diff, so untracked
	// files alone leave it nothing to review (they are still reported above); codex
	// reads the working tree itself and is still asked about them.
	if strings.TrimSpace(diff) == "" && (backend == selfReviewBackendGLM || len(untracked) == 0) {
		out.Base = ""
		if len(untracked) > 0 {
			// Changes exist; a diff just cannot carry them. Do not read as a clean tree.
			return finish(selfReviewInconclusive("the only changes in the " + scope + " scope of this tree are untracked files, which a diff cannot carry; see excluded_untracked for the files this review could not see"))
		}
		return finish(selfReviewInconclusive("no change to review in the " + scope + " scope of this tree"))
	}

	model := strings.TrimSpace(req.GetString(selfReviewModelArg, ""))
	if backend == selfReviewBackendCodex {
		return finish(codexSelfReview(ctx, rs, model))
	}
	return finish(glmSelfReview(ctx, diff, model))
}

// selfReviewInconclusive is the fail-open result for a call that never reached a
// reviewer. It carries no "fall back to the auditor" step: the caller asked for
// a self-review, not an audit.
func selfReviewInconclusive(summary string) ReviewOutput {
	return ReviewOutput{Verdict: VerdictInconclusive, Summary: summary, Findings: []Finding{}, NextSteps: []string{}}
}

// selfReviewMaterial collects what a diff-based reviewer is shown and what it
// cannot be shown: `git diff <base> -- .` with the runtime-managed prefixes
// excluded by pathspec (tracked changes, staged or not, committed since base
// included), and the untracked paths outside those prefixes.
func selfReviewMaterial(root, base string) (diff string, untracked []string, err error) {
	// A plain unified diff whatever the user's git config says: no external diff
	// driver, no colour escapes, no textconv filter.
	args := []string{"diff", "--no-ext-diff", "--no-color", "--no-textconv", base, "--", "."}
	for _, p := range reviewGateRuntimePrefixes {
		args = append(args, ":(exclude)"+p)
	}
	untracked = []string{} // never nil: the result field is a JSON array on every path
	if diff, err = runReviewGit(root, args...); err != nil {
		return "", nil, fmt.Errorf("git diff %s: %w", base, err)
	}
	// -z: NUL-separated and never quoted, so non-ASCII names stay readable and
	// reach the runtime-prefix filter as real paths (core.quotepath is ignored).
	list, err := runReviewGit(root, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return "", nil, fmt.Errorf("git ls-files: %w", err)
	}
	for p := range strings.SplitSeq(list, "\x00") {
		if p != "" && !isRuntimeManagedPath(p) {
			untracked = append(untracked, p)
		}
	}
	return diff, untracked, nil
}

// codexSelfReview sends the gate's own review request for the resolved scope
// through the gate's pin-free driver: no audit model pin, no receipt, no
// required-gate conversion. A missing binary is `inconclusive`, never `fail`.
func codexSelfReview(ctx context.Context, rs reviewScope, model string) ReviewOutput {
	binaryPath, err := codexLookPath(codexBinaryName)
	if err != nil {
		return inconclusiveReview("codex binary not found in PATH")
	}
	params := reviewRequestParams(rs)
	if model != "" {
		params["model"] = model
	}
	ctx, cancel := context.WithTimeout(ctx, config.DefaultCodexReviewGateTimeout)
	defer cancel()
	out, _ := runCodexReviewRPC(ctx, binaryPath, codexMethodReviewStart, params) // fail-open inside
	return out
}

// glmSelfReview posts the diff to z.ai under the delegation default model (or the
// caller's), with no audit pin.
//
// callGLMAudit narrates progress to the MCP client whenever its context carries
// the MCP server. The self-review tools carry no heartbeat, so the call runs on
// a context that keeps the caller's cancellation but not its values.
//
// @MX:NOTE: [AUTO] context.AfterFunc spawns a goroutine that only forwards cancellation; the detached ctx is what keeps callGLMAudit's progress narration off this tool
// @MX:SPEC: SPEC-CODEX-REVIEW-OWNERSHIP-001
func glmSelfReview(ctx context.Context, diff, model string) ReviewOutput {
	key := glmKeyLoader()
	if key == "" {
		return glmInconclusive("GLM API key not configured (~/.moai/.env.glm)")
	}
	if model == "" {
		model = resolveGLMTaskModel()
	}
	detached, cancel := context.WithTimeout(context.Background(), config.DefaultCodexReviewGateTimeout)
	defer cancel()
	stop := context.AfterFunc(ctx, cancel)
	defer stop()
	return callGLMAudit(detached, key, model, "", "", truncateDiff(diff), nil)
}
