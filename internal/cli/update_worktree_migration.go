package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/session"
)

// legacyWorktree is one registered Git worktree from the porcelain listing.
type legacyWorktree struct {
	path, head, branch string
	locked, prunable   bool
}

type worktreeMigrationPlan struct {
	source, destination string
	head, branch        string
	skip                string
}

func gitForWorktreeMigration(root string, args ...string) (string, error) {
	argv := append([]string{"-C", root}, args...)
	out, err := exec.Command("git", argv...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %s: %w", strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return string(out), nil
}

func listRegisteredWorktrees(root string) ([]legacyWorktree, error) {
	out, err := gitForWorktreeMigration(root, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var entries []legacyWorktree
	var current *legacyWorktree
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			entries = append(entries, legacyWorktree{path: filepath.FromSlash(strings.TrimPrefix(line, "worktree "))})
			current = &entries[len(entries)-1]
		case current == nil:
			continue
		case strings.HasPrefix(line, "HEAD "):
			current.head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			current.branch = strings.TrimPrefix(line, "branch ")
		case line == "locked" || strings.HasPrefix(line, "locked "):
			current.locked = true
		case line == "prunable" || strings.HasPrefix(line, "prunable "):
			current.prunable = true
		}
	}
	return entries, nil
}

func worktreeMigrationRegistryReadable(root, tree string) error {
	for _, base := range []string{root, tree} {
		registry := session.NewRegistry(filepath.Join(base, session.DefaultRegistryPath), nil)
		if _, err := registry.Query(""); err != nil {
			return err
		}
	}
	return nil
}

// planLegacyWorktreeMigration never writes. It names every registered tree
// below the old root and explains why a tree cannot be moved now.
func planLegacyWorktreeMigration(root string, now time.Time) ([]worktreeMigrationPlan, error) {
	return planLegacyWorktreeMigrationFor(root, now, "")
}

