// migrate_kept_optional_test.go — U-12 (AC-RC30-13, REQ-RC30-001): an optional
// agent or .agents/skills copy that the migration keeps project-side is either
// refreshed in place or named in the update output. A kept copy must not drop
// out of the output without a trace.
package cli

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

// TestKeptOptionalAssetRefresh: an outdated optional-pack agent copy and an
// outdated optional-pack .agents/skills copy, both template-managed with no
// user counterpart installed, are each refreshed or named in the output.
func TestKeptOptionalAssetRefresh(t *testing.T) {
	kept := []string{
		".claude/agents/moai/e2e-tester.md",  // consult pack agent
		".agents/skills/moai-clean/SKILL.md", // commands pack skill
	}
	root := buildMigrationFixture(t)
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatal(err)
	}
	stale := map[string]string{}
	for _, rel := range kept {
		current, err := fs.ReadFile(embedded, rel)
		if err != nil {
			t.Fatal(err)
		}
		// An outdated copy: the bytes an earlier release installed. Its record
		// carries the hash of those bytes, so the copy is not user-modified.
		body := string(current) + "\n<!-- installed by an earlier release -->\n"
		stale[rel] = body
		writeFixtureFile(t, root, rel, body)
		if err := mgr.Track(rel, manifest.TemplateManaged, manifest.HashBytes([]byte(body))); err != nil {
			t.Fatal(err)
		}
	}
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}
	home := installMigrationUserCounterparts(t)

	var lines []string
	report := func(format string, args ...interface{}) {
		lines = append(lines, fmt.Sprintf(format, args...))
	}
	if _, err := migrateProjectCommonAssets(root, home, true, nil, report); err != nil {
		t.Fatal(err)
	}
	for _, rel := range kept {
		// The AC's two arms: refreshed in place, or a notice that names the copy.
		refreshed := readFixtureFile(t, root, rel) != stale[rel]
		named := false
		for _, l := range lines {
			if strings.Contains(l, rel) {
				named = true
			}
		}
		if !refreshed && !named {
			t.Errorf("kept copy %s is neither refreshed nor named in the output; lines: %q", rel, lines)
		}
	}
}
