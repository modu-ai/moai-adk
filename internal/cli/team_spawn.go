package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Error constants for team spawn operations.
const (
	ErrORCWorktreeRequired      = "ORC_WORKTREE_REQUIRED"
	ErrORCUnknownRoleProfile    = "ORC_UNKNOWN_ROLE_PROFILE"
	ErrORCTeamRosterLimit       = "ORC_TEAM_ROSTER_LIMIT"
	ErrORCBroadcastNotPermitted = "ORC_BROADCAST_NOT_PERMITTED"
	ErrORCStaticTeamAgent       = "ORC_STATIC_TEAM_AGENT_PROHIBITED"
	ErrORCReadonlyIsolation     = "ORC_READONLY_ISOLATION_FORBIDDEN"
)

// Mailbox message types.
const (
	MailboxTypeMessage          = "message"
	MailboxTypeShutdownRequest  = "shutdown_request"
	MailboxTypeShutdownResponse = "shutdown_response"
	MailboxTypeBlockerReport    = "blocker_report"
	MailboxTypeTaskHandoff      = "task_handoff"
)

// Write-heavy role profiles that require worktree isolation.
const WriteHeavyRoles = "implementer,tester,designer"

// MailboxMessage represents a typed envelope for SendMessage.
type MailboxMessage struct {
	Type      string         `yaml:"type"`       // message|shutdown_request|shutdown_response|blocker_report|task_handoff
	RequestID string         `yaml:"request_id"` // correlation ID
	Content   string         `yaml:"-"`          // markdown body (after frontmatter)
	Payload   map[string]any `yaml:"payload"`    // optional structured payload
}

// TeamTaskEntry represents a single task in the team task ledger.
type TeamTaskEntry struct {
	TaskID    string `yaml:"task_id"`
	Subject   string `yaml:"subject"`
	Status    string `yaml:"status"`    // pending|claimed|completed|blocked
	ClaimedBy string `yaml:"claimed_by"`
	BlockedBy string `yaml:"blocked_by,omitempty"`
	Timestamp string `yaml:"timestamp"`
}

// RoleProfile represents a validated role profile from workflow.yaml.
type RoleProfile struct {
	Name        string
	Mode        string // plan|acceptEdits
	Model       string // haiku|sonnet|opus
	Isolation   string // none|worktree
	WriteHeavy  bool
	Description string
}

// ValidateRoleProfile checks if a role exists in the provided profiles map.
// Returns ORC_UNKNOWN_ROLE_PROFILE error if role not found (REQ-017).
func ValidateRoleProfile(role string, profiles map[string]RoleProfile) error {
	if _, exists := profiles[role]; !exists {
		return fmt.Errorf("%s: role '%s' not found in workflow.yaml role_profiles", ErrORCUnknownRoleProfile, role)
	}
	return nil
}

// ValidateSpawn performs comprehensive validation for team spawn operations.
// - Validates role exists (REQ-017)
// - For write-heavy roles, requires isolation=="worktree" (REQ-007)
// - Rejects isolation upgrade on read-only roles (REQ-009)
func ValidateSpawn(role string, profiles map[string]RoleProfile, isolation string) error {
	// Check role exists
	profile, exists := profiles[role]
	if !exists {
		return fmt.Errorf("%s: role '%s' not found in workflow.yaml role_profiles", ErrORCUnknownRoleProfile, role)
	}

	// Write-heavy roles require worktree isolation
	if profile.WriteHeavy && isolation != "worktree" {
		return fmt.Errorf("%s: role '%s' is write-heavy and requires 'isolation: worktree' (REQ-007)", ErrORCWorktreeRequired, role)
	}

	// Read-only roles cannot use worktree isolation
	if !profile.WriteHeavy && isolation == "worktree" {
		return fmt.Errorf("%s: role '%s' is read-only and cannot use 'isolation: worktree' (REQ-009)", ErrORCReadonlyIsolation, role)
	}

	return nil
}

// ValidateRoster checks that the team size does not exceed the maximum limit.
// Returns ORC_TEAM_ROSTER_LIMIT error if len(roster) > maxTeammates (REQ-011).
func ValidateRoster(roster []string, maxTeammates int) error {
	if len(roster) > maxTeammates {
		return fmt.Errorf("%s: team size %d exceeds maximum %d (REQ-011)", ErrORCTeamRosterLimit, len(roster), maxTeammates)
	}
	return nil
}

// ValidateMessage checks that message target is not empty.
// Returns ORC_BROADCAST_NOT_PERMITTED if target is empty (REQ-018).
func ValidateMessage(target string) error {
	if target == "" {
		return fmt.Errorf("%s: SendMessage requires a specific target teammate; broadcast is not permitted (REQ-018)", ErrORCBroadcastNotPermitted)
	}
	return nil
}

