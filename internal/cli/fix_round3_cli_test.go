// fix_round3_cli_test.go — item 1 (fix round 3) lives in the cli package.
package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// Item 1: a failed read or an empty recorded template hash PRESERVES the
// file — deletion requires successful read + valid hash match + confirmed
// counterpart (fix-round-3 gate repro: user-modified files deleted).
func TestFR3_1_MigrationPreservesOnUnverifiable(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".claude", "skills", "moai-seeded"), 0o755); err != nil {
		t.Fatal(err)
	}
	live := filepath.Join(root, ".claude", "skills", "moai-seeded", "SKILL.md")
	if err := os.WriteFile(live, []byte("user's post-install edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatal(err)
	}
	// TemplateManaged with an EMPTY TemplateHash — unverifiable.
	if err := mgr.Track(".claude/skills/moai-seeded/SKILL.md", manifest.TemplateManaged, ""); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}

	if err := migrateProjectCommonAssets(root, home, nil, func(string, ...interface{}) {}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(live); err != nil {
		t.Errorf("Item 1: unverifiable file deleted by the migration: %v", err)
	}
}
