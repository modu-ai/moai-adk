package hook

import (
	"context"
	"encoding/json"
	"io"
	"slices"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// DefaultHookTimeout is the default timeout for hook execution (30 seconds).
const DefaultHookTimeout = 30 * time.Second

// EventType represents a Claude Code hook event type.
type EventType string

const (
	// EventSessionStart is triggered when a new Claude Code session begins.
	EventSessionStart EventType = "SessionStart"

	// EventPreToolUse is triggered before a tool is executed.
	EventPreToolUse EventType = "PreToolUse"

	// EventPostToolUse is triggered after a tool has been executed.
	EventPostToolUse EventType = "PostToolUse"

	// EventSessionEnd is triggered when a Claude Code session ends.
	EventSessionEnd EventType = "SessionEnd"

	// EventStop is triggered when Claude Code requests a stop.
	EventStop EventType = "Stop"

	// EventSubagentStop is triggered when a subagent stops.
	EventSubagentStop EventType = "SubagentStop"

	// EventPreCompact is triggered before context compaction.
	EventPreCompact EventType = "PreCompact"

	// EventPostToolUseFailure is triggered when a tool execution fails.
	EventPostToolUseFailure EventType = "PostToolUseFailure"

	// EventNotification is triggered when Claude Code sends a notification.
	EventNotification EventType = "Notification"

	// EventSubagentStart is triggered when a subagent starts.
	EventSubagentStart EventType = "SubagentStart"

	// EventUserPromptSubmit is triggered when a user submits a prompt.
	EventUserPromptSubmit EventType = "UserPromptSubmit"

	// EventPermissionRequest is triggered when a permission check occurs.
	EventPermissionRequest EventType = "PermissionRequest"

	// EventTeammateIdle is triggered when a teammate goes idle in Agent Teams.
	EventTeammateIdle EventType = "TeammateIdle"

	// EventTaskCompleted is triggered when a task is completed in Agent Teams.
	EventTaskCompleted EventType = "TaskCompleted"

	// EventWorktreeCreate is triggered when a worktree is created for an agent with isolation: worktree.
	// Available since Claude Code v2.1.49+.
	EventWorktreeCreate EventType = "WorktreeCreate"

	// EventWorktreeRemove is triggered when a worktree is removed after an isolated agent terminates.
	// Available since Claude Code v2.1.49+.
	EventWorktreeRemove EventType = "WorktreeRemove"

	// EventPostCompact is triggered after context compaction completes.
	// Available since Claude Code v2.1.76+.
	EventPostCompact EventType = "PostCompact"

	// EventInstructionsLoaded is triggered when CLAUDE.md or .claude/rules/*.md files are loaded.
	// Available since Claude Code v2.1.69+.
	EventInstructionsLoaded EventType = "InstructionsLoaded"

	// EventStopFailure is triggered when a turn ends due to an API error (rate limit, auth failure).
	// Available since Claude Code v2.1.78+.
	EventStopFailure EventType = "StopFailure"

	// EventSetup is triggered via --init, --init-only, or --maintenance CLI flags.
	// Available since Claude Code v2.1.10+.
	EventSetup EventType = "Setup"

	// EventConfigChange is triggered when configuration files change during a session.
	// Available since Claude Code v2.1.49+.
	EventConfigChange EventType = "ConfigChange"

	// EventTaskCreated is triggered when a task is created via TaskCreate.
	// Available since Claude Code v2.1.84+.
	EventTaskCreated EventType = "TaskCreated"

	// EventCwdChanged is triggered when the working directory changes during a session.
	// Available since Claude Code v2.1.83+.
	EventCwdChanged EventType = "CwdChanged"

	// EventFileChanged is triggered when a file is changed externally during a session.
	// Available since Claude Code v2.1.83+.
	EventFileChanged EventType = "FileChanged"

	// EventElicitation is triggered when an MCP server requests user input.
	// Available since Claude Code v2.1.76+.
	EventElicitation EventType = "Elicitation"

	// EventElicitationResult is triggered after user responds to MCP elicitation.
	// Available since Claude Code v2.1.76+.
	EventElicitationResult EventType = "ElicitationResult"

	// EventPermissionDenied is triggered after auto mode classifier denies a tool call.
	// Return {retry: true} in hook output to allow the model to retry the operation.
	// Available since Claude Code v2.1.89+.
	EventPermissionDenied EventType = "PermissionDenied"
)

// ValidEventTypes returns all valid event types.
func ValidEventTypes() []EventType {
	return []EventType{
		EventSessionStart,
		EventPreToolUse,
		EventPostToolUse,
		EventSessionEnd,
		EventStop,
		EventSubagentStop,
		EventPreCompact,
		EventPostToolUseFailure,
		EventNotification,
		EventSubagentStart,
		EventUserPromptSubmit,
		EventPermissionRequest,
		EventTeammateIdle,
		EventTaskCompleted,
		EventWorktreeCreate,
		EventWorktreeRemove,
		EventPostCompact,
		EventInstructionsLoaded,
		EventStopFailure,
		EventSetup,
		EventConfigChange,
		EventTaskCreated,
		EventCwdChanged,
		EventFileChanged,
		EventElicitation,
		EventElicitationResult,
		EventPermissionDenied,
	}
}

// IsValidEventType checks if the given event type is valid.
func IsValidEventType(et EventType) bool {
	return slices.Contains(ValidEventTypes(), et)
}

// Permission decision constants for PreToolUse hooks (Claude Code protocol).
const (
	DecisionAllow = "allow"
	DecisionDeny  = "deny"
	DecisionAsk   = "ask"
	DecisionDefer = "defer" // Pause headless session; resume with --resume (v2.1.89+)
)

// Top-level decision constants for Stop, PostToolUse, etc. (Claude Code protocol).
const (
	DecisionBlock = "block" // Used in top-level decision field for Stop, PostToolUse, etc.
)

// HookInput represents the JSON payload received from Claude Code via stdin.
// Fields follow the official Claude Code hooks protocol.
type HookInput struct {
	// Common fields (all events)
	SessionID      string `json:"session_id,omitempty"`
	TranscriptPath string `json:"transcript_path,omitempty"`
	CWD            string `json:"cwd,omitempty"`
	PermissionMode string `json:"permission_mode,omitempty"` // default, plan, acceptEdits, dontAsk, bypassPermissions
	HookEventName  string `json:"hook_event_name,omitempty"`

	// Tool-related fields (PreToolUse, PostToolUse, PostToolUseFailure, PermissionRequest)
	ToolName     string          `json:"tool_name,omitempty"`
	ToolInput    json.RawMessage `json:"tool_input,omitempty"`
	ToolOutput   json.RawMessage `json:"tool_output,omitempty"`   // Legacy field
	ToolResponse json.RawMessage `json:"tool_response,omitempty"` // PostToolUse result
	ToolUseID    string          `json:"tool_use_id,omitempty"`

	// SessionStart fields
	Source    string `json:"source,omitempty"`     // startup, resume, clear, compact
	Model     string `json:"model,omitempty"`      // Model identifier
	AgentType string `json:"agent_type,omitempty"` // Custom agent name if --agent flag used

	// SessionEnd fields
	Reason string `json:"reason,omitempty"` // clear, logout, prompt_input_exit, bypass_permissions_disabled, other

	// Stop/SubagentStop fields
	StopHookActive bool `json:"stop_hook_active,omitempty"` // True when already continuing due to stop hook

	// SubagentStart/SubagentStop fields
	AgentID             string `json:"agent_id,omitempty"`
	AgentTranscriptPath string `json:"agent_transcript_path,omitempty"`

	// PreCompact fields
	Trigger            string `json:"trigger,omitempty"`             // manual, auto
	CustomInstructions string `json:"custom_instructions,omitempty"` // User instructions for /compact

	// PostToolUseFailure fields
	Error       string `json:"error,omitempty"`
	IsInterrupt bool   `json:"is_interrupt,omitempty"`

	// UserPromptSubmit fields
	Prompt string `json:"prompt,omitempty"`

	// Notification fields
	Message          string `json:"message,omitempty"`
	Title            string `json:"title,omitempty"`
	NotificationType string `json:"notification_type,omitempty"`

	// Legacy/internal field (deprecated, use CWD instead)
	ProjectDir string `json:"project_dir,omitempty"`

	// TeammateIdle and TaskCompleted fields (Agent Teams v2.1.33+)
	TeamName        string `json:"team_name,omitempty"`
	TeammateName    string `json:"teammate_name,omitempty"`
	TaskID          string `json:"task_id,omitempty"`
	TaskSubject     string `json:"task_subject,omitempty"`
	TaskDescription string `json:"task_description,omitempty"`

	// WorktreeCreate and WorktreeRemove fields (v2.1.49+)
	WorktreePath   string `json:"worktree_path,omitempty"`   // Absolute path to the worktree directory
	WorktreeBranch string `json:"worktree_branch,omitempty"` // Branch name for the worktree
	AgentName      string `json:"agent_name,omitempty"`      // Name of the agent using the worktree

	// ConfigChange fields (v2.1.49+)
	ConfigFilePath string `json:"config_file_path,omitempty"` // Path to the changed configuration file

	// TaskCreated fields (v2.1.84+)
	// Reuses TaskID, TaskSubject, TaskDescription from TeammateIdle/TaskCompleted

	// FileChanged fields (v2.1.83+)
	FilePath string `json:"file_path,omitempty"` // Path to the changed file

	// Elicitation fields (v2.1.76+)
	ElicitationServerName string `json:"elicitation_server_name,omitempty"` // MCP server requesting input

	// Internal data (not serialized to JSON)
	Data json.RawMessage `json:"-"`
}

// HookSpecificOutput represents the hookSpecificOutput field for PreToolUse/PostToolUse.
// UserPromptSubmit에서 sessionTitle을 설정하면 Claude Code가 세션 타이틀을 변경한다 (v2.1.94+).
type HookSpecificOutput struct {
	HookEventName            string `json:"hookEventName,omitempty"`
	PermissionDecision       string `json:"permissionDecision,omitempty"`
	PermissionDecisionReason string `json:"permissionDecisionReason,omitempty"`
	AdditionalContext        string `json:"additionalContext,omitempty"`
	// SessionTitle은 UserPromptSubmit 훅에서만 사용되며,
	// Claude Code 2.1.94+에서 세션 타이틀 자동 설정을 지원한다.
	SessionTitle string `json:"sessionTitle,omitempty"`
}

// HookOutput represents the JSON payload written to stdout for Claude Code.
// The format varies by event type per Claude Code protocol.
type HookOutput struct {
	// Universal fields (all events)
	Continue       bool   `json:"continue,omitempty"`       // If false, Claude stops processing entirely
	StopReason     string `json:"stopReason,omitempty"`     // Message shown when continue is false
	SystemMessage  string `json:"systemMessage,omitempty"`  // Warning message shown to user
	SuppressOutput bool   `json:"suppressOutput,omitempty"` // If true, hides stdout from verbose mode

	// Top-level decision fields (Stop, SubagentStop, PostToolUse, PostToolUseFailure, UserPromptSubmit)
	// Use "block" to prevent the action; omit to allow
	Decision string `json:"decision,omitempty"` // "block" to prevent action
	Reason   string `json:"reason,omitempty"`   // Explanation when decision is "block"

	// For PreToolUse/PostToolUse/PermissionRequest: hook-specific output
	HookSpecificOutput *HookSpecificOutput `json:"hookSpecificOutput,omitempty"`

	// UpdatedInput is used by UserPromptSubmit to modify the user's prompt.
	UpdatedInput string `json:"updatedInput,omitempty"`

	// Retry signals that the model should retry the denied operation (PermissionDenied, v2.1.89+).
	Retry bool `json:"retry,omitempty"`

	// ExitCode allows handlers to signal a specific process exit code.
	// Not serialized to JSON. Used for exit code 2 protocol (TeammateIdle, TaskCompleted).
	ExitCode int `json:"-"`

	// Internal data (not serialized to JSON)
	Data json.RawMessage `json:"-"`
}

// @MX:ANCHOR: [AUTO] PreToolUse 허용 응답 팩토리. 모든 hook 핸들러에서 기본 허용 응답을 생성할 때 사용.
// @MX:REASON: fan_in=19, 가장 많이 사용되는 hook 응답 팩토리, 변경 시 전체 핸들러 동작에 영향
// NewAllowOutput creates a HookOutput with permissionDecision "allow" for PreToolUse.
// Per Claude Code protocol, PreToolUse uses hookSpecificOutput.permissionDecision.
func NewAllowOutput() *HookOutput {
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: DecisionAllow,
		},
	}
}

