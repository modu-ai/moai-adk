package template

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestGoalAutoWorkflowContractAndMirrorParity(t *testing.T) {
	source, err := os.ReadFile("../../.claude/skills/moai/workflows/goal.md")
	if err != nil {
		t.Fatal(err)
	}
	mirror, err := os.ReadFile("templates/.claude/skills/moai/workflows/goal.md")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(source, mirror) {
		t.Fatal("goal workflow source/template mirror differ")
	}
	body := string(source)
	start := strings.Index(body, "## `/moai goal --auto`")
	end := strings.Index(body[start+1:], "\n## ")
	if start < 0 || end < 0 {
		t.Fatal("auto workflow section missing or unbounded")
	}
	auto := body[start : start+1+end]
	for _, required := range []string{
		"Capture → Clarify → Organize → Reflect → Engage",
		"versioned sealed contract",
		"Implementation/mission approval",
		"moai goal --auto",
		"moai goal approve",
		"moai goal status",
		"moai goal run",
		"super-advisor",
		"non-binding",
		"mission-governor",
		"deterministic validator",
		"publish → pick → disk dispatch",
		"manager-develop",
		"manager-git",
		"local develop",
		"--no-ff",
		"independent audit",
		"authoritative readback",
		"completion evidence",
		"persisted blocked",
		"active-session-only",
		"external content",
		"moai gpt",
	} {
		if !strings.Contains(auto, required) {
			t.Errorf("auto workflow missing contract token %q", required)
		}
	}
	if strings.Count(auto, "AskUserQuestion") != 1 {
		t.Errorf("auto workflow must have exactly one approval prompt reference, got %d", strings.Count(auto, "AskUserQuestion"))
	}
	if strings.Contains(auto, "moai cc") {
		t.Error("auto workflow recommends the forbidden Claude Code launcher")
	}
}
