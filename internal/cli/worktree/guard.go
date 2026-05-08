package worktree

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	wtg "github.com/modu-ai/moai-adk/internal/worktree"
)

// @MX:NOTE: `moai worktree snapshot|verify|restore` are the orchestrator-facing
// CLI primitives for SPEC-V3R3-CI-AUTONOMY-001 Wave 5. The orchestrator (Claude
// Code runtime) invokes these via Bash before/after each Agent(isolation:) call.

// ExitCodeError carries a structured exit code so cmd/moai/main.go can propagate
// 0/1/2/3 instead of cobra's default 1-on-any-error behavior. Bit 1 = divergence
// detected, bit 2 = suspect (empty worktreePath).
type ExitCodeError struct {
	Code int
}

func (e *ExitCodeError) Error() string { return fmt.Sprintf("worktree guard exit code %d", e.Code) }

// ExitCode satisfies the ExitCoder interface in cmd/moai/main.go.
func (e *ExitCodeError) ExitCode() int { return e.Code }

// VerifyResult is the JSON structure printed to stdout from the verify subcommand.
type VerifyResult struct {
	SnapshotID  string           `json:"snapshot_id"`
	Divergence  wtg.Divergence   `json:"divergence"`
	SuspectFlag *wtg.SuspectFlag `json:"suspect_flag,omitempty"`
	ReportPath  string           `json:"report_path,omitempty"`
	JSONSidecar string           `json:"json_sidecar,omitempty"`
	ExitCode    int              `json:"exit_code"`
}

// projectRoot resolves the project root via `git rev-parse --show-toplevel`.
func projectRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("project root: %w (run from inside a git repo)", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// newGuardSnapshotCmd creates the `moai worktree snapshot` subcommand.
func newGuardSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Capture working tree state snapshot for guard verification",
		Long: "Capture a Snapshot of working tree state (HEAD, branch, porcelain, " +
			"untracked under .moai/specs/) and write JSON to .moai/state/. " +
			"Used by the orchestrator before each Agent(isolation: \"worktree\") call.",
		RunE: runGuardSnapshot,
	}
	cmd.Flags().String("out", "", "Output snapshot path (default: .moai/state/worktree-snapshot-<id>.json)")
	cmd.Flags().String("agent-name", "", "Agent name (informational, recorded for downstream verify)")
	return cmd
}

func runGuardSnapshot(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()
	root, err := projectRoot()
	if err != nil {
		return err
	}
	outPath, _ := cmd.Flags().GetString("out")

	ctx := context.Background()
	snap, err := wtg.Capture(ctx, wtg.CaptureOptions{RepoDir: root})
	if err != nil {
		return fmt.Errorf("capture snapshot: %w", err)
	}
	if outPath == "" {
		outPath = wtg.SnapshotPath(root, snap.SnapshotID)
	} else if !filepath.IsAbs(outPath) {
		outPath = filepath.Join(root, outPath)
	}
	if err := wtg.SaveSnapshot(snap, outPath); err != nil {
		return err
	}
	payload := map[string]string{
		"snapshot_id": snap.SnapshotID,
		"path":        outPath,
		"head_sha":    snap.HeadSHA,
		"branch":      snap.Branch,
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

// newGuardVerifyCmd creates the `moai worktree verify` subcommand.
func newGuardVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify working tree state against snapshot + check agent response",
		Long: "Compare current working tree against a previously-captured snapshot " +
			"and optionally inspect an agent response JSON for empty worktreePath. " +
			"Exit codes: 0=clean, 1=divergence, 2=suspect (empty worktreePath), 3=both.",
		RunE: runGuardVerify,
	}
	cmd.Flags().String("snapshot", "", "Path to pre-state snapshot JSON (required)")
	cmd.Flags().String("agent-response", "", "Path to agent response JSON (optional, for suspect detection)")
	cmd.Flags().String("agent-name", "", "Agent name to record in divergence/suspect logs")
	_ = cmd.MarkFlagRequired("snapshot")
	return cmd
}

