package cli

// AC-ITI-001 / AC-ITI-002 (SPEC-INIT-TUX-I18N-001 M4): neither `moai init` nor
// `moai update` may open the profile confirmation or run the profile wizard,
// and the interactive init flow asks the conversation-language question
// exactly once, as its first question.
//
// A revived profile entry can take two shapes, so there are two independent
// observations and neither alone is sufficient:
//
//   - routed through the runProfileSetupFn seam — the seam counter in
//     TestInitUpdateEntry_NeverRunsProfileWizard sees it;
//   - written directly against runProfileSetup, bypassing the seam — only the
//     source scan in TestInitUpdateSource_CarriesNoProfileEntry sees it.
//
// prepareSafeInitHome sets environment variables, so no test in this file may
// run in parallel (Go testing panics if one tries).

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

// profilePathQuestionRecorder records the ids of every question the run
// issued, across BOTH wizards, so the conversation-language question can be
// counted over the whole execution rather than per wizard.
type profilePathQuestionRecorder struct {
	ids []string
}

func (r *profilePathQuestionRecorder) count(id string) int {
	n := 0
	for _, got := range r.ids {
		if got == id {
			n++
		}
	}
	return n
}

// installProfileSeamSpy swaps the profile-wizard seam for a counter and
// returns a pointer to the call count. The spy never runs the real wizard, so
// no test reaches a huh form. When rec is non-nil the spy records the one
// conversation-language question the real wizard issues as its Step 1
// ("Select your language", profile_setup.go).
func installProfileSeamSpy(t *testing.T, rec *profilePathQuestionRecorder) *int {
	t.Helper()
	calls := 0
	orig := runProfileSetupFn
	runProfileSetupFn = func(_ *cobra.Command, _ []string) error {
		calls++
		if rec != nil {
			rec.ids = append(rec.ids, "conversation_language")
		}
		return nil
	}
	t.Cleanup(func() { runProfileSetupFn = orig })
	return &calls
}

// profilePathWizardResult is the injected init-wizard answer set: the four
// kept answers plus the fixed defaults RunWithDefaults seeds.
func profilePathWizardResult() *wizard.WizardResult {
	return &wizard.WizardResult{
		ConversationLang:          "en",
		UserName:                  "tester",
		AgentWiring:               "claude",
		AutonomyTier:              "semi-auto",
		LSPEnabled:                true,
		EnforceQuality:            true,
		CoverageExemptionsEnabled: false,
		DesignEnabled:             true,
		ClaudeDesignEnabled:       true,
	}
}

// runInitForProfilePath executes the real runInit into a fresh temp project.
func runInitForProfilePath(t *testing.T, interactive, nonInteractive bool) error {
	t.Helper()
	prepareSafeInitHome(t)

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return interactive }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		return profilePathWizardResult(), nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })

	cmd := newInitTestCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	if nonInteractive {
		if err := cmd.Flags().Set("non-interactive", "true"); err != nil {
			t.Fatalf("set --non-interactive: %v", err)
		}
	}
	return runInit(cmd, []string{filepath.Join(t.TempDir(), "profile-path-proj")})
}

// runUpdateForProfilePath executes the real runUpdate entry. --check with a
// nil dependency set returns right after the entry block, so the run exercises
// the entry without touching a project tree.
func runUpdateForProfilePath(t *testing.T, interactive, yes bool) error {
	t.Helper()
	prepareSafeInitHome(t)

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return interactive }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	cmd := &cobra.Command{Use: "update"}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.Flags().Bool("check", true, "")
	cmd.Flags().Bool("shell-env", false, "")
	cmd.Flags().Bool("config", false, "")
	cmd.Flags().Bool("binary", false, "")
	cmd.Flags().Bool("templates-only", false, "")
	cmd.Flags().Bool("yes", yes, "")
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("verbose", false, "")
	return runUpdate(cmd, nil)
}

// TestInitUpdateEntry_NeverRunsProfileWizard is the AC-ITI-001 behavioural
// half: across the interactive-stdin seam (true / false) crossed with the flag
// combinations, the profile-wizard seam is invoked zero times.
func TestInitUpdateEntry_NeverRunsProfileWizard(t *testing.T) {
	flows := []struct {
		name string
		exec func(t *testing.T, interactive bool) error
	}{
		{"init", func(t *testing.T, interactive bool) error {
			return runInitForProfilePath(t, interactive, false)
		}},
		{"init --non-interactive", func(t *testing.T, interactive bool) error {
			return runInitForProfilePath(t, interactive, true)
		}},
		{"update", func(t *testing.T, interactive bool) error {
			return runUpdateForProfilePath(t, interactive, false)
		}},
		{"update --yes", func(t *testing.T, interactive bool) error {
			return runUpdateForProfilePath(t, interactive, true)
		}},
	}
	for _, flow := range flows {
		for _, interactive := range []bool{true, false} {
			name := flow.name + "/stdin=" + map[bool]string{true: "tty", false: "pipe"}[interactive]
			t.Run(name, func(t *testing.T) {
				calls := installProfileSeamSpy(t, nil)
				if err := flow.exec(t, interactive); err != nil {
					t.Fatalf("%s: %v", flow.name, err)
				}
				if *calls != 0 {
					t.Errorf("profile-wizard seam calls = %d, want 0 (REQ-ITI-001: %s carries no profile entry)", *calls, flow.name)
				}
			})
		}
	}
}

