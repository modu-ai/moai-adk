package template

import (
	"io/fs"
	"path"
	"strconv"
	"strings"
	"testing"
)

// skillDirToken is the Claude-Code-only environment variable that expands to
// the invoked skill's own directory. Under a non-Claude harness (codex) no
// process exports it, so every path built from it expands wrong.
//
// The distributed skill tree therefore addresses its own files by
// project-root-relative path (`.claude/skills/<name>/…`), which both harnesses
// resolve to the same file. This guard fails the build if the token is
// reintroduced there.
//
// Scope note: the guard watches the SKILL TREE ONLY
// (`.claude/skills/**` inside internal/template/templates/). It is deliberately
// NOT widened to:
//
//   - the whole template tree — `.claude/rules/moai/development/skill-authoring.md`
//     legitimately names the variable in a table of Claude-Code-provided
//     capabilities, which is a true statement of fact and must survive;
//   - the deployed local `.claude/skills/` copy — that tree is redeployed from
//     the installed binary, so a binary lagging the source tree would make this
//     guard red for a reason other than the one its name states.
const skillDirToken = "CLAUDE_SKILL_DIR"

// skillTreeRoot is the guarded subtree, relative to the embedded templates FS root.
const skillTreeRoot = ".claude/skills"

// TestSkillTreeHasNoClaudeSkillDirToken walks the embedded skill tree and
// reports every line still carrying the harness-specific path token.
//
// The census (matching line count) is reported on both the pass and fail paths
// so a run's output distinguishes "the guard looked and found nothing" from
// "the guard looked at nothing".
func TestSkillTreeHasNoClaudeSkillDirToken(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	filesWalked := 0
	var offenders []string

	walkErr := fs.WalkDir(fsys, skillTreeRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch path.Ext(p) {
		case ".md", ".sh", ".mjs", ".js", ".py", ".yaml", ".yml", ".json":
		default:
			return nil
		}
		data, readErr := fs.ReadFile(fsys, p)
		if readErr != nil {
			return readErr
		}
		filesWalked++
		for i, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, skillDirToken) {
				offenders = append(offenders, formatOffender(p, i+1, line))
			}
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk %s: %v", skillTreeRoot, walkErr)
	}

	// A guard that walked no files proves nothing; treat that as a failure of
	// the guard itself rather than as a clean tree.
	if filesWalked == 0 {
		t.Fatalf("guard walked 0 files under %s — the guard is watching nothing", skillTreeRoot)
	}

	if len(offenders) > 0 {
		t.Errorf("census: %d line(s) carrying %s across %d file(s) walked — expected 0.\n"+
			"Replace ${%s}/X with the project-root-relative form .claude/skills/<skill-name>/X.\n%s",
			len(offenders), skillDirToken, filesWalked, skillDirToken,
			strings.Join(offenders, "\n"))
		return
	}

	t.Logf("census: 0 lines carrying %s across %d files walked under %s", skillDirToken, filesWalked, skillTreeRoot)
}

func formatOffender(p string, line int, text string) string {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) > 120 {
		trimmed = trimmed[:120] + "…"
	}
	return "  " + p + ":" + strconv.Itoa(line) + ": " + trimmed
}