// planLegacyWorktreeMigrationFor checks one source when revalidating a move.
// An empty source retains the full preview used before migration begins.
func planLegacyWorktreeMigrationFor(root string, now time.Time, source string) ([]worktreeMigrationPlan, error) {
	entries, err := listRegisteredWorktrees(root)
	if err != nil {
		return nil, err
	}
	oldRoot := canonicalTreePath(filepath.Join(root, claudeNativeWorktreeSubdir))
	newRoot := filepath.Join(root, sessionWorktreeSubdir)
	for _, parent := range []string{filepath.Join(root, ".moai"), newRoot} {
		if info, statErr := os.Lstat(parent); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("worktree destination parent is a symlink: %s", parent)
		} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return nil, fmt.Errorf("inspect worktree destination parent %s: %w", parent, statErr)
		}
	}
	cwd, _ := os.Getwd()
	var plans []worktreeMigrationPlan
	var processCWDs []string
	var processCWDsRead bool
	var processCWDErr error
	for _, entry := range entries {
		if source != "" && canonicalTreePath(entry.path) != canonicalTreePath(source) {
			continue
		}
		rel, relErr := filepath.Rel(oldRoot, canonicalTreePath(entry.path))
		if relErr != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		if !processCWDsRead {
			processCWDs, processCWDErr = activeProcessCWDs()
			processCWDsRead = true
		}
		plan := worktreeMigrationPlan{
			source: entry.path, destination: filepath.Join(newRoot, rel),
			head: entry.head, branch: entry.branch,
		}
		switch {
		case entry.prunable:
			plan.skip = "Git lists this tree as prunable"
		case entry.locked:
			plan.skip = "Git worktree is locked"
		case isUnderWorktreePrefix(cwd, entry.path) || canonicalTreePath(cwd) == canonicalTreePath(entry.path):
			plan.skip = "current process is inside this worktree"
		case processCWDErr != nil:
			plan.skip = fmt.Sprintf("cannot inspect active process working directories: %v", processCWDErr)
		case processUsesWorktree(processCWDs, entry.path):
			plan.skip = "an active process is inside this worktree"
		case worktreeMigrationRegistryReadable(root, entry.path) != nil:
			plan.skip = "session registry cannot be read"
		case len(session.LiveAnchoredSessionsForProject(entry.path, root, now)) > 0:
			plan.skip = "active session is anchored in this worktree"
		default:
			if info, statErr := os.Lstat(entry.path); statErr != nil || !info.IsDir() {
				plan.skip = "source path is missing or not a directory"
			} else if _, statErr := os.Lstat(plan.destination); statErr == nil {
				plan.skip = "destination already exists"
			} else if !errors.Is(statErr, os.ErrNotExist) {
				plan.skip = "destination cannot be inspected"
			}
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

func processUsesWorktree(directories []string, tree string) bool {
	for _, directory := range directories {
		if canonicalTreePath(directory) == canonicalTreePath(tree) || isUnderWorktreePrefix(directory, tree) {
			return true
		}
	}
	return false
}

// moveLegacyWorktree uses Git's metadata-aware move and checks that the
// registered path, commit, branch, and tracked/untracked status survive.
// The caller must pass a fresh, unskipped plan and hold the update lock.
func moveLegacyWorktree(root string, plan worktreeMigrationPlan) error {
	if plan.skip != "" {
		return fmt.Errorf("worktree %s is not movable: %s", plan.source, plan.skip)
	}
	current, err := planLegacyWorktreeMigrationFor(root, time.Now(), plan.source)
	if err != nil {
		return err
	}
	fresh := false
	for _, item := range current {
		if canonicalTreePath(item.source) == canonicalTreePath(plan.source) &&
			item.destination == plan.destination && item.head == plan.head && item.branch == plan.branch && item.skip == "" {
			fresh = true
			break
		}
	}
	if !fresh {
		return fmt.Errorf("worktree %s changed since migration planning; inspect and retry", plan.source)
	}
	status, err := gitForWorktreeMigration(plan.source, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(plan.destination), 0o755); err != nil {
		return fmt.Errorf("create worktree destination parent: %w", err)
	}
	if _, err := gitForWorktreeMigration(root, "worktree", "move", plan.source, plan.destination); err != nil {
		// Git may have moved the tree before returning an error. Never retry a
		// side effect until its registered and on-disk state is known.
		entries, inspectErr := listRegisteredWorktrees(root)
		if inspectErr != nil {
			return fmt.Errorf("%w; cannot inspect migration outcome: %v", err, inspectErr)
		}
		for _, entry := range entries {
			if canonicalTreePath(entry.path) == canonicalTreePath(plan.destination) {
				return fmt.Errorf("%w; tree was registered at %s despite the error; inspect before retry", err, plan.destination)
			}
		}
		return err
	}
	rollback := func(cause error) error {
		if _, err := gitForWorktreeMigration(root, "worktree", "move", plan.destination, plan.source); err != nil {
			return fmt.Errorf("%w; rollback failed: %v; tree remains at %s", cause, err, plan.destination)
		}
		return fmt.Errorf("%w; moved tree restored to %s", cause, plan.source)
	}
	entries, err := listRegisteredWorktrees(root)
	if err != nil {
		return rollback(err)
	}
	registered := false
	for _, entry := range entries {
		if canonicalTreePath(entry.path) == canonicalTreePath(plan.destination) && entry.head == plan.head && entry.branch == plan.branch {
			registered = true
		}
		if canonicalTreePath(entry.path) == canonicalTreePath(plan.source) {
			return rollback(fmt.Errorf("old worktree path is still registered: %s", plan.source))
		}
	}
	if !registered {
		return rollback(fmt.Errorf("new worktree path is not registered with its original HEAD and branch: %s", plan.destination))
	}
	after, err := gitForWorktreeMigration(plan.destination, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return rollback(err)
	}
	if after != status {
		return rollback(fmt.Errorf("worktree status changed during migration: %s", plan.source))
	}
	return nil
}

// runUpdateWorktreeMigration is idempotent and only acts in the primary Git
// checkout. Linked worktree updates cannot move siblings out from under their
// owner. A skipped or failed tree remains available at its old path for retry.
func runUpdateWorktreeMigration(root string, dryRun bool, out io.Writer) error {
	gitDir, err := os.Lstat(filepath.Join(root, ".git"))
	if errors.Is(err, os.ErrNotExist) || err == nil && !gitDir.IsDir() {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect Git directory: %w", err)
	}
	plans, err := planLegacyWorktreeMigration(root, time.Now())
	if err != nil {
		return err
	}
	if len(plans) == 0 {
		return nil
	}
	if !dryRun {
		_, _ = fmt.Fprintf(out, "Worktree migration: checking %d legacy worktrees\n", len(plans))
	}
	var moved, skipped, failed int
	for _, plan := range plans {
		if plan.skip != "" {
			skipped++
			_, _ = fmt.Fprintf(out, "Worktree migration skipped %s: %s\n", plan.source, plan.skip)
			continue
		}
		if dryRun {
			_, _ = fmt.Fprintf(out, "Worktree migration planned %s -> %s\n", plan.source, plan.destination)
			continue
		}
		if err := moveLegacyWorktree(root, plan); err != nil {
			failed++
			_, _ = fmt.Fprintf(out, "Worktree migration failed %s: %v\n", plan.source, err)
			continue
		}
		moved++
		if moved%25 == 0 {
			_, _ = fmt.Fprintf(out, "Worktree migration: %d moved, %d skipped, %d failed\n", moved, skipped, failed)
		}
	}
	if !dryRun {
		_, _ = fmt.Fprintf(out, "Worktree migration: %d moved, %d skipped, %d failed\n", moved, skipped, failed)
	}
	return nil
}
