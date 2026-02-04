package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/modu-ai/moai-adk-go/internal/rank"
	"github.com/spf13/cobra"
)

var rankCmd = &cobra.Command{
	Use:   "rank",
	Short: "MoAI Rank leaderboard management",
	Long:  "Manage MoAI Rank leaderboard: authenticate, view rankings, sync metrics, and configure exclusions.",
}

func init() {
	rootCmd.AddCommand(rankCmd)

	rankCmd.AddCommand(
		newRankLoginCmd(),
		newRankStatusCmd(),
		newRankLogoutCmd(),
		newRankSyncCmd(),
		newRankExcludeCmd(),
		newRankIncludeCmd(),
		newRankRegisterCmd(),
	)
}

func newRankLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with MoAI Cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if deps == nil || deps.RankCredStore == nil {
				return fmt.Errorf("rank system not initialized")
			}

			// Get or create context
			ctx := cmd.Context()
			if ctx == nil {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), rank.DefaultOAuthTimeout)
				defer cancel()
			}

			// Create OAuth handler with browser opener.
			// Use injected browser if available (for testing), otherwise use real browser.
			browser := deps.RankBrowser
			if browser == nil {
				browser = rank.NewBrowser()
			}
			handler := rank.NewOAuthHandler(rank.OAuthConfig{
				BaseURL: rank.DefaultBaseURL,
				Browser: browser,
			})

			// Start OAuth flow.
			fmt.Fprintln(out, "Opening browser for MoAI Cloud authentication...")
			fmt.Fprintln(out, "Complete authentication in your browser.")

			creds, err := handler.StartOAuthFlow(ctx, rank.DefaultOAuthTimeout)
			if err != nil {
				return fmt.Errorf("oauth flow: %w", err)
			}

			// Store credentials.
			if err := deps.RankCredStore.Save(creds); err != nil {
				return fmt.Errorf("save credentials: %w", err)
			}

			fmt.Fprintf(out, "Authenticated as %s (ID: %s)\n", creds.Username, creds.UserID)
			return nil
		},
	}
}

func newRankStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show ranking status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if deps == nil {
				fmt.Fprintln(out, "Rank client not configured. Run 'moai rank login' first.")
				return nil
			}

			// Lazily initialize Rank client
			if err := deps.EnsureRank(); err != nil {
				fmt.Fprintln(out, "Rank client not configured. Run 'moai rank login' first.")
				return nil
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			userRank, err := deps.RankClient.GetUserRank(ctx)
			if err != nil {
				return fmt.Errorf("get rank: %w", err)
			}

			fmt.Fprintf(out, "User: %s\n", userRank.Username)
			fmt.Fprintf(out, "Total tokens: %d\n", userRank.TotalTokens)
			fmt.Fprintf(out, "Sessions: %d\n", userRank.TotalSessions)
			return nil
		},
	}
}

func newRankLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if deps == nil || deps.RankCredStore == nil {
				return fmt.Errorf("rank system not initialized")
			}

			if err := deps.RankCredStore.Delete(); err != nil {
				return fmt.Errorf("delete credentials: %w", err)
			}

			fmt.Fprintln(out, "Logged out from MoAI Cloud.")
			return nil
		},
	}
}

func newRankSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "sync",
		Short:  "Sync metrics to MoAI Cloud",
		Hidden: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if deps == nil {
				return fmt.Errorf("rank system not initialized")
			}

			// Ensure rank client is initialized
			if err := deps.EnsureRank(); err != nil {
				fmt.Fprintln(out, "Not logged in. Run 'moai rank login' first.")
				return nil
			}

			fmt.Fprintln(out, "Syncing metrics to MoAI Cloud...")

			// TODO: Collect local session/metrics data
			// For now, just verify connection
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			status, err := deps.RankClient.CheckStatus(ctx)
			if err != nil {
				return fmt.Errorf("check rank status: %w", err)
			}

			fmt.Fprintf(out, "Connected to MoAI Cloud (status: %s)\n", status.Status)

			// TODO: Submit session data when available
			// Use SubmitSession() or SubmitSessionsBatch() to send metrics
			fmt.Fprintln(out, "Metrics collection not yet implemented.")
			fmt.Fprintln(out, "Sync complete.")
			return nil
		},
	}
}

func newRankExcludeCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "exclude [pattern]",
		Short:  "Add exclusion pattern for metrics",
		Long:   "Add a glob pattern to exclude from metrics sync. Patterns are stored in ~/.moai/config/rank.yaml.",
		Hidden: false,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			store, err := rank.NewPatternStore("")
			if err != nil {
				return fmt.Errorf("create pattern store: %w", err)
			}

			pattern := args[0]
			if err := store.AddExclude(pattern); err != nil {
				return fmt.Errorf("add exclude pattern: %w", err)
			}

			fmt.Fprintf(out, "Exclusion pattern added: %s\n", pattern)
			return nil
		},
	}
}

func newRankIncludeCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "include [pattern]",
		Short:  "Add inclusion pattern for metrics",
		Long:   "Add a glob pattern to include in metrics sync. Patterns are stored in ~/.moai/config/rank.yaml.",
		Hidden: false,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			store, err := rank.NewPatternStore("")
			if err != nil {
				return fmt.Errorf("create pattern store: %w", err)
			}

			pattern := args[0]
			if err := store.AddInclude(pattern); err != nil {
				return fmt.Errorf("add include pattern: %w", err)
			}

			fmt.Fprintf(out, "Inclusion pattern added: %s\n", pattern)
			return nil
		},
	}
}

func newRankRegisterCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "register [org-name]",
		Short:  "Register organization with MoAI Cloud",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Warning: rank register is experimental and not yet implemented")
			fmt.Fprintf(out, "Organization registration initiated: %s\n", args[0])
			return nil
		},
	}
}
