package cli

// The codex review gate's binary-age policy (card t1528) — the sibling of the
// tree_scope / primary_scope policies in codex_review_tree_scope.go.
//
// The gate binary is itself a judge: its handler embodies the review policies
// (tree_scope, primary_scope, the self-gate, the receipt rules) that decide
// what the review covers. A binary whose build commit is a strict ancestor of
// the session tree's HEAD (binlag.StatusBehind) is running policy OLDER than
// the code it would judge — the observed t1528 defect shape, where the leader's
// primary checkout carried a pre-SPEC-CODEX-GATE-SCOPING-001 bin/moai, so the
// primary-scope skip that the source prescribed never existed in the executing
// handler and every turn-end reviewed the whole tree.
//
// The skip is therefore fail-open in the REVIEW direction: only a decidable
// behind verdict skips; fresh, divergent, ahead and not-applicable all review
// exactly as before. An undecidable age (a development build with no commit
// metadata, a non-git session tree) is not evidence of staleness.
//
// Like treeScopeSkipApplies, this policy runs on BOTH automatic paths — the
// Claude Stop hook and the Codex Stop-chain member 6 — so they cannot disagree
// (REQ-CRO-006). The explicit producer `moai verify codex-review` never calls
// it.
//
// The verdict itself is NOT re-implemented here: REQ-ABI-006 keeps
// binlag.Evaluate the one binary-lag judge, and this policy only reads it.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modu-ai/moai-adk/internal/binlag"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// reviewGateLagEvaluator is the injectable seam over the one binary-lag judge
// (binlag.Evaluate); tests stub the verdict instead of shelling out to git.
var reviewGateLagEvaluator = binlag.Evaluate

// reviewGateBinaryIdentity supplies the judging build's commit and version
// coordinates (ldflags-injected); tests stub them.
var reviewGateBinaryIdentity = func() (commit, ver string) {
	return version.Commit, version.Version
}

// binaryAgeSkipLogger is the policy-observation sink; tests swap it to capture
// the row.
var binaryAgeSkipLogger = logBinaryAgeSkip

// staleBinarySkipApplies reports whether the gate must let the turn through
// because the executing binary predates the session tree it would judge.
// sessionDir is the already-resolved session tree (the scope's Dir); an empty
// dir is not decidable and reviews as before.
func staleBinarySkipApplies(sessionDir string) bool {
	commit, ver := reviewGateBinaryIdentity()
	verdict := reviewGateLagEvaluator(context.Background(), binlag.Request{
		Dir:           sessionDir,
		BinaryCommit:  commit,
		BinaryVersion: ver,
	})
	if verdict.Status != binlag.StatusBehind {
		return false
	}
	binaryAgeSkipLogger(verdict)
	return true
}

// logBinaryAgeSkip writes the skip row to stderr, the gate's diagnostic channel
// (stdout stays the pure HookOutput contract), distinguishable per axis
// (REQ-CGSC-011): the row names the binary_age policy and both coordinates the
// verdict compared.
func logBinaryAgeSkip(verdict binlag.Verdict) {
	row := map[string]any{
		"gate":          "codex-review-gate",
		"policy":        "binary_age",
		"verdict":       verdict.Status,
		"binary_commit": binlag.Short(verdict.BinaryCommit),
		"tree_head":     binlag.Short(verdict.SourceHead),
		"basis":         "gate binary predates the session tree — its review policies are older than the code under review",
	}
	b, err := json.Marshal(row)
	if err != nil {
		return
	}
	fmt.Fprintln(os.Stderr, string(b))
}
