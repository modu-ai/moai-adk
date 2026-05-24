// @MX:ANCHOR: [AUTO] CLI session subcommand — surface for SPEC-V3R6-MULTI-SESSION-COORD-001 L1 primitive
// @MX:REASON: fan_in via 5 verb subcommands (register, heartbeat, deregister, list, purge). Future orchestrator pre-spawn batch (L4) depends on `moai session list --json` exit semantics. Output stability is contract.
package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/session"
)

// newSessionCmd builds the `moai session` cobra command tree with 5
// verbs (register, heartbeat, deregister, list, purge) per REQ-COORD-021.
//
// IMPORTANT: This CLI surface MUST NOT invoke AskUserQuestion (per
// CONST-V3R5-001..003 + CONST-V3R5-030 boundary). The orchestrator owns
// all user interaction; this CLI returns exit codes + JSON only. Verified
// by internal/session/subagent_boundary_test.go static-grep CI guard.
//
// SPEC-V3R6-MULTI-SESSION-COORD-001 REQ-COORD-021, REQ-COORD-023.
func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage active-sessions coordination registry (multi-session race mitigation)",
		Long: `Manage the multi-session coordination registry at .moai/state/active-sessions.json.

The registry tracks active Claude Code sessions on this host so the
orchestrator can detect race conditions before spawning implementation
agents. Each session registers itself at SessionStart, heartbeats during
work, and deregisters at SessionEnd. Stale entries (>30 minutes) are
purged automatically.

This subcommand is the user-facing CLI for the L1 primitive of
SPEC-V3R6-MULTI-SESSION-COORD-001. The orchestrator consumes 'session
list --json --filter-spec=<SPEC-ID>' as the 3rd command in the pre-spawn
sync check batch.`,
		GroupID: "tools",
	}

	cmd.AddCommand(newSessionRegisterCmd())
	cmd.AddCommand(newSessionHeartbeatCmd())
	cmd.AddCommand(newSessionDeregisterCmd())
	cmd.AddCommand(newSessionListCmd())
	cmd.AddCommand(newSessionPurgeCmd())
	return cmd
}

func newSessionRegisterCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "register <session_id> <spec_id> <phase>",
		Short: "Register a new active session in the registry",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID, specID, phase := args[0], args[1], args[2]
			if err := session.RegisterSession(sessionID, specID, phase); err != nil {
				return fmt.Errorf("register: %w", err)
			}
			return emitOK(cmd, jsonOutput, map[string]any{
				"action":     "register",
				"session_id": sessionID,
				"spec_id":    specID,
				"phase":      phase,
			})
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON output")
	return cmd
}

func newSessionHeartbeatCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "heartbeat <session_id>",
		Short: "Update last_heartbeat for an existing session (idempotent)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			if err := session.Heartbeat(sessionID); err != nil {
				return fmt.Errorf("heartbeat: %w", err)
			}
			return emitOK(cmd, jsonOutput, map[string]any{
				"action":     "heartbeat",
				"session_id": sessionID,
			})
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON output")
	return cmd
}

func newSessionDeregisterCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "deregister <session_id>",
		Short: "Remove a session from the registry (idempotent)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			if err := session.DeregisterSession(sessionID); err != nil {
				return fmt.Errorf("deregister: %w", err)
			}
			return emitOK(cmd, jsonOutput, map[string]any{
				"action":     "deregister",
				"session_id": sessionID,
			})
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON output")
	return cmd
}

func newSessionListCmd() *cobra.Command {
	var jsonOutput bool
	var filterSpec string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active sessions (optionally filtered by --filter-spec)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := session.QueryActiveWork(filterSpec)
			if err != nil {
				return fmt.Errorf("list: %w", err)
			}
			if jsonOutput {
				out, err := json.MarshalIndent(entries, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(out))
				return nil
			}
			if len(entries) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "(no active sessions)")
				return nil
			}
			for _, e := range entries {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(),
					"session=%s spec=%s phase=%s started=%s last_hb=%s pid=%d host=%s\n",
					shortID(e.SessionID), e.SpecID, e.Phase,
					e.StartedAt.Format("2006-01-02T15:04:05Z"),
					e.LastHeartbeat.Format("2006-01-02T15:04:05Z"),
					e.PID, e.Host,
				)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON output (orchestrator pre-spawn check format)")
	cmd.Flags().StringVar(&filterSpec, "filter-spec", "", "only return entries matching this spec_id")
	return cmd
}

func newSessionPurgeCmd() *cobra.Command {
	var jsonOutput bool
	var thresholdMinutes int
	cmd := &cobra.Command{
		Use:   "purge",
		Short: "Remove stale entries (default >30 minutes since last heartbeat)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			purged, err := session.PurgeStale(thresholdMinutes)
			if err != nil {
				return fmt.Errorf("purge: %w", err)
			}
			return emitOK(cmd, jsonOutput, map[string]any{
				"action":            "purge",
				"threshold_minutes": thresholdMinutes,
				"purged_count":      purged,
			})
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON output")
	cmd.Flags().IntVar(&thresholdMinutes, "threshold-minutes", session.DefaultStaleMinutes, "stale heartbeat cutoff in minutes")
	return cmd
}

// emitOK writes either human-readable or JSON success output to stdout.
// All session subcommands route through this helper so JSON parsability
// is uniform.
func emitOK(cmd *cobra.Command, jsonOutput bool, payload map[string]any) error {
	if jsonOutput {
		out, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return nil
	}
	parts := make([]string, 0, len(payload))
	for k, v := range payload {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "OK "+strings.Join(parts, " "))
	return nil
}

// shortID returns the first 8 characters of a UUID for compact display.
func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

func init() {
	// SPEC-V3R6-MULTI-SESSION-COORD-001 REQ-COORD-021: register session subcommand
	rootCmd.AddCommand(newSessionCmd())
}
