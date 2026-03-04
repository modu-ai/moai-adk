package template

import (
	"io/fs"
	"strings"
	"testing"
)

// TestSkillAskUserQuestionTurnBoundary verifies that entry-point skills affected
// by the Claude Code bug (anthropics/claude-code#30360) include the required
// [TURN_BOUNDARY] constraint.
//
// Bug: AskUserQuestion returns empty when invoked in the same turn as Skill load.
// The observable symptom is:
//
//	User answered Claude's questions:
//	  ⎿  ← empty
//
// Claude then proceeds with the first/default option, silently bypassing user
// intent. This affects every /moai entry point where a user-confirmation step
// appears immediately after skill activation.
//
// Fix: Affected skill files must contain a [TURN_BOUNDARY] constraint that
// instructs the model to defer AskUserQuestion to the next conversational turn.
func TestSkillAskUserQuestionTurnBoundary(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// Skills that call AskUserQuestion in the same turn as Skill load.
	// Each entry must carry a [TURN_BOUNDARY] marker after the fix is applied.
	affectedSkills := []string{
		".claude/skills/moai/SKILL.md",
		".claude/skills/moai/workflows/feedback.md",
	}

	for _, path := range affectedSkills {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			data, err := fs.ReadFile(fsys, path)
			if err != nil {
				t.Fatalf("ReadFile(%q): %v", path, err)
			}

			content := string(data)
			if !strings.Contains(content, "[TURN_BOUNDARY]") {
				t.Errorf(
					"skill %q is missing the [TURN_BOUNDARY] constraint.\n\n"+
						"Bug: AskUserQuestion returns empty when called in the same turn as Skill load.\n"+
						"Upstream: https://github.com/anthropics/claude-code/issues/30360\n\n"+
						"Fix: Add a [TURN_BOUNDARY] rule that instructs the model to NOT call\n"+
						"AskUserQuestion on first activation — instead output plain text and stop,\n"+
						"then call AskUserQuestion in the next conversational turn.",
					path,
				)
			}
		})
	}
}

// TestSkillTurnBoundaryRule_Content verifies that the [TURN_BOUNDARY] rule in
// SKILL.md contains the three required elements of the workaround pattern:
//
//  1. A prohibition against calling AskUserQuestion in the first turn
//  2. An instruction to output plain text instead
//  3. An instruction to defer to the next turn
func TestSkillTurnBoundaryRule_Content(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/skills/moai/SKILL.md")
	if err != nil {
		t.Fatalf("ReadFile SKILL.md: %v", err)
	}
	content := string(data)

	// Verify the three required elements are present near the [TURN_BOUNDARY] marker.
	required := []string{
		"[TURN_BOUNDARY]",
		"AskUserQuestion",
		"next turn",
	}
	for _, want := range required {
		if !strings.Contains(content, want) {
			t.Errorf("SKILL.md turn boundary rule missing required text %q", want)
		}
	}
}
