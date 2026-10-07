// route_b_precheck_test.go — the manager-git bundle precondition at ALL
// Route B entry points (SPEC-USER-ASSET-INSTALL-001 final-class item 4 +
// directed repair R-b; AC-018's Route-B arm; C4's actionable-report rule).
//
// Route B (Tier L OR explicit --pr) delegates git work to manager-git, who
// is NOT one of the L0 five (D-Q1) — the role body ships in the opt-in
// `delivery` bundle (D-Q5). Both entry points — the sync delivery Route B
// row AND the run flow's Route B at task-decomposition (Phase 19) — must
// verify the role body is installed user-side and, when it is not, refuse
// with the named `moai bundle add delivery` remediation: never a
// missing-file error mid-flow.
package template

import (
	"io/fs"
	"strings"
	"testing"
)

func fsReadAll(fsys fs.FS, path string) ([]byte, error) {
	return fs.ReadFile(fsys, path)
}

// TestRouteBPrecheckCoversAllEntryPoints pins the precheck + remediation on
// both manager-git entry points.
func TestRouteBPrecheckCoversAllEntryPoints(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	cases := []struct {
		path string
	}{
		{".claude/skills/moai/workflows/sync/delivery.md"},
		{".claude/skills/moai/workflows/run/task-decomposition.md"},
	}
	for _, tc := range cases {
		data, readErr := fsReadAll(fsys, tc.path)
		if readErr != nil {
			t.Fatalf("read %s: %v", tc.path, readErr)
		}
		doc := string(data)
		if !strings.Contains(doc, "moai bundle add delivery") {
			t.Errorf("ROUTEB_PRECHECK_MISSING: %s does not carry the `moai bundle add delivery` remediation (R-b: the refusal must name the recovery command)", tc.path)
		}
		if !strings.Contains(doc, "Route B") {
			t.Errorf("ROUTEB_PRECHECK_MISSING: %s carries no Route B row to hang the precheck on", tc.path)
		}
	}
}