// NewMailboxMessage creates a new typed mailbox message.
// Defaults type to "message" if empty (REQ-004).
func NewMailboxMessage(msgType, requestID, content string, payload map[string]any) MailboxMessage {
	if msgType == "" {
		msgType = MailboxTypeMessage
	}

	// Validate known message types
	validTypes := map[string]bool{
		MailboxTypeMessage:          true,
		MailboxTypeShutdownRequest:  true,
		MailboxTypeShutdownResponse: true,
		MailboxTypeBlockerReport:    true,
		MailboxTypeTaskHandoff:      true,
	}

	if !validTypes[msgType] {
		// Warn but allow - backward compatibility
		fmt.Fprintf(os.Stderr, "Warning: unknown message type '%s', defaulting to 'message'\n", msgType)
		msgType = MailboxTypeMessage
	}

	return MailboxMessage{
		Type:      msgType,
		RequestID: requestID,
		Content:   content,
		Payload:   payload,
	}
}

// ParseMailboxMessage parses a YAML frontmatter + markdown body from byte slice (REQ-008).
// Defaults type to "message" on unrecognized type with warning.
func ParseMailboxMessage(data []byte) (MailboxMessage, error) {
	// Split frontmatter and body
	parts := bytes.SplitN(data, []byte("---"), 3)
	if len(parts) < 3 {
		return MailboxMessage{}, fmt.Errorf("invalid mailbox message format: missing frontmatter delimiters")
	}

	frontmatterText := parts[1]
	bodyText := parts[2]

	// Parse frontmatter key-value pairs
	frontmatterMap := make(map[string]string)
	lines := strings.Split(string(frontmatterText), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			frontmatterMap[key] = value
		}
	}

	// Extract fields
	msgType := frontmatterMap["type"]
	if msgType == "" {
		msgType = MailboxTypeMessage
	}

	// Validate type
	validTypes := map[string]bool{
		MailboxTypeMessage:          true,
		MailboxTypeShutdownRequest:  true,
		MailboxTypeShutdownResponse: true,
		MailboxTypeBlockerReport:    true,
		MailboxTypeTaskHandoff:      true,
	}

	if !validTypes[msgType] {
		fmt.Fprintf(os.Stderr, "Warning: unknown message type '%s', defaulting to 'message'\n", msgType)
		msgType = MailboxTypeMessage
	}

	// Trim leading newline from bodyText
	content := strings.TrimPrefix(string(bodyText), "\n")
	content = strings.TrimPrefix(content, "\r\n")

	return MailboxMessage{
		Type:      msgType,
		RequestID: frontmatterMap["request_id"],
		Content:   content,
		Payload:   nil, // Payload parsing not needed for MVP
	}, nil
}

// SerializeMailboxMessage serializes a message to YAML frontmatter + markdown body format.
func SerializeMailboxMessage(msg MailboxMessage) []byte {
	var buf bytes.Buffer

	buf.WriteString("---\n")
	fmt.Fprintf(&buf, "type: %s\n", msg.Type)
	if msg.RequestID != "" {
		fmt.Fprintf(&buf, "request_id: %s\n", msg.RequestID)
	}
	if len(msg.Payload) > 0 {
		buf.WriteString("payload:\n")
		for k, v := range msg.Payload {
			fmt.Fprintf(&buf, "  %s: %v\n", k, v)
		}
	}
	buf.WriteString("---\n")
	buf.WriteString(msg.Content)

	return buf.Bytes()
}

