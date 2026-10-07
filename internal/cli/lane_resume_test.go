package cli

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/spf13/cobra"
)

// TestLaneJoinChildArgv (AC-SCV-008, REQ-SCV-008) — the pure assembler
// returns the child argv carrying the name pair, the settings pair, and the
// pass-through tokens in a deterministic order; it mutates no input, and two
// runs over the same inputs produce identical argv. Purity is by
// construction: the function takes no environment, filesystem, or process
// parameter and returns []string.
func TestLaneJoinChildArgv(t *testing.T) {
	launcherArgs := []string{"--name", "lane-3", "--", "--resume", "<session-id>"}
	settings := []string{"--settings", "<file>"}

	got := laneJoinChildArgv(launcherArgs, "lane-3", settings)
	want := []string{"--name", "lane-3", "--", "--resume", "<session-id>", "--settings", "<file>"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argv = %q, want %q", got, want)
	}
	again := laneJoinChildArgv(launcherArgs, "lane-3", settings)
	if !reflect.DeepEqual(got, again) {
		t.Fatalf("assembler is not deterministic: %q vs %q", got, again)
	}
	if !reflect.DeepEqual(launcherArgs, []string{"--name", "lane-3", "--", "--resume", "<session-id>"}) {
		t.Fatalf("the assembler mutated its input: %q", launcherArgs)
	}

	// Nameless launcher args get the injected pair before the pass-through
	// marker, so the desugared lane name can never land in the child's
	// pass-through region.
	got = laneJoinChildArgv([]string{"--", "--resume", "<session-id>"}, "lane-7", nil)
	want = []string{"--name", "lane-7", "--", "--resume", "<session-id>"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("nameless argv = %q, want %q", got, want)
	}
}

// TestResumeRequiresValue (AC-SCV-009, REQ-SCV-009) — a --resume token with
// no following value refuses before any launch, naming the required
// --resume <session-id> form, and the launch seam is never called. The
// refusing shapes run through runClaudeEntry with a counting launch stub; the
// accepting shapes assert the validation itself passes them.
func TestResumeRequiresValue(t *testing.T) {
	refusing := []struct {
		name string
		args []string
	}{
		{"space form at argv end", []string{"--", "--resume"}},
		{"space form before another flag", []string{"--", "--resume", "--settings", "x"}},
		{"equals form empty value", []string{"--", "--resume="}},
	}
	for _, tc := range refusing {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			launch := func(string, string, []string) error {
				calls++
				return nil
			}
			cmd := &cobra.Command{}
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			err := runClaudeEntry(cmd, tc.args, "cc", "claude", factory.BackendClaude, launch)
			if err == nil {
				t.Fatal("expected the valueless --resume to refuse")
			}
			if !strings.Contains(err.Error(), "--resume <session-id>") {
				t.Errorf("refusal must name --resume <session-id>; got: %v", err)
			}
			if calls != 0 {
				t.Errorf("launch seam called %d time(s), want 0", calls)
			}
		})
	}

	t.Run("well-formed values pass validation", func(t *testing.T) {
		if err := validateResumeArgs([]string{"--", "--resume", "<session-id>"}); err != nil {
			t.Errorf("space form refused: %v", err)
		}
		if err := validateResumeArgs([]string{"--", "--resume=<session-id>"}); err != nil {
			t.Errorf("equals form refused: %v", err)
		}
		if err := validateResumeArgs([]string{"--name", "lane-3", "--"}); err != nil {
			t.Errorf("a launch with no --resume token refused: %v", err)
		}
	})
}

