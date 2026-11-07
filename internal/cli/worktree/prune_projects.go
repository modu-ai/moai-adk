package worktree

// prune_projects.go — `moai worktree prune-projects`: removes registry
// project folders (<MoaiHome>/worktrees/projects/<key>) whose recorded owner
// repository has vanished. Test binaries that resolved a repository under a
// temporary directory used to leave one such folder behind per run; every
// leaked folder costs an inode and slows each registry scan.

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

const (
	pruneDead = "DEAD"
	pruneKept = "KEPT"

	// ownerlessNoticeThreshold is the ownerless-entry count above which the
	// sweep tells the operator to run prune-projects.
	ownerlessNoticeThreshold = 100
)

// pruneProjectEntry is one registry project folder's evaluation.
type pruneProjectEntry struct {
	Key     string `json:"key"`
	Owner   string `json:"owner"`
	Verdict string `json:"verdict"`
	Reason  string `json:"reason"`
	path    string
}

// projectsRoot returns <MoaiHome>/worktrees/projects.
func projectsRoot() (string, error) {
	wt, err := paths.WorktreesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(wt, "projects"), nil
}

func newPruneProjectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune-projects",
		Short: "Remove registry project folders whose owner repository no longer exists",
		Long: `Remove folders under <MoaiHome>/worktrees/projects whose owner marker
(.moai-project.json) names a repository path that no longer exists.

A folder is DEAD only when ALL hold: the owner path does not exist; the folder
is older than --older-than (default 1h); and no .git file or directory lives
anywhere inside it (a folder that still holds a worktree is never removed).
Everything else is KEPT with a reason. Previews by default; pass --yes to remove.`,
		Args: cobra.NoArgs,
		RunE: runPruneProjects,
	}
	cmd.Flags().Bool("yes", false, "Actually remove the DEAD folders instead of previewing them")
	cmd.Flags().Duration("older-than", time.Hour, "Keep folders modified more recently than this")
	cmd.Flags().Bool("json", false, "Report every folder's evaluation as JSON")
	return cmd
}

func runPruneProjects(cmd *cobra.Command, _ []string) error {
	apply, _ := cmd.Flags().GetBool("yes")
	asJSON, _ := cmd.Flags().GetBool("json")
	olderThan, _ := cmd.Flags().GetDuration("older-than")

	root, err := projectsRoot()
	if err != nil {
		return fmt.Errorf("resolve projects root: %w", err)
	}
	entries, err := evaluateProjects(root, olderThan, time.Now())
	if err != nil {
		return err
	}

	removed, failed := 0, 0
	if apply {
		for _, e := range entries {
			if e.Verdict != pruneDead {
				continue
			}
			if rerr := os.RemoveAll(e.path); rerr != nil {
				failed++
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "prune-projects: remove %s: %v\n", e.Key, rerr)
				continue
			}
			removed++
		}
	}

	out := cmd.OutOrStdout()
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	}
	dead := 0
	for _, e := range entries {
		if e.Verdict == pruneDead {
			dead++
		}
	}
	_, _ = fmt.Fprintf(out, "%d project folder(s) evaluated, %d dead.\n", len(entries), dead)
	for _, e := range entries {
		if e.Verdict == pruneDead || dead == 0 {
			_, _ = fmt.Fprintf(out, "  %s %s owner=%q: %s\n", e.Verdict, e.Key, e.Owner, e.Reason)
		}
	}
	switch {
	case apply:
		_, _ = fmt.Fprintf(out, "Removed %d folder(s), %d failed.\n", removed, failed)
	case dead > 0:
		_, _ = fmt.Fprintln(out, "This was a preview. Re-run with --yes to remove the DEAD folders.")
	}
	return nil
}

// evaluateProjects classifies every folder under root. A missing root is an
// empty result, not an error.
func evaluateProjects(root string, olderThan time.Duration, now time.Time) ([]pruneProjectEntry, error) {
	dirents, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", root, err)
	}
	var out []pruneProjectEntry
	for _, d := range dirents {
		if !d.IsDir() {
			continue
		}
		p := filepath.Join(root, d.Name())
		e := pruneProjectEntry{Key: d.Name(), path: p}
		e.Verdict, e.Owner, e.Reason = classifyProject(p, olderThan, now)
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}

// classifyProject returns a DEAD verdict only when every predicate affirms
// it; any unreadable fact keeps the folder.
func classifyProject(dir string, olderThan time.Duration, now time.Time) (verdict, owner, reason string) {
	owner, ok := worktreeroot.OwnerMarkerOf(dir)
	if !ok || owner == "" {
		return pruneKept, owner, "owner marker absent, unreadable or empty"
	}
	if !filepath.IsAbs(owner) {
		// writeMarker records canonical absolute paths only; a relative owner
		// would be judged against the working directory, so it is never DEAD.
		return pruneKept, owner, "owner path is not absolute"
	}
	if _, err := os.Stat(owner); err == nil {
		return pruneKept, owner, "owner path exists"
	} else if !os.IsNotExist(err) {
		return pruneKept, owner, fmt.Sprintf("owner path unverifiable: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		return pruneKept, owner, fmt.Sprintf("folder unverifiable: %v", err)
	}
	if now.Sub(info.ModTime()) < olderThan {
		return pruneKept, owner, fmt.Sprintf("modified within %s", olderThan)
	}
	if found, err := holdsGitEntry(dir); err != nil {
		return pruneKept, owner, fmt.Sprintf("content unverifiable: %v", err)
	} else if found {
		return pruneKept, owner, "holds a .git entry (possible live worktree)"
	}
	return pruneDead, owner, "owner path is gone and the folder holds no worktree"
}

// holdsGitEntry reports whether any .git file or directory exists below dir.
func holdsGitEntry(dir string) (bool, error) {
	found := false
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == ".git" && p != dir {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	return found, err
}

// ownerlessProjectCount counts folders under root whose marker names a
// vanished owner path. Folders without a readable marker are not counted.
func ownerlessProjectCount(root string) int {
	dirents, err := os.ReadDir(root)
	if err != nil {
		return 0
	}
	n := 0
	for _, d := range dirents {
		if !d.IsDir() {
			continue
		}
		owner, ok := worktreeroot.OwnerMarkerOf(filepath.Join(root, d.Name()))
		if !ok || owner == "" {
			continue
		}
		if _, err := os.Stat(owner); os.IsNotExist(err) {
			n++
		}
	}
	return n
}

// noticeOwnerlessProjects prints the one-line sweep hint when the registry
// holds more than ownerlessNoticeThreshold ownerless folders.
func noticeOwnerlessProjects(cmd *cobra.Command) {
	root, err := projectsRoot()
	if err != nil {
		return
	}
	if n := ownerlessProjectCount(root); n > ownerlessNoticeThreshold {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "worktree sweep: %d project entries have no owner; run `moai worktree prune-projects` to remove them\n", n)
	}
}
