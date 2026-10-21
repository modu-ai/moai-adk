package backup

// retention_probe_p2_test.go — AC-UBR-02 (probe P2, SPEC-UPDATE-BACKUP-RETENTION-001):
// the current version does not take a budget slot. Covers REQ-UBR-003 and
// REQ-UBR-004.

import (
	"path/filepath"
	"testing"
)

func TestAC_UBR_02_CurrentVersionDoesNotTakeABudgetSlot(t *testing.T) {
	root := t.TempDir()
	// One file backed up in five runs as v1, v2, v3, v4, v4 (oldest first).
	names := make([]string, 0, 5)
	for i, v := range []string{"v1", "v2", "v3", "v4", "v4"} {
		names = append(names, t1733NamespaceRun(t, root, i+1, v))
	}

	deleted := CleanupOldBackups(root, 1)
	rem := t1733DirNames(filepath.Join(root, ".moai", "backups"))
	t.Logf("AC-UBR-02 deleted=%d remaining=%v", deleted, rem)

	// The runs holding v1, v2 and v3 remain: each is the only holder of a past
	// version inside the budget of 3.
	for i, v := range []string{"v1", "v2", "v3"} {
		if !t1733Contains(rem, names[i]) {
			t.Errorf("AC-UBR-02: the run holding %s (%s) was removed; want it retained (budget 3, window 1)", v, names[i])
		}
	}
	// The older run holding v4 is redundant with the window run that holds v4.
	if t1733Contains(rem, names[3]) {
		t.Errorf("AC-UBR-02: the older run holding v4 (%s) remains; want it removed as redundant with the window", names[3])
	}
	// The newest run (v4) is the window.
	if !t1733Contains(rem, names[4]) {
		t.Errorf("AC-UBR-02: the newest run (%s) was removed; want it kept as the window", names[4])
	}
	if deleted != 1 {
		t.Errorf("AC-UBR-02: deleted=%d, want 1 (only the older v4 run)", deleted)
	}
}
