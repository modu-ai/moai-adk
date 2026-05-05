package cli

import (
"context"
"encoding/json"
"fmt"
"os"
"path/filepath"
"time"

"github.com/spf13/cobra"

"github.com/modu-ai/moai-adk/internal/harness"
"github.com/modu-ai/moai-adk/internal/hook"
"github.com/modu-ai/moai-adk/internal/hook/dbsync"
)

var hookCmd = &cobra.Command{
Use: "hook",
Short: "Execute hook event handlers",
GroupID: "tools",
Long: "Execute Claude Code hook event handlers. Called by Claude Code settings.json hook configuration.",
}

func init() {
rootCmd.AddCommand(hookCmd)

// Register all hook subcommands
hookSubcommands := []struct {
use string
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
{"setup", "Handle setup event", hook.EventSetup},
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
Use: sub.use,
Short: sub.short,
RunE: func(cmd *cobra.Command, _ []string) error {
return runHookEvent(cmd, event)
},
}
hookCmd.AddCommand(cmd)
}

// Add "list" subcommand
hookCmd.AddCommand(&cobra.Command{
Use: "list",
Short: "List registered hook handlers",
RunE: runHookList,
})

// Add "agent" subcommand for agent-specific hooks
hookCmd.AddCommand(&cobra.Command{
Use: "agent [action]",
Short: "Execute agent-specific hook action",
Long: "Execute agent-specific hook actions like ddd-pre-transformation, backend-validation, etc.",
Args: cobra.ExactArgs(1),
RunE: runAgentHook,
})

// harness-observe Add subcommands (SPEC-V3R3-HARNESS-LEARNING-001 T-P1-03)
// Read PostToolUse hook stdin JSON and record event to usage-log.jsonl.
hookCmd.AddCommand(&cobra.Command{
Use: "harness-observe",
Short: "Record PostToolUse event to harness usage log",
Long: "Reads hook stdin JSON and appends an event to .moai/harness/usage-log.jsonl. Called from handle-harness-observe.sh.",
RunE: runHarnessObserve,
})

// Add "db-schema-sync" subcommand (SPEC-DB-SYNC-001)
dbSchemaSyncCmd := &cobra.Command{
Use: "db-schema-sync",
Short: "Handle DB schema change detection from PostToolUse hook",
Long: "Detect migration file changes, apply debounce, parse, and write proposal.json for user approval.",
RunE: runDBSchemaSync,
}
dbSchemaSyncCmd.Flags().String("file", "", "File path from PostToolUse hook stdin")
hookCmd.AddCommand(dbSchemaSyncCmd)

// Add "spec-status" subcommand (SPEC-STATUS-AUTO-001)
specStatusCmd := &cobra.Command{
Use: "spec-status",
Short: "Auto-update SPEC status on git commit",
Long: "Extract SPEC-IDs from git commit messages and update their status to 'implemented'. Called from handle-spec-status.sh.",
RunE: runSpecStatus,
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

ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
defer cancel()

output, err := deps.HookRegistry.Dispatch(ctx, event, input)
if err != nil {
return fmt.Errorf("dispatch hook: %w", err)
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
FilePath: filePath,
MigrationPatterns: patterns,
ExcludedPatterns: dbsync.DefaultExcludedPatterns,
StateFile: filepath.Join(cwd, ".moai", "cache", "db-sync", "last-seen.json"),
ProposalFile: filepath.Join(cwd, ".moai", "cache", "db-sync", "proposal.json"),
ErrorLogFile: filepath.Join(cwd, ".moai", "logs", "db-sync-errors.log"),
DebounceWindow: 10 * time.Second,
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
// @MX:NOTE: [AUTO] learning.enabled Configuration/Settings gate Phase 4from/in/at addition planned/scheduled (T-P4-XX).
func runHarnessObserve(cmd *cobra.Command, _ []string) error {
// read stdin JSON
var hookInput struct {
ToolName string `json:"toolName"`
}

decoder := json.NewDecoder(os.Stdin)
// on parsing failure also exit 0 (non-blocking: does not block parent tool call)
_ = decoder.Decode(&hookInput)

// detect project route: based on cwd
cwd, err := os.Getwd()
if err != nil {
cwd = "."
}

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
