package wizard

// SPEC-CLI-WIZARD-RESTRUCTURE-001 — 3-page restructure acceptance tests.
//
// This file carries the NEW assertions introduced by the restructure
// (AC-WIZ-001..005, AC-WIZ-007, AC-WIZ-012a). It is a new file so the M1-M3
// milestones do not collide with the pre-existing wizard test files, which are
// reconciled in place.

import (
	"slices"
	"testing"
)

// Page group labels (plan.md §A.1 D1).
const (
	pageBasic          = "Basic"
	pageModelReport    = "Model & Report"
	pageQualityWorkflw = "Quality & Workflow"
)

// questionIDsInGroup returns the IDs of the questions carrying the given Group
// label, in set order.
func questionIDsInGroup(questions []Question, group string) []string {
	var ids []string
	for i := range questions {
		if questions[i].Group == group {
			ids = append(ids, questions[i].ID)
		}
	}
	return ids
}

// expectedGroupCount is an independent restatement of the buildFormGroups
// partition contract: consecutive UNCONDITIONAL questions sharing a Group label
// merge into one huh group; every CONDITIONAL question becomes its own group.
// Restating the rule here (rather than hardcoding a number) keeps the test
// valid as questions are removed across milestones while still verifying that
// buildFormGroups actually merges.
func expectedGroupCount(questions []Question) int {
	count, pending := 0, 0
	pendingLabel := ""
	flush := func() {
		if pending > 0 {
			count++
			pending = 0
		}
	}
	for i := range questions {
		q := &questions[i]
		if q.Condition != nil {
			flush()
			count++
			continue
		}
		if pending > 0 && q.Group != pendingLabel {
			flush()
		}
		pendingLabel = q.Group
		pending++
	}
	flush()
	return count
}

// TestInitPages_Membership pins AC-WIZ-002: the three topic pages and their
// exact membership + order.
func TestInitPages_Membership(t *testing.T) {
	t.Parallel()
	questions := InitQuestions(t.TempDir())

	cases := []struct {
		page string
		want []string
	}{
		{pageBasic, []string{"conversation_language", "user_name", "project_name"}},
		{pageModelReport, []string{"model_policy", "report_format"}},
		{pageQualityWorkflw, []string{
			"lsp_enabled", "enforce_quality", "project_mode",
			"design_enabled", "claude_design_enabled",
		}},
	}
	for _, tc := range cases {
		got := questionIDsInGroup(questions, tc.page)
		if !slices.Equal(got, tc.want) {
			t.Errorf("page %q membership = %v, want %v", tc.page, got, tc.want)
		}
	}

	// design_enabled MUST precede claude_design_enabled so huh has the design
	// answer before evaluating the nested hide func (plan.md §A.1 D1).
	idx := func(id string) int {
		for i := range questions {
			if questions[i].ID == id {
				return i
			}
		}
		return -1
	}
	if d, c := idx("design_enabled"), idx("claude_design_enabled"); d < 0 || c < 0 || d >= c {
		t.Errorf("design_enabled (%d) must precede claude_design_enabled (%d)", d, c)
	}
}

// TestInitPages_MergeIntoOneGroupPerPage pins the AC-WIZ-001 page-group half:
// each topic page's unconditional questions collapse into exactly ONE huh
// group, and buildFormGroups honours the partition contract.
func TestInitPages_MergeIntoOneGroupPerPage(t *testing.T) {
	t.Parallel()
	result := &WizardResult{}
	locale := ""
	questions := InitQuestions(t.TempDir())

	groups := buildFormGroups(questions, result, &locale)
	if want := expectedGroupCount(questions); len(groups) != want {
		t.Errorf("buildFormGroups yielded %d groups, want %d (partition contract)", len(groups), want)
	}

	// Each page's unconditional questions must form a single contiguous
	// same-label run — otherwise the page would split across huh groups.
	for _, page := range []string{pageBasic, pageModelReport, pageQualityWorkflw} {
		runs, inRun := 0, false
		for i := range questions {
			q := &questions[i]
			match := q.Group == page && q.Condition == nil
			if match && !inRun {
				runs++
			}
			inRun = match
		}
		if runs != 1 {
			t.Errorf("page %q unconditional questions form %d runs, want exactly 1 (page would split)", page, runs)
		}
	}
}

