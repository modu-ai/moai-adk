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
//
// Issue: https://github.com/modu-ai/moai-adk/issues/1059

// globToRegex translates a Claude-Code-style `paths:` glob to an anchored Go
// regexp. The glob syntax supported here is the gitignore-style subset
// Claude Code's rules engine uses for the `paths:` field:
//
//   - Leading `**/` matches any directory prefix INCLUDING the empty prefix
//     (so `**/foo.md` matches both `foo.md` and `dir/foo.md`).
//   - `**` elsewhere matches any sequence of characters (including `/`).
//   - `*` matches any sequence of non-`/` characters.
//   - any other character is matched literally.
//
// This is intentionally a minimal implementation sufficient for the patterns
// actually used in `skill-authoring.md` and the secondary `coding-standards.md`
// gates.
func globToRegex(glob string) string {
	var b strings.Builder
	b.WriteByte('^')
	i := 0
	for i < len(glob) {
		// Handle the gitignore `**/` prefix specially: it matches any directory
		// prefix INCLUDING empty. Anywhere — not just at the start of the glob —
		// because intermediate `**/` segments behave the same way.
		if i+2 < len(glob) && glob[i] == '*' && glob[i+1] == '*' && glob[i+2] == '/' {
			b.WriteString("(?:.*/)?")
			i += 3
			continue
		}
		c := glob[i]
		switch c {
		case '*':
			if i+1 < len(glob) && glob[i+1] == '*' {
				// Trailing or middle `**` (not followed by `/`) → match anything
				// including `/`.
				b.WriteString(".*")
				i += 2
			} else {
				// `*` → match any non-`/` sequence
				b.WriteString("[^/]*")
				i++
			}
		case '.', '+', '(', ')', '|', '^', '$', '{', '}', '[', ']', '\\', '?':
			// Regex metacharacters → escape them so they match literally.
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

// matchGlob reports whether path matches a single glob.
func matchGlob(glob, path string) bool {
	re := regexp.MustCompile(globToRegex(glob))
	return re.MatchString(path)
}

// matchAnyGlob splits a `paths:` value on commas and returns true if any glob
// matches. The `paths:` field accepts a comma-separated list of globs per the
// MoAI conventions documented in `.claude/rules/moai/development/coding-standards.md`
// § Paths Frontmatter.
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

// TestMatchGlob_Cases unit-tests the glob matcher itself before we use it to
// assert the skill-authoring.md gate behavior.
func TestMatchGlob_Cases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		// Original (broad) gate — matches everything under .claude/skills/
		{"broad matches SKILL.md", "**/.claude/skills/**", ".claude/skills/moai-foundation-cc/SKILL.md", true},
		{"broad matches workflow body", "**/.claude/skills/**", ".claude/skills/moai/workflows/run.md", true},
		{"broad matches nested subskill body", "**/.claude/skills/**", ".claude/skills/moai/workflows/plan/spec-assembly.md", true},

		// Narrowed (fixed) gate — only matches files literally named SKILL.md
		{"narrow matches SKILL.md", "**/.claude/skills/**/SKILL.md", ".claude/skills/moai-foundation-cc/SKILL.md", true},
		{"narrow matches nested SKILL.md", "**/.claude/skills/**/SKILL.md", ".claude/skills/moai/sub/SKILL.md", true},
		{"narrow rejects workflow body", "**/.claude/skills/**/SKILL.md", ".claude/skills/moai/workflows/run.md", false},
		{"narrow rejects plan.md", "**/.claude/skills/**/SKILL.md", ".claude/skills/moai/workflows/plan.md", false},
		{"narrow rejects sync.md", "**/.claude/skills/**/SKILL.md", ".claude/skills/moai/workflows/sync.md", false},
		{"narrow rejects design.md", "**/.claude/skills/**/SKILL.md", ".claude/skills/moai/workflows/design.md", false},
		{"narrow rejects non-skill md outside .claude/skills/", "**/.claude/skills/**/SKILL.md", ".claude/agents/moai/manager-spec.md", false},

		// `*` (single star) must not cross `/`
		{"single star bounded", "*.md", "subdir/file.md", false},
		{"single star matches one segment", "*.md", "file.md", true},

		// Sanity checks
		{"literal match", "CLAUDE.md", "CLAUDE.md", true},
		{"literal mismatch", "CLAUDE.md", "CLAUDE.local.md", false},
		{"dot is literal", "skill-authoring.md", "skill-authoringXmd", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := matchAnyGlob(tc.pattern, tc.path)
			if got != tc.want {
				t.Errorf("matchAnyGlob(%q, %q) = %v, want %v (regex=%q)",
					tc.pattern, tc.path, got, tc.want, globToRegex(tc.pattern))
			}
		})
	}
}

// TestSkillAuthoringPathsGlob_DoesNotMatchWorkflowBodies verifies the issue #1059
// invariant: the `paths:` glob in `skill-authoring.md` MUST NOT match the
// workflow body files MoAI ships under `.claude/skills/moai/workflows/`.
//
// If this test fails, the rule will auto-load on every `/moai` command because
// every workflow command Reads its workflow body, and that read matches the
// over-broad gate.
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

	// Real workflow body paths MoAI ships — reading these MUST NOT trigger the
	// skill-authoring rule, because the user is executing a `/moai` command,
	// not authoring a skill.
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

	// Sanity: a real SKILL.md authoring target MUST still match, so the rule
	// keeps firing when a user actually authors a skill.
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
