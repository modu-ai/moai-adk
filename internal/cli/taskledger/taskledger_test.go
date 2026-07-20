package taskledger

// Migrated coverage tests for the task-ledger primitives
// (SPEC-AGENT-TEAM-RETIRE-001 M0, REQ-ATR-002). TestClaimTask and
// TestClaimTaskConcurrent moved from internal/cli/team_spawn_test.go and
// re-pointed at the taskledger package symbols; the setup replaces the retired
// InitTeamState helper with a direct tasklist.md fixture write (same on-disk
// shape: header + append-only entries).

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// initTasklist creates the team dir and a tasklist.md with the canonical header,
// mirroring the tasklist portion of the retired InitTeamState.
func initTasklist(t *testing.T, stateDir, teamID string) string {
	t.Helper()
	teamDir := filepath.Join(stateDir, "team", teamID)
	if err := os.MkdirAll(teamDir, 0755); err != nil {
		t.Fatalf("mkdir team dir: %v", err)
	}
	header := `# Team Task Ledger

This file tracks all tasks assigned to the team. Tasks are append-only; modifications are prohibited (REQ-013).

Format:
- **[TIMESTAMP]** [TASK_ID] - [SUBJECT] - Status: [STATUS] - Claimed by: [CLAIMED_BY] - Blocked by: [BLOCKED_BY]

`
	tasklistPath := filepath.Join(teamDir, "tasklist.md")
	if err := os.WriteFile(tasklistPath, []byte(header), 0644); err != nil {
		t.Fatalf("write tasklist.md: %v", err)
	}
	return tasklistPath
}

// TestClaimTask tests atomic task claiming with filesystem lock.
func TestClaimTask(t *testing.T) {
	tempDir := t.TempDir()
	stateDir := filepath.Join(tempDir, ".moai", "state")
	teamID := "test-team-789"

	tasklistPath := initTasklist(t, stateDir, teamID)

	// Append a pending task
	entry := TeamTaskEntry{
		TaskID:    "SPEC-002",
		Subject:   "Implement authorization",
		Status:    "pending",
		ClaimedBy: "",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := AppendTask(stateDir, teamID, entry); err != nil {
		t.Fatalf("AppendTask() error: %v", err)
	}

	// Claim the task
	err := ClaimTask(stateDir, teamID, "teammate-1", "SPEC-002")
	if err != nil {
		t.Fatalf("ClaimTask() error: %v", err)
	}

	// Verify claim was appended
	content, err := os.ReadFile(tasklistPath)
	if err != nil {
		t.Fatalf("read tasklist.md: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "CLAIMED") {
		t.Errorf("claim entry not found in tasklist.md")
	}
	if !strings.Contains(contentStr, "teammate-1") {
		t.Errorf("teammate ID not found in claim entry")
	}
	if !strings.Contains(contentStr, "SPEC-002") {
		t.Errorf("task ID not found in claim entry")
	}

	// Test concurrent claims result in distinct task IDs
	// (This is a simplified test; real concurrency requires goroutines)
	t.Log("AC-05: Two concurrent claims result in distinct task IDs (simplified test)")
}

// TestClaimTaskConcurrent tests concurrent task claiming.
func TestClaimTaskConcurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent test in short mode")
	}

	tempDir := t.TempDir()
	stateDir := filepath.Join(tempDir, ".moai", "state")
	teamID := "test-team-concurrent"

	initTasklist(t, stateDir, teamID)

	// Append multiple pending tasks
	tasks := []string{"SPEC-001", "SPEC-002", "SPEC-003"}
	for _, taskID := range tasks {
		entry := TeamTaskEntry{
			TaskID:    taskID,
			Subject:   fmt.Sprintf("Task %s", taskID),
			Status:    "pending",
			ClaimedBy: "",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
		if err := AppendTask(stateDir, teamID, entry); err != nil {
			t.Fatalf("AppendTask() error: %v", err)
		}
	}

	// Attempt concurrent claims using TaskClaimer
	claimer := NewTaskClaimer()
	var wg sync.WaitGroup
	claimCount := make(map[string]int)
	var mu sync.Mutex

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			teammateID := fmt.Sprintf("teammate-%d", idx)
			err := claimer.Claim(stateDir, teamID, teammateID, "")
			if err != nil {
				t.Logf("Claim failed for %s: %v", teammateID, err)
				return
			}

			mu.Lock()
			claimCount[teammateID]++
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify distinct claims
	if len(claimCount) == 0 {
		t.Error("no successful claims")
	}

	t.Log("AC-05: Concurrent claims processed")
}

// TestAppendTask tests task entry appending with blocked-by rendering.
func TestAppendTask(t *testing.T) {
	tempDir := t.TempDir()
	stateDir := filepath.Join(tempDir, ".moai", "state")
	teamID := "test-team-append"

	tasklistPath := initTasklist(t, stateDir, teamID)

	entry := TeamTaskEntry{
		TaskID:    "SPEC-010",
		Subject:   "Implement authentication",
		Status:    "pending",
		ClaimedBy: "",
		BlockedBy: "SPEC-009",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := AppendTask(stateDir, teamID, entry); err != nil {
		t.Fatalf("AppendTask() error: %v", err)
	}

	content, err := os.ReadFile(tasklistPath)
	if err != nil {
		t.Fatalf("read tasklist.md: %v", err)
	}
	contentStr := string(content)
	for _, want := range []string{"SPEC-010", "Implement authentication", "Status: pending", "Blocked by: SPEC-009"} {
		if !strings.Contains(contentStr, want) {
			t.Errorf("tasklist.md missing %q", want)
		}
	}
	// Header must survive (append-only).
	if !strings.HasPrefix(contentStr, "# Team Task Ledger") {
		t.Error("tasklist.md header was overwritten (append-only violated)")
	}
}

// TestClaimTaskNoClaimable verifies the no-claimable-task error path when the
// ledger has no pending unblocked entries and no explicit taskID is given.
func TestClaimTaskNoClaimable(t *testing.T) {
	tempDir := t.TempDir()
	stateDir := filepath.Join(tempDir, ".moai", "state")
	teamID := "test-team-empty"

	initTasklist(t, stateDir, teamID)

	err := ClaimTask(stateDir, teamID, "teammate-1", "")
	if err == nil {
		t.Fatal("ClaimTask() expected error for empty ledger, got nil")
	}
	if !strings.Contains(err.Error(), "no claimable task found") {
		t.Errorf("error = %v, want 'no claimable task found'", err)
	}
}

// TestClaimTaskMissingLedger verifies the open-error path when tasklist.md
// does not exist.
func TestClaimTaskMissingLedger(t *testing.T) {
	tempDir := t.TempDir()

	err := ClaimTask(tempDir, "no-such-team", "teammate-1", "TASK-001")
	if err == nil {
		t.Fatal("ClaimTask() expected error for missing tasklist, got nil")
	}
	if !strings.Contains(err.Error(), "open tasklist") {
		t.Errorf("error = %v, want 'open tasklist' wrap", err)
	}
}

// TestAppendTaskMissingLedger verifies the open-error path for AppendTask.
func TestAppendTaskMissingLedger(t *testing.T) {
	tempDir := t.TempDir()

	err := AppendTask(tempDir, "no-such-team", TeamTaskEntry{TaskID: "TASK-001"})
	if err == nil {
		t.Fatal("AppendTask() expected error for missing tasklist, got nil")
	}
	if !strings.Contains(err.Error(), "open tasklist") {
		t.Errorf("error = %v, want 'open tasklist' wrap", err)
	}
}
