package template

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mirrorIgnorePattern is the narrow entry the deployed .gitignore must carry:
// the mirror is a build product, not a source file.
const mirrorIgnorePattern = ".agents/skills/moai*"

// wholeAgentsRootPatterns are the forms that would ignore the entire .agents/
// root. The mirror is the only thing this project generates under it; ignoring
// the root would also hide a user's own .agents/skills/hns-* entries and any
// source file a later milestone puts there.
var wholeAgentsRootPatterns = map[string]struct{}{
	".agents":     {},
	".agents/":    {},
	".agents/*":   {},
	".agents/**":  {},
	"/.agents":    {},
	"/.agents/":   {},
	"/.agents/*":  {},
	"/.agents/**": {},
}

func gitignoreLines(t *testing.T, content string) []string {
	t.Helper()
	var out []string
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out
}

// TestGitignore_IgnoresSkillMirrorOnly covers AC-CSC-015.
func TestGitignore_IgnoresSkillMirrorOnly(t *testing.T) {
	sourcePath := filepath.Join("templates", ".gitignore")
	sourceRaw, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read template .gitignore: %v", err)
	}
	sourceLines := gitignoreLines(t, string(sourceRaw))

	// 1. the narrow mirror pattern is present.
	found := false
	for _, line := range sourceLines {
		if line == mirrorIgnorePattern {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("template .gitignore has no %q entry — every user project would gain the mirror as a commit candidate, and on a Windows checkout each link materializes as a text file", mirrorIgnorePattern)
	}

	// 2. the whole .agents/ root is NOT ignored.
	for _, line := range sourceLines {
		if _, bad := wholeAgentsRootPatterns[line]; bad {
			t.Errorf("template .gitignore ignores the whole .agents root via %q — narrow the pattern to %q", line, mirrorIgnorePattern)
		}
	}

	// 3. the entry is readable through the embedded FS, which is what a user
	//    actually receives. A source-only edit with no rebuild fails here.
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded templates: %v", err)
	}
	embeddedRawContent, err := fs.ReadFile(embedded, ".gitignore")
	if err != nil {
		t.Fatalf("read embedded .gitignore: %v", err)
	}
	embeddedFound := false
	for _, line := range gitignoreLines(t, string(embeddedRawContent)) {
		if line == mirrorIgnorePattern {
			embeddedFound = true
		}
		if _, bad := wholeAgentsRootPatterns[line]; bad {
			t.Errorf("embedded .gitignore ignores the whole .agents root via %q", line)
		}
	}
	if !embeddedFound {
		t.Errorf("embedded .gitignore has no %q entry — the template source was edited without rebuilding the binary", mirrorIgnorePattern)
	}
}
