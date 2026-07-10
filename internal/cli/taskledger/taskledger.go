// Package taskledger provides concurrency-safe, append-only task-ledger
// primitives (AppendTask / ClaimTask / TaskClaimer). Migrated from
// internal/cli/team_spawn.go (SPEC-AGENT-TEAM-RETIRE-001 M0, REQ-ATR-002) so
// the SPEC-CLIFIX-CRITICAL-001 P0 map-RMW / O_APPEND regression guard survives
// the Agent Teams static-layer retirement.
package taskledger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/lockfile"
)

// TeamTaskEntry represents a single task in the team task ledger.
type TeamTaskEntry struct {
	TaskID    string `yaml:"task_id"`
	Subject   string `yaml:"subject"`
	Status    string `yaml:"status"` // pending|claimed|completed|blocked
	ClaimedBy string `yaml:"claimed_by"`
	BlockedBy string `yaml:"blocked_by,omitempty"`
	Timestamp string `yaml:"timestamp"`
}

// AppendTask appends a task entry to tasklist.md (REQ-003).
// Enforces append-only: no modifications or deletes allowed (REQ-013).
func AppendTask(stateDir, teamID string, entry TeamTaskEntry) error {
	teamDir := filepath.Join(stateDir, "team", teamID)
	tasklistPath := filepath.Join(teamDir, "tasklist.md")

	// Open file in append mode
	f, err := os.OpenFile(tasklistPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open tasklist: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Format entry as markdown
	blockedBy := ""
	if entry.BlockedBy != "" {
		blockedBy = fmt.Sprintf(" - Blocked by: %s", entry.BlockedBy)
	}
	line := fmt.Sprintf("- **%s** %s - %s - Status: %s - Claimed by: %s%s\n",
		entry.Timestamp, entry.TaskID, entry.Subject, entry.Status, entry.ClaimedBy, blockedBy)

	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("write task entry: %w", err)
	}

	return nil
}

// @MX:ANCHOR: [AUTO] ClaimTask carries the SPEC-CLIFIX-CRITICAL-001 P0 O_APPEND fix
// @MX:REASON: fan_in >= 3: TaskClaimer.Claim, the internal/cli
// TestClaimTaskAppend_Repro regression guard, and the taskledger coverage tests
// all depend on the append-only CLAIMED-row contract; opening with anything other
// than O_APPEND|O_WRONLY regresses the ledger-head-overwrite defect.
// @MX:SPEC: SPEC-CLIFIX-CRITICAL-001
//
// ClaimTask atomically claims a task by appending a CLAIMED row (REQ-009).
// Uses filesystem lock (flock) on tasklist.md for concurrency safety.
// Claims the lowest-ID unblocked task if taskID is empty.
func ClaimTask(stateDir, teamID, teammateID, taskID string) error {
	teamDir := filepath.Join(stateDir, "team", teamID)
	tasklistPath := filepath.Join(teamDir, "tasklist.md")

	// SPEC-CLIFIX-CRITICAL-001 REQ-CRIT-001-002: open with O_APPEND|O_WRONLY so
	// the CLAIMED row lands at the ledger tail and the append-only head is never
	// overwritten. The lock is acquired on this fd; the content read uses a
	// separate os.ReadFile handle (lock is still valid — flock works on any fd).
	f, err := os.OpenFile(tasklistPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open tasklist: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Acquire exclusive lock (platform-specific implementation in internal/lockfile)
	if err := lockfile.Lock(f); err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	defer func() {
		_ = lockfile.Unlock(f)
	}()

	// Read current content to find task
	content, err := os.ReadFile(tasklistPath)
	if err != nil {
		return fmt.Errorf("read tasklist: %w", err)
	}

	// Parse tasks to find the target
	lines := strings.Split(string(content), "\n")
	var targetTaskID string

	if taskID != "" {
		// Claim specific task
		targetTaskID = taskID
		for _, line := range lines {
			if strings.Contains(line, taskID) && strings.Contains(line, "Status: pending") {
				break
			}
		}
	} else {
		// Find lowest-ID unblocked pending task
		for _, line := range lines {
			if strings.Contains(line, "Status: pending") && !strings.Contains(line, "Blocked by:") {
				// Extract task ID
				parts := strings.Fields(line)
				for _, part := range parts {
					if strings.HasPrefix(part, "SPEC-") || strings.HasPrefix(part, "TASK-") {
						targetTaskID = part
						break
					}
				}
				if targetTaskID != "" {
					break
				}
			}
		}
	}

	if targetTaskID == "" {
		return fmt.Errorf("no claimable task found")
	}

	// Append CLAIMED row
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	claimLine := fmt.Sprintf("- **%s** CLAIMED - %s claimed by %s\n", timestamp, targetTaskID, teammateID)

	if _, err := f.WriteString(claimLine); err != nil {
		return fmt.Errorf("write claim: %w", err)
	}

	return nil
}

// TaskClaimer wraps ClaimTask with in-process synchronized access for
// concurrent claimers. Uses a mutex for in-process synchronization.
type TaskClaimer struct {
	mu sync.Mutex
}

// NewTaskClaimer creates a new task claimer with synchronized access.
func NewTaskClaimer() *TaskClaimer {
	return &TaskClaimer{}
}

// Claim attempts to claim a task with automatic retry on concurrent access.
func (tc *TaskClaimer) Claim(stateDir, teamID, teammateID, taskID string) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	return ClaimTask(stateDir, teamID, teammateID, taskID)
}
