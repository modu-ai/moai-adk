package cli

import (
	"context"
	"errors"
	"strings"

	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/spf13/cobra"
)

type gptCommandServices struct {
	Login, Logout, Status func(context.Context) error
	Launch                func(string, string, []string) error
}

func newGPTCommand(services gptCommandServices) *cobra.Command {
	cmd := &cobra.Command{Use: "gpt", Short: "Launch Claude Code with the GPT gateway", Long: "Launch Claude Code with GPT (default: gpt-5.6-sol).\nUse gpt login, gpt logout, or gpt status to manage the private GPT login.\nLauncher flags include -p, --spawn, -w, --continue, and --model.", GroupID: "launch", DisableFlagParsing: true, SilenceUsage: true}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
			if len(args) != 1 {
				return errors.New("gpt login, logout and status do not accept additional arguments")
			}
			var operation func(context.Context) error
			switch args[0] {
			case "login":
				operation = services.Login
			case "logout":
				operation = services.Logout
			case "status":
				operation = services.Status
			default:
				return errors.New("unknown gpt command; use login, logout, status, or launcher flags")
			}
			if operation == nil {
				return errors.New("GPT authentication service is unavailable")
			}
			return operation(cmd.Context())
		}
		launch := services.Launch
		if launch == nil {
			for _, arg := range args {
				if arg == "--" {
					break
				}
				if arg == "--help" || arg == "-h" {
					return cmd.Help()
				}
			}
			if err := guardCGLaunchMode("gpt"); err != nil {
				return err
			}

			return errors.New("GPT gateway launch is awaiting transport verification; use moai gpt status to inspect login")
		}
		if err := rejectRetiredGPTKanban(args); err != nil {
			return err
		}
		return runClaudeEntry(cmd, args, "gpt", "gpt", kanban.BackendGPT, launch)
	}
	return cmd
}

func rejectRetiredGPTKanban(args []string) error {
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if arg == kanbanFlagShort || arg == kanbanFlagLong || strings.HasPrefix(arg, kanbanFlagShort+"=") || strings.HasPrefix(arg, kanbanFlagLong+"=") {
			return errors.New("moai gpt -k/--kanban is retired; use moai gpt -f [N] for Factory Mode")
		}
	}
	return nil
}
