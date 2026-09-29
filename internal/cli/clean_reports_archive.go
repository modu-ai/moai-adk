package cli

// Reports-archive action for `moai clean` — SPEC-REPORTS-LIFECYCLE-001
// REQ-RLC-007 (AC-RLC-009 / AC-RLC-010).
//
// Move-only: the action MOVES aging evidence directories into
// .moai/reports/archive/<YYYY-MM>/ and never deletes anything. Selection is
// a default-deny predicate over top-level entries of .moai/reports/ — a
// candidate must match ALL of:
//  1. evidence-shaped name: t<digits> or SPEC-<DOMAIN>-<NNN>;
//  2. mtime older than the retention window (default
//     config.DefaultReportsArchiveRetentionDays, --reports-archive-days);
//  3. zero git-tracked files (git ls-files) — the t338/t528/t530/t229
//     fixtures protect themselves by this rule.
//
// Explicitly protected entries — historical/ (the relocated legacy reports),
// plan-audit/ (audit records), worktrees/ (hoist output), archive/ (the
// destination itself) — are out of scope by construction, whatever their
// name or age.
//
// Like every `moai clean` scope, the default is a report-only dry-run and
// --force performs the move.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/config"
)

// evidenceNameRe is the evidence-directory naming shape (REQ-RLC-007).
var evidenceNameRe = regexp.MustCompile(`^(t[0-9]+|SPEC-[A-Z0-9-]+[0-9]{3})$`)

// protectedReportsEntries are never archive candidates, whatever the
// predicate says (REQ-RLC-007 explicit protection set).
var protectedReportsEntries = map[string]bool{
	"historical": true,
	"plan-audit": true,
	"worktrees":  true,
	"archive":    true,
}

// reportsTrackedFilesFunc reports whether rel (slash-separated, relative to
// root) contains git-tracked files. An unreadable answer conservatively
// reports tracked (default-deny: unable to prove zero is not zero).
// Overridable in tests.
var reportsTrackedFilesFunc = func(root, rel string) (bool, error) {
	out, err := exec.Command("git", "-C", root, "ls-files", "--", rel).Output()
	if err != nil {
		return true, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// reportsArchiveWarnNeeded is the 1GiB advisory predicate, factored out so
// the threshold logic is directly testable without writing a gigabyte.
func reportsArchiveWarnNeeded(totalBytes, warnThreshold int64) bool {
	return totalBytes > warnThreshold
}

type archiveCandidate struct {
	name    string
	path    string
	size    int64
	modTime time.Time
}

// runCleanReportsArchive is the clean-command entry; it resolves the project
// root and delegates. The root comes from git (the reports directory lives
// in the repository), resolved fresh — a state-dir-style resolution would
// silently archive the wrong checkout when an env var names another tree.
func runCleanReportsArchive(p printer.Printer, force bool, retentionDays int) error {
	root, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return fmt.Errorf("reports-archive requires a git repository: %w", err)
	}
	return runCleanReportsArchiveWithRoot(p, force, retentionDays, strings.TrimSpace(string(root)))
}

// runCleanReportsArchiveWithRoot runs the action against an explicit project
// root (the tested form).
func runCleanReportsArchiveWithRoot(p printer.Printer, force bool, retentionDays int, root string) error {
	reportsDir := filepath.Join(root, ".moai", "reports")
	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		if os.IsNotExist(err) {
			p.Info("no .moai/reports/ directory; nothing to archive")
			return nil
		}
		return fmt.Errorf("read reports dir: %w", err)
	}
	if retentionDays <= 0 {
		retentionDays = config.DefaultReportsArchiveRetentionDays
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	var candidates []archiveCandidate
	var totalBytes int64
	for _, entry := range entries {
		name := entry.Name()
		if protectedReportsEntries[name] {
			continue
		}
		if !evidenceNameRe.MatchString(name) {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			continue // unreadable entry is not a provable candidate
		}
		if !info.ModTime().Before(cutoff) {
			continue
		}
		rel := filepath.ToSlash(filepath.Join(".moai", "reports", name))
		tracked, trackedErr := reportsTrackedFilesFunc(root, rel)
		if trackedErr != nil || tracked {
			continue // unable to prove zero tracked files is not zero
		}
		candidates = append(candidates, archiveCandidate{name: name, path: filepath.Join(reportsDir, name), size: info.Size(), modTime: info.ModTime()})
		totalBytes += info.Size()
	}

	if len(candidates) == 0 {
		p.Info("no archive candidates under .moai/reports/ (older than %d days, evidence-named, untracked)", retentionDays)
		return nil
	}
	if reportsArchiveWarnNeeded(totalBytes, config.DefaultReportsArchiveWarnBytes) {
		p.Warn("%d candidate(s) total %d bytes exceeds the %d-byte advisory threshold — review the list before --force",
			len(candidates), totalBytes, config.DefaultReportsArchiveWarnBytes)
	}

	if !force {
		p.Info("[dry-run] %d candidate(s), %d byte(s) would move to .moai/reports/archive/<YYYY-MM>/ — run with --force to move",
			len(candidates), totalBytes)
		return nil
	}

	movedBytes := int64(0)
	moved := 0
	for _, c := range candidates {
		shard := filepath.Join(reportsDir, "archive", c.modTime.Format("2006-01"))
		if err := os.MkdirAll(shard, 0o755); err != nil {
			return fmt.Errorf("create archive shard: %w", err)
		}
		if err := os.Rename(c.path, filepath.Join(shard, c.name)); err != nil {
			p.Warn("failed to move %s: %v", c.name, err)
			continue
		}
		moved++
		movedBytes += c.size
	}
	p.Info("Archived %d candidate(s), %d byte(s) into .moai/reports/archive/", moved, movedBytes)
	return nil
}
