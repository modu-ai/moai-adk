// backlog_json_disclosure_mirror_test.go — SPEC-BACKLOG-JSON-DISCLOSURE-001
// (card t395), AC-BJD-015 / AC-BJD-016: the distributed copy carries the
// same repair, and the binary carries the distributed copy.
//
// THREE files carried FOUR historical claims: `workflows/todo.md` held both
// the primary-checkout assertion and the home-fallback one. The home-scoped
// SQLite layout removes both old path shapes; two patterns remain necessary
// because neither can detect the other by construction.
package template

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// backlogJSONDisclosureMirroredFiles are the three files the repair
// touches, as deployment-relative paths.
var backlogJSONDisclosureMirroredFiles = []string{
	".claude/skills/moai/SKILL.md",
	".claude/skills/moai/workflows/todo.md",
	".claude/skills/moai-kanban-foreman/SKILL.md",
}

// TestBacklogJSONDisclosure_EmbeddedTemplatesMatchSource — AC-BJD-016. The
// binary was rebuilt after the template edit, so each embedded copy is
// byte-for-byte its `internal/template/templates/` source.
func TestBacklogJSONDisclosure_EmbeddedTemplatesMatchSource(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded templates: %v", err)
	}
	for _, rel := range backlogJSONDisclosureMirroredFiles {
		embedded, err := fs.ReadFile(fsys, rel)
		if err != nil {
			t.Errorf("embedded %s: %v", rel, err)
			continue
		}
		source, err := os.ReadFile(filepath.Join("templates", rel))
		if err != nil {
			t.Errorf("source %s: %v", rel, err)
			continue
		}
		if string(embedded) != string(source) {
			t.Errorf("%s: the embedded copy differs from its template source — the binary predates the edit (run `make build`)", rel)
		}
	}
}

// TestBacklogJSONDisclosure_TemplateMirrorIsComplete — AC-BJD-015, run as
// its two enumerated patterns rather than one.
func TestBacklogJSONDisclosure_TemplateMirrorIsComplete(t *testing.T) {
	// Pattern 1: the former project-local path. The home-scoped layout leaves
	// no current deployment claim using it.
	primary := regexp.MustCompile(`state/todo/backlog\.json`)
	// Pattern 2: the home-fallback site. Pattern 1 cannot see this one.
	fallback := regexp.MustCompile(`moai/todo/<project-key>/backlog\.json`)

	var primaryHits, fallbackHits []string
	err := filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		raw, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		for i, line := range strings.Split(string(raw), "\n") {
			loc := filepath.ToSlash(path) + ":" + strconv.Itoa(i+1)
			if primary.MatchString(line) {
				primaryHits = append(primaryHits, loc+" — "+strings.TrimSpace(line))
			}
			if fallback.MatchString(line) {
				fallbackHits = append(fallbackHits, loc+" — "+strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk templates: %v", err)
	}

	if len(primaryHits) != 0 {
		t.Errorf("pattern 1 (state/todo/backlog.json) expected zero current claims, got %d:\n%s",
			len(primaryHits), strings.Join(primaryHits, "\n"))
	}
	if len(fallbackHits) != 0 {
		t.Errorf("pattern 2 (home-fallback) expected zero matches, got %d:\n%s",
			len(fallbackHits), strings.Join(fallbackHits, "\n"))
	}
}
