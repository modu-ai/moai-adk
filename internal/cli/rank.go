package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/rank"
	"github.com/spf13/cobra"
)

const (
	// rankBatchSize is the number of sessions to submit per batch
	rankBatchSize = 100
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
			_, _ = fmt.Fprintln(out, "Opening browser for MoAI Cloud authentication...")
			_, _ = fmt.Fprintln(out, "Complete authentication in your browser.")

			creds, err := handler.StartOAuthFlow(ctx, rank.DefaultOAuthTimeout)
			if err != nil {
				return fmt.Errorf("oauth flow: %w", err)
			}

			// Set device ID for multi-device tracking
			deviceInfo := rank.GetDeviceInfo()
			creds.DeviceID = deviceInfo.DeviceID

			// Store credentials.
			if err := deps.RankCredStore.Save(creds); err != nil {
				return fmt.Errorf("save credentials: %w", err)
			}

			_, _ = fmt.Fprintln(out, renderSuccessCard(
				fmt.Sprintf("Authenticated as %s", creds.Username),
				fmt.Sprintf("ID: %s", creds.UserID),
			))

			// Install Claude Code hook for session submission
			if err := installRankHook(); err != nil {
				_, _ = fmt.Fprintf(out, "Warning: failed to install session hook: %v\n", err)
				_, _ = fmt.Fprintln(out, "Session metrics will not be submitted automatically.")
			} else {
				_, _ = fmt.Fprintln(out, "Session hook installed successfully.")
			}

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
				_, _ = fmt.Fprintln(out, "Rank client not configured. Run 'moai rank login' first.")
				return nil
			}

			// Lazily initialize Rank client
			if err := deps.EnsureRank(); err != nil {
				_, _ = fmt.Fprintln(out, "Rank client not configured. Run 'moai rank login' first.")
				return nil
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			userRank, err := deps.RankClient.GetUserRank(ctx)
			if err != nil {
				return fmt.Errorf("get rank: %w", err)
			}

			pairs := []kvPair{
				{"User", userRank.Username},
			}
			if userRank.Stats != nil {
				pairs = append(pairs,
					kvPair{"Tokens", fmt.Sprintf("%d", userRank.Stats.TotalTokens)},
					kvPair{"Sessions", fmt.Sprintf("%d", userRank.Stats.TotalSessions)},
				)
			}
			_, _ = fmt.Fprintln(out, renderCard("MoAI Rank", renderKeyValueLines(pairs)))
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

			// Remove Claude Code hook for session submission
			if err := removeRankHook(); err != nil {
				_, _ = fmt.Fprintf(out, "Warning: failed to remove session hook: %v\n", err)
			}

			if err := deps.RankCredStore.Delete(); err != nil {
				return fmt.Errorf("delete credentials: %w", err)
			}

			_, _ = fmt.Fprintln(out, renderSuccessCard("Logged out from MoAI Cloud"))
			return nil
		},
	}
}

func newRankSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
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
				_, _ = fmt.Fprintln(out, "Not logged in. Run 'moai rank login' first.")
				return nil
			}

			// Load sync state for incremental sync
			syncState, err := rank.NewSyncState("")
			if err != nil {
				_, _ = fmt.Fprintf(out, "Warning: could not load sync state: %v\n", err)
				// Continue without sync state (will sync everything)
			}

			// Get device info for multi-device tracking
			deviceInfo := rank.GetDeviceInfo()

			// Check for --force flag
			force, _ := cmd.Flags().GetBool("force")
			if force && syncState != nil {
				syncState.Reset()
			}

			_, _ = fmt.Fprintln(out, "Syncing metrics to MoAI Cloud...")
			_, _ = fmt.Fprintf(out, "Device: %s (%s)\n", deviceInfo.HostName, deviceInfo.DeviceID)

			// Find all transcript files
			transcripts, err := rank.FindTranscripts()
			if err != nil {
				return fmt.Errorf("find transcripts: %w", err)
			}

			if len(transcripts) == 0 {
				_, _ = fmt.Fprintln(out, "No transcripts found. Run Claude Code sessions to generate transcript files for syncing.")
				return nil
			}

			// Parse transcripts and build session submissions
			var sessions []*rank.SessionSubmission
			var sessionTranscriptPaths []string
			parsedCount := 0
			skippedCount := 0
			alreadySyncedCount := 0

			for _, transcriptPath := range transcripts {
				// Skip already-synced transcripts (unless --force)
				if syncState != nil && syncState.IsSynced(transcriptPath) {
					alreadySyncedCount++
					continue
				}

				usage, err := rank.ParseTranscript(transcriptPath)
				if err != nil {
					skippedCount++
					continue
				}

				// Skip sessions with no token usage
				if usage.InputTokens == 0 && usage.OutputTokens == 0 {
					skippedCount++
					continue
				}

				parsedCount++

				// Generate session hash
				sessionHash := rank.ComputeSessionHash(usage.EndedAt, usage.InputTokens, usage.OutputTokens, usage.CacheCreationTokens, usage.CacheReadTokens, usage.ModelName)

				// Extract session ID from filename
				sessionID := filepath.Base(transcriptPath)
				sessionID = strings.TrimSuffix(sessionID, ".jsonl")
				sessionID = strings.TrimSuffix(sessionID, ".json")

				// Anonymize project ID - hash and truncate to 16 chars (server limit)
				projHash := sha256.Sum256([]byte(sessionID))
				anonymousProjectID := hex.EncodeToString(projHash[:])[:16]

				submission := &rank.SessionSubmission{
					SessionHash:         sessionHash,
					EndedAt:             usage.EndedAt,
					InputTokens:         usage.InputTokens,
					OutputTokens:        usage.OutputTokens,
					CacheCreationTokens: usage.CacheCreationTokens,
					CacheReadTokens:     usage.CacheReadTokens,
					AnonymousProjectID:  anonymousProjectID,
					StartedAt:           usage.StartedAt,
					DurationSeconds:     int(usage.DurationSeconds),
					TurnCount:           usage.TurnCount,
					ModelName:           usage.ModelName,
					DeviceID:            deviceInfo.DeviceID,
				}

				sessions = append(sessions, submission)
				sessionTranscriptPaths = append(sessionTranscriptPaths, transcriptPath)
			}

			_, _ = fmt.Fprintf(out, "Found %d transcript(s), %d new, %d already synced\n", len(transcripts), parsedCount, alreadySyncedCount)
			if skippedCount > 0 {
				_, _ = fmt.Fprintf(out, "Skipped %d (parse errors or no token usage)\n", skippedCount)
			}

			if len(sessions) == 0 {
				_, _ = fmt.Fprintln(out, "No new sessions to sync.")
				return nil
			}

			// Submit in batches
			ctx, cancel := context.WithTimeout(cmd.Context(), 300*time.Second)
			defer cancel()

			batchResult := submitSyncBatches(ctx, deps.RankClient, sessions, sessionTranscriptPaths, syncState, out)

			// Save sync state
			if syncState != nil {
				syncState.CleanStale()
				if err := syncState.Save(); err != nil {
					_, _ = fmt.Fprintf(out, "Warning: could not save sync state: %v\n", err)
				}
			}

			succeededTotal := batchResult.Submitted - batchResult.FailedTotal
			_, _ = fmt.Fprintf(out, "Sync complete. Submitted %d session(s), %d succeeded, %d failed, %d errored.\n",
				batchResult.Submitted, succeededTotal, batchResult.FailedTotal, batchResult.ErroredTotal)
			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Force resync all transcripts")

	return cmd
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

			_, _ = fmt.Fprintln(out, renderSuccessCard(fmt.Sprintf("Exclusion pattern added: %s", pattern)))
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

			_, _ = fmt.Fprintln(out, renderSuccessCard(fmt.Sprintf("Inclusion pattern added: %s", pattern)))
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
			_, _ = fmt.Fprintln(out, "Warning: rank register is experimental and not yet implemented")
			_, _ = fmt.Fprintf(out, "Organization registration initiated: %s\n", args[0])
			return nil
		},
	}
}

// syncBatchResult holds the summary counts from a batch sync operation.
type syncBatchResult struct {
	Submitted    int
	FailedTotal  int
	ErroredTotal int
}

