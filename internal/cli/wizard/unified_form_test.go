package wizard

// SPEC-CLI-TUX-V3-002 M2c — unified multi-group form tests (REQ-TUX2-006/008,
// AC-TUX2-006/007). The form is driven programmatically with bubbletea v2
// messages (the M2a spike technique): no TTY, no form.Run, cross-platform.
// The driver itself lives in internal/cli/ptycaptest and is shared with the
// profile-wizard golden tests.

import (
	"errors"
	"strings"
	"testing"

	"charm.land/huh/v2"

	"github.com/modu-ai/moai-adk/internal/cli/ptycaptest"
)

// newFormDriver hands the shared ptycaptest driver to this package's tests
// under the name the call sites use.
func newFormDriver(t *testing.T, f *huh.Form) *ptycaptest.FormDriver {
	return ptycaptest.NewFormDriver(t, f)
}

// TestUnifiedForm_MultiGroupSinglePage asserts the one-question-one-form
// workaround is gone: a topic page renders multiple fields on ONE page
// (REQ-TUX2-006) and the stepper note carries the dynamic denominator.
// Post-restructure the merged pages are "Basic" (3 fields) and
// "Model & Report" (2 fields).
func TestUnifiedForm_MultiGroupSinglePage(t *testing.T) {
	result := &WizardResult{}
	questions := DefaultQuestions("/tmp/unified-page")
	form := buildUnifiedForm(questions, result, "")
	d := newFormDriver(t, form)

	// Initial page is the merged "Basic" group (question 1 of 5 visible:
	// conversation_language, user_name, project_name, model_policy,
	// report_format — the init set asks nothing about Git, and the
	// advanced_bridge gate is retired by C1).
	// The stepper note renders the dynamic denominator "1 / 5" (REQ-TUX2-008).
	if initial := d.View(); !strings.Contains(initial, "1 / 5") {
		t.Errorf("initial stepper note must render dynamic denominator '1 / 5', frame:\n%s", initial)
	}

	// Page 1 "Basic" renders all three of its fields together.
	frame := d.View()
	for _, want := range []string{
		"Select conversation language",
		"Enter your name",
		"Enter project name",
	} {
		if !strings.Contains(frame, want) {
			t.Errorf("Basic group page must render %q (unified multi-field page), frame:\n%s", want, frame)
		}
	}

	// Advance past the three Basic fields to reach the merged Model & Report page.
	d.Enter() // conversation_language = en
	d.Enter() // user_name (empty)
	d.Enter() // project_name (default) -> Model & Report page

	frame = d.View()
	for _, want := range []string{
		"Select model policy",
		"Select report format",
	} {
		if !strings.Contains(frame, want) {
			t.Errorf("Model & Report group page must render %q, frame:\n%s", want, frame)
		}
	}
}

// TestUnifiedForm_ConditionalGroupsAppear drives the personal+github path:
// conditional git groups must appear once git_mode is answered, and the
// harvested WizardResult must match the v1 per-question-form behavior.
func TestUnifiedForm_ConditionalGroupsAppear(t *testing.T) {
	result := &WizardResult{}
	// Git conditionals live on the reconfigure set only.
	questions := ReconfigureQuestions("/tmp/unified-cond")
	form := buildUnifiedForm(questions, result, "")
	d := newFormDriver(t, form)

	// Page 1 (Language): default "en" selected.
	d.Enter() // conversation_language = en -> Identity page
	// Page 2 (Identity): user_name input.
	d.TypeText("octo-dev")
	d.Enter() // user_name -> Project page

	// Group (Project): project name + 3 selects. The input pre-fills the
	// default (directory basename, v1-preserved behavior) — clear it first,
	// then type a fresh name.
	for range len("unified-cond") {
		d.Backspace()
	}
	d.TypeText("uniproj")
	d.Enter() // project_name -> model_policy
	d.Enter() // model_policy (medium — the new default)
	d.Enter() // report_format (html+md) -> next group

	// Group (Git): git_mode manual -> personal (one cursor down).
	frame := d.View()
	if !strings.Contains(frame, "Select Git automation mode") {
		t.Fatalf("expected git_mode group, frame:\n%s", frame)
	}
	d.Down()
	d.Enter() // git_mode = personal -> conditional git_provider group appears

	frame = d.View()
	if !strings.Contains(frame, "Select your Git provider") {
		t.Fatalf("conditional git_provider group must appear for personal mode, frame:\n%s", frame)
	}
	// Dynamic denominator: base 6 (language, user_name, project_name,
	// model_policy, report_format, git_mode) + git_provider = 7;
	// git_provider is question 7. Provider answer pending so github/gitlab
	// sub-questions are still hidden.
	if !strings.Contains(frame, "7 / 7") {
		t.Errorf("git_provider stepper must render '7 / 7' (dynamic), frame:\n%s", frame)
	}
	d.Enter() // git_provider = github -> github_username group

	frame = d.View()
	if !strings.Contains(frame, "GitHub username") {
		t.Fatalf("github_username group must appear for github provider, frame:\n%s", frame)
	}
	// Provider answered: github_username + github_token now visible. Total = base
	// 6 + git_provider + github_username + github_token = 9;
	// github_username is question 8.
	if !strings.Contains(frame, "8 / 9") {
		t.Errorf("github_username stepper must render '8 / 9' (dynamic), frame:\n%s", frame)
	}
	d.TypeText("octocat")
	d.Enter() // github_username
	d.Enter() // github_token (empty, optional) -> form complete

	if form.State != huh.StateCompleted {
		t.Fatalf("form must complete, state=%v", form.State)
	}

	want := WizardResult{
		ConversationLang: "en",
		UserName:         "octo-dev",
		ProjectName:      "uniproj",
		ModelPolicy:      "medium",
		ReportFormat:     "html+md",
		GitMode:          "personal",
		GitProvider:      "github",
		GitHubUsername:   "octocat",
	}
	if *result != want {
		t.Errorf("WizardResult mismatch:\n got: %+v\nwant: %+v", *result, want)
	}
}

