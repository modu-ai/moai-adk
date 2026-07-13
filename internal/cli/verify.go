// @MX:NOTE: [AUTO] moai verify CLI — the mechanical surface of the shared
// diagnostic snapshot contract. Doctrine-level consumers (gate / run / sync /
// loop workflow bodies) invoke `moai verify record|check` instead of
// hand-writing snapshot JSON; the stop-goal evaluator consumes the same store
// through internal/verify.Source.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/verify"
)

// verifyKeyTimeout bounds the key computation (3 git subprocesses) on the CLI
// path. The CLI is not the latency-sensitive Stop-hook path, so the bound is
// generous — it exists to keep a wedged git from hanging the command.
const verifyKeyTimeout = 30 * time.Second

// newVerifyCmd builds the `moai verify` cobra command tree: record / check.
// It exposes the internal/verify snapshot engine (schema, key, freshness,
// atomic store) so workflow doctrine invokes the contract mechanically.
//
// This CLI surface MUST NOT invoke AskUserQuestion (subagent boundary): it
// returns exit codes + structured stdout only — exit 0 fresh / exit 1 stale
// on `check`.
func newVerifyCmd() *cobra.Command {
	var projectRoot string

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Record and check shared diagnostic snapshots (verification evidence reuse)",
		Long: `Shared diagnostic snapshot contract.

Quality-check results (tests, lint, coverage, ...) recorded under the current
working-tree key can be reused by sibling verification layers instead of
re-executing, under a strict freshness rule: the stored key must equal the key
recomputed from the current tree AND the recorded timestamp must be within the
TTL (default 10 minutes). A stale snapshot is never citable evidence — on any
mismatch the consumer re-executes the check.

Verbs:
  verify record   record one executed check result under the current-tree key
  verify check    freshness query — exit 0 fresh / exit 1 stale`,
		GroupID:      "tools",
		SilenceUsage: true,
	}
	cmd.PersistentFlags().StringVar(&projectRoot, "project-root", "", "project root (default: $CLAUDE_PROJECT_DIR or cwd)")

	cmd.AddCommand(newVerifyRecordCmd(&projectRoot))
	cmd.AddCommand(newVerifyCheckCmd(&projectRoot))
	return cmd
}

// verifyResolveRoot resolves the project root for snapshot I/O: an explicit
// --project-root wins (absolutized — never joined onto cwd); otherwise the
// shared resolver ($CLAUDE_PROJECT_DIR then cwd).
func verifyResolveRoot(flagValue string) (string, error) {
	if flagValue != "" {
		abs, err := filepath.Abs(flagValue)
		if err != nil {
			return "", fmt.Errorf("verify: resolve --project-root: %w", err)
		}
		return abs, nil
	}
	root := resolveProjectDir()
	if root == "" {
		return "", fmt.Errorf("verify: cannot resolve project root (set CLAUDE_PROJECT_DIR or pass --project-root)")
	}
	return root, nil
}