// TestResumeShortAliasRequiresValue (card-review round 1, P1) — `-r` is
// Claude's short alias of --resume, so a valueless alias refuses exactly
// like the long form, in both spellings, before any launch; well-formed
// aliases pass.
func TestResumeShortAliasRequiresValue(t *testing.T) {
	refusing := []struct {
		name string
		args []string
	}{
		{"alias space form at argv end", []string{"--", "-r"}},
		{"alias equals form empty value", []string{"--", "-r="}},
	}
	for _, tc := range refusing {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			launch := func(string, string, []string) error {
				calls++
				return nil
			}
			cmd := &cobra.Command{}
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			err := runClaudeEntry(cmd, tc.args, "cc", "claude", factory.BackendClaude, launch)
			if err == nil {
				t.Fatal("expected the valueless -r alias to refuse")
			}
			if !strings.Contains(err.Error(), "--resume <session-id>") {
				t.Errorf("refusal must name --resume <session-id>; got: %v", err)
			}
			if calls != 0 {
				t.Errorf("launch seam called %d time(s), want 0", calls)
			}
		})
	}

	t.Run("well-formed aliases pass validation", func(t *testing.T) {
		if err := validateResumeArgs([]string{"--", "-r", "<session-id>"}); err != nil {
			t.Errorf("alias space form refused: %v", err)
		}
		if err := validateResumeArgs([]string{"--", "-r=<session-id>"}); err != nil {
			t.Errorf("alias equals form refused: %v", err)
		}
	})
}

// TestResumeAliasExactTokenOnly (card-review round 1, P1; refined round 2) —
// the VALIDATOR judges exact resume tokens only: `-rx`, `-root`,
// `--resumex` are other tokens and are never refused by it. (The GUARD is
// different by design since round 2: any `-r`-prefixed token counts as a
// resume carrier — see TestGuardRecognizesAttachedShortForm.)
func TestResumeAliasExactTokenOnly(t *testing.T) {
	for _, token := range []string{"-rx", "-root", "--resumex"} {
		if err := validateResumeArgs([]string{"--", token}); err != nil {
			t.Errorf("%q judged by the validation: %v", token, err)
		}
	}
}

// TestValidatorSkipsOptionValues (card-review round 2, P2) — a token that is
// the VALUE of another value-taking option (space form) is prompt data, not
// an option: the reviewer's repro `moai cc -- --append-system-prompt
// '--resume'` must reach launch as it did at base.
func TestValidatorSkipsOptionValues(t *testing.T) {
	for _, args := range [][]string{
		{"--append-system-prompt", "--resume"},
		{"--settings", "--resume"},
		{"--model", "--resume"},
		{"--", "--append-system-prompt", "--resume"},
	} {
		if err := validateResumeArgs(args); err != nil {
			t.Errorf("validateResumeArgs(%q) = %v, want nil (the token is another option's value)", args, err)
		}
	}
}

// TestValidatorStillRefusesValueless (card-review round 2, P2 pins) — the
// value-skip must consume exactly one token: a resume token that is nobody's
// value still refuses.
func TestValidatorStillRefusesValueless(t *testing.T) {
	for _, args := range [][]string{
		{"--resume"},
		{"--resume="},
		{"--append-system-prompt", "--resume", "--resume"}, // the second one is bare
		{"--append-system-prompt", "--", "--resume"},       // the `--` is a value; this resume is bare
	} {
		err := validateResumeArgs(args)
		if err == nil {
			t.Errorf("validateResumeArgs(%q) = nil, want the valueless-resume refusal", args)
			continue
		}
		if !strings.Contains(err.Error(), "--resume <session-id>") {
			t.Errorf("refusal must name --resume <session-id>; got: %v", err)
		}
	}
}

// TestAmbiguityNeverRefusesInValidator (card-review round 3, P1 — the
// validator's half of the principle) — ambiguity is never a refusal: behind
// an ambiguous option the next token may be its value, so the validator
// passes it silently. The ONLY refusal stays the definitive valueless
// resume.
func TestAmbiguityNeverRefusesInValidator(t *testing.T) {
	for _, args := range [][]string{
		{"-w", "--resume"},
		{"-w", "--resume", "<session-id>"},
		{"-d", "--resume"},
		{"--teleport", "-rabc"},
	} {
		if err := validateResumeArgs(args); err != nil {
			t.Errorf("validateResumeArgs(%q) = %v, want nil (ambiguity never refuses)", args, err)
		}
	}
}

// TestPostSeparatorLauncherFlagsAreInert (card-review round 4, P2) —
// interpretation splits at MoAI's separator: after it, `-p` is Claude's
// boolean --print, not the launcher's value-taking --profile — the
// launcher-side value flags are not claude's options in that region and must
// not consume a phantom value, so a valueless `--resume` behind them is
// refused.
func TestPostSeparatorLauncherFlagsAreInert(t *testing.T) {
	for _, args := range [][]string{
		{"--", "-p", "--resume"},
		{"--", "--profile", "--resume"},
		{"--", "--branch", "--resume"},
	} {
		err := validateResumeArgs(args)
		if err == nil {
			t.Errorf("validateResumeArgs(%q) = nil, want the valueless-resume refusal (the launcher flag is inert post-separator)", args)
			continue
		}
		if !strings.Contains(err.Error(), "--resume <session-id>") {
			t.Errorf("refusal must name --resume <session-id>; got: %v", err)
		}
	}
}

