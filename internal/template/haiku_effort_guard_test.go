package template

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// haikuModelFrontmatterRegex matches a `model: haiku` line in YAML frontmatter.
// The invariant keys on the model FIELD VALUE (haiku), not on hardcoded agent
// names, so it generalizes to any future haiku-tier agent (REQ-HEI-002).
var haikuModelFrontmatterRegex = regexp.MustCompile(`(?m)^model:\s*haiku\b`)

// findProjectRootForHaikuGuard walks up from the test's working directory
// (the package dir internal/template) until it finds go.mod, returning the
// repository root. This lets the guard scan both agent trees by absolute path
// (REQ-HEI-006 / AC-HEI-004) without depending on the caller's cwd.
func findProjectRootForHaikuGuard(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found; cannot determine project root")
		}
		dir = parent
	}
}

// extractAgentFrontmatter returns the YAML frontmatter block (the text between
// the opening `---\n` and the next `\n---`) of a markdown file, or "" when no
// frontmatter is present. Scoping the invariant checks to the frontmatter
// block avoids false positives from body prose that happens to mention
// `model:` or `effort:`.
func extractAgentFrontmatter(content string) string {
	const open = "---\n"
	if !strings.HasPrefix(content, open) {
		return ""
	}
	rest := content[len(open):]
	closingIdx := strings.Index(rest, "\n---")
	if closingIdx == -1 {
		return ""
	}
	return rest[:closingIdx]
}

// TestHaikuAgentsHaveNoEffort enforces the `model: haiku ⇒ no effort` invariant
// (REQ-HEI-002 / REQ-HEI-006 / AC-HEI-004) across BOTH agent trees:
//   - the local mirror at .claude/agents/moai/*.md
//   - the template source at internal/template/templates/.claude/agents/moai/*.md
//
// Haiku does not support the `effort` frontmatter field (per
// code.claude.com/docs/en/model-config: "Models not listed here do not support
// effort"; Haiku is not listed). An `effort:` field on a `model: haiku` agent is
// therefore silently inert — a misleading configuration. This guard fails when
// any file declaring `model: haiku` in EITHER tree also declares an `effort:`
// field, keyed on the model value (not agent names) so a future haiku agent is
// covered automatically.
//
// Sentinel on failure: HAIKU_EFFORT_INERT.
func TestHaikuAgentsHaveNoEffort(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootForHaikuGuard(t)

	agentDirs := []string{
		filepath.Join(projectRoot, ".claude", "agents", "moai"),
		filepath.Join(projectRoot, "internal", "template", "templates", ".claude", "agents", "moai"),
	}

	for _, dir := range agentDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read agents dir %q: %v", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read agent file %q: %v", path, err)
			}
			fm := extractAgentFrontmatter(string(content))
			if fm == "" {
				continue // no frontmatter — invariant does not apply
			}
			if !haikuModelFrontmatterRegex.MatchString(fm) {
				continue // not a haiku agent — invariant does not apply
			}
			// effortLineRegex ((?m)^effort:\s*\S+) is defined in model_policy.go.
			if effortLineRegex.Match([]byte(fm)) {
				t.Errorf("HAIKU_EFFORT_INERT: agent %q declares model: haiku AND an effort: field; "+
					"Haiku does not support effort levels (effort is silently inert). Remove the effort: line.", path)
			}
		}
	}
}
