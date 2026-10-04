package cli

// SPEC-CODEX-REVIEW-OWNERSHIP-001 REQ-CRO-001..006 + SPEC-CODEX-GATE-SCOPING-001
// REQ-CGSC-001..004 — the codex review gate's tree_scope policy and its
// primary-checkout sibling.
//
// A session whose scope class is TREE (codex_review_scope.go) and which carries
// no WT- branch evidence has no card to attribute its working tree to: the gate
// either reviews the whole uncommitted tree (workflow.codex.review_gate.
// tree_scope: review, the default and today's behavior) or lets the turn
// through (skip). Since SPEC-CODEX-GATE-SCOPING-001 a tree-scope session whose
// tree IS the repository's primary working tree skips by default regardless of
// tree_scope (primary_scope, REQ-CGSC-002), and only an explicit
// primary_scope: review restores the whole-tree review (REQ-CGSC-004). The
// decision takes exactly three inputs — the resolver's result (class, branch,
// primary determination) and the two key values. No environment, role or
// launcher mode is read here, and this file is kept free of every environment
// reference on purpose: TestTreeScopePolicy_SourceReadsNoEnvironment scans it
// for exactly that.
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

// primaryScopeReader is the primary-scope key-reader seam, the tree_scope
// reader's sibling (SPEC-CODEX-GATE-SCOPING-001 REQ-CGSC-002).
var primaryScopeReader = readCodexReviewGatePrimaryScope

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

// readCodexReviewGatePrimaryScope reads workflow.codex.review_gate.primary_scope
// from the same NESTED path, the sibling of readCodexReviewGateTreeScope with
// the REVERSED fail direction (SPEC-CODEX-GATE-SCOPING-001 §F.2): an empty
// root, a missing or unreadable file, a YAML error, a missing key (the flat
// form, a commented-out key and the key under another block included) and an
// unknown value all read as skip — the distributed default, because the gate
// does not review a primary checkout's unattributable changes (REQ-CGSC-002)
// — and only an explicit review restores the pre-SPEC behavior (REQ-CGSC-004).
// Value comparison is the single config.NormalizeCodexReviewGatePrimaryScope.
func readCodexReviewGatePrimaryScope(projectDir string) string {
	if projectDir == "" {
		return config.CodexReviewGatePrimaryScopeSkip
	}
	data, err := os.ReadFile(filepath.Join(projectDir, ".moai", "config", "sections", "workflow.yaml"))
	if err != nil {
		return config.CodexReviewGatePrimaryScopeSkip
	}
	var doc struct {
		Workflow struct {
			Codex struct {
				ReviewGate struct {
					PrimaryScope string `yaml:"primary_scope"`
				} `yaml:"review_gate"`
			} `yaml:"codex"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return config.CodexReviewGatePrimaryScopeSkip
	}
	return config.NormalizeCodexReviewGatePrimaryScope(doc.Workflow.Codex.ReviewGate.PrimaryScope)
}

// treeScopeSkipApplies reports whether the gate must let the turn through
// without a review. Two sibling policies decide at this one shared spot, so
// both automatic paths cannot disagree (REQ-CRO-006, REQ-CGSC-003): the
// primary-checkout skip — a tree-scope session whose tree IS the repository's
// primary working tree has no card to attribute its changes to, and the
// default primary_scope skips it regardless of tree_scope (REQ-CGSC-002) — and
// the tree_scope policy, which keeps governing non-primary tree sessions (and
// a primary session explicitly restored to review, REQ-CGSC-004). configRoot
// is lazy so a card session — the common case under a gate — costs no extra
// root resolution. On a skip it logs one row and the caller returns
// immediately, before the self-gate detector, the reviewer lookup and any
// review.
//
// @MX:NOTE: the one policy both automatic paths share (REQ-CRO-006, REQ-CGSC-003); a WT- branch is never skipped even when its merge base is unavailable (REQ-CRO-004)
func treeScopeSkipApplies(scope reviewScope, configRoot func() string) bool {
	if scope.Class != reviewScopeTree || cardScopeFromBranch(scope.Branch) {
		return false
	}
	root := configRoot()
	if scope.Primary && primaryScopeReader(root) != config.CodexReviewGatePrimaryScopeReview {
		primary := scope
		primary.Basis = "primary checkout (git dir == git common dir): " + scope.Dir
		treeScopeSkipLogger(primary, config.CodexReviewGatePrimaryScopeSkip)
		return true
	}
	if reviewGateTreeScopeReader(root) != config.CodexReviewGateTreeScopeSkip {
		return false
	}
	// The tree_scope arm decided, so the row's axis is tree_scope even on a
	// primary tree explicitly restored to review: the flag names the
	// primary-checkout policy, and this is not that policy's row.
	tree := scope
	tree.Primary = false
	treeScopeSkipLogger(tree, config.CodexReviewGateTreeScopeSkip)
	return true
}

// treeScopeSkipRow is the structured skip row: the gate, the scope class, the
// policy axis and value in effect, and the resolver's basis — distinguishable
// per reason (REQ-CGSC-011): a primary-checkout skip names primary_scope and a
// basis carrying the primary reason plus the resolved tree; a tree_scope skip
// names tree_scope. Distinct from the scope row (reviewGateScopeLogger),
// which every gate turn writes.
func treeScopeSkipRow(scope reviewScope, value string) map[string]any {
	axis := "tree_scope"
	if scope.Primary {
		axis = "primary_scope"
	}
	return map[string]any{"gate": "codex-review-gate", "scope": scope.Class, axis: value, "basis": scope.Basis}
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