func runGuardVerify(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()
	root, err := projectRoot()
	if err != nil {
		return err
	}
	snapPath, _ := cmd.Flags().GetString("snapshot")
	agentResp, _ := cmd.Flags().GetString("agent-response")
	agentName, _ := cmd.Flags().GetString("agent-name")
	if !filepath.IsAbs(snapPath) {
		snapPath = filepath.Join(root, snapPath)
	}

	pre, err := wtg.LoadSnapshot(snapPath)
	if err != nil {
		return err
	}
	ctx := context.Background()
	post, err := wtg.Capture(ctx, wtg.CaptureOptions{RepoDir: root, SnapshotID: pre.SnapshotID + "-post"})
	if err != nil {
		return fmt.Errorf("capture post-state: %w", err)
	}
	div := wtg.Diff(pre, post)
	result := VerifyResult{SnapshotID: pre.SnapshotID, Divergence: div}

	var suspect *wtg.SuspectFlag
	if agentResp != "" {
		respPath := agentResp
		if !filepath.IsAbs(respPath) {
			respPath = filepath.Join(root, respPath)
		}
		empty, err := agentResponseHasEmptyWorktreePath(respPath)
		if err != nil {
			return fmt.Errorf("inspect agent response: %w", err)
		}
		if empty {
			f := wtg.SuspectFlag{
				SnapshotID:  pre.SnapshotID,
				AgentName:   agentName,
				Reason:      wtg.SuspectReasonEmptyWorktreePath,
				DetectedAt:  time.Now().UTC(),
				PushBlocked: true,
			}
			flagPath, err := wtg.WriteSuspectFlag(root, f)
			if err != nil {
				return err
			}
			suspect = &f
			result.SuspectFlag = suspect
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "suspect flag written: %s\n", flagPath)
		}
	}

	exitCode := 0
	if div.IsDivergent() {
		exitCode |= 1
	}
	if suspect != nil {
		exitCode |= 2
	}
	result.ExitCode = exitCode

	if div.IsDivergent() || suspect != nil {
		entry := wtg.DivergenceLogEntry{
			Timestamp:   time.Now().UTC(),
			SnapshotID:  pre.SnapshotID,
			AgentName:   agentName,
			Divergence:  div,
			SuspectFlag: suspect,
		}
		mdPath, jsonPath, err := wtg.AppendDivergenceLog(root, entry)
		if err != nil {
			return err
		}
		result.ReportPath = mdPath
		result.JSONSidecar = jsonPath
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		return err
	}
	if exitCode != 0 {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		return &ExitCodeError{Code: exitCode}
	}
	return nil
}

// agentResponseHasEmptyWorktreePath reports whether the agent response JSON has
// an empty worktreePath (missing field, {}, "", or null).
func agentResponseHasEmptyWorktreePath(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return false, fmt.Errorf("parse %s: %w", path, err)
	}
	val, ok := raw["worktreePath"]
	if !ok {
		return true, nil
	}
	s := strings.TrimSpace(string(val))
	switch s {
	case "{}", `""`, "null":
		return true, nil
	}
	return false, nil
}

// newGuardRestoreCmd creates the `moai worktree restore` subcommand.
func newGuardRestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore working tree to a snapshot's HEAD state",
		Long: "Run git restore --source=<snapshot HEAD> --staged --worktree :/. " +
			"Untracked files cannot be recreated from git; this command lists their " +
			"snapshot paths and instructs the user to recreate them manually.",
		RunE: runGuardRestore,
	}
	cmd.Flags().String("snapshot", "", "Path to snapshot JSON (required)")
	cmd.Flags().Bool("dry-run", false, "Print the git command without executing")
	_ = cmd.MarkFlagRequired("snapshot")
	return cmd
}

func runGuardRestore(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()
	root, err := projectRoot()
	if err != nil {
		return err
	}
	snapPath, _ := cmd.Flags().GetString("snapshot")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if !filepath.IsAbs(snapPath) {
		snapPath = filepath.Join(root, snapPath)
	}
	snap, err := wtg.LoadSnapshot(snapPath)
	if err != nil {
		return err
	}
	if snap.HeadSHA == "" {
		return fmt.Errorf("snapshot %s has empty HeadSHA — cannot restore", snapPath)
	}
	gitArgs := []string{"restore", "--source=" + snap.HeadSHA, "--staged", "--worktree", ":/"}
	short := snap.HeadSHA
	if len(short) > 8 {
		short = short[:8]
	}
	_, _ = fmt.Fprintf(out, "WARNING: this will discard local changes against tracked files at HEAD %s\n", short)
	_, _ = fmt.Fprintf(out, "command: git %s\n", strings.Join(gitArgs, " "))
	if len(snap.UntrackedSpecs) > 0 {
		_, _ = fmt.Fprintln(out, "untracked files in snapshot (manual recreation required — git restore cannot recover untracked content):")
		for _, p := range snap.UntrackedSpecs {
			_, _ = fmt.Fprintf(out, "  - %s\n", p)
		}
	}
	if dryRun {
		_, _ = fmt.Fprintln(out, "dry-run: command not executed")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), wtg.DefaultGitTimeout)
	defer cancel()
	gc := exec.CommandContext(ctx, "git", gitArgs...)
	gc.Dir = root
	if combined, gerr := gc.CombinedOutput(); gerr != nil {
		return fmt.Errorf("git restore: %w\n%s", gerr, combined)
	}
	_, _ = fmt.Fprintln(out, "restore complete")
	return nil
}
