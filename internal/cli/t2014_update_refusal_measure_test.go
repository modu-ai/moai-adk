// t2014_update_refusal_measure_test.go — card t2014: measures what the
// user-asset phase of `moai update` does when its protections refuse. The
// tests pin observed behaviour (the refusals stay in force); they weaken
// nothing.
package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/testguard"
	"github.com/modu-ai/moai-adk/internal/userassets"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// seedT2014RefusalFixture builds the operator's state in a temp home: a record
// written under v3.2.0 (a user-edited file recorded as its baseline), a
// manifest key whose root slug carries an `@profile` suffix (written by a
// producer this binary does not know), then switches the binary to rc.38.
func seedT2014RefusalFixture(t *testing.T) (home, target string, edited []byte) {
	t.Helper()
	home = testguard.RedirectHome(t)

	prev := version.Version
	t.Cleanup(func() { version.Version = prev })

	version.Version = "v3.2.0"
	if err := ensureUserAssetsLocked(home, nil, &bytes.Buffer{}); err != nil {
		t.Fatalf("record run under v3.2.0: %v", err)
	}
	const key = "claude-skills/moai-workflow-tdd/SKILL.md"
	target = filepath.Join(home, ".claude", "skills", "moai-workflow-tdd", "SKILL.md")
	edited = []byte("user edits recorded as the baseline\n")
	if err := os.WriteFile(target, edited, 0o644); err != nil {
		t.Fatal(err)
	}
	manifestPath := userassets.ManifestPath(home)
	m, err := userassets.Load(manifestPath)
	if err != nil {
		t.Fatalf("load the recorded manifest: %v", err)
	}
	fe := m.Files[key]
	fe.SHA256 = shaHex(edited)
	m.Files[key] = fe
	m.Files["claude-skills@moai-adk/moai-workflow-tdd/SKILL.md"] = fe
	if err := m.Save(manifestPath); err != nil {
		t.Fatal(err)
	}
	version.Version = "v3.2.0-rc.38"
	return home, target, edited
}

// TestT2014_PhaseReportsRefusalsButReturnsNil reproduces the operator's report
// through the phase function: the refusals are reported as per-file failure
// rows, the refused bytes stay untouched, and the phase returns nil.
func TestT2014_PhaseReportsRefusalsButReturnsNil(t *testing.T) {
	home, target, edited := seedT2014RefusalFixture(t)
	var out bytes.Buffer
	err := runUserAssetUpdatePhase(home, &out)
	t.Logf("phase output:\n%s", out.String())
	if err != nil {
		t.Fatalf("phase returned %v; the measurement expects nil (refusals are only printed)", err)
	}
	got, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("read the refused file back: %v", readErr)
	}
	if string(got) != string(edited) {
		t.Errorf("refused refresh still changed the file: %q", got)
	}
	if !strings.Contains(out.String(), "is newer than this binary") {
		t.Errorf("no downgrade refusal row in the summary")
	}
}

// TestT2014_RefusalsReachTheClosingBlock pins the repair: a run whose
// user-asset phase refused or failed files must say so in the closing
// "Post-update actions" block, because the "Up to date" line that follows
// judges the project templates only.
func TestT2014_RefusalsReachTheClosingBlock(t *testing.T) {
	home, _, _ := seedT2014RefusalFixture(t)
	updateLedger.reset()
	t.Cleanup(updateLedger.reset)
	var out bytes.Buffer
	if err := runUserAssetUpdatePhase(home, &out); err != nil {
		t.Fatalf("phase: %v", err)
	}
	// The closing row must carry the number of failure rows the phase printed
	// above it, not merely mention "user-asset".
	printed := 0
	for _, line := range strings.Split(out.String(), "\n") {
		if strings.HasPrefix(line, "  failure:") {
			printed++
		}
	}
	if printed == 0 {
		t.Fatalf("the fixture produced no failure rows; output:\n%s", out.String())
	}
	want := fmt.Sprintf("%d file(s) refused or failed", printed)
	found := false
	for _, row := range updateLedger.required {
		if strings.Contains(row.msg, want) {
			found = true
		}
	}
	if !found {
		t.Errorf("no closing row carrying %q; required=%+v", want, updateLedger.required)
	}
}

// TestT2014_JournalPersistFailuresAreNotCountedAsRefusedFiles pins audit
// finding F1. When the install journal cannot be saved, install.go adds one
// failure per written file against the journal path (the file HAS been
// written). The closing row must count only real file failures, must not claim
// the journal-failed files were left as they were, and must name the journal
// problem once, in its own row.
func TestT2014_JournalPersistFailuresAreNotCountedAsRefusedFiles(t *testing.T) {
	home := testguard.RedirectHome(t)
	journal := userassets.JournalPath(home)
	res := &userassets.Result{Failures: []userassets.FileOutcome{
		{Path: journal, Reason: "persist completion flag: injected"},
		{Path: journal, Reason: "persist completion flag: injected"},
		{Path: journal, Reason: "persist completion flag: injected"},
		{Path: "claude-skills/moai-workflow-tdd/SKILL.md", Reason: "refresh refused: injected"},
	}}
	updateLedger.reset()
	t.Cleanup(updateLedger.reset)
	recordUserAssetFailures(home, res)

	var fileRow, journalRow string
	for _, row := range updateLedger.required {
		switch {
		case strings.Contains(row.msg, "refused or failed"):
			fileRow = row.msg
		case strings.Contains(row.msg, "journal"):
			journalRow = row.msg
		}
	}
	if !strings.Contains(fileRow, "1 file(s) refused or failed") {
		t.Errorf("JOURNAL_FAILURES_COUNTED_AS_FILES: file row = %q, want a count of 1", fileRow)
	}
	if journalRow == "" {
		t.Errorf("JOURNAL_FAILURE_NOT_NAMED: no row names the journal problem; required=%+v", updateLedger.required)
	}
	if strings.Contains(journalRow, "left as they were") {
		t.Errorf("a journal-persist failure must not claim the files were left as they were: %q", journalRow)
	}
	// Journal-only failures: no file row at all.
	updateLedger.reset()
	recordUserAssetFailures(home, &userassets.Result{Failures: res.Failures[:3]})
	for _, row := range updateLedger.required {
		if strings.Contains(row.msg, "refused or failed") {
			t.Errorf("journal-only failures produced a refused-file row: %q", row.msg)
		}
	}
}

// TestT2014_CleanRunLeavesTheClosingBlockEmpty is the negative control: a run
// with no refusal adds no required row (unconverted references are
// informational).
func TestT2014_CleanRunLeavesTheClosingBlockEmpty(t *testing.T) {
	home := testguard.RedirectHome(t)
	if err := ensureUserAssetsLocked(home, nil, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	updateLedger.reset()
	t.Cleanup(updateLedger.reset)
	if err := runUserAssetUpdatePhase(home, &bytes.Buffer{}); err != nil {
		t.Fatalf("phase: %v", err)
	}
	if len(updateLedger.required) != 0 {
		t.Errorf("clean run recorded required rows: %+v", updateLedger.required)
	}
}
