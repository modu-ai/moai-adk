package cli

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/harness"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/hook/dbsync"
)

var hookCmd = &cobra.Command{
	Use:     "hook",
	Short:   "Execute hook event handlers",
	GroupID: "tools",
	Long:    "Execute Claude Code hook event handlers. Called by Claude Code settings.json hook configuration.",
}

func init() {
	rootCmd.AddCommand(hookCmd)

	// Register all hook subcommands
	hookSubcommands := []struct {
		use   string
		short string
		event hook.EventType
	}{
		{"session-start", "Handle session start event", hook.EventSessionStart},
		{"pre-tool", "Handle pre-tool-use event", hook.EventPreToolUse},
		{"post-tool", "Handle post-tool-use event", hook.EventPostToolUse},
		{"session-end", "Handle session end event", hook.EventSessionEnd},
		{"stop", "Handle stop event", hook.EventStop},
		{"compact", "Handle pre-compact event", hook.EventPreCompact},
		{"post-tool-failure", "Handle post-tool-use failure event", hook.EventPostToolUseFailure},
		{"notification", "Handle notification event", hook.EventNotification},
		{"subagent-start", "Handle subagent start event", hook.EventSubagentStart},
		{"user-prompt-submit", "Handle user prompt submit event", hook.EventUserPromptSubmit},
		{"permission-request", "Handle permission request event", hook.EventPermissionRequest},
		{"teammate-idle", "Handle teammate idle event", hook.EventTeammateIdle},
		{"task-completed", "Handle task completed event", hook.EventTaskCompleted},
		{"subagent-stop", "Handle subagent stop event", hook.EventSubagentStop},
		{"worktree-create", "Handle worktree create event", hook.EventWorktreeCreate},
		{"worktree-remove", "Handle worktree remove event", hook.EventWorktreeRemove},
		{"post-compact", "Handle post-compact event", hook.EventPostCompact},
		{"instructions-loaded", "Handle instructions loaded event", hook.EventInstructionsLoaded},
		{"stop-failure", "Handle stop failure event", hook.EventStopFailure},
		{"config-change", "Handle config change event", hook.EventConfigChange},
		{"task-created", "Handle task created event", hook.EventTaskCreated},
		{"cwd-changed", "Handle cwd changed event", hook.EventCwdChanged},
		{"file-changed", "Handle file changed event", hook.EventFileChanged},
		{"elicitation", "Handle MCP elicitation event", hook.EventElicitation},
		{"elicitation-result", "Handle MCP elicitation result event", hook.EventElicitationResult},
		{"permission-denied", "Handle permission denied event", hook.EventPermissionDenied},
	}

	for _, sub := range hookSubcommands {
		event := sub.event // capture for closure
		cmd := &cobra.Command{
			Use:   sub.use,
			Short: sub.short,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return runHookEvent(cmd, event)
			},
		}
		hookCmd.AddCommand(cmd)
	}

	// Add "list" subcommand
	hookCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List registered hook handlers",
		RunE:  runHookList,
	})

	// Add "agent" subcommand for agent-specific hooks
	hookCmd.AddCommand(&cobra.Command{
		Use:   "agent [action]",
		Short: "Execute agent-specific hook action",
		Long:  "Execute agent-specific hook actions like ddd-pre-transformation, backend-validation, etc.",
		Args:  cobra.ExactArgs(1),
		RunE:  runAgentHook,
	})

	// harness-observe Add subcommands (SPEC-V3R3-HARNESS-LEARNING-001 T-P1-03)
	// Read PostToolUse hook stdin JSON and record event to usage-log.jsonl.
	hookCmd.AddCommand(&cobra.Command{
		Use:   "harness-observe",
		Short: "Record PostToolUse event to harness usage log",
		Long:  "Reads hook stdin JSON and appends an event to .moai/harness/usage-log.jsonl. Called from handle-harness-observe.sh.",
		RunE:  runHarnessObserve,
	})

	// Multi-event observer subcommands (SPEC-V3R4-HARNESS-002 Wave A)
	hookCmd.AddCommand(&cobra.Command{
		Use:   "harness-observe-stop",
		Short: "Record Stop event to harness usage log",
		Long:  "Reads Stop hook stdin JSON and appends a session_stop event to .moai/harness/usage-log.jsonl.",
		RunE:  runHarnessObserveStop,
	})
	hookCmd.AddCommand(&cobra.Command{
		Use:   "harness-observe-subagent-stop",
		Short: "Record SubagentStop event to harness usage log",
		Long:  "Reads SubagentStop hook stdin JSON and appends a subagent_stop event to .moai/harness/usage-log.jsonl.",
		RunE:  runHarnessObserveSubagentStop,
	})
	hookCmd.AddCommand(&cobra.Command{
		Use:   "harness-observe-user-prompt-submit",
		Short: "Record UserPromptSubmit event to harness usage log",
		Long:  "Reads UserPromptSubmit hook stdin JSON and appends a user_prompt event. Default strategy: SHA-256 hash + length.",
		RunE:  runHarnessObserveUserPromptSubmit,
	})

	// Add "db-schema-sync" subcommand (SPEC-DB-SYNC-001)
	dbSchemaSyncCmd := &cobra.Command{
		Use:   "db-schema-sync",
		Short: "Handle DB schema change detection from PostToolUse hook",
		Long:  "Detect migration file changes, apply debounce, parse, and write proposal.json for user approval.",
		RunE:  runDBSchemaSync,
	}
	dbSchemaSyncCmd.Flags().String("file", "", "File path from PostToolUse hook stdin")
	hookCmd.AddCommand(dbSchemaSyncCmd)

	// Add "spec-status" subcommand (SPEC-STATUS-AUTO-001)
	specStatusCmd := &cobra.Command{
		Use:   "spec-status",
		Short: "Auto-update SPEC status on git commit",
		Long:  "Extract SPEC-IDs from git commit messages and update their status to 'implemented'. Called from handle-spec-status.sh.",
		RunE:  runSpecStatus,
	}
	hookCmd.AddCommand(specStatusCmd)
}

