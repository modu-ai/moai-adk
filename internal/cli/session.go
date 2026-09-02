// @MX:ANCHOR: [AUTO] CLI session subcommand — surface for SPEC-V3R6-MULTI-SESSION-COORD-001 L1 primitive
// @MX:REASON: fan_in via 5 verb subcommands (register, heartbeat, deregister, list, purge). Future orchestrator pre-spawn batch (L4) depends on `moai session list --json` exit semantics. Output stability is contract.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/session"
)

// newSessionCmd builds the `moai session` cobra command tree with 7
// verbs (register, heartbeat, deregister, list, purge, current, doctor).
// The first 5 are the L1 primitives per REQ-COORD-021; `current` and
// `doctor` are the P2 read-path + P1 diagnostic added by
// SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 (REQ-RDP-001, REQ-WPR-001).
//
// IMPORTANT: This CLI surface MUST NOT invoke AskUserQuestion (per
// CONST-V3R5-001..003 + CONST-V3R5-030 boundary). The orchestrator owns
// all user interaction; this CLI returns exit codes + JSON only. Verified
// by internal/session/subagent_boundary_test.go static-grep CI guard.
//
// SPEC-V3R6-MULTI-SESSION-COORD-001 REQ-COORD-021, REQ-COORD-023.
// SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 REQ-RDP-001, REQ-WPR-001.
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
sync check batch.

SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 adds:
  - 'current': resolve this orchestrator's own session UUID (P2 read path)
  - 'doctor':  diagnose why the registry is empty (P1 write-path diagnostic)`,
		GroupID: "tools",
	}

	cmd.AddCommand(newSessionRegisterCmd())
	cmd.AddCommand(newSessionHeartbeatCmd())
	cmd.AddCommand(newSessionDeregisterCmd())
	cmd.AddCommand(newSessionListCmd())
	cmd.AddCommand(newSessionPurgeCmd())
	cmd.AddCommand(newSessionCurrentCmd())
	cmd.AddCommand(newSessionDoctorCmd())
	return cmd
}

func newSessionRegisterCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "register <session_id> <spec_id> <phase>",
		Short: "Register a new active session in the registry",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// SPEC-SESSION-WORKTREE-001 M8: on-touch PR-merge cleanup fires
			// here (REQ-SW-022). Gated by the AutoCleanup toggle inside
			// prMergeCleanup; fail-open, non-blocking.
			prMergeCleanup(loadSessionWorktreeConfig(cmd), cmd.ErrOrStderr())
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
			// SPEC-SESSION-WORKTREE-001 M8: on-touch PR-merge cleanup fires
			// here (REQ-SW-022). Gated by the AutoCleanup toggle inside
			// prMergeCleanup; fail-open, non-blocking.
			prMergeCleanup(loadSessionWorktreeConfig(cmd), cmd.ErrOrStderr())
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

// CanonicalFallbackSessionID is the single canonical environment-fallback
// string for `source_session_id` per SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001
// REQ-FBC-001. When the runtime does not expose session.id to the CLI
// subprocess, `moai session current` emits this string (REQ-RDP-006) and
// the paste-ready resume template (session-handoff.md Block 2) cites it
// verbatim as the graceful-degradation pattern.
//
// This constant is the single anchor; doctrine edits referencing the
// fallback MUST use this string verbatim (REQ-FBC-004 / AC-FBC-005).
const CanonicalFallbackSessionID = "source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>"

// resolveCurrentSessionID resolves this orchestrator's own session UUID.
//
// Stage 0 is the process environment: Claude Code stamps
// CLAUDE_CODE_SESSION_ID into every subprocess it spawns, matching the
// session_id it hands hooks on stdin. That value travels with the process,
// so it names THIS session and no other — it is the authoritative source.
//
// Stage 1 is the side-channel file written by the SessionStart hook. It is a
// single slot per project, unconditionally overwritten on every SessionStart,
// so under two or more concurrent sessions it holds whichever session started
// most recently — possibly a foreign one. It is kept only as the degraded
// path for runtimes that do not stamp the env var (and for a session resumed
// where the env is absent); its answer is not a function of the asking
// session, so it is consulted only when Stage 0 is empty.
//
// Stage 2 is the canonical environment-fallback string (REQ-RDP-006).
//
// Returns (uuid, source, ok). When ok is false, uuid is the canonical
// fallback string and source is "fallback".
//
// SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 REQ-RDP-001/002/003/006.
func resolveCurrentSessionID() (uuid, source string, ok bool) {
	// Stage 0: per-process env var stamped by the Claude Code runtime.
	if id := strings.TrimSpace(os.Getenv(config.EnvClaudeCodeSessionID)); id != "" {
		return id, "env:" + config.EnvClaudeCodeSessionID, true
	}
	// Stage 1: side-channel file written by the SessionStart hook (M3).
	if projectDir := resolveProjectDir(); projectDir != "" {
		sidecar := filepath.Join(projectDir, session.CurrentSideChannelFile)
		if data, err := os.ReadFile(sidecar); err == nil {
			if id := strings.TrimSpace(string(data)); id != "" {
				return id, "side-channel:" + session.CurrentSideChannelFile, true
			}
		}
	}
	// Stage 2: canonical fallback.
	return CanonicalFallbackSessionID, "fallback", false
}

// sessionIDSourceIsAuthoritative reports whether a source string returned by
// resolveCurrentSessionID names the per-process env var. Only that source is
// a function of the asking session; the side-channel file is not, and callers
// that guard against foreign-id attribution key on this.
func sessionIDSourceIsAuthoritative(source string) bool {
	return source == "env:"+config.EnvClaudeCodeSessionID
}

// resolveProjectDir returns the project root using $CLAUDE_PROJECT_DIR
// when set, otherwise the current working directory. The SessionStart hook
// anchors the registry + side-channel file to input.ProjectDir; the CLI
// runs in the same project context so this resolves identically.
func resolveProjectDir() string {
	dir, _ := resolveProjectDirWithSource()
	return dir
}

// resolveProjectDirWithSource reports which tier produced the dir:
// "env:CLAUDE_PROJECT_DIR", "server-cwd", or "unresolved" (t236 / issue
// #1640). The MCP catalog tools surface this as response provenance so a
// fallback resolution — which froze at server spawn — is distinguishable
// from an explicit project_root argument instead of silent.
func resolveProjectDirWithSource() (string, string) {
	if dir := os.Getenv("CLAUDE_PROJECT_DIR"); dir != "" {
		return dir, "env:CLAUDE_PROJECT_DIR"
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd, "server-cwd"
	}
	return "", "unresolved"
}

// newSessionCurrentCmd implements `moai session current` (P2 Stage 1,
// REQ-RDP-001). It resolves and prints this orchestrator's own session
// UUID from the side-channel file (M3) or the canonical fallback string
// (REQ-RDP-006). Exit 0 in both cases; the fallback is graceful
// degradation, not an error.
//
// SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 REQ-RDP-001/002/003/006.
func newSessionCurrentCmd() *cobra.Command {
	var jsonOutput bool
	var showFallback bool
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Print this orchestrator's own session UUID (P2 read path)",
		Long: `Print this orchestrator's own session UUID.

