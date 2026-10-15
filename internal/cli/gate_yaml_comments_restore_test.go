package cli

// gate_yaml_comments_restore_test.go — the gate.yaml comment-survival contract
// for the restore step (SPEC-UPDATE-REPAIR-RC30-001 AC-RC30-22, U-23;
// REQ-RC30-007). The user's comment lines in a sections config file must
// survive the update cycle's backup, wipe, deploy, and restore steps.
//
// The fixture is the REAL embedded template gate.yaml with twelve user comment
// lines above a template-carried key, the way a user annotates the file. The
// driver is the production template-sync cycle (runTemplateSyncAt): the backup
// keeps every line, the deploy step writes the template file over the user's,
// and the restore merge (backup.RestoreMoaiConfigRetained, called from the
// template-sync path) must carry the lines back. This test is the RED that
// precedes the restore-merge fix in internal/cli/update/backup/node_merge.go.

import (
	"fmt"
	"strings"
	"testing"
)

// gateNoteCount is the number of user comment lines in the fixture (U-23).
const gateNoteCount = 12

// gateNoteMarker names user comment line i. The assertion looks for the marker
// token, not the whole line, so a line the merge re-indents still counts as
// present, and a missing marker names exactly the line that was lost.
func gateNoteMarker(i int) string {
	return fmt.Sprintf("RC30-U23-NOTE-%02d", i)
}

// buildGateNotesUserYAML derives the user's gate.yaml from the embedded
// template: the twelve comment lines sit above skip_tests, and the user has
// switched skip_tests on. The needle must appear exactly once in the template
// (replaceOnce), so template drift fails the builder instead of yielding a
// vacuous green.
func buildGateNotesUserYAML(t *testing.T, base []byte) []byte {
	t.Helper()
	var notes strings.Builder
	for i := 1; i <= gateNoteCount; i++ {
		notes.WriteString("  # " + gateNoteMarker(i) + ": user gate note, kept across the update cycle\n")
	}
	s := replaceOnce(t, string(base), "  skip_tests: false\n", notes.String()+"  skip_tests: true\n")
	return []byte(s)
}

// TestGateYAMLCommentsSurviveRestore is AC-RC30-22: after the production
// template-sync cycle (backup, wipe, deploy, restore), all twelve user comment
// lines are still in gate.yaml. The failure names each line that was lost.
func TestGateYAMLCommentsSurviveRestore(t *testing.T) {
	fixture := buildGateNotesUserYAML(t, readEmbeddedGateYAML(t))
	if got := strings.Count(string(fixture), "RC30-U23-NOTE-"); got != gateNoteCount {
		t.Fatalf("fixture premise broken: %d user comment lines in the user gate.yaml, want %d", got, gateNoteCount)
	}
	root := makeGatePreserveFixture(t, fixture)

	runTemplateSyncAt(t, root)

	raw := gateYAMLRaw(t, root)
	var lost []string
	for i := 1; i <= gateNoteCount; i++ {
		if !strings.Contains(raw, gateNoteMarker(i)) {
			lost = append(lost, gateNoteMarker(i))
		}
	}
	if len(lost) > 0 {
		t.Errorf("%d of %d user comment lines lost across backup/wipe/deploy/restore; lost: %s\npost-update gate.yaml:\n%s",
			len(lost), gateNoteCount, strings.Join(lost, ", "), raw)
	}
}