// @MX:ANCHOR: [AUTO] runHookEvent is the central dispatcher for all Claude Code hook events
// @MX:REASON: [AUTO] fan_in=3, called from hook.go init(), coverage_test.go, hook_e2e_test.go
// runHookEvent dispatches a hook event by reading JSON from stdin and writing to stdout.
func runHookEvent(cmd *cobra.Command, event hook.EventType) error {
	if deps == nil || deps.HookProtocol == nil || deps.HookRegistry == nil {
		return fmt.Errorf("hook system not initialized")
	}

	input, err := deps.HookProtocol.ReadInput(os.Stdin)
	if err != nil {
		return fmt.Errorf("read hook input: %w", err)
	}

	// Inject event name from CLI subcommand when Claude Code omits it.
	if input.HookEventName == "" || input.HookEventName == "unknown" {
		input.HookEventName = string(event)
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()

	output, err := deps.HookRegistry.Dispatch(ctx, event, input)
	if err != nil {
		return fmt.Errorf("dispatch hook: %w", err)
	}

	if err := writeHookOutput(event, input, output); err != nil {
		return err
	}

	// Exit code 2 for explicit exit code (TeammateIdle, TaskCompleted)
	if output != nil && output.ExitCode == 2 {
		os.Exit(2)
	}

	// Exit code 2 for deny decisions per Claude Code protocol
	if output != nil && output.Decision == hook.DecisionDeny {
		os.Exit(2)
	}

	return nil
}

// writeHookOutput dispatches stdout writing per event-specific contract.
//
// WorktreeCreate / WorktreeRemove (Claude Code v2.1.49+): the runtime parses
// stdout as the worktree directory path (not JSON). The hook MUST echo the
// directory path as plain text — emitting an empty JSON object yields
// "WorktreeCreate hook returned a path that is not a directory: {}". We echo
// input.WorktreePath unchanged, treating the hook as a passthrough observer.
// When input.WorktreePath is absent the stdout is left empty (fail-safe).
//
// All other events use the JSON HookOutput protocol via HookProtocol.WriteOutput.
func writeHookOutput(event hook.EventType, input *hook.HookInput, output *hook.HookOutput) error {
	if event == hook.EventWorktreeCreate || event == hook.EventWorktreeRemove {
		if input != nil && input.WorktreePath != "" {
			if _, writeErr := fmt.Fprintln(os.Stdout, input.WorktreePath); writeErr != nil {
				return fmt.Errorf("write worktree path: %w", writeErr)
			}
		}
		return nil
	}

	if writeErr := deps.HookProtocol.WriteOutput(os.Stdout, output); writeErr != nil {
		return fmt.Errorf("write hook output: %w", writeErr)
	}
	return nil
}

// runHookList displays all registered hook handlers.
func runHookList(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	if deps == nil || deps.HookRegistry == nil {
		_, _ = fmt.Fprintln(out, renderInfoCard("Registered Hook Handlers", "Hook system not initialized."))
		return nil
	}

	events := hook.ValidEventTypes()
	totalHandlers := 0
	var pairs []kvPair
	for _, event := range events {
		handlers := deps.HookRegistry.Handlers(event)
		count := len(handlers)
		totalHandlers += count
		if count > 0 {
			label := "handler"
			if count > 1 {
				label = "handlers"
			}
			pairs = append(pairs, kvPair{string(event), fmt.Sprintf("%d %s", count, label)})
		}
	}

	if totalHandlers == 0 {
		_, _ = fmt.Fprintln(out, renderInfoCard("Registered Hook Handlers", "No handlers registered."))
	} else {
		_, _ = fmt.Fprintln(out, renderCard("Registered Hook Handlers", renderKeyValueLines(pairs)))
	}

	return nil
}

// runAgentHook executes an agent-specific hook action.
// Agent actions are like: ddd-pre-transformation, backend-validation, etc.
func runAgentHook(cmd *cobra.Command, args []string) error {
	if deps == nil || deps.HookProtocol == nil || deps.HookRegistry == nil {
		return fmt.Errorf("hook system not initialized")
	}

	action := args[0]

	// Read hook input from stdin
	input, err := deps.HookProtocol.ReadInput(os.Stdin)
	if err != nil {
		return fmt.Errorf("read hook input: %w", err)
	}

	// Determine the event type based on the action suffix
	// PreToolUse: *-validation, *-pre-transformation, *-pre-implementation
	// PostToolUse: *-verification, *-post-transformation, *-post-implementation
	// SubagentStop: *-completion
	var event hook.EventType
	switch {
	case endsWithAny(action, "-validation", "-pre-transformation", "-pre-implementation"):
		event = hook.EventPreToolUse
	case endsWithAny(action, "-verification", "-post-transformation", "-post-implementation"):
		event = hook.EventPostToolUse
	case endsWith(action, "-completion"):
		event = hook.EventSubagentStop
	default:
		// Default to PreToolUse for unknown actions
		event = hook.EventPreToolUse
	}

	// Add action to input for handler identification
	input.Data = fmt.Appendf(nil, `{"action":"%s"}`, action)

	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()

	output, err := deps.HookRegistry.Dispatch(ctx, event, input)
	if err != nil {
		return fmt.Errorf("dispatch agent hook: %w", err)
	}

	if writeErr := deps.HookProtocol.WriteOutput(os.Stdout, output); writeErr != nil {
		return fmt.Errorf("write hook output: %w", writeErr)
	}

	// Exit code 2 for explicit exit code (TeammateIdle, TaskCompleted)
	if output != nil && output.ExitCode == 2 {
		os.Exit(2)
	}

	// Exit code 2 for deny decisions per Claude Code protocol
	if output != nil && output.Decision == hook.DecisionDeny {
		os.Exit(2)
	}

	return nil
}

// endsWith checks if a string ends with any of the given suffixes.
func endsWith(s string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix {
			return true
		}
	}
	return false
}