// NewAllowOutputWithData creates a HookOutput with permissionDecision "allow" and internal data.
// Per Claude Code protocol, PreToolUse uses hookSpecificOutput.permissionDecision.
func NewAllowOutputWithData(data json.RawMessage) *HookOutput {
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: DecisionAllow,
		},
		Data: data,
	}
}

// @MX:ANCHOR: [AUTO] PreToolUse 차단 응답 팩토리. 도구 실행을 거부할 때 사용되는 핵심 팩토리.
// @MX:REASON: fan_in=7, 차단 결정의 단일 진입점, 프로토콜 준수 필수
// NewDenyOutput creates a HookOutput with permissionDecision "deny" for PreToolUse.
// Per Claude Code protocol, PreToolUse uses hookSpecificOutput.permissionDecision.
func NewDenyOutput(reason string) *HookOutput {
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       DecisionDeny,
			PermissionDecisionReason: reason,
		},
	}
}

// @MX:ANCHOR: [AUTO] PreToolUse 권한 요청 응답 팩토리. 사용자 확인이 필요한 경우 사용.
// @MX:REASON: fan_in=3, 권한 요청 응답의 단일 생성 지점
// NewAskOutput creates a HookOutput with permissionDecision "ask" for PreToolUse.
// Per Claude Code protocol, PreToolUse uses hookSpecificOutput.permissionDecision.
func NewAskOutput(reason string) *HookOutput {
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       DecisionAsk,
			PermissionDecisionReason: reason,
		},
	}
}

