package cli

// TEMPORARY RED-stage stub (SPEC-CODEX-REVIEW-OWNERSHIP-001 M1). Declarations
// only, with deliberately wrong behaviour, so the new tests compile and fail at
// their intended assertions (EXPECTED_RED) instead of at the compiler
// (TOOL_FAILURE). Removed in M2 when codex_review_tree_scope.go lands.

var reviewGateTreeScopeReader = readCodexReviewGateTreeScope

var treeScopeSkipLogger = func(reviewScope, string) {}

func readCodexReviewGateTreeScope(string) string { return "review" }

func treeScopeSkipRow(reviewScope, string) map[string]any { return map[string]any{} }