// endsWithAny is an alias for endsWith for readability.
func endsWithAny(s string, suffixes ...string) bool {
	return endsWith(s, suffixes...)
}

// runDBSchemaSync executes the db-schema-sync hook handler (SPEC-DB-SYNC-001).
// It accepts --file <path> and always exits 0 (non-blocking per REQ-011).
//
// @MX:NOTE: [AUTO] runDBSchemaSync wires CLI flag to dbsync.HandleDBSchemaSync; always non-blocking.
func runDBSchemaSync(cmd *cobra.Command, _ []string) error {
	filePath, _ := cmd.Flags().GetString("file")

	// Resolve project root from cwd
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	// Load migration patterns from db.yaml if available; use defaults otherwise.
	patterns := loadMigrationPatterns(cwd)

	cfg := dbsync.Config{
		FilePath:          filePath,
		MigrationPatterns: patterns,
		ExcludedPatterns:  dbsync.DefaultExcludedPatterns,
		StateFile:         filepath.Join(cwd, ".moai", "cache", "db-sync", "last-seen.json"),
		ProposalFile:      filepath.Join(cwd, ".moai", "cache", "db-sync", "proposal.json"),
		ErrorLogFile:      filepath.Join(cwd, ".moai", "logs", "db-sync-errors.log"),
		DebounceWindow:    10 * time.Second,
	}

	result := dbsync.HandleDBSchemaSync(cfg)

	// REQ-010: emit decision JSON to stdout so orchestrator can act on it.
	out := map[string]string{"decision": result.Decision}
	if encErr := json.NewEncoder(cmd.OutOrStdout()).Encode(out); encErr != nil {
		// Non-fatal: stdout write failure is logged but does not block.
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "db-schema-sync: write output:", encErr)
	}

	// Always exit 0 (non-blocking, REQ-011)
	_ = result.ExitCode
	return nil
}