// NewBlockOutput creates a HookOutput with permissionDecision "deny" for PreToolUse.
// This is an alias for NewDenyOutput. For Stop/PostToolUse, use NewStopBlockOutput instead.
func NewBlockOutput(reason string) *HookOutput {
	return NewDenyOutput(reason)
}

// NewSuppressOutput creates a HookOutput that suppresses output.
func NewSuppressOutput() *HookOutput {
	return &HookOutput{SuppressOutput: true}
}

// NewSessionOutput creates a HookOutput for SessionStart/SessionEnd events.
func NewSessionOutput(continueSession bool, message string) *HookOutput {
	return &HookOutput{
		Continue:      continueSession,
		SystemMessage: message,
	}
}

// NewPostToolOutput creates a HookOutput with additionalContext for PostToolUse.
func NewPostToolOutput(context string) *HookOutput {
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     "PostToolUse",
			AdditionalContext: context,
		},
	}
}

// NewStopBlockOutput creates a HookOutput that prevents Claude from stopping.
// Use this for Stop and SubagentStop hooks when you want Claude to continue working.
// Per Claude Code protocol, Stop hooks use top-level decision/reason, not hookSpecificOutput.
func NewStopBlockOutput(reason string) *HookOutput {
	return &HookOutput{
		Decision: DecisionBlock,
		Reason:   reason,
	}
}

