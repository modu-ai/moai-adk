package cli

// spec_view_duplicate_test.go — `moai spec view` on a duplicate inline AC id
// (card t564).
//
// Before the repair the view returned "parse error" and printed nothing, while
// `moai spec lint` never saw the duplicate at all. The view now warns on stderr
// and renders the tree, which carries the REQ mapping of every line with the id.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpecView_DuplicateACIDWarnsAndRendersTree(t *testing.T) {
	root := t.TempDir()
	specDir := filepath.Join(root, ".moai", "specs", "SPEC-DUPVIEW-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	markdown := `# SPEC-DUPVIEW-001

## Acceptance Criteria

- AC-DUPVIEW-001-01: cited here only as a cross-reference note
- AC-DUPVIEW-001-01: Given a SPEC, When it is viewed, Then the tree renders (maps REQ-DUPVIEW-001-001)
`
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(markdown), 0o644); err != nil {
		t.Fatal(err)
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return root, nil }
	defer func() { findProjectRootFn = origFn }()

	cmd := newSpecViewCmd()
	var out, errOut strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	if err := viewAcceptanceCriteria(cmd, "SPEC-DUPVIEW-001", false); err != nil {
		t.Fatalf("viewAcceptanceCriteria returned %v, want nil (duplicate is a warning)", err)
	}
	if !strings.Contains(errOut.String(), "Warning: duplicate acceptance criteria ID: AC-DUPVIEW-001-01") {
		t.Errorf("stderr = %q, want the duplicate warning", errOut.String())
	}
	for _, want := range []string{"AC-DUPVIEW-001-01", "REQ-DUPVIEW-001-001"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("stdout = %q, want it to contain %s", out.String(), want)
		}
	}
}
