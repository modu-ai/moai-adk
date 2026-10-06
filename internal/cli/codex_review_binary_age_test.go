package cli

import (
	"context"
	"testing"

	"github.com/modu-ai/moai-adk/internal/binlag"
)

// stubLag installs a lag-evaluator stub returning status and restores the real
// seam on cleanup; captured records whether the policy consulted the judge.
func stubLag(t *testing.T, status binlag.Status) *bool {
	t.Helper()
	called := false
	prev := reviewGateLagEvaluator
	reviewGateLagEvaluator = func(_ context.Context, _ binlag.Request) binlag.Verdict {
		called = true
		return binlag.Verdict{Status: status, BinaryCommit: "aaa111111", SourceHead: "bbb222222"}
	}
	t.Cleanup(func() { reviewGateLagEvaluator = prev })
	return &called
}

// stubIdentity pins the binary's ldflags coordinates.
func stubIdentity(t *testing.T, commit, ver string) {
	t.Helper()
	prev := reviewGateBinaryIdentity
	reviewGateBinaryIdentity = func() (string, string) { return commit, ver }
	t.Cleanup(func() { reviewGateBinaryIdentity = prev })
}

func TestStaleBinarySkip_BehindSkips(t *testing.T) {
	called := stubLag(t, binlag.StatusBehind)
	stubIdentity(t, "aaa111111", "v1")
	var rows []binlag.Verdict
	prevLog := binaryAgeSkipLogger
	binaryAgeSkipLogger = func(v binlag.Verdict) { rows = append(rows, v) }
	t.Cleanup(func() { binaryAgeSkipLogger = prevLog })

	if !staleBinarySkipApplies("/some/tree") {
		t.Fatalf("behind verdict must skip")
	}
	if !*called {
		t.Fatalf("policy did not consult the binary-lag judge")
	}
	if len(rows) != 1 {
		t.Fatalf("behind skip must log exactly one row, got %d", len(rows))
	}
}

func TestStaleBinarySkip_NonBehindVerdictsReview(t *testing.T) {
	for _, status := range []binlag.Status{
		binlag.StatusFresh,
		binlag.StatusDivergent,
		binlag.StatusAhead,
		binlag.StatusNotApplicable,
	} {
		t.Run(string(status), func(t *testing.T) {
			stubLag(t, status)
			stubIdentity(t, "aaa111111", "v1")
			if staleBinarySkipApplies("/some/tree") {
				t.Fatalf("%s verdict must not skip — only a decidable behind verdict is staleness", status)
			}
		})
	}
}

// The real judge's undecidable path: a development build carries no commit
// metadata, so the policy must review (fail-open in the review direction) —
// this exercises binlag.Evaluate itself, not a stub.
func TestStaleBinarySkip_DevBuildReviews(t *testing.T) {
	prev := reviewGateLagEvaluator
	reviewGateLagEvaluator = binlag.Evaluate
	t.Cleanup(func() { reviewGateLagEvaluator = prev })
	stubIdentity(t, "", "dev")

	if staleBinarySkipApplies("/some/tree") {
		t.Fatalf("a development build (no commit metadata) must not skip")
	}
}

// The gate-level contract: on a behind verdict the Claude Stop path returns
// ALLOW with the binary_age row logged — the row is the observable that the
// binary-age policy decided the outcome, not a later fail-open arm.
func TestHandleCodexReviewGate_StaleBinaryAllowsWithRow(t *testing.T) {
	stubLag(t, binlag.StatusBehind)
	stubIdentity(t, "aaa111111", "v1")
	var rows int
	prevLog := binaryAgeSkipLogger
	binaryAgeSkipLogger = func(binlag.Verdict) { rows++ }
	t.Cleanup(func() { binaryAgeSkipLogger = prevLog })

	out, err := HandleCodexReviewGate(gateInput(false), true /* enabled */, "/proj")
	if err != nil {
		t.Fatalf("gate error: %v", err)
	}
	if out == nil || out.Decision != "" {
		t.Fatalf("stale-binary path must ALLOW with an empty decision, got %+v", out)
	}
	if rows != 1 {
		t.Fatalf("stale-binary ALLOW must carry exactly one binary_age row, got %d", rows)
	}
}

// Contrast: on a fresh verdict the policy stays silent — the gate proceeds past
// the binary-age arm with no row.
func TestHandleCodexReviewGate_FreshBinaryLogsNoRow(t *testing.T) {
	stubLag(t, binlag.StatusFresh)
	stubIdentity(t, "aaa111111", "v1")
	var rows int
	prevLog := binaryAgeSkipLogger
	binaryAgeSkipLogger = func(binlag.Verdict) { rows++ }
	t.Cleanup(func() { binaryAgeSkipLogger = prevLog })

	if _, err := HandleCodexReviewGate(gateInput(false), true /* enabled */, "/proj"); err != nil {
		t.Fatalf("gate error: %v", err)
	}
	if rows != 0 {
		t.Fatalf("fresh binary must not log a binary_age row, got %d", rows)
	}
}