// NewPostToolBlockOutput creates a HookOutput that blocks after tool execution.
// Use this for PostToolUse hooks when you want to provide feedback that stops Claude.
// Per Claude Code protocol, PostToolUse uses top-level decision/reason.
func NewPostToolBlockOutput(reason string, additionalContext string) *HookOutput {
	output := &HookOutput{
		Decision: DecisionBlock,
		Reason:   reason,
	}
	if additionalContext != "" {
		output.HookSpecificOutput = &HookSpecificOutput{
			HookEventName:     "PostToolUse",
			AdditionalContext: additionalContext,
		}
	}
	return output
}

// NewPermissionRequestOutput creates a HookOutput for PermissionRequest events.
// Per Claude Code protocol (v2.1.59+), hookSpecificOutput.hookEventName must be
// "PermissionRequest" for PermissionRequest events.
func NewPermissionRequestOutput(decision, reason string) *HookOutput {
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:            "PermissionRequest",
			PermissionDecision:       decision,
			PermissionDecisionReason: reason,
		},
	}
}

// NewUserPromptBlockOutput creates a HookOutput that blocks user prompt processing.
func NewUserPromptBlockOutput(reason string) *HookOutput {
	return &HookOutput{
		Decision: DecisionBlock,
		Reason:   reason,
	}
}

