package worktree

// list.go — SPEC-WORKTREE-HOME-ROOT-001 M4: `moai worktree list [--json]`
// (REQ-WHR-010).
//
// Read-only enumeration of every worktree registered with the repository.
// The classification is the M1 shared classifier (classify.go), so list and
// migrate cannot disagree; the collapse of two spellings of one directory
// is the M1 spelling-safe equality (Same); the dirty flag, the
// not-on-any-remote commit count and the anchored flag are read per tree
// from git and the session registry. The command writes nothing.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
	"github.com/spf13/cobra"
)

// listRow is one row of the inventory: REQ-WHR-010's nine reported fields.
type listRow struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Root     string `json:"root"`
	Class    string `json:"class"`
	Branch   string `json:"branch"`
	Head     string `json:"head"`
	Dirty    bool   `json:"dirty"`
	Unpushed int    `json:"unpushed"`
	Anchored bool   `json:"anchored"`
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List every worktree registered with the repository",
		Long: `List every worktree registered with the repository, classified by
creation class through the shared classifier (the same function the
migration and the disposal verbs use), reporting for each tree its name,
path, root kind, class, branch, short HEAD, dirty flag, unpushed-commit
count and anchored flag. Read-only: the command writes nothing.
With --json the inventory is a JSON array of row objects.`,
		Args: cobra.NoArgs,
		RunE: runList,
	}
	cmd.Flags().Bool("json", false, "emit the inventory as a JSON array")
	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve the working directory: %w", err)
	}
	rows, err := buildListRows(cwd)
	if err != nil {
		return err
	}
	asJSON, _ := cmd.Flags().GetBool("json")
	if asJSON {
		body, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(body))
		return nil
	}
	for _, r := range rows {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\t%s\t%s\tdirty=%v unpushed=%d anchored=%v\n",
			r.Name, r.Class, r.Root, r.Branch, r.Head, r.Path, r.Dirty, r.Unpushed, r.Anchored)
	}
	return nil
}

// buildListRows enumerates the registered worktrees of root's repository
// and classifies each through the shared classifier.
func buildListRows(root string) ([]listRow, error) {
	commonDir, err := listGit(root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return nil, fmt.Errorf("resolve the repository of %s: %w", root, err)
	}
	primary := filepath.Dir(commonDir)
	porcelain, err := listGit(root, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("list the worktrees of %s: %w", root, err)
	}
	records, err := ParseWorktreeList(porcelain)
	if err != nil {
		return nil, err
	}

	ctx := ClassifyContext{
		PrimaryRoot:       primary,
		IntegrationBranch: config.LoadGitFlowIntegrationConfig(primary).IntegrationTarget,
		HomeProjectsDir:   homeProjectsDir(),
		CodexWorktreesDir: codexWorktreesDir(),
		MarkerOwners:      homeMarkerOwners(homeProjectsDir()),
	}
	now := time.Now()
	var rows []listRow
	for _, rec := range records {
		if rec.Bare {
			// A bare registration has no working tree to enumerate.
			continue
		}
		path := filepath.Clean(rec.Path)
		dup := false
		for _, r := range rows {
			if worktreeroot.Same(r.Path, path) {
				dup = true // two spellings of one directory are one tree
				break
			}
		}
		if dup {
			continue
		}
		cls := Classify(WorktreeRecord{Path: path, Branch: rec.Branch}, ctx)
		// REQ-WHR-010's root-kind enum has five values; the primary
		// checkout's unpinned kind reports as "other".
		rootKind := string(cls.Root)
		if rootKind == "" {
			rootKind = "other"
		}
		head := rec.Head
		if len(head) > 7 {
			head = head[:7]
		}
		rows = append(rows, listRow{
			Name:     filepath.Base(path),
			Path:     path,
			Root:     rootKind,
			Class:    string(cls.Class),
			Branch:   rec.Branch,
			Head:     head,
			Dirty:    listIsDirty(path),
			Unpushed: listUnpushedCount(path),
			Anchored: len(session.LiveAnchoredSessions(path, now)) > 0,
		})
	}
	return rows, nil
}

// listGit runs one git command in dir and returns its stdout.
func listGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", "-C", dir)
	cmd.Args = append(cmd.Args, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// listIsDirty reports whether the tree at path holds uncommitted or
// untracked work.
func listIsDirty(path string) bool {
	out, err := listGit(path, "status", "--porcelain")
	if err != nil {
		return false
	}
	return out != ""
}

// listUnpushedCount counts the commits of the tree's HEAD that no
// remote-tracking ref carries.
func listUnpushedCount(path string) int {
	out, err := listGit(path, "rev-list", "--count", "HEAD", "--not", "--remotes")
	if err != nil {
		return 0
	}
	n := 0
	fmt.Sscanf(out, "%d", &n)
	return n
}

func homeProjectsDir() string {
	home, err := paths.MoaiHome()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "worktrees", "projects")
}

func codexWorktreesDir() string {
	home, err := paths.Home()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".codex", "worktrees")
}

// homeMarkerOwners reads every project folder's owner marker under the home
// projects directory — the classifier's marker witness as data.
func homeMarkerOwners(projectsDir string) map[string]string {
	if projectsDir == "" {
		return nil
	}
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil
	}
	owners := make(map[string]string, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if owner, ok := worktreeroot.OwnerMarkerOf(filepath.Join(projectsDir, e.Name())); ok {
			owners[e.Name()] = owner
		}
	}
	return owners
}
