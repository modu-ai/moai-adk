package backup

// retention_probe_p3_test.go — AC-UBR-03 (probe P3, SPEC-UPDATE-BACKUP-RETENTION-001):
// redundant copies are not all kept. Covers REQ-UBR-005.

import (
	"path/filepath"
	"testing"
)

func TestAC_UBR_03_RedundantCopiesAreNotAllKept(t *testing.T) {
	root := t.TempDir()
	// Ten runs that each hold v1 of one file, then a newer run that holds v2.
	names := make([]string, 0, 11)
	for i := 1; i <= 10; i++ {
		names = append(names, t1733NamespaceRun(t, root, i, "v1"))
	}
	names = append(names, t1733NamespaceRun(t, root, 11, "v2"))

	deleted := CleanupOldBackups(root, 1)
	rem := t1733DirNames(filepath.Join(root, ".moai", "backups"))
	t.Logf("AC-UBR-03 deleted=%d remaining=%v", deleted, rem)

	// Exactly one run holding v1 remains, and it is the newest of the ten.
	if !t1733Contains(rem, names[9]) {
		t.Errorf("AC-UBR-03: the newest v1 holder (%s) was removed; want it retained as the sole v1 copy", names[9])
	}
	v1Holders := 0
	for _, n := range names[:10] {
		if t1733Contains(rem, n) {
			v1Holders++
		}
	}
	if v1Holders != 1 {
		t.Errorf("AC-UBR-03: %d runs holding v1 remain, want exactly 1", v1Holders)
	}
	// The run holding v2 is the window.
	if !t1733Contains(rem, names[10]) {
		t.Errorf("AC-UBR-03: the run holding v2 (%s) was removed; want it kept as the window", names[10])
	}
	if deleted != 9 {
		t.Errorf("AC-UBR-03: deleted=%d, want 9", deleted)
	}
}
