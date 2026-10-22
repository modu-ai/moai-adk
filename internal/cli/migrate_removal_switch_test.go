// migrate_removal_switch_test.go — the project-side removal in
// migrateProjectCommonAssets ships behind a switch. With the default
// configuration a copy whose user counterpart is confirmed stays project-side.
package cli

import (
	"fmt"
	"slices"
	"strings"
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

// TestMigrationRemovalGate is AC-RC30-02 (REQ-RC30-003, card t1668): without
// an AC-RC30-01 pass recorded on the tree being updated — the default
// configuration, removalEnabled=false, with no production activation path —
// a project copy that meets EVERY removal precondition (template-identical
// bytes, confirmed user counterpart, install succeeded) stays in place: the
// removal goes to the "removal is not enabled yet" preservation branch.
//
// The criterion is a preservation criterion, so its failure side is observed
// through the mutant recorded in the card evidence (the switch default
// flipped on): this test fails there — the copy is removed — and passes on
// the shipped default. TestMigrationDefaultKeepsConfirmedProjectCopy above
// remains the no-criterion regression guard.
func TestMigrationRemovalGate(t *testing.T) {
	root := buildMigrationFixture(t)
	home := installMigrationUserCounterparts(t)

	var lines []string
	removed, err := migrateProjectCommonAssets(root, home, true, nil, func(format string, args ...interface{}) {
		lines = append(lines, fmt.Sprintf(format, args...))
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 0 {
		t.Errorf("default-configuration migration removed %v; the gate must keep every removal-preconditioned copy", removed)
	}
	assertFilePresent(t, root, migIdenticalSkill)
	if !slices.ContainsFunc(lines, func(l string) bool {
		return strings.Contains(l, "removal is not enabled yet") && strings.Contains(l, migIdenticalSkill)
	}) {
		t.Errorf("no removal-gate preservation line for %s in the migration output: %q", migIdenticalSkill, lines)
	}
}
