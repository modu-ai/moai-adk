package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// localInstructionsBackupDir is where the migration verb keeps its copies and
// the renamed originals, relative to the project root. Every entry lives in
// its own fresh timestamped directory; the verb never writes a second backup
// for a completed migration, and never touches existing entries
// (REQ-CIF-015: historical backups are immutable).
const localInstructionsBackupDir = ".moai/backups/local-instructions"

// Test seams for the M2 lossless fixtures (AC-CIF-010). Production defaults
// carry the production behavior; tests override them to inject the read->write
// window edit, a refused rename, and a failed destination write.
var (
	migrateGuardHook func()                      // called between the read and the destination write
	migrateRename    = os.Rename                 // the final atomic step and the preserve rename
	migrateWriteDest = writeExclusiveKeepPartial // the destination's exclusive write
)

// instructionPair is one legacy->canonical migration path. Both pairs are
// mandatory (REQ-CIF-005): the shared file and the local file each migrate
// under the same semantics.
type instructionPair struct {
	legacy    string
	canonical string
}

// instructionMigrationPairs is the pair table the verb drives.
var instructionMigrationPairs = []instructionPair{
	{legacy: legacyLocalInstructionName, canonical: canonicalLocalInstructionName},
	{legacy: "AGENTS.md", canonical: "CLAUDE.md"},
}

// Instruction-file name constants (SPEC-CLAUDE-INSTRUCTION-FILES-001):
// CLAUDE.local.md is the canonical local instruction file and AGENTS.local.md
// its legacy predecessor, moved only by the migration verb.
const (
	canonicalLocalInstructionName = "CLAUDE.local.md"
	legacyLocalInstructionName    = "AGENTS.local.md"
)

func init() { migrateCmd.AddCommand(newMigrateLocalInstructionsCmd(findProjectRoot)) }