// TestInitUpdateSource_CarriesNoProfileEntry is the AC-ITI-001 source half —
// the same judgement the RED ledger L2/L4 rows make with `git grep`, so a
// revived call written directly against runProfileSetup (bypassing the seam)
// is still caught. The two controls make the zeros attributable: an unreadable
// file or a mistyped needle would zero the whole table, and they do not.
func TestInitUpdateSource_CarriesNoProfileEntry(t *testing.T) {
	read := func(name string) string {
		t.Helper()
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		return string(b)
	}

	removed := []struct{ file, needle string }{
		{"init.go", "No profile found"},
		{"update.go", "No profile found"},
		{"init.go", "runProfileSetup("},
		{"update.go", "runProfileSetup("},
	}
	for _, tc := range removed {
		if got := strings.Count(read(tc.file), tc.needle); got != 0 {
			t.Errorf("%s contains %q %d time(s), want 0 (REQ-ITI-001)", tc.file, tc.needle, got)
		}
	}

	// Control 1: the needle-bearing scan is live — init.go still carries the
	// cancellation line, so a zero above is a removal, not an empty read.
	if got := strings.Count(read("init.go"), "Initialization cancelled"); got < 1 {
		t.Errorf("control: init.go contains %q %d time(s), want at least 1", "Initialization cancelled", got)
	}
	// Control 2: the explicit entry survives — profile.go still names
	// runProfileSetup exactly once, behind the seam (REQ-ITI-002).
	if got := strings.Count(read("profile.go"), "runProfileSetup("); got != 1 {
		t.Errorf("control: profile.go contains %q %d time(s), want exactly 1", "runProfileSetup(", got)
	}
}

// TestInitInteractive_AsksConversationLanguageExactlyOnce is AC-ITI-002: over
// the whole interactive init run, the conversation-language question is issued
// once and is the first question the init wizard asks.
func TestInitInteractive_AsksConversationLanguageExactlyOnce(t *testing.T) {
	prepareSafeInitHome(t)

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	rec := &profilePathQuestionRecorder{}
	calls := installProfileSeamSpy(t, rec)

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	var wizardIDs []string
	origWizard := runWizardFn
	// The recording seam mirrors production: RunWithDefaults issues exactly
	// InitQuestions(projectRoot), in order.
	runWizardFn = func(rootFlag, _, _ string) (*wizard.WizardResult, error) {
		for _, q := range wizard.InitQuestions(rootFlag) {
			wizardIDs = append(wizardIDs, q.ID)
			rec.ids = append(rec.ids, q.ID)
		}
		return profilePathWizardResult(), nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })

	cmd := newInitTestCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	if err := runInit(cmd, []string{filepath.Join(t.TempDir(), "conv-lang-proj")}); err != nil {
		t.Fatalf("runInit: %v (stderr: %s)", err, errBuf.String())
	}

	if *calls != 0 {
		t.Errorf("profile-wizard seam calls = %d, want 0 (REQ-ITI-002: the wizard starts only from an explicit entry)", *calls)
	}
	if len(wizardIDs) == 0 {
		t.Fatal("the init wizard seam issued no questions; the interactive branch did not run")
	}
	if wizardIDs[0] != "conversation_language" {
		t.Errorf("first init question = %q, want %q", wizardIDs[0], "conversation_language")
	}
	if got := rec.count("conversation_language"); got != 1 {
		t.Errorf("conversation_language questions issued in the whole run = %d, want 1 (issued: %v)", got, rec.ids)
	}
}

// TestProfileExplicitEntries_RunProfileWizardOnce is the AC-ITI-002 tail: both
// explicit entries still reach the profile wizard, exactly once each.
func TestProfileExplicitEntries_RunProfileWizardOnce(t *testing.T) {
	t.Run("moai profile setup", func(t *testing.T) {
		calls := installProfileSeamSpy(t, nil)
		cmd := &cobra.Command{Use: "setup"}
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		if err := profileSetupCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("profile setup: %v", err)
		}
		if *calls != 1 {
			t.Errorf("profile-wizard seam calls = %d, want exactly 1", *calls)
		}
	})

	t.Run("moai profile --setup", func(t *testing.T) {
		calls := installProfileSeamSpy(t, nil)
		cmd := &cobra.Command{Use: "profile"}
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.Flags().BoolP("setup", "s", true, "")
		if err := runProfileCmd(cmd, nil); err != nil {
			t.Fatalf("profile --setup: %v", err)
		}
		if *calls != 1 {
			t.Errorf("profile-wizard seam calls = %d, want exactly 1", *calls)
		}
	})
}