// runSpecStatus handles the spec-status hook subcommand.
// It reads hook input from stdin and dispatches to the spec status handler.
func runSpecStatus(cmd *cobra.Command, _ []string) error {
	if deps == nil || deps.HookProtocol == nil || deps.HookRegistry == nil {
		return fmt.Errorf("hook system not initialized")
	}

	// Read hook input from stdin
	input, err := deps.HookProtocol.ReadInput(os.Stdin)
	if err != nil {
		return fmt.Errorf("read hook input: %w", err)
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
	defer cancel()

	// Create spec status handler and execute
	handler := hook.NewSpecStatusHandler()
	output, err := handler.Handle(ctx, input)
	if err != nil {
		// Log but don't fail - hook is non-blocking
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "spec-status: error:", err)
		return nil
	}

	if writeErr := deps.HookProtocol.WriteOutput(cmd.OutOrStdout(), output); writeErr != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "spec-status: write output:", writeErr)
	}

	// Always exit 0 (non-blocking)
	return nil
}

// defaultMigrationPatterns are the built-in migration patterns from SPEC-DB-SYNC-001.
var defaultMigrationPatterns = []string{
	"prisma/schema.prisma",
	"alembic/versions/**/*.py",
	"db/migrate/**/*.rb",
	"migrations/**/*.sql",
	"supabase/migrations/**/*.sql",
	"sql/migrations/**/*.sql",
}

// loadMigrationPatterns reads migration_patterns from .moai/config/sections/db.yaml.
// Falls back to defaultMigrationPatterns if the file is absent or unparseable.
func loadMigrationPatterns(projectRoot string) []string {
	dbYAML := filepath.Join(projectRoot, ".moai", "config", "sections", "db.yaml")
	data, err := os.ReadFile(dbYAML)
	if err != nil {
		return defaultMigrationPatterns
	}

	// Simple line-based extraction for migration_patterns list.
	// Full YAML parsing would require gopkg.in/yaml.v3 which is already a dep,
	// but the file format is simple enough for this lightweight approach.
	var patterns []string
	inPatterns := false
	for _, line := range splitLines(string(data)) {
		trimmed := trimSpace(line)
		if trimmed == "migration_patterns:" {
			inPatterns = true
			continue
		}
		if inPatterns {
			if len(trimmed) > 0 && trimmed[0] == '-' {
				pat := trimSpace(trimmed[1:])
				if pat != "" {
					patterns = append(patterns, pat)
				}
			} else if len(trimmed) > 0 && trimmed[0] != ' ' && trimmed[0] != '\t' {
				// New top-level key — stop collecting
				break
			}
		}
	}

	if len(patterns) == 0 {
		return defaultMigrationPatterns
	}
	return patterns
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	left := 0
	for left < len(s) && (s[left] == ' ' || s[left] == '\t') {
		left++
	}
	right := len(s)
	for right > left && (s[right-1] == ' ' || s[right-1] == '\t' || s[right-1] == '\r') {
		right--
	}
	return s[left:right]
}