// newMigrateLocalInstructionsCmd builds `moai migrate local-instructions`, the
// only path that moves AGENTS.md/AGENTS.local.md to CLAUDE.md/CLAUDE.local.md.
// Nothing else — update, init, doctor, a launcher, or a hook — performs the
// move.
func newMigrateLocalInstructionsCmd(rootFn func() (string, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local-instructions",
		Short: "Move AGENTS.md/AGENTS.local.md to CLAUDE.md/CLAUDE.local.md, keeping backups",
		Long: "Moves the legacy AGENTS.md and AGENTS.local.md to the canonical CLAUDE.md and\n" +
			"CLAUDE.local.md, keeping byte-exact backups under " + localInstructionsBackupDir + ".\n" +
			"Refuses without changing anything when both files of a pair exist.",
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

// @MX:WARN: [AUTO] Moves user-authored files; every exit leaves the user's latest content at the canonical path, in a backup, or at the untouched legacy path.
// @MX:REASON: The files may carry content found nowhere else, so no exit of the verb may cost it (REQ-CIF-010 lossless invariant).
// migrateLocalInstructions decides for every pair first — any refusal means
// nothing is written anywhere — then migrates each legacy-only pair: a
// byte-exact backup copy before the destination write, a read->write change
// guard, and a final single atomic os.Rename of the original into a fresh
// backup path (whatever bytes exist at that instant land in the backup by
// construction; there is no re-check-then-remove window).
func migrateLocalInstructions(root string, out io.Writer, now time.Time) error {
	// Decision phase: validate every pair before touching anything.
	type plan struct {
		pair  instructionPair
		body  []byte
		perm  os.FileMode
		mtime time.Time
	}
	var plans []plan
	for _, pair := range instructionMigrationPairs {
		legacyPath := filepath.Join(root, pair.legacy)
		canonPath := filepath.Join(root, pair.canonical)
		legacyPresent, canonicalPresent, err := instructionPairPresent(legacyPath, canonPath)
		if err != nil {
			return err
		}
		switch {
		case legacyPresent && canonicalPresent:
			return fmt.Errorf("refusing to migrate: %s and %s both exist; merge %s into %s by hand, then remove %s",
				pair.legacy, pair.canonical, pair.legacy, pair.canonical, pair.legacy)
		case !legacyPresent:
			continue
		}
		// The verb takes no path argument — the fixed names cannot escape the
		// root, and a non-regular entry (symlink included, whatever it points
		// at) is refused rather than followed (REQ-CIF-009).
		info, err := os.Lstat(legacyPath)
		if err != nil {
			return fmt.Errorf("stat %s: %w", pair.legacy, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("refusing to migrate: %s is not a regular file (%s); move its content by hand",
				pair.legacy, info.Mode().Type())
		}
		body, err := os.ReadFile(legacyPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", pair.legacy, err)
		}
		plans = append(plans, plan{pair: pair, body: body, perm: info.Mode().Perm(), mtime: info.ModTime()})
	}
	if len(plans) == 0 {
		_, _ = fmt.Fprintf(out, "Nothing to migrate: no legacy AGENTS instruction files in %s.\n", root)
		return nil
	}

	for _, p := range plans {
		if err := migrateOnePair(root, p.pair, p.body, p.perm, p.mtime, now, out); err != nil {
			return err
		}
	}
	return nil
}

// migrateOnePair performs one pair's migration. Every failure exit preserves
// the user's latest content and reports the exact partial outcome.
func migrateOnePair(root string, pair instructionPair, body []byte, perm os.FileMode, mtime time.Time, now time.Time, out io.Writer) error {
	legacyPath := filepath.Join(root, pair.legacy)
	canonPath := filepath.Join(root, pair.canonical)

	// REQ-CIF-006: a byte-exact, permission-preserving backup copy under a
	// fresh timestamped directory, BEFORE the destination write.
	copyRel, err := writeLocalInstructionsBackup(root, pair.legacy, body, perm, now)
	if err != nil {
		return err
	}

	// REQ-CIF-010 front guard: abort before the destination write when the
	// source's content or mtime changed since the read; the changed original
	// stays untouched at the legacy path.
	if migrateGuardHook != nil {
		migrateGuardHook()
	}
	if changed, cerr := sourceChangedSinceRead(legacyPath, body, mtime); cerr != nil {
		return fmt.Errorf("re-check %s: %w", pair.legacy, cerr)
	} else if changed {
		return fmt.Errorf("refusing to migrate: %s changed between the read and the write; aborting before the destination write (the file is untouched)", pair.legacy)
	}

	// Destination write. The partial file is kept on failure so the caller can
	// preserve it (the plain writeExclusive removes it — that behavior stays
	// for backup copies, where nothing else exists to preserve).
	werr := migrateWriteDest(canonPath, body, perm)
	if werr != nil {
		// N12 disambiguation: EEXIST means another process created the
		// destination after the presence check raced. That file is NOT "the
		// partial destination" — the O_EXCL creation in THIS process is the
		// ownership proof. Leave both files in place and refuse; a foreign
		// file is never renamed into the backup.
		if errors.Is(werr, os.ErrExist) {
			return fmt.Errorf("refusing to migrate: %s already exists (created by another process) while %s is present; merge %s into %s by hand, then remove %s",
				pair.canonical, pair.legacy, pair.legacy, pair.canonical, pair.legacy)
		}
		// Any other write failure: preserve the partial destination —
		// unconditionally, in whatever state it is (REQ-CIF-011; no
		// edited-vs-untouched classification exists).
		if _, serr := os.Lstat(canonPath); serr == nil {
			preserved, perr := preserveIntoBackup(root, pair.canonical, canonPath, now)
			if perr != nil {
				return residualReport(pair, legacyPath, canonPath, fmt.Sprintf("write %s failed: %v; preserving the partial destination failed: %v", pair.canonical, werr, perr))
			}
			return fmt.Errorf("write %s failed: %w; the partial destination was preserved at %s, and %s is unchanged — no migration happened", pair.canonical, werr, preserved, pair.legacy)
		}
		return fmt.Errorf("write %s: %w", pair.canonical, werr)
	}

	// REQ-CIF-010 final step: a single atomic rename of the original into a
	// fresh timestamped backup path — never a re-check-then-remove pair.
	finalPath, err := freshBackupPath(root, pair.legacy, now)
	if err != nil {
		return preserveOrReport(root, pair, legacyPath, canonPath, now, fmt.Sprintf("create final backup path: %v", err))
	}
	if err := migrateRename(legacyPath, finalPath); err != nil {
		return preserveOrReport(root, pair, legacyPath, canonPath, now, fmt.Sprintf("final atomic rename: %v", err))
	}

	_, _ = fmt.Fprintf(out, "Migrated %s to %s (backup copy: %s; original preserved: %s).\n",
		pair.legacy, pair.canonical, copyRel, relFromRoot(root, finalPath))
	return nil
}

// preserveOrReport preserves the destination this process created (the O_EXCL
// write above is the ownership proof) into a fresh backup entry —
// unconditionally — and reports the failure without claiming a migration. If
// the preserving rename itself fails, both files stay in place and the report
// names the residual paths and the failed preservation (REQ-CIF-011
// rollback-failure exit).
func preserveOrReport(root string, pair instructionPair, legacyPath, canonPath string, now time.Time, cause string) error {
	preserved, perr := preserveIntoBackup(root, pair.canonical, canonPath, now)
	if perr != nil {
		return residualReport(pair, legacyPath, canonPath, fmt.Sprintf("%s; preserving the destination failed: %v", cause, perr))
	}
	return fmt.Errorf("%s; the destination was preserved at %s, and %s is unchanged in place — no migration happened", cause, preserved, pair.legacy)
}

// residualReport is the exit of last resort: nothing more can be moved, both
// files remain exactly where they are, and the report says so.
func residualReport(pair instructionPair, legacyPath, canonPath string, cause string) error {
	return fmt.Errorf("%s; both files are left in place (%s, %s) — resolve by hand; the failed preservation means the exact partial state is only on disk", cause, legacyPath, canonPath)
}

// preserveIntoBackup renames path into a fresh timestamped backup entry named
// name, returning the project-relative path of the preserved entry.
func preserveIntoBackup(root, name, path string, now time.Time) (string, error) {
	dst, err := freshBackupPath(root, name, now)
	if err != nil {
		return "", err
	}
	if err := migrateRename(path, dst); err != nil {
		return "", err
	}
	return relFromRoot(root, dst), nil
}

// writeLocalInstructionsBackup copies body under a fresh timestamped
// directory as name and returns the copy's project-relative path.
func writeLocalInstructionsBackup(root, name string, body []byte, perm os.FileMode, now time.Time) (string, error) {
	path, err := freshBackupPath(root, name, now)
	if err != nil {
		return "", err
	}
	if err := writeExclusive(path, body, perm); err != nil {
		return "", fmt.Errorf("write backup: %w", err)
	}
	return relFromRoot(root, path), nil
}

// freshBackupPath returns a never-before-used path under the backup root:
// a fresh timestamped directory joined with name.
func freshBackupPath(root, name string, now time.Time) (string, error) {
	base := filepath.Join(root, filepath.FromSlash(localInstructionsBackupDir))
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}
	dir, err := os.MkdirTemp(base, now.UTC().Format("20060102T150405Z")+"-")
	if err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}
	return filepath.Join(dir, name), nil
}

func relFromRoot(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

// sourceChangedSinceRead reports whether the file at path differs from the
// bytes and mtime captured at the migration's read. The re-check never
// follows a symlink (sync-audit F8): a non-unix Lstat pre-check refuses a
// link by mode, and the unix handle opens O_NOFOLLOW — a link swapped into
// the window resolves as "changed", never as its target's identical
// mtime/content. The open is NONBLOCKING and verifies the HANDLE's own
// Stat is a regular file before reading (audit P2-3): a same-mtime FIFO is
// refused by the handle's mode, never waited out by a blocking read.
func sourceChangedSinceRead(path string, body []byte, mtime time.Time) (bool, error) {
	if info, lerr := os.Lstat(path); lerr == nil && info.Mode()&os.ModeSymlink != 0 {
		return true, nil // link swap: refuse without following (REQ-CIF-009)
	}
	f, err := openInstructionHandle(path)
	if err != nil {
		return true, err // unopenable (a writer-less FIFO refuses O_NONBLOCK too; a link refuses O_NOFOLLOW)
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return true, err
	}
	if !info.Mode().IsRegular() {
		return true, nil // non-regular swap (FIFO/socket/device): refuse
	}
	if !info.ModTime().Equal(mtime) {
		return true, nil
	}
	cur, err := io.ReadAll(f)
	if err != nil {
		return true, err
	}
	return !bytes.Equal(cur, body), nil
}

// writeExclusive creates path (failing if it exists), writes body, and syncs.
// A failed write removes the partial file.
func writeExclusive(path string, body []byte, perm os.FileMode) error {
	return writeExclusiveOpts(path, body, perm, false)
}

// writeExclusiveKeepPartial is writeExclusive without the remove-on-failure:
// whatever partial bytes landed stay on disk so the caller can preserve them
// (REQ-CIF-011).
func writeExclusiveKeepPartial(path string, body []byte, perm os.FileMode) error {
	return writeExclusiveOpts(path, body, perm, true)
}

func writeExclusiveOpts(path string, body []byte, perm os.FileMode, keepPartial bool) error {
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
	if werr != nil && !keepPartial {
		_ = os.Remove(path)
	}
	return werr
}

// instructionPairPresent reports which of the pair's two files exist. Any
// directory entry counts, links included: the question is whether a second
// live read path could exist, not whether it is readable.
func instructionPairPresent(legacyPath, canonPath string) (legacy, canonical bool, err error) {
	present := func(path string) (bool, error) {
		if _, err := os.Lstat(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	}
	if legacy, err = present(legacyPath); err != nil {
		return false, false, err
	}
	if canonical, err = present(canonPath); err != nil {
		return false, false, err
	}
	return legacy, canonical, nil
}

// @MX:ANCHOR: [AUTO] One advisory text for the surfaces that report a legacy local instruction file.
// @MX:REASON: update and doctor must say the same thing (REQ-IFU-012); shared text keeps them in lockstep.
// localInstructionsAdvisory returns the operator advisory for the given file
// set, or "" when no legacy local file is present. Both advising branches
// name the migration verb.
func localInstructionsAdvisory(canonicalPresent, legacyPresent bool) string {
	switch {
	case !legacyPresent:
		return ""
	case canonicalPresent:
		return fmt.Sprintf("Advisory: %s and %s both exist; merge %s into %s by hand and remove it "+
			"(`moai migrate local-instructions` refuses while both exist).",
			canonicalLocalInstructionName, legacyLocalInstructionName, legacyLocalInstructionName, canonicalLocalInstructionName)
	default:
		return fmt.Sprintf("Advisory: %s is a legacy local instruction file; "+
			"run `moai migrate local-instructions` to move it to %s.",
			legacyLocalInstructionName, canonicalLocalInstructionName)
	}
}

// emitLocalInstructionsAdvisory writes the advisory for root to w, if any. It
// only reads: neither file is moved, renamed, or rewritten (REQ-IFU-011).
func emitLocalInstructionsAdvisory(w io.Writer, root string) {
	legacyPath := filepath.Join(root, legacyLocalInstructionName)
	canonPath := filepath.Join(root, canonicalLocalInstructionName)
	legacy, canonical, err := instructionPairPresent(legacyPath, canonPath)
	if err != nil {
		return
	}
	if msg := localInstructionsAdvisory(canonical, legacy); msg != "" {
		_, _ = fmt.Fprintln(w, msg)
	}
}

// sharedInstructionAdvisory returns the operator advisory for the SHARED
// instruction pair, or "" when no legacy AGENTS.md is present. The
// deployment layer skips creating CLAUDE.md next to a legacy AGENTS.md
// (sync-audit F1); this advisory names the migration verb as the way out.
func sharedInstructionAdvisory(canonicalPresent, legacyPresent bool) string {
	switch {
	case !legacyPresent:
		return ""
	case canonicalPresent:
		return fmt.Sprintf("Advisory: %s and %s both exist; merge %s into %s by hand and remove it "+
			"(`moai migrate local-instructions` refuses while both exist).",
			"CLAUDE.md", "AGENTS.md", "AGENTS.md", "CLAUDE.md")
	default:
		return fmt.Sprintf("Advisory: %s is a legacy instruction file and template deployment of %s was skipped to keep it active; "+
			"run `moai migrate local-instructions` to move it to %s.",
			"AGENTS.md", "CLAUDE.md", "CLAUDE.md")
	}
}

// emitSharedInstructionAdvisory writes the shared-pair advisory for root to
// w, if any. Read-only, like the local-pair emitter.
func emitSharedInstructionAdvisory(w io.Writer, root string) {
	legacyPath := filepath.Join(root, "AGENTS.md")
	canonPath := filepath.Join(root, "CLAUDE.md")
	legacy, canonical, err := instructionPairPresent(legacyPath, canonPath)
	if err != nil {
		return
	}
	if msg := sharedInstructionAdvisory(canonical, legacy); msg != "" {
		_, _ = fmt.Fprintln(w, msg)
	}
}