Resolves the UUID from CLAUDE_CODE_SESSION_ID, the per-process variable
Claude Code stamps into every subprocess it spawns. That value travels
with the process, so it names THIS session even when several sessions
share one checkout.

When the variable is absent, the command falls back to the side-channel
file written by the SessionStart hook
(.moai/state/current-session-id.txt). That file is one slot per project,
overwritten on every SessionStart, so with concurrent sessions it holds
whichever started most recently — read it as a degraded answer, not an
authoritative one. With neither source available the command emits the
canonical fallback string instead of an opaque error. The fallback is
graceful degradation, NOT a failure (exit 0).

Use --show-fallback to print ONLY the canonical fallback string (for
paste-ready resume template emission when the orchestrator knows it has
no UUID source).

SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 REQ-RDP-001/002/003/006.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// --show-fallback short-circuits to the canonical string.
			if showFallback {
				if jsonOutput {
					return emitOK(cmd, true, map[string]any{
						"action":             "current",
						"source":             "fallback",
						"session_id":         CanonicalFallbackSessionID,
						"available":          false,
						"canonical_fallback": CanonicalFallbackSessionID,
					})
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), CanonicalFallbackSessionID)
				return nil
			}

			uuid, source, available := resolveCurrentSessionID()
			if jsonOutput {
				return emitOK(cmd, true, map[string]any{
					"action":             "current",
					"source":             source,
					"session_id":         uuid,
					"available":          available,
					"canonical_fallback": CanonicalFallbackSessionID,
				})
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), uuid)
			if !available {
				// Human-readable hint to stderr (non-blocking; exit 0 per
				// REQ-RDP-006 — the fallback is graceful degradation).
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
					"hint: session.id not available from the runtime; emitted canonical fallback. "+
						"Run 'moai session doctor' to diagnose.\n")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON output")
	cmd.Flags().BoolVar(&showFallback, "show-fallback", false, "print ONLY the canonical fallback string (for paste-ready resume emission)")
	return cmd
}