// isHarnessLearningEnabled reports whether the harness learning subsystem is
// enabled for the project rooted at projectRoot, per REQ-HRN-FND-009 of
// SPEC-V3R4-HARNESS-001.
//
// The gate reads `.moai/config/sections/harness.yaml` and inspects the
// top-level `learning.enabled` key. Truth table:
//
//   - file missing / unreadable          → true  (treat as enabled by default)
//   - YAML parse error                   → true  (fail-open; preserve baseline)
//   - `learning` block absent            → true  (default enabled)
//   - `learning.enabled` absent          → true  (default enabled)
//   - `learning.enabled: true`           → true
//   - `learning.enabled: false`          → false (observer must be a no-op)
//
// The gate is intentionally fail-open so that a corrupted or missing config
// does not silently disable the observer. Explicit `false` is required to
// suppress observation, matching the EARS state-driven semantics of REQ-HRN-FND-009.
func isHarnessLearningEnabled(projectRoot string) bool {
	configPath := filepath.Join(projectRoot, ".moai", "config", "sections", "harness.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return true
	}
	var doc struct {
		Learning struct {
			Enabled *bool `yaml:"enabled,omitempty"`
		} `yaml:"learning,omitempty"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return true
	}
	if doc.Learning.Enabled == nil {
		return true
	}
	return *doc.Learning.Enabled
}

// runHarnessObserve reads PostToolUse hook stdin JSON and records event to usage-log.jsonl.
// T-P1-03: handle-harness-observe.sh → moai hook harness-observe routing implementation.
//
// stdin JSON structure (PostToolUse standard):
//
// {
// "toolName": "Bash" | "Edit" | "Write" | "Agent" | "AskUserQuestion",
// "toolInput": { ... }
// }
//
// @MX:NOTE: [AUTO] REQ-HRN-FND-009 (SPEC-V3R4-HARNESS-001): When learning.enabled
// in harness.yaml resolves to false, this handler is a complete no-op — no read,
// no write, no append to usage-log.jsonl. Existing log entries are not deleted.
// Gate is implemented by isHarnessLearningEnabled (fail-open semantics: missing
// config or parse error preserves baseline observation).
func runHarnessObserve(cmd *cobra.Command, _ []string) error {
	// detect project route: based on cwd
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	// REQ-HRN-FND-009 gate: if learning.enabled is explicitly false, exit no-op.
	// stdin is NOT consumed in the no-op path; the hook exits 0 immediately so
	// the PostToolUse pipeline is non-blocking and leaves usage-log.jsonl untouched.
	if !isHarnessLearningEnabled(cwd) {
		return nil
	}

	// read stdin JSON
	var hookInput struct {
		ToolName string `json:"toolName"`
	}

	decoder := json.NewDecoder(os.Stdin)
	// on parsing failure also exit 0 (non-blocking: does not block parent tool call)
	_ = decoder.Decode(&hookInput)

	logPath := filepath.Join(cwd, ".moai", "harness", "usage-log.jsonl")
	archiveDir := filepath.Join(cwd, ".moai", "harness", "learning-history", "archive")

	retention := harness.NewRetention(logPath, archiveDir, nil)
	obs := harness.NewObserverWithRetention(logPath, retention)

	// use tool name as subject, context hash is empty string
	subject := hookInput.ToolName
	if subject == "" {
		subject = "unknown"
	}

	// log error to stderr but return exit 0 (non-blocking)
	if err := obs.RecordEvent(harness.EventTypeAgentInvocation, subject, ""); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "harness-observe: event recording failed: %v\n", err)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────
// T-A3: Stop 훅 핸들러
// ─────────────────────────────────────────────────────────────

// runHarnessObserveStop는 Claude Code Stop 훅 stdin JSON을 읽고
// session_stop 이벤트를 usage-log.jsonl에 기록한다.
//
// stdin JSON 스키마 (T-A3 spec):
//
//	{"last_assistant_message": "...", "session": {"id": "..."}}
//
// @MX:NOTE: [AUTO] REQ-HRN-OBS-002 + REQ-HRN-OBS-004 + REQ-HRN-OBS-008: Stop 훅 핸들러.
// isHarnessLearningEnabled 게이트 재사용 (REQ-HRN-FND-009).
// last_assistant_message_hash = SHA-256[:16] hex, last_assistant_message_len = byte length.
// 에러는 stderr에 기록하고 exit 0 반환 (비블로킹).
func runHarnessObserveStop(cmd *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	if !isHarnessLearningEnabled(cwd) {
		return nil
	}

	// T-A3 spec: nested stdin JSON — last_assistant_message + session.id
	var hookInput struct {
		LastAssistantMessage string `json:"last_assistant_message"`
		Session              struct {
			ID string `json:"id"`
		} `json:"session"`
	}
	_ = json.NewDecoder(os.Stdin).Decode(&hookInput)

	// subject: cwd 기반 SPEC-ID 탐지 (없으면 빈 문자열)
	subject := detectSpecIDFromCwd(cwd)

	// last_assistant_message_hash + len 계산 (비어있지 않을 때만)
	msgHash, msgLen := assistantMessageFields(hookInput.LastAssistantMessage)

	logPath := filepath.Join(cwd, ".moai", "harness", "usage-log.jsonl")
	archiveDir := filepath.Join(cwd, ".moai", "harness", "learning-history", "archive")
	obs := harness.NewObserverWithRetention(logPath, harness.NewRetention(logPath, archiveDir, nil))

	evt := harness.Event{
		EventType:                harness.EventTypeSessionStop,
		Subject:                  subject,
		ContextHash:              "",
		TierIncrement:            0,
		SessionID:                hookInput.Session.ID,
		LastAssistantMessageHash: msgHash,
		LastAssistantMessageLen:  msgLen,
	}

	if err := obs.RecordExtendedEvent(evt); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "harness-observe-stop: event recording failed: %v\n", err)
	}

	return nil
}

// assistantMessageFields는 last_assistant_message로부터
// last_assistant_message_hash (SHA-256 첫 16 hex 문자)와
// last_assistant_message_len (UTF-8 바이트 길이)을 계산한다.
// msg가 빈 문자열이면 ("", 0)을 반환한다.
//
// @MX:NOTE: [AUTO] REQ-HRN-OBS-002 + REQ-HRN-OBS-004: Stop/SubagentStop 핸들러에서 공유.
// SHA-256[:16]는 16 hex chars = 8 bytes 강도 (충돌 방지보다 빠른 중복 탐지 목적).
func assistantMessageFields(msg string) (hash string, length int) {
	if msg == "" {
		return "", 0
	}
	h := sha256.Sum256([]byte(msg))
	return fmt.Sprintf("%x", h)[:16], len([]byte(msg))
}

// ─────────────────────────────────────────────────────────────
// T-A4: SubagentStop 훅 핸들러
// ─────────────────────────────────────────────────────────────

// runHarnessObserveSubagentStop는 Claude Code SubagentStop 훅 stdin JSON을 읽고
// subagent_stop 이벤트를 usage-log.jsonl에 기록한다.
//
// stdin JSON 스키마 (T-A4 spec):
//
//	{"agentType": "...", "agentName": "...", "last_assistant_message": "...",
//	 "agent_id": "...", "agent_transcript_path": "...", "session": {"id": "..."}}
//
// @MX:NOTE: [AUTO] REQ-HRN-OBS-003 + REQ-HRN-OBS-005 + REQ-HRN-OBS-008: SubagentStop 훅 핸들러.
// isHarnessLearningEnabled 게이트 재사용 (REQ-HRN-FND-009).
// parent_session_id는 session.id (nested) 에서 추출한다.
func runHarnessObserveSubagentStop(cmd *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	if !isHarnessLearningEnabled(cwd) {
		return nil
	}

	// T-A4 spec: camelCase agentType/agentName, nested session.id
	var hookInput struct {
		AgentType            string `json:"agentType"`
		AgentName            string `json:"agentName"`
		LastAssistantMessage string `json:"last_assistant_message"`
		AgentID              string `json:"agent_id"`
		AgentTranscriptPath  string `json:"agent_transcript_path"`
		Session              struct {
			ID string `json:"id"`
		} `json:"session"`
	}
	_ = json.NewDecoder(os.Stdin).Decode(&hookInput)

	// subject: 에이전트 이름 (없으면 "unknown")
	subject := hookInput.AgentName
	if subject == "" {
		subject = "unknown"
	}

	logPath := filepath.Join(cwd, ".moai", "harness", "usage-log.jsonl")
	archiveDir := filepath.Join(cwd, ".moai", "harness", "learning-history", "archive")
	obs := harness.NewObserverWithRetention(logPath, harness.NewRetention(logPath, archiveDir, nil))

	evt := harness.Event{
		EventType:       harness.EventTypeSubagentStop,
		Subject:         subject,
		ContextHash:     "",
		TierIncrement:   0,
		AgentName:       hookInput.AgentName,
		AgentType:       hookInput.AgentType,
		AgentID:         hookInput.AgentID,
		ParentSessionID: hookInput.Session.ID,
	}

	if err := obs.RecordExtendedEvent(evt); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "harness-observe-subagent-stop: event recording failed: %v\n", err)
	}

	return nil
}

// ─────────────────────────────────────────────────────────────
// T-A5: UserPromptSubmit 훅 핸들러 + PII 전략
// ─────────────────────────────────────────────────────────────

// UserPromptStrategy는 UserPromptSubmit 이벤트의 PII 처리 전략 열거형.
// REQ-HRN-OBS-014: 기본값 = StrategyHash (SHA-256 + 길이 + 언어).
type UserPromptStrategy string

const (
	// UserPromptStrategyHash는 기본 전략: SHA-256 해시 + 길이 + 언어 감지.
	// 프롬프트 원문은 기록하지 않음 (PII 최소화).
	UserPromptStrategyHash UserPromptStrategy = "hash"

	// UserPromptStrategyPreview는 opt-in: 프롬프트 앞 200자 미리보기 포함.
	UserPromptStrategyPreview UserPromptStrategy = "preview"

	// UserPromptStrategyFull는 opt-in: 프롬프트 전문 포함 (명시적 동의 필요).
	UserPromptStrategyFull UserPromptStrategy = "full"

	// UserPromptStrategyNone은 UserPromptSubmit 이벤트를 기록하지 않음.
	UserPromptStrategyNone UserPromptStrategy = "none"
)

// specIDRegexp는 SPEC-ID 패턴을 탐지하는 정규식.
var specIDRegexp = regexp.MustCompile(`SPEC-[A-Z][A-Z0-9]+-[0-9]+`)

// resolveUserPromptStrategy는 harness.yaml의 learning.user_prompt_content 값을
// UserPromptStrategy 열거형으로 변환한다.
// REQ-HRN-OBS-014: 알 수 없는 값은 UserPromptStrategyHash로 폴백 (fail-open).
//
// @MX:ANCHOR: [AUTO] resolveUserPromptStrategy는 PII 처리 전략의 단일 결정 지점.
// @MX:REASON: [AUTO] fan_in >= 3: runHarnessObserveUserPromptSubmit, test helpers, future config reload
func resolveUserPromptStrategy(raw string) UserPromptStrategy {
	switch raw {
	case "hash", "":
		return UserPromptStrategyHash
	case "preview":
		return UserPromptStrategyPreview
	case "full":
		return UserPromptStrategyFull
	case "none":
		return UserPromptStrategyNone
	default:
		// fail-open: 알 수 없는 값은 Strategy A로 폴백
		return UserPromptStrategyHash
	}
}

// detectPromptLang은 Unicode 블록 분석으로 프롬프트 언어를 추정한다.
// 반환값: "ko", "ja", "zh", "en", "" (감지 불가).
func detectPromptLang(prompt string) string {
	for _, r := range prompt {
		switch {
		case r >= 0xAC00 && r <= 0xD7A3: // Hangul syllables
			return "ko"
		case (r >= 0x3040 && r <= 0x309F) || (r >= 0x30A0 && r <= 0x30FF): // Hiragana/Katakana
			return "ja"
		case r >= 0x4E00 && r <= 0x9FFF: // CJK Unified Ideographs
			return "zh"
		case r <= 0x007F && unicode.IsLetter(r): // ASCII letters
			return "en"
		}
	}
	return ""
}

// detectSpecIDFromCwd는 cwd 경로에서 SPEC-ID를 탐지한다.
// worktree 경로 패턴에서 SPEC-ID를 추출하는 데 사용된다.
func detectSpecIDFromCwd(cwd string) string {
	match := specIDRegexp.FindString(cwd)
	return match
}

// readUserPromptContentStrategy는 harness.yaml에서 learning.user_prompt_content 값을 읽는다.
func readUserPromptContentStrategy(projectRoot string) string {
	configPath := filepath.Join(projectRoot, ".moai", "config", "sections", "harness.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}
	var doc struct {
		Learning struct {
			UserPromptContent string `yaml:"user_prompt_content,omitempty"`
		} `yaml:"learning,omitempty"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return ""
	}
	return doc.Learning.UserPromptContent
}