// TestPage3_NoStandardModeGate pins AC-WIZ-003 + the C3 half of AC-WIZ-001: the
// Quality & Workflow questions are visible with NO gate, and the nested design
// condition survives with StandardMode dropped from it.
func TestPage3_NoStandardModeGate(t *testing.T) {
	t.Parallel()
	questions := InitQuestions(t.TempDir())

	for _, id := range []string{"lsp_enabled", "enforce_quality", "project_mode", "design_enabled"} {
		q := QuestionByID(questions, id)
		if q == nil {
			t.Fatalf("%s missing from the init question set", id)
		}
		if q.Condition != nil {
			t.Errorf("%s must be unconditional (no StandardMode gate) so it merges into Page 3", id)
		}
	}

	cd := QuestionByID(questions, "claude_design_enabled")
	if cd == nil {
		t.Fatal("claude_design_enabled missing from the init question set")
	}
	if cd.Condition == nil {
		t.Fatal("claude_design_enabled must stay conditional (nested on design_enabled)")
	}
	// The condition collapses from (StandardMode && DesignEnabled) to
	// DesignEnabled: it must now be TRUE with DesignEnabled alone.
	if !cd.Condition(&WizardResult{DesignEnabled: true}) {
		t.Error("claude_design_enabled must be visible when DesignEnabled=true regardless of StandardMode")
	}
	if cd.Condition(&WizardResult{DesignEnabled: false}) {
		t.Error("claude_design_enabled must be hidden when DesignEnabled=false")
	}
}

// TestReconfigureMembershipExcludesPage3 pins AC-WIZ-012a: the Page-3 questions
// must NOT leak into the `moai update --reconfigure` set.
func TestReconfigureMembershipExcludesPage3(t *testing.T) {
	t.Parallel()
	reconf := ReconfigureQuestions(t.TempDir())

	for _, id := range []string{
		"lsp_enabled", "enforce_quality", "project_mode",
		"design_enabled", "claude_design_enabled",
	} {
		if QuestionByID(reconf, id) != nil {
			t.Errorf("Page-3 question %q leaked into ReconfigureQuestions", id)
		}
	}
	// The pre-restructure member set is retained (Basic + Model + Git).
	for _, id := range []string{
		"conversation_language", "user_name", "project_name",
		"model_policy", "report_format",
		"git_mode", "git_provider", "gitlab_instance_url",
		"github_username", "github_token", "gitlab_username", "gitlab_token",
	} {
		if QuestionByID(reconf, id) == nil {
			t.Errorf("ReconfigureQuestions lost pre-restructure member %q", id)
		}
	}
}

// TestBasicPage_LocaleLiveRender pins AC-WIZ-004: changing conversation_language
// re-renders the SIBLING Page-1 questions (they now share one group with it).
func TestBasicPage_LocaleLiveRender(t *testing.T) {
	t.Parallel()
	questions := InitQuestions(t.TempDir())
	result := &WizardResult{}
	locale := "en"

	before := map[string]string{}
	for _, id := range []string{"user_name", "project_name"} {
		before[id] = GetLocalizedQuestion(QuestionByID(questions, id), locale).Title
	}

	saveAnswer("conversation_language", "ko", result, &locale)
	if locale != "ko" {
		t.Fatalf("live locale = %q, want %q", locale, "ko")
	}
	for _, id := range []string{"user_name", "project_name"} {
		got := GetLocalizedQuestion(QuestionByID(questions, id), locale).Title
		if got == before[id] || got == "" {
			t.Errorf("%s title did not re-render after the locale switch (still %q)", id, got)
		}
	}
}
