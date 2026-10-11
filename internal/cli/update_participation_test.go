package cli

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// countingPromptReader counts how many times the prompt source was opened —
// one open per promptBool call — and feeds the scripted answers.
type countingPromptReader struct {
	answers []string
	pos     int
	opens   int
}

func (c *countingPromptReader) Read(p []byte) (int, error) {
	if c.pos >= len(c.answers) {
		return 0, errors.New("scripted answers exhausted")
	}
	ans := c.answers[c.pos]
	c.pos++
	c.opens++
	if n := copy(p, ans); n < len(p) {
		return n, nil
	}
	return len(p), nil
}

// updateParticipationFixture swaps the seams the step depends on and points
// MOAI_HOME at a temporary directory. Restore is registered with t.Cleanup.
func updateParticipationFixture(t *testing.T, answers ...string) (promptOpens *int, consentPath string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	t.Setenv("CI", "")
	consentPath = filepath.Join(home, "config", "participation.yaml")

	reader := &countingPromptReader{answers: answers}
	opens := 0
	prevSource := stdinPromptSource
	prevInteractive := isInteractiveStdin
	prevLocale := participationLocaleFn
	stdinPromptSource = func() io.Reader {
		opens++
		return reader
	}
	isInteractiveStdin = func() bool { return true }
	participationLocaleFn = func() string { return "en" }
	t.Cleanup(func() {
		stdinPromptSource = prevSource
		isInteractiveStdin = prevInteractive
		participationLocaleFn = prevLocale
	})
	return &opens, consentPath
}

func newBareUpdateCmd() *cobra.Command {
	// The participation gates read these flags from a PRIVATE set (review
	// finding, r7): AddFlagSet shares the pflag.Flag POINTERS, so a value
	// another test left on the package-level updateCmd — a Set survives the
	// test that made it — leaked into a "bare" command and tripped the
	// mode-flag gate for every subtest (the CI-only divergence: prompt
	// opens = 0 in the full suite, green under -run). Each flag is
	// re-registered here with the ORIGINAL's type and DEFAULT value, so the
	// gates read fresh defaults whatever the shared set holds.
	names := append(append([]string{}, updateParticipationModeFlags...), updateParticipationModifierFlags...)
	fs := pflag.NewFlagSet("update", pflag.ContinueOnError)
	for _, name := range names {
		orig := updateCmd.Flags().Lookup(name)
		if orig == nil {
			continue
		}
		if orig.Value.Type() == "bool" {
			def, _ := strconv.ParseBool(orig.DefValue)
			_ = fs.Bool(name, def, "")
		} else {
			_ = fs.String(name, orig.DefValue, "")
		}
	}
	c := &cobra.Command{Use: "update"}
	c.Flags().AddFlagSet(fs)
	return c
}