// submitSyncBatches submits sessions to the server in batches and tracks sync state.
// It returns aggregate counts of submitted, server-failed, and HTTP-errored sessions.
func submitSyncBatches(ctx context.Context, client rank.Client, sessions []*rank.SessionSubmission, paths []string, syncState *rank.SyncState, out io.Writer) syncBatchResult {
	var res syncBatchResult
	for i := 0; i < len(sessions); i += rankBatchSize {
		end := min(i+rankBatchSize, len(sessions))

		batch := sessions[i:end]
		result, err := client.SubmitSessionsBatch(ctx, batch)
		if err != nil {
			_, _ = fmt.Fprintf(out, "Batch %d-%d failed: %v\n", i, end-1, err)
			res.ErroredTotal += len(batch)
			continue
		}

		res.Submitted += len(batch)
		_, _ = fmt.Fprintf(out, "Submitted %d sessions (batch %d-%d)\n", len(batch), i, end-1)

		// Track actual server-side failures
		if result != nil && result.Failed > 0 {
			res.FailedTotal += result.Failed
			_, _ = fmt.Fprintf(out, "  Failed: %d sessions\n", result.Failed)
		}

		// Only mark synced if ALL sessions in batch succeeded (result.Failed == 0).
		// When any session fails server-side, skip marking so they can be retried
		// on subsequent sync runs without --force.
		// Note: when result.Failed > 0, we skip marking ALL sessions in the batch as
		// synced — even the ones that succeeded — so they can be retried. This relies
		// on the server deduplicating by SessionHash on subsequent sync runs.
		if syncState != nil && (result == nil || result.Failed == 0) {
			batchPaths := paths[i:end]
			for _, transcriptPath := range batchPaths {
				_ = syncState.MarkSynced(transcriptPath)
			}
		}
	}
	return res
}

// --- Claude Code Settings.json Hook Management ---

// claudeSettings represents the structure of ~/.claude/settings.json
type claudeSettings struct {
	Hooks map[string][]hookGroup `json:"hooks,omitempty"`
	// Other fields are omitted as we only modify hooks
}

// hookGroup represents a hook group with optional matcher
type hookGroup struct {
	Matcher string      `json:"matcher,omitempty"`
	Hooks   []hookEntry `json:"hooks"`
}

// hookEntry represents a single hook command
type hookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

// installRankHook installs the rank session hook in ~/.claude/settings.json.
// This enables automatic session metrics submission to MoAI Rank across all projects.
// It deploys a wrapper script to ~/.claude/hooks/moai-rank-session-end.sh (global)
// and configures settings.json to call this script.
func installRankHook() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	// Step 1: Deploy the wrapper script to ~/.claude/hooks/
	if err := deployGlobalRankHookScript(homeDir); err != nil {
		return fmt.Errorf("deploy global hook script: %w", err)
	}

	// Step 2: Update ~/.claude/settings.json with global hook
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
	// Use $HOME environment variable for user-independent path
	hookCommand := `"$HOME/.claude/hooks/rank-submit.sh"`

	// Read existing settings or create new structure
	var settings claudeSettings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read settings.json: %w", err)
		}
		// File doesn't exist, create new structure
		settings = claudeSettings{
			Hooks: make(map[string][]hookGroup),
		}
	} else {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parse settings.json: %w", err)
		}
	}

	// Ensure hooks map exists
	if settings.Hooks == nil {
		settings.Hooks = make(map[string][]hookGroup)
	}

	// Check if SessionEnd hook already exists with our script
	sessionEndHooks := settings.Hooks["SessionEnd"]
	hookExists := false
	for _, group := range sessionEndHooks {
		for _, h := range group.Hooks {
			// Check for either old name (moai-rank-session-end.sh) or new name (rank-submit.sh)
			if strings.Contains(h.Command, "moai-rank-session-end.sh") || strings.Contains(h.Command, "rank-submit.sh") {
				hookExists = true
				break
			}
		}
		if hookExists {
			break
		}
	}

	// Add hook if not exists
	if !hookExists {
		settings.Hooks["SessionEnd"] = append(settings.Hooks["SessionEnd"], hookGroup{
			Matcher: "",
			Hooks: []hookEntry{
				{
					Type:    "command",
					Command: hookCommand,
					Timeout: 10,
				},
			},
		})
	}

	// Write back to file
	newData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return fmt.Errorf("create settings directory: %w", err)
	}

	if err := os.WriteFile(settingsPath, newData, 0o644); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}

	return nil
}

