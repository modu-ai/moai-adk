// migrate_noncatalog_message_test.go — U-09 (AC-RC30-10, REQ-RC30-005): a
// managed skill the catalog does not carry has no installer source, so no
// bundle opts it in. The migration must say so, never name an opt-in command.
package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// TestNonCatalogSkillMessage: a template-managed skill with no catalog entry
// that stays project-side gets a distinct no-catalog message, not the
// "opt in via 'moai bundle add'" message.
func TestNonCatalogSkillMessage(t *testing.T) {
	const retired = ".claude/skills/moai-retired-skill/SKILL.md"
	const body = "# moai-retired-skill: a managed skill the catalog no longer carries\n"
	root := buildMigrationFixture(t)
	writeFixtureFile(t, root, retired, body)
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Track(retired, manifest.TemplateManaged, manifest.HashBytes([]byte(body))); err != nil {
		t.Fatal(err)
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

	var line string
	for _, l := range lines {
		if strings.Contains(l, retired) {
			line = l
		}
	}
	if line == "" {
		t.Fatalf("no migration line names %s; lines: %q", retired, lines)
	}
	if strings.Contains(line, "opt in via") {
		t.Errorf("non-catalog skill gets the opt-in message: %s", line)
	}
	if !strings.Contains(line, "no catalog entry") {
		t.Errorf("non-catalog skill gets no distinct no-catalog message: %s", line)
	}
}
