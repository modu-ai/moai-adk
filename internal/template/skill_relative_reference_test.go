package template_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestSubAgentReferenceRelativeLinksResolve measures that every relative
// path reference inside the sub-agent reference family actually resolves on
// disk.
//
// Measured 2026-10-07 (card t1540): sub-agent-examples.md pointed readers at
// `../../moai-foundation-core/modules/agents-reference.md` — one level short
// of the sibling skill (the target lives at
// .claude/skills/moai-foundation-core/modules/agents-reference.md, three
// levels up from reference/sub-agents/). The broken pointer survived in all
// three mirror copies (template source, deployed .claude, plugin mirror)
// because nothing resolved relative references mechanically.
//
// Scope: the moai-foundation-cc/reference/sub-agents family in the template
// source — the SSOT the deployed mirrors derive from. The regex catches both
// reference shapes the family uses: markdown links `](../...)` and
// backtick-quoted prose pointers “ > `../...` “.
func TestSubAgentReferenceRelativeLinksResolve(t *testing.T) {
	t.Parallel()

	root := hocProjectRoot(t)
	familyDir := filepath.Join(root, "internal", "template", "templates", ".claude", "skills", "moai-foundation-cc", "reference", "sub-agents")

	entries, err := os.ReadDir(familyDir)
	if err != nil {
		t.Fatalf("read sub-agent reference family dir %s: %v", familyDir, err)
	}

	// Relative references: a `](...)` markdown link target or a backtick
	// quoted path, both starting with ./ or ../.
	refRe := regexp.MustCompile("(?:`|\\]\\()((?:\\.\\./|\\./)[A-Za-z0-9_./-]+)")

	files := 0
	refs := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}
		files++
		filePath := filepath.Join(familyDir, name)
		data, readErr := os.ReadFile(filePath)
		if readErr != nil {
			t.Fatalf("read %s: %v", name, readErr)
		}

		for lineNo, line := range strings.Split(string(data), "\n") {
			for _, m := range refRe.FindAllStringSubmatch(line, -1) {
				refs++
				ref := m[1]
				target := filepath.Join(familyDir, ref)
				if _, statErr := os.Stat(target); statErr != nil {
					t.Errorf("%s:%d relative reference does not resolve: %s (target %s missing) — fix the path depth or repoint to the real location (card t1540 shape: one more ../ for a sibling skill)",
						name, lineNo+1, ref, ref)
				}
			}
		}
	}

	// Guard-of-the-guard: an empty or moved family dir, or a regex that lost
	// its match shape, would pass vacuously — fail loudly instead.
	if files < 3 {
		t.Fatalf("only %d markdown files found under %s — family path wrong; check before trusting this guard", files, familyDir)
	}
	if refs < 1 {
		t.Fatalf("no relative references found across %d files — regex lost its match shape; check before trusting this guard", files)
	}
}
