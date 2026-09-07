package cli

// update_mirror_heal_wiring_test.go — SPEC-UPDATE-MIRROR-HEAL-001, the
// call-site reachability guard for REQ-UMH-001.
//
// Every criterion in update_mirror_heal_test.go drives
// repairSkillMirrorBestEffortAt directly, so none of them would notice the
// call site being dropped from runUpdate. This guard is the analogue of
// TestRunUpdate_CodexRefreshRunsEvenWhenSyncSkips, and its placement half is
// what makes the version-MATCHED update — the whole defect — reach the repair
// at all: after the `if syncSkipped` block, runUpdate has already returned.

import (
	"os"
	"strings"
	"testing"
)

func TestUpdateMirrorHeal_WiredBeforeSyncSkippedReturn(t *testing.T) {
	src, err := os.ReadFile("update.go")
	if err != nil {
		t.Fatalf("read update.go: %v", err)
	}
	body := string(src)

	callIdx := strings.Index(body, "repairSkillMirrorBestEffort(")
	if callIdx < 0 {
		t.Fatal("runUpdate does not call repairSkillMirrorBestEffort — a version-matched update cannot restore a deleted mirror (REQ-UMH-001)")
	}
	skipIdx := strings.Index(body, "if syncSkipped {")
	if skipIdx < 0 {
		t.Fatal("syncSkipped early-return block not found in update.go — test premise stale")
	}
	if callIdx > skipIdx {
		t.Error("repairSkillMirrorBestEffort sits AFTER the syncSkipped early return — a version-matched update never reaches it")
	}
}