// deployGlobalRankHookScript deploys the rank-submit.sh wrapper script
// to ~/.claude/hooks/. This global hook submits session metrics to MoAI Rank.
func deployGlobalRankHookScript(homeDir string) error {
	// Detect Go bin path for script
	goBinPath := detectGoBinPath(homeDir)

	// Script content - directly generated without template file
	scriptContent := fmt.Sprintf(`#!/bin/bash
# MoAI Rank Session Submission Hook - Generated by moai-adk
# This script forwards stdin JSON to the moai hook session-end command.
# Location: ~/.claude/hooks/rank-submit.sh
# Scope: Global (applies to all projects)
# Action: Submits session metrics to MoAI Rank leaderboard

# Create temp file to store stdin
temp_file=$(mktemp)
trap 'rm -f "$temp_file"' EXIT

# Read stdin into temp file
cat > "$temp_file"

# Try moai command in PATH
if command -v moai &> /dev/null; then
	exec moai hook session-end < "$temp_file"
fi

# Try detected Go bin path from initialization
if [ -f "%s/moai" ]; then
	exec "%s/moai" hook session-end < "$temp_file"
fi

# Try default ~/go/bin/moai
if [ -f "%s/go/bin/moai" ]; then
	exec "%s/go/bin/moai" hook session-end < "$temp_file"
fi

# Not found - exit silently (Claude Code handles missing hooks gracefully)
exit 0
`, goBinPath, goBinPath, homeDir, homeDir)

	// Determine destination path in global hooks directory
	// Use action-based name: rank-submit.sh instead of moai-rank-session-end.sh
	destPath := filepath.Join(homeDir, ".claude", "hooks", "rank-submit.sh")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("create global hooks directory: %w", err)
	}

	// Write the script with executable permissions
	if err := os.WriteFile(destPath, []byte(scriptContent), 0o755); err != nil {
		return fmt.Errorf("write global hook script: %w", err)
	}

	return nil
}

// detectGoBinPath detects the Go binary installation path using `go env`.
func detectGoBinPath(homeDir string) string {
	// Try GOBIN first (explicit override)
	if output, err := exec.Command("go", "env", "GOBIN").Output(); err == nil {
		if goBin := strings.TrimSpace(string(output)); goBin != "" {
			return goBin
		}
	}
	// Try GOPATH/bin (user's Go workspace)
	if output, err := exec.Command("go", "env", "GOPATH").Output(); err == nil {
		if goPath := strings.TrimSpace(string(output)); goPath != "" {
			return filepath.Join(goPath, "bin")
		}
	}
	// Fallback to default ~/go/bin
	if homeDir != "" {
		return filepath.Join(homeDir, "go", "bin")
	}
	return "/usr/local/go/bin"
}

// removeRankHook removes the rank session hook from ~/.claude/settings.json
// and deletes the global wrapper script from ~/.claude/hooks/.
func removeRankHook() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
	hookScriptPath := filepath.Join(homeDir, ".claude", "hooks", "rank-submit.sh")

	// Read existing settings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, just try to remove the script
			_ = os.Remove(hookScriptPath)
			return nil
		}
		return fmt.Errorf("read settings.json: %w", err)
	}

	var settings claudeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("parse settings.json: %w", err)
	}

	// Remove our hook from SessionEnd
	sessionEndHooks := settings.Hooks["SessionEnd"]
	var newHooks []hookGroup

	for _, group := range sessionEndHooks {
		var filteredEntries []hookEntry
		for _, h := range group.Hooks {
			// Filter out hooks that reference our global hook script
			// Check for both old name (moai-rank-session-end.sh) and new name (rank-submit.sh)
			if !strings.Contains(h.Command, "moai-rank-session-end.sh") && !strings.Contains(h.Command, "rank-submit.sh") {
				filteredEntries = append(filteredEntries, h)
			}
		}
		if len(filteredEntries) > 0 {
			newHooks = append(newHooks, hookGroup{
				Matcher: group.Matcher,
				Hooks:   filteredEntries,
			})
		}
	}

	if len(newHooks) == 0 {
		delete(settings.Hooks, "SessionEnd")
	} else {
		settings.Hooks["SessionEnd"] = newHooks
	}

	// Write back to file
	newData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, newData, 0o644); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}

	// Remove the global hook script
	_ = os.Remove(hookScriptPath)

	return nil
}
