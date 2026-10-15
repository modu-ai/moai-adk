// migrate_removal_switch_test.go — the project-side removal in
// migrateProjectCommonAssets ships behind a switch. With the default
// configuration a copy whose user counterpart is confirmed stays project-side.
package cli

import (
	"testing"
)

// TestMigrationDefaultKeepsConfirmedProjectCopy is a regression guard, no
// criterion (AC-RC30-02 belongs to card t1668). The fixture's project copy
// meets the migration's removal preconditions (the bytes match the template and
// the user counterpart is confirmed), so the default configuration must keep it.
func TestMigrationDefaultKeepsConfirmedProjectCopy(t *testing.T) {
	root := buildMigrationFixture(t)
	home := installMigrationUserCounterparts(t)
	if _, err := migrateProjectCommonAssets(root, home, true, nil, func(string, ...interface{}) {}); err != nil {
		t.Fatal(err)
	}
	assertFilePresent(t, root, migIdenticalSkill)
}

// enableRemovalForTest turns the removal switch on for one test and restores
// the default when the test ends. A test that exercises the removal mechanics
// calls it; TestMigrationDefaultKeepsConfirmedProjectCopy pins the default.
func enableRemovalForTest(t *testing.T) {
	t.Helper()
	prev := removalEnabled
	removalEnabled = true
	t.Cleanup(func() { removalEnabled = prev })
}
