package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anthropics/moai-adk-go/internal/output"
	"github.com/anthropics/moai-adk-go/internal/rank"
	"github.com/spf13/cobra"
)

// NewRankCommand creates the rank command group
func NewRankCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rank",
		Short: "MoAI Rank leaderboard commands",
		Long:  `Manage MoAI Rank leaderboard registration, status, and session tracking.`,
	}

	cmd.AddCommand(
		newRankLoginCmd(),
		newRankStatusCmd(),
		newRankLogoutCmd(),
		newRankExcludeCmd(),
		newRankIncludeCmd(),
		newRankSyncCmd(),
	)

	return cmd
}

func newRankLoginCmd() *cobra.Command {
	var apiKey string

	cmd := &cobra.Command{
		Use:     "login",
		Aliases: []string{"register"},
		Short:   "Login to MoAI Rank via GitHub OAuth",
		RunE: func(cmd *cobra.Command, args []string) error {
			if rank.IsLoggedIn() {
				creds, err := rank.LoadCredentials()
				if err == nil {
					fmt.Println(output.WarningStyle.Render(fmt.Sprintf("Already logged in as %s", creds.Username)))
					fmt.Println(output.MutedStyle.Render("  Use 'moai rank logout' first to switch accounts."))
					return nil
				}
			}

			var creds *rank.Credentials
			var err error

			if apiKey != "" {
				fmt.Println(output.InfoStyle.Render("Validating API key..."))
				creds, err = rank.LoginWithAPIKey(apiKey)
			} else {
				fmt.Println(output.InfoStyle.Render("Opening browser for GitHub authentication..."))
				fmt.Println(output.MutedStyle.Render("  Waiting for authorization..."))
				creds, err = rank.Login()
			}

			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("Logged in as %s", creds.Username)))

			// Install session tracking hook
			if hookErr := rank.InstallHook(); hookErr != nil {
				fmt.Println(output.WarningStyle.Render(fmt.Sprintf("Warning: failed to install hook: %v", hookErr)))
				fmt.Println(output.MutedStyle.Render("  Session tracking may not work. Try 'moai rank login' again."))
			} else {
				fmt.Println(output.SuccessStyle.Render("Session tracking hook installed"))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&apiKey, "api-key", "", "login with an API key instead of OAuth")
	return cmd
}

func newRankStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show your rank and statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			status, err := rank.GetStatus()
			if err != nil && !status.LoggedIn {
				fmt.Println(output.MutedStyle.Render("Not logged in."))
				fmt.Println(output.MutedStyle.Render("  Run 'moai rank login' to get started."))
				return nil
			}

			fmt.Println(output.HeaderStyle.Render("MoAI Rank Status"))
			fmt.Println()

			// User info
			fmt.Printf("  %s %s\n", output.InfoStyle.Render("User:"), status.Username)

			// Hook status
			hookStatus := output.SuccessStyle.Render("installed")
			if !rank.IsHookInstalled() {
				hookStatus = output.WarningStyle.Render("not installed")
			}
			fmt.Printf("  %s %s\n", output.InfoStyle.Render("Hook:"), hookStatus)
			fmt.Println()

			if err != nil {
				fmt.Println(output.WarningStyle.Render(fmt.Sprintf("  Could not fetch rank data: %v", err)))
				return nil
			}

			// Rankings
			if len(status.Ranks) > 0 {
				fmt.Println(output.SectionStyle.Render("  Rankings"))
				periods := []string{"daily", "weekly", "monthly", "all_time"}
				periodNames := map[string]string{
					"daily": "Daily", "weekly": "Weekly",
					"monthly": "Monthly", "all_time": "All Time",
				}
				for _, period := range periods {
					if info, ok := status.Ranks[period]; ok {
						medal := rank.RankMedal(info.Position)
						fmt.Printf("    %s %s #%d / %d (score: %.1f)\n",
							periodNames[period], medal, info.Position,
							info.TotalParticipants, info.CompositeScore)
					}
				}
				fmt.Println()
			}

			// Token stats
			fmt.Println(output.SectionStyle.Render("  Token Usage"))
			fmt.Printf("    Total:  %s\n", rank.FormatTokens(status.Tokens.Total))
			fmt.Printf("    Input:  %s\n", rank.FormatTokens(status.Tokens.Input))
			fmt.Printf("    Output: %s\n", rank.FormatTokens(status.Tokens.Output))

			return nil
		},
	}
}

func newRankLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout and remove credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !rank.IsLoggedIn() {
				fmt.Println(output.MutedStyle.Render("Not logged in."))
				return nil
			}

			// Uninstall hook
			if hookErr := rank.UninstallHook(); hookErr != nil {
				fmt.Println(output.WarningStyle.Render(fmt.Sprintf("Warning: failed to remove hook: %v", hookErr)))
			}

			// Delete credentials
			if err := rank.DeleteCredentials(); err != nil {
				return fmt.Errorf("failed to remove credentials: %w", err)
			}

			fmt.Println(output.SuccessStyle.Render("Logged out successfully"))
			fmt.Println(output.MutedStyle.Render("  Session tracking hook removed."))
			return nil
		},
	}
}

func newRankExcludeCmd() *cobra.Command {
	var listExcluded bool

	cmd := &cobra.Command{
		Use:   "exclude [path]",
		Short: "Exclude a project from rank tracking",
		RunE: func(cmd *cobra.Command, args []string) error {
			if listExcluded {
				cfg, err := rank.LoadRankConfig()
				if err != nil {
					return err
				}
				if len(cfg.ExcludeProjects) == 0 {
					fmt.Println(output.MutedStyle.Render("No excluded projects."))
					return nil
				}
				fmt.Println(output.HeaderStyle.Render("Excluded Projects"))
				for _, p := range cfg.ExcludeProjects {
					fmt.Printf("  %s %s\n", output.MutedStyle.Render("-"), p)
				}
				return nil
			}

			var projectPath string
			if len(args) > 0 {
				projectPath = args[0]
			} else {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				projectPath = cwd
			}

			absPath, err := filepath.Abs(projectPath)
			if err != nil {
				return err
			}

			if err := rank.AddExclusion(absPath); err != nil {
				return err
			}

			fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("Excluded: %s", absPath)))
			return nil
		},
	}

	cmd.Flags().BoolVar(&listExcluded, "list", false, "list excluded projects")
	return cmd
}

func newRankIncludeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "include [path]",
		Short: "Remove a project from the exclusion list",
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectPath string
			if len(args) > 0 {
				projectPath = args[0]
			} else {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				projectPath = cwd
			}

			absPath, err := filepath.Abs(projectPath)
			if err != nil {
				return err
			}

			if err := rank.RemoveExclusion(absPath); err != nil {
				return err
			}

			fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("Included: %s", absPath)))
			return nil
		},
	}
}

func newRankSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync session transcripts to MoAI Rank",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !rank.IsLoggedIn() {
				return fmt.Errorf("not logged in. Run 'moai rank login' first")
			}

			fmt.Println(output.InfoStyle.Render("Syncing sessions..."))
			fmt.Println(output.MutedStyle.Render("  This feature requires moai-adk Python package."))
			fmt.Println(output.MutedStyle.Render("  Use 'moai rank sync' for full sync support."))
			return nil
		},
	}
}