// TestUnifiedForm_ManualModeSkipsConditionals asserts the manual git path
// never surfaces provider questions (visibility semantics preserved).
func TestUnifiedForm_ManualModeSkipsConditionals(t *testing.T) {
	result := &WizardResult{}
	questions := ReconfigureQuestions("/tmp/unified-manual")
	form := buildUnifiedForm(questions, result, "")
	d := newFormDriver(t, form)

	d.Enter() // conversation_language = en -> Identity page
	d.Enter() // user_name (empty) -> Project page
	d.Enter() // project_name (keep default) -> model_policy
	d.Enter() // model_policy
	d.Enter() // report_format -> Git page

	frame := d.View()
	if strings.Contains(frame, "Select your Git provider") {
		t.Fatalf("git_provider must stay hidden before git_mode is answered, frame:\n%s", frame)
	}
	d.Enter() // git_mode = manual -> all git conditionals hidden -> complete

	if form.State != huh.StateCompleted {
		t.Fatalf("form must complete after manual git_mode, state=%v", form.State)
	}
	if result.GitMode != "manual" || result.GitProvider != "" {
		t.Errorf("manual path result mismatch: %+v", *result)
	}
}

// TestBuildFormGroups_Partition asserts the partition rule: consecutive
// unconditional questions sharing a Group label merge; conditional questions
// become their own hideable groups.
func TestBuildFormGroups_Partition(t *testing.T) {
	result := &WizardResult{}
	locale := ""

	// Init set: "Basic" (conversation_language, user_name, project_name) +
	// "Model & Report" (model_policy, report_format) = 1 + 1 = 2 groups.
	// No Git groups, and the advanced_bridge group is retired by C1.
	initGroups := buildFormGroups(DefaultQuestions("/tmp/unified-partition"), result, &locale)
	if len(initGroups) != 2 {
		t.Errorf("expected 2 init groups (Basic, Model & Report), got %d", len(initGroups))
	}

	// Reconfigure set adds "Git" (git_mode) + 6 conditional git questions,
	// each its own hideable group = 2 + 1 + 6 = 9 groups.
	groups := buildFormGroups(ReconfigureQuestions("/tmp/unified-partition"), result, &locale)
	if len(groups) != 9 {
		t.Errorf("expected 9 groups (Basic, Model & Report, Git, 6 git conditionals), got %d", len(groups))
	}
	for i, g := range append(initGroups, groups...) {
		if g == nil {
			t.Fatalf("group %d is nil", i)
		}
	}
}

// TestMapFormErr covers the huh error mapping (cancel path preserved).
func TestMapFormErr(t *testing.T) {
	if got := mapFormErr(huh.ErrUserAborted); !errors.Is(got, ErrCancelled) {
		t.Errorf("ErrUserAborted must map to ErrCancelled, got %v", got)
	}
	boom := errors.New("boom")
	got := mapFormErr(boom)
	if !errors.Is(got, boom) || !strings.Contains(got.Error(), "wizard error") {
		t.Errorf("non-abort errors must wrap with 'wizard error', got %v", got)
	}
}
