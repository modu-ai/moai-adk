package template

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// TestSkillAuthoringRuleGate_SkipsWorkflowBodiesButMatchesSkillFiles verifies
// issue #1059: the skill-authoring rule must not auto-load when MoAI reads its
// own workflow bodies under .claude/skills/moai/workflows/, but it must still
// load when a real SKILL.md is touched.
func TestSkillAuthoringRuleGate_SkipsWorkflowBodiesButMatchesSkillFiles(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".claude/rules/moai/development/skill-authoring.md")
	if err != nil {
		t.Fatalf("ReadFile(skill-authoring.md) error: %v", err)
	}

	fm, _, parseErr := parseFrontmatterAndBody(string(data))
	if parseErr != "" {
		t.Fatalf("frontmatter parse error: %s", parseErr)
	}

	patterns := fm["paths"]
	if patterns == "" {
		t.Fatal("skill-authoring.md must declare a paths frontmatter gate")
	}

	workflowBody := ".claude/skills/moai/workflows/run.md"
	if matchesPathFrontmatter(patterns, workflowBody) {
		t.Fatalf("SKILL_AUTHORING_RULE_FALSE_POSITIVE: paths=%q must not match built-in workflow body %q", patterns, workflowBody)
	}

	userSkill := ".claude/skills/custom-example/SKILL.md"
	if !matchesPathFrontmatter(patterns, userSkill) {
		t.Fatalf("SKILL_AUTHORING_RULE_FALSE_NEGATIVE: paths=%q must still match authored skill file %q", patterns, userSkill)
	}
}

func matchesPathFrontmatter(patterns, target string) bool {
	for _, pattern := range strings.Split(patterns, ",") {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if globToRegexp(pattern).MatchString(target) {
			return true
		}
	}
	return false
}

func globToRegexp(pattern string) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(pattern); {
		if strings.HasPrefix(pattern[i:], "**/") {
			b.WriteString(`(?:.*/)?`)
			i += 3
			continue
		}
		switch pattern[i] {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				b.WriteString(".*")
				i += 2
				continue
			}
			b.WriteString(`[^/]*`)
		case '?':
			b.WriteString(`[^/]`)
		default:
			b.WriteString(regexp.QuoteMeta(pattern[i : i+1]))
		}
		i++
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String())
}
