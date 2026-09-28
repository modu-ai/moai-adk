package template

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// agentModelEffortLineRegex matches a model: or effort: key at the start of a
// frontmatter line.
var agentModelEffortLineRegex = regexp.MustCompile(`(?m)^(model|effort):`)

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

// TestAgentsDeclareNoModelOrEffort enforces the inheritance invariant
// (SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-003, AC-AMI-003): no MoAI agent
// definition declares a model: or effort: frontmatter key, so every subagent
// inherits the main session's model and effort (Claude Code resolves an absent
// model to the main conversation's model and an absent effort to the session's).
// It scans the template source, the local mirror, and the local harness
// agents. It supersedes the former `model: haiku ⇒ no effort` guard, which this
// invariant implies.
//
// Sentinel on failure: AGENT_MODEL_EFFORT_DECLARED.
func TestAgentsDeclareNoModelOrEffort(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootForHaikuGuard(t)

	agentDirs := []string{
		filepath.Join(projectRoot, ".claude", "agents", "moai"),
		filepath.Join(projectRoot, ".claude", "agents", "harness"),
		filepath.Join(projectRoot, "internal", "template", "templates", ".claude", "agents", "moai"),
	}

	scanned := 0
	for _, dir := range agentDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) && strings.HasSuffix(dir, "harness") {
				continue // the local harness tree is optional
			}
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
				continue
			}
			scanned++
			if m := agentModelEffortLineRegex.FindAllString(fm, -1); len(m) > 0 {
				t.Errorf("AGENT_MODEL_EFFORT_DECLARED: agent %q declares %v; subagents inherit the main session's model and effort — remove the key(s)", path, m)
			}
		}
	}
	if scanned == 0 {
		t.Fatal("scanned no agent frontmatter — the guard read nothing")
	}
}
