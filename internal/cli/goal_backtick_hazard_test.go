package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/goal"
)

// The backtick hazard: a misclassified prose condition is not merely a condition
// that can never hold — it is a string handed to `sh -c`, so any command
// substitution inside it RUNS. The canonical ac_converge paragraph embeds
// `go test ./...` in backticks, which makes the misclassification a full local
// test-suite execution rather than a failed predicate.
//
// These tests pin how far the repair reaches. They exercise the ARM path only,
// which never executes the condition; every backtick body below is `true` so
// that even an unexpected execution is harmless.

// TestBacktickProse_ClassifiedModelNeverReachesTheShell is the first line of
// defence: the canonical paragraph carries "conversation", so it classifies
// model and its cmd field stays empty. Nothing is ever handed to the shell.
func TestBacktickProse_ClassifiedModelNeverReachesTheShell(t *testing.T) {
	t.Parallel()

	cond := parseCondition(canonicalAcConvergeProse)
	if cond.Type != goal.ConditionModel {
		t.Fatalf("canonical prose classified %q, want model", cond.Type)
	}
	if cond.Cmd != "" {
		t.Fatalf("model condition carries cmd %q — the backticked `go test ./...` "+
			"would be command-substituted by sh -c", cond.Cmd)
	}
}

// TestBacktickProse_WithoutReferentIsRefusedAtArm is the second line: strip the
// referent token and the same paragraph classifies mechanical, reaching the arm
// gate. Its first word ("Every") resolves to no command, so the arm is refused
// and the backticks are never substituted.
func TestBacktickProse_WithoutReferentIsRefusedAtArm(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	prose := strings.ReplaceAll(canonicalAcConvergeProse, "conversation", "chat")
	if got := parseCondition(prose).Type; got != goal.ConditionMechanical {
		t.Fatalf("referent-stripped prose classified %q, want mechanical "+
			"(the fixture no longer reproduces the hazard)", got)
	}
	if !strings.Contains(prose, "`go test ./...`") {
		t.Fatalf("fixture lost its backticks — this test no longer covers the hazard")
	}

	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal", "arm", prose, "--session", "BACKTICK"})
	if err := rc.Execute(); err == nil {
		t.Fatalf("backticked prose was armed (out=%s)", buf.String())
	}
}

// TestBacktickHazard_KnownGaps pins what the arm gate does NOT catch, so the
// limit is measured rather than assumed. Both shapes below are armed today: the
// gate judges only the FIRST word, and a backtick anywhere after it is invisible
// to that judgement. The eval-time exit-127 backstop does not help either —
// these commands resolve and run.
//
// This is a deliberate consequence of the gate refusing only on positive
// evidence: a first-token heuristic that tried to reason about the rest of the
// string would refuse legitimate commands, which is the worse error. The
// remaining exposure is recorded here and in the card verdict.
func TestBacktickHazard_KnownGaps(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
	}{
		// A resolvable first word: the gate passes it, backticks and all.
		{"resolvable first word", "true `true`"},
		// A leading backtick: the gate skips the shape it cannot parse.
		{"leading backtick", "`true` && true"},
	} {
		cond := parseCondition(tc.in)
		if cond.Type != goal.ConditionMechanical {
			t.Fatalf("%s: classified %q, want mechanical", tc.name, cond.Type)
		}
		if tok, bad := unrunnableCommandToken(t.Context(), cond.Cmd); bad {
			t.Errorf("%s: gate refused on %q — the known gap has closed; "+
				"update the card verdict rather than this expectation", tc.name, tok)
		}
	}
}
