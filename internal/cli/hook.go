package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/hook"
)

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Execute hook event handlers",
	Long:  "Execute Claude Code hook event handlers. Called by Claude Code settings.json hook configuration.",
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