// InitTeamState creates the team state directory structure (REQ-006).
// Creates {stateDir}/team/{teamID}/ with:
// - team-config.yaml (snapshot of role profiles)
// - tasklist.md (empty task ledger with header)
// - mailbox/ subdirectory
//
// Note: stateDir should be the full path to .moai/state directory.
func InitTeamState(stateDir, teamID string, profiles map[string]RoleProfile) error {
	teamDir := filepath.Join(stateDir, "team", teamID)

	// Create team directory
	if err := os.MkdirAll(teamDir, 0755); err != nil {
		return fmt.Errorf("create team dir: %w", err)
	}

	// Create mailbox subdirectory
	mailboxDir := filepath.Join(teamDir, "mailbox")
	if err := os.MkdirAll(mailboxDir, 0755); err != nil {
		return fmt.Errorf("create mailbox dir: %w", err)
	}

	// Write team-config.yaml with role profiles snapshot
	configPath := filepath.Join(teamDir, "team-config.yaml")
	var configBuf bytes.Buffer
	configBuf.WriteString("# Team Configuration Snapshot\n")
	configBuf.WriteString("# Generated at: " + time.Now().Format(time.RFC3339) + "\n\n")
	configBuf.WriteString("role_profiles:\n")
	for name, profile := range profiles {
		fmt.Fprintf(&configBuf, "  %s:\n", name)
		fmt.Fprintf(&configBuf, "    mode: %s\n", profile.Mode)
		fmt.Fprintf(&configBuf, "    model: %s\n", profile.Model)
		fmt.Fprintf(&configBuf, "    isolation: %s\n", profile.Isolation)
		fmt.Fprintf(&configBuf, "    write_heavy: %t\n", profile.WriteHeavy)
		fmt.Fprintf(&configBuf, "    description: %s\n", profile.Description)
	}

	if err := os.WriteFile(configPath, configBuf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write team-config.yaml: %w", err)
	}

	// Create empty tasklist.md with header
	tasklistPath := filepath.Join(teamDir, "tasklist.md")
	header := `# Team Task Ledger

This file tracks all tasks assigned to the team. Tasks are append-only; modifications are prohibited (REQ-013).

Format:
- **[TIMESTAMP]** [TASK_ID] - [SUBJECT] - Status: [STATUS] - Claimed by: [CLAIMED_BY] - Blocked by: [BLOCKED_BY]

`
	if err := os.WriteFile(tasklistPath, []byte(header), 0644); err != nil {
		return fmt.Errorf("write tasklist.md: %w", err)
	}

	return nil
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

// ClaimTask atomically claims a task by appending a CLAIMED row (REQ-009).
// Uses filesystem lock (flock) on tasklist.md for concurrency safety.
// Claims the lowest-ID unblocked task if taskID is empty.
func ClaimTask(stateDir, teamID, teammateID, taskID string) error {
	teamDir := filepath.Join(stateDir, "team", teamID)
	tasklistPath := filepath.Join(teamDir, "tasklist.md")

	// Open file with lock
	f, err := os.OpenFile(tasklistPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("open tasklist: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Acquire exclusive lock (platform-specific implementation in team_spawn_lock_*.go)
	if err := lockFile(f); err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	defer func() {
		_ = unlockFile(f)
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

// ArchiveTeamState archives team state to team-archive/ with timestamp (REQ-010).
// Renames {stateDir}/team/{teamID}/ to {stateDir}/team-archive/{teamID}-{timestamp}/
func ArchiveTeamState(stateDir, teamID string) error {
	teamDir := filepath.Join(stateDir, "team", teamID)
	archiveDir := filepath.Join(stateDir, "team-archive")

	// Create archive directory if not exists
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("create archive dir: %w", err)
	}

	// Generate timestamped archive name
	timestamp := time.Now().Format("20060102-150405")
	archiveName := fmt.Sprintf("%s-%s", teamID, timestamp)
	archivePath := filepath.Join(archiveDir, archiveName)

	// Rename to archive
	if err := os.Rename(teamDir, archivePath); err != nil {
		return fmt.Errorf("archive team state: %w", err)
	}

	return nil
}

// LoadRoleProfiles parses workflow.yaml role_profiles section.
// Uses simple YAML parsing (no external deps).
// Determines WriteHeavy based on known write-heavy roles: implementer, tester, designer.
func LoadRoleProfiles(workflowPath string) (map[string]RoleProfile, error) {
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		return nil, fmt.Errorf("read workflow.yaml: %w", err)
	}

	profiles := make(map[string]RoleProfile)

	// Parse role_profiles section
	lines := strings.Split(string(content), "\n")
	inRoleProfiles := false
	var currentProfile *RoleProfile
	var currentName string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "role_profiles:") {
			inRoleProfiles = true
			continue
		}

		if inRoleProfiles {
			// Check for indentation level (2 spaces = profile name, 4 spaces = property)
			indent := len(line) - len(strings.TrimLeft(line, " "))

			if indent == 2 && !strings.HasPrefix(trimmed, "#") && trimmed != "" {
				// New profile name
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					currentName = strings.TrimSpace(parts[0])
					currentProfile = &RoleProfile{
						Name: currentName,
					}
					profiles[currentName] = *currentProfile
				}
			} else if indent == 4 && currentProfile != nil {
				// Profile property
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])

					switch key {
					case "mode":
						currentProfile.Mode = value
						profiles[currentName] = *currentProfile
					case "model":
						currentProfile.Model = value
						profiles[currentName] = *currentProfile
					case "isolation":
						currentProfile.Isolation = value
						profiles[currentName] = *currentProfile
					case "description":
						currentProfile.Description = value
						profiles[currentName] = *currentProfile
					}
				}
			} else if indent == 0 && inRoleProfiles && trimmed != "" {
				// End of role_profiles section
				break
			}
		}
	}

	// Determine WriteHeavy based on role name
	writeHeavySet := make(map[string]bool)
	for _, role := range strings.Split(WriteHeavyRoles, ",") {
		writeHeavySet[strings.TrimSpace(role)] = true
	}

	for name, profile := range profiles {
		if writeHeavySet[name] {
			profile.WriteHeavy = true
			profiles[name] = profile
		}
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no role_profiles found in workflow.yaml")
	}

	return profiles, nil
}

// claimTaskWithRetry wraps ClaimTask with retry logic for concurrent access.
// Uses mutex for in-process synchronization.
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
