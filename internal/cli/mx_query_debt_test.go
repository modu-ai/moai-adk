package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// TestMxQueryCmd_DebtKindAccepted covers AC-SL-009 at the CLI boundary:
// `moai mx query --kind DEBT` is accepted (no "InvalidQuery" rejection) and
// returns the DEBT tag.
func TestMxQueryCmd_DebtKindAccepted(t *testing.T) {
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".moai", "state")
	buildTestSidecarForCLI(t, stateDir, []mx.Tag{
		makeCLITag(mx.MXDebt, filepath.Join(tmpDir, "cache.go"), "in-memory cache", 3, "no-trigger"),
	})

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	stdout, stderr, err := executeQueryCmd(t, []string{"--kind", "DEBT", "--format", "json"})
	if err != nil {
		t.Fatalf("--kind DEBT should be accepted, got err=%v stderr=%s", err, stderr)
	}
	if strings.Contains(stderr, "InvalidQuery") {
		t.Errorf("--kind DEBT wrongly rejected as InvalidQuery: %s", stderr)
	}
	if !strings.Contains(stdout, `"kind": "DEBT"`) {
		t.Errorf("expected a DEBT tag in JSON output, got: %s", stdout)
	}
}

// TestMxQueryCmd_DebtRotRiskJSON covers AC-SL-008 / Scenario 3 at the exact
// CLI-output contract: `moai mx query --kind DEBT --json` emits
// "rotRisk": "no-trigger" for a no-@MX:UPGRADE marker, and a marker with an
// upgrade trigger carries no "no-trigger" token (RotRisk omitted under omitempty).
//
// Binary PASS contract (acceptance.md AC-SL-008):
//
//	moai mx query --kind DEBT --json | grep -c '"rotRisk": "no-trigger"' == 1
func TestMxQueryCmd_DebtRotRiskJSON(t *testing.T) {
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".moai", "state")
	buildTestSidecarForCLI(t, stateDir, []mx.Tag{
		// rotting DEBT: no upgrade trigger.
		makeCLITag(mx.MXDebt, filepath.Join(tmpDir, "rot.go"), "hardcoded retry count", 3, "no-trigger"),
		// healthy DEBT: upgrade trigger present (empty RotRisk -> omitted in JSON).
		makeCLITag(mx.MXDebt, filepath.Join(tmpDir, "ok.go"), "in-memory cache", 4, ""),
	})

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	stdout, stderr, err := executeQueryCmd(t, []string{"--kind", "DEBT", "--format", "json"})
	if err != nil {
		t.Fatalf("query failed: err=%v stderr=%s", err, stderr)
	}

	// Exactly one '"rotRisk": "no-trigger"' occurrence (the rotting DEBT only).
	count := strings.Count(stdout, `"rotRisk": "no-trigger"`)
	if count != 1 {
		t.Errorf("expected exactly 1 '\"rotRisk\": \"no-trigger\"' token, got %d.\nJSON:\n%s", count, stdout)
	}
}

// makeCLITag builds a Tag for CLI-level sidecar fixtures, including RotRisk.
func makeCLITag(kind mx.TagKind, file, body string, line int, rotRisk string) mx.Tag {
	return mx.Tag{
		Kind:      kind,
		File:      file,
		Line:      line,
		Body:      body,
		RotRisk:   rotRisk,
		CreatedBy: "test",
	}
}
