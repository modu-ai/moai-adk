package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

// agentFrontmatterValue returns the value of a top-level frontmatter key, or ""
// when the key is absent.
func agentFrontmatterValue(fm, key string) string {
	for _, line := range strings.Split(fm, "\n") {
		if v, ok := strings.CutPrefix(line, key+":"); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// TestAgentRoleModelEffortDefaults pins the per-role defaults (card t1808):
// the judging roles declare model opus, and no agent definition — live,
// harness, or template mirror — declares effort max or ultra (those levels are
// never a shipped default).
//
// Sentinel on failure: AGENT_ROLE_DEFAULT_VIOLATION.
func TestAgentRoleModelEffortDefaults(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootForHaikuGuard(t)
	agentDirs := []string{
		filepath.Join(projectRoot, ".claude", "agents", "moai"),
		filepath.Join(projectRoot, ".claude", "agents", "harness"),
		filepath.Join(projectRoot, "internal", "template", "templates", ".claude", "agents", "moai"),
	}
	judging := map[string]bool{"plan-auditor": true, "sync-auditor": true, "super-advisor": true}

	scanned, judgedSeen := 0, 0
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
			if effort := agentFrontmatterValue(fm, "effort"); effort == "max" || effort == "ultra" {
				t.Errorf("AGENT_ROLE_DEFAULT_VIOLATION: %q declares effort %q; max/ultra are never a shipped default", path, effort)
			}
			if judging[strings.TrimSuffix(entry.Name(), ".md")] {
				judgedSeen++
				if model := agentFrontmatterValue(fm, "model"); model != "opus" {
					t.Errorf("AGENT_ROLE_DEFAULT_VIOLATION: judging role %q declares model %q, want opus", path, model)
				}
			}
		}
	}
	if scanned == 0 {
		t.Fatal("scanned no agent frontmatter — the guard read nothing")
	}
	if judgedSeen != 2*len(judging) {
		t.Errorf("saw %d judging-role files, want %d (live + template for each of %d roles)", judgedSeen, 2*len(judging), len(judging))
	}
}
