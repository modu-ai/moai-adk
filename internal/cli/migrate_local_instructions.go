package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// localInstructionsBackupDir is where the migration verb keeps the original
// CLAUDE.local.md, relative to the project root. One timestamped directory per
// migration; the verb never writes a second backup for the same file.
const localInstructionsBackupDir = ".moai/backups/local-instructions"

func init() { migrateCmd.AddCommand(newMigrateLocalInstructionsCmd(findProjectRoot)) }

// newMigrateLocalInstructionsCmd builds `moai migrate local-instructions`, the
// only path that moves CLAUDE.local.md to AGENTS.local.md. Nothing else —
// update, init, doctor, the Codex launcher, or a hook — performs the rename.
func newMigrateLocalInstructionsCmd(rootFn func() (string, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local-instructions",
		Short: "Move CLAUDE.local.md to AGENTS.local.md, keeping a backup",
		Long: "Moves the legacy CLAUDE.local.md to AGENTS.local.md, the local instruction file\n" +
			"both Claude Code and Codex read, and keeps the original under " + localInstructionsBackupDir + ".\n" +
			"Refuses without changing anything when both files exist.",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		root, err := rootFn()
		if err != nil {
			return fmt.Errorf("resolve project root: %w", err)
		}
		return migrateLocalInstructions(root, cmd.OutOrStdout(), time.Now())
	}
	return cmd
}

// @MX:WARN: [AUTO] Moves a user-authored file; every exit before the final remove leaves the original in place.
// @MX:REASON: The file may carry content found nowhere else, so a partial move must never cost it.
// migrateLocalInstructions performs the move. The refusal branch comes first:
// with both files present the verb cannot know which content the user means to
// keep, so it modifies neither (REQ-IFU-010b).
func migrateLocalInstructions(root string, out io.Writer, now time.Time) error {
	agentsPath := filepath.Join(root, codexLocalInstructionName)
	claudePath := filepath.Join(root, codexClaudeLocalName)
	agentsPresent, claudePresent, err := localInstructionFilesPresent(root)
	if err != nil {
		return err
	}
	switch {
	case agentsPresent && claudePresent:
		return fmt.Errorf("refusing to migrate: %s and %s both exist; merge %s into %s by hand, then remove %s",
			codexLocalInstructionName, codexClaudeLocalName, codexClaudeLocalName, codexLocalInstructionName, codexClaudeLocalName)
	case !claudePresent && agentsPresent:
		_, _ = fmt.Fprintf(out, "Nothing to migrate: %s is already the local instruction file.\n", codexLocalInstructionName)
		return nil
	case !claudePresent:
		_, _ = fmt.Fprintf(out, "Nothing to migrate: no %s in %s.\n", codexClaudeLocalName, root)
		return nil
	}

	info, err := os.Lstat(claudePath)
	if err != nil {
		return fmt.Errorf("stat %s: %w", codexClaudeLocalName, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("refusing to migrate: %s is not a regular file (%s); move its content by hand",
			codexClaudeLocalName, info.Mode().Type())
	}
	body, err := os.ReadFile(claudePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", codexClaudeLocalName, err)
	}
	perm := info.Mode().Perm()

	backupRel, err := writeLocalInstructionsBackup(root, body, perm, now)
	if err != nil {
		return err
	}
	if err := writeExclusive(agentsPath, body, perm); err != nil {
		return fmt.Errorf("write %s: %w", codexLocalInstructionName, err)
	}
	if err := os.Remove(claudePath); err != nil {
		// Leaving both files would recreate the coexistence the verb exists to end.
		if rbErr := os.Remove(agentsPath); rbErr != nil {
			return fmt.Errorf("remove %s: %w (rollback of %s also failed: %v)", codexClaudeLocalName, err, codexLocalInstructionName, rbErr)
		}
		return fmt.Errorf("remove %s: %w", codexClaudeLocalName, err)
	}

	_, _ = fmt.Fprintf(out, "Migrated %s to %s (backup: %s).\n", codexClaudeLocalName, codexLocalInstructionName, backupRel)
	_, _ = fmt.Fprintf(out, "Claude Code reads %s through the @%s import in CLAUDE.md; Codex reads it through `moai codex`.\n",
		codexLocalInstructionName, codexLocalInstructionName)
	return nil
}

// writeLocalInstructionsBackup copies body into a fresh timestamped directory
// and returns the copy's project-relative path.
func writeLocalInstructionsBackup(root string, body []byte, perm os.FileMode, now time.Time) (string, error) {
	base := filepath.Join(root, filepath.FromSlash(localInstructionsBackupDir))
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}
	dir, err := os.MkdirTemp(base, now.UTC().Format("20060102T150405Z")+"-")
	if err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}
	path := filepath.Join(dir, codexClaudeLocalName)
	if err := writeExclusive(path, body, perm); err != nil {
		return "", fmt.Errorf("write backup: %w", err)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path, nil
	}
	return filepath.ToSlash(rel), nil
}

// writeExclusive creates path (failing if it exists), writes body, and syncs.
// A failed write removes the partial file.
func writeExclusive(path string, body []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}
	_, werr := f.Write(body)
	if werr == nil {
		werr = f.Sync()
	}
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		_ = os.Remove(path)
	}
	return werr
}

// localInstructionFilesPresent reports which of the two local instruction files
// exist at root. Any directory entry counts, links included: the question is
// whether a second live read path could exist, not whether it is readable.
func localInstructionFilesPresent(root string) (agents, claude bool, err error) {
	present := func(name string) (bool, error) {
		if _, err := os.Lstat(filepath.Join(root, name)); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return false, nil
			}
			return false, fmt.Errorf("stat %s: %w", name, err)
		}
		return true, nil
	}
	if agents, err = present(codexLocalInstructionName); err != nil {
		return false, false, err
	}
	if claude, err = present(codexClaudeLocalName); err != nil {
		return false, false, err
	}
	return agents, claude, nil
}

// @MX:ANCHOR: [AUTO] One advisory text for the three surfaces that report a legacy local instruction file.
// @MX:REASON: update, doctor, and the Codex launcher must say the same thing (REQ-IFU-012); fan_in 3.
// localInstructionsAdvisory returns the operator advisory for the given file
// set, or "" when CLAUDE.local.md is absent. Both advising branches name the
// migration verb.
func localInstructionsAdvisory(agentsPresent, claudePresent bool) string {
	switch {
	case !claudePresent:
		return ""
	case agentsPresent:
		return fmt.Sprintf("Advisory: %s and %s both exist; merge %s into %s by hand and remove it "+
			"(`moai migrate local-instructions` refuses while both exist).",
			codexLocalInstructionName, codexClaudeLocalName, codexClaudeLocalName, codexLocalInstructionName)
	default:
		return fmt.Sprintf("Advisory: %s is a legacy local instruction file; "+
			"run `moai migrate local-instructions` to move it to %s.",
			codexClaudeLocalName, codexLocalInstructionName)
	}
}

// emitLocalInstructionsAdvisory writes the advisory for root to w, if any. It
// only reads: neither file is moved, renamed, or rewritten (REQ-IFU-011).
func emitLocalInstructionsAdvisory(w io.Writer, root string) {
	agents, claude, err := localInstructionFilesPresent(root)
	if err != nil {
		return
	}
	if msg := localInstructionsAdvisory(agents, claude); msg != "" {
		_, _ = fmt.Fprintln(w, msg)
	}
}
