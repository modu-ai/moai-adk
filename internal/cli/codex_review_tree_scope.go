package cli

// SPEC-CODEX-REVIEW-OWNERSHIP-001 REQ-CRO-001..006 — the codex review gate's
// tree_scope policy.
//
// A session whose scope class is TREE (codex_review_scope.go) and which carries
// no WT- branch evidence has no card to attribute its working tree to: the gate
// either reviews the whole uncommitted tree (workflow.codex.review_gate.
// tree_scope: review, the default and today's behavior) or lets the turn
// through (skip). The decision takes exactly two inputs — the resolver's result
// and the key value (REQ-CRO-005). No environment, role or launcher mode is
// read here, and this file is kept free of every environment reference on
// purpose: TestTreeScopePolicy_SourceReadsNoEnvironment scans it for exactly
// that.
//
// A WT- branch whose merge base cannot be computed also falls to the tree class
// (resolveReviewScope), but it carries card evidence, so it is never skipped
// (REQ-CRO-004): an unidentifiable card base is not a licence to stop reviewing.
//
// Both automatic gate paths — the Claude Stop hook and Codex Stop-chain
// member 6 — call treeScopeSkipApplies, so they cannot disagree on the policy
// for the same session state (REQ-CRO-006). The explicit producer
// `moai verify codex-review` never calls it.
//
// @MX:SPEC: SPEC-CODEX-REVIEW-OWNERSHIP-001

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/config"

	"gopkg.in/yaml.v3"
)

// reviewGateTreeScopeReader is the injectable key-reader seam (the
// reviewScopeResolver precedent): tests record which root each path reads.
var reviewGateTreeScopeReader = readCodexReviewGateTreeScope

// treeScopeSkipLogger is the policy-observation sink (REQ-CRO-002); tests swap
// it to capture the row.
var treeScopeSkipLogger = logTreeScopeSkip

// readCodexReviewGateTreeScope reads workflow.codex.review_gate.tree_scope from
// `.moai/config/sections/workflow.yaml` under projectDir and returns
// config.CodexReviewGateTreeScopeSkip or ...Review. It is the sibling of
// readCodexReviewGateEnabled and reads the same NESTED path: the flat form, a
// commented-out key, the key under another block, an unknown value, a missing
// or unreadable file and a YAML error all read as review — the direction that
// never silently reviews less (REQ-CRO-001). Value comparison is the single
// config.NormalizeCodexReviewGateTreeScope.
func readCodexReviewGateTreeScope(projectDir string) string {
	if projectDir == "" {
		return config.CodexReviewGateTreeScopeReview
	}
	data, err := os.ReadFile(filepath.Join(projectDir, ".moai", "config", "sections", "workflow.yaml"))
	if err != nil {
		return config.CodexReviewGateTreeScopeReview
	}
	var doc struct {
		Workflow struct {
			Codex struct {
				ReviewGate struct {
					TreeScope string `yaml:"tree_scope"`
				} `yaml:"review_gate"`
			} `yaml:"codex"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return config.CodexReviewGateTreeScopeReview
	}
	return config.NormalizeCodexReviewGateTreeScope(doc.Workflow.Codex.ReviewGate.TreeScope)
}

// treeScopeSkipApplies reports whether the gate must let the turn through
// without a review: the scope class is tree, the resolver found no WT- branch,
// and tree_scope reads skip in the root configRoot names. configRoot is lazy so
// a card session — the common case under a gate — costs no extra root
// resolution. On a skip it logs one row and the caller returns immediately,
// before the self-gate detector, the reviewer lookup and any review.
//
// @MX:NOTE: the one policy both automatic paths share (REQ-CRO-006); a WT- branch is never skipped even when its merge base is unavailable (REQ-CRO-004)
func treeScopeSkipApplies(scope reviewScope, configRoot func() string) bool {
	if scope.Class != reviewScopeTree || cardScopeFromBranch(scope.Branch) {
		return false
	}
	if reviewGateTreeScopeReader(configRoot()) != config.CodexReviewGateTreeScopeSkip {
		return false
	}
	treeScopeSkipLogger(scope, config.CodexReviewGateTreeScopeSkip)
	return true
}

// treeScopeSkipRow is the structured skip row: the gate, the scope class, the
// key value and the resolver's basis — distinct from the scope row
// (reviewGateScopeLogger), which every gate turn writes.
func treeScopeSkipRow(scope reviewScope, value string) map[string]any {
	return map[string]any{"gate": "codex-review-gate", "scope": scope.Class, "tree_scope": value, "basis": scope.Basis}
}

// logTreeScopeSkip writes the skip row to stderr, the gate's diagnostic channel
// (stdout stays the pure HookOutput contract).
func logTreeScopeSkip(scope reviewScope, value string) {
	b, err := json.Marshal(treeScopeSkipRow(scope, value))
	if err != nil {
		return
	}
	fmt.Fprintln(os.Stderr, string(b))
}
