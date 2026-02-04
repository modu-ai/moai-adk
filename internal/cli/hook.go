package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk-go/internal/hook"
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
}

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

	// Exit code 2 for deny decisions per Claude Code protocol
	if output != nil && output.Decision == hook.DecisionDeny {
		os.Exit(2)
	}

	return nil
}

// runHookList displays all registered hook handlers.
func runHookList(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	fmt.Fprintln(out, "Registered Hook Handlers")
	fmt.Fprintln(out, "========================")
	fmt.Fprintln(out)

	if deps == nil || deps.HookRegistry == nil {
		fmt.Fprintln(out, "  Hook system not initialized.")
		return nil
	}

	events := hook.ValidEventTypes()
	totalHandlers := 0
	for _, event := range events {
		handlers := deps.HookRegistry.Handlers(event)
		count := len(handlers)
		totalHandlers += count
		if count > 0 {
			fmt.Fprintf(out, "  %s: %d handler(s)\n", string(event), count)
		}
	}

	if totalHandlers == 0 {
		fmt.Fprintln(out, "  No handlers registered.")
	}

	return nil
}
