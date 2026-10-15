package cli

// U-16 (SPEC-UPDATE-REPAIR-RC30-001 M6; SPEC-UPDATE-MIGRATION-001 REQ-UPM-030,
// count-by-default): operator ruling d-20261012T221654Z-9ced. The user-visible
// summary reports the preserved user-owned files as a COUNT by default and
// lists each path only under --verbose. The leader applies the same rule to
// the --dry-run reconciliation preview.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

// TestPreviewPreservedLinesCountByDefault seeds two local-only skills (both
// classified user-owned, both preserved) and checks the preview's preserved
// disposition under both flag states.
//
// It is deliberately NOT parallel: it sets the package-level updateVerboseMode,
// which the merge-fallback ledger and the reconciliation preview read. The
// original value is restored in t.Cleanup.
func TestPreviewPreservedLinesCountByDefault(t *testing.T) {
	root := t.TempDir()
	makeSkillDir(t, root, legacySkillIDs[0], "# local-only customization")
	makeSkillDir(t, root, legacySkillIDs[1], "# local-only customization")

	preservedA := ".claude/skills/" + legacySkillIDs[0] + "/SKILL.md"
	preservedB := ".claude/skills/" + legacySkillIDs[1] + "/SKILL.md"

	prev := updateVerboseMode
	t.Cleanup(func() { updateVerboseMode = prev })

	// Default (no --verbose): the preserved files are counted, not listed.
	updateVerboseMode = false
	var dflt bytes.Buffer
	if err := previewReconciliation(root, template.DeployModeLocal, &dflt); err != nil {
		t.Fatalf("previewReconciliation (default): %v", err)
	}
	gotDefault := dflt.String()
	for _, rel := range []string{preservedA, preservedB} {
		if strings.Contains(gotDefault, "[dry-run] preserved: "+rel) {
			t.Errorf("default preview lists the preserved path %s; the count is the default, got:\n%s", rel, gotDefault)
		}
	}
	if !strings.Contains(gotDefault, "[dry-run] preserved: 2 file(s)") {
		t.Errorf("default preview must carry the preserved count line, got:\n%s", gotDefault)
	}
	if !strings.Contains(gotDefault, "run with --verbose to list") {
		t.Errorf("default preview must name the --verbose route to the list, got:\n%s", gotDefault)
	}

	// --verbose: each preserved path is listed, exactly those two, and the
	// count hint is absent because the list is already shown.
	updateVerboseMode = true
	var verbose bytes.Buffer
	if err := previewReconciliation(root, template.DeployModeLocal, &verbose); err != nil {
		t.Fatalf("previewReconciliation (verbose): %v", err)
	}
	gotVerbose := verbose.String()
	for _, rel := range []string{preservedA, preservedB} {
		if !strings.Contains(gotVerbose, "[dry-run] preserved: "+rel) {
			t.Errorf("verbose preview must list the preserved path %s, got:\n%s", rel, gotVerbose)
		}
	}
	if n := strings.Count(gotVerbose, "[dry-run] preserved: "); n != 2 {
		t.Errorf("verbose preview must list exactly the 2 preserved files, listed %d, got:\n%s", n, gotVerbose)
	}
	if strings.Contains(gotVerbose, "run with --verbose to list") {
		t.Errorf("verbose preview must not carry the count hint, got:\n%s", gotVerbose)
	}
}

// withVerboseMode turns the package-level updateVerboseMode on for one test and
// restores the previous value when the test ends. Callers must not call
// t.Parallel: the flag is package state that the preview reads. The --verbose
// twins of the count-by-default tests use it, so each per-path assertion that
// the default now replaces with a count still runs under --verbose.
func withVerboseMode(t *testing.T) {
	t.Helper()
	prev := updateVerboseMode
	t.Cleanup(func() { updateVerboseMode = prev })
	updateVerboseMode = true
}
