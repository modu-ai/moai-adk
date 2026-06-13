package template

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// skill_authoring_paths_glob_test.go: reproduction + regression guard for issue #1059.
//
// Background: `.claude/rules/moai/development/skill-authoring.md` gates rule loading
// via a YAML frontmatter `paths:` glob. Because MoAI ships its own `/moai` slash
// commands as **skill bodies** under `.claude/skills/moai/workflows/`, every `/moai`
// command Reads one of those files. If the rule's `paths:` glob is too broad
// (e.g. `**/.claude/skills/**`), Claude Code's native rules engine auto-loads
// the ~9.5 KB `skill-authoring.md` rule on every `/moai` command — pure waste,
// since the rule is meant to fire only when a user authors a SKILL.md.
//
// The fix: narrow the glob to `**/.claude/skills/**/SKILL.md` so only actual
// skill authoring (which always touches the SKILL.md artifact) triggers the rule.
// MoAI workflow body reads (`.claude/skills/moai/workflows/run.md`, etc.) MUST NOT
// trigger it.

func globToRegex(glob string) string {
	var b strings.Builder
	b.WriteByte('^')
	i := 0
	for i < len(glob) {
		if i+2 < len(glob) && glob[i] == '*' && glob[i+1] == '*' && glob[i+2] == '/' {
			b.WriteString("(?:.*/)?")
			i += 3
			continue
		}
		c := glob[i]
		switch c {
		case '*':
			if i+1 < len(glob) && glob[i+1] == '*' {
				b.WriteString(".*")
				i += 2
			} else {
				b.WriteString("[^/]*")
				i++
			}
		case '.', '+', '(', ')', '|', '^', '$', '{', '}', '[', ']', '\\', '?':
			b.WriteByte('\\')
			b.WriteByte(c)
			i++
		default:
			b.WriteByte(c)
			i++
		}
	}
	b.WriteByte('$')
	return b.String()
}

func matchGlob(glob, path string) bool {
	re := regexp.MustCompile(globToRegex(glob))
	return re.MatchString(path)
}

func matchAnyGlob(pathsField, candidate string) bool {
	for _, raw := range strings.Split(pathsField, ",") {
		g := strings.TrimSpace(raw)
		if g == "" {
			continue
		}
		if matchGlob(g, candidate) {
			return true
		}
	}
	return false
}

func TestMatchGlob_Cases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{"broad matches SKILL.md", "**/.claude/skills/**", ".claude/skills/moai-foundation-cc/SKILL.md", true},
		{"broad matches workflow body", "**/.claude/skills/**", ".claude/skills/moai/workflows/run.md", true},
		{"broad matches nested subskill body", "**/.claude/skills/**", ".claude/skills/moai/workflows/plan/spec-assembly.md", true},
		{"narrow matches SKILL.md", "**/.claude/skills/**/SKILL.md", ".claude/skills/moai-foundation-cc/SKILL.md", true},
		{"narrow matches nested SKILL.md", "**/.claude/skills/**/SKILL.md", ".claude/skills/moai/sub/SKILL.md", true},
		{"narrow rejects workflow body", "**/.claude/skills/**/SKILL.md", ".claude/skills/moai/workflows/run.md", false},
		{"narrow rejects plan.md", "**/.claude/skills/**/SKILL.md", ".claude/skills/moai/workflows/plan.md", false},
		{"narrow rejects sync.md", "**/.claude/skills/**/SKILL.md", ".claude/skills/moai/workflows/sync.md", false},
		{"narrow rejects design.md", "**/.claude/skills/**/SKILL.md", ".claude/skills/moai/workflows/design.md", false},
		{"narrow rejects non-skill md outside .claude/skills/", "**/.claude/skills/**/SKILL.md", ".claude/agents/moai/manager-spec.md", false},
		{"single star bounded", "*.md", "subdir/file.md", false},
		{"single star matches one segment", "*.md", "file.md", true},
		{"literal match", "CLAUDE.md", "CLAUDE.md", true},
		{"literal mismatch", "CLAUDE.md", "CLAUDE.local.md", false},
		{"dot is literal", "skill-authoring.md", "skill-authoringXmd", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := matchAnyGlob(tc.pattern, tc.path)
			if got != tc.want {
				t.Errorf("matchAnyGlob(%q, %q) = %v, want %v (regex=%q)", tc.pattern, tc.path, got, tc.want, globToRegex(tc.pattern))
			}
		})
	}
}

func TestSkillAuthoringPathsGlob_DoesNotMatchWorkflowBodies(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const rulePath = ".claude/rules/moai/development/skill-authoring.md"
	data, readErr := fs.ReadFile(fsys, rulePath)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error: %v", rulePath, readErr)
	}

	fm, _, parseErr := parseFrontmatterAndBody(string(data))
	if parseErr != "" {
		t.Fatalf("frontmatter parse failed for %s: %s", rulePath, parseErr)
	}

	paths, ok := fm["paths"]
	if !ok || paths == "" {
		t.Fatalf("%s: missing or empty paths frontmatter field (was %q)", rulePath, paths)
	}

	mustNotMatch := []string{
		".claude/skills/moai/workflows/plan.md",
		".claude/skills/moai/workflows/run.md",
		".claude/skills/moai/workflows/sync.md",
		".claude/skills/moai/workflows/design.md",
		".claude/skills/moai/workflows/project.md",
		".claude/skills/moai/workflows/fix.md",
		".claude/skills/moai/workflows/loop.md",
	}
	for _, p := range mustNotMatch {
		if matchAnyGlob(paths, p) {
			t.Errorf("issue #1059 regression: paths=%q wrongly matches workflow body %q — every /moai command will auto-load skill-authoring.md", paths, p)
		}
	}

	mustMatch := []string{
		".claude/skills/moai-foundation-cc/SKILL.md",
		".claude/skills/moai-workflow-tdd/SKILL.md",
		".claude/skills/user-custom-skill/SKILL.md",
	}
	for _, p := range mustMatch {
		if !matchAnyGlob(paths, p) {
			t.Errorf("paths=%q fails to match SKILL.md authoring target %q — the rule will not load when it should", paths, p)
		}
	}
}
