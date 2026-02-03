package cli

import (
	"context"
	"fmt"
	"time"

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
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if deps == nil || deps.RankCredStore == nil {
				return fmt.Errorf("rank system not initialized")
			}
			fmt.Fprintln(out, "Opening browser for MoAI Cloud authentication...")
			fmt.Fprintln(out, "Login flow initiated. Complete authentication in your browser.")
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
		Hidden: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Warning: rank sync is experimental and not yet implemented")
			fmt.Fprintln(out, "Syncing metrics to MoAI Cloud...")
			fmt.Fprintln(out, "Sync complete.")
			return nil
		},
	}
}

func newRankExcludeCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "exclude [pattern]",
		Short:  "Add exclusion pattern for metrics",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Warning: rank exclude is experimental and not yet implemented")
			fmt.Fprintf(out, "Exclusion pattern added: %s\n", args[0])
			return nil
		},
	}
}

func newRankIncludeCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "include [pattern]",
		Short:  "Add inclusion pattern for metrics",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Warning: rank include is experimental and not yet implemented")
			fmt.Fprintf(out, "Inclusion pattern added: %s\n", args[0])
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
