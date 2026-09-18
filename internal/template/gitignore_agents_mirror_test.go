package template

import (
	"fmt"
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

// checkRootAgentsInvariant applies the same two checks
// TestGitignore_IgnoresSkillMirrorOnly runs against the template source and
// embedded FS — no whole-.agents-root ignore pattern, and the narrow mirror
// pattern present — to an arbitrary line set. It is a pure function, so the
// invariant is expressible without touching disk.
//
// While card t912's repair was outstanding this took a second argument: a
// self-expiring KnownStale marker (the idiom in internal/harness/rosterguard)
// declaring the then-known gap — the root file's `.agents/*` deny-all, and the
// narrow mirror pattern absent from it. That repair and this deletion landed
// in the same merge: the root file now satisfies the invariant outright, the
// marker expired exactly as designed, and the exception branch went with it. A
// forbidden form is now reported unconditionally, because a branch that
// tolerates a declared one is a bypass available to anyone willing to declare.
func checkRootAgentsInvariant(lines []string) []string {
	var bad []string

	// Report every whole-.agents-root line, not just the first — an
	// additional forbidden form must not hide behind an earlier one.
	for _, line := range lines {
		if _, isBad := wholeAgentsRootPatterns[line]; isBad {
			bad = append(bad, fmt.Sprintf(
				"root .gitignore ignores the whole .agents root via %q — narrow the pattern to %q",
				line, mirrorIgnorePattern))
		}
	}

	hasMirror := false
	for _, line := range lines {
		if line == mirrorIgnorePattern {
			hasMirror = true
			break
		}
	}
	if !hasMirror {
		bad = append(bad, fmt.Sprintf(
			"root .gitignore has no %q entry", mirrorIgnorePattern))
	}

	return bad
}

// TestGitignore_RootFollowsNarrowAgentsInvariant is the root-file half of
// AC-CSC-015: TestGitignore_IgnoresSkillMirrorOnly above reads only
// templates/.gitignore and the embedded FS copy, so a forbidden
// whole-.agents-root pattern on the repository root file — which is what
// this project's own git status actually reads day to day — went
// unchecked. Root and template are not required to be identical; only the
// narrow invariant is required on both.
func TestGitignore_RootFollowsNarrowAgentsInvariant(t *testing.T) {
	rootPath := filepath.Join("..", "..", ".gitignore")
	raw, err := os.ReadFile(rootPath)
	if err != nil {
		t.Fatalf("read repository root .gitignore: %v", err)
	}
	lines := gitignoreLines(t, string(raw))
	for _, msg := range checkRootAgentsInvariant(lines) {
		t.Error(msg)
	}
}
