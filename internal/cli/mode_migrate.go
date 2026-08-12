package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config/atomicfile"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// modeMigrateCmd implements `moai config mode-migrate` — a dry-run-first,
// operator-approval-gated one-time migration that widens permission bits of
// .moai/config/** files back toward defs.FilePerm.
//
// Default invocation is a dry-run (no-op on disk); --apply widens each
// enumerated candidate through the shared atomic-write helper followed by a
// single os.Chmod(path, defs.FilePerm) site (the AC-MIG-006 scoped exemption —
// the atomic helper preserves the pre-existing destination mode, so widening
// REQUIRES this post-write chmod via the named constant).
//
// SPEC-CONFIG-MODE-MIGRATE-001.
var modeMigrateCmd = &cobra.Command{
	Use:   "mode-migrate",
	Short: "Widen narrowed .moai/config file modes toward defs.FilePerm",
	Long: `Scan .moai/config/** for files whose permission bits are a proper subset of
defs.FilePerm and report them as widening candidates.

By default this is a DRY-RUN: it lists candidates (path + current mode -> target
mode) and modifies nothing. Pass --apply to widen each candidate toward
defs.FilePerm via the shared atomic-write helper.

The migration only widens (never narrows), is scoped to .moai/config/ only, and
is idempotent: a tree already at defs.FilePerm yields an empty candidate list.
Symlinks under .moai/config/** are Lstat-detected and skipped.`,
	RunE: runModeMigrateCmd,
}

// configCmd is the parent for config-surface subcommands. Sibling of the
// existing `moai doctor config` diagnostics tree; this parent owns mutating
// config operations (mode-migrate is the first).
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration-surface operations",
	Long: `Operations on the .moai/config tree.

Subcommands perform mutating config operations (as opposed to ` + "`moai doctor config`" + `
which is read-only diagnostics).`,
}

func init() {
	modeMigrateCmd.Flags().Bool("apply", false,
		"Widen each enumerated candidate toward defs.FilePerm (default is dry-run)")

	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(modeMigrateCmd)
}

// runModeMigrateCmd is the thin cobra handler: it resolves the project's
// .moai/config directory and delegates to runModeMigrate.
func runModeMigrateCmd(cmd *cobra.Command, _ []string) error {
	apply, err := cmd.Flags().GetBool("apply")
	if err != nil {
		return fmt.Errorf("mode-migrate: read --apply flag: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("mode-migrate: resolve working directory: %w", err)
	}
	configDir := filepath.Join(cwd, ".moai", "config")

	return runModeMigrate(cmd.OutOrStdout(), configDir, apply)
}

// Candidate describes a file under .moai/config/** whose permission bits are a
// proper subset of defs.FilePerm and would therefore be widened by --apply.
type Candidate struct {
	// Path is the absolute (or configDir-rooted) filesystem path of the file.
	Path string
	// CurrentMode is the file's current permission bits (os.Lstat .Perm()).
	CurrentMode os.FileMode
	// TargetMode is always defs.FilePerm; carried per-candidate so the dry-run
	// report and the apply path share one source of truth.
	TargetMode os.FileMode
}

// IsWideningCandidate reports whether mode is a proper subset of defs.FilePerm:
// every bit set in mode is also set in defs.FilePerm, AND mode is not already
// equal to defs.FilePerm. This is the spec.md §D.2 Predicate definition applied
// verbatim — widening such a file toward defs.FilePerm only ADDS bits.
//
// A mode carrying a bit NOT present in defs.FilePerm (exec bits, group-write,
// set-uid/gid, sticky) is excluded, because setting it to defs.FilePerm would
// DROP that bit — a narrow, forbidden by REQ-MIG-002.
func IsWideningCandidate(mode os.FileMode) bool {
	cur := mode.Perm()
	target := defs.FilePerm.Perm()
	return (cur|target) == target && cur != target
}

// ScanConfigDir walks dir recursively and returns (a) every regular file whose
// permission bits make it a widening candidate, and (b) the relative paths of
// symlinks encountered (reported as skipped so the operator is aware).
//
// Symlinks are detected via os.Lstat (NOT os.Stat) so the symlink entry itself
// is inspected — it is never followed, and os.Chmod can therefore never land on
// an out-of-scope target (AC-MIG-008 scope-leak closure).
func ScanConfigDir(dir string) (candidates []Candidate, skipped []string, err error) {
	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, werr error) error {
		if werr != nil {
			// Surface walk errors (e.g. unreadable subdir) rather than silently
			// skipping; the operator decides whether to proceed.
			return fmt.Errorf("walk %s: %w", path, werr)
		}
		// filepath.Walk provides Lstat-derived info; detect symlinks explicitly
		// and never follow them (AC-MIG-008).
		if info.Mode()&os.ModeSymlink != 0 {
			rel, rerr := filepath.Rel(dir, path)
			if rerr != nil {
				rel = path
			}
			skipped = append(skipped, rel)
			return nil
		}
		if info.IsDir() {
			return nil
		}
		// Only regular files are candidates (skip sockets, pipes, devices).
		if !info.Mode().IsRegular() {
			return nil
		}
		mode := info.Mode()
		if IsWideningCandidate(mode) {
			candidates = append(candidates, Candidate{
				Path:        path,
				CurrentMode: mode.Perm(),
				TargetMode:  defs.FilePerm,
			})
		}
		return nil
	})
	if walkErr != nil {
		// A missing .moai/config directory is not an error — there is simply
		// nothing to migrate. Surface every other walk failure wrapped.
		if os.IsNotExist(walkErr) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("mode-migrate: scan %s: %w", dir, walkErr)
	}
	return candidates, skipped, nil
}