// doctorRootCauses enumerates the candidate root causes for an empty
// registry per SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 REQ-WPR-002 /
// research.md §D.1/D.2/D.3. Order is stable for test stability.
func doctorRootCauses() []string {
	return []string{
		"empty session_id: SessionStart hook received input.SessionID == \"\" (research.md §D.1); the L66 gate bypassed Register. Check ~/.moai/logs/hook-stderr.log for the REQ-WPR-003 warning.",
		"hook wrapper silent-exit: handle-session-start.sh 3-tier fallback (PATH → ~/go/bin/moai → ~/.local/bin/moai) failed to locate the moai binary and exited 0 silently (research.md §D.2). Verify 'command -v moai' or ~/go/bin/moai exists.",
		"registry write failure: reg.Register was called but failed (slog.Warn non-blocking at session_start.go). Check ~/.moai/logs/hook-stderr.log for 'RegisterSession failed'.",
	}
}

// newSessionDoctorCmd implements `moai session doctor` (P1 diagnostic,
// REQ-WPR-001). It reports (a) whether the registry file exists, (b) the
// entry count, (c) the likely root causes when the registry is empty.
//
// SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 REQ-WPR-001/002, AC-WPR-001/002.
func newSessionDoctorCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose why the session registry is empty (P1 write-path diagnostic)",
		Long: `Diagnose why the multi-session coordination registry is empty.

Reports:
  1. Whether .moai/state/active-sessions.json exists
  2. The number of entries currently registered
  3. When the registry is empty on a host with an active session, the
     likely root causes (empty session_id, hook wrapper silent-exit,
     registry write failure)

This is the P1 diagnostic surface for the SessionStart write-path
investigation (SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 §B.1 P1,
REQ-WPR-001/002).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir := resolveProjectDir()
			registryPath := session.DefaultRegistryPath
			if projectDir != "" {
				registryPath = projectDir + string(os.PathSeparator) + registryPath
			}
			_, statErr := os.Stat(registryPath)
			registryExists := statErr == nil

			entries, queryErr := session.QueryActiveWork("")
			entryCount := 0
			if queryErr == nil {
				entryCount = len(entries)
			}

			candidates := doctorRootCauses()
			// When the registry has entries, root causes are informational only.
			// We still emit them so the orchestrator can correlate.

			payload := map[string]any{
				"action":                "doctor",
				"registry_path":         session.DefaultRegistryPath,
				"registry_exists":       registryExists,
				"entry_count":           entryCount,
				"root_cause_candidates": candidates,
				"side_channel_file":     session.CurrentSideChannelFile,
				"session_id_env":        config.EnvClaudeCodeSessionID,
				"session_id_env_set":    os.Getenv(config.EnvClaudeCodeSessionID) != "",
			}
			if queryErr != nil {
				payload["query_error"] = queryErr.Error()
			}

			if jsonOutput {
				return emitOK(cmd, true, payload)
			}

			// Human-readable.
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "registry: %s\n", session.DefaultRegistryPath)
			if registryExists {
				_, _ = fmt.Fprintf(out, "  exists: yes (%d entries)\n", entryCount)
			} else {
				_, _ = fmt.Fprintf(out, "  exists: no (registry file absent)\n")
			}
			_, _ = fmt.Fprintln(out, "root-cause candidates (when registry is empty on an active host):")
			for i, c := range candidates {
				_, _ = fmt.Fprintf(out, "  %d. %s\n", i+1, c)
			}
			if os.Getenv(config.EnvClaudeCodeSessionID) != "" {
				_, _ = fmt.Fprintf(out, "session id: %s is set (authoritative, per-process)\n", config.EnvClaudeCodeSessionID)
			} else {
				_, _ = fmt.Fprintf(out, "session id: %s is NOT set; falling back to the shared side-channel slot\n", config.EnvClaudeCodeSessionID)
			}
			_, _ = fmt.Fprintf(out, "side-channel: %s (degraded fallback; one slot per project, last SessionStart wins)\n", session.CurrentSideChannelFile)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "emit machine-readable JSON output")
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
