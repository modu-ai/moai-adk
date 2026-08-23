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
		{pageQualityWorkflw, []string{"project_mode", "worktree_auto_create", "todo_enabled", "feedback_auto_submit", "audit_model", "audit_gate_claude", "audit_gate_codex", "audit_gate_glm", "codex_audit_enabled", "mcp_provision"}},
		{"Autonomy", []string{"autonomy_tier"}},
	}
	for _, tc := range cases {
		got := questionIDsInGroup(questions, tc.page)
		if !slices.Equal(got, tc.want) {
			t.Errorf("page %q membership = %v, want %v", tc.page, got, tc.want)
		}
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

// TestPage3_NoModeGate pins AC-WIZ-003 + the C3 half of AC-WIZ-001: the
// Quality & Workflow questions are visible with NO gate. With the four
// default-true confirms removed (2026-08-03), only project_mode remains on
// page 3, and it must be unconditional.
func TestPage3_NoModeGate(t *testing.T) {
	t.Parallel()
	questions := InitQuestions(t.TempDir())

	pm := QuestionByID(questions, "project_mode")
	if pm == nil {
		t.Fatal("project_mode missing from the init question set")
	}
	if pm.Condition != nil {
		t.Error("project_mode must be unconditional (no mode gate) so it stays on Page 3")
	}
}

// TestReconfigureMembershipExcludesPage3 pins AC-WIZ-012a: the Page-3 questions
// must NOT leak into the `moai update --reconfigure` set.
func TestReconfigureMembershipExcludesPage3(t *testing.T) {
	t.Parallel()
	reconf := ReconfigureQuestions(t.TempDir())

	// Every page-3 question (now just project_mode after the 2026-08-03 removal
	// of the four default-true confirms) must NOT leak into reconfigure.
	for _, q := range Page3Questions(t.TempDir()) {
		if QuestionByID(reconf, q.ID) != nil {
			t.Errorf("Page-3 question %q leaked into ReconfigureQuestions", q.ID)
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

// TestAdvancedBridgeRemoved pins the advanced_bridge half of AC-WIZ-001,
// AC-WIZ-011 and AC-WIZ-013 (plan.md C1/C14/C17): the gate question is absent
// from BOTH question sets, leaves no ko/ja/zh orphan translation, and no longer
// has an answer-capture branch that could reopen the retired path.
func TestAdvancedBridgeRemoved(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// AC-WIZ-001: absent from the init set (and from reconfigure, where it used
	// to sit at the tail after the Git splice).
	if QuestionByID(InitQuestions(root), "advanced_bridge") != nil {
		t.Error("advanced_bridge must be absent from InitQuestions — Page 3 is ungated")
	}
	if QuestionByID(ReconfigureQuestions(root), "advanced_bridge") != nil {
		t.Error("advanced_bridge must be absent from ReconfigureQuestions")
	}

	// AC-WIZ-011: no orphan translation entry survives the removal.
	for _, locale := range localizableLocales {
		langTrans, ok := translations[locale]
		if !ok {
			t.Fatalf("translations for locale %q not found", locale)
		}
		if _, exists := langTrans["advanced_bridge"]; exists {
			t.Errorf("locale %q: orphan translation entry for removed question %q", locale, "advanced_bridge")
		}
	}

	// AC-WIZ-013: the saveBoolAnswer case is gone, so answering the removed ID
	// mutates nothing. The two mode fields it used to write are themselves
	// retired (REQ-WIZ-018), so the assertion is now the stronger "no field
	// changes at all" form rather than a single-field check.
	r := &WizardResult{}
	saveBoolAnswer("advanced_bridge", true, r)
	if *r != (WizardResult{}) {
		t.Errorf("saveBoolAnswer still captures advanced_bridge (%+v) — the case must be gone", *r)
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
