// todo_ghost_notice_test.go — SPEC-TODO-SURFACE-POLISH-001 (card t1349)
// AC-TSP-041: the once-per-class ghost artifact notice.
//
// The first read verb that discovers a ghost class says so on stderr,
// exactly once; later reads are silent; the notice marker lands in the
// ALIVE state directory while the ghost directories themselves gain
// nothing; stdout is byte-identical across a noticing and a silent run.
package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

func TestGhostNoticeOnce(t *testing.T) {
	root, homeDB, legacyDB := staleStoreFixture(t)
	// A project-local ghost json beside the divergent store, and a ghost
	// in the PURE ghost directory (the legacy kanban dir) whose entry set
	// must not change.
	localGhost := filepath.Join(root, ".moai", "state", "todo", "backlog.json")
	if err := os.WriteFile(localGhost, []byte(strings.Repeat("g", 40)), 0o600); err != nil {
		t.Fatalf("plant project-local ghost: %v", err)
	}
	pureGhostDir := filepath.Join(root, ".moai", "state", "kanban")
	if err := os.MkdirAll(pureGhostDir, 0o755); err != nil {
		t.Fatalf("create pure ghost dir: %v", err)
	}
	pureGhost := filepath.Join(pureGhostDir, "backlog.json.migrated")
	if err := os.WriteFile(pureGhost, []byte("quarantine"), 0o600); err != nil {
		t.Fatalf("plant pure ghost: %v", err)
	}

	ghostSum := func(path string) string {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		sum := sha256.Sum256(raw)
		return hex.EncodeToString(sum[:])
	}
	localGhostSum := ghostSum(localGhost)
	pureGhostSum := ghostSum(pureGhost)
	homeEntries := dirEntryNames(t, filepath.Dir(homeDB))
	pureGhostEntries := dirEntryNames(t, pureGhostDir)

	out1, errOut1, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("first list: %v (stderr %q)", err, errOut1)
	}
	if !strings.Contains(errOut1, "ghost queue artifact") {
		t.Fatalf("first list stderr = %q, want the ghost notice line", errOut1)
	}
	if !strings.Contains(errOut1, "legacy-json") {
		t.Errorf("first list stderr = %q, want the legacy-json class named", errOut1)
	}

	// Second read: silent on every ghost class the marker now records.
	out2, errOut2, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("second list: %v (stderr %q)", err, errOut2)
	}
	if strings.Contains(errOut2, "ghost queue artifact") {
		t.Errorf("second list stderr = %q, want silence — the notice fires exactly once per class", errOut2)
	}
	if out1 != out2 {
		t.Errorf("stdout changed across a noticing and a silent run:\nfirst:  %q\nsecond: %q — stdout must be disclosure-independent", out1, out2)
	}

	// The marker lives in the alive state directory and records the class.
	marker := filepath.Join(factory.RuntimeStateDirForRoot(root), "ghost-notices.json")
	raw, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("read notice marker %s: %v", marker, err)
	}
	if !strings.Contains(string(raw), "legacy-json") {
		t.Errorf("notice marker = %q, want it to record the legacy-json class", raw)
	}

	// The ghost evidence is untouched: no byte changed, and the pure ghost
	// directory and home queue directory gained no entry (REQ-TSS-003).
	if ghostSum(localGhost) != localGhostSum || ghostSum(pureGhost) != pureGhostSum {
		t.Error("a ghost file changed across the notice runs — the notice path is read-only over the evidence")
	}
	if strings.Join(dirEntryNames(t, filepath.Dir(homeDB)), ",") != strings.Join(homeEntries, ",") {
		t.Errorf("home queue directory changed across the notice runs: %v -> %v", homeEntries, dirEntryNames(t, filepath.Dir(homeDB)))
	}
	if strings.Join(dirEntryNames(t, pureGhostDir), ",") != strings.Join(pureGhostEntries, ",") {
		t.Errorf("pure ghost directory changed across the notice runs: %v -> %v — no file may be added beside the evidence",
			pureGhostEntries, dirEntryNames(t, pureGhostDir))
	}
	_ = legacyDB
}