// FormatDryRun renders the candidate list and skipped-symlink report to w in a
// human-readable form. The footer announces the candidate count and the
// --apply re-run instruction, and explicitly states no files were modified.
func FormatDryRun(w io.Writer, candidates []Candidate, skipped []string) {
	_, _ = fmt.Fprintf(w, "Scanning for files narrower than defs.FilePerm (%04o)...\n\n", defs.FilePerm)

	if len(candidates) == 0 && len(skipped) == 0 {
		_, _ = fmt.Fprintln(w, "0 candidate(s) found. Review the list and re-run with --apply to widen.")
		_, _ = fmt.Fprintln(w, "No files were modified.")
		return
	}

	if len(candidates) > 0 {
		_, _ = fmt.Fprintf(w, "  %-40s %-10s %-10s\n", "PATH", "CURRENT", "TARGET")
		for _, c := range candidates {
			_, _ = fmt.Fprintf(w, "  %-40s %-10s %-10s\n", c.Path, fmt.Sprintf("%04o", c.CurrentMode), fmt.Sprintf("%04o", c.TargetMode))
		}
		_, _ = fmt.Fprintf(w, "\n%d candidate(s) found. Review the list and re-run with --apply to widen.\n", len(candidates))
	} else {
		_, _ = fmt.Fprintf(w, "0 candidate(s) found. Re-run with --apply to widen (no-op).\n")
	}

	for _, s := range skipped {
		_, _ = fmt.Fprintf(w, "  skipped (symlink): %s\n", s)
	}
	_, _ = fmt.Fprintln(w, "No files were modified.")
}

// ApplyWidening widens each candidate toward defs.FilePerm. For each candidate
// it (1) re-reads the file content, (2) writes it back unchanged via the shared
// atomic-write helper (atomic + mode-preserving), and (3) performs the single
// permitted os.Chmod(path, defs.FilePerm) that overrides the helper's
// mode-preservation to actually widen (AC-MIG-006 scoped exemption).
//
// The chmod references the named constant defs.FilePerm — never a numeric
// literal (CLAUDE.local.md §14). Bare os.WriteFile is never used.
func ApplyWidening(candidates []Candidate) error {
	for _, c := range candidates {
		content, err := os.ReadFile(c.Path)
		if err != nil {
			return fmt.Errorf("mode-migrate: read %s: %w", c.Path, err)
		}
		// Route through the shared atomic-write helper (atomic + mode-preserving).
		// defaultMode = defs.FilePerm applies only when the destination does not
		// already exist (it does, so the helper preserves the current mode); the
		// post-write chmod below performs the actual widening.
		if err := atomicfile.Write(c.Path, content, defs.FilePerm); err != nil {
			return fmt.Errorf("mode-migrate: atomic write %s: %w", c.Path, err)
		}
		// The single permitted os.Chmod site (AC-MIG-006): named constant only.
		if err := os.Chmod(c.Path, defs.FilePerm); err != nil {
			return fmt.Errorf("mode-migrate: chmod %s: %w", c.Path, err)
		}
	}
	return nil
}

// runModeMigrate is the testable core: scan configDir, then either print the
// dry-run report (apply == false) or widen each candidate (apply == true).
// Separating it from the cobra handler lets every AC test drive it with a
// t.TempDir()-rooted configDir without spawning a subprocess.
func runModeMigrate(w io.Writer, configDir string, apply bool) error {
	candidates, skipped, err := ScanConfigDir(configDir)
	if err != nil {
		return err
	}

	if !apply {
		FormatDryRun(w, candidates, skipped)
		return nil
	}

	if len(candidates) == 0 {
		// Idempotent no-op: empty candidate list means nothing to widen. No
		// marker file, no output noise beyond a brief confirmation.
		_, _ = fmt.Fprintln(w, "No candidates found. Nothing to widen.")
		return nil
	}

	if err := ApplyWidening(candidates); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "Widened %d file(s) toward defs.FilePerm (%04o):\n", len(candidates), defs.FilePerm)
	for _, c := range candidates {
		_, _ = fmt.Fprintf(w, "  %s  %04o -> %04o\n", c.Path, c.CurrentMode, c.TargetMode)
	}
	for _, s := range skipped {
		_, _ = fmt.Fprintf(w, "  skipped (symlink): %s\n", s)
	}
	return nil
}
