package cli

// AC-ITI-008 (REQ-ITI-007): the absorbed profile wizard's screens, drawn per
// locale through the REAL option-label bridge, match their goldens, and no
// localizable English string leaks onto a ko/ja/zh screen outside the closed
// exception list X.
//
// The form is constructed per locale. huh v2 key maps and option labels are
// static per form, so the judged run is the saved-locale path the wizard
// actually serves: the form builds in the run's initial locale (the stored
// conversation language) and the language question confirms it. Question
// titles and descriptions additionally re-render mid-form through the locale
// pointer (REQ-ITI-007's in-form path), which the goldens pin via the
// per-locale frames.
//
// Set E (the strings asserted absent) is the en value of every localizable
// text whose target-locale rendering DIFFERS from en — strings that were
// actually translated. A text whose locale rendering equals its en value (the
// schema's fixed empty-option literals, the untranslated model-policy
// composition labels) is its own rendering in every locale, not English
// leakage; naming each of them in X would widen the closed list for texts the
// SPEC never provided translations for. This interpretation is recorded in
// progress.md §E.2 (M5).

import (
	"slices"
	"strings"
	"testing"

	"charm.land/huh/v2"

	"github.com/modu-ai/moai-adk/internal/cli/ptycaptest"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// profileGoldenDir holds the profile wizard's per-locale View() goldens.
const profileGoldenDir = "testdata/profilewizard"

// goldenLocaleAnswers is the pre-selected answer set the goldens render: the
// stored values a configured profile carries.
func goldenLocaleAnswers(loc string) wizard.ProfileResult {
	return wizard.ProfileResult{
		ConversationLang: loc,
		UserName:         "",
		GitCommitLang:    "en",
		CodeCommentLang:  "en",
		DocLang:          "en",
		Model:            "opus[1m]",
		ModelPolicy:      "high",
		EffortLevel:      "high",
		PermissionMode:   "acceptEdits",
		DevelopmentMode:  "tdd",
	}
}

// drawProfileGroups confirms the language question and walks the remaining
// groups, returning each group's ANSI-stripped frame (groups 2..5).
func drawProfileGroups(t *testing.T, loc string) []string {
	t.Helper()
	form := wizard.NewProfileForm(buildProfileOptions(getProfileText(loc)), goldenLocaleAnswers(loc), loc)
	d := ptycaptest.NewFormDriver(t, form)
	d.Enter() // conversation_language = loc -> identity page

	var frames []string
	capture := func() {
		frames = append(frames, ptycaptest.StripANSI(d.View()))
	}
	capture()     // identity page
	d.Enter()     // user_name -> languages page
	capture()     // languages page
	for range 3 { // three language selects -> model page
		d.Enter()
	}
	capture()     // model page
	for range 4 { // model, policy, effort, permission -> project page
		d.Enter()
	}
	capture() // project page
	d.Enter() // development_mode -> complete
	if form.State != huh.StateCompleted {
		t.Fatalf("profile form must complete after the last group, state=%v", form.State)
	}
	return frames
}

// TestProfileWizardGolden_LocaleFrames is AC-ITI-008's golden half: the
// ko/ja/zh walks match their goldens line for line. Rerun with
// -update-golden to rewrite them after an intentional render change.
func TestProfileWizardGolden_LocaleFrames(t *testing.T) {
	for _, loc := range []string{"ko", "ja", "zh"} {
		t.Run(loc, func(t *testing.T) {
			got := strings.Join(drawProfileGroups(t, loc), "\n")
			if err := ptycaptest.CompareGolden(profileGoldenDir, "groups-"+loc, got, *updateViewGoldens); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// goldenExceptions is AC-ITI-008's closed exception list X: the strings that
// legitimately appear on a localized screen. Never widen it — a string
// outside X appearing on a localized screen is a defect to fix.
func goldenExceptions() []string {
	x := []string{
		// Language option labels — native names, never translated.
		"English", "Korean (한국어)", "Japanese (日本語)", "Chinese (中文)",
		// Values drawn where no localized label exists.
		"low", "medium", "high", "xhigh", "max",
		"acceptEdits", "auto", "default", "plan", "bypassPermissions", "dontAsk",
		// Help-line key notations.
		"enter", "tab", "shift+tab", "esc", "↑", "↓", "←/→", "/", "x", "y", "n",
		"ctrl+e", "ctrl+u", "ctrl+d", "ctrl+a", "g/home", "G/end",
		// Proper nouns.
		"MoAI", "Claude",
	}
	x = append(x, template.ModelAliasPickerValues()...)
	for _, mode := range models.ValidDevelopmentModes() {
		x = append(x, string(mode))
	}
	return x
}

// eStringsFor computes set E for one locale: the en values of the profile
// wizard's localizable texts whose locale rendering differs from en. See the
// file comment for the differs-from-en rule.
func eStringsFor(t *testing.T, loc string) []string {
	t.Helper()
	initial := goldenLocaleAnswers("en")
	optsEn := buildProfileOptions(getProfileText("en"))
	optsLoc := buildProfileOptions(getProfileText(loc))
	qsEn := wizard.ProfileQuestions(optsEn, initial)
	qsLoc := wizard.ProfileQuestions(optsLoc, initial)
	if len(qsEn) != len(qsLoc) {
		t.Fatalf("profile question sets diverge: %d en vs %d %s", len(qsEn), len(qsLoc), loc)
	}

	var e []string
	for i := range qsEn {
		enQ := wizard.LocalizeProfileQuestion(&qsEn[i], "en")
		locQ := wizard.LocalizeProfileQuestion(&qsLoc[i], loc)
		if enQ.Title != locQ.Title {
			e = append(e, enQ.Title)
		}
		if enQ.Description != locQ.Description {
			e = append(e, enQ.Description)
		}
		for j, o := range enQ.Options {
			if j < len(locQ.Options) && o.Label != locQ.Options[j].Label {
				e = append(e, o.Label)
			}
		}
	}
	locales := wizard.HelpActionLabels(loc)
	for _, s := range wizard.HelpActionLabels("en") {
		if !slices.Contains(locales, s) {
			e = append(e, s)
		}
	}
	return e
}

// TestProfileWizardGolden_NoEnglishLeak is AC-ITI-008's set-E half: no
// localizable English string (outside the closed list X) may appear on a
// ko/ja/zh screen.
func TestProfileWizardGolden_NoEnglishLeak(t *testing.T) {
	for _, loc := range []string{"ko", "ja", "zh"} {
		t.Run(loc, func(t *testing.T) {
			frames := drawProfileGroups(t, loc)
			e := eStringsFor(t, loc)
			if len(e) == 0 {
				t.Fatal("set E came out empty; the leak assertion would be vacuous")
			}
			x := goldenExceptions()
			for _, s := range e {
				if slices.Contains(x, s) {
					continue
				}
				for gi, frame := range frames {
					if strings.Contains(frame, s) {
						t.Errorf("group %d frame leaks the English %q (not in the closed exception list X); frame:\n%s", gi+2, s, frame)
					}
				}
			}
		})
	}
}