func TestUpdateParticipationStep(t *testing.T) {
	t.Run("first_run_prompts_once_and_persists_both_values", func(t *testing.T) {
		opens, path := updateParticipationFixture(t, "n\n")
		err := runParticipationStep(newBareUpdateCmd(), os.Stdout)
		if err != nil {
			t.Fatalf("runParticipationStep: %v", err)
		}
		if *opens != 1 {
			t.Fatalf("prompt opens = %d, want 1", *opens)
		}
		body, _ := os.ReadFile(path)
		for _, want := range []string{"enabled: false", "asked: true"} {
			if !strings.Contains(string(body), want) {
				t.Errorf("consent file missing %q:\n%s", want, body)
			}
		}
	})

	t.Run("empty_answer_persists_enabled_false_asked_true", func(t *testing.T) {
		opens, path := updateParticipationFixture(t, "\n")
		if err := runParticipationStep(newBareUpdateCmd(), os.Stdout); err != nil {
			t.Fatalf("runParticipationStep: %v", err)
		}
		if *opens != 1 {
			t.Fatalf("prompt opens = %d, want 1", *opens)
		}
		body, _ := os.ReadFile(path)
		if !strings.Contains(string(body), "enabled: false") || !strings.Contains(string(body), "asked: true") {
			t.Errorf("empty answer did not persist enabled false + asked true:\n%s", body)
		}
	})

	t.Run("second_run_does_not_prompt", func(t *testing.T) {
		opens, _ := updateParticipationFixture(t, "n\n")
		cmd := newBareUpdateCmd()
		if err := runParticipationStep(cmd, os.Stdout); err != nil {
			t.Fatalf("first run: %v", err)
		}
		first := *opens
		if err := runParticipationStep(cmd, os.Stdout); err != nil {
			t.Fatalf("second run: %v", err)
		}
		if *opens != first {
			t.Fatalf("second run prompted %d more time(s)", *opens-first)
		}
	})

	t.Run("stdin_not_a_terminal_never_prompts", func(t *testing.T) {
		opens, _ := updateParticipationFixture(t)
		prev := isInteractiveStdin
		isInteractiveStdin = func() bool { return false }
		t.Cleanup(func() { isInteractiveStdin = prev })
		if err := runParticipationStep(newBareUpdateCmd(), os.Stdout); err != nil {
			t.Fatalf("runParticipationStep: %v", err)
		}
		if *opens != 0 {
			t.Fatalf("prompt opens = %d, want 0", *opens)
		}
	})

	t.Run("ci_environment_never_prompts", func(t *testing.T) {
		opens, _ := updateParticipationFixture(t)
		t.Setenv("CI", "1")
		if err := runParticipationStep(newBareUpdateCmd(), os.Stdout); err != nil {
			t.Fatalf("runParticipationStep: %v", err)
		}
		if *opens != 0 {
			t.Fatalf("prompt opens = %d, want 0", *opens)
		}
	})

	for _, modeFlag := range updateParticipationModeFlags {
		modeFlag := modeFlag
		t.Run("mode_flag_"+modeFlag+"_never_prompts", func(t *testing.T) {
			opens, path := updateParticipationFixture(t)
			// The flag is set on THIS subtest's own bare command: the bare
			// command's flag set is private (r7), so setting the shared
			// updateCmd's flag would not reach the gates — and would
			// re-create the leak the private set exists to prevent.
			cmd := newBareUpdateCmd()
			f := cmd.Flags().Lookup(modeFlag)
			if f == nil {
				t.Fatalf("bare command has no flag %q", modeFlag)
			}
			if err := f.Value.Set(boolOrText(modeFlag, "on")); err != nil {
				t.Fatalf("set --%s: %v", modeFlag, err)
			}
			if err := runParticipationStep(cmd, os.Stdout); err != nil {
				t.Fatalf("runParticipationStep: %v", err)
			}
			if *opens != 0 {
				t.Fatalf("--%s prompted %d time(s)", modeFlag, *opens)
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("--%s wrote the consent file: %v", modeFlag, err)
			}
		})
	}

	for _, modFlag := range updateParticipationModifierFlags {
		modFlag := modFlag
		t.Run("modifier_flag_"+modFlag+"_still_prompts", func(t *testing.T) {
			opens, _ := updateParticipationFixture(t, "n\n")
			// Same private-set rule as the mode flags above: the modifier
			// is set on this subtest's own bare command.
			cmd := newBareUpdateCmd()
			f := cmd.Flags().Lookup(modFlag)
			if f == nil {
				t.Fatalf("bare command has no flag %q", modFlag)
			}
			if err := f.Value.Set(boolOrText(modFlag, "on")); err != nil {
				t.Fatalf("set --%s: %v", modFlag, err)
			}
			if err := runParticipationStep(cmd, os.Stdout); err != nil {
				t.Fatalf("runParticipationStep: %v", err)
			}
			if *opens != 1 {
				t.Fatalf("--%s prompted %d time(s), want 1", modFlag, *opens)
			}
		})
	}

	t.Run("yes_never_sets_enabled_true", func(t *testing.T) {
		opens, path := updateParticipationFixture(t)
		cmd := newBareUpdateCmd()
		f := cmd.Flags().Lookup("yes")
		if f == nil {
			t.Fatal("bare command has no flag \"yes\"")
		}
		if err := f.Value.Set("true"); err != nil {
			t.Fatalf("set --yes: %v", err)
		}
		if err := runParticipationStep(cmd, os.Stdout); err != nil {
			t.Fatalf("runParticipationStep: %v", err)
		}
		if *opens != 0 {
			t.Fatalf("--yes prompted %d time(s), want 0", *opens)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("--yes wrote the consent file: %v", err)
		}
	})

	t.Run("prompt_text_equals_init_question_text_per_locale", func(t *testing.T) {
		_, _ = updateParticipationFixture(t, "n\n")
		var out strings.Builder
		if err := runParticipationStep(newBareUpdateCmd(), &out); err != nil {
			t.Fatalf("runParticipationStep: %v", err)
		}
		// The en prompt carries the init question's title and description —
		// the single source of truth, not a second copy.
		enQ := wizard.QuestionByID(wizard.InitQuestions(t.TempDir()), wizard.ParticipationQuestionID)
		if enQ == nil {
			t.Fatal("no participation question in the init set")
		}
		if !strings.Contains(out.String(), enQ.Title) {
			t.Errorf("prompt does not carry the init question title")
		}
		if !strings.Contains(out.String(), enQ.Description) {
			t.Errorf("prompt does not carry the init question description")
		}
	})
}

// boolOrText returns the value string a flag of either shape accepts for
// "on": bool flags take "true", string flags take a non-empty marker.
func boolOrText(name, on string) string {
	switch name {
	case "restore", "version", "profile":
		return "some-" + name + "-value"
	default:
		return "true"
	}
}

// TestUpdateFlagInventoryClassified enumerates every flag on the update
// command and fails when one is in neither the mode list nor the modifier
// list, so a new flag forces a decision instead of silently asking.
func TestUpdateFlagInventoryClassified(t *testing.T) {
	classified := map[string]string{}
	for _, name := range updateParticipationModeFlags {
		classified[name] = "mode"
	}
	for _, name := range updateParticipationModifierFlags {
		classified[name] = "modifier"
	}
	if _, ok := classified["yes"]; !ok {
		t.Fatal("--yes is missing from the mode list")
	}

	unclassified := []string{}
	updateCmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}
		if _, ok := classified[f.Name]; !ok {
			unclassified = append(unclassified, f.Name)
		}
	})
	if len(unclassified) != 0 {
		t.Fatalf("update flags in neither list: %v — add each to the mode or the modifier list", unclassified)
	}
}
