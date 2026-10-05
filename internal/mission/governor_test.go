package mission

import (
	"os"
	"strings"
	"testing"
)

func TestAutoMissionQuestionBoundary(t *testing.T) {
	inside := SuppressAutoMissionQuestions(true, "")
	if inside.Blocked || inside.AskUserQuestion || !inside.Proceed {
		t.Fatalf("inside boundary = %+v", inside)
	}
	outside := SuppressAutoMissionQuestions(false, "new authority required")
	if !outside.Blocked || outside.AskUserQuestion || outside.Proceed || outside.Report == "" {
		t.Fatalf("outside boundary = %+v", outside)
	}
}

// TestManagerTodoJudgmentSubRoleBoundary pins the read-only discipline of the
// sealed-snapshot judgment sub-role. The primary mission of manager-todo is
// todo-queue management and legitimately carries Bash/Write/Edit; the judgment
// sub-role's read-only contract therefore lives in the agent body as a prompt
// contract (the former judgment agent carried it in frontmatter mode), and
// this test asserts that contract survives in the body text.
func TestManagerTodoJudgmentSubRoleBoundary(t *testing.T) {
	data, err := os.ReadFile("../../.claude/agents/moai/manager-todo.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	marker := "## Judgment Sub-Role"
	idx := strings.Index(text, marker)
	if idx < 0 {
		t.Fatalf("manager-todo body missing %q section", marker)
	}
	sub := strings.Join(strings.Fields(text[idx:]), " ")
	for _, clause := range []string{
		"Never write files or state",
		"mutate the",
		"dispatch a lane",
		"audit verdict",
		"never applies the",
	} {
		if !strings.Contains(sub, clause) {
			t.Fatalf("judgment sub-role section missing read-only clause %q", clause)
		}
	}
}

// TestManagerTodoJevBoundaryNamesGrade3 pins the Jev decision boundary in the
// agent body: the grade-3 authority list (queue mutation, completion verdict,
// merge approval, operator gates) is present as PROHIBITED uses, and the
// degraded mode is a labelled non-finding with the cycle proceeding on lead
// judgment (REQ-MT-014/015).
func TestManagerTodoJevBoundaryNamesGrade3(t *testing.T) {
	data, err := os.ReadFile("../../.claude/agents/moai/manager-todo.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	idx := strings.Index(text, "## Jev Decision Boundary")
	if idx < 0 {
		t.Fatal("manager-todo body missing the Jev Decision Boundary section")
	}
	section := text[idx:]
	for _, prohibited := range []string{
		"queue mutation",
		"completion verdict",
		"merge approval",
		"operator-gate",
	} {
		if !strings.Contains(section, prohibited) {
			t.Errorf("Jev boundary section missing prohibited authority %q", prohibited)
		}
	}
	for _, want := range []string{
		"display-only",
		"labelled non-finding",
		"lead judgment",
	} {
		if !strings.Contains(section, want) {
			t.Errorf("Jev boundary section missing %q", want)
		}
	}
}
