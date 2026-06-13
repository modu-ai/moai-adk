package template

import (
	"io/fs"
	"path"
	"regexp"
	"strings"
	"testing"
)

// TestLanguageRuleRelativeLinksExist verifies that every shipped language rule
// under .claude/rules/moai/languages/ is free of dead relative Markdown links.
//
// Sentinel: LANGUAGE_RULE_BROKEN_RELATIVE_LINK
// Source: GitHub issue #1056 — csharp.md (and six sibling language rules:
// kotlin, elixir, cpp, scala, flutter, swift) pointed their "Module Index"
// sections at non-existent modules/ subfiles plus reference.md / examples.md.
// Those files were never shipped, so each link 404'd. The fix replaced the dead
// link lists with self-contained "Coverage Areas" prose. This test is the
// generalized regression guard: it scans ALL language rules (table-driven over
// the discovered file set) rather than a single hard-coded language, so a future
// language rule that reintroduces a dead relative link is caught immediately.
func TestLanguageRuleRelativeLinksExist(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const langDir = ".claude/rules/moai/languages"

	entries, err := fs.ReadDir(fsys, langDir)
	if err != nil {
		t.Fatalf("ReadDir(%q) error: %v", langDir, err)
	}

	// Collect every *.md language rule into a table so each is a named subtest.
	var rulePaths []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		rulePaths = append(rulePaths, path.Join(langDir, e.Name()))
	}

	if len(rulePaths) == 0 {
		t.Fatalf("no language rule .md files found under %q", langDir)
	}

	// Markdown inline link: [text](target)
	linkRE := regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

	for _, rulePath := range rulePaths {
		t.Run(path.Base(rulePath), func(t *testing.T) {
			t.Parallel()

			data, readErr := fs.ReadFile(fsys, rulePath)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", rulePath, readErr)
			}

			baseDir := path.Dir(rulePath)
			matches := linkRE.FindAllStringSubmatch(string(data), -1)

			for _, match := range matches {
				target := strings.TrimSpace(match[1])
				if target == "" {
					continue
				}
				// Skip external links and intra-document anchors.
				if strings.HasPrefix(target, "http://") ||
					strings.HasPrefix(target, "https://") ||
					strings.HasPrefix(target, "#") {
					continue
				}
				// Strip any anchor fragment from a relative link (e.g. foo.md#bar).
				if idx := strings.IndexByte(target, '#'); idx >= 0 {
					target = target[:idx]
				}
				if target == "" {
					continue
				}

				resolved := path.Clean(path.Join(baseDir, target))
				if _, statErr := fs.Stat(fsys, resolved); statErr != nil {
					t.Errorf(
						"LANGUAGE_RULE_BROKEN_RELATIVE_LINK: %s links to %q, but embedded template path %q does not exist",
						rulePath,
						target,
						resolved,
					)
				}
			}
		})
	}
}