func newVerifyRecordCmd(projectRoot *string) *cobra.Command {
	var (
		checkID    string
		command    string
		exitCode   int
		durationMS int64
		fromStdin  bool
	)
	cmd := &cobra.Command{
		Use:   "record",
		Short: "Record one executed check result under the current-tree key",
		Long: `Record one executed check result into the snapshot for the current
working-tree key. Flag-driven (--check-id/--command/--exit) or --stdin with a
JSON check entry ({"check_id","command","exit_code","duration_ms","conditions"}).
An entry with the identical command replaces the previous record for this key.`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, _ []string) error {
			root, err := verifyResolveRoot(*projectRoot)
			if err != nil {
				return err
			}
			entry := verify.CheckEntry{
				CheckID:    checkID,
				Command:    command,
				ExitCode:   exitCode,
				DurationMS: durationMS,
			}
			if fromStdin {
				data, readErr := io.ReadAll(c.InOrStdin())
				if readErr != nil {
					return fmt.Errorf("verify record: read stdin: %w", readErr)
				}
				if unmarshalErr := json.Unmarshal(data, &entry); unmarshalErr != nil {
					return fmt.Errorf("verify record: parse stdin JSON: %w", unmarshalErr)
				}
			}
			if entry.Command == "" {
				return fmt.Errorf("verify record: --command is required (or provide it via --stdin JSON)")
			}
			if entry.RecordedAt.IsZero() {
				entry.RecordedAt = time.Now()
			}

			ctx, cancel := context.WithTimeout(c.Context(), verifyKeyTimeout)
			defer cancel()
			key, err := verify.Key(ctx, root)
			if err != nil {
				return fmt.Errorf("verify record: %w", err)
			}
			if _, err := verify.RecordCheck(root, key, entry); err != nil {
				// Fail-open contract lives at the CONSUMER: an unwritable state
				// dir means recording is skipped with an explicit note and the
				// consumer re-executes. The CLI surfaces the error (exit 1) so
				// the note is observable — it never fabricates a record.
				return fmt.Errorf("verify record: %w", err)
			}
			return verifyEmitJSON(c, map[string]any{
				"snapshot":  verify.SnapshotPath(root, key),
				"key":       key,
				"check_id":  entry.CheckID,
				"command":   entry.Command,
				"exit_code": entry.ExitCode,
			})
		},
	}
	cmd.Flags().StringVar(&checkID, "check-id", "", "check identifier (test | lint | format | type | coverage | free-form)")
	cmd.Flags().StringVar(&command, "command", "", "the exact command string that was executed")
	cmd.Flags().IntVar(&exitCode, "exit", 0, "the command's exit code")
	cmd.Flags().Int64Var(&durationMS, "duration-ms", 0, "execution duration in milliseconds")
	cmd.Flags().BoolVar(&fromStdin, "stdin", false, "read the check entry as JSON from stdin")
	return cmd
}

func newVerifyCheckCmd(projectRoot *string) *cobra.Command {
	var (
		keyCurrent bool
		checkID    string
		ttl        time.Duration
	)
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Freshness query for the current-tree snapshot (exit 0 fresh / exit 1 stale)",
		Long: `Query whether a snapshot recorded for the CURRENT working-tree key is
fresh: key equality AND recorded-at within the TTL (default 10 minutes,
configurable via --ttl). Prints snapshot path + key + matched entries for
citation. Exit 0 fresh / exit 1 stale — on stale, re-execute the check instead
of reusing.`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, _ []string) error {
			_ = keyCurrent // current-tree keying is the only supported mode
			root, err := verifyResolveRoot(*projectRoot)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(c.Context(), verifyKeyTimeout)
			defer cancel()
			key, err := verify.Key(ctx, root)
			if err != nil {
				return fmt.Errorf("verify check: %w", err)
			}
			snap, err := verify.Load(root, key)
			if err != nil {
				return fmt.Errorf("verify check: %w", err)
			}
			path := verify.SnapshotPath(root, key)
			stale := func(reason string) error {
				_ = verifyEmitJSON(c, map[string]any{
					"fresh":  false,
					"key":    key,
					"reason": reason,
				})
				return fmt.Errorf("stale: %s", reason)
			}
			if snap == nil {
				return stale("no snapshot recorded for the current working-tree key")
			}
			now := time.Now()
			matched := make([]verify.CheckEntry, 0, len(snap.Checks))
			for _, e := range snap.Checks {
				if checkID != "" && e.CheckID != checkID {
					continue
				}
				if verify.Fresh(snap.Key, key, e.RecordedAt, now, ttl) {
					matched = append(matched, e)
				}
			}
			if len(matched) == 0 {
				if checkID != "" {
					return stale(fmt.Sprintf("no fresh entry for check id %q (miss or TTL expired)", checkID))
				}
				return stale("snapshot exists but no entry is within the TTL")
			}
			return verifyEmitJSON(c, map[string]any{
				"fresh":    true,
				"snapshot": path,
				"key":      key,
				"checks":   matched,
			})
		},
	}
	cmd.Flags().BoolVar(&keyCurrent, "key-current", true, "check freshness against the current working-tree key (the only supported mode)")
	cmd.Flags().StringVar(&checkID, "check", "", "restrict the freshness query to one check id")
	cmd.Flags().DurationVar(&ttl, "ttl", 0, "freshness TTL override (default 10m)")
	return cmd
}

func verifyEmitJSON(c *cobra.Command, payload map[string]any) error {
	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("verify: marshal output: %w", err)
	}
	_, _ = fmt.Fprintln(c.OutOrStdout(), string(out))
	return nil
}

func init() {
	// Cross-file reachability: the verb group must be registered on the root
	// command tree — a defined-but-unregistered verb is inert.
	rootCmd.AddCommand(newVerifyCmd())
}