// NewDeferOutput creates a HookOutput with permissionDecision "defer" for PreToolUse.
// In headless sessions (-p mode), this pauses execution until resumed with --resume (v2.1.89+).
func NewDeferOutput(reason string) *HookOutput {
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       DecisionDefer,
			PermissionDecisionReason: reason,
		},
	}
}

// NewPermissionDeniedRetryOutput creates a HookOutput for PermissionDenied that signals retry.
// Return this from a PermissionDenied hook handler to allow the model to retry the denied operation.
func NewPermissionDeniedRetryOutput() *HookOutput {
	return &HookOutput{Retry: true}
}

// NewTeammateKeepWorkingOutput creates a HookOutput that signals exit code 2 for TeammateIdle.
// Exit code 2 tells Claude Code to keep the teammate working.
func NewTeammateKeepWorkingOutput() *HookOutput {
	return &HookOutput{ExitCode: 2}
}

// NewTaskRejectedOutput creates a HookOutput that signals exit code 2 for TaskCompleted.
// Exit code 2 tells Claude Code to reject the task completion.
func NewTaskRejectedOutput() *HookOutput {
	return &HookOutput{ExitCode: 2}
}

// Handler processes a specific hook event type.
type Handler interface {
	// Handle processes the hook input and returns output.
	// ctx carries cancellation and timeout signals.
	Handle(ctx context.Context, input *HookInput) (*HookOutput, error)

	// EventType returns the event type this handler processes.
	EventType() EventType
}

// Registry manages handler registration and event dispatching.
type Registry interface {
	// Register adds a handler to the registry for its declared event type.
	Register(handler Handler)

	// Dispatch sends an event to all registered handlers for the given event type.
	// Handlers are executed sequentially. If any handler returns Decision "block",
	// remaining handlers are skipped and the block result is returned immediately.
	Dispatch(ctx context.Context, event EventType, input *HookInput) (*HookOutput, error)

	// Handlers returns all handlers registered for the given event type.
	Handlers(event EventType) []Handler
}

// Protocol handles JSON communication with Claude Code via stdin/stdout.
type Protocol interface {
	// ReadInput reads and parses JSON from the given reader.
	ReadInput(r io.Reader) (*HookInput, error)

	// WriteOutput serializes the output as JSON to the given writer.
	WriteOutput(w io.Writer, output *HookOutput) error
}

// Contract defines the hook execution contract per ADR-012.
type Contract interface {
	// Validate checks that the execution environment meets contract requirements.
	Validate(ctx context.Context) error

	// Guarantees returns the list of guaranteed execution conditions.
	Guarantees() []string

	// NonGuarantees returns the list of non-guaranteed execution conditions.
	NonGuarantees() []string
}

// ConfigProvider provides read access to application configuration.
// It is satisfied by *config.ConfigManager.
type ConfigProvider interface {
	Get() *config.Config
}
