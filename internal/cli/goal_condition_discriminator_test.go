package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/goal"
)

// The fixtures below are the separation experiment that established WHICH
// property of a condition string actually drives classification. They are
// pinned as a regression because the answer is counter-intuitive: neither the
// multi-line shape nor the ac_converge document frame contributes anything —
// the sole operative property is the presence of a modelConditionReferents
// token ("transcript" / "conversation", case-insensitive). Every string lacking
// one of those two ENGLISH words classifies mechanical and is handed to
// `sh -c`, which is why non-English prose (cases E/F) armed a goal that could
// never exit 0 and blocked every turn-end to the ceiling.
//
// SAFETY: none of these strings may be EXECUTED by a test.
// canonicalAcConvergeProse embeds `go test ./...` in backticks; routed to
// `sh -c` it would command-substitute and run the full suite.
const (
	// C: one line, no document frame, carries the referent.
	oneLineWithReferent = "the build log is surfaced in the conversation"

	// D: multi-line, ac_converge-shaped, no referent token.
	multilineNoReferent = "Every AC in acceptance.md has PASS evidence shown in the chat log;\n" +
		"AND go test exit 0 is shown.\nStop when all hold."

	// E: Korean prose — the reported defect shape (no English referent at all).
	koreanProse = "모든 차단 AC가 통과 증거와 함께 대화에 표시된다"

	// F: Korean prose, multi-line, ac_converge-shaped.
	koreanProseMultiline = ".moai/specs/SPEC-XXX/acceptance.md 의 모든 차단 AC 가 통과 증거를 갖는다;\n" +
		"그리고 go test ./... 종료코드 0 이 표시된다.\n모두 성립하면 중단한다."
)

// TestParseCondition_ReferentTokenIsTheSoleDiscriminator pins the separation
// result: shape and frame contribute NOTHING; only the referent token decides.
func TestParseCondition_ReferentTokenIsTheSoleDiscriminator(t *testing.T) {
	t.Parallel()

	tokenStripped := strings.ReplaceAll(canonicalAcConvergeProse, "conversation", "chat")

	cases := []struct {
		name string
		in   string
		want goal.ConditionType
	}{
		{"A canonical: multiline + frame + referent", canonicalAcConvergeProse, goal.ConditionModel},
		{"B canonical minus referent", tokenStripped, goal.ConditionMechanical},
		{"C one line, no frame, referent", oneLineWithReferent, goal.ConditionModel},
		{"D multiline + frame, no referent", multilineNoReferent, goal.ConditionMechanical},
		{"E korean prose", koreanProse, goal.ConditionMechanical},
		{"F korean prose, multiline + frame", koreanProseMultiline, goal.ConditionMechanical},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := parseCondition(tc.in)
			if got.Type != tc.want {
				t.Fatalf("parseCondition(%.40q) = %q, want %q", tc.in, got.Type, tc.want)
			}
		})
	}
}

// TestParseCondition_ExplicitPrefixWins covers the `model:` / `cmd:` declaration
// prefix: the EXPLICIT rule REQ-GLE-032 asks for. The prefix must beat the
// substring heuristic in BOTH directions, because the heuristic is an English
// allowlist and an allowlist fails silently on what it omits.
func TestParseCondition_ExplicitPrefixWins(t *testing.T) {
	t.Parallel()

	t.Run("model prefix forces model", func(t *testing.T) {
		t.Parallel()
		for _, in := range []string{
			"model:" + koreanProse,
			"model: " + koreanProse,
			"  MODEL:   " + koreanProse + "  ",
			"Model: go test ./...",
		} {
			cond := parseCondition(in)
			if cond.Type != goal.ConditionModel {
				t.Errorf("parseCondition(%q) = %q, want model", in, cond.Type)
			}
			if cond.Cmd != "" {
				t.Errorf("parseCondition(%q) carries cmd = %q, want empty", in, cond.Cmd)
			}
		}
	})

	t.Run("model prefix is stripped from the claim", func(t *testing.T) {
		t.Parallel()
		cond := parseCondition("model:  " + koreanProse)
		if cond.Claim != koreanProse {
			t.Errorf("claim = %q, want %q", cond.Claim, koreanProse)
		}
	})

	t.Run("cmd prefix forces mechanical over the referent heuristic", func(t *testing.T) {
		t.Parallel()
		cond := parseCondition("cmd: all AC rows show PASS in the transcript")
		if cond.Type != goal.ConditionMechanical {
			t.Fatalf("type = %q, want mechanical", cond.Type)
		}
		if cond.Cmd != "all AC rows show PASS in the transcript" {
			t.Errorf("cmd = %q, want the prefix-stripped text", cond.Cmd)
		}
	})

	t.Run("cmd prefix still parses the trailing exits clause", func(t *testing.T) {
		t.Parallel()
		cond := parseCondition("CMD: grep -q TODO main.go exits 1")
		if cond.Type != goal.ConditionMechanical {
			t.Fatalf("type = %q, want mechanical", cond.Type)
		}
		if cond.Cmd != "grep -q TODO main.go" {
			t.Errorf("cmd = %q, want %q", cond.Cmd, "grep -q TODO main.go")
		}
		if cond.ExpectExit != 1 {
			t.Errorf("expect_exit = %d, want 1", cond.ExpectExit)
		}
	})

	t.Run("no prefix falls back to the referent heuristic unchanged", func(t *testing.T) {
		t.Parallel()
		if got := parseCondition(oneLineWithReferent).Type; got != goal.ConditionModel {
			t.Errorf("unprefixed referent string = %q, want model", got)
		}
		if got := parseCondition("go test ./internal/cli/...").Type; got != goal.ConditionMechanical {
			t.Errorf("unprefixed command = %q, want mechanical", got)
		}
	})

	t.Run("a bare colon word is not a prefix", func(t *testing.T) {
		t.Parallel()
		// "modelling:" and "models:" must NOT be read as the model: prefix.
		for _, in := range []string{"modelling: something", "cmdline: something"} {
			cond := parseCondition(in)
			if cond.Cmd != in {
				t.Errorf("parseCondition(%q).Cmd = %q, want the string untouched", in, cond.Cmd)
			}
		}
	})
}
