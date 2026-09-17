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

func TestMissionGovernorPermissionBoundary(t *testing.T) {
	data, err := os.ReadFile("../../.claude/agents/moai/mission-governor.md")
	if err != nil {
		t.Fatal(err)
	}
	front := strings.SplitN(string(data), "---", 3)
	if len(front) != 3 {
		t.Fatal("mission-governor frontmatter missing")
	}
	for _, forbidden := range []string{"Bash", "Write", "Edit", "Agent"} {
		if strings.Contains(front[1], forbidden) {
			t.Fatalf("mission-governor frontmatter grants %s: %s", forbidden, front[1])
		}
	}
}
