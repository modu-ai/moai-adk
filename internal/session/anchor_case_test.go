package session

// anchor_case_test.go — card t1640 M8S, finding F4: the live-session
// containment folds path case where the platform's volumes fold (darwin
// APFS; research §R1 measured 25 of the real repository's trees registered
// under the lowercase spelling), so a session parked through a
// case-aliased cwd anchors its tree (REQ-WHR-016's case-aliased arm,
// AC-WHR-023's containment clause). On a non-folding volume the literal
// comparison stands: two differently-cased paths there are two different
// trees. The volume's behavior is probed independently (os.SameFile of the
// two spellings) — never through the code under test.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathWithinTree_CaseAliasedCwd(t *testing.T) {
	base := t.TempDir()
	upper := filepath.Join(base, "T")
	if err := os.MkdirAll(filepath.Join(upper, "internal", "cli"), 0o755); err != nil {
		t.Fatalf("mkdir upper: %v", err)
	}
	lower := filepath.Join(base, "t")
	child := filepath.Join(lower, "internal", "cli")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("mkdir lower: %v", err)
	}
	// The independent volume probe: the two spellings name one directory.
	ri, rierr := os.Stat(upper)
	li, lierr := os.Stat(lower)
	if rierr != nil || lierr != nil {
		t.Fatalf("stat the alias spellings: %v / %v", rierr, lierr)
	}
	folds := os.SameFile(ri, li)

	got := pathWithinTree(child, upper)
	if folds && !got {
		t.Errorf("pathWithinTree(%q, %q) = false on a case-folding volume; the case-aliased cwd must anchor the tree", child, upper)
	}
	if !folds && got {
		t.Errorf("pathWithinTree(%q, %q) = true on a case-sensitive volume; the literal comparison must stand", child, upper)
	}
}

func TestUnderPath_FoldsCaseOnlyWhereTheVolumeDoes(t *testing.T) {
	base := t.TempDir()
	upper := filepath.Join(base, "MoAI", "t")
	lower := filepath.Join(base, "moai", "t")
	if err := os.MkdirAll(upper, 0o755); err != nil {
		t.Fatalf("mkdir upper: %v", err)
	}
	if err := os.MkdirAll(lower, 0o755); err != nil {
		t.Fatalf("mkdir lower: %v", err)
	}
	ri, rierr := os.Stat(upper)
	li, lierr := os.Stat(lower)
	if rierr != nil || lierr != nil {
		t.Fatalf("stat the alias spellings: %v / %v", rierr, lierr)
	}
	folds := os.SameFile(ri, li)

	got := underPath(filepath.Join(lower, "sub"), upper)
	if folds && !got {
		t.Errorf("underPath does not fold case on a folding volume (moai vs MoAI)")
	}
	if !folds && got {
		t.Errorf("underPath folds case on a non-folding volume")
	}
}
