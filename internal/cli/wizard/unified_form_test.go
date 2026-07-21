package wizard

// SPEC-CLI-TUX-V3-002 M2c — unified multi-group form tests (REQ-TUX2-006/008,
// AC-TUX2-006/007). The form is driven programmatically with bubbletea v2
// messages (the M2a spike technique): no TTY, no form.Run, cross-platform.

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// formDriver drives a huh v2 form via Update messages, executing returned
// commands with a bounded timeout so sleeping tick commands (cursor blink)
// are abandoned instead of recursing.
type formDriver struct {
	t *testing.T
	m huh.Model
}

func newFormDriver(t *testing.T, f *huh.Form) *formDriver {
	t.Helper()
	d := &formDriver{t: t, m: f}
	d.drain(f.Init())
	d.send(tea.WindowSizeMsg{Width: 80, Height: 40})
	return d
}

func (d *formDriver) send(msg tea.Msg) {
	nm, cmd := d.m.Update(msg)
	d.m = nm
	d.drain(cmd)
}

func (d *formDriver) drain(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	msg := execBoundedCmd(cmd)
	if msg == nil {
		return
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			d.drain(c)
		}
		return
	}
	nm, next := d.m.Update(msg)
	d.m = nm
	d.drain(next)
}

// execBoundedCmd runs a tea.Cmd with a timeout: huh's internal routing
// messages return instantly; sleeping tick commands are abandoned.
func execBoundedCmd(cmd tea.Cmd) tea.Msg {
	ch := make(chan tea.Msg, 1)
	go func() { ch <- cmd() }()
	select {
	case msg := <-ch:
		return msg
	case <-time.After(50 * time.Millisecond):
		return nil
	}
}