// runHarnessObserveUserPromptSubmit는 Claude Code UserPromptSubmit 훅 stdin JSON을 읽고
// user_prompt 이벤트를 usage-log.jsonl에 기록한다.
// 기본 전략(Strategy A): SHA-256 해시 + 길이 + 언어 (PII 최소화).
//
// @MX:WARN: [AUTO] PII 민감 핸들러 — 사용자 프롬프트 원문이 Strategy C 활성화 시 로그에 기록됨.
// @MX:REASON: [AUTO] REQ-HRN-OBS-014: 기본값은 Strategy A로 원문 미기록. opt-in만 원문 기록.
func runHarnessObserveUserPromptSubmit(cmd *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	if !isHarnessLearningEnabled(cwd) {
		return nil
	}

	var hookInput struct {
		Prompt string `json:"prompt"`
	}
	_ = json.NewDecoder(os.Stdin).Decode(&hookInput)

	prompt := hookInput.Prompt

	// SPEC-ID 탐지: 프롬프트에서 SPEC-ID 추출 (없으면 cwd 기반 탐지)
	subject := specIDRegexp.FindString(prompt)
	if subject == "" {
		subject = detectSpecIDFromCwd(cwd)
	}

	// PII 전략 결정 (REQ-HRN-OBS-014)
	strategyRaw := readUserPromptContentStrategy(cwd)
	strategy := resolveUserPromptStrategy(strategyRaw)

	// Strategy None: 이벤트 기록 없음
	if strategy == UserPromptStrategyNone {
		return nil
	}

	// SHA-256 해시 [:16] + byte 길이 + 언어 (Strategy A 기본, B/C 포함)
	// REQ-HRN-OBS-006 / AC-HRN-OBS-004: prompt_hash = SHA-256(prompt) 앞 16 hex 문자,
	// prompt_len = UTF-8 byte 수 (rune 수 아님 — multi-byte 언어 byte 크기 보존)
	h := sha256.Sum256([]byte(prompt))
	promptHash := fmt.Sprintf("%x", h)[:16]
	promptLen := len([]byte(prompt))
	promptLang := detectPromptLang(prompt)

	logPath := filepath.Join(cwd, ".moai", "harness", "usage-log.jsonl")
	archiveDir := filepath.Join(cwd, ".moai", "harness", "learning-history", "archive")
	obs := harness.NewObserverWithRetention(logPath, harness.NewRetention(logPath, archiveDir, nil))

	evt := harness.Event{
		EventType:     harness.EventTypeUserPrompt,
		Subject:       subject,
		ContextHash:   "",
		TierIncrement: 0,
		PromptHash:    promptHash,
		PromptLen:     promptLen,
		PromptLang:    promptLang,
	}

	// opt-in: preview (Strategy B) — REQ-HRN-OBS-013: 첫 64바이트 (UTF-8 경계 안전).
	if strategy == UserPromptStrategyPreview && len(prompt) > 0 {
		b := []byte(prompt)
		end := min(len(b), 64)
		// UTF-8 경계 안전: 64바이트 지점이 멀티바이트 룬 중간이면 유효한 경계까지 후퇴.
		for end > 0 && !utf8.Valid(b[:end]) {
			end--
		}
		evt.PromptPreview = string(b[:end])
	}

	// opt-in: full (Strategy C)
	if strategy == UserPromptStrategyFull {
		evt.PromptContent = prompt
	}

	if err := obs.RecordExtendedEvent(evt); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "harness-observe-user-prompt-submit: event recording failed: %v\n", err)
	}

	return nil
}