// TestPreSeparatorLauncherFlagsKeepValueBehavior (card-review round 4 pins) —
// before MoAI's separator the launcher-side value flags still consume their
// values (r1-r3 behavior unchanged).
func TestPreSeparatorLauncherFlagsKeepValueBehavior(t *testing.T) {
	for _, args := range [][]string{
		{"-p", "--resume", "<session-id>"},
		{"-m", "--resume"},
		{"--profile", "--resume"},
	} {
		if err := validateResumeArgs(args); err != nil {
			t.Errorf("validateResumeArgs(%q) = %v, want nil (pre-separator launcher value behavior)", args, err)
		}
	}
}

// TestSeparatorWinsOverAmbiguousValue (card-review round 3, P1 repro 3) —
// `moai cc -w -- -- --resume` reaches launch as it did at base: the
// ambiguous `-w` meets `--`, and the separator wins — a `--` is never
// consumed as an option's value, so both `--` tokens count and the trailing
// resume is post-separator prompt text.
func TestSeparatorWinsOverAmbiguousValue(t *testing.T) {
	if err := validateResumeArgs([]string{"-w", "--", "--", "--resume"}); err != nil {
		t.Fatalf("the separator must win over an ambiguous option's value-ness: %v", err)
	}

	t.Run("one-shot parity: the resume validation never refuses", func(t *testing.T) {
		calls := 0
		launch := func(string, string, []string) error {
			calls++
			return nil
		}
		cmd := &cobra.Command{}
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)

		// End to end the launcher's own -w handling decides what `-w --`
		// means (base-identical: that code is untouched by this SPEC); the
		// round-3 contract pins that the RESUME validation is not what
		// refuses, so the launch seam either runs or the error carries some
		// other surface's text — never the resume refusal.
		err := runClaudeEntry(cmd, []string{"-w", "--", "--", "--resume"}, "cc", "claude", factory.BackendClaude, launch)
		if err != nil && strings.Contains(err.Error(), "--resume <session-id>") {
			t.Fatalf("the resume validation refused behind the separator rule: %v", err)
		}
		if err == nil && calls != 1 {
			t.Errorf("launch seam called %d time(s), want 1 when the launch is reached", calls)
		}
	})
}

// TestResumeScanStopsAtClaudeSeparator (card-review round 1, P2) — the scan
// distinguishes the two `--` tokens: the FIRST is MoAI's pass-through
// separator (the emergency form's resume tokens live after it and stay
// judged); the SECOND is Claude's own argument separator, after which every
// token is prompt text and must never be judged. (c) pins the base parity:
// a well-formed token after Claude's separator reaches launch.
func TestResumeScanStopsAtClaudeSeparator(t *testing.T) {
	t.Run("resume after MoAI's separator stays validated", func(t *testing.T) {
		if err := validateResumeArgs([]string{"--", "--resume", "<session-id>"}); err != nil {
			t.Fatalf("the emergency form must stay validated: %v", err)
		}
	})
	t.Run("valueless token after Claude's separator is never judged", func(t *testing.T) {
		if err := validateResumeArgs([]string{"--", "--", "--resume"}); err != nil {
			t.Fatalf("prompt text after Claude's separator judged: %v", err)
		}
	})
	t.Run("double separator reaches launch as at base", func(t *testing.T) {
		calls := 0
		launch := func(string, string, []string) error {
			calls++
			return nil
		}
		cmd := &cobra.Command{}
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)

		err := runClaudeEntry(cmd, []string{"--", "--", "--resume", "x"}, "cc", "claude", factory.BackendClaude, launch)
		if err != nil {
			t.Fatalf("double-separator input must reach launch as at base: %v", err)
		}
		if calls != 1 {
			t.Errorf("launch seam called %d time(s), want 1 (base parity)", calls)
		}
	})
}