func (d *formDriver) typeText(s string) {
	for _, r := range s {
		d.send(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}

func (d *formDriver) enter() { d.send(tea.KeyPressMsg{Code: tea.KeyEnter}) }
func (d *formDriver) down()  { d.send(tea.KeyPressMsg{Code: tea.KeyDown}) }

func (d *formDriver) view() string { return d.m.(*huh.Form).View() }

// TestUnifiedForm_MultiGroupSinglePage asserts the one-question-one-form
// workaround is gone: the Project group renders multiple fields on ONE page
// (REQ-TUX2-006) and the stepper note carries the dynamic denominator.
func TestUnifiedForm_MultiGroupSinglePage(t *testing.T) {
	result := &WizardResult{}
	questions := DefaultQuestions("/tmp/unified-page")
	form := buildUnifiedForm(questions, result, "")
	d := newFormDriver(t, form)

	// Initial page is the Language select (question 1 of 7 visible:
	// conversation_language, user_name, project_name, model_policy, report_format,
	// git_mode, advanced_bridge; git conditionals hidden). The stepper note
	// renders the dynamic denominator "1 / 7" (REQ-TUX2-008).
	if initial := d.view(); !strings.Contains(initial, "1 / 7") {
		t.Errorf("initial stepper note must render dynamic denominator '1 / 7', frame:\n%s", initial)
	}

	// Page 1 is the Language select, page 2 is the Identity (user_name) input;
	// advance past both to reach the unified Project page.
	d.enter() // conversation_language = en -> Identity page
	d.enter() // user_name (empty) -> Project page

	frame := d.view()
	for _, want := range []string{
		"Enter project name",
		"Select model policy",
		"Select report format",
	} {
		if !strings.Contains(frame, want) {
			t.Errorf("Project group page must render %q (unified multi-field page), frame:\n%s", want, frame)
		}
	}
}

// TestUnifiedForm_ConditionalGroupsAppear drives the personal+github path:
// conditional git groups must appear once git_mode is answered, and the
// harvested WizardResult must match the v1 per-question-form behavior.
func TestUnifiedForm_ConditionalGroupsAppear(t *testing.T) {
	result := &WizardResult{}
	questions := DefaultQuestions("/tmp/unified-cond")
	form := buildUnifiedForm(questions, result, "")
	d := newFormDriver(t, form)

	// Page 1 (Language): default "en" selected.
	d.enter() // conversation_language = en -> Identity page
	// Page 2 (Identity): user_name input.
	d.typeText("octo-dev")
	d.enter() // user_name -> Project page

	// Group (Project): project name + 3 selects. The input pre-fills the
	// default (directory basename, v1-preserved behavior) — clear it first,
	// then type a fresh name.
	for range len("unified-cond") {
		d.send(tea.KeyPressMsg{Code: tea.KeyBackspace})
	}
	d.typeText("uniproj")
	d.enter() // project_name -> model_policy
	d.enter() // model_policy (high)
	d.enter() // report_format (html+md) -> next group

	// Group (Git): git_mode manual -> personal (one cursor down).
	frame := d.view()
	if !strings.Contains(frame, "Select Git automation mode") {
		t.Fatalf("expected git_mode group, frame:\n%s", frame)
	}
	d.down()
	d.enter() // git_mode = personal -> conditional git_provider group appears

	frame = d.view()
	if !strings.Contains(frame, "Select your Git provider") {
		t.Fatalf("conditional git_provider group must appear for personal mode, frame:\n%s", frame)
	}
	// Dynamic denominator: base 6 (language, user_name, project_name,
	// model_policy, report_format, git_mode) + git_provider + advanced_bridge = 8;
	// git_provider is question 7. Provider answer pending so github/gitlab
	// sub-questions are still hidden.
	if !strings.Contains(frame, "7 / 8") {
		t.Errorf("git_provider stepper must render '7 / 8' (dynamic), frame:\n%s", frame)
	}
	d.enter() // git_provider = github -> github_username group

	frame = d.view()
	if !strings.Contains(frame, "GitHub username") {
		t.Fatalf("github_username group must appear for github provider, frame:\n%s", frame)
	}
	// Provider answered: github_username + github_token now visible. Total = base
	// 6 + git_provider + github_username + github_token + advanced_bridge = 10;
	// github_username is question 8.
	if !strings.Contains(frame, "8 / 10") {
		t.Errorf("github_username stepper must render '8 / 10' (dynamic), frame:\n%s", frame)
	}
	d.typeText("octocat")
	d.enter() // github_username
	d.enter() // github_token (empty, optional) -> advanced_bridge group

	frame = d.view()
	if !strings.Contains(frame, "advanced settings") {
		t.Fatalf("advanced_bridge group must appear at end of quick mode, frame:\n%s", frame)
	}
	d.enter() // advanced_bridge = No (default) -> form complete

	if form.State != huh.StateCompleted {
		t.Fatalf("form must complete, state=%v", form.State)
	}

	want := WizardResult{
		ConversationLang: "en",
		UserName:         "octo-dev",
		ProjectName:      "uniproj",
		ModelPolicy:      "high",
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
	questions := DefaultQuestions("/tmp/unified-manual")
	form := buildUnifiedForm(questions, result, "")
	d := newFormDriver(t, form)

	d.enter() // conversation_language = en -> Identity page
	d.enter() // user_name (empty) -> Project page
	d.enter() // project_name (keep default) -> model_policy
	d.enter() // model_policy
	d.enter() // report_format -> Git page

	frame := d.view()
	if strings.Contains(frame, "Select your Git provider") {
		t.Fatalf("git_provider must stay hidden before git_mode is answered, frame:\n%s", frame)
	}
	d.enter() // git_mode = manual -> all git conditionals hidden -> advanced_bridge

	frame = d.view()
	if !strings.Contains(frame, "advanced settings") {
		t.Fatalf("advanced_bridge group must appear after manual git_mode, frame:\n%s", frame)
	}
	d.enter() // advanced_bridge = No (default) -> complete

	if form.State != huh.StateCompleted {
		t.Fatalf("form must complete after manual git_mode + advanced_bridge, state=%v", form.State)
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
	questions := DefaultQuestions("/tmp/unified-partition")

	groups := buildFormGroups(questions, result, &locale)
	// DefaultQuestions groups: "Language" (conversation_language) + "Identity"
	// (user_name) + "Project" (project_name, model_policy, report_format) + "Git"
	// (git_mode) + 6 conditional git questions + advanced_bridge (conditional) =
	// 1 + 1 + 1 + 1 + 6 + 1 = 11 groups.
	if len(groups) != 11 {
		t.Errorf("expected 11 groups (Language, Identity, Project, Git, 6 git conditionals, advanced_bridge), got %d", len(groups))
	}
	for i, g := range groups {
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
