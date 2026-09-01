package wizard

// SPEC-INIT-HARNESS-PROMPT-001 — the agent-harness wizard question.
//
// The harness selection ({claude, codex, both}) used to be reachable only
// through `moai init --agent`, so an interactive user was never asked and
// silently received the claude default (spec.md §1). These tests pin the
// question's presence, shape, English source text, capture into WizardResult,
// and the D3 leak guard that keeps it out of `moai update --reconfigure`.
//
// @MX:SPEC: SPEC-INIT-HARNESS-PROMPT-001

import (
	"slices"
	"testing"
)

// TestAgentWiringQuestion_InInitSetWithClosedOptionSet asserts AC-IHP-001: the
// interactive init set carries exactly one harness question, of Select type,
// offering the same closed set the --agent flag accepts, with claude
// pre-selected as the recommended default.
func TestAgentWiringQuestion_InInitSetWithClosedOptionSet(t *testing.T) {
	t.Parallel()
	questions := InitQuestions(t.TempDir())

	var found int
	for i := range questions {
		if questions[i].ID == "agent_wiring" {
			found++
		}
	}
	if found != 1 {
		t.Fatalf("InitQuestions contains %d questions with ID %q, want exactly 1 (AC-IHP-001)", found, "agent_wiring")
	}

	q := QuestionByID(questions, "agent_wiring")
	if q == nil {
		t.Fatal("QuestionByID(InitQuestions, \"agent_wiring\") = nil — the harness question is absent from the init set (AC-IHP-001)")
	}
	if q.Type != QuestionTypeSelect {
		t.Errorf("agent_wiring type = %v, want %v", q.Type, QuestionTypeSelect)
	}

	var values []string
	for _, opt := range q.Options {
		values = append(values, opt.Value)
	}
	if want := []string{"claude", "codex", "both"}; !slices.Equal(values, want) {
		t.Errorf("agent_wiring option values = %v, want %v (the --agent closed set, REQ-IHP-011)", values, want)
	}
	if q.Default != "claude" {
		t.Errorf("agent_wiring default = %q, want %q (recommended default, REQ-IHP-001)", q.Default, "claude")
	}

	// plan.md §B Decision B1: the question is asked unconditionally, and so is
	// mcp_provision. A Condition func here would contradict a settled decision.
	if q.Condition != nil {
		t.Error("agent_wiring must carry no Condition func (plan.md §B Decision B1)")
	}
}

// TestAgentWiringQuestion_PrecedesMCPProvision asserts the §F M3 placement:
// the harness question sits immediately before mcp_provision, so the
// overriding answer is given first (spec.md §4 D3).
func TestAgentWiringQuestion_PrecedesMCPProvision(t *testing.T) {
	t.Parallel()
	questions := Page3Questions(t.TempDir())

	harness, mcp := -1, -1
	for i := range questions {
		switch questions[i].ID {
		case "agent_wiring":
			harness = i
		case "mcp_provision":
			mcp = i
		}
	}
	if harness < 0 {
		t.Fatal("agent_wiring is not in Page3Questions (spec.md §4 D3)")
	}
	if mcp < 0 {
		t.Fatal("mcp_provision is not in Page3Questions")
	}
	if harness != mcp-1 {
		t.Errorf("agent_wiring is at index %d and mcp_provision at %d; the harness question must sit immediately before it", harness, mcp)
	}
	if got := questions[harness].Group; got != "Quality & Workflow" {
		t.Errorf("agent_wiring group = %q, want %q", got, "Quality & Workflow")
	}
}

// TestAgentWiringQuestion_EnglishSourceTextPresent asserts AC-IHP-006b:
// English is the source language and carries no translation table, so the
// question literal's own text IS the en rendering — every field non-empty.
func TestAgentWiringQuestion_EnglishSourceTextPresent(t *testing.T) {
	t.Parallel()
	q := QuestionByID(InitQuestions(t.TempDir()), "agent_wiring")
	if q == nil {
		t.Fatal("agent_wiring question absent (AC-IHP-006b)")
	}
	if q.Title == "" {
		t.Error("agent_wiring Title is empty — English is the source language and has no translation table")
	}
	if q.Description == "" {
		t.Error("agent_wiring Description is empty")
	}
	for i, opt := range q.Options {
		if opt.Label == "" {
			t.Errorf("agent_wiring option %d has an empty Label", i)
		}
		if opt.Desc == "" {
			t.Errorf("agent_wiring option %d (%s) has an empty Desc", i, opt.Value)
		}
	}
}

// TestSaveAnswer_CapturesAgentWiring asserts AC-IHP-002: the select-answer
// handler stores the harness selection on WizardResult, so it is readable
// after the wizard returns.
func TestSaveAnswer_CapturesAgentWiring(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"claude", "codex", "both"} {
		result := &WizardResult{}
		locale := "en"
		saveAnswer("agent_wiring", value, result, &locale)
		if result.AgentWiring != value {
			t.Errorf("after saveAnswer(agent_wiring, %q): WizardResult.AgentWiring = %q, want %q", value, result.AgentWiring, value)
		}
	}
}

// TestReconfigureQuestions_NoHarnessLeak asserts AC-IHP-008: the harness
// question lives in Page3Questions, which ReconfigureQuestions deliberately
// does not splice, so `moai update --reconfigure` keeps the exact ID sequence
// it had at HEAD 2c18091d1.
func TestReconfigureQuestions_NoHarnessLeak(t *testing.T) {
	t.Parallel()
	questions := ReconfigureQuestions(t.TempDir())

	if q := QuestionByID(questions, "agent_wiring"); q != nil {
		t.Error("agent_wiring leaked into ReconfigureQuestions — page-3 questions must not reach `moai update --reconfigure` (AC-WIZ-012a, AC-IHP-008)")
	}

	var got []string
	for i := range questions {
		got = append(got, questions[i].ID)
	}
	// Pinned verbatim at HEAD 2c18091d1 before the change.
	want := []string{
		"conversation_language",
		"user_name",
		"project_name",
		"model_policy",
		"report_format",
		"git_mode",
		"git_provider",
		"gitlab_instance_url",
		"github_username",
		"github_token",
		"gitlab_username",
		"gitlab_token",
	}
	if !slices.Equal(got, want) {
		t.Errorf("ReconfigureQuestions ID sequence = %v, want %v (unchanged from HEAD 2c18091d1)", got, want)
	}
}
