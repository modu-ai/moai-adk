package template

import (
	"io/fs"
	"path"
	"regexp"
	"strings"
	"testing"
)

// TestCSharpLanguageRuleRelativeLinksExist verifies that the shipped C# language
// rule does not contain dead relative Markdown links.
//
// Sentinel: CSHARP_RULE_BROKEN_RELATIVE_LINK
// Source: GitHub issue #1056 — csharp.md pointed at non-existent modules/
// and reference/examples files.
func TestCSharpLanguageRuleRelativeLinksExist(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const rulePath = ".claude/rules/moai/languages/csharp.md"
	data, err := fs.ReadFile(fsys, rulePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error: %v", rulePath, err)
	}

	linkRE := regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	matches := linkRE.FindAllStringSubmatch(string(data), -1)
	baseDir := path.Dir(rulePath)

	for _, match := range matches {
		target := strings.TrimSpace(match[1])
		if target == "" {
			continue
		}
		if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
			continue
		}
		if strings.HasPrefix(target, "#") {
			continue
		}

		resolved := path.Clean(path.Join(baseDir, target))
		if _, err := fs.Stat(fsys, resolved); err != nil {
			t.Errorf(
				"CSHARP_RULE_BROKEN_RELATIVE_LINK: %s links to %q, but embedded template path %q does not exist",
				rulePath,
				target,
				resolved,
			)
		}
	}
}
