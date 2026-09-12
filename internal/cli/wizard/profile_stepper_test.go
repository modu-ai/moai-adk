package wizard

// AC-ITI-009 (REQ-ITI-008): the absorbed profile wizard renders the same
// step-indicator format as the init wizard — on the first line of each
// group's View() (ANSI stripped) the ●/○ count equals the visible question
// count N, and the line ends with "<k> / <N>", k being the group's first
// question's position. The init wizard is the control group: the same rule
// must hold there, so a stepper generalization that changed the init output
// cannot hide.

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/huh/v2"

	"github.com/modu-ai/moai-adk/internal/cli/ptycaptest"
)

// profileStepperOptions builds the option rows the selects render, mirroring
// the real conversation_language shape (4 native-name rows) so the layout
// walk covers real option-list geometry; the labels are otherwise irrelevant
// to the stepper rule under test.
func profileStepperOptions() ProfileOptions {
	lang := []Option{
		{Label: "English", Value: "en", Desc: "English"},
		{Label: "Korean (한국어)", Value: "ko", Desc: "한국어"},
		{Label: "Japanese (日本語)", Value: "ja", Desc: "日本語"},
		{Label: "Chinese (中文)", Value: "zh", Desc: "中文"},
	}
	return ProfileOptions{
		Language:        lang,
		Model:           []Option{{Label: "(runtime default)", Value: ""}, {Label: "opus[1m]", Value: "opus[1m]"}},
		ModelPolicy:     []Option{{Label: "(none)", Value: ""}, {Label: "High", Value: "high"}},
		EffortLevel:     []Option{{Label: "high", Value: "high"}},
		PermissionMode:  []Option{{Label: "plan", Value: "plan"}},
		DevelopmentMode: []Option{{Label: "(project default)", Value: ""}, {Label: "tdd", Value: "tdd"}},
	}
}

// firstStepperLine returns the first non-empty line of an ANSI-stripped frame.
func firstStepperLine(t *testing.T, frame string) string {
	t.Helper()
	for _, line := range strings.Split(frame, "\n") {
		if strings.TrimSpace(line) != "" {
			return line
		}
	}
	t.Fatal("frame has no non-empty line")
	return ""
}

// assertStepperLine applies the AC-ITI-009 rule to one first line. The huh
// frame pads its lines to the viewport width, so the suffix check trims the
// trailing padding first.
func assertStepperLine(t *testing.T, line string, k, n int) {
	t.Helper()
	line = strings.TrimRight(line, " ")
	if got := strings.Count(line, "●") + strings.Count(line, "○"); got != n {
		t.Errorf("first line carries %d ●/○ marks, want N=%d; line: %q", got, n, line)
	}
	if want := fmt.Sprintf("%d / %d", k, n); !strings.HasSuffix(line, want) {
		t.Errorf("first line %q must end with %q", line, want)
	}
}

func TestProfileWizardStepper_SameFormatAsInit(t *testing.T) {
	// Profile half: N = 10 (design.md §2.2, no conditional questions), five
	// groups whose first questions sit at positions 1, 2, 3, 6, 10.
	initial := ProfileResult{
		ConversationLang: "en",
		GitCommitLang:    "en",
		CodeCommentLang:  "en",
		DocLang:          "en",
	}
	form := NewProfileForm(profileStepperOptions(), initial, "en")
	d := ptycaptest.NewFormDriver(t, form)

	profileGroups := []struct {
		questions int // fields on this group's page (enter count to leave it)
		first     int // the group's first question's 1-based position
	}{
		{questions: 1, first: 1},  // profile-language
		{questions: 1, first: 2},  // profile-identity
		{questions: 3, first: 3},  // profile-languages
		{questions: 4, first: 6},  // profile-model
		{questions: 1, first: 10}, // profile-project
	}
	for _, g := range profileGroups {
		assertStepperLine(t, firstStepperLine(t, ptycaptest.StripANSI(d.View())), g.first, 10)
		for range g.questions {
			d.Enter()
		}
	}
	if form.State != huh.StateCompleted {
		t.Fatalf("profile form must complete after the last group, state=%v", form.State)
	}

	// Init control (the same rule with init's own visible N): InitQuestions —
	// what `moai init` runs — renders two pages after the Q5 regroup, Basic
	// (conversation_language + user_name, first at 1) and Agents & Autonomy
	// (agent_wiring + autonomy_tier, first at 3); N = 4.
	result := &WizardResult{}
	initForm := buildUnifiedForm(InitQuestions("/tmp/stepper-control"), result, "")
	id := ptycaptest.NewFormDriver(t, initForm)
	initGroups := []struct {
		questions int
		first     int
	}{
		{questions: 2, first: 1},
		{questions: 2, first: 3},
	}
	for _, g := range initGroups {
		assertStepperLine(t, firstStepperLine(t, ptycaptest.StripANSI(id.View())), g.first, 4)
		for range g.questions {
			id.Enter()
		}
	}
	if initForm.State != huh.StateCompleted {
		t.Fatalf("init form must complete after the last group, state=%v", initForm.State)
	}
}
