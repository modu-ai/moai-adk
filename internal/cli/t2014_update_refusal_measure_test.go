// t2014_update_refusal_measure_test.go — card t2014: measures what the
// user-asset phase of `moai update` does when its protections refuse. The
// tests pin observed behaviour (the refusals stay in force); they weaken
// nothing.
package cli

import (
	"bytes"
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
	m, _ := userassets.Load(manifestPath)
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
	got, _ := os.ReadFile(target)
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
	if err := runUserAssetUpdatePhase(home, &bytes.Buffer{}); err != nil {
		t.Fatalf("phase: %v", err)
	}
	found := false
	for _, row := range updateLedger.required {
		if strings.Contains(row.msg, "user-asset") {
			found = true
		}
	}
	if !found {
		t.Errorf("no user-asset row in the closing block; required=%+v", updateLedger.required)
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
